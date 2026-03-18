package main

import (
	"fmt"
	"os"

	ll "github.com/rft0/quill"
)

func Form(ctx *ll.Context) *ll.Node {
	name := ll.UseRef(ctx, ll.InputState{})
	email := ll.UseRef(ctx, ll.InputState{})
	focus := ll.UseRef(ctx, ll.FocusGroup{})

	ll.UseEffect(ctx, func() {
		*focus = ll.NewFocusGroup(name, email)
	})

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		switch key.Type {
		case ll.KeyCtrlC, ll.KeyEscape:
			ctx.Quit()
		case ll.KeyTab, ll.KeyEnter:
			focus.Next()
		default:
			focus.Update(key)
		}
	})

	return ll.Box(ll.FlexColumn, ll.BorderRounded, ll.PadXY(1, 1),
		ll.Text("Sign Up", ll.Bold, ll.TextColor(ll.Cyan)),

		ll.Box(ll.MarginTop(1),
			ll.Text("Name: ", ll.TextColor(ll.White), ll.Bold),
			ll.Input(name, ll.TextColor(ll.Yellow)),
		),

		ll.Box(
			ll.Text("Email: ", ll.TextColor(ll.White), ll.Bold),
			ll.Input(email, ll.TextColor(ll.Yellow)),
		),

		ll.Text("TAB to switch · ESC to quit", ll.MarginTop(1), ll.TextColor(ll.Gray)),
	)
}

func main() {
	app := ll.New(Form)
	app.SetCursor(ll.CursorBar)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
