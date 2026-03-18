package main

import (
	"fmt"
	"os"

	"github.com/rft0/quill"
)

func Form(ctx *quill.Context) *quill.Node {
	name := quill.UseRef(ctx, quill.InputState{})
	email := quill.UseRef(ctx, quill.InputState{})
	focus := quill.UseRef(ctx, quill.FocusGroup{})

	quill.UseEffect(ctx, func() {
		*focus = quill.NewFocusGroup(name, email)
	})

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		case quill.KeyTab, quill.KeyEnter:
			focus.Next()
		default:
			focus.Update(key)
		}
	})

	return quill.Box(quill.FlexColumn, quill.BorderRounded, quill.PadXY(1, 1),
		quill.Text("Sign Up", quill.Bold, quill.TextColor(quill.Cyan)),

		quill.Box(quill.MarginTop(1),
			quill.Text("Name: ", quill.TextColor(quill.White), quill.Bold),
			quill.Input(name, quill.TextColor(quill.Yellow)),
		),

		quill.Box(
			quill.Text("Email: ", quill.TextColor(quill.White), quill.Bold),
			quill.Input(email, quill.TextColor(quill.Yellow)),
		),

		quill.Text("TAB to switch · ESC to quit", quill.MarginTop(1), quill.TextColor(quill.Gray)),
	)
}

func main() {
	app := quill.New(Form)
	app.SetCursor(quill.CursorBar)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
