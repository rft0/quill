package quill

import (
	"math"
	"testing"
)

const eps = 0.001

func approx(a, b float64) bool {
	return math.Abs(a-b) < eps
}

func assertLayout(t *testing.T, label string, n *Node, x, y, w, h float64) {
	t.Helper()
	l := n.Layout
	if !approx(l.X, x) || !approx(l.Y, y) || !approx(l.Width, w) || !approx(l.Height, h) {
		t.Errorf("%s: expected (%.1f, %.1f, %.1f, %.1f), got (%.1f, %.1f, %.1f, %.1f)",
			label, x, y, w, h, l.X, l.Y, l.Width, l.Height)
	}
}

func TestSingleChildStretch(t *testing.T) {
	// A single child should stretch to fill the container (default align=stretch).
	root := NewNode(DefaultStyle())
	child := NewNode(DefaultStyle())
	root.AddChild(child)

	Compute(root, 80, 24)

	assertLayout(t, "root", root, 0, 0, 80, 24)
	// Row direction: child main=width gets 0 base (auto) but flex-shrink won't shrink below 0.
	// Cross=height stretches to 24.
	assertLayout(t, "child", child, 0, 0, 0, 24)
}

func TestFixedSizeChildren_Row(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)

	b := NewNode(DefaultStyle())
	b.Style.Width = Px(30)
	b.Style.Height = Px(10)

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 80, 24)

	assertLayout(t, "root", root, 0, 0, 80, 24)
	// Explicit Height wins over AlignStretch (spec-correct; only Auto cross dims stretch).
	assertLayout(t, "a", a, 0, 0, 20, 10)
	assertLayout(t, "b", b, 20, 0, 30, 10)
}

func TestFixedSizeChildren_Column(t *testing.T) {
	s := DefaultStyle()
	s.Direction = FlexColumn
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(5)

	b := NewNode(DefaultStyle())
	b.Style.Width = Px(30)
	b.Style.Height = Px(8)

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 80, 24)

	assertLayout(t, "root", root, 0, 0, 80, 24)
	// Column: main=height, cross=width. Explicit Width wins over AlignStretch.
	assertLayout(t, "a", a, 0, 0, 20, 5)
	assertLayout(t, "b", b, 0, 5, 30, 8)
}

func TestFlexGrow(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.FlexGrow = 1

	b := NewNode(DefaultStyle())
	b.Style.Width = Px(20)
	b.Style.FlexGrow = 2

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 80, 24)

	// Free space = 80 - 40 = 40. a gets 40/3 ≈ 13.33, b gets 80/3 ≈ 26.67.
	assertLayout(t, "a", a, 0, 0, 20+40.0/3, 24)
	assertLayout(t, "b", b, 20+40.0/3, 0, 20+80.0/3, 24)
}

func TestFlexShrink(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(50)
	a.Style.FlexShrink = 1

	b := NewNode(DefaultStyle())
	b.Style.Width = Px(50)
	b.Style.FlexShrink = 1

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 60, 24)

	// Total base = 100, container = 60. Overflow = 40.
	// Equal shrink factors and equal bases -> each shrinks by 20.
	assertLayout(t, "a", a, 0, 0, 30, 24)
	assertLayout(t, "b", b, 30, 0, 30, 24)
}

func TestJustifyCenter(t *testing.T) {
	s := DefaultStyle()
	s.Justify = JustifyCenter
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.FlexShrink = 0

	root.AddChild(a)

	Compute(root, 80, 24)

	// Centered: offset = (80 - 20) / 2 = 30.
	assertLayout(t, "a", a, 30, 0, 20, 24)
}

func TestJustifySpaceBetween(t *testing.T) {
	s := DefaultStyle()
	s.Justify = JustifySpaceBetween
	root := NewNode(s)

	for i := 0; i < 3; i++ {
		c := NewNode(DefaultStyle())
		c.Style.Width = Px(10)
		c.Style.FlexShrink = 0
		root.AddChild(c)
	}

	Compute(root, 80, 24)

	// space-between: first at 0, last at 80-10=70, middle at 35.
	// gap = (80 - 30) / 2 = 25.
	assertLayout(t, "child0", root.Children[0], 0, 0, 10, 24)
	assertLayout(t, "child1", root.Children[1], 35, 0, 10, 24)
	assertLayout(t, "child2", root.Children[2], 70, 0, 10, 24)
}

func TestJustifySpaceAround(t *testing.T) {
	s := DefaultStyle()
	s.Justify = JustifySpaceAround
	root := NewNode(s)

	for i := 0; i < 2; i++ {
		c := NewNode(DefaultStyle())
		c.Style.Width = Px(10)
		c.Style.FlexShrink = 0
		root.AddChild(c)
	}

	Compute(root, 80, 24)

	// space-around: gap = (80-20)/2 = 30. offset = 30/2 = 15.
	// child0 at 15, child1 at 15+10+30 = 55.
	assertLayout(t, "child0", root.Children[0], 15, 0, 10, 24)
	assertLayout(t, "child1", root.Children[1], 55, 0, 10, 24)
}

func TestAlignCenter(t *testing.T) {
	s := DefaultStyle()
	s.Align = AlignCenter
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)

	root.AddChild(a)

	Compute(root, 80, 24)

	// Cross axis centered: (24 - 10) / 2 = 7.
	assertLayout(t, "a", a, 0, 7, 20, 10)
}

func TestAlignFlexEnd(t *testing.T) {
	s := DefaultStyle()
	s.Align = AlignFlexEnd
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)

	root.AddChild(a)

	Compute(root, 80, 24)

	// Cross axis flex-end: 24 - 10 = 14.
	assertLayout(t, "a", a, 0, 14, 20, 10)
}

func TestPadding(t *testing.T) {
	s := DefaultStyle()
	s.Padding = UniformEdges(2)
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)

	root.AddChild(a)

	Compute(root, 80, 24)

	// Inner area: 76 x 20. Child starts at root.X+padding.Left, root.Y+padding.Top.
	// Explicit Height wins over AlignStretch.
	assertLayout(t, "a", a, 2, 2, 20, 10)
}

func TestPaddingExplicitHeight(t *testing.T) {
	// When a child has explicit height, stretch should NOT override it.
	s := DefaultStyle()
	s.Padding = UniformEdges(2)
	s.Align = AlignFlexStart // don't stretch
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)

	root.AddChild(a)

	Compute(root, 80, 24)

	assertLayout(t, "a", a, 2, 2, 20, 10)
}

func TestMargin(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.Height = Px(10)
	a.Style.Margin = Edges{Left: 5, Top: 3}

	root.AddChild(a)

	Compute(root, 80, 24)

	// Margin offsets position. Cross-axis: align=stretch, but height is explicit(10).
	// Actually with Stretch+Auto height, it would stretch. Height is Px(10) so cross stays 10.
	assertLayout(t, "a", a, 5, 3, 20, 10)
}

func TestTextNode(t *testing.T) {
	s := DefaultStyle()
	s.Direction = FlexColumn
	root := NewNode(s)

	txt := NewText("Hello, World!")
	root.AddChild(txt)

	Compute(root, 80, 24)

	// "Hello, World!" = 13 chars. Column: main=height, cross=width.
	// MeasureFunc returns (13, 1). Main = height = 1, cross = width = 13.
	// But align=stretch, width is auto -> stretches to 80.
	assertLayout(t, "txt", txt, 0, 0, 80, 1)
}

func TestNestedContainers(t *testing.T) {
	// Root (row) -> left (column, grow=1) -> [a, b]
	//            -> right (column, grow=1) -> [c]
	root := NewNode(DefaultStyle())

	leftStyle := DefaultStyle()
	leftStyle.Direction = FlexColumn
	leftStyle.FlexGrow = 1
	left := NewNode(leftStyle)

	rightStyle := DefaultStyle()
	rightStyle.Direction = FlexColumn
	rightStyle.FlexGrow = 1
	right := NewNode(rightStyle)

	a := NewNode(DefaultStyle())
	a.Style.Height = Px(5)

	b := NewNode(DefaultStyle())
	b.Style.Height = Px(5)

	c := NewNode(DefaultStyle())
	c.Style.Height = Px(10)

	left.AddChild(a)
	left.AddChild(b)
	right.AddChild(c)
	root.AddChild(left)
	root.AddChild(right)

	Compute(root, 80, 24)

	// Both left and right have flexGrow=1, base=0: each gets 40.
	assertLayout(t, "left", left, 0, 0, 40, 24)
	assertLayout(t, "right", right, 40, 0, 40, 24)

	// Inside left (column): a and b are 5 tall each, stretch to width 40.
	assertLayout(t, "a", a, 0, 0, 40, 5)
	assertLayout(t, "b", b, 0, 5, 40, 5)

	// Inside right (column): c is 10 tall, stretch to width 40.
	assertLayout(t, "c", c, 40, 0, 40, 10)
}

func TestFlexBasis(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.FlexBasis = Px(30)
	a.Style.FlexGrow = 1

	b := NewNode(DefaultStyle())
	b.Style.FlexBasis = Px(10)
	b.Style.FlexGrow = 1

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 80, 24)

	// Base: 30 + 10 = 40. Free = 40, split equally: a=50, b=30.
	assertLayout(t, "a", a, 0, 0, 50, 24)
	assertLayout(t, "b", b, 50, 0, 30, 24)
}

func TestPercentWidth(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Pct(50)
	a.Style.FlexShrink = 0

	root.AddChild(a)

	Compute(root, 80, 24)

	assertLayout(t, "a", a, 0, 0, 40, 24)
}

func TestMinMaxConstraints(t *testing.T) {
	root := NewNode(DefaultStyle())

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(10)
	a.Style.MinWidth = Px(20)
	a.Style.FlexShrink = 0

	b := NewNode(DefaultStyle())
	b.Style.Width = Px(100)
	b.Style.MaxWidth = Px(50)
	b.Style.FlexShrink = 0

	root.AddChild(a)
	root.AddChild(b)

	Compute(root, 200, 24)

	assertLayout(t, "a", a, 0, 0, 20, 24)  // min wins
	assertLayout(t, "b", b, 20, 0, 50, 24) // max wins
}

func TestTextNode_ColumnAlignCenter(t *testing.T) {
	// Regression: MeasureFunc was called with swapped (height, width) in
	// Column layout, producing a cross size of 1 instead of len(text).
	s := DefaultStyle()
	s.Direction = FlexColumn
	s.Align = AlignCenter
	root := NewNode(s)

	txt := NewText("Hello!")
	root.AddChild(txt)

	Compute(root, 80, 24)

	// "Hello!" = 6 chars. With AlignCenter, width = intrinsic 6, centered.
	assertLayout(t, "txt", txt, 37, 0, 6, 1)
}

func TestJustifyFlexEnd(t *testing.T) {
	s := DefaultStyle()
	s.Justify = JustifyFlexEnd
	root := NewNode(s)

	a := NewNode(DefaultStyle())
	a.Style.Width = Px(20)
	a.Style.FlexShrink = 0

	root.AddChild(a)

	Compute(root, 80, 24)

	assertLayout(t, "a", a, 60, 0, 20, 24)
}
