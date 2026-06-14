package dockerx

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yuasalily/contalyst/internal/engine"
)

// ComposeAvailable reports whether the `docker compose` v2 plugin is usable.
// Compose operations shell out to it (DR-5); when it is missing the compose
// view stays read-only and explains why (R9 / FR-CMP7).
func (c *Client) ComposeAvailable(ctx context.Context) bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return exec.CommandContext(ctx, "docker", "compose", "version").Run() == nil
}

// Compose runs a compose lifecycle operation for a project by delegating to the
// `docker compose` CLI, which honours depends_on ordering for free (FR-CMP5).
// The project's config-file and working-directory labels are passed through so
// up/build can find the compose file even though Contalyst never reads it.
func (c *Client) Compose(ctx context.Context, p engine.ComposeProject, op engine.ComposeOp) error {
	args := composeBaseArgs(p)
	args = append(args, op.Args()...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("docker compose %s: %w", strings.Join(op.Args(), " "), err)
		}
		return fmt.Errorf("docker compose %s: %s", strings.Join(op.Args(), " "), lastLine(msg))
	}
	return nil
}

// composeBaseArgs builds the `docker compose -p … [-f …] [--project-directory …]`
// prefix shared by every operation.
func composeBaseArgs(p engine.ComposeProject) []string {
	args := []string{"compose", "-p", p.Name}
	if p.WorkingDir != "" {
		args = append(args, "--project-directory", p.WorkingDir)
	}
	for _, f := range strings.Split(p.ConfigFiles, ",") {
		if f = strings.TrimSpace(f); f != "" {
			args = append(args, "-f", f)
		}
	}
	return args
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	return lines[len(lines)-1]
}
