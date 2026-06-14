package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

// feed applies a sequence of messages to the model, discarding commands (which
// would otherwise hit a live daemon). It lets us exercise Update/View logic
// deterministically without a TTY or Docker.
func feed(t *testing.T, m model, msgs ...tea.Msg) model {
	t.Helper()
	for _, msg := range msgs {
		nm, _ := m.Update(msg)
		m = nm.(model)
	}
	return m
}

func keyRunes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func sampleContainers() containersMsg {
	return containersMsg{
		{ID: "id-web", Name: "ct-web", Image: "nginx:alpine", State: "running", Status: "Up 2 min", Ports: "8088→80/tcp"},
		{ID: "id-db", Name: "ct-db", Image: "postgres:16", State: "exited", Status: "Exited (0)"},
		{ID: "id-cache", Name: "ct-cache", Image: "redis:7", State: "paused", Status: "Paused"},
	}
}

func ready(t *testing.T) model {
	m := New(nil)
	return feed(t, m,
		tea.WindowSizeMsg{Width: 120, Height: 40},
		connectedMsg{ver: "29.5.3"},
		sampleContainers(),
	)
}

func TestFunctional_ListView(t *testing.T) {
	m := ready(t)
	out := m.View()
	for _, want := range []string{"Contalyst", "Containers", "ct-web", "running", "8088→80/tcp", "start/stop", "docker 29.5.3"} {
		if !strings.Contains(out, want) {
			t.Errorf("list view missing %q\n---\n%s", want, out)
		}
	}
}

func TestFunctional_DetailView(t *testing.T) {
	m := ready(t)
	nm, _ := m.enterDetail("id-web", "ct-web")
	m = nm.(model)
	m = feed(t, m,
		logLineMsg(dockerx.LogLine{Text: "server started on :80"}),
		statsMsg(dockerx.Stats{CPUPercent: 12.5, MemUsage: 256 << 20, MemLimit: 1 << 30, MemPercent: 25, Pids: 7}),
	)
	out := m.View()
	for _, want := range []string{"Logs", "Stats", "server started on :80", "CPU", "12.5%", "follow"} {
		if !strings.Contains(out, want) {
			t.Errorf("detail view missing %q\n---\n%s", want, out)
		}
	}
}

func TestFunctional_HelpOverlay(t *testing.T) {
	m := ready(t)
	m = feed(t, m, keyRunes("?"))
	if m.overlay != ovHelp {
		t.Fatalf("expected help overlay, got %v", m.overlay)
	}
	out := m.View()
	if !strings.Contains(out, "keybindings") || !strings.Contains(out, "exec shell") {
		t.Errorf("help overlay missing content\n---\n%s", out)
	}
}

func TestFunctional_ConfirmDialogSafeDefault(t *testing.T) {
	m := ready(t)
	m = feed(t, m, keyRunes("d")) // delete selected container
	if m.overlay != ovConfirm {
		t.Fatalf("expected confirm overlay, got %v", m.overlay)
	}
	if m.confirm.yes {
		t.Error("confirm dialog must default to the safe (Cancel) option")
	}
	out := m.View()
	if !strings.Contains(out, "Remove container") || !strings.Contains(out, "ct-web") {
		t.Errorf("confirm dialog missing content\n---\n%s", out)
	}
	// Esc cancels.
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.overlay != ovNone {
		t.Error("esc should close the confirm dialog")
	}
}

func TestFunctional_ThemeCycle(t *testing.T) {
	m := ready(t)
	first := m.th.Name
	m = feed(t, m, keyRunes("T"))
	if m.th.Name == first {
		t.Errorf("theme did not change from %q", first)
	}
}

func TestFunctional_Filter(t *testing.T) {
	m := ready(t)
	m = feed(t, m, keyRunes("/"))
	if m.overlay != ovFilter {
		t.Fatalf("expected filter overlay")
	}
	m = feed(t, m, keyRunes("web"))
	if m.filter != "web" {
		t.Fatalf("filter not applied: %q", m.filter)
	}
	// Only ct-web should remain.
	if got := len(m.lst.rows); got != 1 {
		t.Errorf("expected 1 filtered row, got %d", got)
	}
}

func TestFunctional_SwitchToImages(t *testing.T) {
	m := ready(t)
	// Open command palette, type "images", enter.
	m = feed(t, m, keyRunes(":"))
	m.cmdInput.SetValue("images")
	nm, _ := m.runCommand("images")
	m = nm.(model)
	if m.kind != kindImages {
		t.Fatalf("expected images kind, got %v", m.kind)
	}
	m = feed(t, m, imagesMsg{{ID: "abc123", Repo: "nginx", Tag: "alpine", Size: 12 << 20}})
	out := m.View()
	for _, want := range []string{"Images", "REPOSITORY", "nginx", "alpine"} {
		if !strings.Contains(out, want) {
			t.Errorf("images view missing %q\n---\n%s", want, out)
		}
	}
}

func TestFunctional_LogSearch(t *testing.T) {
	m := ready(t)
	nm, _ := m.enterDetail("id-web", "ct-web")
	m = nm.(model)
	m = feed(t, m,
		logLineMsg(dockerx.LogLine{Text: "listening on :80"}),
		logLineMsg(dockerx.LogLine{Text: "ERROR: connection refused"}),
		logLineMsg(dockerx.LogLine{Text: "retrying error in 1s"}),
	)
	// Open search and type a query that matches two lines (case-insensitive).
	m = feed(t, m, keyRunes("/"))
	if m.overlay != ovLogSearch {
		t.Fatalf("expected log-search overlay, got %v", m.overlay)
	}
	m = feed(t, m, keyRunes("error"))
	if m.detail.search != "error" {
		t.Fatalf("search query not applied: %q", m.detail.search)
	}
	if got := len(m.detail.matches); got != 2 {
		t.Fatalf("expected 2 matches, got %d (%v)", got, m.detail.matches)
	}
	// Bottom bar reports the match count while searching.
	if out := m.bottomView(); !strings.Contains(out, "1/2") {
		t.Errorf("search bar missing match counter\n---\n%s", out)
	}
	// Enter commits the search; n/N cycle through matches.
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.overlay != ovNone {
		t.Fatalf("enter should close the search overlay")
	}
	m = feed(t, m, keyRunes("n"))
	if m.detail.matchIdx != 1 {
		t.Errorf("n should advance to match 1, got %d", m.detail.matchIdx)
	}
	m = feed(t, m, keyRunes("n")) // wraps back to first
	if m.detail.matchIdx != 0 {
		t.Errorf("n should wrap to match 0, got %d", m.detail.matchIdx)
	}
	// Esc clears the search.
	m = feed(t, m, keyRunes("/"), tea.KeyMsg{Type: tea.KeyEsc})
	if m.detail.search != "" || len(m.detail.matches) != 0 {
		t.Errorf("esc should clear the search, got %q / %v", m.detail.search, m.detail.matches)
	}
}

func TestFunctional_CompactHintsToggle(t *testing.T) {
	m := ready(t)
	if m.compactHints {
		t.Fatal("compact hints should start disabled")
	}
	m = feed(t, m, keyRunes("H"))
	if !m.compactHints {
		t.Error("H should enable compact hints")
	}
	// In compact mode the hint bar is a single line.
	if got := strings.Count(m.hintView(), "\n"); got != 0 {
		t.Errorf("compact hint bar should be one line, got %d newlines", got)
	}
	m = feed(t, m, keyRunes("H"))
	if m.compactHints {
		t.Error("H should toggle compact hints back off")
	}
}

func TestFunctional_FrameToggle(t *testing.T) {
	m := ready(t)
	if !m.rounded {
		t.Fatal("frames should start rounded")
	}
	m = feed(t, m, keyRunes("F"))
	if m.rounded {
		t.Error("F should switch to square frames")
	}
	if m.toast == "" || !strings.Contains(m.toast, "square") {
		t.Errorf("frame toggle should announce the mode, got %q", m.toast)
	}
	m = feed(t, m, keyRunes("F"))
	if !m.rounded {
		t.Error("F should switch back to rounded frames")
	}
}

func TestFunctional_TinyTerminalNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("View panicked at tiny size: %v", r)
		}
	}()
	m := New(nil)
	m = feed(t, m,
		tea.WindowSizeMsg{Width: 20, Height: 6},
		connectedMsg{ver: "29"},
		sampleContainers(),
	)
	_ = m.View()
	// Also detail at tiny size.
	nm, _ := m.enterDetail("id-web", "ct-web")
	m = nm.(model)
	_ = m.View()
}
