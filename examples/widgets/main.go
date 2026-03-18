package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rft0/quill"
)

func Widgets(ctx *quill.Context) *quill.Node {
	// Spinner state
	frame := quill.UseState(ctx, 0)
	quill.UseInterval(ctx, 80*time.Millisecond, func() {
		frame.Set((frame.Get() + 1) % len(quill.SpinnerDots))
	})

	// Progress bar state
	progress := quill.UseState(ctx, 0.0)
	quill.UseInterval(ctx, 200*time.Millisecond, func() {
		v := progress.Get() + 0.01
		if v > 1 {
			v = 0
		}
		progress.Set(v)
	})

	// Checkbox state
	check1 := quill.UseRef(ctx, quill.CheckboxState{Checked: true})
	check2 := quill.UseRef(ctx, quill.CheckboxState{})

	// Select state
	sel := quill.UseRef(ctx, quill.SelectState{
		Options: []string{"Option A", "Option B", "Option C", "Option D"},
		Focused: true,
	})

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		case quill.KeySpace:
			check1.Toggle()
		default:
			sel.Update(key)
		}
	})

	return quill.Box(quill.FlexColumn, quill.Gap(1),
		quill.Text("Widget Showcase", quill.Bold, quill.TextColor(quill.Yellow)),

		// Spinner
		quill.Box(quill.FlexRow, quill.Gap(1),
			quill.Spinner(frame.Get(), quill.SpinnerDots, quill.TextColor(quill.Cyan)),
			quill.Text("Loading...", quill.TextColor(quill.Gray)),
		),

		// Progress bar
		quill.Box(quill.FlexColumn,
			quill.Text(fmt.Sprintf("Progress: %.0f%%", progress.Get()*100), quill.TextColor(quill.White)),
			quill.ProgressBar(progress.Get(), quill.TextColor(quill.White)),
		),

		// Checkboxes
		quill.Box(quill.FlexColumn,
			quill.Text("Checkboxes:", quill.Bold),
			quill.Checkbox(check1, "Enable feature", quill.TextColor(quill.Green)),
			quill.Checkbox(check2, "Dark mode", quill.TextColor(quill.Blue)),
		),

		// Select
		quill.Box(quill.FlexColumn,
			quill.Text("Select (j/k to navigate):", quill.Bold),
			quill.Select(sel, quill.TextColor(quill.Cyan)),
		),

		quill.Text("space: toggle checkbox · j/k: select · esc: quit", quill.TextColor(quill.Gray)),
	)
}

func main() {
	if err := quill.New(Widgets).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
