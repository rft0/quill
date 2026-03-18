package main

import (
	"fmt"
	"os"

	"github.com/rft0/quill"
)

func MouseDemo(ctx *quill.Context) *quill.Node {
	lastEvent := quill.UseState(ctx, "click anywhere...")

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		if key.Type == quill.KeyCtrlC || key.Type == quill.KeyEscape {
			ctx.Quit()
		}
	})

	quill.OnMouse(ctx, func(m quill.MouseMsg) {
		var action string
		switch m.Type {
		case quill.MouseLeft:
			action = "Left click"
		case quill.MouseRight:
			action = "Right click"
		case quill.MouseMiddle:
			action = "Middle click"
		case quill.MouseRelease:
			action = "Release"
		case quill.MouseWheelUp:
			action = "Wheel up"
		case quill.MouseWheelDown:
			action = "Wheel down"
		case quill.MouseMotion:
			action = "Motion"
		}
		lastEvent.Set(fmt.Sprintf("%s at (%d, %d)", action, m.X, m.Y))
	})

	return quill.Box(quill.FlexColumn, quill.Gap(1),
		quill.Text("Mouse Demo", quill.Bold, quill.TextColor(quill.Yellow)),
		quill.Box(quill.FlexRow, quill.Gap(1),
			quill.Text("Last event:", quill.TextColor(quill.Gray)),
			quill.Text(lastEvent.Get(), quill.TextColor(quill.Cyan)),
		),
		quill.Text("click/scroll anywhere · esc: quit", quill.TextColor(quill.Gray)),
	)
}

func main() {
	app := quill.New(MouseDemo)
	app.EnableMouse()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
