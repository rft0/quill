package quill

import (
	"strings"
	"testing"
)

func TestRenderToString(t *testing.T) {
	root := Box(FlexRow, Text("hello"), Text(" world"))
	s := RenderToString(root, 20)
	if !strings.Contains(s, "hello") || !strings.Contains(s, "world") {
		t.Errorf("output = %q, want to contain 'hello' and 'world'", s)
	}
}

func TestRenderToStringNilChildren(t *testing.T) {
	// Nil children should be silently skipped.
	root := Box(FlexRow, Text("a"), nil, Text("b"))
	s := RenderToString(root, 20)
	if !strings.Contains(s, "a") || !strings.Contains(s, "b") {
		t.Errorf("output = %q, want 'a' and 'b'", s)
	}
}

func TestRenderToStringEllipsis(t *testing.T) {
	root := Box(Width(Px(5)), Text("hello world", Ellipsis))
	s := RenderToString(root, 5)
	if !strings.Contains(s, "\u2026") {
		t.Errorf("output = %q, should contain ellipsis", s)
	}
	// Should be at most 5 chars wide per line.
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		if stringWidth(line) > 5 {
			t.Errorf("line %q is wider than 5", line)
		}
	}
}

func TestRenderToStringClip(t *testing.T) {
	root := Box(Width(Px(5)), Text("hello world", ClipText))
	s := RenderToString(root, 5)
	lines := strings.Split(s, "\n")
	// Clipped text should be single line, max 5 chars.
	if len(lines) > 1 {
		t.Errorf("clipped text should be 1 line, got %d", len(lines))
	}
	if stringWidth(lines[0]) > 5 {
		t.Errorf("clipped line %q wider than 5", lines[0])
	}
}

func TestColorInheritance(t *testing.T) {
	root := Box(TextColor(Red), Text("hi"))
	canvas := RenderToCanvas(root, 20)
	cell := canvas.Get(0, 0)
	if cell.FG != Red {
		t.Errorf("inherited FG = %v, want Red", cell.FG)
	}
}

func TestColorInheritanceOverride(t *testing.T) {
	root := Box(TextColor(Red), Text("hi", TextColor(Blue)))
	canvas := RenderToCanvas(root, 20)
	cell := canvas.Get(0, 0)
	if cell.FG != Blue {
		t.Errorf("overridden FG = %v, want Blue", cell.FG)
	}
}

func TestTestAppSendKey(t *testing.T) {
	counter := 0
	comp := func(ctx *Context) *Node {
		count := UseState(ctx, 0)
		OnKey(ctx, func(key KeyMsg) {
			if key.Rune == '+' {
				count.Set(count.Get() + 1)
			}
		})
		counter = count.Get()
		return Text("ok")
	}

	app := NewTestApp(comp, 20)
	if counter != 0 {
		t.Fatalf("initial counter = %d, want 0", counter)
	}

	app.SendKey(KeyMsg{Type: KeyRune, Rune: '+'})
	if counter != 1 {
		t.Errorf("after key: counter = %d, want 1", counter)
	}
}

func TestSubContext(t *testing.T) {
	// Two sub-contexts should have isolated state.
	comp := func(ctx *Context) *Node {
		sub1 := ctx.SubContext("a")
		s1 := UseState(sub1, 10)

		sub2 := ctx.SubContext("b")
		s2 := UseState(sub2, 20)

		_ = s1.Get()
		_ = s2.Get()
		return Text("ok")
	}

	app := NewTestApp(comp, 20)
	_ = app.Output()
}

func TestSubContextPreservesState(t *testing.T) {
	var val1, val2 int
	comp := func(ctx *Context) *Node {
		sub1 := ctx.SubContext("x")
		s1 := UseState(sub1, 100)

		sub2 := ctx.SubContext("y")
		s2 := UseState(sub2, 200)

		val1 = s1.Get()
		val2 = s2.Get()

		// Modify state on first key press.
		OnKey(ctx, func(key KeyMsg) {
			s1.Set(111)
		})
		return Text("ok")
	}

	app := NewTestApp(comp, 20)
	if val1 != 100 || val2 != 200 {
		t.Fatalf("initial: val1=%d val2=%d", val1, val2)
	}

	app.SendKey(KeyMsg{Type: KeyRune, Rune: 'x'})
	if val1 != 111 {
		t.Errorf("after key: val1=%d, want 111", val1)
	}
	if val2 != 200 {
		t.Errorf("after key: val2=%d, want 200 (unchanged)", val2)
	}
}

func TestAbsolutePositioning(t *testing.T) {
	root := NewNode(DefaultStyle())
	root.Layout.Width = 80
	root.Layout.Height = 24

	child := NewNode(DefaultStyle())
	child.Style.Position = PositionAbsolute
	child.Style.Left = Px(10)
	child.Style.Top = Px(5)
	child.Style.Width = Px(20)
	child.Style.Height = Px(3)
	root.AddChild(child)

	Compute(root, 80, 24)
	if child.Layout.X != 10 {
		t.Errorf("X = %v, want 10", child.Layout.X)
	}
	if child.Layout.Y != 5 {
		t.Errorf("Y = %v, want 5", child.Layout.Y)
	}
	if child.Layout.Width != 20 {
		t.Errorf("Width = %v, want 20", child.Layout.Width)
	}
}

func TestZIndexRenderOrder(t *testing.T) {
	// Two overlapping boxes, higher z-index should render on top.
	root := NewNode(DefaultStyle())
	root.Style.Width = Px(10)
	root.Style.Height = Px(3)

	bg := NewText("AAAAAAAAAA")
	bg.Style.Position = PositionAbsolute
	bg.Style.Width = Px(10)
	bg.Style.ZIndex = 0
	root.AddChild(bg)

	fg := NewText("BB")
	fg.Style.Position = PositionAbsolute
	fg.Style.Width = Px(2)
	fg.Style.ZIndex = 1
	root.AddChild(fg)

	Compute(root, 10, 3)
	canvas := NewCanvas(10, 3)
	Render(root, canvas)

	// "BB" (z=1) should overwrite first 2 cells of "AA..." (z=0)
	if c := canvas.Get(0, 0); c.Rune != 'B' {
		t.Errorf("cell(0,0) = %q, want 'B' (higher z-index)", c.Rune)
	}
	if c := canvas.Get(2, 0); c.Rune != 'A' {
		t.Errorf("cell(2,0) = %q, want 'A'", c.Rune)
	}
}

func TestShiftTab(t *testing.T) {
	msgs := parseInput([]byte("\x1b[Z"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyShiftTab {
		t.Errorf("got %v, want KeyShiftTab", msgs[0])
	}
}

func TestKeyMsgString(t *testing.T) {
	tests := []struct {
		msg  KeyMsg
		want string
	}{
		{KeyMsg{Type: KeyRune, Rune: 'a'}, "a"},
		{KeyMsg{Type: KeyRune, Rune: 'x', Alt: true}, "Alt+x"},
		{KeyMsg{Type: KeyEnter}, "Enter"},
		{KeyMsg{Type: KeyCtrlC}, "Ctrl+C"},
		{KeyMsg{Type: KeyShiftTab}, "Shift+Tab"},
		{KeyMsg{Type: KeyF1}, "F1"},
	}
	for _, tt := range tests {
		if got := tt.msg.String(); got != tt.want {
			t.Errorf("%+v.String() = %q, want %q", tt.msg, got, tt.want)
		}
	}
}

func TestMouseMsgString(t *testing.T) {
	m := MouseMsg{Type: MouseLeft, X: 5, Y: 10}
	s := m.String()
	if s != "MouseLeft(5,10)" {
		t.Errorf("got %q, want %q", s, "MouseLeft(5,10)")
	}
}

func TestInputOnSubmit(t *testing.T) {
	submitted := ""
	s := &InputState{Value: "hello", Cursor: 5, Focused: true, OnSubmit: func(v string) {
		submitted = v
	}}

	s.Update(KeyMsg{Type: KeyEnter})
	if submitted != "hello" {
		t.Errorf("submitted = %q, want %q", submitted, "hello")
	}
}

func TestInputVisualCol(t *testing.T) {
	s := &InputState{Value: "a中b", Cursor: 4} // after "a中"
	if col := s.VisualCol(); col != 3 {
		t.Errorf("VisualCol = %d, want 3 (1 for 'a' + 2 for '中')", col)
	}
}

func TestNilChildren(t *testing.T) {
	// Should not panic.
	root := Box(FlexRow, Text("a"), nil, Text("b"), nil)
	if len(root.Children) != 2 {
		t.Errorf("children = %d, want 2", len(root.Children))
	}
}
