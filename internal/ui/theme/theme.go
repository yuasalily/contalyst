// Package theme defines Contalyst's color palettes and the semantic mapping from
// container state to color/glyph. Colors are stored as truecolor hex; termenv
// degrades them to 256/16 colors automatically on poorer terminals (inception
// NFR-U3). State is always conveyed by a text glyph as well as color so the UI
// remains legible without color (NFR-U2 / accessibility).
package theme

import "github.com/charmbracelet/lipgloss"

// Theme is a complete color palette.
type Theme struct {
	Name string

	Fg     lipgloss.Color // primary text
	Muted  lipgloss.Color // de-emphasized text
	Subtle lipgloss.Color // borders / faint text

	Accent  lipgloss.Color // primary accent (focus, titles)
	Accent2 lipgloss.Color // secondary accent

	// State semantics.
	Running    lipgloss.Color
	Exited     lipgloss.Color
	Paused     lipgloss.Color
	Transition lipgloss.Color // created / restarting / removing
	Danger     lipgloss.Color

	// Chrome.
	Border      lipgloss.Color
	BorderFocus lipgloss.Color
	SelBg       lipgloss.Color
	SelFg       lipgloss.Color
}

// StateColor returns the semantic color for a container state string.
func (t Theme) StateColor(state string) lipgloss.Color {
	switch state {
	case "running":
		return t.Running
	case "paused":
		return t.Paused
	case "restarting", "created":
		return t.Transition
	case "exited", "dead", "removing":
		return t.Exited
	default:
		return t.Muted
	}
}

// StateGlyph returns a small status glyph for a container state. ASCII-safe
// fallbacks are used so it renders on any terminal.
func StateGlyph(state string) string {
	switch state {
	case "running":
		return "●"
	case "paused":
		return "‖"
	case "restarting", "created", "removing":
		return "◐"
	case "exited", "dead":
		return "○"
	default:
		return "•"
	}
}

// Catalyst is the default theme (Catppuccin Mocha) — colorful on a dark
// background, the look the inception brief calls for.
var Catalyst = Theme{
	Name:        "Catalyst",
	Fg:          "#cdd6f4",
	Muted:       "#6c7086",
	Subtle:      "#45475a",
	Accent:      "#cba6f7",
	Accent2:     "#89dceb",
	Running:     "#a6e3a1",
	Exited:      "#6c7086",
	Paused:      "#f9e2af",
	Transition:  "#89b4fa",
	Danger:      "#f38ba8",
	Border:      "#45475a",
	BorderFocus: "#cba6f7",
	SelBg:       "#313244",
	SelFg:       "#f5e0dc",
}

// Aurora is a Tokyo Night-inspired alternative.
var Aurora = Theme{
	Name:        "Aurora",
	Fg:          "#c0caf5",
	Muted:       "#565f89",
	Subtle:      "#3b4261",
	Accent:      "#bb9af7",
	Accent2:     "#7dcfff",
	Running:     "#9ece6a",
	Exited:      "#565f89",
	Paused:      "#e0af68",
	Transition:  "#7aa2f7",
	Danger:      "#f7768e",
	Border:      "#3b4261",
	BorderFocus: "#bb9af7",
	SelBg:       "#283457",
	SelFg:       "#c0caf5",
}

// Mono is a high-contrast, mostly-grayscale theme with a single accent, for
// terminals or users where heavy color is unwanted.
var Mono = Theme{
	Name:        "Mono",
	Fg:          "#e4e4e4",
	Muted:       "#8a8a8a",
	Subtle:      "#444444",
	Accent:      "#ffffff",
	Accent2:     "#bcbcbc",
	Running:     "#d7ffd7",
	Exited:      "#8a8a8a",
	Paused:      "#dadada",
	Transition:  "#bcbcbc",
	Danger:      "#ff8787",
	Border:      "#444444",
	BorderFocus: "#ffffff",
	SelBg:       "#303030",
	SelFg:       "#ffffff",
}

var themes = []Theme{Catalyst, Aurora, Mono}

// Default returns the default theme.
func Default() Theme { return Catalyst }

// Next returns the theme after the named one, cycling around — used for live
// theme switching (inception US-11).
func Next(name string) Theme {
	for i, t := range themes {
		if t.Name == name {
			return themes[(i+1)%len(themes)]
		}
	}
	return Default()
}
