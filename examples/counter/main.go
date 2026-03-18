package main

import (
	"fmt"
	"os"
	"time"

	ll "github.com/rft0/quill"
)

func Counter(ctx *ll.Context) *ll.Node {
	count := ll.UseState(ctx, 0)

	ll.UseInterval(ctx, time.Millisecond*10, func() {
		count.Set(count.Get() + 1)
	})

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		if key.Type == ll.KeyEscape || key.Type == ll.KeyCtrlC {
			ctx.Quit()
		}
	})

	return ll.Box(ll.Width(ll.Pct(100)), ll.FlexRow, ll.AlignCenter, ll.JustifySpaceBetween,
		ll.Box(
			ll.Text("Count: ", ll.TextColor(ll.Green), ll.Bold),
			ll.Text(fmt.Sprintf("%d", count.Get()), ll.TextColor(ll.Cyan), ll.Bold),
		),
		ll.Text("ESC to quit", ll.TextColor(ll.Gray)),
	)
}

func main() {
	if err := ll.New(Counter).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
