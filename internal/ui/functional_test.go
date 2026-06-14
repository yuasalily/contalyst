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
