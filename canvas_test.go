package quill

import "testing"

func TestCanvasSetGet(t *testing.T) {
	c := NewCanvas(10, 5)
	cell := Cell{Rune: 'X', FG: Red}
	c.Set(3, 2, cell)

	got := c.Get(3, 2)
	if got != cell {
		t.Errorf("Get(3,2) = %v, want %v", got, cell)
	}
}

func TestCanvasOutOfBounds(t *testing.T) {
	c := NewCanvas(5, 5)

	// Out-of-bounds Get returns emptyCell.
	if got := c.Get(-1, 0); got != emptyCell {
		t.Errorf("Get(-1,0) = %v, want emptyCell", got)
	}
	if got := c.Get(5, 0); got != emptyCell {
		t.Errorf("Get(5,0) = %v, want emptyCell", got)
	}

	// Out-of-bounds Set doesn't panic.
	c.Set(-1, 0, Cell{Rune: 'X'})
	c.Set(100, 100, Cell{Rune: 'X'})
}

func TestCanvasClear(t *testing.T) {
	c := NewCanvas(3, 3)
	c.Set(1, 1, Cell{Rune: 'A', FG: Green})
	c.Clear()

	if got := c.Get(1, 1); got != emptyCell {
		t.Errorf("after Clear, Get(1,1) = %v, want emptyCell", got)
	}
}

func TestCanvasFillRect(t *testing.T) {
	c := NewCanvas(10, 10)
	c.FillRect(2, 2, 3, 3, Blue)

	// Inside the rect.
	for y := 2; y < 5; y++ {
		for x := 2; x < 5; x++ {
			if got := c.Get(x, y); got.BG != Blue {
				t.Errorf("Get(%d,%d).BG = %v, want Blue", x, y, got.BG)
			}
		}
	}
	// Outside the rect.
	if got := c.Get(0, 0); got.BG != ColorDefault {
		t.Errorf("Get(0,0).BG = %v, want default", got.BG)
	}
}

func TestCanvasDiffIdentical(t *testing.T) {
	a := NewCanvas(5, 3)
	b := NewCanvas(5, 3)
	a.Set(1, 1, Cell{Rune: 'X'})
	b.Set(1, 1, Cell{Rune: 'X'})

	changed := false
	for i := range a.cells {
		if a.cells[i] != b.cells[i] {
			changed = true
			break
		}
	}
	if changed {
		t.Error("identical canvases should not report changes")
	}
}

func TestCanvasDiffChanged(t *testing.T) {
	a := NewCanvas(5, 3)
	b := NewCanvas(5, 3)
	a.Set(1, 1, Cell{Rune: 'X'})
	b.Set(1, 1, Cell{Rune: 'Y'})

	changed := false
	for i := range a.cells {
		if a.cells[i] != b.cells[i] {
			changed = true
			break
		}
	}
	if !changed {
		t.Error("different canvases should report changes")
	}
}

func TestSGR(t *testing.T) {
	// Default: just reset.
	s := sgr(ColorDefault, ColorDefault, CellAttrs{})
	if s != "\x1b[0m" {
		t.Errorf("default sgr = %q, want \\x1b[0m", s)
	}

	// Bold.
	s = sgr(ColorDefault, ColorDefault, CellAttrs{Bold: true})
	if s != "\x1b[0;1m" {
		t.Errorf("bold sgr = %q, want \\x1b[0;1m", s)
	}

	// ANSI foreground.
	s = sgr(ANSIColor(1), ColorDefault, CellAttrs{})
	if s != "\x1b[0;31m" {
		t.Errorf("red fg sgr = %q, want \\x1b[0;31m", s)
	}

	// RGB background.
	s = sgr(ColorDefault, RGBColor(255, 0, 128), CellAttrs{})
	if s != "\x1b[0;48;2;255;0;128m" {
		t.Errorf("rgb bg sgr = %q, want \\x1b[0;48;2;255;0;128m", s)
	}
}
