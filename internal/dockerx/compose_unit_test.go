package dockerx

import (
	"reflect"
	"testing"

	"github.com/yuasalily/contalyst/internal/engine"
)

func TestUnit_ComposeBaseArgs(t *testing.T) {
	p := engine.ComposeProject{Name: "app", ConfigFiles: "/a/compose.yaml,/a/override.yaml", WorkingDir: "/a"}
	got := composeBaseArgs(p)
	want := []string{"compose", "-p", "app", "--project-directory", "/a", "-f", "/a/compose.yaml", "-f", "/a/override.yaml"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("composeBaseArgs=%v\nwant %v", got, want)
	}

	// No labels: bare project flag only.
	bare := composeBaseArgs(engine.ComposeProject{Name: "x"})
	if !reflect.DeepEqual(bare, []string{"compose", "-p", "x"}) {
		t.Errorf("bare composeBaseArgs=%v", bare)
	}
}

func TestUnit_LastLine(t *testing.T) {
	if got := lastLine("a\nb\nerror: boom\n"); got != "error: boom" {
		t.Errorf("lastLine=%q", got)
	}
	if got := lastLine("single"); got != "single" {
		t.Errorf("lastLine single=%q", got)
	}
}
