package dockerx

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/client"
)

// DockerContext is a Docker CLI context (a named connection target), used by the
// host/context switcher (U11 / FR-H1). Contalyst reads the contexts from the
// `docker context ls` CLI rather than re-implementing the on-disk format
// (~/.docker/contexts), consistent with the shell-out approach taken elsewhere
// (DR-5/DR-6).
type DockerContext struct {
	Name    string
	Host    string // DOCKER_HOST endpoint, e.g. unix:///var/run/docker.sock
	Current bool
}

// Contexts lists the available Docker contexts. It returns an empty slice
// (no error) when the docker CLI is unavailable, so the switcher degrades to
// "current connection only" rather than failing.
func Contexts(ctx context.Context) ([]DockerContext, error) {
	if _, err := exec.LookPath("docker"); err != nil {
		return nil, nil
	}
	out, err := exec.CommandContext(ctx, "docker", "context", "ls", "--format", "{{json .}}").Output()
	if err != nil {
		return nil, fmt.Errorf("listing docker contexts: %w", err)
	}
	var ctxs []DockerContext
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var raw struct {
			Name           string `json:"Name"`
			Current        bool   `json:"Current"`
			DockerEndpoint string `json:"DockerEndpoint"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}
		ctxs = append(ctxs, DockerContext{Name: raw.Name, Host: raw.DockerEndpoint, Current: raw.Current})
	}
	return ctxs, nil
}

// NewClientForHost connects to a specific Docker host endpoint, negotiating the
// API version like NewClient. An empty host falls back to the environment
// (DOCKER_HOST / default socket), matching NewClient's behaviour.
func NewClientForHost(host string) (*Client, error) {
	opts := []client.Opt{client.WithAPIVersionNegotiation()}
	if host == "" {
		opts = append([]client.Opt{client.FromEnv}, opts...)
	} else {
		opts = append(opts, client.WithHost(host))
	}
	api, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("connecting to %s: %w", host, err)
	}
	return &Client{api: api}, nil
}
