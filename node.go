package quill

// LayoutResult holds the computed absolute position and size of a node
// after the layout pass.
type LayoutResult struct {
	X, Y          float64
	Width, Height float64
}

// Node is a single element in the layout tree. It can be a container
// (with children) or a leaf (with text content and intrinsic size).
type Node struct {
	Style    Style
	Paint    Paint
	Children []*Node
	Parent   *Node

	// Text is the content for leaf nodes. The layout engine uses
	// MeasureFunc to determine its intrinsic size.
	Text string

	// MeasureFunc is called on leaf nodes (no children) to determine
	// their intrinsic content size given available space. If nil and
	// the node has Text, a default measure based on string length is used.
	MeasureFunc func(availWidth, availHeight float64) (width, height float64)

	// TextOverflow controls wrapping / truncation for text content.
	TextOverflow TextOverflow

	// showCursor and cursorPos are set by TextInput to render a cursor
	// as a reversed character at the given rune index.
	showCursor bool
	cursorPos  int

	// isProgress marks this node as a progress bar, rendered dynamically
	// based on available width. progressValue is the fill ratio (0–1).
	isProgress    bool
	progressValue float64

	// scrollState is set by ScrollView for clipped scrolling.
	scrollState *ScrollState

	// borderTitle is optional text rendered on the top border edge.
	borderTitle string

	// debug causes colored outlines to be drawn around this node
	// and all descendants, for visualizing the layout tree.
	debug bool

	// Layout is populated after Compute().
	Layout LayoutResult
}

// NewNode creates a container node with the given style and children.
func NewNode(style Style, children ...*Node) *Node {
	n := &Node{Style: style}
	for _, c := range children {
		if c == nil {
			continue
		}
		c.Parent = n
		n.Children = append(n.Children, c)
	}
	return n
}

// NewText creates a leaf node displaying text. By default it measures
// as len(text) wide and 1 tall (single terminal line).
func NewText(text string) *Node {
	s := DefaultStyle()
	s.FlexShrink = 0
	n := &Node{Style: s, Text: text}
	n.MeasureFunc = func(availWidth, availHeight float64) (float64, float64) {
		w := float64(stringWidth(text))
		// Ellipsis and clip modes never wrap — always single line.
		if n.TextOverflow == TextOverflowEllipsis || n.TextOverflow == TextOverflowClip {
			if availWidth > 0 && w > availWidth {
				return availWidth, 1
			}
			return w, 1
		}
		if availWidth > 0 && w > availWidth {
			lines := (w + availWidth - 1) / availWidth
			if lines < 1 {
				lines = 1
			}
			return availWidth, lines
		}
		return w, 1
	}
	return n
}

// AddChild appends a child node.
func (n *Node) AddChild(child *Node) *Node {
	if child == nil {
		return n
	}
	child.Parent = n
	n.Children = append(n.Children, child)
	return n
}

// IsLeaf returns true if the node has no children.
func (n *Node) IsLeaf() bool {
	return len(n.Children) == 0
}

// ContentBox returns the available inner dimensions after subtracting
// padding and border.
func (n *Node) ContentBox() (width, height float64) {
	bw := n.Style.BorderWidth()
	return n.Layout.Width - n.Style.Padding.Horizontal() - bw*2,
		n.Layout.Height - n.Style.Padding.Vertical() - bw*2
}
