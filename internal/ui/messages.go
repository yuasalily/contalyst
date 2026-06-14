package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

// resourceKind enumerates the resource lists the app can show.
type resourceKind int

const (
	kindContainers resourceKind = iota
	kindImages
	kindVolumes
	kindNetworks
)

func (k resourceKind) String() string {
	switch k {
	case kindImages:
		return "Images"
	case kindVolumes:
		return "Volumes"
	case kindNetworks:
		return "Networks"
	default:
		return "Containers"
	}
}

// --- messages ---

type errMsg struct{ err error }

type containersMsg []dockerx.Container
type imagesMsg []dockerx.Image
type volumesMsg []dockerx.Volume
type networksMsg []dockerx.Network

type tickMsg time.Time

// actionDoneMsg is the result of a one-shot action (start/stop/remove/…).
type actionDoneMsg struct {
	ok  string // success toast text
	err error
}

type inspectMsg struct {
	title string
	text  string
	err   error
}

// streaming
type logStartedMsg struct{ ch <-chan dockerx.LogLine }
type logLineMsg dockerx.LogLine
type logClosedMsg struct{}

type statsStartedMsg struct{ ch <-chan dockerx.Stats }
type statsMsg dockerx.Stats
type statsClosedMsg struct{}

type toastClearMsg struct{}

// --- commands ---

func tickCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// loadCmd fetches the currently selected resource kind.
func (m *model) loadCmd() tea.Cmd {
	c := m.client
	kind := m.kind
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		switch kind {
		case kindImages:
			v, err := c.Images(ctx)
			if err != nil {
				return errMsg{err}
			}
			return imagesMsg(v)
		case kindVolumes:
			v, err := c.Volumes(ctx)
			if err != nil {
				return errMsg{err}
			}
			return volumesMsg(v)
		case kindNetworks:
			v, err := c.Networks(ctx)
			if err != nil {
				return errMsg{err}
			}
			return networksMsg(v)
		default:
			v, err := c.Containers(ctx)
			if err != nil {
				return errMsg{err}
			}
			return containersMsg(v)
		}
	}
}

// action wraps a daemon call into an actionDoneMsg with a success toast.
func action(ok string, fn func(ctx context.Context) error) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := fn(ctx); err != nil {
			return actionDoneMsg{err: err}
		}
		return actionDoneMsg{ok: ok}
	}
}

func inspectCmd(c *dockerx.Client, id, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		text, err := c.Inspect(ctx, id)
		return inspectMsg{title: name, text: text, err: err}
	}
}

// log stream: started by startLogCmd (using a stored cancelable ctx), then each
// line is awaited by waitLogCmd which re-issues itself in Update.
func startLogCmd(ctx context.Context, c *dockerx.Client, id string, ts bool) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.LogStream(ctx, id, ts)
		if err != nil {
			return errMsg{err}
		}
		return logStartedMsg{ch: ch}
	}
}

func waitLogCmd(ch <-chan dockerx.LogLine) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logClosedMsg{}
		}
		return logLineMsg(line)
	}
}

func startStatsCmd(ctx context.Context, c *dockerx.Client, id string) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.StatsStream(ctx, id)
		if err != nil {
			return errMsg{err}
		}
		return statsStartedMsg{ch: ch}
	}
}

func waitStatsCmd(ch <-chan dockerx.Stats) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return statsClosedMsg{}
		}
		return statsMsg(s)
	}
}

func toastClearCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg { return toastClearMsg{} })
}
