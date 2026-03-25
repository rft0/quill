package quill

// prop is implemented by anything that can configure a Node.
type prop interface {
	apply(n *Node)
}

// Style enums double as props — pass them directly to Box().
func (d FlexDirection) apply(n *Node)  { n.Style.Direction = d }
func (j JustifyContent) apply(n *Node) { n.Style.Justify = j }
func (a AlignItems) apply(n *Node)     { n.Style.Align = a }
func (a AlignSelf) apply(n *Node)      { n.Style.AlignSelf = a }
func (b BorderStyle) apply(n *Node)    { n.Style.Border = b }
func (w FlexWrap) apply(n *Node)       { n.Style.Wrap = w }

// toDimension converts a value to a Dimension. Accepts Dimension, int, or float64.
// Bare numbers are treated as Px values.
func toDimension(v any) Dimension {
	switch d := v.(type) {
	case Dimension:
		return d
	case int:
		return Px(float64(d))
	case float64:
		return Px(d)
	default:
		return AutoDim()
	}
}

// --- Layout props ---

type growProp float64

func (g growProp) apply(n *Node) { n.Style.FlexGrow = float64(g) }

// Grow sets the flex-grow factor. A value of 1 makes the node expand to fill
// available space along the main axis.
func Grow(v float64) growProp { return growProp(v) }

type shrinkProp float64

func (s shrinkProp) apply(n *Node) { n.Style.FlexShrink = float64(s) }

// Shrink sets the flex-shrink factor (default 1). Set to 0 to prevent a node
// from shrinking below its basis.
func Shrink(v float64) shrinkProp { return shrinkProp(v) }

type basisProp Dimension

func (b basisProp) apply(n *Node) { n.Style.FlexBasis = Dimension(b) }

// Basis sets the flex-basis (initial main-axis size before grow/shrink).
// Accepts a [Dimension], int, or float64 (bare numbers default to [Px]).
func Basis(v any) basisProp { return basisProp(toDimension(v)) }

type widthProp Dimension

func (w widthProp) apply(n *Node) { n.Style.Width = Dimension(w) }

// Width sets the node's width. Pass a number for cells or [Pct] for percentage.
func Width(v any) widthProp { return widthProp(toDimension(v)) }

type heightProp Dimension

func (h heightProp) apply(n *Node) { n.Style.Height = Dimension(h) }

// Height sets the node's height. Pass a number for cells or [Pct] for percentage.
func Height(v any) heightProp { return heightProp(toDimension(v)) }

type minWidthProp Dimension

func (m minWidthProp) apply(n *Node) { n.Style.MinWidth = Dimension(m) }

// MinWidth sets the minimum width constraint.
func MinWidth(v any) minWidthProp { return minWidthProp(toDimension(v)) }

type minHeightProp Dimension

func (m minHeightProp) apply(n *Node) { n.Style.MinHeight = Dimension(m) }

// MinHeight sets the minimum height constraint.
func MinHeight(v any) minHeightProp { return minHeightProp(toDimension(v)) }

type maxWidthProp Dimension

func (m maxWidthProp) apply(n *Node) { n.Style.MaxWidth = Dimension(m) }

// MaxWidth sets the maximum width constraint.
func MaxWidth(v any) maxWidthProp { return maxWidthProp(toDimension(v)) }

type maxHeightProp Dimension

func (m maxHeightProp) apply(n *Node) { n.Style.MaxHeight = Dimension(m) }

// MaxHeight sets the maximum height constraint.
func MaxHeight(v any) maxHeightProp { return maxHeightProp(toDimension(v)) }

type gapProp float64

func (g gapProp) apply(n *Node) { n.Style.Gap = float64(g) }

// Gap sets the spacing (in cells) between children along the main axis.
func Gap(v float64) gapProp { return gapProp(v) }

// --- Padding props ---

type padProp Edges

func (p padProp) apply(n *Node) { n.Style.Padding = Edges(p) }

// Padding sets equal padding on all four sides.
func Padding(v float64) padProp { return padProp(UniformEdges(v)) }

// PadX sets left and right padding.
func PadX(v float64) padProp { return padProp(Edges{Left: v, Right: v}) }

// PadY sets top and bottom padding.
func PadY(v float64) padProp { return padProp(Edges{Top: v, Bottom: v}) }

// PadXY sets horizontal (x) and vertical (y) padding.
func PadXY(x, y float64) padProp { return padProp(Edges{Top: y, Right: x, Bottom: y, Left: x}) }

type padTopProp float64

func (p padTopProp) apply(n *Node) { n.Style.Padding.Top = float64(p) }

// PadTop sets the top padding.
func PadTop(v float64) padTopProp { return padTopProp(v) }

type padRightProp float64

func (p padRightProp) apply(n *Node) { n.Style.Padding.Right = float64(p) }

// PadRight sets the right padding.
func PadRight(v float64) padRightProp { return padRightProp(v) }

type padBottomProp float64

func (p padBottomProp) apply(n *Node) { n.Style.Padding.Bottom = float64(p) }

// PadBottom sets the bottom padding.
func PadBottom(v float64) padBottomProp { return padBottomProp(v) }

type padLeftProp float64

func (p padLeftProp) apply(n *Node) { n.Style.Padding.Left = float64(p) }

// PadLeft sets the left padding.
func PadLeft(v float64) padLeftProp { return padLeftProp(v) }

// --- Margin props ---

type marginProp Edges

func (m marginProp) apply(n *Node) { n.Style.Margin = Edges(m) }

// Margin sets equal margin on all four sides.
func Margin(v float64) marginProp { return marginProp(UniformEdges(v)) }

// MarginX sets left and right margin.
func MarginX(v float64) marginProp { return marginProp(Edges{Left: v, Right: v}) }

// MarginY sets top and bottom margin.
func MarginY(v float64) marginProp { return marginProp(Edges{Top: v, Bottom: v}) }

// MarginXY sets horizontal (x) and vertical (y) margin.
func MarginXY(x, y float64) marginProp { return marginProp(Edges{Top: y, Right: x, Bottom: y, Left: x}) }

type marginTopProp float64

func (m marginTopProp) apply(n *Node) { n.Style.Margin.Top = float64(m) }

// MarginTop sets the top margin.
func MarginTop(v float64) marginTopProp { return marginTopProp(v) }

type marginRightProp float64

func (m marginRightProp) apply(n *Node) { n.Style.Margin.Right = float64(m) }

// MarginRight sets the right margin.
func MarginRight(v float64) marginRightProp { return marginRightProp(v) }

type marginBottomProp float64

func (m marginBottomProp) apply(n *Node) { n.Style.Margin.Bottom = float64(m) }

// MarginBottom sets the bottom margin.
func MarginBottom(v float64) marginBottomProp { return marginBottomProp(v) }

type marginLeftProp float64

func (m marginLeftProp) apply(n *Node) { n.Style.Margin.Left = float64(m) }

// MarginLeft sets the left margin.
func MarginLeft(v float64) marginLeftProp { return marginLeftProp(v) }

// --- Paint props ---

type textColorProp struct{ c Color }

func (p textColorProp) apply(n *Node) { n.Paint.FG = p.c }

// TextColor sets the foreground text color.
func TextColor(c Color) textColorProp { return textColorProp{c} }

type backgroundColorProp struct{ c Color }

func (p backgroundColorProp) apply(n *Node) { n.Paint.BG = p.c }

// BackgroundColor sets the background color.
func BackgroundColor(c Color) backgroundColorProp { return backgroundColorProp{c} }

type borderColorProp struct{ c Color }

func (p borderColorProp) apply(n *Node) { n.Paint.BorderFG = p.c }

// BorderColor sets the color of the node's border.
func BorderColor(c Color) borderColorProp { return borderColorProp{c} }

type titleProp struct{ t string }

func (p titleProp) apply(n *Node) { n.borderTitle = p.t }

// Title sets a text label rendered on the top border of a box.
func Title(t string) titleProp { return titleProp{t} }

type boldProp struct{}

func (boldProp) apply(n *Node) { n.Paint.Bold = true }

// Bold renders text in bold.
var Bold prop = boldProp{}

type italicProp struct{}

func (italicProp) apply(n *Node) { n.Paint.Italic = true }

// Italic renders text in italic (terminal support varies).
var Italic prop = italicProp{}

type underlineProp struct{}

func (underlineProp) apply(n *Node) { n.Paint.Underline = true }

// Underline renders text with an underline.
var Underline prop = underlineProp{}

type dimProp struct{}

func (dimProp) apply(n *Node) { n.Paint.Dim = true }

// Dim renders text with reduced intensity.
var Dim prop = dimProp{}

type strikethroughProp struct{}

func (strikethroughProp) apply(n *Node) { n.Paint.Strikethrough = true }

// Strikethrough renders text with a horizontal line through it.
var Strikethrough prop = strikethroughProp{}

type reverseProp struct{}

func (reverseProp) apply(n *Node) { n.Paint.Reverse = true }

// Reverse swaps foreground and background colors.
var Reverse prop = reverseProp{}

// --- Text overflow props ---

type textOverflowProp TextOverflow

func (p textOverflowProp) apply(n *Node) { n.TextOverflow = TextOverflow(p) }

// Ellipsis truncates text with "…" when it exceeds the available width.
var Ellipsis prop = textOverflowProp(TextOverflowEllipsis)

// ClipText hard-clips text at the available width without wrapping.
var ClipText prop = textOverflowProp(TextOverflowClip)

// --- Position props ---

type positionProp Position

func (p positionProp) apply(n *Node) { n.Style.Position = Position(p) }

// Absolute removes a node from flex flow and positions it relative to
// its parent's content box using Left() and Top().
var Absolute prop = positionProp(PositionAbsolute)

type leftProp Dimension

func (p leftProp) apply(n *Node) { n.Style.Left = Dimension(p) }

// Left sets the left offset for absolutely positioned nodes.
func Left(v any) leftProp { return leftProp(toDimension(v)) }

type topProp Dimension

func (p topProp) apply(n *Node) { n.Style.Top = Dimension(p) }

// Top sets the top offset for absolutely positioned nodes.
func Top(v any) topProp { return topProp(toDimension(v)) }

type rightProp Dimension

func (p rightProp) apply(n *Node) { n.Style.Right = Dimension(p) }

// Right sets the right offset for absolutely positioned nodes.
// If both Left and Right are set, Left takes priority.
func Right(v any) rightProp { return rightProp(toDimension(v)) }

type bottomProp Dimension

func (p bottomProp) apply(n *Node) { n.Style.Bottom = Dimension(p) }

// Bottom sets the bottom offset for absolutely positioned nodes.
// If both Top and Bottom are set, Top takes priority.
func Bottom(v any) bottomProp { return bottomProp(toDimension(v)) }

// --- Z-index prop ---

type zIndexProp int

func (z zIndexProp) apply(n *Node) { n.Style.ZIndex = int(z) }

// ZIndex controls render order. Higher values render on top.
func ZIndex(v int) zIndexProp { return zIndexProp(v) }

// --- Cursor style ---

// CursorStyle controls the terminal cursor shape.
type CursorStyle int

const (
	CursorDefault        CursorStyle = iota // terminal default
	CursorBlockBlink                        // blinking block ▊
	CursorBlock                             // steady block ▊
	CursorUnderlineBlink                    // blinking underline _
	CursorUnderline                         // steady underline _
	CursorBarBlink                          // blinking bar |
	CursorBar                               // steady bar |
	CursorHidden                            // invisible
)

// Standard ANSI terminal colors. Use [RGBColor] for 24-bit colors.
var (
	Black         = ANSIColor(0)
	Red           = ANSIColor(1)
	Green         = ANSIColor(2)
	Yellow        = ANSIColor(3)
	Blue          = ANSIColor(4)
	Magenta       = ANSIColor(5)
	Cyan          = ANSIColor(6)
	White         = ANSIColor(7)
	Gray          = ANSIColor(8)
	BrightRed     = ANSIColor(9)
	BrightGreen   = ANSIColor(10)
	BrightYellow  = ANSIColor(11)
	BrightBlue    = ANSIColor(12)
	BrightMagenta = ANSIColor(13)
	BrightCyan    = ANSIColor(14)
	BrightWhite   = ANSIColor(15)
)

// --- Element constructors ---

// Box creates a container node. Arguments can be props (styling) or
// *Node children — they're sorted out automatically.
//
//	Box(FlexColumn, JustifyCenter, Padding(1),
//	    Text("hello", TextColor(Cyan), Bold),
//	    Text("world"),
//	)
func Box(args ...any) *Node {
	n := NewNode(DefaultStyle())
	n.Style.FlexGrow = 1
	for _, arg := range args {
		switch v := arg.(type) {
		case *Node:
			if v != nil {
				n.AddChild(v)
			}
		case []*Node:
			for _, child := range v {
				if child != nil {
					n.AddChild(child)
				}
			}
		case prop:
			v.apply(n)
		}
	}
	return n
}

// Text creates a styled text leaf node.
//
//	Text("hello world", TextColor(Green), Bold)
func Text(content string, args ...any) *Node {
	n := NewText(content)
	for _, arg := range args {
		if p, ok := arg.(prop); ok {
			p.apply(n)
		}
	}
	return n
}

// --- Conditional rendering ---

// If returns node when cond is true, nil otherwise.
// Nil nodes are safely ignored by [Box].
//
//	Box(FlexColumn,
//	    If(showHeader, Text("Header", Bold)),
//	    Text("Always visible"),
//	)
func If(cond bool, node *Node) *Node {
	if cond {
		return node
	}
	return nil
}

// IfElse returns a when cond is true, b otherwise.
//
//	Box(FlexColumn,
//	    IfElse(loggedIn, Text("Welcome"), Text("Please log in")),
//	)
func IfElse(cond bool, a, b *Node) *Node {
	if cond {
		return a
	}
	return b
}

// Map converts a slice into a list of nodes using the given function.
// The result can be spread directly into [Box] args.
//
//	Box(FlexColumn,
//	    Text("Header"),
//	    Map(items, func(item string, i int) *Node {
//	        return Text(item)
//	    }),
//	    Text("Footer"),
//	)
func Map[T any](items []T, fn func(item T, index int) *Node) []*Node {
	nodes := make([]*Node, len(items))
	for i, item := range items {
		nodes[i] = fn(item, i)
	}
	return nodes
}

// --- Debug prop ---

type debugProp struct{}

func (debugProp) apply(n *Node) { n.debug = true }

// Debug draws colored outlines around the node and all its descendants,
// making the layout tree visible for debugging.
var Debug prop = debugProp{}

// --- Focus helpers ---

// FocusColor returns active when focused is true, inactive otherwise.
// Use with any color prop to visually indicate focus state.
//
//	Text("Name", TextColor(FocusColor(input.Focused, Cyan, Gray)))
//	Box(BorderColor(FocusColor(input.Focused, Cyan, Gray)), ...)
//	Box(BackgroundColor(FocusColor(input.Focused, Blue, Black)), ...)
func FocusColor(focused bool, active, inactive Color) Color {
	if focused {
		return active
	}
	return inactive
}

// FocusBorderColor is a shorthand for BorderColor(FocusColor(...)).
//
//	Box(BorderRounded, FocusBorderColor(input.Focused, Cyan, Gray),
//	    Input(input),
//	)
func FocusBorderColor(focused bool, active, inactive Color) borderColorProp {
	return borderColorProp{FocusColor(focused, active, inactive)}
}

type pickProp struct{ chosen prop }

func (p pickProp) apply(n *Node) { p.chosen.apply(n) }

// Pick returns a or b based on the condition. Works with any prop.
//
//	Pick(input.Focused, TextColor(Cyan), TextColor(Gray))
//	Pick(checked, Bold, Dim)
func Pick(cond bool, a, b prop) prop {
	if cond {
		return pickProp{a}
	}
	return pickProp{b}
}
