// Package engine defines the backend-neutral port that the rest of the
// application depends on. Concrete container engines (Docker today, Podman or
// others later) live in sub-packages that implement the Engine interface and
// translate their SDK/CLI types into the view models declared here. Frontends
// (the Bubble Tea TUI today, a web UI later) depend only on this package, never
// on a specific engine implementation. See aidlc-docs/inception NFR-M1 / DR-2.
package engine

import "context"

// Engine is the set of container-management operations a frontend needs. Every
// method speaks in this package's neutral view models so that swapping the
// underlying engine (or running against multiple at once) touches no UI code.
type Engine interface {
	// Connection / metadata.
	Ping(ctx context.Context) error
	ServerVersion(ctx context.Context) string
	Close() error

	// Containers.
	Containers(ctx context.Context) ([]Container, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Pause(ctx context.Context, id string) error
	Unpause(ctx context.Context, id string) error
	Kill(ctx context.Context, id string) error
	Remove(ctx context.Context, id string, force bool) error
	Inspect(ctx context.Context, id string) (string, error)
	LogStream(ctx context.Context, id string, timestamps bool) (<-chan LogLine, error)
	StatsStream(ctx context.Context, id string) (<-chan Stats, error)
	// ExecSpec returns the external command that opens an interactive shell in a
	// container. The frontend decides how to run it (the TUI suspends and hands
	// over the terminal; a web UI would bridge it over a PTY socket), so the port
	// only describes what to launch — keeping engine-specific binaries (docker vs
	// podman) out of the frontend (inception FR-C9).
	ExecSpec(id string) ExecSpec

	// Images.
	Images(ctx context.Context) ([]Image, error)
	RemoveImage(ctx context.Context, id string, force bool) error
	ImageHistory(ctx context.Context, id string) ([]Layer, error)

	// Volumes.
	Volumes(ctx context.Context) ([]Volume, error)
	RemoveVolume(ctx context.Context, name string, force bool) error

	// Networks.
	Networks(ctx context.Context) ([]Network, error)
	RemoveNetwork(ctx context.Context, id string) error

	// Maintenance.
	DiskUsage(ctx context.Context) ([]Usage, error)
	Prune(ctx context.Context, k PruneKind) error

	// Compose.
	ComposeAvailable(ctx context.Context) bool
	Compose(ctx context.Context, p ComposeProject, op ComposeOp) error
}

// ExecSpec describes an external command to launch (e.g. an interactive shell
// inside a container). Name is the executable; Args are its arguments.
type ExecSpec struct {
	Name string
	Args []string
}
