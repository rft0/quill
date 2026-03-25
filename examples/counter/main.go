package main

import (
	"fmt"
	"os"
	"time"

	ll "github.com/rft0/quill"
)

func Counter(ctx *ll.Context) *ll.Node {
	count := ll.UseState(ctx, 0)

	ll.UseInterval(ctx, time.Millisecond*100, func() {
		count.Set(count.Get() + 1)
	})

	return ll.Box(
		ll.Text("Count: ", ll.TextColor(ll.Green), ll.Bold),
		ll.Text(fmt.Sprintf("%d", count.Get()), ll.TextColor(ll.Cyan), ll.Bold),
	)
}

func main() {
	app := ll.New(Counter)
	app.ExitOnCtrlC()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
