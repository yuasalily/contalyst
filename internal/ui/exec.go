package ui

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// execContainer suspends the TUI and drops the user into an interactive shell
// inside the container, preferring bash and falling back to sh (inception
// FR-C9). tea.ExecProcess hands the terminal over cleanly and restores the TUI
// on exit. This shells out to the docker CLI, which is virtually always present
// alongside the daemon and gives a correct interactive PTY for free.
func execContainer(id string) tea.Cmd {
	c := exec.Command("docker", "exec", "-it", id, "sh", "-c",
		"command -v bash >/dev/null 2>&1 && exec bash || exec sh")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return execDoneMsg{err: err}
	})
}
