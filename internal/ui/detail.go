package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

const maxLogLines = 5000

type detailState struct {
	id, name string

	logs       viewport.Model
	lines      []string
	follow     bool
	timestamps bool
	logCancel  context.CancelFunc
	logCh      <-chan dockerx.LogLine

	stats       dockerx.Stats
	haveStats   bool
	statsCancel context.CancelFunc
	statsCh     <-chan dockerx.Stats
	statsWidth  int
}

func (m model) enterDetail(id, name string) (tea.Model, tea.Cmd) {
	m.teardownStreams()
	m.state = stateDetail
	m.detail.id = id
	m.detail.name = name
	m.detail.lines = nil
	m.detail.follow = true
	m.detail.haveStats = false
	m.detail.logs.SetContent("")
	m.detail.logs.GotoTop()
	m.recomputeLayout()

	logCtx, logCancel := context.WithCancel(context.Background())
	statsCtx, statsCancel := context.WithCancel(context.Background())
	m.detail.logCancel = logCancel
	m.detail.statsCancel = statsCancel

	return m, tea.Batch(
		startLogCmd(logCtx, m.client, id, m.detail.timestamps),
		startStatsCmd(statsCtx, m.client, id),
	)
}

// teardownStreams cancels any active log/stats streams so their goroutines exit
// (inception NFR-R1: no goroutine leaks).
func (m *model) teardownStreams() {
	if m.detail.logCancel != nil {
		m.detail.logCancel()
		m.detail.logCancel = nil
	}
	if m.detail.statsCancel != nil {
		m.detail.statsCancel()
		m.detail.statsCancel = nil
	}
	m.detail.logCh = nil
	m.detail.statsCh = nil
}

func (m model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Quit):
		m.teardownStreams()
		return m, tea.Quit
	case key.Matches(msg, k.Back):
		m.teardownStreams()
		m.state = stateList
		m.recomputeLayout()
		return m, m.loadCmd()
	case key.Matches(msg, k.Help):
		m.overlay = ovHelp
		return m, nil
	case key.Matches(msg, k.Follow):
		m.detail.follow = !m.detail.follow
		if m.detail.follow {
			m.detail.logs.GotoBottom()
		}
		return m, nil
	case key.Matches(msg, k.Timestamps):
		m.detail.timestamps = !m.detail.timestamps
		if m.detail.logCancel != nil {
			m.detail.logCancel()
		}
		m.detail.lines = nil
		m.detail.logs.SetContent("")
		ctx, cancel := context.WithCancel(context.Background())
		m.detail.logCancel = cancel
		return m, startLogCmd(ctx, m.client, m.detail.id, m.detail.timestamps)
	}

	// Scrolling is delegated to the viewport; following pauses while scrolled up.
	var cmd tea.Cmd
	m.detail.logs, cmd = m.detail.logs.Update(msg)
	m.detail.follow = m.detail.logs.AtBottom()
	return m, cmd
}

func (m model) handleLogLine(line dockerx.LogLine) (tea.Model, tea.Cmd) {
	if line.Err != nil {
		return m, nil // stream ended; keep what we have
	}
	if m.state == stateDetail {
		m.detail.lines = append(m.detail.lines, line.Text)
		if len(m.detail.lines) > maxLogLines {
			m.detail.lines = m.detail.lines[len(m.detail.lines)-maxLogLines:]
		}
		m.detail.logs.SetContent(strings.Join(m.detail.lines, "\n"))
		if m.detail.follow {
			m.detail.logs.GotoBottom()
		}
	}
	if m.detail.logCh != nil {
		return m, waitLogCmd(m.detail.logCh)
	}
	return m, nil
}

func (m model) detailView() string {
	d := m.detail
	follow := "paused"
	if d.follow {
		follow = "following"
	}
	logTitle := m.s.panelTitle.Render("Logs") + m.s.hintDesc.Render("  ("+follow+")")
	logPanel := m.s.panelFocus.Width(d.logs.Width).Height(d.logs.Height).Render(d.logs.View())
	logBox := lipgloss.JoinVertical(lipgloss.Left, logTitle, logPanel)

	statsPanel := m.s.panel.Width(d.statsWidth).Height(d.logs.Height).Render(m.statsContent())
	statsBox := lipgloss.JoinVertical(lipgloss.Left, m.s.panelTitle.Render("Stats"), statsPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, logBox, " ", statsBox)
}

func (m model) statsContent() string {
	if !m.detail.haveStats {
		return m.s.empty.Render("sampling…")
	}
	s := m.detail.stats
	w := m.detail.statsWidth
	label := lipgloss.NewStyle().Foreground(m.th.Muted)
	val := lipgloss.NewStyle().Foreground(m.th.Fg)

	cpuColor := m.th.Running
	if s.CPUPercent > 80 {
		cpuColor = m.th.Danger
	} else if s.CPUPercent > 50 {
		cpuColor = m.th.Paused
	}

	lines := []string{
		label.Render("CPU"),
		bar(s.CPUPercent, w, cpuColor, m.th.Subtle) + " " + val.Render(fmt.Sprintf("%.1f%%", s.CPUPercent)),
		"",
		label.Render("MEM"),
		bar(s.MemPercent, w, m.th.Accent, m.th.Subtle),
		val.Render(fmt.Sprintf("%s / %s", humanSize(int64(s.MemUsage)), humanSize(int64(s.MemLimit)))),
		"",
		label.Render("NET") + "  " + val.Render(fmt.Sprintf("↓%s ↑%s", humanSize(int64(s.NetRx)), humanSize(int64(s.NetTx)))),
		label.Render("BLK") + "  " + val.Render(fmt.Sprintf("r:%s w:%s", humanSize(int64(s.BlkRead)), humanSize(int64(s.BlkWrite)))),
		label.Render("PIDS") + " " + val.Render(fmt.Sprintf("%d", s.Pids)),
	}
	return strings.Join(lines, "\n")
}

// bar renders a horizontal percentage meter.
func bar(pct float64, width int, fill, empty lipgloss.Color) string {
	if width < 4 {
		width = 4
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(float64(width) * pct / 100.0)
	full := lipgloss.NewStyle().Foreground(fill).Render(strings.Repeat("█", filled))
	rest := lipgloss.NewStyle().Foreground(empty).Render(strings.Repeat("░", width-filled))
	return full + rest
}

func (m model) updateInspect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Quit):
		return m, tea.Quit
	case key.Matches(msg, k.Back):
		m.state = stateList
		m.recomputeLayout()
		return m, nil
	case key.Matches(msg, k.Help):
		m.overlay = ovHelp
		return m, nil
	}
	var cmd tea.Cmd
	m.inspect.vp, cmd = m.inspect.vp.Update(msg)
	return m, cmd
}

func (m model) inspectView() string {
	title := m.s.panelTitle.Render("Inspect: ") + m.s.crumb.Render(m.inspect.title)
	return lipgloss.JoinVertical(lipgloss.Left, title, m.inspect.vp.View())
}
