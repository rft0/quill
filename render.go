package quill

import "sort"

// Paint holds the visual styling for a node (colors, text attributes).
// It is independent of the flexbox layout properties in Style.
type Paint struct {
	FG            Color
	BG            Color
	BorderFG      Color // border color; falls back to FG if default
	Bold          bool
	Italic        bool
	Underline     bool
	Dim           bool
	Strikethrough bool
	Reverse       bool
}

// Render walks the layout tree rooted at node and paints each node onto
// canvas. Compute or ComputeInline must be called first.
func Render(root *Node, canvas *Canvas) {
	renderNode(root, canvas, nil, ColorDefault, CellAttrs{}, false, 0)
}

type clipRect struct {
	x, y, w, h int
}

// debugColors are cycled through when Debug is enabled.
var debugColors = [...]Color{
	ANSIColor(1), // red
	ANSIColor(2), // green
	ANSIColor(3), // yellow
	ANSIColor(4), // blue
	ANSIColor(5), // magenta
	ANSIColor(6), // cyan
}

func renderNode(node *Node, canvas *Canvas, clip *clipRect, inheritFG Color, inheritAttrs CellAttrs, debug bool, depth int) {
	x := int(node.Layout.X)
	y := int(node.Layout.Y)
	w := int(node.Layout.Width)
	h := int(node.Layout.Height)
	bw := int(node.Style.BorderWidth())

	// Resolve inherited foreground color.
	effectiveFG := node.Paint.FG
	if effectiveFG.IsDefault() {
		effectiveFG = inheritFG
	}

	// Merge inherited text attributes with this node's own.
	// Reverse and BG do not inherit (same as CSS/Ink).
	effectiveAttrs := CellAttrs{
		Bold:          node.Paint.Bold || inheritAttrs.Bold,
		Italic:        node.Paint.Italic || inheritAttrs.Italic,
		Underline:     node.Paint.Underline || inheritAttrs.Underline,
		Dim:           node.Paint.Dim || inheritAttrs.Dim,
		Strikethrough: node.Paint.Strikethrough || inheritAttrs.Strikethrough,
		Reverse:       node.Paint.Reverse,
	}

	// Fill background across the node's entire box.
	if !node.Paint.BG.IsDefault() {
		fillRectClipped(canvas, x, y, w, h, node.Paint.BG, clip)
	}

	// Draw border (use effective FG for border color inheritance).
	if node.Style.Border != BorderNone {
		borderPaint := node.Paint
		borderPaint.FG = effectiveFG
		drawBorder(canvas, x, y, w, h, node.Style.Border, borderPaint, node.borderTitle, clip)
	}

	// Debug: propagate from node or parent, draw colored outline.
	if node.debug {
		debug = true
	}
	if debug && w >= 2 && h >= 2 {
		dc := debugColors[depth%len(debugColors)]
		drawBorder(canvas, x, y, w, h, BorderSingle, Paint{FG: dc, BorderFG: dc}, "", clip)
	}

	// Draw text content for leaf nodes.
	if node.IsLeaf() && (node.Text != "" || node.showCursor || node.isProgress) {
		contentX := x + int(node.Style.Padding.Left) + bw
		contentY := y + int(node.Style.Padding.Top) + bw
		contentW := w - int(node.Style.Padding.Horizontal()) - bw*2
		if contentW <= 0 {
			contentW = w
		}

		// Progress bar: generate bar text dynamically from available width.
		text := node.Text
		if node.isProgress {
			text = renderProgressBar(contentW, node.progressValue)
		}

		// Apply text overflow.
		switch node.TextOverflow {
		case TextOverflowEllipsis:
			if contentW > 0 && stringWidth(text) > contentW {
				if contentW > 1 {
					text = truncateToWidth(text, contentW-1) + "\u2026"
				} else {
					text = "\u2026"
				}
			}
		case TextOverflowClip:
			if contentW > 0 && stringWidth(text) > contentW {
				text = truncateToWidth(text, contentW)
			}
		}

		attrs := effectiveAttrs

		col, row, runeIdx := 0, 0, 0
		for _, r := range text {
			rw := runeWidth(r)
			if rw == 0 {
				continue // skip combining marks
			}

			// Clip mode: stop on first line overflow.
			if node.TextOverflow == TextOverflowClip && contentW > 0 && col+rw > contentW {
				break
			}

			if contentW > 0 && col+rw > contentW {
				row++
				col = 0
			}
			c := Cell{Rune: r, FG: effectiveFG, BG: node.Paint.BG, Attrs: attrs}
			if node.showCursor && runeIdx == node.cursorPos {
				c.Attrs.Reverse = true
			}
			setClipped(canvas, contentX+col, contentY+row, c, clip)
			col++
			// Place continuation cell for wide characters.
			if rw == 2 {
				setClipped(canvas, contentX+col, contentY+row, Cell{Wide: true}, clip)
				col++
			}
			runeIdx++
		}
		// If cursor is at end of text, draw a reversed space.
		if node.showCursor && node.cursorPos >= runeIdx {
			if contentW > 0 && col >= contentW {
				row++
				col = 0
			}
			setClipped(canvas, contentX+col, contentY+row, Cell{
				Rune:  ' ',
				FG:    effectiveFG,
				BG:    node.Paint.BG,
				Attrs: CellAttrs{Reverse: true},
			}, clip)
		}
	}

	// Handle scroll view: offset children and clip to viewport.
	childClip := clip
	scrollOffset := 0
	if node.scrollState != nil {
		innerX := x + int(node.Style.Padding.Left) + bw
		innerY := y + int(node.Style.Padding.Top) + bw
		innerW := w - int(node.Style.Padding.Horizontal()) - bw*2
		innerH := h - int(node.Style.Padding.Vertical()) - bw*2

		node.scrollState.ViewHeight = innerH

		// Compute total content height from children.
		totalH := 0
		for _, child := range node.Children {
			bottom := int(child.Layout.Y) + int(child.Layout.Height) - y
			if bottom > totalH {
				totalH = bottom
			}
		}
		node.scrollState.ContentHeight = totalH

		scrollOffset = node.scrollState.Offset
		childClip = &clipRect{x: innerX, y: innerY, w: innerW, h: innerH}
	}

	// Sort children by ZIndex (stable to preserve DOM order for equal z).
	children := node.Children
	if len(children) > 1 {
		needsSort := false
		for _, child := range children {
			if child.Style.ZIndex != 0 {
				needsSort = true
				break
			}
		}
		if needsSort {
			sorted := make([]*Node, len(children))
			copy(sorted, children)
			sort.SliceStable(sorted, func(i, j int) bool {
				return sorted[i].Style.ZIndex < sorted[j].Style.ZIndex
			})
			children = sorted
		}
	}

	for _, child := range children {
		if scrollOffset != 0 {
			child.Layout.Y -= float64(scrollOffset)
		}
		renderNode(child, canvas, childClip, effectiveFG, effectiveAttrs, debug, depth+1)
		if scrollOffset != 0 {
			child.Layout.Y += float64(scrollOffset)
		}
	}
}

func setClipped(canvas *Canvas, x, y int, c Cell, clip *clipRect) {
	if clip != nil {
		if x < clip.x || x >= clip.x+clip.w || y < clip.y || y >= clip.y+clip.h {
			return
		}
	}
	canvas.Set(x, y, c)
}

func fillRectClipped(canvas *Canvas, x, y, w, h int, bg Color, clip *clipRect) {
	if clip == nil {
		canvas.FillRect(x, y, w, h, bg)
		return
	}
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			if col >= clip.x && col < clip.x+clip.w && row >= clip.y && row < clip.y+clip.h {
				c := canvas.Get(col, row)
				c.BG = bg
				canvas.Set(col, row, c)
			}
		}
	}
}

func drawBorder(canvas *Canvas, x, y, w, h int, bs BorderStyle, p Paint, title string, clip *clipRect) {
	if w < 2 || h < 2 {
		return
	}
	chars := borderCharSets[bs]

	fg := p.BorderFG
	if fg.IsDefault() {
		fg = p.FG
	}
	bg := p.BG

	cell := func(r rune) Cell {
		return Cell{Rune: r, FG: fg, BG: bg}
	}

	// Corners.
	setClipped(canvas, x, y, cell(chars.TL), clip)
	setClipped(canvas, x+w-1, y, cell(chars.TR), clip)
	setClipped(canvas, x, y+h-1, cell(chars.BL), clip)
	setClipped(canvas, x+w-1, y+h-1, cell(chars.BR), clip)

	// Top edge with optional title.
	titleRunes := []rune(title)
	maxTitle := w - 6 // corners + h-char + space on each side
	if maxTitle < 0 {
		maxTitle = 0
	}
	if len(titleRunes) > maxTitle {
		titleRunes = titleRunes[:maxTitle]
	}
	titleStart := x + 2 // after corner + one horizontal char
	spaceAfter := titleStart + len(titleRunes) + 1
	for col := x + 1; col < x+w-1; col++ {
		if len(titleRunes) > 0 && col == titleStart {
			setClipped(canvas, col, y, Cell{Rune: ' ', FG: fg, BG: bg}, clip)
		} else if len(titleRunes) > 0 && col > titleStart && col <= titleStart+len(titleRunes) {
			setClipped(canvas, col, y, Cell{Rune: titleRunes[col-titleStart-1], FG: fg, BG: bg}, clip)
		} else if len(titleRunes) > 0 && col == spaceAfter {
			setClipped(canvas, col, y, Cell{Rune: ' ', FG: fg, BG: bg}, clip)
		} else {
			setClipped(canvas, col, y, cell(chars.H), clip)
		}
	}

	// Bottom edge.
	for col := x + 1; col < x+w-1; col++ {
		setClipped(canvas, col, y+h-1, cell(chars.H), clip)
	}

	// Vertical edges.
	for row := y + 1; row < y+h-1; row++ {
		setClipped(canvas, x, row, cell(chars.V), clip)
		setClipped(canvas, x+w-1, row, cell(chars.V), clip)
	}
}
