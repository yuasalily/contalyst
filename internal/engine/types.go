package engine

import "time"

// Container is the backend-neutral view model for a container.
type Container struct {
	ID      string
	Name    string
	Image   string
	State   string // running, exited, paused, created, restarting, removing, dead
	Status  string // human status, e.g. "Up 2 minutes"
	Ports   string // compact published-port summary
	Created time.Time

	// Compose metadata, taken from the com.docker.compose.* labels (empty for
	// containers not started by compose). Used by the compose project view (U9).
	Project     string
	Service     string
	ConfigFiles string // com.docker.compose.project.config_files
	WorkingDir  string // com.docker.compose.project.working_dir
}

// Image is the backend-neutral view model for an image.
type Image struct {
	ID      string
	Repo    string
	Tag     string
	Size    int64
	Created time.Time
}

// Volume is the backend-neutral view model for a volume.
type Volume struct {
	Name       string
	Driver     string
	Mountpoint string
}

// Network is the backend-neutral view model for a network.
type Network struct {
	ID     string
	Name   string
	Driver string
	Scope  string
}

// Stats is a single sampled snapshot of a container's resource usage.
type Stats struct {
	CPUPercent float64
	MemUsage   uint64
	MemLimit   uint64
	MemPercent float64
	NetRx      uint64
	NetTx      uint64
	BlkRead    uint64
	BlkWrite   uint64
	Pids       uint64
	Err        error
}

// LogLine is a single line emitted by a log stream, or a terminal error.
type LogLine struct {
	Text string
	Err  error // non-nil signals the stream ended (io.EOF on clean close)
}

// Layer is one entry in an image's history (U12 / FR-L1): a build step with the
// disk space it added and the command that produced it.
type Layer struct {
	Size      int64
	CreatedBy string
	Created   time.Time
}

// PruneKind identifies a category in the prune dashboard (U12 / FR-PR1).
type PruneKind int

const (
	PruneKindImages PruneKind = iota
	PruneKindContainers
	PruneKindVolumes
	PruneKindNetworks
	PruneKindBuildCache
)

func (k PruneKind) Label() string {
	switch k {
	case PruneKindContainers:
		return "stopped containers"
	case PruneKindVolumes:
		return "volumes"
	case PruneKindNetworks:
		return "networks"
	case PruneKindBuildCache:
		return "build cache"
	default:
		return "images"
	}
}

// Usage is the reclaimable-space summary for one prune category.
type Usage struct {
	Kind        PruneKind
	Count       int
	Reclaimable int64
}

// IsUpState reports whether a container state counts as "up" for the purpose of
// compose project health and reclaimable-space accounting.
func IsUpState(s string) bool {
	return s == "running" || s == "restarting" || s == "paused"
}
