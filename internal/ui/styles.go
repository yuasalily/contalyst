package ui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/yuasalily/contalyst/internal/ui/theme"
)

// styles holds every lipgloss.Style the UI needs, derived from a Theme so a
// theme switch is a single rebuild (inception US-11).
type styles struct {
	th theme.Theme

	appName    lipgloss.Style
	crumb      lipgloss.Style
	crumbSep   lipgloss.Style
	statusInfo lipgloss.Style
	headerRule lipgloss.Style

	colHeader  lipgloss.Style
	rowNormal  lipgloss.Style
	rowSel     lipgloss.Style
	cellMuted  lipgloss.Style
	markGutter lipgloss.Style

	hintKey  lipgloss.Style
	hintDesc lipgloss.Style
	hintSep  lipgloss.Style

	panel      lipgloss.Style
	panelFocus lipgloss.Style
	panelTitle lipgloss.Style

	dialog       lipgloss.Style
	dialogTitle  lipgloss.Style
	btnActive    lipgloss.Style
	btnInactive  lipgloss.Style
	btnDanger    lipgloss.Style
	toastInfo    lipgloss.Style
	toastErr     lipgloss.Style
	promptPrefix lipgloss.Style
	empty        lipgloss.Style
	searchHit    lipgloss.Style
}

// newStyles derives every style from a Theme. When rounded is false the framed
// panels/dialog fall back to a plain square border, making the rounded-frame
// decoration opt-in (inception NFR-U5).
func newStyles(th theme.Theme, rounded bool) styles {
	border := lipgloss.NormalBorder()
	if rounded {
		border = lipgloss.RoundedBorder()
	}
	return styles{
		th: th,

		appName:    lipgloss.NewStyle().Bold(true).Foreground(th.Accent),
		crumb:      lipgloss.NewStyle().Foreground(th.Fg),
		crumbSep:   lipgloss.NewStyle().Foreground(th.Muted),
		statusInfo: lipgloss.NewStyle().Foreground(th.Muted),
		headerRule: lipgloss.NewStyle().Foreground(th.Subtle),

		colHeader:  lipgloss.NewStyle().Bold(true).Foreground(th.Accent2),
		rowNormal:  lipgloss.NewStyle().Foreground(th.Fg),
		rowSel:     lipgloss.NewStyle().Background(th.SelBg).Foreground(th.SelFg).Bold(true),
		cellMuted:  lipgloss.NewStyle().Foreground(th.Muted),
		markGutter: lipgloss.NewStyle().Foreground(th.Accent2).Bold(true),

		hintKey:  lipgloss.NewStyle().Bold(true).Foreground(th.Accent),
		hintDesc: lipgloss.NewStyle().Foreground(th.Muted),
		hintSep:  lipgloss.NewStyle().Foreground(th.Subtle),

		panel:      lipgloss.NewStyle().Border(border).BorderForeground(th.Border),
		panelFocus: lipgloss.NewStyle().Border(border).BorderForeground(th.BorderFocus),
		panelTitle: lipgloss.NewStyle().Bold(true).Foreground(th.Accent),

		dialog:       lipgloss.NewStyle().Border(border).BorderForeground(th.Danger).Padding(1, 3),
		dialogTitle:  lipgloss.NewStyle().Bold(true).Foreground(th.Fg),
		btnActive:    lipgloss.NewStyle().Background(th.Accent).Foreground(th.SelBg).Bold(true).Padding(0, 2),
		btnInactive:  lipgloss.NewStyle().Foreground(th.Muted).Padding(0, 2),
		btnDanger:    lipgloss.NewStyle().Background(th.Danger).Foreground(th.SelBg).Bold(true).Padding(0, 2),
		toastInfo:    lipgloss.NewStyle().Foreground(th.Running).Bold(true),
		toastErr:     lipgloss.NewStyle().Foreground(th.Danger).Bold(true),
		promptPrefix: lipgloss.NewStyle().Bold(true).Foreground(th.Accent),
		empty:        lipgloss.NewStyle().Foreground(th.Muted).Italic(true),
		searchHit:    lipgloss.NewStyle().Background(th.Accent2).Foreground(th.SelBg).Bold(true),
	}
}
