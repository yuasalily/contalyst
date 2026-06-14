package dockerx

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// ComposeProject is the UI-facing view model for a docker-compose project,
// reconstructed from the com.docker.compose.* labels on its containers (U9 /
// FR-CMP1). Contalyst does not parse compose files itself; it groups the live
// containers and delegates lifecycle operations to the `docker compose` CLI
// (DR-5).
type ComposeProject struct {
	Name        string
	Services    int    // distinct compose services seen
	Containers  int    // total containers in the project
	Running     int    // containers currently up
	State       string // "up" (all running), "degraded" (some), "down" (none)
	ConfigFiles string // first non-empty config_files label (for -f)
	WorkingDir  string // first non-empty working_dir label (for --project-directory)
}

// ComposeProjects groups containers by their compose project label and returns
// one summary per project, sorted by name. Containers without a project label
// are ignored. Pure function (no daemon) so it is unit-testable.
func ComposeProjects(containers []Container) []ComposeProject {
	byProject := map[string]*ComposeProject{}
	services := map[string]map[string]struct{}{}
	for _, c := range containers {
		if c.Project == "" {
			continue
		}
		p := byProject[c.Project]
		if p == nil {
			p = &ComposeProject{Name: c.Project}
			byProject[c.Project] = p
			services[c.Project] = map[string]struct{}{}
		}
		p.Containers++
		if isUpState(c.State) {
			p.Running++
		}
		if c.Service != "" {
			services[c.Project][c.Service] = struct{}{}
		}
		if p.ConfigFiles == "" && c.ConfigFiles != "" {
			p.ConfigFiles = c.ConfigFiles
		}
		if p.WorkingDir == "" && c.WorkingDir != "" {
			p.WorkingDir = c.WorkingDir
		}
	}

	out := make([]ComposeProject, 0, len(byProject))
	for name, p := range byProject {
		p.Services = len(services[name])
		if p.Services == 0 {
			p.Services = p.Containers
		}
		p.State = composeState(p.Running, p.Containers)
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func isUpState(s string) bool {
	return s == "running" || s == "restarting" || s == "paused"
}

func composeState(running, total int) string {
	switch {
	case total == 0 || running == 0:
		return "down"
	case running == total:
		return "up"
	default:
		return "degraded"
	}
}

// ComposeAvailable reports whether the `docker compose` v2 plugin is usable.
// Compose operations shell out to it (DR-5); when it is missing the compose
// view stays read-only and explains why (R9 / FR-CMP7).
func ComposeAvailable(ctx context.Context) bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}
	return exec.CommandContext(ctx, "docker", "compose", "version").Run() == nil
}

// ComposeOp is a project-level compose lifecycle operation.
type ComposeOp int

const (
	ComposeUp ComposeOp = iota
	ComposeDown
	ComposeRestart
	ComposeBuild
	ComposeBuildNoCache
)

func (o ComposeOp) args() []string {
	switch o {
	case ComposeDown:
		return []string{"down"}
	case ComposeRestart:
		return []string{"restart"}
	case ComposeBuild:
		return []string{"build"}
	case ComposeBuildNoCache:
		return []string{"build", "--no-cache"}
	default: // ComposeUp
		return []string{"up", "-d"}
	}
}

// Compose runs a compose lifecycle operation for a project by delegating to the
// `docker compose` CLI, which honours depends_on ordering for free (FR-CMP5).
// The project's config-file and working-directory labels are passed through so
// up/build can find the compose file even though Contalyst never reads it.
func (c *Client) Compose(ctx context.Context, p ComposeProject, op ComposeOp) error {
	args := composeBaseArgs(p)
	args = append(args, op.args()...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			return fmt.Errorf("docker compose %s: %w", strings.Join(op.args(), " "), err)
		}
		return fmt.Errorf("docker compose %s: %s", strings.Join(op.args(), " "), lastLine(msg))
	}
	return nil
}

// composeBaseArgs builds the `docker compose -p … [-f …] [--project-directory …]`
// prefix shared by every operation.
func composeBaseArgs(p ComposeProject) []string {
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
