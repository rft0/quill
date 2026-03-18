package quill

import (
	"fmt"
	"strings"
)

// Color represents a terminal color. The zero value means "terminal default".
type Color struct {
	mode uint8 // 0=default, 1=ANSI 16, 2=RGB
	r, g, b uint8
}

// ColorDefault is the terminal's default color (no explicit color set).
var ColorDefault = Color{}

// ANSIColor returns one of the 16 standard terminal colors (0–15).
func ANSIColor(n uint8) Color { return Color{mode: 1, r: n} }

// RGBColor returns a 24-bit color.
func RGBColor(r, g, b uint8) Color { return Color{mode: 2, r: r, g: g, b: b} }

// IsDefault reports whether c is the terminal default color.
func (c Color) IsDefault() bool { return c.mode == 0 }

// RGB returns the red, green, and blue components. ANSI colors return
// approximate RGB values; default color returns (0,0,0).
func (c Color) RGB() (r, g, b uint8) {
	if c.mode == 2 {
		return c.r, c.g, c.b
	}
	if c.mode == 1 {
		return ansiToRGB(c.r)
	}
	return 0, 0, 0
}

// LerpColor linearly interpolates between two colors. t is clamped to [0,1].
// Both colors are converted to RGB for interpolation.
func LerpColor(from, to Color, t float64) Color {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}
	r1, g1, b1 := from.RGB()
	r2, g2, b2 := to.RGB()
	return RGBColor(
		uint8(float64(r1)+t*float64(int(r2)-int(r1))),
		uint8(float64(g1)+t*float64(int(g2)-int(g1))),
		uint8(float64(b1)+t*float64(int(b2)-int(b1))),
	)
}

// Gradient returns a slice of n colors interpolated between from and to.
// Returns nil if n < 1, from if n == 1.
func Gradient(from, to Color, n int) []Color {
	if n < 1 {
		return nil
	}
	if n == 1 {
		return []Color{from}
	}
	colors := make([]Color, n)
	for i := range colors {
		colors[i] = LerpColor(from, to, float64(i)/float64(n-1))
	}
	return colors
}

// ansiToRGB returns approximate RGB values for the 16 standard ANSI colors.
func ansiToRGB(idx uint8) (r, g, b uint8) {
	table := [16][3]uint8{
		{0, 0, 0},       // 0 black
		{170, 0, 0},     // 1 red
		{0, 170, 0},     // 2 green
		{170, 85, 0},    // 3 yellow/brown
		{0, 0, 170},     // 4 blue
		{170, 0, 170},   // 5 magenta
		{0, 170, 170},   // 6 cyan
		{170, 170, 170}, // 7 white
		{85, 85, 85},    // 8 bright black (gray)
		{255, 85, 85},   // 9 bright red
		{85, 255, 85},   // 10 bright green
		{255, 255, 85},  // 11 bright yellow
		{85, 85, 255},   // 12 bright blue
		{255, 85, 255},  // 13 bright magenta
		{85, 255, 255},  // 14 bright cyan
		{255, 255, 255}, // 15 bright white
	}
	if int(idx) < len(table) {
		return table[idx][0], table[idx][1], table[idx][2]
	}
	return 0, 0, 0
}

// CellAttrs holds text rendering attributes.
type CellAttrs struct {
	Bold          bool
	Italic        bool
	Underline     bool
	Dim           bool
	Strikethrough bool
	Reverse       bool
}

// Cell is a single terminal cell.
type Cell struct {
	Rune  rune
	FG    Color
	BG    Color
	Attrs CellAttrs
	Wide  bool // continuation cell for wide (2-cell) characters; skip during output
}

var emptyCell = Cell{Rune: ' '}

// Canvas is a 2D grid of terminal cells.
type Canvas struct {
	Width, Height int
	cells         []Cell
}

// NewCanvas creates a canvas of the given dimensions, filled with empty cells.
func NewCanvas(w, h int) *Canvas {
	cells := make([]Cell, w*h)
	for i := range cells {
		cells[i] = emptyCell
	}
	return &Canvas{Width: w, Height: h, cells: cells}
}

func (c *Canvas) inBounds(x, y int) bool {
	return x >= 0 && x < c.Width && y >= 0 && y < c.Height
}

// Set writes a cell at (x, y), silently ignoring out-of-bounds positions.
func (c *Canvas) Set(x, y int, cell Cell) {
	if c.inBounds(x, y) {
		c.cells[y*c.Width+x] = cell
	}
}

// Get returns the cell at (x, y), or emptyCell if out of bounds.
func (c *Canvas) Get(x, y int) Cell {
	if c.inBounds(x, y) {
		return c.cells[y*c.Width+x]
	}
	return emptyCell
}

// Clear resets every cell to the empty/default state.
func (c *Canvas) Clear() {
	for i := range c.cells {
		c.cells[i] = emptyCell
	}
}

// FillRect fills a rectangular region with a background color.
func (c *Canvas) FillRect(x, y, w, h int, bg Color) {
	for row := y; row < y+h; row++ {
		for col := x; col < x+w; col++ {
			cell := c.Get(col, row)
			cell.BG = bg
			c.Set(col, row, cell)
		}
	}
}

// sgr builds a Select Graphic Rendition escape sequence for the given style.
// It always resets first so callers do not need to track cumulative state.
func sgr(fg, bg Color, attrs CellAttrs) string {
	var b strings.Builder
	b.WriteString("\x1b[0") // reset

	if attrs.Bold {
		b.WriteString(";1")
	}
	if attrs.Dim {
		b.WriteString(";2")
	}
	if attrs.Italic {
		b.WriteString(";3")
	}
	if attrs.Underline {
		b.WriteString(";4")
	}
	if attrs.Reverse {
		b.WriteString(";7")
	}
	if attrs.Strikethrough {
		b.WriteString(";9")
	}

	// Foreground color.
	switch fg.mode {
	case 1: // ANSI 16
		if fg.r < 8 {
			fmt.Fprintf(&b, ";%d", 30+fg.r)
		} else {
			fmt.Fprintf(&b, ";%d", 82+fg.r) // bright: 90–97
		}
	case 2: // 24-bit RGB
		fmt.Fprintf(&b, ";38;2;%d;%d;%d", fg.r, fg.g, fg.b)
	}

	// Background color.
	switch bg.mode {
	case 1: // ANSI 16
		if bg.r < 8 {
			fmt.Fprintf(&b, ";%d", 40+bg.r)
		} else {
			fmt.Fprintf(&b, ";%d", 92+bg.r) // bright: 100–107
		}
	case 2: // 24-bit RGB
		fmt.Fprintf(&b, ";48;2;%d;%d;%d", bg.r, bg.g, bg.b)
	}

	b.WriteString("m")
	return b.String()
}
