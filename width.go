package quill

// runeWidth returns the display width of a rune in terminal cells.
// CJK characters and most emoji are 2 cells wide; combining marks are 0.
func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}

	// Combining characters: zero width.
	if isCombining(r) {
		return 0
	}

	// Wide characters: CJK, fullwidth forms, emoji.
	if isWide(r) {
		return 2
	}

	return 1
}

// stringWidth returns the total display width of s in terminal cells.
func stringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

// truncateToWidth returns the longest prefix of s that fits within maxW
// terminal cells, respecting multi-byte and wide characters.
func truncateToWidth(s string, maxW int) string {
	w := 0
	for i, r := range s {
		rw := runeWidth(r)
		if w+rw > maxW {
			return s[:i]
		}
		w += rw
	}
	return s
}

// isCombining returns true for Unicode combining marks (Mn, Mc, Me categories)
// and other zero-width characters.
func isCombining(r rune) bool {
	switch {
	// Combining Diacritical Marks
	case r >= 0x0300 && r <= 0x036F:
		return true
	// Combining Diacritical Marks Extended
	case r >= 0x1AB0 && r <= 0x1AFF:
		return true
	// Combining Diacritical Marks Supplement
	case r >= 0x1DC0 && r <= 0x1DFF:
		return true
	// Combining Diacritical Marks for Symbols
	case r >= 0x20D0 && r <= 0x20FF:
		return true
	// Combining Half Marks
	case r >= 0xFE20 && r <= 0xFE2F:
		return true
	// Zero-width space, joiner, non-joiner
	case r == 0x200B || r == 0x200C || r == 0x200D || r == 0xFEFF:
		return true
	// Variation selectors
	case r >= 0xFE00 && r <= 0xFE0F:
		return true
	case r >= 0xE0100 && r <= 0xE01EF:
		return true
	}
	return false
}

// isWide returns true for characters that occupy 2 terminal cells.
func isWide(r rune) bool {
	switch {
	// CJK Unified Ideographs
	case r >= 0x4E00 && r <= 0x9FFF:
		return true
	// CJK Unified Ideographs Extension A
	case r >= 0x3400 && r <= 0x4DBF:
		return true
	// CJK Unified Ideographs Extension B–F
	case r >= 0x20000 && r <= 0x2FA1F:
		return true
	// CJK Compatibility Ideographs
	case r >= 0xF900 && r <= 0xFAFF:
		return true
	// Hangul Syllables
	case r >= 0xAC00 && r <= 0xD7AF:
		return true
	// Hangul Jamo
	case r >= 0x1100 && r <= 0x115F:
		return true
	case r >= 0x2329 && r <= 0x232A:
		return true
	// CJK Radicals Supplement, Kangxi Radicals
	case r >= 0x2E80 && r <= 0x2FFF:
		return true
	// CJK Symbols and Punctuation, Hiragana, Katakana
	case r >= 0x3000 && r <= 0x33FF:
		return true
	// Bopomofo, CJK Compatibility
	case r >= 0x3100 && r <= 0x31FF:
		return true
	// Katakana Phonetic Extensions
	case r >= 0x31F0 && r <= 0x31FF:
		return true
	// Enclosed CJK Letters and Months
	case r >= 0x3200 && r <= 0x32FF:
		return true
	// CJK Compatibility Forms
	case r >= 0xFE30 && r <= 0xFE4F:
		return true
	// Fullwidth Forms (excluding halfwidth)
	case r >= 0xFF01 && r <= 0xFF60:
		return true
	case r >= 0xFFE0 && r <= 0xFFE6:
		return true
	// Emoji that are typically wide
	case r >= 0x1F300 && r <= 0x1F9FF:
		return true
	case r >= 0x1FA00 && r <= 0x1FA6F:
		return true
	case r >= 0x1FA70 && r <= 0x1FAFF:
		return true
	// Misc Symbols and Pictographs
	case r >= 0x1F600 && r <= 0x1F64F:
		return true
	// Regional indicators (flags)
	case r >= 0x1F1E0 && r <= 0x1F1FF:
		return true
	}
	return false
}
