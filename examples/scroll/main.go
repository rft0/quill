package main

import (
	"fmt"
	"os"

	"github.com/rft0/quill"
)

func ScrollDemo(ctx *quill.Context) *quill.Node {
	scroll := quill.UseRef(ctx, quill.ScrollState{})

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		case quill.KeyUp:
			scroll.ScrollUp(1)
		case quill.KeyDown:
			scroll.ScrollDown(1)
		case quill.KeyPageUp:
			scroll.PageUp()
		case quill.KeyPageDown:
			scroll.PageDown()
		case quill.KeyRune:
			switch key.Rune {
			case 'k':
				scroll.ScrollUp(1)
			case 'j':
				scroll.ScrollDown(1)
			}
		}
	})

	// Build scrollable content.
	args := []any{
		quill.FlexColumn,
		quill.Height(quill.Px(10)),
		quill.BorderRounded,
		quill.BorderColor(quill.Cyan),
	}
	for i := 1; i <= 30; i++ {
		color := quill.White
		if i%2 == 0 {
			color = quill.Gray
		}
		args = append(args, quill.Text(fmt.Sprintf("  Line %2d: some content here", i), quill.TextColor(color)))
	}

	return quill.Box(quill.FlexColumn, quill.Gap(1),
		quill.Text("Scroll Demo", quill.Bold, quill.TextColor(quill.Yellow)),
		quill.Text(fmt.Sprintf("Offset: %d", scroll.Offset), quill.TextColor(quill.Gray)),
		quill.ScrollView(scroll, args...),
		quill.Text("j/k or ↑/↓: scroll · pgup/pgdn: page · esc: quit", quill.TextColor(quill.Gray)),
	)
}

func main() {
	if err := quill.New(ScrollDemo).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
