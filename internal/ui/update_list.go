package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	switch {
	case key.Matches(msg, k.Quit):
		return m, tea.Quit
	case key.Matches(msg, k.Help):
		m.overlay = ovHelp
		return m, nil
	case key.Matches(msg, k.Filter):
		m.overlay = ovFilter
		m.filterInput.SetValue(m.filter)
		m.filterInput.CursorEnd()
		m.filterInput.Focus()
		return m, nil
	case key.Matches(msg, k.Command):
		m.overlay = ovCommand
		m.cmdInput.SetValue("")
		m.cmdInput.Focus()
		return m, nil
	case key.Matches(msg, k.Theme):
		return m, m.cycleTheme()
	case key.Matches(msg, k.CompactHints):
		m.compactHints = !m.compactHints
		m.recomputeLayout()
		return m, nil
	case key.Matches(msg, k.Frame):
		m.rounded = !m.rounded
		m.applyTheme(m.th)
		mode := "square"
		if m.rounded {
			mode = "rounded"
		}
		return m, m.setToast("frames: "+mode, false)
	case key.Matches(msg, k.Refresh):
		return m, m.loadCmd()
	case key.Matches(msg, k.OpLog):
		m.overlay = ovOpLog
		return m, nil
	case key.Matches(msg, k.Back):
		// Esc clears bulk marks first, then leaves a compose project scope.
		if len(m.marked) > 0 {
			m.marked = map[string]bool{}
			return m, nil
		}
		if m.composeScope != "" {
			m.composeScope = ""
			return m, m.setKind(kindCompose)
		}
		return m, nil
	case m.kind == kindContainers && key.Matches(msg, k.Mark):
		if row, ok := m.lst.selected(); ok {
			if m.marked[row.id] {
				delete(m.marked, row.id)
			} else {
				m.marked[row.id] = true
			}
			m.lst.moveDown()
		}
		return m, nil
	case m.kind == kindContainers && key.Matches(msg, k.MarkAll):
		m.toggleMarkAll()
		return m, nil
	case key.Matches(msg, k.Up):
		m.lst.moveUp()
		return m, nil
	case key.Matches(msg, k.Down):
		m.lst.moveDown()
		return m, nil
	case key.Matches(msg, k.Top):
		m.lst.top()
		return m, nil
	case key.Matches(msg, k.Bottom):
		m.lst.bottom()
		return m, nil
	}

	// Resource-specific actions.
	switch m.kind {
	case kindContainers:
		return m.updateContainerActions(msg)
	case kindCompose:
		return m.updateComposeActions(msg)
	default:
		return m.updateResourceActions(msg)
	}
}

// updateComposeActions handles project-level compose operations (U9). Up/build
// run in the background; the destructive `down` is confirmed first.
func (m model) updateComposeActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	p, ok := m.currentComposeProject()
	if !ok {
		return m, nil
	}
	if !m.composeAvail {
		// Read-only without the compose plugin (R9 / FR-CMP7); Enter still works.
		if key.Matches(msg, k.Enter) {
			return m.enterComposeProject(p.Name)
		}
		if key.Matches(msg, k.ComposeUp, k.ComposeDown, k.Restart, k.ComposeBuild, k.ComposeBuildNC) {
			return m, m.setToast("docker compose is not available", true)
		}
		return m, nil
	}
	cl := m.client
	switch {
	case key.Matches(msg, k.Enter):
		return m.enterComposeProject(p.Name)
	case key.Matches(msg, k.ComposeUp):
		return m, action("compose up "+p.Name, func(ctx context.Context) error { return cl.Compose(ctx, p, dockerx.ComposeUp) })
	case key.Matches(msg, k.Restart):
		return m, action("compose restart "+p.Name, func(ctx context.Context) error { return cl.Compose(ctx, p, dockerx.ComposeRestart) })
	case key.Matches(msg, k.ComposeBuild):
		return m, action("compose build "+p.Name, func(ctx context.Context) error { return cl.Compose(ctx, p, dockerx.ComposeBuild) })
	case key.Matches(msg, k.ComposeBuildNC):
		return m, action("compose build --no-cache "+p.Name, func(ctx context.Context) error { return cl.Compose(ctx, p, dockerx.ComposeBuildNoCache) })
	case key.Matches(msg, k.ComposeDown):
		proj := p
		m.openConfirm("Compose down", proj.Name+" (stop & remove)", true,
			action("compose down "+proj.Name, func(ctx context.Context) error { return cl.Compose(ctx, proj, dockerx.ComposeDown) }))
		return m, nil
	}
	return m, nil
}

// enterComposeProject scopes the container list to one project's services.
func (m model) enterComposeProject(name string) (tea.Model, tea.Cmd) {
	m.composeScope = name
	return m, m.setKind(kindContainers)
}

func (m *model) currentComposeProject() (dockerx.ComposeProject, bool) {
	row, ok := m.lst.selected()
	if !ok {
		return dockerx.ComposeProject{}, false
	}
	for _, p := range m.composeProjects {
		if p.Name == row.id {
			return p, true
		}
	}
	return dockerx.ComposeProject{}, false
}

func (m model) updateContainerActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	if len(m.marked) > 0 {
		if nm, cmd, handled := m.updateBulkActions(msg); handled {
			return nm, cmd
		}
	}
	c, ok := m.currentContainer()
	if !ok {
		return m, nil
	}
	cl := m.client
	switch {
	case key.Matches(msg, k.Enter, k.Logs):
		return m.enterDetail(c.ID, c.Name)
	case key.Matches(msg, k.Inspect):
		return m, inspectCmd(cl, c.ID, c.Name)
	case key.Matches(msg, k.StartStop):
		if isUp(c.State) {
			return m, action("stopped "+c.Name, func(ctx context.Context) error { return cl.Stop(ctx, c.ID) })
		}
		return m, action("started "+c.Name, func(ctx context.Context) error { return cl.Start(ctx, c.ID) })
	case key.Matches(msg, k.Restart):
		return m, action("restarted "+c.Name, func(ctx context.Context) error { return cl.Restart(ctx, c.ID) })
	case key.Matches(msg, k.Pause):
		if c.State == "paused" {
			return m, action("unpaused "+c.Name, func(ctx context.Context) error { return cl.Unpause(ctx, c.ID) })
		}
		return m, action("paused "+c.Name, func(ctx context.Context) error { return cl.Pause(ctx, c.ID) })
	case key.Matches(msg, k.Exec):
		if !isUp(c.State) {
			return m, m.setToast("can't exec: container is not running", true)
		}
		return m, execContainer(c.ID)
	case key.Matches(msg, k.Delete):
		id, name := c.ID, c.Name
		m.openConfirm("Remove container", name, true,
			action("removed "+name, func(ctx context.Context) error { return cl.Remove(ctx, id, true) }))
		return m, nil
	case key.Matches(msg, k.Kill):
		id, name := c.ID, c.Name
		m.openConfirm("Kill container", name, true,
			action("killed "+name, func(ctx context.Context) error { return cl.Kill(ctx, id) }))
		return m, nil
	}
	return m, nil
}

// updateBulkActions applies a lifecycle action to the marked container set
// (U10). It returns handled=false for keys it does not own so single-row
// handling can proceed. The destructive remove is confirmed with the count.
func (m model) updateBulkActions(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	k := m.keys
	ids := m.markedIDs()
	if len(ids) == 0 {
		return m, nil, false
	}
	cl := m.client
	n := len(ids)
	switch {
	case key.Matches(msg, k.StartStop): // s = start marked
		m.marked = map[string]bool{}
		return m, bulkAction("started", ids, func(ctx context.Context, id string) error { return cl.Start(ctx, id) }), true
	case key.Matches(msg, k.BulkStop): // S = stop marked
		m.marked = map[string]bool{}
		return m, bulkAction("stopped", ids, func(ctx context.Context, id string) error { return cl.Stop(ctx, id) }), true
	case key.Matches(msg, k.Restart):
		m.marked = map[string]bool{}
		return m, bulkAction("restarted", ids, func(ctx context.Context, id string) error { return cl.Restart(ctx, id) }), true
	case key.Matches(msg, k.Delete):
		m.openConfirm("Remove containers", fmt.Sprintf("%d containers", n), true,
			bulkAction("removed", ids, func(ctx context.Context, id string) error { return cl.Remove(ctx, id, true) }))
		m.marked = map[string]bool{}
		return m, nil, true
	}
	return m, nil, false
}

// markedIDs returns marked container ids that still exist, in list order.
func (m *model) markedIDs() []string {
	var ids []string
	for _, c := range m.containers {
		if m.marked[c.ID] {
			ids = append(ids, c.ID)
		}
	}
	return ids
}

// toggleMarkAll marks every visible row, or clears all if everything visible is
// already marked.
func (m *model) toggleMarkAll() {
	allMarked := len(m.lst.rows) > 0
	for _, r := range m.lst.rows {
		if !m.marked[r.id] {
			allMarked = false
			break
		}
	}
	for _, r := range m.lst.rows {
		if allMarked {
			delete(m.marked, r.id)
		} else {
			m.marked[r.id] = true
		}
	}
}

func (m model) updateResourceActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	row, ok := m.lst.selected()
	if !ok {
		return m, nil
	}
	cl := m.client
	id, name := row.id, row.name
	// Images drill down to their layer history (U12 / FR-L1).
	if m.kind == kindImages && key.Matches(msg, k.Enter, k.Logs) {
		return m, imageLayersCmd(cl, id, name)
	}
	if key.Matches(msg, k.Delete) {
		var del tea.Cmd
		switch m.kind {
		case kindImages:
			del = action("removed image "+name, func(ctx context.Context) error { return cl.RemoveImage(ctx, id, false) })
		case kindVolumes:
			del = action("removed volume "+name, func(ctx context.Context) error { return cl.RemoveVolume(ctx, id, false) })
		case kindNetworks:
			del = action("removed network "+name, func(ctx context.Context) error { return cl.RemoveNetwork(ctx, id) })
		}
		if del != nil {
			m.openConfirm("Remove "+m.kind.String(), name, true, del)
		}
	}
	return m, nil
}

// currentContainer returns the selected container by matching the list row id.
func (m *model) currentContainer() (dockerx.Container, bool) {
	row, ok := m.lst.selected()
	if !ok {
		return dockerx.Container{}, false
	}
	for _, c := range m.containers {
		if c.ID == row.id {
			return c, true
		}
	}
	return dockerx.Container{}, false
}

func isUp(state string) bool {
	return state == "running" || state == "restarting" || state == "paused"
}

// setKind switches the active resource list and reloads it.
func (m *model) setKind(k resourceKind) tea.Cmd {
	m.kind = k
	m.filter = ""
	m.marked = map[string]bool{}
	m.lst.cursor = 0
	m.lst.offset = 0
	m.rebuildList()
	return m.loadCmd()
}
