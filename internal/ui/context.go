package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

// updateContext drives the host/context switcher overlay (U11 / FR-H2). Up/down
// move the cursor; enter switches to the highlighted context; esc closes.
func (m model) updateContext(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Back), key.Matches(msg, k.Quit):
		m.overlay = ovNone
		return m, nil
	case key.Matches(msg, k.Up):
		if m.contextCursor > 0 {
			m.contextCursor--
		}
		return m, nil
	case key.Matches(msg, k.Down):
		if m.contextCursor < len(m.contexts)-1 {
			m.contextCursor++
		}
		return m, nil
	case key.Matches(msg, k.Enter):
		if m.contextCursor < 0 || m.contextCursor >= len(m.contexts) {
			m.overlay = ovNone
			return m, nil
		}
		return m.switchContext(m.contexts[m.contextCursor])
	}
	return m, nil
}

// switchContext tears down the current connection's streams, rebuilds the
// client against the chosen host, and resets view state (R11 / OQ-5). The
// periodic refresh tick keeps running; only the client and data are replaced.
func (m model) switchContext(c dockerx.DockerContext) (tea.Model, tea.Cmd) {
	if c.Name == m.contextName {
		m.overlay = ovNone
		return m, nil
	}
	cl, err := dockerx.NewClientForHost(c.Host)
	if err != nil {
		m.overlay = ovNone
		return m, m.setToast("context switch failed: "+err.Error(), true)
	}
	m.teardownStreams()
	m.client = cl
	m.contextName = c.Name
	m.overlay = ovNone

	// Reset view state so nothing from the old host lingers.
	m.state = stateList
	m.kind = kindContainers
	m.composeScope = ""
	m.marked = map[string]bool{}
	m.filter = ""
	m.containers = nil
	m.images = nil
	m.volumes = nil
	m.networks = nil
	m.composeProjects = nil
	m.lst.cursor = 0
	m.lst.offset = 0
	m.rebuildList()
	m.recomputeLayout()
	return m, reconnectCmd(cl)
}

func (m model) contextBox() string {
	title := m.s.appName.Render("Switch context") + m.s.hintDesc.Render("  — ↑↓ choose · ⏎ switch · esc")
	if len(m.contexts) == 0 {
		body := m.s.empty.Render("no docker contexts found")
		return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
	}
	var lines []string
	for i, c := range m.contexts {
		marker := "  "
		if c.Name == m.contextName {
			marker = m.s.markGutter.Render("● ")
		}
		name := c.Name
		host := m.s.hintDesc.Render(truncMiddle(c.Host, 40))
		row := marker + pad(name, 14) + "  " + host
		if i == m.contextCursor {
			row = m.s.rowSel.Render(" " + marker + pad(name, 14) + "  " + truncMiddle(c.Host, 40) + " ")
		}
		lines = append(lines, row)
	}
	body := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// truncMiddle shortens s to w cells, keeping head and tail around an ellipsis.
func truncMiddle(s string, w int) string {
	if len(s) <= w || w < 5 {
		return s
	}
	half := (w - 1) / 2
	return s[:half] + "…" + s[len(s)-(w-1-half):]
}

func (m model) activeContextLabel() string {
	if m.contextName == "" {
		return ""
	}
	return strings.ToLower(m.contextName)
}
