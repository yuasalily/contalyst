package ui

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/dockerx"
	"github.com/yuasalily/contalyst/internal/engine"
)

// resourceKind enumerates the resource lists the app can show.
type resourceKind int

const (
	kindContainers resourceKind = iota
	kindImages
	kindVolumes
	kindNetworks
	kindCompose
)

func (k resourceKind) String() string {
	switch k {
	case kindImages:
		return "Images"
	case kindVolumes:
		return "Volumes"
	case kindNetworks:
		return "Networks"
	case kindCompose:
		return "Compose"
	default:
		return "Containers"
	}
}

// --- messages ---

type errMsg struct{ err error }

type containersMsg []engine.Container
type imagesMsg []engine.Image
type volumesMsg []engine.Volume
type networksMsg []engine.Network

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
type logStartedMsg struct{ ch <-chan engine.LogLine }
type logLineMsg engine.LogLine
type logClosedMsg struct{}

type statsStartedMsg struct{ ch <-chan engine.Stats }
type statsMsg engine.Stats
type statsClosedMsg struct{}

type toastClearMsg struct{}

// composeAvailMsg reports whether `docker compose` is usable (U9 / FR-CMP7).
type composeAvailMsg struct{ ok bool }

// Multi-host / context switching (U11).
type contextsMsg []dockerx.DockerContext
type reconnectedMsg struct {
	ver string
	err error
}

// --- commands ---

// composeAvailCmd probes whether the engine's compose support is usable, once
// at startup (and after a context switch).
func composeAvailCmd(c engine.Engine) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return composeAvailMsg{ok: c.ComposeAvailable(ctx)}
	}
}

// contextsCmd lists the available Docker contexts for the switcher (U11).
func contextsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ctxs, err := dockerx.Contexts(ctx)
		if err != nil {
			return errMsg{err}
		}
		return contextsMsg(ctxs)
	}
}

// reconnectCmd pings a freshly built client after a context switch and fetches
// its version, without restarting the periodic refresh tick.
func reconnectCmd(c engine.Engine) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := c.Ping(ctx); err != nil {
			return reconnectedMsg{err: err}
		}
		return reconnectedMsg{ver: c.ServerVersion(ctx)}
	}
}

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

// bulkAction applies fn to every id concurrently and reports an aggregated
// result (U10 / FR-B5). Partial failure is tolerated: the toast summarises how
// many succeeded and failed rather than aborting on the first error.
func bulkAction(verb string, ids []string, fn func(ctx context.Context, id string) error) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		errs := make([]error, len(ids))
		for i, id := range ids {
			wg.Add(1)
			go func(i int, id string) {
				defer wg.Done()
				errs[i] = fn(ctx, id)
			}(i, id)
		}
		wg.Wait()
		failed := 0
		var firstErr error
		for _, e := range errs {
			if e != nil {
				failed++
				if firstErr == nil {
					firstErr = e
				}
			}
		}
		ok := len(ids) - failed
		if failed > 0 {
			return actionDoneMsg{err: fmt.Errorf("%s: %d ok, %d failed — %v", verb, ok, failed, firstErr)}
		}
		return actionDoneMsg{ok: fmt.Sprintf("%s: %d ok", verb, ok)}
	}
}

func inspectCmd(c engine.Engine, id, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		text, err := c.Inspect(ctx, id)
		return inspectMsg{title: name, text: text, err: err}
	}
}

// log stream: started by startLogCmd (using a stored cancelable ctx), then each
// line is awaited by waitLogCmd which re-issues itself in Update.
func startLogCmd(ctx context.Context, c engine.Engine, id string, ts bool) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.LogStream(ctx, id, ts)
		if err != nil {
			return errMsg{err}
		}
		return logStartedMsg{ch: ch}
	}
}

func waitLogCmd(ch <-chan engine.LogLine) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logClosedMsg{}
		}
		return logLineMsg(line)
	}
}

func startStatsCmd(ctx context.Context, c engine.Engine, id string) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.StatsStream(ctx, id)
		if err != nil {
			return errMsg{err}
		}
		return statsStartedMsg{ch: ch}
	}
}

func waitStatsCmd(ch <-chan engine.Stats) tea.Cmd {
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
