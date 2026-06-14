package ui

import (
	"strings"
	"testing"

	"github.com/yuasalily/contalyst/internal/dockerx"
)

func TestUnit_FormatLayers(t *testing.T) {
	layers := []dockerx.Layer{
		{Size: 78 << 20, CreatedBy: "/bin/sh -c #(nop)  CMD [\"nginx\"]"},
		{Size: 0, CreatedBy: "/bin/sh -c apt-get update"},
	}
	out := formatLayers(layers)
	if !strings.Contains(out, "SIZE") || !strings.Contains(out, "CREATED BY") {
		t.Errorf("missing header:\n%s", out)
	}
	if !strings.Contains(out, "78.0MB") {
		t.Errorf("missing humanized size:\n%s", out)
	}
	// The /bin/sh -c #(nop) noise should be stripped.
	if strings.Contains(out, "#(nop)") {
		t.Errorf("nop prefix not stripped:\n%s", out)
	}
	if !strings.Contains(out, "apt-get update") {
		t.Errorf("plain command prefix not stripped:\n%s", out)
	}
}

func TestUnit_Plural(t *testing.T) {
	if plural(1) != "y" {
		t.Errorf("plural(1)=%q want y", plural(1))
	}
	if plural(0) != "ies" || plural(3) != "ies" {
		t.Errorf("plural(n>1) wrong")
	}
}

func TestUnit_TruncMiddle(t *testing.T) {
	if got := truncMiddle("short", 40); got != "short" {
		t.Errorf("short string changed: %q", got)
	}
	long := "ssh://user@really-long-host.example.internal:2376"
	got := truncMiddle(long, 20)
	if len([]rune(got)) > 20 {
		t.Errorf("truncMiddle too long: %q (%d)", got, len([]rune(got)))
	}
	if !strings.Contains(got, "…") {
		t.Errorf("truncMiddle should ellipsize: %q", got)
	}
}

func TestUnit_ComposeStateGlyph(t *testing.T) {
	cases := map[string]string{"up": "●", "degraded": "◐", "down": "○"}
	for state, want := range cases {
		if got := composeStateGlyph(state); got != want {
			t.Errorf("composeStateGlyph(%q)=%q want %q", state, got, want)
		}
	}
}
