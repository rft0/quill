package main

import (
	"fmt"
	"os"
	"time"

	ll "github.com/rft0/quill"
)

func Widgets(ctx *ll.Context) *ll.Node {
	// Progress bar state
	progress := ll.UseState(ctx, 0.0)
	ll.UseInterval(ctx, 200*time.Millisecond, func() {
		v := progress.Get() + 0.01
		if v > 1 {
			v = 0
		}
		progress.Set(v)
	})

	// Checkbox state
	check1 := ll.UseRef(ctx, ll.CheckboxState{Checked: true})
	check2 := ll.UseRef(ctx, ll.CheckboxState{})

	// Select state
	sel := ll.UseRef(ctx, ll.SelectState{
		Options: []string{"Option A", "Option B", "Option C", "Option D"},
	})

	// Focus group: Tab/Shift+Tab cycles between checkboxes and select
	focus := ll.UseFocusGroup(ctx, check1, check2, sel)

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		switch key.Type {
		case ll.KeyCtrlC, ll.KeyEscape:
			ctx.Quit()
		case ll.KeyTab:
			focus.Next()
		case ll.KeyShiftTab:
			focus.Prev()
		case ll.KeyDown:
			if sel.Focused && sel.Selected < len(sel.Options)-1 {
				sel.Selected++
			} else {
				focus.Next()
			}
		case ll.KeyUp:
			if sel.Focused && sel.Selected > 0 {
				sel.Selected--
			} else {
				focus.Prev()
			}
		default:
			focus.Update(key)
		}
	})

	return ll.Box(ll.FlexColumn, ll.Gap(1),
		ll.Text("Widget Showcase", ll.Bold, ll.TextColor(ll.Yellow)),

		// Spinner
		ll.Box(ll.FlexRow, ll.Gap(1),
			ll.Spinner(ctx, ll.SpinnerDots, ll.TextColor(ll.Cyan)),
			ll.Text("Loading...", ll.TextColor(ll.Gray)),
		),

		// Progress bar
		ll.Box(ll.FlexColumn,
			ll.Text(fmt.Sprintf("Progress: %.0f%%", progress.Get()*100), ll.TextColor(ll.White)),
			ll.ProgressBar(progress.Get(), ll.TextColor(ll.White)),
		),

		// Checkboxes
		ll.Box(ll.FlexColumn,
			ll.Text("Checkboxes:", ll.Bold),
			ll.Checkbox(check1, "Enable feature", ll.TextColor(ll.Green)),
			ll.Checkbox(check2, "Dark mode", ll.TextColor(ll.Blue)),
		),

		// Select
		ll.Box(ll.FlexColumn,
			ll.Text("Select:", ll.Bold),
			ll.Select(sel, ll.TextColor(ll.Cyan)),
		),

		ll.Text("↑/↓ to navigate · ESC to quit", ll.TextColor(ll.Gray)),
	)
}

func main() {
	if err := ll.New(Widgets).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
