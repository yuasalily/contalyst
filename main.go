// Command contalyst is a modern, colorful TUI for managing Docker containers.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/yuasalily/contalyst/internal/dockerx"
	"github.com/yuasalily/contalyst/internal/ui"
)

// Build metadata, injected at release time via -ldflags (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version", "version":
			fmt.Printf("contalyst %s (commit %s, built %s)\n", version, commit, date)
			return
		case "-h", "--help", "help":
			fmt.Println("contalyst — a modern, colorful Docker TUI")
			fmt.Println("\nUsage: contalyst [--version]")
			fmt.Println("\nConnects to Docker via the standard environment (DOCKER_HOST, etc.).")
			fmt.Println("Run with no arguments to launch the TUI. Press ? inside for keybindings.")
			return
		}
	}

	client, err := dockerx.NewClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, "contalyst:", err)
		os.Exit(1)
	}
	defer func() { _ = client.Close() }()

	p := tea.NewProgram(ui.New(client), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "contalyst:", err)
		os.Exit(1)
	}
}
