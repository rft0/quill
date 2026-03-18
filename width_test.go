package quill

import "testing"

func TestRuneWidthASCII(t *testing.T) {
	for r := 'a'; r <= 'z'; r++ {
		if w := runeWidth(r); w != 1 {
			t.Errorf("runeWidth(%q) = %d, want 1", r, w)
		}
	}
	if w := runeWidth(' '); w != 1 {
		t.Errorf("runeWidth(' ') = %d, want 1", w)
	}
}

func TestRuneWidthCJK(t *testing.T) {
	wide := []rune{'中', '文', '日', '本', '語', '한', '글'}
	for _, r := range wide {
		if w := runeWidth(r); w != 2 {
			t.Errorf("runeWidth(%q) = %d, want 2", r, w)
		}
	}
}

func TestRuneWidthEmoji(t *testing.T) {
	emoji := []rune{'🎉', '🚀', '💡', '🔥'}
	for _, r := range emoji {
		if w := runeWidth(r); w != 2 {
			t.Errorf("runeWidth(%q) = %d, want 2", r, w)
		}
	}
}

func TestRuneWidthCombining(t *testing.T) {
	combining := []rune{0x0300, 0x0301, 0x0302, 0x200B, 0x200D, 0xFE0F}
	for _, r := range combining {
		if w := runeWidth(r); w != 0 {
			t.Errorf("runeWidth(U+%04X) = %d, want 0", r, w)
		}
	}
}

func TestRuneWidthNull(t *testing.T) {
	if w := runeWidth(0); w != 0 {
		t.Errorf("runeWidth(0) = %d, want 0", w)
	}
}

func TestStringWidth(t *testing.T) {
	tests := []struct {
		s    string
		want int
	}{
		{"hello", 5},
		{"", 0},
		{"中文", 4},
		{"a中b", 4},
		{"🎉🚀", 4},
		{"hi🔥", 4},
	}
	for _, tt := range tests {
		if got := stringWidth(tt.s); got != tt.want {
			t.Errorf("stringWidth(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

func TestRuneWidthFullwidth(t *testing.T) {
	// Fullwidth Latin letters (FF01–FF60)
	if w := runeWidth('Ａ'); w != 2 {
		t.Errorf("runeWidth('Ａ') = %d, want 2", w)
	}
}

func TestRuneWidthHangul(t *testing.T) {
	if w := runeWidth('가'); w != 2 {
		t.Errorf("runeWidth('가') = %d, want 2", w)
	}
	if w := runeWidth('힣'); w != 2 {
		t.Errorf("runeWidth('힣') = %d, want 2", w)
	}
}
