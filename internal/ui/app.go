package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/dockerx"
	"github.com/yuasalily/contalyst/internal/ui/theme"
)

type uiState int

const (
	stateList uiState = iota
	stateDetail
	stateInspect
)

type overlay int

const (
	ovNone overlay = iota
	ovFilter
	ovCommand
	ovConfirm
	ovHelp
	ovLogSearch
	ovContext // U11: host/context switcher
	ovOpLog   // U12: operation log
	ovPrune   // U12: prune dashboard
)

type connectedMsg struct{ ver string }
type execDoneMsg struct{ err error }

type model struct {
	client *dockerx.Client
	keys   keyMap
	th     theme.Theme
	s      styles

	width, height int
	serverVer     string
	ready         bool
	fatalErr      error

	state   uiState
	overlay overlay
	kind    resourceKind
	lst     list

	containers []dockerx.Container
	images     []dockerx.Image
	volumes    []dockerx.Volume
	networks   []dockerx.Network

	// Compose (U9): projects are derived from the container list.
	composeProjects []dockerx.ComposeProject
	composeAvail    bool
	composeScope    string // when set, the container list is scoped to this project

	// Bulk multi-select (U10): set of marked container ids.
	marked map[string]bool

	// Multi-host / context switching (U11).
	contexts      []dockerx.DockerContext
	contextName   string // active context name shown in the header
	contextCursor int    // selection cursor in the context overlay

	// Maintenance (U12): operation log + prune dashboard.
	opLog       []opEntry
	pruneUsage  []dockerx.Usage
	pruneSel    []bool
	pruneCursor int

	filter      string
	filterInput textinput.Model
	cmdInput    textinput.Model
	searchInput textinput.Model
	confirm     confirmState

	detail  detailState
	inspect inspectState

	toast        string
	toastErr     bool
	compactHints bool
	rounded      bool
}

type inspectState struct {
	title string
	vp    viewport.Model
}

// New builds the root model around an established Docker client.
func New(client *dockerx.Client) model {
	th := theme.Default()

	fi := textinput.New()
	fi.Prompt = ""
	fi.Placeholder = "filter…"

	ci := textinput.New()
	ci.Prompt = ""
	ci.Placeholder = "command (try: images, volumes, networks, theme, quit)"

	si := textinput.New()
	si.Prompt = ""
	si.Placeholder = "search logs…"

	return model{
		client:      client,
		keys:        defaultKeys(),
		th:          th,
		s:           newStyles(th, true),
		kind:        kindContainers,
		filterInput: fi,
		cmdInput:    ci,
		searchInput: si,
		rounded:     true,
		marked:      map[string]bool{},
	}
}

func (m model) Init() tea.Cmd {
	return connectCmd(m.client)
}

func connectCmd(c *dockerx.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.Ping(ctx); err != nil {
			return errMsg{err}
		}
		return connectedMsg{ver: c.ServerVersion(ctx)}
	}
}

func (m *model) recomputeLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	hintH := 2
	if m.compactHints {
		hintH = 1
	}
	toastH := 0
	if m.toast != "" {
		toastH = 1
	}
	bodyH := max(m.height-headerHeight-hintH-toastH, 3)

	m.lst.width = m.width
	m.lst.height = bodyH - 1 // minus column header
	m.lst.clampOffset()

	// Detail: logs panel on the left, stats on the right. The two bordered
	// panels plus the gap consume 5 columns of chrome (2 borders each + 1 gap),
	// so the log content width is total - stats - 5.
	statsW := 30
	if m.width < 80 {
		statsW = m.width / 3
	}
	logsW := max(m.width-statsW-5, 10)
	m.detail.logs.Width = logsW
	m.detail.logs.Height = max(bodyH-2, 1) // borders
	m.detail.statsWidth = statsW

	m.inspect.vp.Width = m.width
	m.inspect.vp.Height = bodyH
}

const headerHeight = 1

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.detail.logs.Width == 0 {
			m.detail.logs = viewport.New(10, 3)
		}
		if m.inspect.vp.Width == 0 {
			m.inspect.vp = viewport.New(10, 3)
		}
		m.recomputeLayout()
		return m, nil

	case connectedMsg:
		m.ready = true
		m.serverVer = msg.ver
		return m, tea.Batch(m.loadCmd(), tickCmd(), composeAvailCmd())

	case composeAvailMsg:
		m.composeAvail = msg.ok
		return m, nil

	case contextsMsg:
		m.contexts = []dockerx.DockerContext(msg)
		m.contextCursor = 0
		for i, c := range m.contexts {
			if c.Current && m.contextName == "" {
				m.contextName = c.Name
			}
			if c.Name == m.contextName {
				m.contextCursor = i
			}
		}
		m.overlay = ovContext
		return m, nil

	case reconnectedMsg:
		if msg.err != nil {
			return m, m.setToast("context switch failed: "+msg.err.Error(), true)
		}
		m.serverVer = msg.ver
		return m, tea.Batch(m.setToast("switched to "+m.contextName, false), m.loadCmd(), composeAvailCmd())

	case errMsg:
		if !m.ready {
			m.fatalErr = msg.err
			return m, nil
		}
		return m, m.setToast(msg.err.Error(), true)

	case tickMsg:
		if m.state == stateList {
			return m, tea.Batch(m.loadCmd(), tickCmd())
		}
		return m, tickCmd()

	case containersMsg:
		m.containers = []dockerx.Container(msg)
		m.rebuildList()
		return m, nil
	case imagesMsg:
		m.images = []dockerx.Image(msg)
		m.rebuildList()
		return m, nil
	case volumesMsg:
		m.volumes = []dockerx.Volume(msg)
		m.rebuildList()
		return m, nil
	case networksMsg:
		m.networks = []dockerx.Network(msg)
		m.rebuildList()
		return m, nil

	case actionDoneMsg:
		// Every one-shot/bulk/compose/prune action funnels through here, so the
		// operation log (U12 / FR-OL1) is recorded in one place.
		if msg.err != nil {
			m.recordOp(msg.err.Error(), false)
			return m, m.setToast(msg.err.Error(), true)
		}
		m.recordOp(msg.ok, true)
		return m, tea.Batch(m.setToast(msg.ok, false), m.loadCmd())

	case execDoneMsg:
		if msg.err != nil {
			return m, m.setToast("exec failed: "+msg.err.Error(), true)
		}
		return m, m.loadCmd()

	case toastClearMsg:
		m.toast = ""
		m.recomputeLayout()
		return m, nil

	case inspectMsg:
		if msg.err != nil {
			return m, m.setToast(msg.err.Error(), true)
		}
		m.state = stateInspect
		m.inspect.title = msg.title
		m.inspect.vp.SetContent(msg.text)
		m.inspect.vp.GotoTop()
		return m, nil

	case layersMsg:
		if msg.err != nil {
			return m, m.setToast(msg.err.Error(), true)
		}
		m.state = stateInspect
		m.inspect.title = "Layers: " + msg.title
		m.inspect.vp.SetContent(msg.text)
		m.inspect.vp.GotoTop()
		return m, nil

	case pruneUsageMsg:
		if msg.err != nil {
			return m, m.setToast(msg.err.Error(), true)
		}
		m.pruneUsage = msg.usage
		m.pruneSel = make([]bool, len(msg.usage))
		m.pruneCursor = 0
		m.overlay = ovPrune
		return m, nil

	case logStartedMsg:
		m.detail.logCh = msg.ch
		return m, waitLogCmd(msg.ch)
	case logLineMsg:
		return m.handleLogLine(dockerx.LogLine(msg))
	case logClosedMsg:
		return m, nil

	case statsStartedMsg:
		m.detail.statsCh = msg.ch
		return m, waitStatsCmd(msg.ch)
	case statsMsg:
		m.detail.stats = dockerx.Stats(msg)
		m.detail.haveStats = true
		if m.detail.statsCh != nil {
			return m, waitStatsCmd(m.detail.statsCh)
		}
		return m, nil
	case statsClosedMsg:
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey routes key input by the active overlay/state.
func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.fatalErr != nil {
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}
		return m, nil
	}

	switch m.overlay {
	case ovFilter:
		return m.updateFilter(msg)
	case ovCommand:
		return m.updateCommand(msg)
	case ovConfirm:
		return m.updateConfirm(msg)
	case ovHelp:
		if key.Matches(msg, m.keys.Back, m.keys.Help, m.keys.Quit) {
			m.overlay = ovNone
		}
		return m, nil
	case ovLogSearch:
		return m.updateLogSearch(msg)
	case ovContext:
		return m.updateContext(msg)
	case ovPrune:
		return m.updatePrune(msg)
	case ovOpLog:
		if key.Matches(msg, m.keys.Back, m.keys.OpLog, m.keys.Quit) {
			m.overlay = ovNone
		}
		return m, nil
	}

	switch m.state {
	case stateDetail:
		return m.updateDetail(msg)
	case stateInspect:
		return m.updateInspect(msg)
	default:
		return m.updateList(msg)
	}
}

func (m *model) setToast(text string, isErr bool) tea.Cmd {
	m.toast = text
	m.toastErr = isErr
	m.recomputeLayout()
	return toastClearCmd()
}

// applyTheme rebuilds styles after a theme change.
func (m *model) applyTheme(t theme.Theme) {
	m.th = t
	m.s = newStyles(t, m.rounded)
	m.rebuildList()
}

// cycleTheme advances to the next theme and announces it. Shared by the `T` key
// and the `:theme` command so the behaviour stays in one place.
func (m *model) cycleTheme() tea.Cmd {
	m.applyTheme(theme.Next(m.th.Name))
	return m.setToast("theme: "+m.th.Name, false)
}

func (m model) View() string {
	if m.fatalErr != nil {
		return m.fatalView()
	}
	if !m.ready || m.width == 0 {
		return "\n  Connecting to Docker…"
	}

	var body string
	switch m.state {
	case stateDetail:
		body = m.detailView()
	case stateInspect:
		body = m.inspectView()
	default:
		body = m.lst.view(m.s)
	}

	parts := []string{m.headerView(), body}
	if m.toast != "" {
		parts = append(parts, m.toastView())
	}
	parts = append(parts, m.bottomView())

	base := lipgloss.JoinVertical(lipgloss.Left, parts...)
	if overlayView := m.overlayView(); overlayView != "" {
		return overlayCenter(overlayView, m.width, m.height)
	}
	return base
}

func (m model) fatalView() string {
	box := m.s.dialog.Render(
		m.s.toastErr.Render("Cannot connect to Docker") + "\n\n" +
			m.s.crumb.Render(m.fatalErr.Error()) + "\n\n" +
			m.s.hintDesc.Render("Press q to quit."),
	)
	return "\n" + box
}
