package theme

import "testing"

func TestUnit_StateColor(t *testing.T) {
	th := Default()
	cases := map[string]string{
		"running":    string(th.Running),
		"exited":     string(th.Exited),
		"dead":       string(th.Exited),
		"paused":     string(th.Paused),
		"restarting": string(th.Transition),
		"created":    string(th.Transition),
		"weird":      string(th.Muted),
	}
	for state, want := range cases {
		if got := string(th.StateColor(state)); got != want {
			t.Errorf("StateColor(%q)=%q want %q", state, got, want)
		}
	}
}

func TestUnit_StateGlyph(t *testing.T) {
	cases := map[string]string{
		"running": "●",
		"paused":  "‖",
		"exited":  "○",
		"created": "◐",
		"unknown": "•",
	}
	for state, want := range cases {
		if got := StateGlyph(state); got != want {
			t.Errorf("StateGlyph(%q)=%q want %q", state, got, want)
		}
	}
}

func TestUnit_NextCycles(t *testing.T) {
	start := Default().Name
	seen := map[string]bool{start: true}
	name := start
	for i := 0; i < 10; i++ {
		name = Next(name).Name
		if name == start {
			break
		}
		seen[name] = true
	}
	if len(seen) < 2 {
		t.Errorf("Next did not cycle through multiple themes: %v", seen)
	}
	// Cycling from an unknown name returns the default.
	if Next("does-not-exist").Name != Default().Name {
		t.Errorf("Next(unknown) should return default")
	}
}
