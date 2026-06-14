//go:build e2e

// End-to-end test of the full TUI against a real Docker daemon. It runs the
// actual ui.model program (real Docker client, real commands) through teatest's
// simulated terminal, drives it with key input, and asserts on rendered frames.
// Guarded by the `e2e` build tag. Run: go test -tags e2e -run TestE2E ./internal/ui/
package ui

import (
	"bytes"
	"os/exec"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

func TestE2E_TUIShowsContainerAndOpensLogs(t *testing.T) {
	client, err := dockerx.NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	const name = "contalyst-e2e-tui"
	_ = exec.Command("docker", "rm", "-f", name).Run()
	out, err := exec.Command("docker", "run", "-d", "--name", name, "busybox:latest",
		"sh", "-c", "i=0; while true; do echo hello-$i; i=$((i+1)); sleep 1; done").CombinedOutput()
	if err != nil {
		t.Fatalf("docker run: %v\n%s", err, out)
	}
	defer exec.Command("docker", "rm", "-f", name).Run()

	tm := teatest.NewTestModel(t, New(client), teatest.WithInitialTermSize(120, 40))

	// The list should show our container.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte(name))
	}, teatest.WithDuration(15*time.Second), teatest.WithCheckInterval(200*time.Millisecond))

	// Filter down to our container so it is the selected row, then open logs.
	tm.Type("/")
	tm.Type(name)
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // apply filter
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter}) // open detail (logs + stats)

	// Detail view should render the Logs panel and stream a log line.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("Logs")) && bytes.Contains(b, []byte("hello-"))
	}, teatest.WithDuration(15*time.Second), teatest.WithCheckInterval(200*time.Millisecond))

	// Quit cleanly.
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(10*time.Second))
}
