package ui

import (
	"context"

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
	case key.Matches(msg, k.Refresh):
		return m, m.loadCmd()
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
	if m.kind == kindContainers {
		return m.updateContainerActions(msg)
	}
	return m.updateResourceActions(msg)
}

func (m model) updateContainerActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
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

func (m model) updateResourceActions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := m.keys
	row, ok := m.lst.selected()
	if !ok {
		return m, nil
	}
	cl := m.client
	id, name := row.id, row.name
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
	m.lst.cursor = 0
	m.lst.offset = 0
	m.rebuildList()
	return m.loadCmd()
}
