package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/dockerx"
	"github.com/yuasalily/contalyst/internal/ui/theme"
)

// rebuildList regenerates the list's columns and rows from the cached data for
// the active resource kind, applying the current fuzzy filter.
func (m *model) rebuildList() {
	// Reset per-kind list decorations; the kind builder re-enables what it needs.
	m.lst.marked = nil
	m.lst.emptyMsg = ""
	switch m.kind {
	case kindImages:
		m.buildImages()
	case kindVolumes:
		m.buildVolumes()
	case kindNetworks:
		m.buildNetworks()
	case kindCompose:
		m.buildCompose()
	default:
		m.buildContainers()
	}
}

func (m *model) buildContainers() {
	m.lst.marked = m.marked // enables the bulk-select gutter (U10)
	m.lst.cols = []column{
		{title: "", width: 1},
		{title: "NAME", min: 14},
		{title: "IMAGE", min: 16},
		{title: "STATE", width: 10},
		{title: "STATUS", width: 16},
		{title: "PORTS", min: 8},
	}
	accent2 := m.th.Accent2
	var rows []listRow
	for _, c := range m.containers {
		if m.composeScope != "" && c.Project != m.composeScope {
			continue
		}
		if !m.matches(c.Name, c.Image, c.State) {
			continue
		}
		col := m.th.StateColor(c.State)
		rows = append(rows, listRow{
			id:   c.ID,
			name: c.Name,
			cells: []cell{
				{text: theme.StateGlyph(c.State), color: col},
				{text: c.Name},
				{text: c.Image, faint: true},
				{text: c.State, color: col},
				{text: c.Status, faint: true},
				{text: c.Ports, color: accent2},
			},
		})
	}
	m.lst.setRows(rows)
}

// buildCompose lists docker-compose projects grouped from the container cache
// (U9 / FR-CMP1). The aggregate state (up/degraded/down) is color-coded.
func (m *model) buildCompose() {
	m.composeProjects = dockerx.ComposeProjects(m.containers)
	if !m.composeAvail {
		m.lst.emptyMsg = "no compose projects (or `docker compose` is unavailable)"
	}
	m.lst.cols = []column{
		{title: "", width: 1},
		{title: "PROJECT", min: 18},
		{title: "SERVICES", width: 9},
		{title: "RUNNING", width: 9},
		{title: "STATE", width: 10},
	}
	var rows []listRow
	for _, p := range m.composeProjects {
		if !m.matches(p.Name, p.State) {
			continue
		}
		col := composeStateColor(m.th, p.State)
		rows = append(rows, listRow{
			id:   p.Name,
			name: p.Name,
			cells: []cell{
				{text: composeStateGlyph(p.State), color: col},
				{text: p.Name},
				{text: fmt.Sprintf("%d", p.Services)},
				{text: fmt.Sprintf("%d/%d", p.Running, p.Containers), faint: true},
				{text: p.State, color: col},
			},
		})
	}
	m.lst.setRows(rows)
}

func composeStateColor(th theme.Theme, state string) lipgloss.Color {
	switch state {
	case "up":
		return th.Running
	case "degraded":
		return th.Paused
	default:
		return th.Exited
	}
}

func composeStateGlyph(state string) string {
	switch state {
	case "up":
		return "●"
	case "degraded":
		return "◐"
	default:
		return "○"
	}
}

func (m *model) buildImages() {
	m.lst.cols = []column{
		{title: "ID", width: 12},
		{title: "REPOSITORY", min: 18},
		{title: "TAG", min: 8},
		{title: "SIZE", width: 10},
		{title: "CREATED", width: 14},
	}
	var rows []listRow
	for _, im := range m.images {
		if !m.matches(im.Repo, im.Tag, im.ID) {
			continue
		}
		rows = append(rows, listRow{
			id:   im.ID,
			name: im.Repo + ":" + im.Tag,
			cells: []cell{
				{text: im.ID, faint: true},
				{text: im.Repo},
				{text: im.Tag, color: m.th.Accent2},
				{text: humanSize(im.Size)},
				{text: relativeTime(im.Created), faint: true},
			},
		})
	}
	m.lst.setRows(rows)
}

func (m *model) buildVolumes() {
	m.lst.cols = []column{
		{title: "NAME", min: 20},
		{title: "DRIVER", width: 10},
		{title: "MOUNTPOINT", min: 24},
	}
	var rows []listRow
	for _, v := range m.volumes {
		if !m.matches(v.Name, v.Driver, v.Mountpoint) {
			continue
		}
		rows = append(rows, listRow{
			id:   v.Name,
			name: v.Name,
			cells: []cell{
				{text: v.Name},
				{text: v.Driver, color: m.th.Accent2},
				{text: v.Mountpoint, faint: true},
			},
		})
	}
	m.lst.setRows(rows)
}

func (m *model) buildNetworks() {
	m.lst.cols = []column{
		{title: "ID", width: 12},
		{title: "NAME", min: 18},
		{title: "DRIVER", width: 12},
		{title: "SCOPE", width: 8},
	}
	var rows []listRow
	for _, n := range m.networks {
		if !m.matches(n.Name, n.Driver, n.Scope) {
			continue
		}
		rows = append(rows, listRow{
			id:   n.ID,
			name: n.Name,
			cells: []cell{
				{text: n.ID, faint: true},
				{text: n.Name},
				{text: n.Driver, color: m.th.Accent2},
				{text: n.Scope, faint: true},
			},
		})
	}
	m.lst.setRows(rows)
}

// matches reports whether any field fuzzy-matches the active filter.
func (m *model) matches(fields ...string) bool {
	if m.filter == "" {
		return true
	}
	pat := strings.ToLower(m.filter)
	for _, f := range fields {
		if fuzzyMatch(pat, strings.ToLower(f)) {
			return true
		}
	}
	return false
}

// fuzzyMatch reports whether pat is a subsequence of s (both lowercased).
func fuzzyMatch(pat, s string) bool {
	if pat == "" {
		return true
	}
	pi := 0
	for i := 0; i < len(s) && pi < len(pat); i++ {
		if s[i] == pat[pi] {
			pi++
		}
	}
	return pi == len(pat)
}

func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
