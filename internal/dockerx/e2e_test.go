//go:build e2e

// Package dockerx end-to-end tests run against a real Docker daemon. They are
// guarded by the `e2e` build tag so the default `go test ./...` (unit +
// functional) needs no daemon. Run with: go test -tags e2e -run TestE2E ./...
//
// Fixtures are created/destroyed via the docker CLI; the code under test
// (dockerx) is exercised for the read/stream/lifecycle paths.
package dockerx

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func dockerRun(t *testing.T, args ...string) {
	t.Helper()
	out, err := exec.Command("docker", append([]string{"run", "-d"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker run %v: %v\n%s", args, err, out)
	}
}

func dockerRm(name string) {
	_ = exec.Command("docker", "rm", "-f", name).Run()
}

func newClientOrFatal(t *testing.T) (*Client, context.Context) {
	t.Helper()
	c, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	if err := c.Ping(ctx); err != nil {
		t.Fatalf("Docker daemon not reachable: %v", err)
	}
	return c, ctx
}

func TestE2E_ContainerLifecycleAndStreams(t *testing.T) {
	c, ctx := newClientOrFatal(t)

	name := "contalyst-e2e-logger"
	dockerRm(name)
	dockerRun(t, "--name", name, "busybox:latest", "sh", "-c",
		"i=0; while true; do echo line-$i; i=$((i+1)); sleep 1; done")
	defer dockerRm(name)

	// List must include our container, running.
	var target Container
	waitFor(t, 10*time.Second, func() bool {
		cs, err := c.Containers(ctx)
		if err != nil {
			t.Fatalf("Containers: %v", err)
		}
		for _, x := range cs {
			if x.Name == name {
				target = x
				return x.State == "running"
			}
		}
		return false
	}, "container to appear as running")

	// Logs (non-TTY → exercises stdcopy demux).
	lctx, lcancel := context.WithTimeout(ctx, 6*time.Second)
	defer lcancel()
	ch, err := c.LogStream(lctx, target.ID, false)
	if err != nil {
		t.Fatalf("LogStream: %v", err)
	}
	var lines int
	for ln := range ch {
		if ln.Err != nil {
			break
		}
		if strings.HasPrefix(ln.Text, "line-") {
			lines++
		}
		if lines >= 2 {
			lcancel()
		}
	}
	if lines < 1 {
		t.Errorf("expected log lines, got %d", lines)
	}

	// Stats.
	sctx, scancel := context.WithTimeout(ctx, 6*time.Second)
	defer scancel()
	sch, err := c.StatsStream(sctx, target.ID)
	if err != nil {
		t.Fatalf("StatsStream: %v", err)
	}
	got := false
	for s := range sch {
		if s.Err != nil {
			break
		}
		got = true
		scancel()
	}
	if !got {
		t.Error("expected at least one stats sample")
	}

	// Inspect.
	if js, err := c.Inspect(ctx, target.ID); err != nil || len(js) == 0 {
		t.Errorf("Inspect: err=%v len=%d", err, len(js))
	}

	// Lifecycle: stop → verify exited → remove → verify gone.
	if err := c.Stop(ctx, target.ID); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	waitFor(t, 15*time.Second, func() bool {
		cs, _ := c.Containers(ctx)
		for _, x := range cs {
			if x.Name == name {
				return x.State == "exited"
			}
		}
		return false
	}, "container to be exited")

	if err := c.Remove(ctx, target.ID, true); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	waitFor(t, 10*time.Second, func() bool {
		cs, _ := c.Containers(ctx)
		for _, x := range cs {
			if x.Name == name {
				return false
			}
		}
		return true
	}, "container to be removed")
}

func TestE2E_ResourceLists(t *testing.T) {
	c, ctx := newClientOrFatal(t)
	// These should never error against a live daemon.
	if _, err := c.Images(ctx); err != nil {
		t.Errorf("Images: %v", err)
	}
	if _, err := c.Volumes(ctx); err != nil {
		t.Errorf("Volumes: %v", err)
	}
	if _, err := c.Networks(ctx); err != nil {
		t.Errorf("Networks: %v", err)
	}
}

// TestE2E_MaintenanceReads covers the v2 read-only maintenance paths (U12):
// disk-usage summary and image-layer history against the live daemon.
func TestE2E_MaintenanceReads(t *testing.T) {
	c, ctx := newClientOrFatal(t)

	// Ensure at least one image exists, then read its layer history.
	if out, err := exec.Command("docker", "pull", "busybox:latest").CombinedOutput(); err != nil {
		t.Fatalf("docker pull busybox: %v\n%s", err, out)
	}
	imgs, err := c.Images(ctx)
	if err != nil {
		t.Fatalf("Images: %v", err)
	}
	if len(imgs) == 0 {
		t.Skip("no images to read history from")
	}
	layers, err := c.ImageHistory(ctx, imgs[0].ID)
	if err != nil {
		t.Fatalf("ImageHistory: %v", err)
	}
	if len(layers) == 0 {
		t.Errorf("expected at least one layer for %s", imgs[0].ID)
	}

	// Disk usage must return the full set of prune categories without erroring.
	usage, err := c.DiskUsage(ctx)
	if err != nil {
		t.Fatalf("DiskUsage: %v", err)
	}
	if len(usage) != 5 {
		t.Errorf("expected 5 prune categories, got %d", len(usage))
	}
}

// TestE2E_ComposeAndContexts covers the CLI-backed v2 paths (U9/U11): the
// compose-availability probe and context enumeration. Both must degrade
// gracefully rather than error.
func TestE2E_ComposeAndContexts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Just assert it does not panic / hang; the result depends on the host.
	_ = ComposeAvailable(ctx)

	ctxs, err := Contexts(ctx)
	if err != nil {
		t.Fatalf("Contexts: %v", err)
	}
	// When the docker CLI is present there is always at least the default context.
	if _, lookErr := exec.LookPath("docker"); lookErr == nil && len(ctxs) == 0 {
		t.Error("expected at least one docker context")
	}
}

func waitFor(t *testing.T, timeout time.Duration, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
