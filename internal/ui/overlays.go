package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- confirmation dialog (inception R5: destructive ops always confirmed, with
// the safe option focused by default) ---

type confirmState struct {
	title, body string
	danger      bool
	onYes       tea.Cmd
	yes         bool // cursor: false = Cancel (default/safe), true = Confirm
}

func (m *model) openConfirm(title, body string, danger bool, onYes tea.Cmd) {
	m.confirm = confirmState{title: title, body: body, danger: danger, onYes: onYes, yes: false}
	m.overlay = ovConfirm
}

func (m model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "tab", "h", "l":
		m.confirm.yes = !m.confirm.yes
		return m, nil
	case "y", "Y":
		m.overlay = ovNone
		return m, m.confirm.onYes
	case "n", "N", "esc":
		m.overlay = ovNone
		return m, nil
	case "enter":
		m.overlay = ovNone
		if m.confirm.yes {
			return m, m.confirm.onYes
		}
		return m, nil
	}
	return m, nil
}

// --- filter (live fuzzy) ---

func (m model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.overlay = ovNone
		m.filter = ""
		m.filterInput.Blur()
		m.rebuildList()
		return m, nil
	case "enter":
		m.overlay = ovNone
		m.filterInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.filter = m.filterInput.Value()
	m.rebuildList()
	return m, cmd
}

// --- in-log search (inception FR-C5) ---

func (m model) updateLogSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.overlay = ovNone
		m.searchInput.Blur()
		m.detail.search = ""
		m.detail.matches = nil
		m.detail.matchIdx = 0
		m.detail.logs.SetContent(strings.Join(m.detail.lines, "\n"))
		return m, nil
	case "enter":
		m.overlay = ovNone
		m.searchInput.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.detail.search = m.searchInput.Value()
	m.detail.matchIdx = 0
	m.computeMatches()
	m.detail.logs.SetContent(m.renderLogContent())
	m.jumpToMatch()
	return m, cmd
}

// --- command palette ---

var commandNames = []string{"containers", "compose", "images", "volumes", "networks", "context", "prune", "oplog", "theme", "help", "quit"}

func (m model) updateCommand(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.overlay = ovNone
		m.cmdInput.Blur()
		return m, nil
	case "enter":
		return m.runCommand(m.cmdInput.Value())
	case "tab":
		if s := commandSuggest(m.cmdInput.Value()); s != "" {
			m.cmdInput.SetValue(s)
			m.cmdInput.CursorEnd()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.cmdInput, cmd = m.cmdInput.Update(msg)
	return m, cmd
}

func (m model) runCommand(raw string) (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(strings.ToLower(raw))
	m.overlay = ovNone
	m.cmdInput.Blur()
	switch cmd {
	case "", "esc":
		return m, nil
	case "containers", "container", "ps", "c":
		m.composeScope = ""
		return m, m.setKind(kindContainers)
	case "compose", "comp", "stacks":
		m.composeScope = ""
		return m, m.setKind(kindCompose)
	case "images", "image", "img", "i":
		return m, m.setKind(kindImages)
	case "volumes", "volume", "vol", "v":
		return m, m.setKind(kindVolumes)
	case "networks", "network", "net", "n":
		return m, m.setKind(kindNetworks)
	case "theme":
		return m, m.cycleTheme()
	case "help", "?":
		m.overlay = ovHelp
		return m, nil
	case "quit", "q", "exit":
		return m, tea.Quit
	case "context", "contexts", "hosts", "host", "ctx":
		return m, contextsCmd()
	case "oplog", "ops", "log", "history":
		m.overlay = ovOpLog
		return m, nil
	case "prune":
		return m, pruneUsageCmd(m.client)
	default:
		return m, m.setToast("unknown command: "+cmd, true)
	}
}

func commandSuggest(prefix string) string {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		return ""
	}
	for _, c := range commandNames {
		if strings.HasPrefix(c, prefix) {
			return c
		}
	}
	return ""
}

// --- header ---

func (m model) headerView() string {
	left := m.s.appName.Render("Contalyst")
	sep := m.s.crumbSep.Render(" › ")
	crumbs := m.s.crumb.Render(m.kind.String())
	if m.composeScope != "" && m.kind == kindContainers {
		crumbs = m.s.crumb.Render("Compose") + sep + m.s.crumb.Render(m.composeScope)
	}
	switch m.state {
	case stateDetail:
		crumbs += sep + m.s.crumb.Render(m.detail.name)
	case stateInspect:
		crumbs += sep + m.s.crumb.Render(m.inspect.title)
	}
	if m.filter != "" {
		crumbs += "  " + m.s.hintDesc.Render("/"+m.filter)
	}
	hostInfo := ""
	if label := m.activeContextLabel(); label != "" {
		hostInfo = label + " · "
	}
	right := m.s.statusInfo.Render(fmt.Sprintf("%sdocker %s · %s", hostInfo, m.serverVer, m.th.Name))

	line := left + sep + crumbs
	gap := m.width - lipgloss.Width(line) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
		right = ""
	}
	return line + strings.Repeat(" ", gap) + right
}

// --- bottom area: filter bar / command bar / hint bar ---

func (m model) bottomView() string {
	switch m.overlay {
	case ovFilter:
		return m.s.promptPrefix.Render("/") + " " + m.filterInput.View()
	case ovCommand:
		sug := commandSuggest(m.cmdInput.Value())
		suffix := ""
		if sug != "" && sug != strings.ToLower(strings.TrimSpace(m.cmdInput.Value())) {
			suffix = "   " + m.s.hintDesc.Render("⇥ "+sug)
		}
		return m.s.promptPrefix.Render(":") + " " + m.cmdInput.View() + suffix
	case ovLogSearch:
		count := ""
		if m.detail.search != "" {
			if n := len(m.detail.matches); n > 0 {
				count = "   " + m.s.hintDesc.Render(fmt.Sprintf("%d/%d", m.detail.matchIdx+1, n))
			} else {
				count = "   " + m.s.hintDesc.Render("no matches")
			}
		}
		return m.s.promptPrefix.Render("/") + " " + m.searchInput.View() + count
	default:
		return m.hintView()
	}
}

func (m model) hintView() string {
	var tokens [][2]string
	switch m.state {
	case stateDetail:
		tokens = [][2]string{{"↑↓", "scroll"}, {"f", "follow"}, {"t", "timestamps"}, {"/", "search"}}
		if m.detail.search != "" {
			tokens = append(tokens, [2]string{"n/N", "next/prev"})
		}
		tokens = append(tokens, [][2]string{{"esc", "back"}, {"?", "help"}, {"q", "quit"}}...)
	case stateInspect:
		tokens = [][2]string{{"↑↓", "scroll"}, {"esc", "back"}, {"?", "help"}, {"q", "quit"}}
	default:
		switch m.kind {
		case kindCompose:
			tokens = [][2]string{{"↑↓", "move"}, {"⏎", "services"}, {"u", "up"}, {"d", "down"}, {"r", "restart"}, {"b", "build"}, {"B", "no-cache"}, {":", "cmd"}, {"?", "help"}, {"q", "quit"}}
		case kindContainers:
			tokens = [][2]string{{"↑↓", "move"}, {"⏎", "logs"}, {"s", "start/stop"}, {"r", "restart"}, {"e", "exec"}, {"i", "inspect"}, {"d", "delete"}, {"space", "mark"}, {"/", "filter"}, {":", "cmd"}, {"?", "help"}, {"q", "quit"}}
			if len(m.marked) > 0 {
				tokens = [][2]string{{"↑↓", "move"}, {"space", "mark"}, {"a", "all"}, {"s", "start"}, {"S", "stop"}, {"r", "restart"}, {"d", "delete"}, {"esc", "clear"}, {":", "cmd"}, {"?", "help"}, {"q", "quit"}}
			}
		case kindImages:
			tokens = [][2]string{{"↑↓", "move"}, {"⏎", "layers"}, {"d", "delete"}, {"/", "filter"}, {":", "cmd"}, {"T", "theme"}, {"?", "help"}, {"q", "quit"}}
		default:
			tokens = [][2]string{{"↑↓", "move"}, {"d", "delete"}, {"/", "filter"}, {":", "cmd"}, {"T", "theme"}, {"?", "help"}, {"q", "quit"}}
		}
	}
	maxLines := 2
	if m.compactHints {
		maxLines = 1
	}
	return m.packHints(tokens, maxLines)
}

func (m model) packHints(tokens [][2]string, maxLines int) string {
	rendered := make([]string, len(tokens))
	for i, t := range tokens {
		rendered[i] = m.s.hintKey.Render(t[0]) + " " + m.s.hintDesc.Render(t[1])
	}
	sep := m.s.hintSep.Render("  ·  ")
	sepW := lipgloss.Width(sep)

	var lines []string
	var cur string
	curW := 0
	for _, r := range rendered {
		rw := lipgloss.Width(r)
		switch {
		case cur == "":
			cur, curW = r, rw
		case curW+sepW+rw <= m.width:
			cur += sep + r
			curW += sepW + rw
		default:
			lines = append(lines, cur)
			cur, curW = r, rw
			if len(lines) == maxLines {
				// out of room; append an ellipsis to the last line and stop
				lines[maxLines-1] += m.s.hintDesc.Render(" …")
				return strings.Join(lines, "\n")
			}
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}

func (m model) toastView() string {
	st := m.s.toastInfo
	prefix := "✓ "
	if m.toastErr {
		st = m.s.toastErr
		prefix = "✕ "
	}
	return st.Render(prefix + m.toast)
}

// --- centered overlays: help & confirm ---

func (m model) overlayView() string {
	switch m.overlay {
	case ovHelp:
		return m.helpBox()
	case ovConfirm:
		return m.confirmBox()
	case ovContext:
		return m.contextBox()
	case ovOpLog:
		return m.opLogBox()
	case ovPrune:
		return m.pruneBox()
	}
	return ""
}

func (m model) confirmBox() string {
	c := m.confirm
	var cancelBtn, confirmBtn string
	if c.yes {
		cancelBtn = m.s.btnInactive.Render("Cancel")
		confirmBtn = m.s.btnDanger.Render("Confirm")
	} else {
		cancelBtn = m.s.btnActive.Render("Cancel")
		confirmBtn = m.s.btnInactive.Render("Confirm")
	}
	btns := lipgloss.JoinHorizontal(lipgloss.Top, cancelBtn, "   ", confirmBtn)
	hint := m.s.hintDesc.Render("←/→ choose · enter · y/n · esc")
	content := lipgloss.JoinVertical(lipgloss.Center,
		m.s.dialogTitle.Render(c.title),
		"",
		m.s.crumb.Render(c.body),
		"",
		btns,
		"",
		hint,
	)
	return m.s.dialog.Render(content)
}

func (m model) helpBox() string {
	section := func(title string, pairs [][2]string) string {
		lines := []string{m.s.panelTitle.Render(title)}
		for _, p := range pairs {
			lines = append(lines, "  "+m.s.hintKey.Render(pad(p[0], 8))+m.s.hintDesc.Render(p[1]))
		}
		return strings.Join(lines, "\n")
	}
	global := section("Global", [][2]string{
		{"↑/k ↓/j", "move"}, {"g / G", "top / bottom"}, {"/", "fuzzy filter"},
		{":", "command palette"}, {"T", "cycle theme"}, {"R", "refresh"},
		{"H", "compact hints"}, {"F", "frame style"}, {"@", "operation log"},
		{"?", "this help"}, {"q", "quit"},
	})
	containers := section("Containers", [][2]string{
		{"⏎ / l", "logs + stats"}, {"i", "inspect"}, {"s", "start/stop"},
		{"r", "restart"}, {"p", "pause/unpause"}, {"e", "exec shell"},
		{"d", "remove"}, {"K", "kill"}, {"space / a", "mark / mark all"},
		{"s/S/r/d", "bulk on marked"},
	})
	detail := section("Logs / Detail", [][2]string{
		{"↑↓ pgup", "scroll"}, {"f", "toggle follow"}, {"t", "timestamps"},
		{"/", "search logs"}, {"n / N", "next / prev match"}, {"esc", "back"},
	})
	compose := section("Compose", [][2]string{
		{"⏎", "services"}, {"u", "up -d"}, {"d", "down"}, {"r", "restart"},
		{"b / B", "build / --no-cache"},
	})
	cmds := section("Commands ( : )", [][2]string{
		{"compose", "compose projects"}, {"images", "images (⏎ = layers)"},
		{"volumes / networks", "other resources"}, {"context", "switch host"},
		{"prune", "prune dashboard"}, {"oplog", "operation log"},
	})
	left := lipgloss.JoinVertical(lipgloss.Left, global, containers)
	right := lipgloss.JoinVertical(lipgloss.Left, detail, compose, cmds)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "    ", right)
	title := m.s.appName.Render("Contalyst") + m.s.hintDesc.Render("  — keybindings")
	return m.s.panel.Padding(1, 2).Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

// overlayCenter places box in the center of a w×h area (the box replaces the
// screen while active — a deliberate modal focus).
func overlayCenter(box string, w, h int) string {
	return lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, box)
}
