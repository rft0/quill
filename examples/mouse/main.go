package main

import (
	"fmt"
	"os"

	ll "github.com/rft0/quill"
)

func MouseDemo(ctx *ll.Context) *ll.Node {
	lastEvent := ll.UseState(ctx, "click anywhere...")

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		if key.Type == ll.KeyCtrlC || key.Type == ll.KeyEscape {
			ctx.Quit()
		}
	})

	ll.OnMouse(ctx, func(m ll.MouseMsg) {
		var action string
		switch m.Type {
		case ll.MouseLeft:
			action = "Left click"
		case ll.MouseRight:
			action = "Right click"
		case ll.MouseMiddle:
			action = "Middle click"
		case ll.MouseRelease:
			action = "Release"
		case ll.MouseWheelUp:
			action = "Wheel up"
		case ll.MouseWheelDown:
			action = "Wheel down"
		case ll.MouseMotion:
			action = "Motion"
		}
		lastEvent.Set(fmt.Sprintf("%s at (%d, %d)", action, m.X, m.Y))
	})

	return ll.Box(ll.FlexColumn, ll.Gap(1),
		ll.Text("Mouse Demo", ll.Bold, ll.TextColor(ll.Yellow)),
		ll.Box(ll.FlexRow, ll.Gap(1),
			ll.Text("Last event:", ll.TextColor(ll.Gray)),
			ll.Text(lastEvent.Get(), ll.TextColor(ll.Cyan)),
		),
		ll.Text("click/scroll anywhere · ESC to quit", ll.TextColor(ll.Gray)),
	)
}

func main() {
	app := ll.New(MouseDemo)
	app.EnableMouse()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
