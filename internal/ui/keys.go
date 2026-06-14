package ui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines every key binding. Using bubbles/key keeps bindings and their
// help text in one place and feeds the hint bar / help overlay (inception
// FR-NAV5/7). Terminal-reserved chords (Ctrl-D/R/S) are deliberately avoided
// (inception R4 / NFR-U4).
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Top     key.Binding
	Bottom  key.Binding
	Enter   key.Binding
	Back    key.Binding
	Filter  key.Binding
	Command key.Binding
	Help    key.Binding
	Theme   key.Binding
	Refresh key.Binding
	Quit    key.Binding

	// Layout / decoration toggles.
	CompactHints key.Binding
	Frame        key.Binding

	// Container actions.
	StartStop key.Binding
	Restart   key.Binding
	Pause     key.Binding
	Exec      key.Binding
	Logs      key.Binding
	Inspect   key.Binding
	Delete    key.Binding
	Kill      key.Binding

	// Detail view.
	Follow     key.Binding
	Timestamps key.Binding
	Search     key.Binding
	SearchNext key.Binding
	SearchPrev key.Binding
}

func defaultKeys() keyMap {
	return keyMap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Top:     key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "top")),
		Bottom:  key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "bottom")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("⏎", "open")),
		Back:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Filter:  key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Command: key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		Help:    key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Theme:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "theme")),
		Refresh: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh")),
		Quit:    key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),

		CompactHints: key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "compact hints")),
		Frame:        key.NewBinding(key.WithKeys("F"), key.WithHelp("F", "frame style")),

		StartStop: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start/stop")),
		Restart:   key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restart")),
		Pause:     key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pause")),
		Exec:      key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "exec")),
		Logs:      key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "logs")),
		Inspect:   key.NewBinding(key.WithKeys("i"), key.WithHelp("i", "inspect")),
		Delete:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete")),
		Kill:      key.NewBinding(key.WithKeys("K"), key.WithHelp("K", "kill")),

		Follow:     key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow")),
		Timestamps: key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "timestamps")),
		Search:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		SearchNext: key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
		SearchPrev: key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev match")),
	}
}
