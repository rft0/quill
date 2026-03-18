package main

import (
	"fmt"
	"os"

	ll "github.com/rft0/quill"
)

func App(ctx *ll.Context) *ll.Node {
	ll.OnKey(ctx, func(key ll.KeyMsg) {
		if key.Type == ll.KeyEscape || key.Type == ll.KeyCtrlC {
			ctx.Quit()
		}
	})

	return ll.Box(ll.BorderRounded,
		ll.Text("Hello, World!", ll.TextColor(ll.White), ll.Bold),
	)
}

func main() {
	if err := ll.New(App).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
