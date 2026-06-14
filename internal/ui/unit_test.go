package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func TestUnit_HighlightMatches(t *testing.T) {
	style := lipgloss.NewStyle() // no-op styling keeps output comparable
	cases := []struct {
		line, query, want string
	}{
		{"hello world", "", "hello world"},      // empty query is a no-op
		{"hello world", "xyz", "hello world"},   // no match is verbatim
		{"hello world", "o", "hello world"},     // matches survive round-trip
		{"Error: boom", "error", "Error: boom"}, // case-insensitive, original case kept
		{"aXaXa", "x", "aXaXa"},                 // multiple matches
	}
	for _, c := range cases {
		got := highlightMatches(c.line, c.query, style)
		if got != c.want {
			t.Errorf("highlightMatches(%q,%q)=%q, want %q", c.line, c.query, got, c.want)
		}
	}
}

func TestUnit_Pad(t *testing.T) {
	cases := []struct {
		in    string
		w     int
		wantW int
	}{
		{"hi", 5, 5},    // pads right
		{"hello", 3, 3}, // truncates with ellipsis
		{"", 4, 4},      // empty pads
		{"abc", 0, 0},   // zero width
		{"日本語", 4, 4},   // wide runes truncated to display width
	}
	for _, c := range cases {
		got := pad(c.in, c.w)
		if w := runewidth.StringWidth(got); w != c.wantW {
			t.Errorf("pad(%q,%d)=%q has display width %d, want %d", c.in, c.w, got, w, c.wantW)
		}
	}
}

func TestUnit_ResolveWidths(t *testing.T) {
	l := &list{
		cols: []column{
			{title: "A", width: 1},
			{title: "B", min: 10},
			{title: "C", width: 8},
		},
		width: 100,
	}
	w := l.resolveWidths()
	if len(w) != 3 {
		t.Fatalf("want 3 widths, got %d", len(w))
	}
	if w[0] != 1 || w[2] != 8 {
		t.Errorf("fixed columns wrong: %v", w)
	}
	if w[1] < 10 {
		t.Errorf("flexible column below min: %d", w[1])
	}
	// Total (incl. 2-cell gaps) must not exceed the available width.
	total := w[0] + w[1] + w[2] + 2*2
	if total > l.width {
		t.Errorf("columns overflow width: total=%d width=%d", total, l.width)
	}
}

func TestUnit_FuzzyMatch(t *testing.T) {
	cases := []struct {
		pat, s string
		want   bool
	}{
		{"ng", "nginx", true},
		{"nx", "nginx", true},
		{"xyz", "nginx", false},
		{"", "anything", true},
		{"abc", "ab", false},
	}
	for _, c := range cases {
		if got := fuzzyMatch(c.pat, c.s); got != c.want {
			t.Errorf("fuzzyMatch(%q,%q)=%v want %v", c.pat, c.s, got, c.want)
		}
	}
}

func TestUnit_CommandSuggest(t *testing.T) {
	cases := map[string]string{
		"im":  "images",
		"vol": "volumes",
		"net": "networks",
		"q":   "quit",
		"zzz": "",
		"":    "",
	}
	for in, want := range cases {
		if got := commandSuggest(in); got != want {
			t.Errorf("commandSuggest(%q)=%q want %q", in, got, want)
		}
	}
}

func TestUnit_HumanSize(t *testing.T) {
	cases := map[int64]string{
		512:     "512B",
		1024:    "1.0KB",
		1536:    "1.5KB",
		1 << 20: "1.0MB",
		5 << 30: "5.0GB",
	}
	for in, want := range cases {
		if got := humanSize(in); got != want {
			t.Errorf("humanSize(%d)=%q want %q", in, got, want)
		}
	}
}

func TestUnit_RelativeTime(t *testing.T) {
	now := time.Now()
	cases := []struct {
		t    time.Time
		want string
	}{
		{now.Add(-30 * time.Second), "just now"},
		{now.Add(-5 * time.Minute), "5m ago"},
		{now.Add(-3 * time.Hour), "3h ago"},
		{now.Add(-49 * time.Hour), "2d ago"},
	}
	for _, c := range cases {
		if got := relativeTime(c.t); got != c.want {
			t.Errorf("relativeTime(%v)=%q want %q", c.t, got, c.want)
		}
	}
}

func TestUnit_PackHintsRespectsMaxLines(t *testing.T) {
	m := New(nil)
	m.width = 40 // narrow, forces wrapping
	tokens := [][2]string{
		{"a", "alpha"}, {"b", "bravo"}, {"c", "charlie"},
		{"d", "delta"}, {"e", "echo"}, {"f", "foxtrot"},
	}
	out := m.packHints(tokens, 2)
	if lines := strings.Count(out, "\n") + 1; lines > 2 {
		t.Errorf("packHints produced %d lines, want <= 2:\n%s", lines, out)
	}
	if !strings.Contains(out, "a") {
		t.Errorf("packHints dropped the first token:\n%s", out)
	}
}
