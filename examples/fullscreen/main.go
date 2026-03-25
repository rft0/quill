package main

import (
	"fmt"
	"os"
	"time"

	ll "github.com/rft0/quill"
)

func App(ctx *ll.Context) *ll.Node {
	clock := ll.UseState(ctx, "")
	width := ll.UseState(ctx, 0)
	height := ll.UseState(ctx, 0)

	ll.UseInterval(ctx, time.Second, func() {
		clock.Set(time.Now().Format("15:04:05"))
	})

	ll.OnResize(ctx, func(w, h int) {
		width.Set(w)
		height.Set(h)
	})

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		switch key.Type {
		case ll.KeyCtrlC, ll.KeyEscape:
			ctx.Quit()
		}
	})

	return ll.Box(ll.FlexColumn, ll.JustifyCenter, ll.AlignCenter,
		ll.Grow(1),

		// Title
		ll.Text("Fullscreen Mode", ll.Bold, ll.TextColor(ll.BrightCyan)),

		// Clock
		ll.Box(ll.FlexRow, ll.Gap(1), ll.MarginY(1),
			ll.Spinner(ctx, ll.SpinnerDots, ll.TextColor(ll.Yellow)),
			ll.Text(clock.Get(), ll.TextColor(ll.BrightWhite), ll.Bold),
		),

		// Terminal size
		ll.Text(
			fmt.Sprintf("Terminal: %dx%d", width.Get(), height.Get()),
			ll.TextColor(ll.Gray),
		),

		// Hint
		ll.Box(ll.MarginTop(1),
			ll.Text("ESC to quit", ll.TextColor(ll.Gray), ll.Dim),
		),
	)
}

func main() {
	if err := ll.New(App, ll.WithFullscreen()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
