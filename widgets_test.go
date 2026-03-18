package quill

import "testing"

func TestSelectUpdate(t *testing.T) {
	s := &SelectState{
		Options:  []string{"A", "B", "C"},
		Selected: 0,
		Focused:  true,
	}

	// Move down with arrow.
	s.Update(KeyMsg{Type: KeyDown})
	if s.Selected != 1 {
		t.Errorf("after down: selected = %d, want 1", s.Selected)
	}

	// Move down with j.
	s.Update(KeyMsg{Type: KeyRune, Rune: 'j'})
	if s.Selected != 2 {
		t.Errorf("after j: selected = %d, want 2", s.Selected)
	}

	// Can't go past end.
	s.Update(KeyMsg{Type: KeyDown})
	if s.Selected != 2 {
		t.Errorf("past end: selected = %d, want 2", s.Selected)
	}

	// Move up.
	s.Update(KeyMsg{Type: KeyUp})
	if s.Selected != 1 {
		t.Errorf("after up: selected = %d, want 1", s.Selected)
	}

	// Move up with k.
	s.Update(KeyMsg{Type: KeyRune, Rune: 'k'})
	if s.Selected != 0 {
		t.Errorf("after k: selected = %d, want 0", s.Selected)
	}

	// Can't go past start.
	s.Update(KeyMsg{Type: KeyUp})
	if s.Selected != 0 {
		t.Errorf("past start: selected = %d, want 0", s.Selected)
	}
}

func TestSelectUnfocused(t *testing.T) {
	s := &SelectState{
		Options:  []string{"A", "B"},
		Selected: 0,
		Focused:  false,
	}
	s.Update(KeyMsg{Type: KeyDown})
	if s.Selected != 0 {
		t.Errorf("unfocused: selected = %d, want 0", s.Selected)
	}
}

func TestCheckboxToggle(t *testing.T) {
	s := &CheckboxState{Checked: false, Focused: true}

	s.Update(KeyMsg{Type: KeySpace})
	if !s.Checked {
		t.Error("after space: should be checked")
	}

	s.Update(KeyMsg{Type: KeyEnter})
	if s.Checked {
		t.Error("after enter: should be unchecked")
	}
}

func TestCheckboxUnfocused(t *testing.T) {
	s := &CheckboxState{Checked: false, Focused: false}
	s.Update(KeyMsg{Type: KeySpace})
	if s.Checked {
		t.Error("unfocused: should not toggle")
	}
}

func TestInputUpdate(t *testing.T) {
	s := &InputState{Focused: true}

	// Type "hi".
	s.Update(KeyMsg{Type: KeyRune, Rune: 'h'})
	s.Update(KeyMsg{Type: KeyRune, Rune: 'i'})
	if s.Value != "hi" {
		t.Errorf("value = %q, want %q", s.Value, "hi")
	}

	// Backspace.
	s.Update(KeyMsg{Type: KeyBackspace})
	if s.Value != "h" {
		t.Errorf("after backspace: value = %q, want %q", s.Value, "h")
	}

	// Left arrow then type.
	s.Update(KeyMsg{Type: KeyLeft})
	s.Update(KeyMsg{Type: KeyRune, Rune: 'a'})
	if s.Value != "ah" {
		t.Errorf("after insert: value = %q, want %q", s.Value, "ah")
	}
}

func TestInputHomeEnd(t *testing.T) {
	s := &InputState{Value: "abc", Cursor: 3, Focused: true}

	s.Update(KeyMsg{Type: KeyHome})
	if s.Cursor != 0 {
		t.Errorf("after Home: cursor = %d, want 0", s.Cursor)
	}

	s.Update(KeyMsg{Type: KeyEnd})
	if s.Cursor != 3 {
		t.Errorf("after End: cursor = %d, want 3", s.Cursor)
	}
}

func TestInputCtrlU(t *testing.T) {
	s := &InputState{Value: "abcdef", Cursor: 3, Focused: true}

	s.Update(KeyMsg{Type: KeyCtrlU})
	if s.Value != "def" {
		t.Errorf("after Ctrl+U: value = %q, want %q", s.Value, "def")
	}
	if s.Cursor != 0 {
		t.Errorf("after Ctrl+U: cursor = %d, want 0", s.Cursor)
	}
}

func TestInputCtrlK(t *testing.T) {
	s := &InputState{Value: "abcdef", Cursor: 3, Focused: true}

	s.Update(KeyMsg{Type: KeyCtrlK})
	if s.Value != "abc" {
		t.Errorf("after Ctrl+K: value = %q, want %q", s.Value, "abc")
	}
}

func TestInputDelete(t *testing.T) {
	s := &InputState{Value: "abc", Cursor: 1, Focused: true}

	s.Update(KeyMsg{Type: KeyDelete})
	if s.Value != "ac" {
		t.Errorf("after delete: value = %q, want %q", s.Value, "ac")
	}
}

func TestInputUnfocused(t *testing.T) {
	s := &InputState{Value: "hi", Focused: false}
	s.Update(KeyMsg{Type: KeyRune, Rune: 'x'})
	if s.Value != "hi" {
		t.Errorf("unfocused: value = %q, want %q", s.Value, "hi")
	}
}

func TestFocusGroupCycle(t *testing.T) {
	a := &InputState{}
	b := &InputState{}
	c := &InputState{}
	fg := NewFocusGroup(a, b, c)

	if !a.Focused {
		t.Error("first input should be focused")
	}

	fg.Next()
	if a.Focused || !b.Focused || c.Focused {
		t.Error("after Next: b should be focused")
	}

	fg.Next()
	if !c.Focused {
		t.Error("after Next×2: c should be focused")
	}

	fg.Next()
	if !a.Focused {
		t.Error("after Next×3: should wrap to a")
	}

	fg.Prev()
	if !c.Focused {
		t.Error("after Prev: should wrap to c")
	}
}

func TestScrollState(t *testing.T) {
	s := &ScrollState{Offset: 0, ContentHeight: 50, ViewHeight: 10}

	s.ScrollDown(5)
	if s.Offset != 5 {
		t.Errorf("after ScrollDown(5): offset = %d, want 5", s.Offset)
	}

	s.ScrollUp(3)
	if s.Offset != 2 {
		t.Errorf("after ScrollUp(3): offset = %d, want 2", s.Offset)
	}

	// Can't scroll past 0.
	s.ScrollUp(100)
	if s.Offset != 0 {
		t.Errorf("scroll up past 0: offset = %d, want 0", s.Offset)
	}

	// Can't scroll past max.
	s.ScrollDown(999)
	max := s.ContentHeight - s.ViewHeight
	if s.Offset != max {
		t.Errorf("scroll down past max: offset = %d, want %d", s.Offset, max)
	}

	// PageUp/PageDown.
	s.Offset = 20
	s.PageDown()
	if s.Offset != 30 {
		t.Errorf("after PageDown: offset = %d, want 30", s.Offset)
	}
	s.PageUp()
	if s.Offset != 20 {
		t.Errorf("after PageUp: offset = %d, want 20", s.Offset)
	}
}

func TestRenderProgressBarWidths(t *testing.T) {
	// Zero width.
	if bar := renderProgressBar(0, 0.5); bar != "" {
		t.Errorf("width=0: got %q, want empty", bar)
	}

	// Negative width.
	if bar := renderProgressBar(-1, 0.5); bar != "" {
		t.Errorf("width=-1: got %q, want empty", bar)
	}

	// Width 1, 100%.
	bar := renderProgressBar(1, 1.0)
	if bar != "█" {
		t.Errorf("width=1 100%%: got %q, want █", bar)
	}
}
