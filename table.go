package quill

import "strings"

// TableAlign controls column text alignment.
type TableAlign int

const (
	TableAlignLeft TableAlign = iota
	TableAlignCenter
	TableAlignRight
)

// TableColumn defines a column in a Table.
type TableColumn struct {
	Header string
	Width  int        // fixed width in cells; 0 = auto-fit content
	Align  TableAlign // text alignment within the column
}

type tableChars struct {
	h, v                   rune
	tl, tm, tr             rune // top border
	ml, mm, mr             rune // separator (between header and body)
	bl, bm, br             rune // bottom border
}

var tableCharSets = map[BorderStyle]tableChars{
	BorderSingle: {
		h: '─', v: '│',
		tl: '┌', tm: '┬', tr: '┐',
		ml: '├', mm: '┼', mr: '┤',
		bl: '└', bm: '┴', br: '┘',
	},
	BorderDouble: {
		h: '═', v: '║',
		tl: '╔', tm: '╦', tr: '╗',
		ml: '╠', mm: '╬', mr: '╣',
		bl: '╚', bm: '╩', br: '╝',
	},
	BorderRounded: {
		h: '─', v: '│',
		tl: '╭', tm: '┬', tr: '╮',
		ml: '├', mm: '┼', mr: '┤',
		bl: '╰', bm: '┴', br: '╯',
	},
	BorderThick: {
		h: '━', v: '┃',
		tl: '┏', tm: '┳', tr: '┓',
		ml: '┣', mm: '╋', mr: '┫',
		bl: '┗', bm: '┻', br: '┛',
	},
}

// Table creates a table element with column headers and data rows.
// Pass a BorderStyle (e.g. BorderRounded) in args to draw box-drawing borders.
// Without a border, columns are separated by spaces with a dashed separator.
//
//	quill.Table(
//	    []quill.TableColumn{
//	        {Header: "Name", Width: 20},
//	        {Header: "Age", Width: 5, Align: quill.TableAlignRight},
//	    },
//	    [][]string{
//	        {"Alice", "30"},
//	        {"Bob", "25"},
//	    },
//	    quill.BorderRounded,
//	)
func Table(columns []TableColumn, rows [][]string, args ...any) *Node {
	widths := resolveColumnWidths(columns, rows)

	var bs BorderStyle
	for _, arg := range args {
		if b, ok := arg.(BorderStyle); ok {
			bs = b
		}
	}

	var children []any
	children = append(children, FlexColumn)

	// Collect non-border props.
	for _, arg := range args {
		if _, ok := arg.(BorderStyle); ok {
			continue
		}
		if _, ok := arg.(prop); ok {
			children = append(children, arg)
		}
	}

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}

	if bs != BorderNone {
		tc := tableCharSets[bs]
		children = append(children, Text(tableHLine(tc.tl, tc.tm, tc.tr, tc.h, widths), ClipText))
		children = append(children, Text(tableDataLine(tc.v, columns, widths, headers), Bold, ClipText))
		children = append(children, Text(tableHLine(tc.ml, tc.mm, tc.mr, tc.h, widths), ClipText))
		for _, row := range rows {
			children = append(children, Text(tableDataLine(tc.v, columns, widths, row), ClipText))
		}
		children = append(children, Text(tableHLine(tc.bl, tc.bm, tc.br, tc.h, widths), ClipText))
	} else {
		children = append(children, Text(tableSimpleLine(columns, widths, headers), Bold, ClipText))
		sep := tableSimpleSep(widths)
		children = append(children, Text(sep, ClipText))
		for _, row := range rows {
			children = append(children, Text(tableSimpleLine(columns, widths, row), ClipText))
		}
	}

	return Box(children...)
}

func resolveColumnWidths(columns []TableColumn, rows [][]string) []int {
	widths := make([]int, len(columns))
	for i, col := range columns {
		if col.Width > 0 {
			widths[i] = col.Width
		} else {
			w := stringWidth(col.Header)
			for _, row := range rows {
				if i < len(row) {
					rw := stringWidth(row[i])
					if rw > w {
						w = rw
					}
				}
			}
			widths[i] = w
		}
	}
	return widths
}

func tableHLine(left, mid, right, h rune, widths []int) string {
	var b strings.Builder
	b.WriteRune(left)
	for i, w := range widths {
		if i > 0 {
			b.WriteRune(mid)
		}
		for j := 0; j < w+2; j++ {
			b.WriteRune(h)
		}
	}
	b.WriteRune(right)
	return b.String()
}

func tableDataLine(v rune, columns []TableColumn, widths []int, cells []string) string {
	var b strings.Builder
	b.WriteRune(v)
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		align := TableAlignLeft
		if i < len(columns) {
			align = columns[i].Align
		}
		b.WriteString(" " + alignTableText(cell, w, align) + " ")
		b.WriteRune(v)
	}
	return b.String()
}

func tableSimpleLine(columns []TableColumn, widths []int, cells []string) string {
	var b strings.Builder
	for i, w := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		align := TableAlignLeft
		if i < len(columns) {
			align = columns[i].Align
		}
		b.WriteString(alignTableText(cell, w, align))
	}
	return b.String()
}

func tableSimpleSep(widths []int) string {
	var b strings.Builder
	for i, w := range widths {
		if i > 0 {
			b.WriteString("  ")
		}
		b.WriteString(strings.Repeat("─", w))
	}
	return b.String()
}

func alignTableText(text string, width int, align TableAlign) string {
	tw := stringWidth(text)
	if tw >= width {
		return truncateToWidth(text, width)
	}
	pad := width - tw
	switch align {
	case TableAlignRight:
		return strings.Repeat(" ", pad) + text
	case TableAlignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", pad)
	}
}
