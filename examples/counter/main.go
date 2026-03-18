package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rft0/quill"
)

func Counter(ctx *quill.Context) *quill.Node {
	count := quill.UseState(ctx, 0)

	quill.UseInterval(ctx, time.Second, func() {
		count.Set(count.Get() + 1)
	})

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		case quill.KeyUp:
			count.Set(count.Get() + 1)
		case quill.KeyDown:
			count.Set(count.Get() - 1)
		case quill.KeyRune:
			switch key.Rune {
			case 'k':
				count.Set(count.Get() + 1)
			case 'j':
				count.Set(count.Get() - 1)
			}
		}
	})

	return quill.Box(quill.FlexRow, quill.AlignCenter, quill.JustifySpaceBetween,
		quill.Box(
			quill.Text("Count: ", quill.TextColor(quill.Green), quill.Bold),
			quill.Text(fmt.Sprintf("%d", count.Get()), quill.TextColor(quill.Cyan), quill.Bold),
		),
		quill.Text("j/k or arrows · ESC to quit", quill.TextColor(quill.Gray)),
	)
}

func main() {
	if err := quill.New(Counter).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
