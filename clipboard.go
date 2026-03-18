package quill

import (
	"encoding/base64"
	"fmt"
	"os"
)

// CopyToClipboard returns a Cmd that writes text to the system clipboard
// using the OSC 52 escape sequence. This is supported by most modern
// terminal emulators (iTerm2, WezTerm, Kitty, Alacritty, Windows Terminal,
// tmux with set-clipboard on, etc.).
//
// Use with ctx.Exec:
//
//	ctx.Exec(quill.CopyToClipboard("hello"))
func CopyToClipboard(text string) Cmd {
	return func() Msg {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		fmt.Fprintf(os.Stdout, "\x1b]52;c;%s\x07", encoded)
		return nil
	}
}
