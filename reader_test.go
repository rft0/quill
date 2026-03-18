package quill

import "testing"

func TestParseInputASCII(t *testing.T) {
	msgs := parseInput([]byte("abc"))
	if len(msgs) != 3 {
		t.Fatalf("len = %d, want 3", len(msgs))
	}
	for i, ch := range "abc" {
		key := msgs[i].(KeyMsg)
		if key.Type != KeyRune || key.Rune != ch {
			t.Errorf("msgs[%d] = %+v, want KeyRune %q", i, key, ch)
		}
	}
}

func TestParseInputEnter(t *testing.T) {
	// CR
	msgs := parseInput([]byte{0x0D})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyEnter {
		t.Errorf("got %+v, want KeyEnter", msgs[0])
	}

	// LF
	msgs = parseInput([]byte{0x0A})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyEnter {
		t.Errorf("got %+v, want KeyEnter", msgs[0])
	}

	// CRLF should be single Enter.
	msgs = parseInput([]byte{0x0D, 0x0A})
	if len(msgs) != 1 {
		t.Fatalf("CRLF: len = %d, want 1", len(msgs))
	}
}

func TestParseInputEscape(t *testing.T) {
	msgs := parseInput([]byte{0x1b})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyEscape {
		t.Errorf("got %+v, want KeyEscape", msgs[0])
	}
}

func TestParseInputArrows(t *testing.T) {
	tests := []struct {
		input []byte
		want  KeyType
	}{
		{[]byte("\x1b[A"), KeyUp},
		{[]byte("\x1b[B"), KeyDown},
		{[]byte("\x1b[C"), KeyRight},
		{[]byte("\x1b[D"), KeyLeft},
	}
	for _, tt := range tests {
		msgs := parseInput(tt.input)
		if len(msgs) != 1 {
			t.Errorf("input %q: len = %d, want 1", tt.input, len(msgs))
			continue
		}
		if msgs[0].(KeyMsg).Type != tt.want {
			t.Errorf("input %q: got %+v, want %v", tt.input, msgs[0], tt.want)
		}
	}
}

func TestParseInputCtrl(t *testing.T) {
	msgs := parseInput([]byte{0x03}) // Ctrl+C
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyCtrlC {
		t.Errorf("got %+v, want KeyCtrlC", msgs[0])
	}
}

func TestParseInputFKeys(t *testing.T) {
	tests := []struct {
		input []byte
		want  KeyType
	}{
		{[]byte("\x1bOP"), KeyF1},
		{[]byte("\x1bOQ"), KeyF2},
		{[]byte("\x1bOR"), KeyF3},
		{[]byte("\x1bOS"), KeyF4},
		{[]byte("\x1b[15~"), KeyF5},
		{[]byte("\x1b[17~"), KeyF6},
	}
	for _, tt := range tests {
		msgs := parseInput(tt.input)
		if len(msgs) != 1 {
			t.Errorf("input %v: len = %d, want 1", tt.input, len(msgs))
			continue
		}
		if msgs[0].(KeyMsg).Type != tt.want {
			t.Errorf("input %v: got %+v, want %v", tt.input, msgs[0], tt.want)
		}
	}
}

func TestParseInputAltKey(t *testing.T) {
	msgs := parseInput([]byte{0x1b, 'x'})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	key := msgs[0].(KeyMsg)
	if key.Type != KeyRune || key.Rune != 'x' || !key.Alt {
		t.Errorf("got %+v, want Alt+x", key)
	}
}

func TestParseInputTab(t *testing.T) {
	msgs := parseInput([]byte{0x09})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyTab {
		t.Errorf("got %+v, want KeyTab", msgs[0])
	}
}

func TestParseInputBackspace(t *testing.T) {
	// 0x7F
	msgs := parseInput([]byte{0x7F})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyBackspace {
		t.Errorf("got %+v, want KeyBackspace", msgs[0])
	}

	// 0x08
	msgs = parseInput([]byte{0x08})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(KeyMsg).Type != KeyBackspace {
		t.Errorf("got %+v, want KeyBackspace", msgs[0])
	}
}

func TestParseInputSpace(t *testing.T) {
	msgs := parseInput([]byte{0x20})
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	key := msgs[0].(KeyMsg)
	if key.Type != KeySpace || key.Rune != ' ' {
		t.Errorf("got %+v, want KeySpace with rune ' '", key)
	}
}

func TestParseInputHomeEnd(t *testing.T) {
	tests := []struct {
		input []byte
		want  KeyType
	}{
		{[]byte("\x1b[H"), KeyHome},
		{[]byte("\x1b[F"), KeyEnd},
		{[]byte("\x1b[1~"), KeyHome},
		{[]byte("\x1b[4~"), KeyEnd},
	}
	for _, tt := range tests {
		msgs := parseInput(tt.input)
		if len(msgs) != 1 {
			t.Errorf("input %v: len = %d, want 1", tt.input, len(msgs))
			continue
		}
		if msgs[0].(KeyMsg).Type != tt.want {
			t.Errorf("input %v: got %+v, want %v", tt.input, msgs[0], tt.want)
		}
	}
}

func TestParseInputUTF8(t *testing.T) {
	msgs := parseInput([]byte("é"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	key := msgs[0].(KeyMsg)
	if key.Type != KeyRune || key.Rune != 'é' {
		t.Errorf("got %+v, want KeyRune 'é'", key)
	}
}

func TestParseSGRMouse(t *testing.T) {
	// Left click at (10, 5): \x1b[<0;11;6M
	msgs := parseInput([]byte("\x1b[<0;11;6M"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	m := msgs[0].(MouseMsg)
	if m.Type != MouseLeft {
		t.Errorf("type = %v, want MouseLeft", m.Type)
	}
	if m.X != 10 || m.Y != 5 {
		t.Errorf("pos = (%d,%d), want (10,5)", m.X, m.Y)
	}
}

func TestParseSGRMouseRelease(t *testing.T) {
	// Release: lowercase m
	msgs := parseInput([]byte("\x1b[<0;1;1m"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	m := msgs[0].(MouseMsg)
	if m.Type != MouseRelease {
		t.Errorf("type = %v, want MouseRelease", m.Type)
	}
}

func TestParseSGRMouseWheel(t *testing.T) {
	// Wheel up: btn=64
	msgs := parseInput([]byte("\x1b[<64;1;1M"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(MouseMsg).Type != MouseWheelUp {
		t.Errorf("got %v, want MouseWheelUp", msgs[0].(MouseMsg).Type)
	}

	// Wheel down: btn=65
	msgs = parseInput([]byte("\x1b[<65;1;1M"))
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].(MouseMsg).Type != MouseWheelDown {
		t.Errorf("got %v, want MouseWheelDown", msgs[0].(MouseMsg).Type)
	}
}
