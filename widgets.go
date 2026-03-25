package quill

import (
	"math"
	"strings"
	"time"
)

// --- Select ---

// SelectState holds the state of a select/list picker.
type SelectState struct {
	Options  []string
	Selected int
	Focused  bool
}

// Focus gives the select keyboard focus.
func (s *SelectState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *SelectState) Blur() { s.Focused = false }

// Update processes keyboard events for the select.
func (s *SelectState) Update(msg Msg) {
	if !s.Focused {
		return
	}
	key, ok := msg.(KeyMsg)
	if !ok {
		return
	}
	switch key.Type {
	case KeyUp:
		if s.Selected > 0 {
			s.Selected--
		}
	case KeyDown:
		if s.Selected < len(s.Options)-1 {
			s.Selected++
		}
	case KeyRune:
		switch key.Rune {
		case 'k':
			if s.Selected > 0 {
				s.Selected--
			}
		case 'j':
			if s.Selected < len(s.Options)-1 {
				s.Selected++
			}
		}
	}
}

// Select creates a list picker element.
//
//	quill.Select(&state, quill.BorderRounded)
func Select(state *SelectState, args ...any) *Node {
	children := make([]any, 0, len(state.Options)+len(args))

	// Collect props first.
	children = append(children, FlexColumn)
	for _, arg := range args {
		if _, ok := arg.(prop); ok {
			children = append(children, arg)
		}
	}

	for i, opt := range state.Options {
		prefix := "  "
		if i == state.Selected {
			prefix = "> "
		}

		item := Text(prefix+opt, TextColor(White))
		if i == state.Selected && state.Focused {
			item.Paint.Reverse = true
		}
		children = append(children, item)
	}

	return Box(children...)
}

// --- Checkbox ---

// CheckboxState holds the state of a checkbox.
type CheckboxState struct {
	Checked bool
	Focused bool
}

// Focus gives the checkbox keyboard focus.
func (s *CheckboxState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *CheckboxState) Blur() { s.Focused = false }

// Toggle flips the checked state.
func (s *CheckboxState) Toggle() { s.Checked = !s.Checked }

// Update processes keyboard events for the checkbox.
func (s *CheckboxState) Update(msg Msg) {
	if !s.Focused {
		return
	}
	key, ok := msg.(KeyMsg)
	if !ok {
		return
	}
	if key.Type == KeySpace || key.Type == KeyEnter {
		s.Toggle()
	}
}

// Checkbox creates a checkbox element with a label.
//
//	quill.Checkbox(&state, "Accept terms", quill.TextColor(quill.Green))
func Checkbox(state *CheckboxState, label string, args ...any) *Node {
	check := "[ ] "
	if state.Checked {
		check = "[x] "
	}
	n := Text(check+label, args...)
	if state.Focused {
		n.Paint.Reverse = true
	}
	return n
}

// --- ProgressBar ---

// ProgressBar creates a progress bar element. Value should be 0.0 to 1.0.
//
//	quill.ProgressBar(0.75, quill.TextColor(quill.Green))
func ProgressBar(value float64, args ...any) *Node {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}

	s := DefaultStyle()
	s.FlexGrow = 1
	n := &Node{
		Style: s,
		MeasureFunc: func(availWidth, availHeight float64) (float64, float64) {
			return availWidth, 1
		},
	}

	n.isProgress = true
	n.progressValue = value

	for _, arg := range args {
		if p, ok := arg.(prop); ok {
			p.apply(n)
		}
	}

	return n
}

// --- Spinner ---

// Spinner frame sets.
var (
	SpinnerDots  = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	SpinnerLine  = []string{"|", "/", "-", "\\"}
	SpinnerBlock = []string{"▖", "▘", "▝", "▗"}
)

// Spinner creates a self-animating spinner element. It manages its own
// timing internally — just pass a frame set and optional styling props.
//
//	quill.Spinner(ctx, quill.SpinnerDots, quill.TextColor(quill.Cyan))
func Spinner(ctx *Context, frames []string, args ...any) *Node {
	if len(frames) == 0 {
		return Text(" ", args...)
	}
	frame := UseState(ctx, 0)
	UseInterval(ctx, 80*time.Millisecond, func() {
		frame.Set((frame.Get() + 1) % len(frames))
	})
	return Text(frames[frame.Get()%len(frames)], args...)
}

// --- ScrollView ---

// ScrollState holds the scroll offset for a scrollable container.
type ScrollState struct {
	Offset        int
	ContentHeight int
	ViewHeight    int
}

// ScrollUp scrolls up by n lines.
func (s *ScrollState) ScrollUp(n int) {
	s.Offset -= n
	if s.Offset < 0 {
		s.Offset = 0
	}
}

// ScrollDown scrolls down by n lines.
func (s *ScrollState) ScrollDown(n int) {
	s.Offset += n
	max := s.ContentHeight - s.ViewHeight
	if max < 0 {
		max = 0
	}
	if s.Offset > max {
		s.Offset = max
	}
}

// PageUp scrolls up by one page.
func (s *ScrollState) PageUp() { s.ScrollUp(s.ViewHeight) }

// PageDown scrolls down by one page.
func (s *ScrollState) PageDown() { s.ScrollDown(s.ViewHeight) }

// ScrollView creates a scrollable container. The container clips its
// children to the visible area and offsets them by the scroll state.
//
//	quill.ScrollView(&scrollState, quill.Height(Px(10)),
//	    quill.Text("line 1"),
//	    quill.Text("line 2"),
//	    // ... many children
//	)
func ScrollView(state *ScrollState, args ...any) *Node {
	n := Box(args...)
	n.scrollState = state
	return n
}

// --- List ---

// ListState holds the state for a virtualized scrollable list.
type ListState struct {
	Selected int
	Offset   int  // first visible item index
	Focused  bool
	total    int
	viewSize int
}

// Focus gives the list keyboard focus.
func (s *ListState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *ListState) Blur() { s.Focused = false }

// Total returns the total number of items.
func (s *ListState) Total() int { return s.total }

// ViewSize returns the number of visible items.
func (s *ListState) ViewSize() int { return s.viewSize }

// ensureVisible scrolls to keep the selected item in view.
func (s *ListState) ensureVisible() {
	if s.Selected < s.Offset {
		s.Offset = s.Selected
	}
	if s.viewSize > 0 && s.Selected >= s.Offset+s.viewSize {
		s.Offset = s.Selected - s.viewSize + 1
	}
	if s.Offset < 0 {
		s.Offset = 0
	}
}

// Update processes keyboard events for the list.
func (s *ListState) Update(msg Msg) {
	if !s.Focused {
		return
	}
	key, ok := msg.(KeyMsg)
	if !ok {
		return
	}
	switch key.Type {
	case KeyUp:
		if s.Selected > 0 {
			s.Selected--
			s.ensureVisible()
		}
	case KeyDown:
		if s.Selected < s.total-1 {
			s.Selected++
			s.ensureVisible()
		}
	case KeyPageUp:
		s.Selected -= s.viewSize
		if s.Selected < 0 {
			s.Selected = 0
		}
		s.ensureVisible()
	case KeyPageDown:
		s.Selected += s.viewSize
		if s.Selected >= s.total {
			s.Selected = s.total - 1
		}
		if s.Selected < 0 {
			s.Selected = 0
		}
		s.ensureVisible()
	case KeyHome:
		s.Selected = 0
		s.ensureVisible()
	case KeyEnd:
		s.Selected = s.total - 1
		s.ensureVisible()
	case KeyRune:
		switch key.Rune {
		case 'k':
			if s.Selected > 0 {
				s.Selected--
				s.ensureVisible()
			}
		case 'j':
			if s.Selected < s.total-1 {
				s.Selected++
				s.ensureVisible()
			}
		}
	}
}

// List creates a virtualized scrollable list that only renders visible items.
// viewHeight is the number of visible rows. renderItem is called for each
// visible item with its index and whether it is selected.
//
//	quill.List(state, 10, len(items), func(i int, selected bool) *quill.Node {
//	    return quill.Text(items[i])
//	})
func List(state *ListState, viewHeight int, total int, renderItem func(index int, selected bool) *Node, args ...any) *Node {
	state.total = total
	state.viewSize = viewHeight
	state.ensureVisible()

	children := make([]any, 0, viewHeight+len(args))
	children = append(children, FlexColumn)
	for _, arg := range args {
		if _, ok := arg.(prop); ok {
			children = append(children, arg)
		}
	}

	end := state.Offset + viewHeight
	if end > total {
		end = total
	}
	for i := state.Offset; i < end; i++ {
		node := renderItem(i, i == state.Selected)
		if node != nil {
			children = append(children, node)
		}
	}

	return Box(children...)
}

// --- Notify ---

// Notify creates an absolutely positioned toast/notification at the top-right.
// Control visibility with [If] and timing with UseAfter.
//
//	If(showToast.Get(), Notify(
//	    Text("Saved!", TextColor(Green)),
//	    BackgroundColor(Black), BorderRounded,
//	))
func Notify(content *Node, args ...any) *Node {
	n := Box(
		Absolute, Top(0), Right(0),
		ZIndex(99),
		content,
	)
	for _, arg := range args {
		if p, ok := arg.(prop); ok {
			p.apply(n)
		}
	}
	return n
}

// --- Modal ---

// Modal creates an absolutely positioned overlay that centers its content.
// It fills the parent's content box and renders above other children via ZIndex.
// Add [BackgroundColor] for a dimmed backdrop.
//
//	Modal(
//	    Box(BorderRounded, PadXY(2, 1), BackgroundColor(Black),
//	        Text("Are you sure?"),
//	    ),
//	    BackgroundColor(RGBColor(0, 0, 0)), // dim backdrop
//	)
func Modal(content *Node, args ...any) *Node {
	overlay := Box(
		Absolute, Left(0), Top(0),
		FlexColumn, JustifyCenter, AlignCenter,
		ZIndex(100),
		content,
	)
	for _, arg := range args {
		if p, ok := arg.(prop); ok {
			p.apply(overlay)
		}
	}
	return overlay
}

// renderProgressBar generates the bar text for a progress node.
func renderProgressBar(width int, value float64) string {
	if width <= 0 {
		return ""
	}
	filled := int(math.Round(float64(width) * value))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
