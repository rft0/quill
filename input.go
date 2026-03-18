package quill

// InputState holds the mutable state of a text input.
// Keep this in your component struct and pass a pointer to Input().
type InputState struct {
	Value    string
	Cursor   int
	Focused  bool
	OnSubmit func(string) // called when Enter is pressed while focused
}

// Focus gives the input keyboard focus.
func (s *InputState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *InputState) Blur() { s.Focused = false }

// Update processes a keyboard event. Call from your component's Update.
func (s *InputState) Update(msg Msg) Cmd {
	if !s.Focused {
		return nil
	}

	key, ok := msg.(KeyMsg)
	if !ok {
		return nil
	}

	if key.Type == KeyEnter {
		if s.OnSubmit != nil {
			s.OnSubmit(s.Value)
		}
		return nil
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

	return nil
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
	n := NewText(state.Value)
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
	Update(Msg) Cmd
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

// Update forwards the message to the currently focused item.
func (fg *FocusGroup) Update(msg Msg) Cmd {
	if len(fg.items) == 0 {
		return nil
	}
	return fg.items[fg.index].Update(msg)
}
