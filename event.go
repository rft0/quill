package quill

import "fmt"

// Msg is the interface for all events in the application.
type Msg interface{}

// KeyMsg represents a keyboard input event.
type KeyMsg struct {
	Type KeyType
	Rune rune // only meaningful when Type == KeyRune
	Alt  bool
}

// String returns a human-readable description of the key event.
func (k KeyMsg) String() string {
	if k.Type == KeyRune {
		if k.Alt {
			return fmt.Sprintf("Alt+%c", k.Rune)
		}
		return string(k.Rune)
	}
	name := keyTypeName(k.Type)
	if k.Alt {
		return "Alt+" + name
	}
	return name
}

func keyTypeName(kt KeyType) string {
	switch kt {
	case KeyEnter:
		return "Enter"
	case KeyTab:
		return "Tab"
	case KeyShiftTab:
		return "Shift+Tab"
	case KeyBackspace:
		return "Backspace"
	case KeyEscape:
		return "Escape"
	case KeySpace:
		return "Space"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyDelete:
		return "Delete"
	case KeyInsert:
		return "Insert"
	case KeyCtrlA:
		return "Ctrl+A"
	case KeyCtrlB:
		return "Ctrl+B"
	case KeyCtrlC:
		return "Ctrl+C"
	case KeyCtrlD:
		return "Ctrl+D"
	case KeyCtrlE:
		return "Ctrl+E"
	case KeyCtrlF:
		return "Ctrl+F"
	case KeyCtrlG:
		return "Ctrl+G"
	case KeyCtrlK:
		return "Ctrl+K"
	case KeyCtrlL:
		return "Ctrl+L"
	case KeyCtrlN:
		return "Ctrl+N"
	case KeyCtrlO:
		return "Ctrl+O"
	case KeyCtrlP:
		return "Ctrl+P"
	case KeyCtrlQ:
		return "Ctrl+Q"
	case KeyCtrlR:
		return "Ctrl+R"
	case KeyCtrlS:
		return "Ctrl+S"
	case KeyCtrlT:
		return "Ctrl+T"
	case KeyCtrlU:
		return "Ctrl+U"
	case KeyCtrlV:
		return "Ctrl+V"
	case KeyCtrlW:
		return "Ctrl+W"
	case KeyCtrlX:
		return "Ctrl+X"
	case KeyCtrlY:
		return "Ctrl+Y"
	case KeyCtrlZ:
		return "Ctrl+Z"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	default:
		return fmt.Sprintf("Key(%d)", int(kt))
	}
}

// KeyType identifies the key pressed.
type KeyType int

const (
	KeyRune KeyType = iota // printable character stored in Rune
	KeyEnter
	KeyTab
	KeyShiftTab
	KeyBackspace
	KeyEscape
	KeySpace
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyDelete
	KeyInsert
	KeyCtrlA
	KeyCtrlB
	KeyCtrlC
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlK
	KeyCtrlL
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// ResizeMsg is sent when the terminal window changes size.
type ResizeMsg struct {
	Width, Height int
}

// MouseType identifies the mouse action.
type MouseType int

const (
	MouseLeft MouseType = iota
	MouseMiddle
	MouseRight
	MouseRelease
	MouseWheelUp
	MouseWheelDown
	MouseMotion
)

// MouseMsg represents a mouse input event.
type MouseMsg struct {
	Type MouseType
	X, Y int // 0-based cell coordinates
}

// String returns a human-readable description of the mouse event.
func (m MouseMsg) String() string {
	return fmt.Sprintf("%s(%d,%d)", mouseTypeName(m.Type), m.X, m.Y)
}

func mouseTypeName(mt MouseType) string {
	switch mt {
	case MouseLeft:
		return "MouseLeft"
	case MouseMiddle:
		return "MouseMiddle"
	case MouseRight:
		return "MouseRight"
	case MouseRelease:
		return "MouseRelease"
	case MouseWheelUp:
		return "MouseWheelUp"
	case MouseWheelDown:
		return "MouseWheelDown"
	case MouseMotion:
		return "MouseMotion"
	default:
		return fmt.Sprintf("Mouse(%d)", int(mt))
	}
}

// quitMsg signals the application to exit.
type quitMsg struct{}

// batchMsg carries multiple commands to run in parallel.
type batchMsg []Cmd
