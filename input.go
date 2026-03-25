package quill

import "strings"

// InputState holds the mutable state of a text input.
// Keep this in your component struct and pass a pointer to Input().
type InputState struct {
	Value    string
	Cursor   int
	Focused  bool
	Hidden   bool         // when true, display mask characters instead of actual value
	OnSubmit func(string) // called when Enter is pressed while focused
}

// Focus gives the input keyboard focus.
func (s *InputState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *InputState) Blur() { s.Focused = false }

// Update processes a keyboard event. Call from your component's OnKey handler.
func (s *InputState) Update(msg Msg) {
	if !s.Focused {
		return
	}

	key, ok := msg.(KeyMsg)
	if !ok {
		return
	}

	if key.Type == KeyEnter {
		if s.OnSubmit != nil {
			s.OnSubmit(s.Value)
		}
		return
	}

	runes := []rune(s.Value)
	cur := s.runePos()

	switch key.Type {
	case KeyRune, KeySpace:
		ch := key.Rune
		if key.Type == KeySpace {
			ch = ' '
		}
		runes = append(runes[:cur], append([]rune{ch}, runes[cur:]...)...)
		s.Value = string(runes)
		s.Cursor = len(string(runes[:cur+1]))

	case KeyBackspace:
		if cur > 0 {
			runes = append(runes[:cur-1], runes[cur:]...)
			s.Value = string(runes)
			s.Cursor = len(string(runes[:cur-1]))
		}

	case KeyDelete:
		if cur < len(runes) {
			runes = append(runes[:cur], runes[cur+1:]...)
			s.Value = string(runes)
		}

	case KeyLeft:
		if cur > 0 {
			s.Cursor = len(string(runes[:cur-1]))
		}

	case KeyRight:
		if cur < len(runes) {
			s.Cursor = len(string(runes[:cur+1]))
		}

	case KeyHome, KeyCtrlA:
		s.Cursor = 0

	case KeyEnd, KeyCtrlE:
		s.Cursor = len(s.Value)

	case KeyCtrlU:
		s.Value = string(runes[cur:])
		s.Cursor = 0

	case KeyCtrlK:
		s.Value = string(runes[:cur])
	}
}

func (s *InputState) runePos() int {
	return len([]rune(s.Value[:s.Cursor]))
}

// VisualCol returns the display column of the cursor, accounting for
// double-width characters (CJK, emoji).
func (s *InputState) VisualCol() int {
	return stringWidth(s.Value[:s.Cursor])
}

// Input creates a text input element. Pass a pointer to an InputState
// plus any styling props.
//
//	quill.Input(&f.name, quill.TextColor(quill.Yellow))
func Input(state *InputState, args ...any) *Node {
	display := state.Value
	if state.Hidden {
		display = strings.Repeat("•", len([]rune(state.Value)))
	}
	n := NewText(display)
	n.Style.FlexGrow = 1

	for _, arg := range args {
		if p, ok := arg.(prop); ok {
			p.apply(n)
		}
	}

	if state.Focused {
		n.cursorPos = state.runePos()
		n.showCursor = true
	}

	return n
}

// Focusable is implemented by any widget state that supports focus management.
type Focusable interface {
	Focus()
	Blur()
	Update(Msg)
}

// FocusGroup manages focus across multiple focusable widgets.
// It handles Tab/Shift-Tab cycling and routes key events to the focused widget.
type FocusGroup struct {
	items []Focusable
	index int
}

// NewFocusGroup creates a focus group from the given focusable widgets.
// The first item is focused automatically.
func NewFocusGroup(items ...Focusable) FocusGroup {
	fg := FocusGroup{items: items}
	if len(items) > 0 {
		items[0].Focus()
	}
	return fg
}

// UseFocusGroup creates a [FocusGroup] as a hook, eliminating the
// UseRef + UseEffect boilerplate. The first item is focused automatically
// on mount.
//
//	focus := quill.UseFocusGroup(ctx, name, email, password)
func UseFocusGroup(ctx *Context, items ...Focusable) *FocusGroup {
	fg := UseRef(ctx, FocusGroup{})
	UseEffect(ctx, func() {
		*fg = NewFocusGroup(items...)
	})
	return fg
}

// Next moves focus to the next item.
func (fg *FocusGroup) Next() {
	if len(fg.items) == 0 {
		return
	}
	fg.items[fg.index].Blur()
	fg.index = (fg.index + 1) % len(fg.items)
	fg.items[fg.index].Focus()
}

// Prev moves focus to the previous item.
func (fg *FocusGroup) Prev() {
	if len(fg.items) == 0 {
		return
	}
	fg.items[fg.index].Blur()
	fg.index = (fg.index - 1 + len(fg.items)) % len(fg.items)
	fg.items[fg.index].Focus()
}

// FormConfig configures a [UseForm] hook.
type FormConfig struct {
	Fields   []Focusable // form fields in tab order
	OnSubmit func()      // called when Enter is pressed on the last field
}

// Form manages a group of form fields with focus cycling and submission.
type Form struct {
	focus    FocusGroup
	onSubmit func()
}

// UseForm creates a form that manages focus cycling (Tab/Shift+Tab),
// routes key events to the focused field, and calls OnSubmit when
// Enter is pressed on the last field.
//
//	form := quill.UseForm(ctx, quill.FormConfig{
//	    Fields:   []quill.Focusable{name, email, password},
//	    OnSubmit: func() { /* submit */ },
//	})
func UseForm(ctx *Context, cfg FormConfig) *Form {
	f := UseRef(ctx, Form{})
	UseEffect(ctx, func() {
		f.focus = NewFocusGroup(cfg.Fields...)
		f.onSubmit = cfg.OnSubmit
	})

	OnKey(ctx, func(key KeyMsg) {
		switch key.Type {
		case KeyTab:
			f.focus.Next()
		case KeyShiftTab:
			f.focus.Prev()
		case KeyEnter:
			if f.focus.IsLast() && f.onSubmit != nil {
				f.onSubmit()
			} else {
				f.focus.Next()
			}
		default:
			f.focus.Update(key)
		}
	})

	return f
}

// IsLast returns true if the last item in the group is focused.
func (fg *FocusGroup) IsLast() bool {
	return len(fg.items) > 0 && fg.index == len(fg.items)-1
}

// Update forwards the message to the currently focused item.
func (fg *FocusGroup) Update(msg Msg) {
	if len(fg.items) == 0 {
		return
	}
	fg.items[fg.index].Update(msg)
}
