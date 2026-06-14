package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/engine"
)

func sampleComposeContainers() containersMsg {
	return containersMsg{
		{ID: "c1", Name: "app-web-1", Image: "nginx:alpine", State: "running", Project: "app", Service: "web"},
		{ID: "c2", Name: "app-db-1", Image: "postgres:16", State: "exited", Project: "app", Service: "db"},
		{ID: "c3", Name: "data-minio-1", Image: "minio", State: "running", Project: "data", Service: "minio"},
	}
}

func TestFunctional_ComposeProjectListAndDrillDown(t *testing.T) {
	m := ready(t)
	// Switch to the compose view and feed compose-labelled containers.
	nm, _ := m.runCommand("compose")
	m = nm.(model)
	if m.kind != kindCompose {
		t.Fatalf("expected compose kind, got %v", m.kind)
	}
	m = feed(t, m, sampleComposeContainers())

	out := m.View()
	for _, want := range []string{"Compose", "PROJECT", "app", "data", "degraded"} {
		if !strings.Contains(out, want) {
			t.Errorf("compose view missing %q\n---\n%s", want, out)
		}
	}

	// Drill into the first project (app, sorted first) → scoped container list.
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.composeScope != "app" {
		t.Fatalf("expected composeScope=app, got %q", m.composeScope)
	}
	if m.kind != kindContainers {
		t.Fatalf("drill-down should switch to containers, got %v", m.kind)
	}
	m = feed(t, m, sampleComposeContainers()) // reload after setKind
	out = m.View()
	if !strings.Contains(out, "app-web-1") || !strings.Contains(out, "app-db-1") {
		t.Errorf("scoped list should show app services\n---\n%s", out)
	}
	if strings.Contains(out, "data-minio-1") {
		t.Errorf("scoped list must not show other projects\n---\n%s", out)
	}
	// Esc returns to the project list.
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.composeScope != "" || m.kind != kindCompose {
		t.Errorf("esc should return to compose list, scope=%q kind=%v", m.composeScope, m.kind)
	}
}

func TestFunctional_BulkMarkAndDeleteConfirm(t *testing.T) {
	m := ready(t)
	// Mark the first two rows with space (the mark advances the cursor).
	m = feed(t, m, keyRunes(" "), keyRunes(" "))
	if len(m.marked) != 2 {
		t.Fatalf("expected 2 marked, got %d (%v)", len(m.marked), m.marked)
	}
	// The hint bar switches to the bulk affordances.
	if !strings.Contains(m.hintView(), "stop") {
		t.Errorf("bulk hint bar missing stop affordance:\n%s", m.hintView())
	}
	// Bulk delete opens a count-aware confirm with the safe default.
	m = feed(t, m, keyRunes("d"))
	if m.overlay != ovConfirm {
		t.Fatalf("bulk delete should open confirm, got %v", m.overlay)
	}
	if m.confirm.yes {
		t.Error("bulk confirm must default to the safe option")
	}
	if !strings.Contains(m.confirm.body, "2 containers") {
		t.Errorf("confirm body should report the count, got %q", m.confirm.body)
	}
}

func TestFunctional_BulkMarkAllAndClear(t *testing.T) {
	m := ready(t)
	m = feed(t, m, keyRunes("a"))
	if len(m.marked) != len(m.lst.rows) || len(m.marked) == 0 {
		t.Fatalf("mark-all should mark every row, got %d/%d", len(m.marked), len(m.lst.rows))
	}
	m = feed(t, m, keyRunes("a"))
	if len(m.marked) != 0 {
		t.Errorf("mark-all again should clear, got %d", len(m.marked))
	}
	// Esc also clears marks.
	m = feed(t, m, keyRunes("a"))
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if len(m.marked) != 0 {
		t.Errorf("esc should clear marks, got %d", len(m.marked))
	}
}

func TestFunctional_ContextSwitcherOverlay(t *testing.T) {
	m := ready(t)
	m = feed(t, m, contextsMsg{
		{Name: "default", Host: "unix:///var/run/docker.sock", Current: true},
		{Name: "prod", Host: "ssh://prod"},
	})
	if m.overlay != ovContext {
		t.Fatalf("expected context overlay, got %v", m.overlay)
	}
	if m.contextName != "default" {
		t.Errorf("active context should be the current one, got %q", m.contextName)
	}
	out := m.View()
	for _, want := range []string{"Switch context", "default", "prod"} {
		if !strings.Contains(out, want) {
			t.Errorf("context overlay missing %q\n---\n%s", want, out)
		}
	}
	// Header reflects the active host.
	if !strings.Contains(m.headerView(), "default") {
		t.Errorf("header missing active context\n---\n%s", m.headerView())
	}
	// Down then esc closes.
	m = feed(t, m, keyRunes("j"))
	if m.contextCursor != 1 {
		t.Errorf("j should move context cursor, got %d", m.contextCursor)
	}
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.overlay != ovNone {
		t.Error("esc should close the context overlay")
	}
}

func TestFunctional_PruneDashboard(t *testing.T) {
	m := ready(t)
	m = feed(t, m, pruneUsageMsg{usage: []engine.Usage{
		{Kind: engine.PruneKindImages, Count: 12, Reclaimable: 1900 << 20},
		{Kind: engine.PruneKindBuildCache, Count: 5, Reclaimable: 940 << 20},
		{Kind: engine.PruneKindVolumes, Count: 4, Reclaimable: 320 << 20},
	}})
	if m.overlay != ovPrune {
		t.Fatalf("expected prune overlay, got %v", m.overlay)
	}
	out := m.View()
	for _, want := range []string{"Prune", "images", "build cache", "items"} {
		if !strings.Contains(out, want) {
			t.Errorf("prune dashboard missing %q\n---\n%s", want, out)
		}
	}
	// Select the first category and confirm → count-aware confirm dialog.
	m = feed(t, m, keyRunes(" "))
	if !m.pruneSel[0] {
		t.Error("space should select the cursor category")
	}
	m = feed(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.overlay != ovConfirm {
		t.Fatalf("enter with a selection should confirm, got %v", m.overlay)
	}
	if !strings.Contains(m.confirm.body, "reclaim") {
		t.Errorf("confirm should report reclaimable space, got %q", m.confirm.body)
	}
}

func TestFunctional_OperationLog(t *testing.T) {
	m := ready(t)
	// A successful and a failed action both get recorded.
	m = feed(t, m, actionDoneMsg{ok: "started ct-web"})
	m = feed(t, m, actionDoneMsg{err: errString("boom")})
	if len(m.opLog) != 2 {
		t.Fatalf("expected 2 op-log entries, got %d", len(m.opLog))
	}
	m = feed(t, m, keyRunes("@"))
	if m.overlay != ovOpLog {
		t.Fatalf("@ should open the op log, got %v", m.overlay)
	}
	out := m.View()
	for _, want := range []string{"Operation log", "started ct-web", "boom"} {
		if !strings.Contains(out, want) {
			t.Errorf("op log missing %q\n---\n%s", want, out)
		}
	}
}

func TestFunctional_ImageLayerView(t *testing.T) {
	m := ready(t)
	m = feed(t, m, layersMsg{title: "nginx:alpine", text: formatLayers([]engine.Layer{
		{Size: 78 << 20, CreatedBy: "/bin/sh -c apt-get install -y build-essential"},
		{Size: 12 << 20, CreatedBy: "COPY . /app"},
	})})
	if m.state != stateInspect {
		t.Fatalf("layers should open the inspect-style view, got %v", m.state)
	}
	out := m.View()
	for _, want := range []string{"Layers: nginx:alpine", "SIZE", "78.0MB", "build-essential"} {
		if !strings.Contains(out, want) {
			t.Errorf("layer view missing %q\n---\n%s", want, out)
		}
	}
}

// errString is a tiny error type for tests.
type errString string

func (e errString) Error() string { return string(e) }
