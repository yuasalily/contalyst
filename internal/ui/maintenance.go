package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/engine"
)

const maxOpLog = 200

// opEntry is one recorded operation in the session operation log (U12 /
// FR-OL1). The log lives only in memory for the session (OQ-6).
type opEntry struct {
	when time.Time
	text string
	ok   bool
}

// recordOp appends an action result to the operation log (ring-buffered).
func (m *model) recordOp(text string, ok bool) {
	m.opLog = append(m.opLog, opEntry{when: time.Now(), text: text, ok: ok})
	if len(m.opLog) > maxOpLog {
		m.opLog = m.opLog[len(m.opLog)-maxOpLog:]
	}
}

func (m model) opLogBox() string {
	title := m.s.appName.Render("Operation log") + m.s.hintDesc.Render("  — esc / @ to close")
	if len(m.opLog) == 0 {
		body := m.s.empty.Render("no operations yet")
		return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
	}
	// Show the most recent entries last; cap to what fits a sensible box.
	start := 0
	if len(m.opLog) > 18 {
		start = len(m.opLog) - 18
	}
	var lines []string
	for _, e := range m.opLog[start:] {
		ts := m.s.hintDesc.Render(e.when.Format("15:04:05"))
		mark := m.s.toastInfo.Render("✓")
		text := e.text
		if !e.ok {
			mark = m.s.toastErr.Render("✕")
		}
		lines = append(lines, ts+" "+mark+" "+text)
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// --- prune dashboard (U12 / FR-PR1, FR-PR2) ---

// pruneUsageMsg carries the reclaimable-space summary for the dashboard.
type pruneUsageMsg struct {
	usage []engine.Usage
	err   error
}

func pruneUsageCmd(c engine.Engine) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		u, err := c.DiskUsage(ctx)
		return pruneUsageMsg{usage: u, err: err}
	}
}

func (m model) updatePrune(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Back), key.Matches(msg, k.Quit):
		m.overlay = ovNone
		return m, nil
	case key.Matches(msg, k.Up):
		if m.pruneCursor > 0 {
			m.pruneCursor--
		}
		return m, nil
	case key.Matches(msg, k.Down):
		if m.pruneCursor < len(m.pruneUsage)-1 {
			m.pruneCursor++
		}
		return m, nil
	case key.Matches(msg, k.Mark):
		if m.pruneCursor >= 0 && m.pruneCursor < len(m.pruneSel) {
			m.pruneSel[m.pruneCursor] = !m.pruneSel[m.pruneCursor]
		}
		return m, nil
	case key.Matches(msg, k.Enter):
		return m.confirmPrune()
	}
	return m, nil
}

// confirmPrune gathers the selected categories and opens a confirm dialog
// reporting how much space will be reclaimed (R12: explicit + safe default).
func (m model) confirmPrune() (tea.Model, tea.Cmd) {
	var kinds []engine.PruneKind
	var total int64
	var labels []string
	for i, sel := range m.pruneSel {
		if sel && i < len(m.pruneUsage) {
			u := m.pruneUsage[i]
			kinds = append(kinds, u.Kind)
			total += u.Reclaimable
			labels = append(labels, u.Kind.Label())
		}
	}
	if len(kinds) == 0 {
		return m, m.setToast("select a category to prune (space)", true)
	}
	cl := m.client
	body := fmt.Sprintf("%s — reclaim ~%s", strings.Join(labels, ", "), humanSize(total))
	m.overlay = ovNone
	m.openConfirm("Prune", body, true, bulkPrune(cl, kinds))
	return m, nil
}

// bulkPrune prunes every selected category and reports an aggregated result.
func bulkPrune(cl engine.Engine, kinds []engine.PruneKind) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		failed := 0
		var firstErr error
		for _, k := range kinds {
			if err := cl.Prune(ctx, k); err != nil {
				failed++
				if firstErr == nil {
					firstErr = err
				}
			}
		}
		done := len(kinds) - failed
		if failed > 0 {
			return actionDoneMsg{err: fmt.Errorf("prune: %d ok, %d failed — %v", done, failed, firstErr)}
		}
		return actionDoneMsg{ok: fmt.Sprintf("pruned %d categor%s", done, plural(done))}
	}
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

func (m model) pruneBox() string {
	title := m.s.appName.Render("Prune") + m.s.hintDesc.Render("  — space select · ⏎ prune · esc")
	var selTotal int64
	var lines []string
	for i, u := range m.pruneUsage {
		box := "[ ]"
		if i < len(m.pruneSel) && m.pruneSel[i] {
			box = m.s.markGutter.Render("[✓]")
			selTotal += u.Reclaimable
		}
		size := "–"
		if u.Reclaimable > 0 {
			size = humanSize(u.Reclaimable)
		}
		row := fmt.Sprintf("%s %s %s %s",
			box,
			pad(u.Kind.Label(), 18),
			pad(size, 10),
			m.s.hintDesc.Render(fmt.Sprintf("%d items", u.Count)),
		)
		if i == m.pruneCursor {
			row = m.s.rowSel.Render(" " + row + " ")
		}
		lines = append(lines, row)
	}
	footer := m.s.panelTitle.Render(fmt.Sprintf("selected: ~%s", humanSize(selTotal)))
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", footer))
}

// --- image layer view (U12 / FR-L1) ---

// layersMsg carries an image's formatted history for the inspect-style viewer.
type layersMsg struct {
	title string
	text  string
	err   error
}

func imageLayersCmd(c engine.Engine, id, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		layers, err := c.ImageHistory(ctx, id)
		if err != nil {
			return layersMsg{err: err}
		}
		return layersMsg{title: name, text: formatLayers(layers)}
	}
}

// formatLayers renders the layer table shown in the layer view. Pure helper so
// it is unit-testable.
func formatLayers(layers []engine.Layer) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-10s  %s\n", "SIZE", "CREATED BY")
	for _, l := range layers {
		cmd := strings.TrimSpace(l.CreatedBy)
		cmd = strings.TrimPrefix(cmd, "/bin/sh -c #(nop) ")
		cmd = strings.TrimPrefix(cmd, "/bin/sh -c ")
		cmd = strings.ReplaceAll(cmd, "\t", " ")
		fmt.Fprintf(&b, "%-10s  %s\n", humanSize(l.Size), cmd)
	}
	return b.String()
}
