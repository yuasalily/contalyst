// Package dockerx isolates all Docker Engine SDK access behind a small set of
// domain types and methods. The rest of the application depends on this package
// rather than on github.com/docker/docker directly, so that the upstream SDK
// (which is migrating toward github.com/moby/moby) can be swapped without
// touching the UI. See aidlc-docs/inception NFR-M1 / DR-2.
package dockerx

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// Client is a thin wrapper over the Docker Engine API client.
type Client struct {
	api *client.Client
}

// NewClient connects to the Docker daemon using the standard environment
// (DOCKER_HOST, etc.) and negotiates the API version so a single binary works
// across a range of Engine versions (avoids the dry/ctop crash-on-launch class
// of bug — see inception R2).
func NewClient() (*Client, error) {
	api, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}
	return &Client{api: api}, nil
}

// Ping verifies the daemon is reachable, returning a human-actionable error.
func (c *Client) Ping(ctx context.Context) error {
	if _, err := c.api.Ping(ctx); err != nil {
		return fmt.Errorf("cannot reach the Docker daemon — is it running and is DOCKER_HOST correct?\n  %w", err)
	}
	return nil
}

// ServerVersion returns the daemon version string (best-effort).
func (c *Client) ServerVersion(ctx context.Context) string {
	v, err := c.api.ServerVersion(ctx)
	if err != nil {
		return "unknown"
	}
	return v.Version
}

// Close releases the underlying client.
func (c *Client) Close() error { return c.api.Close() }

// Container is the UI-facing view model for a container.
type Container struct {
	ID      string
	Name    string
	Image   string
	State   string // running, exited, paused, created, restarting, removing, dead
	Status  string // human status, e.g. "Up 2 minutes"
	Ports   string // compact published-port summary
	Created time.Time
}

// Containers lists all containers (running and stopped), newest first.
func (c *Client) Containers(ctx context.Context) ([]Container, error) {
	summaries, err := c.api.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	out := make([]Container, 0, len(summaries))
	for _, s := range summaries {
		out = append(out, Container{
			ID:      s.ID,
			Name:    primaryName(s.Names),
			Image:   shortImage(s.Image),
			State:   s.State,
			Status:  s.Status,
			Ports:   formatPorts(s.Ports),
			Created: time.Unix(s.Created, 0),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Created.After(out[j].Created) })
	return out, nil
}

func (c *Client) Start(ctx context.Context, id string) error {
	return c.api.ContainerStart(ctx, id, container.StartOptions{})
}

func (c *Client) Stop(ctx context.Context, id string) error {
	return c.api.ContainerStop(ctx, id, container.StopOptions{})
}

func (c *Client) Restart(ctx context.Context, id string) error {
	return c.api.ContainerRestart(ctx, id, container.StopOptions{})
}

func (c *Client) Pause(ctx context.Context, id string) error {
	return c.api.ContainerPause(ctx, id)
}

func (c *Client) Unpause(ctx context.Context, id string) error {
	return c.api.ContainerUnpause(ctx, id)
}

func (c *Client) Kill(ctx context.Context, id string) error {
	return c.api.ContainerKill(ctx, id, "KILL")
}

func (c *Client) Remove(ctx context.Context, id string, force bool) error {
	return c.api.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}

// Inspect returns the raw, indented JSON for a container.
func (c *Client) Inspect(ctx context.Context, id string) (string, error) {
	_, raw, err := c.api.ContainerInspectWithRaw(ctx, id, false)
	if err != nil {
		return "", err
	}
	return indentJSON(raw), nil
}

// HasTTY reports whether the container was created with a TTY, which determines
// how its log/attach stream must be decoded (see logs.go, inception R1).
func (c *Client) HasTTY(ctx context.Context, id string) (bool, error) {
	info, err := c.api.ContainerInspect(ctx, id)
	if err != nil {
		return false, err
	}
	return info.Config != nil && info.Config.Tty, nil
}

// PruneContainers removes stopped containers.
func (c *Client) PruneContainers(ctx context.Context) error {
	_, err := c.api.ContainersPrune(ctx, filters.Args{})
	return err
}

func primaryName(names []string) string {
	if len(names) == 0 {
		return "<none>"
	}
	return strings.TrimPrefix(names[0], "/")
}

func shortImage(img string) string {
	// Drop the sha256: digest form which is unreadable in a list.
	if strings.HasPrefix(img, "sha256:") {
		return img[7:19]
	}
	return img
}

func formatPorts(ports []container.Port) string {
	seen := map[string]struct{}{}
	var parts []string
	for _, p := range ports {
		if p.PublicPort == 0 {
			continue // only show published ports to keep the column compact
		}
		s := fmt.Sprintf("%d→%d/%s", p.PublicPort, p.PrivatePort, p.Type)
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		parts = append(parts, s)
	}
	return strings.Join(parts, " ")
}
