package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rft0/quill"
)

func App(ctx *quill.Context) *quill.Node {
	clock := quill.UseState(ctx, "")
	frame := quill.UseState(ctx, 0)
	width := quill.UseState(ctx, 0)
	height := quill.UseState(ctx, 0)

	quill.UseInterval(ctx, time.Second, func() {
		clock.Set(time.Now().Format("15:04:05"))
	})

	quill.UseInterval(ctx, 80*time.Millisecond, func() {
		frame.Set((frame.Get() + 1) % len(quill.SpinnerDots))
	})

	quill.OnResize(ctx, func(w, h int) {
		width.Set(w)
		height.Set(h)
	})

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		}
	})

	return quill.Box(quill.FlexColumn, quill.JustifyCenter, quill.AlignCenter,
		quill.Grow(1),

		// Title
		quill.Text("Fullscreen Mode", quill.Bold, quill.TextColor(quill.BrightCyan)),

		// Clock
		quill.Box(quill.FlexRow, quill.Gap(1), quill.MarginY(1),
			quill.Spinner(frame.Get(), quill.SpinnerDots, quill.TextColor(quill.Yellow)),
			quill.Text(clock.Get(), quill.TextColor(quill.BrightWhite), quill.Bold),
		),

		// Terminal size
		quill.Text(
			fmt.Sprintf("Terminal: %dx%d", width.Get(), height.Get()),
			quill.TextColor(quill.Gray),
		),

		// Hint
		quill.Box(quill.MarginTop(1),
			quill.Text("ESC to quit", quill.TextColor(quill.Gray), quill.Dim),
		),
	)
}

func main() {
	app := quill.New(App)
	app.Fullscreen()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
