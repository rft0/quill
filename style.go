package quill

// FlexDirection controls the main axis of a container.
type FlexDirection int

const (
	FlexRow FlexDirection = iota
	FlexColumn
)

// JustifyContent controls distribution of children along the main axis.
type JustifyContent int

const (
	JustifyFlexStart JustifyContent = iota
	JustifyFlexEnd
	JustifyCenter
	JustifySpaceBetween
	JustifySpaceAround
)

// AlignItems controls alignment of children along the cross axis.
type AlignItems int

const (
	AlignFlexStart AlignItems = iota
	AlignFlexEnd
	AlignCenter
	AlignStretch
)

// AlignSelf overrides the parent's AlignItems for a single child.
type AlignSelf int

const (
	AlignSelfAuto AlignSelf = iota // inherit from parent
	AlignSelfFlexStart
	AlignSelfFlexEnd
	AlignSelfCenter
	AlignSelfStretch
)

// FlexWrap controls whether children wrap onto multiple lines.
type FlexWrap int

const (
	FlexNoWrap FlexWrap = iota // default, single line
	FlexWrapWrap               // wrap onto multiple lines
)

// BorderStyle controls the border drawn around a node.
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSingle  // ┌─┐│└─┘
	BorderDouble  // ╔═╗║╚═╝
	BorderRounded // ╭─╮│╰─╯
	BorderThick   // ┏━┓┃┗━┛
)

type borderChars struct {
	TL, TR, BL, BR rune
	H, V           rune
}

var borderCharSets = [...]borderChars{
	BorderSingle:  {'┌', '┐', '└', '┘', '─', '│'},
	BorderDouble:  {'╔', '╗', '╚', '╝', '═', '║'},
	BorderRounded: {'╭', '╮', '╰', '╯', '─', '│'},
	BorderThick:   {'┏', '┓', '┗', '┛', '━', '┃'},
}

// DimensionUnit specifies how a dimension value is interpreted.
type DimensionUnit int

const (
	Auto DimensionUnit = iota
	Fixed
	Percent
)

// Dimension represents a length that can be fixed, percentage, or auto.
type Dimension struct {
	Value float64
	Unit  DimensionUnit
}

// Helpers for constructing dimensions.
func Px(v float64) Dimension  { return Dimension{Value: v, Unit: Fixed} }
func Pct(v float64) Dimension { return Dimension{Value: v, Unit: Percent} }
func AutoDim() Dimension      { return Dimension{Unit: Auto} }

// Position controls whether a node participates in flow or is positioned absolutely.
type Position int

const (
	PositionRelative Position = iota // default, normal flex flow
	PositionAbsolute                 // removed from flow, positioned relative to parent
)

// TextOverflow controls how text behaves when it exceeds available width.
type TextOverflow int

const (
	TextOverflowWrap     TextOverflow = iota // default: wrap to next line
	TextOverflowEllipsis                     // truncate with "…"
	TextOverflowClip                         // hard clip, no wrap
)

// Edges represents top/right/bottom/left values (padding, margin, border).
type Edges struct {
	Top, Right, Bottom, Left float64
}

// Horizontal returns Left + Right.
func (e Edges) Horizontal() float64 { return e.Left + e.Right }

// Vertical returns Top + Bottom.
func (e Edges) Vertical() float64 { return e.Top + e.Bottom }

// UniformEdges creates edges with the same value on all sides.
func UniformEdges(v float64) Edges {
	return Edges{Top: v, Right: v, Bottom: v, Left: v}
}

// Style holds all layout-relevant properties for a node.
type Style struct {
	Direction  FlexDirection
	Justify    JustifyContent
	Align      AlignItems
	AlignSelf  AlignSelf
	FlexGrow   float64
	FlexShrink float64
	FlexBasis  Dimension

	Width  Dimension
	Height Dimension

	MinWidth  Dimension
	MinHeight Dimension
	MaxWidth  Dimension
	MaxHeight Dimension

	Padding Edges
	Margin  Edges
	Gap  float64
	Wrap FlexWrap

	Border BorderStyle

	Position Position  // relative (default) or absolute
	Left     Dimension // offset from parent left (absolute only)
	Top      Dimension // offset from parent top (absolute only)
	Right    Dimension // offset from parent right (absolute only)
	Bottom   Dimension // offset from parent bottom (absolute only)
	ZIndex   int       // render order; higher values render on top
}

// BorderWidth returns the cell width consumed by the border (0 or 1).
func (s *Style) BorderWidth() float64 {
	if s.Border != BorderNone {
		return 1
	}
	return 0
}

// DefaultStyle returns a style with sensible defaults matching CSS flexbox:
// row direction, flex-start justify, stretch align, shrink=1.
func DefaultStyle() Style {
	return Style{
		Direction:  FlexRow,
		Justify:    JustifyFlexStart,
		Align:      AlignStretch,
		FlexGrow:   0,
		FlexShrink: 1,
		FlexBasis:  AutoDim(),
		Width:      AutoDim(),
		Height:     AutoDim(),
		MinWidth:   Dimension{Value: 0, Unit: Fixed},
		MinHeight:  Dimension{Value: 0, Unit: Fixed},
		MaxWidth:   AutoDim(),
		MaxHeight:  AutoDim(),
	}
}
