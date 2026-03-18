package main

import (
	"fmt"
	"os"
	"time"

	"github.com/rft0/quill"
)

// Timer is a child component — just a function that takes ctx and returns a node.
func Timer(ctx *quill.Context) *quill.Node {
	elapsed := quill.UseState(ctx, 0)

	quill.UseInterval(ctx, time.Second, func() {
		elapsed.Set(elapsed.Get() + 1)
	})

	m := elapsed.Get() / 60
	s := elapsed.Get() % 60

	return quill.Box(quill.FlexRow, quill.Gap(1),
		quill.Text("Elapsed:", quill.Bold),
		quill.Text(fmt.Sprintf("%02d:%02d", m, s), quill.TextColor(quill.Cyan), quill.Bold),
	)
}

// TodoItem is a child component that renders a single todo.
func TodoItem(ctx *quill.Context, text string, done bool) *quill.Node {
	check := "[ ]"
	color := quill.White
	if done {
		check = "[x]"
		color = quill.Green
	}

	return quill.Box(quill.FlexRow, quill.Gap(1),
		quill.Text(check, quill.TextColor(color), quill.Bold),
		quill.Text(text, quill.TextColor(color)),
	)
}

type todo struct {
	text string
	done bool
}

// App is the root component that composes Timer and TodoItem.
func App(ctx *quill.Context) *quill.Node {
	todos := quill.UseState(ctx, []todo{
		{text: "Build a TUI framework"},
		{text: "Add hooks system"},
		{text: "Write nested component example"},
	})
	cursor := quill.UseState(ctx, 0)

	quill.OnKey(ctx, func(key quill.KeyMsg) {
		switch key.Type {
		case quill.KeyCtrlC, quill.KeyEscape:
			ctx.Quit()
		case quill.KeyUp:
			if cursor.Get() > 0 {
				cursor.Set(cursor.Get() - 1)
			}
		case quill.KeyDown:
			if cursor.Get() < len(todos.Get())-1 {
				cursor.Set(cursor.Get() + 1)
			}
		case quill.KeySpace, quill.KeyEnter:
			items := todos.Get()
			items[cursor.Get()].done = !items[cursor.Get()].done
			todos.Set(items)
		}
	})

	// Build todo list items using the child component.
	items := todos.Get()
	todoNodes := make([]any, len(items))
	for i, item := range items {
		node := TodoItem(ctx, item.text, item.done)
		if i == cursor.Get() {
			node.Paint.Reverse = true
		}
		todoNodes[i] = node
	}

	// Compose: timer on top, todo list below, help text at bottom.
	children := []any{
		quill.FlexColumn, quill.BorderRounded, quill.PadXY(2, 1), quill.Gap(1),
		quill.Text("Todo List", quill.Bold, quill.TextColor(quill.Yellow)),
		Timer(ctx),
	}
	children = append(children, todoNodes...)
	children = append(children,
		quill.Text("↑/↓ move · space toggle · esc quit", quill.MarginTop(1), quill.TextColor(quill.Gray)),
	)

	return quill.Box(children...)
}

func main() {
	if err := quill.New(App).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
