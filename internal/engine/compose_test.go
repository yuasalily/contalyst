package engine

import (
	"reflect"
	"testing"
)

func TestUnit_ComposeProjects(t *testing.T) {
	containers := []Container{
		{ID: "1", Project: "app", Service: "web", State: "running", ConfigFiles: "/srv/app/compose.yaml", WorkingDir: "/srv/app"},
		{ID: "2", Project: "app", Service: "db", State: "running"},
		{ID: "3", Project: "app", Service: "web", State: "exited"}, // 2nd replica of web, stopped
		{ID: "4", Project: "data", Service: "minio", State: "exited"},
		{ID: "5", Project: "", Service: "", State: "running"}, // not a compose container
	}
	got := ComposeProjects(containers)
	if len(got) != 2 {
		t.Fatalf("want 2 projects, got %d (%+v)", len(got), got)
	}
	// Sorted by name: app, data.
	app := got[0]
	if app.Name != "app" {
		t.Fatalf("want app first, got %q", app.Name)
	}
	if app.Containers != 3 || app.Running != 2 {
		t.Errorf("app containers/running = %d/%d, want 3/2", app.Containers, app.Running)
	}
	if app.Services != 2 { // web + db (web counted once)
		t.Errorf("app services = %d, want 2", app.Services)
	}
	if app.State != "degraded" {
		t.Errorf("app state = %q, want degraded", app.State)
	}
	if app.ConfigFiles != "/srv/app/compose.yaml" || app.WorkingDir != "/srv/app" {
		t.Errorf("app config/workdir not captured: %q / %q", app.ConfigFiles, app.WorkingDir)
	}
	data := got[1]
	if data.State != "down" {
		t.Errorf("data state = %q, want down", data.State)
	}
}

func TestUnit_ComposeState(t *testing.T) {
	cases := []struct {
		running, total int
		want           string
	}{
		{0, 0, "down"},
		{0, 3, "down"},
		{2, 3, "degraded"},
		{3, 3, "up"},
	}
	for _, c := range cases {
		if got := composeState(c.running, c.total); got != c.want {
			t.Errorf("composeState(%d,%d)=%q want %q", c.running, c.total, got, c.want)
		}
	}
}

func TestUnit_ComposeOpArgs(t *testing.T) {
	cases := map[ComposeOp][]string{
		ComposeUp:           {"up", "-d"},
		ComposeDown:         {"down"},
		ComposeRestart:      {"restart"},
		ComposeBuild:        {"build"},
		ComposeBuildNoCache: {"build", "--no-cache"},
	}
	for op, want := range cases {
		if got := op.Args(); !reflect.DeepEqual(got, want) {
			t.Errorf("op %d args=%v want %v", op, got, want)
		}
	}
}
