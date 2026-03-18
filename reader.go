package quill

import (
	"io"
	"unicode/utf8"
)

// readInput continuously reads from r, parses terminal input into Msgs,
// and sends them on ch. It returns when r is closed or errors.
func readInput(r io.Reader, ch chan<- Msg) {
	buf := make([]byte, 256)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return
		}
		for _, msg := range parseInput(buf[:n]) {
			ch <- msg
		}
	}
}

// parseInput converts raw terminal bytes into a slice of KeyMsg events.
func parseInput(data []byte) []Msg {
	var msgs []Msg
	i := 0
	for i < len(data) {
		b := data[i]

		// --- Escape sequences ---
		if b == 0x1b {
			if i+1 >= len(data) {
				msgs = append(msgs, KeyMsg{Type: KeyEscape})
				i++
				continue
			}
			switch data[i+1] {
			case '[': // CSI sequence
				msg, advance := parseCSI(data[i+2:])
				msgs = append(msgs, msg)
				i += 2 + advance
			case 'O': // SS3 (F1–F4)
				if i+2 < len(data) {
					switch data[i+2] {
					case 'P':
						msgs = append(msgs, KeyMsg{Type: KeyF1})
					case 'Q':
						msgs = append(msgs, KeyMsg{Type: KeyF2})
					case 'R':
						msgs = append(msgs, KeyMsg{Type: KeyF3})
					case 'S':
						msgs = append(msgs, KeyMsg{Type: KeyF4})
					default:
						msgs = append(msgs, KeyMsg{Type: KeyEscape})
					}
					i += 3
				} else {
					msgs = append(msgs, KeyMsg{Type: KeyEscape})
					i += 2
				}
			default: // Alt + key
				msgs = append(msgs, KeyMsg{Type: KeyRune, Rune: rune(data[i+1]), Alt: true})
				i += 2
			}
			continue
		}

		// --- Control characters and ASCII ---
		switch {
		case b == 0x0D:
			msgs = append(msgs, KeyMsg{Type: KeyEnter})
			// Skip LF following CR (Windows CRLF) so Enter isn't doubled.
			if i+1 < len(data) && data[i+1] == 0x0A {
				i++
			}
		case b == 0x0A:
			msgs = append(msgs, KeyMsg{Type: KeyEnter})
		case b == 0x09:
			msgs = append(msgs, KeyMsg{Type: KeyTab})
		case b == 0x08 || b == 0x7F:
			msgs = append(msgs, KeyMsg{Type: KeyBackspace})
		case b == 0x20:
			msgs = append(msgs, KeyMsg{Type: KeySpace, Rune: ' '})
		case b >= 0x01 && b <= 0x1A:
			msgs = append(msgs, KeyMsg{Type: ctrlKey(b)})
		case b >= 0x21 && b <= 0x7E:
			msgs = append(msgs, KeyMsg{Type: KeyRune, Rune: rune(b)})
		default:
			// UTF-8 multi-byte
			r, size := utf8.DecodeRune(data[i:])
			if r != utf8.RuneError && size > 0 {
				msgs = append(msgs, KeyMsg{Type: KeyRune, Rune: r})
				i += size
				continue
			}
		}
		i++
	}
	return msgs
}

// parseCSI parses a CSI sequence (bytes after "\x1b[") and returns
// the decoded Msg plus how many bytes were consumed.
func parseCSI(data []byte) (Msg, int) {
	if len(data) == 0 {
		return KeyMsg{Type: KeyEscape}, 0
	}

	// SGR mouse: \x1b[<btn;col;row;M or m
	if data[0] == '<' {
		msg, advance := parseSGRMouse(data[1:])
		return msg, 1 + advance
	}

	// Simple single-letter sequences (arrows, Home, End).
	switch data[0] {
	case 'A':
		return KeyMsg{Type: KeyUp}, 1
	case 'B':
		return KeyMsg{Type: KeyDown}, 1
	case 'C':
		return KeyMsg{Type: KeyRight}, 1
	case 'D':
		return KeyMsg{Type: KeyLeft}, 1
	case 'H':
		return KeyMsg{Type: KeyHome}, 1
	case 'F':
		return KeyMsg{Type: KeyEnd}, 1
	case 'Z':
		return KeyMsg{Type: KeyShiftTab}, 1
	}

	// Numeric parameter sequences: <number> ~
	num := 0
	j := 0
	for j < len(data) && data[j] >= '0' && data[j] <= '9' {
		num = num*10 + int(data[j]-'0')
		j++
	}

	if j < len(data) && data[j] == '~' {
		switch num {
		case 1:
			return KeyMsg{Type: KeyHome}, j + 1
		case 2:
			return KeyMsg{Type: KeyInsert}, j + 1
		case 3:
			return KeyMsg{Type: KeyDelete}, j + 1
		case 4:
			return KeyMsg{Type: KeyEnd}, j + 1
		case 5:
			return KeyMsg{Type: KeyPageUp}, j + 1
		case 6:
			return KeyMsg{Type: KeyPageDown}, j + 1
		case 15:
			return KeyMsg{Type: KeyF5}, j + 1
		case 17:
			return KeyMsg{Type: KeyF6}, j + 1
		case 18:
			return KeyMsg{Type: KeyF7}, j + 1
		case 19:
			return KeyMsg{Type: KeyF8}, j + 1
		case 20:
			return KeyMsg{Type: KeyF9}, j + 1
		case 21:
			return KeyMsg{Type: KeyF10}, j + 1
		case 23:
			return KeyMsg{Type: KeyF11}, j + 1
		case 24:
			return KeyMsg{Type: KeyF12}, j + 1
		}
	}

	return KeyMsg{Type: KeyEscape}, j
}

// parseSGRMouse parses an SGR mouse sequence (bytes after "\x1b[<").
// Format: btn;col;rowM (press) or btn;col;rowm (release).
func parseSGRMouse(data []byte) (Msg, int) {
	// Parse three semicolon-separated numbers followed by M or m.
	nums := [3]int{}
	ni := 0
	j := 0
	for j < len(data) {
		b := data[j]
		if b >= '0' && b <= '9' {
			nums[ni] = nums[ni]*10 + int(b-'0')
		} else if b == ';' {
			ni++
			if ni > 2 {
				return KeyMsg{Type: KeyEscape}, j + 1
			}
		} else if b == 'M' || b == 'm' {
			btn := nums[0]
			x := nums[1] - 1 // 1-based to 0-based
			y := nums[2] - 1
			if x < 0 {
				x = 0
			}
			if y < 0 {
				y = 0
			}

			var mt MouseType
			if b == 'm' {
				mt = MouseRelease
			} else {
				switch {
				case btn&64 != 0:
					if btn&1 != 0 {
						mt = MouseWheelDown
					} else {
						mt = MouseWheelUp
					}
				case btn&32 != 0:
					mt = MouseMotion
				case btn&3 == 0:
					mt = MouseLeft
				case btn&3 == 1:
					mt = MouseMiddle
				case btn&3 == 2:
					mt = MouseRight
				default:
					mt = MouseLeft
				}
			}

			return MouseMsg{Type: mt, X: x, Y: y}, j + 1
		} else {
			break
		}
		j++
	}
	return KeyMsg{Type: KeyEscape}, j
}

// ctrlKey maps a control byte (0x01–0x1A) to its KeyType.
// Bytes that alias common keys (0x08=BS, 0x09=Tab, 0x0A=LF, 0x0D=CR)
// are handled before this function is called.
func ctrlKey(b byte) KeyType {
	switch b {
	case 0x01:
		return KeyCtrlA
	case 0x02:
		return KeyCtrlB
	case 0x03:
		return KeyCtrlC
	case 0x04:
		return KeyCtrlD
	case 0x05:
		return KeyCtrlE
	case 0x06:
		return KeyCtrlF
	case 0x07:
		return KeyCtrlG
	case 0x0B:
		return KeyCtrlK
	case 0x0C:
		return KeyCtrlL
	case 0x0E:
		return KeyCtrlN
	case 0x0F:
		return KeyCtrlO
	case 0x10:
		return KeyCtrlP
	case 0x11:
		return KeyCtrlQ
	case 0x12:
		return KeyCtrlR
	case 0x13:
		return KeyCtrlS
	case 0x14:
		return KeyCtrlT
	case 0x15:
		return KeyCtrlU
	case 0x16:
		return KeyCtrlV
	case 0x17:
		return KeyCtrlW
	case 0x18:
		return KeyCtrlX
	case 0x19:
		return KeyCtrlY
	case 0x1A:
		return KeyCtrlZ
	default:
		return KeyRune
	}
}
