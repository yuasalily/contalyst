package engine

import "sort"

// ComposeProject is the backend-neutral view model for a docker-compose project,
// reconstructed from the com.docker.compose.* labels on its containers (U9 /
// FR-CMP1). Contalyst does not parse compose files itself; it groups the live
// containers and delegates lifecycle operations to the engine's compose CLI
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

// ComposeOp is a project-level compose lifecycle operation.
type ComposeOp int

const (
	ComposeUp ComposeOp = iota
	ComposeDown
	ComposeRestart
	ComposeBuild
	ComposeBuildNoCache
)

// Args returns the compose subcommand arguments for the operation (e.g.
// {"up", "-d"}). Exported so backend adapters can build the CLI invocation.
func (o ComposeOp) Args() []string {
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

// ComposeProjects groups containers by their compose project label and returns
// one summary per project, sorted by name. Containers without a project label
// are ignored. Pure function (no daemon) so it is unit-testable and shared
// across backends.
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
		if IsUpState(c.State) {
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
