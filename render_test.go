package quill

import "testing"

func TestRenderTextASCII(t *testing.T) {
	node := NewText("hello")
	Compute(node, 80, 24)
	canvas := NewCanvas(80, 24)
	Render(node, canvas)

	// "hello" should be at (0,0)–(4,0)
	for i, ch := range "hello" {
		cell := canvas.Get(i, 0)
		if cell.Rune != ch {
			t.Errorf("cell(%d,0) = %q, want %q", i, cell.Rune, ch)
		}
	}
}

func TestRenderTextWide(t *testing.T) {
	node := NewText("A中B")
	Compute(node, 80, 24)
	canvas := NewCanvas(80, 24)
	Render(node, canvas)

	// 'A' at col 0 (width 1)
	if c := canvas.Get(0, 0); c.Rune != 'A' {
		t.Errorf("col 0 = %q, want 'A'", c.Rune)
	}
	// '中' at col 1 (width 2)
	if c := canvas.Get(1, 0); c.Rune != '中' {
		t.Errorf("col 1 = %q, want '中'", c.Rune)
	}
	// col 2 should be Wide continuation
	if c := canvas.Get(2, 0); !c.Wide {
		t.Error("col 2 should be Wide continuation cell")
	}
	// 'B' at col 3
	if c := canvas.Get(3, 0); c.Rune != 'B' {
		t.Errorf("col 3 = %q, want 'B'", c.Rune)
	}
}

func TestRenderProgressBar(t *testing.T) {
	bar := renderProgressBar(10, 0.5)
	if len([]rune(bar)) != 10 {
		t.Errorf("progress bar length = %d, want 10", len([]rune(bar)))
	}

	bar0 := renderProgressBar(10, 0.0)
	for _, r := range bar0 {
		if r != '░' {
			t.Errorf("0%% bar contains %q, want '░'", r)
			break
		}
	}

	bar100 := renderProgressBar(10, 1.0)
	for _, r := range bar100 {
		if r != '█' {
			t.Errorf("100%% bar contains %q, want '█'", r)
			break
		}
	}
}

func TestRenderBorder(t *testing.T) {
	s := DefaultStyle()
	s.Border = BorderRounded
	s.Width = Px(5)
	s.Height = Px(3)
	node := NewNode(s)

	Compute(node, 80, 24)
	canvas := NewCanvas(80, 24)
	Render(node, canvas)

	chars := borderCharSets[BorderRounded]
	if c := canvas.Get(0, 0); c.Rune != chars.TL {
		t.Errorf("top-left = %q, want %q", c.Rune, chars.TL)
	}
	if c := canvas.Get(4, 0); c.Rune != chars.TR {
		t.Errorf("top-right = %q, want %q", c.Rune, chars.TR)
	}
	if c := canvas.Get(0, 2); c.Rune != chars.BL {
		t.Errorf("bottom-left = %q, want %q", c.Rune, chars.BL)
	}
	if c := canvas.Get(4, 2); c.Rune != chars.BR {
		t.Errorf("bottom-right = %q, want %q", c.Rune, chars.BR)
	}
}

func TestRenderBackgroundFill(t *testing.T) {
	s := DefaultStyle()
	s.Width = Px(4)
	s.Height = Px(2)
	node := NewNode(s)
	node.Paint.BG = Red

	Compute(node, 80, 24)
	canvas := NewCanvas(80, 24)
	Render(node, canvas)

	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			if c := canvas.Get(x, y); c.BG != Red {
				t.Errorf("cell(%d,%d) BG = %v, want Red", x, y, c.BG)
			}
		}
	}
}

func TestNewTextMeasureWide(t *testing.T) {
	// "中文" is 4 cells wide
	node := NewText("中文")
	w, h := node.MeasureFunc(80, 24)
	if w != 4 {
		t.Errorf("width = %v, want 4", w)
	}
	if h != 1 {
		t.Errorf("height = %v, want 1", h)
	}
}

func TestNewTextMeasureWrap(t *testing.T) {
	node := NewText("abcdefghij") // 10 chars
	w, h := node.MeasureFunc(4, 24)
	if w != 4 {
		t.Errorf("width = %v, want 4", w)
	}
	if h < 3 {
		t.Errorf("height = %v, want >= 3", h)
	}
}
