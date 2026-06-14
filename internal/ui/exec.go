package ui

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/engine"
)

// execContainer suspends the TUI and drops the user into an interactive shell
// inside the container (inception FR-C9). tea.ExecProcess hands the terminal
// over cleanly and restores the TUI on exit. The command to run is supplied by
// the engine (engine.ExecSpec) so the choice of CLI — docker, podman, … — stays
// in the adapter rather than hard-coded here.
func execContainer(spec engine.ExecSpec) tea.Cmd {
	c := exec.Command(spec.Name, spec.Args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execDoneMsg{err: err}
	})
}
