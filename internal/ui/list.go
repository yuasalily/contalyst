package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// The list is a custom table renderer (rather than bubbles/table) so individual
// cells can carry their own color — e.g. the state glyph green/red — which is
// central to the colorful look the brief asks for. It owns cursor movement,
// vertical scrolling, and width allocation across columns.

type column struct {
	title string
	width int // fixed width in cells; 0 means flexible (shares leftover space)
	min   int // minimum width when flexible
}

type cell struct {
	text  string
	color lipgloss.Color // zero value -> default row color
	faint bool
}

type listRow struct {
	id    string // docker id / name used by actions
	name  string // display name used in breadcrumbs and dialogs
	cells []cell
}

type list struct {
	cols   []column
	rows   []listRow
	cursor int
	offset int
	width  int
	height int // visible body rows (excludes the column header line)

	// emptyMsg overrides the "(nothing here)" placeholder when set (e.g. the
	// compose view explaining the plugin is missing). marked carries the bulk
	// multi-select set (U10); a "✓" gutter is drawn when it is non-nil.
	emptyMsg string
	marked   map[string]bool
}

func (l *list) setRows(rows []listRow) {
	l.rows = rows
	if l.cursor >= len(rows) {
		l.cursor = max(0, len(rows)-1)
	}
	l.clampOffset()
}

func (l *list) selected() (listRow, bool) {
	if l.cursor < 0 || l.cursor >= len(l.rows) {
		return listRow{}, false
	}
	return l.rows[l.cursor], true
}

func (l *list) moveUp() {
	if l.cursor > 0 {
		l.cursor--
	}
	l.clampOffset()
}

func (l *list) moveDown() {
	if l.cursor < len(l.rows)-1 {
		l.cursor++
	}
	l.clampOffset()
}

func (l *list) top()    { l.cursor = 0; l.clampOffset() }
func (l *list) bottom() { l.cursor = len(l.rows) - 1; l.clampOffset() }

func (l *list) clampOffset() {
	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.height <= 0 {
		l.offset = 0
		return
	}
	if l.cursor < l.offset {
		l.offset = l.cursor
	}
	if l.cursor >= l.offset+l.height {
		l.offset = l.cursor - l.height + 1
	}
	if l.offset < 0 {
		l.offset = 0
	}
}

// resolveWidths returns the rendered width of each column for the current total.
func (l *list) resolveWidths() []int {
	const gap = 2
	widths := make([]int, len(l.cols))
	fixed, flex := 0, 0
	for i, c := range l.cols {
		if c.width > 0 {
			widths[i] = c.width
			fixed += c.width
		} else {
			flex++
		}
	}
	gaps := gap * (len(l.cols) - 1)
	leftover := l.width - l.gutter() - fixed - gaps
	if flex > 0 {
		each := leftover / flex
		for i, c := range l.cols {
			if c.width == 0 {
				w := each
				if w < c.min {
					w = c.min
				}
				widths[i] = w
			}
		}
	}
	return widths
}

func (l *list) view(s styles) string {
	widths := l.resolveWidths()
	gut := l.gutter()
	var b strings.Builder

	// Column header (with a blank mark gutter when marking is enabled).
	var head []string
	for i, c := range l.cols {
		head = append(head, pad(c.title, widths[i]))
	}
	b.WriteString(s.colHeader.Render(strings.Repeat(" ", gut) + strings.Join(head, "  ")))
	b.WriteString("\n")

	if len(l.rows) == 0 {
		msg := l.emptyMsg
		if msg == "" {
			msg = "(nothing here)"
		}
		b.WriteString(s.empty.Render("  " + msg))
		return b.String()
	}

	end := min(l.offset+l.height, len(l.rows))
	for ri := l.offset; ri < end; ri++ {
		r := l.rows[ri]
		selected := ri == l.cursor
		marked := l.marked != nil && l.marked[r.id]
		var cells []string
		for ci := range l.cols {
			var cl cell
			if ci < len(r.cells) {
				cl = r.cells[ci]
			}
			txt := pad(cl.text, widths[ci])
			if selected {
				// Color is applied uniformly by the row background style below,
				// so emit plain text here to avoid nested ANSI sequences.
				cells = append(cells, txt)
				continue
			}
			st := lipgloss.NewStyle().Foreground(s.th.Fg)
			if cl.color != "" {
				st = st.Foreground(cl.color)
			} else if cl.faint {
				st = st.Foreground(s.th.Muted)
			}
			cells = append(cells, st.Render(txt))
		}

		gutter := ""
		if gut > 0 {
			mark := " "
			if marked {
				mark = "✓"
			}
			if selected {
				gutter = mark + " "
			} else {
				gutter = s.markGutter.Render(mark) + " "
			}
		}

		line := strings.Join(cells, "  ")
		if selected {
			line = s.rowSel.Width(l.width).Render(gutter + line)
		} else {
			line = gutter + line
		}
		b.WriteString(line)
		if ri < end-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

// gutter returns the width reserved for the bulk-mark indicator (0 when the
// active list does not support marking).
func (l *list) gutter() int {
	if l.marked == nil {
		return 0
	}
	return 2
}

// pad truncates (with an ellipsis) or right-pads s to exactly w display cells.
func pad(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return runewidth.FillRight(runewidth.Truncate(s, w, "…"), w)
}
