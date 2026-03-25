package main

import (
	"fmt"
	"os"

	ll "github.com/rft0/quill"
)

func Form(ctx *ll.Context) *ll.Node {
	name := ll.UseRef(ctx, ll.InputState{})
	password := ll.UseRef(ctx, ll.InputState{Hidden: true})
	submitted := ll.UseState(ctx, false)

	ll.UseForm(ctx, ll.FormConfig{
		Fields: []ll.Focusable{name, password},
		OnSubmit: func() {
			submitted.Set(true)
		},
	})

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		if key.Type == ll.KeyCtrlC || key.Type == ll.KeyEscape {
			ctx.Quit()
		}
	})

	if submitted.Get() {
		return ll.Box(ll.FlexColumn,
			ll.Text(fmt.Sprintf("Welcome, %s!", name.Value), ll.TextColor(ll.Green), ll.Bold),
		)
	}

	return ll.Box(ll.Title("Sign Up"), ll.BorderColor(ll.Blue), ll.FlexColumn, ll.BorderRounded, ll.PadXY(1, 1),
		ll.Box(
			ll.Text("Username: ", ll.TextColor(ll.White), ll.Bold),
			ll.Input(name, ll.TextColor(ll.Yellow)),
		),

		ll.Box(
			ll.Text("Password: ", ll.TextColor(ll.White), ll.Bold),
			ll.Input(password, ll.TextColor(ll.Yellow)),
		),

		ll.Text("TAB to switch · ENTER to submit · ESC to quit", ll.MarginTop(1), ll.TextColor(ll.Gray)),
	)
}

func main() {
	app := ll.New(Form)
	app.ExitOnCtrlC()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
