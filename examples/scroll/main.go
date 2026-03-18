package main

import (
	"fmt"
	"os"

	ll "github.com/rft0/quill"
)

func ScrollDemo(ctx *ll.Context) *ll.Node {
	scroll := ll.UseRef(ctx, ll.ScrollState{})

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		switch key.Type {
		case ll.KeyCtrlC, ll.KeyEscape:
			ctx.Quit()
		case ll.KeyUp:
			scroll.ScrollUp(1)
		case ll.KeyDown:
			scroll.ScrollDown(1)
		case ll.KeyPageUp:
			scroll.PageUp()
		case ll.KeyPageDown:
			scroll.PageDown()
		}
	})

	// Build scrollable content.
	args := []any{
		ll.FlexColumn,
		ll.Height(ll.Px(10)),
		ll.BorderRounded,
		ll.BorderColor(ll.Cyan),
	}
	for i := 1; i <= 30; i++ {
		color := ll.White
		if i%2 == 0 {
			color = ll.Gray
		}
		args = append(args, ll.Text(fmt.Sprintf("  Line %2d: some content here", i), ll.TextColor(color)))
	}

	return ll.Box(ll.FlexColumn, ll.Gap(1),
		ll.Text("Scroll Demo", ll.Bold, ll.TextColor(ll.Yellow)),
		ll.Text(fmt.Sprintf("Offset: %d", scroll.Offset), ll.TextColor(ll.Gray)),
		ll.ScrollView(scroll, args...),
		ll.Text("↑/↓ to scroll · pgup/pgdn to change page · ESC to quit", ll.TextColor(ll.Gray)),
	)
}

func main() {
	if err := ll.New(ScrollDemo).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
