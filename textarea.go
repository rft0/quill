package quill

import "strings"

// TextareaState holds the mutable state of a multi-line text input.
type TextareaState struct {
	Value      string
	CursorRow  int
	CursorCol  int // rune index within current line
	Focused    bool
	ViewHeight int          // visible lines; 0 = show all
	OnSubmit   func(string) // called on Alt+Enter if set

	scrollOffset int
}

// Focus gives the textarea keyboard focus.
func (s *TextareaState) Focus() { s.Focused = true }

// Blur removes keyboard focus.
func (s *TextareaState) Blur() { s.Focused = false }

// Lines returns the value split into lines.
func (s *TextareaState) Lines() []string { return strings.Split(s.Value, "\n") }

// Update processes a keyboard event for the textarea.
func (s *TextareaState) Update(msg Msg) Cmd {
	if !s.Focused {
		return nil
	}
	key, ok := msg.(KeyMsg)
	if !ok {
		return nil
	}

	lines := s.Lines()
	s.clampCursor(lines)
	lineRunes := []rune(lines[s.CursorRow])

	switch key.Type {
	case KeyRune, KeySpace:
		ch := key.Rune
		if key.Type == KeySpace {
			ch = ' '
		}
		lineRunes = append(lineRunes[:s.CursorCol], append([]rune{ch}, lineRunes[s.CursorCol:]...)...)
		lines[s.CursorRow] = string(lineRunes)
		s.CursorCol++

	case KeyEnter:
		if key.Alt && s.OnSubmit != nil {
			s.OnSubmit(s.Value)
			return nil
		}
		before := string(lineRunes[:s.CursorCol])
		after := string(lineRunes[s.CursorCol:])
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:s.CursorRow]...)
		newLines = append(newLines, before, after)
		newLines = append(newLines, lines[s.CursorRow+1:]...)
		lines = newLines
		s.CursorRow++
		s.CursorCol = 0

	case KeyBackspace:
		if s.CursorCol > 0 {
			lineRunes = append(lineRunes[:s.CursorCol-1], lineRunes[s.CursorCol:]...)
			lines[s.CursorRow] = string(lineRunes)
			s.CursorCol--
		} else if s.CursorRow > 0 {
			prevRunes := []rune(lines[s.CursorRow-1])
			s.CursorCol = len(prevRunes)
			lines[s.CursorRow-1] = lines[s.CursorRow-1] + string(lineRunes)
			lines = append(lines[:s.CursorRow], lines[s.CursorRow+1:]...)
			s.CursorRow--
		}

	case KeyDelete:
		if s.CursorCol < len(lineRunes) {
			lineRunes = append(lineRunes[:s.CursorCol], lineRunes[s.CursorCol+1:]...)
			lines[s.CursorRow] = string(lineRunes)
		} else if s.CursorRow < len(lines)-1 {
			lines[s.CursorRow] = string(lineRunes) + lines[s.CursorRow+1]
			lines = append(lines[:s.CursorRow+1], lines[s.CursorRow+2:]...)
		}

	case KeyLeft:
		if s.CursorCol > 0 {
			s.CursorCol--
		} else if s.CursorRow > 0 {
			s.CursorRow--
			s.CursorCol = len([]rune(lines[s.CursorRow]))
		}

	case KeyRight:
		if s.CursorCol < len(lineRunes) {
			s.CursorCol++
		} else if s.CursorRow < len(lines)-1 {
			s.CursorRow++
			s.CursorCol = 0
		}

	case KeyUp:
		if s.CursorRow > 0 {
			s.CursorRow--
			maxCol := len([]rune(lines[s.CursorRow]))
			if s.CursorCol > maxCol {
				s.CursorCol = maxCol
			}
		}

	case KeyDown:
		if s.CursorRow < len(lines)-1 {
			s.CursorRow++
			maxCol := len([]rune(lines[s.CursorRow]))
			if s.CursorCol > maxCol {
				s.CursorCol = maxCol
			}
		}

	case KeyHome, KeyCtrlA:
		s.CursorCol = 0

	case KeyEnd, KeyCtrlE:
		s.CursorCol = len(lineRunes)

	case KeyCtrlU:
		lines[s.CursorRow] = string(lineRunes[s.CursorCol:])
		s.CursorCol = 0

	case KeyCtrlK:
		lines[s.CursorRow] = string(lineRunes[:s.CursorCol])
	}

	s.Value = strings.Join(lines, "\n")
	s.ensureScroll()
	return nil
}

func (s *TextareaState) clampCursor(lines []string) {
	if s.CursorRow >= len(lines) {
		s.CursorRow = len(lines) - 1
	}
	if s.CursorRow < 0 {
		s.CursorRow = 0
	}
	maxCol := len([]rune(lines[s.CursorRow]))
	if s.CursorCol > maxCol {
		s.CursorCol = maxCol
	}
	if s.CursorCol < 0 {
		s.CursorCol = 0
	}
}

func (s *TextareaState) ensureScroll() {
	if s.ViewHeight <= 0 {
		return
	}
	if s.CursorRow < s.scrollOffset {
		s.scrollOffset = s.CursorRow
	}
	if s.CursorRow >= s.scrollOffset+s.ViewHeight {
		s.scrollOffset = s.CursorRow - s.ViewHeight + 1
	}
}

// Textarea creates a multi-line text input element. Pass a pointer to a
// TextareaState plus any styling props.
//
//	quill.Textarea(&state, quill.BorderRounded, quill.Height(quill.Px(10)))
func Textarea(state *TextareaState, args ...any) *Node {
	lines := state.Lines()
	state.clampCursor(lines)
	state.ensureScroll()

	startLine := state.scrollOffset
	endLine := len(lines)
	if state.ViewHeight > 0 {
		endLine = startLine + state.ViewHeight
		if endLine > len(lines) {
			endLine = len(lines)
		}
	}

	var children []any
	children = append(children, FlexColumn)

	for _, arg := range args {
		if _, ok := arg.(prop); ok {
			children = append(children, arg)
		}
	}

	for i := startLine; i < endLine; i++ {
		text := lines[i]
		if text == "" {
			text = " " // ensure empty lines have height for cursor
		}
		n := NewText(text)
		n.Style.FlexGrow = 1
		if state.Focused && i == state.CursorRow {
			n.showCursor = true
			n.cursorPos = state.CursorCol
		}
		children = append(children, n)
	}

	return Box(children...)
}
