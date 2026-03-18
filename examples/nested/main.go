package main

import (
	"fmt"
	"os"
	"time"

	ll "github.com/rft0/quill"
)

// Timer is a child component — just a function that takes ctx and returns a node.
func Timer(ctx *ll.Context) *ll.Node {
	elapsed := ll.UseState(ctx, 0)

	ll.UseInterval(ctx, time.Second, func() {
		elapsed.Set(elapsed.Get() + 1)
	})

	m := elapsed.Get() / 60
	s := elapsed.Get() % 60

	return ll.Box(ll.FlexRow, ll.Gap(1),
		ll.Text("Elapsed:", ll.Bold),
		ll.Text(fmt.Sprintf("%02d:%02d", m, s), ll.TextColor(ll.Cyan), ll.Bold),
	)
}

// TodoItem is a child component that renders a single todo.
func TodoItem(ctx *ll.Context, text string, done, selected bool) *ll.Node {
	color := ll.Pick(done, ll.TextColor(ll.Green), ll.TextColor(ll.White))

	return ll.Box(ll.FlexRow, ll.Gap(1),
		ll.Pick(selected, ll.Bold, ll.Dim),
		ll.IfElse(done,
			ll.Text("[x]", color),
			ll.Text("[ ]", color),
		),
		ll.Text(text, color),
	)
}

type todo struct {
	text string
	done bool
}

// App is the root component that composes Timer and TodoItem.
func App(ctx *ll.Context) *ll.Node {
	todos := ll.UseState(ctx, []todo{
		{text: "Build a TUI framework"},
		{text: "Add hooks system"},
		{text: "Write nested component example"},
	})
	cursor := ll.UseState(ctx, 0)

	ll.OnKey(ctx, func(key ll.KeyMsg) {
		switch key.Type {
		case ll.KeyCtrlC, ll.KeyEscape:
			ctx.Quit()
		case ll.KeyUp:
			if cursor.Get() > 0 {
				cursor.Set(cursor.Get() - 1)
			}
		case ll.KeyDown:
			if cursor.Get() < len(todos.Get())-1 {
				cursor.Set(cursor.Get() + 1)
			}
		case ll.KeyEnter:
			items := todos.Get()
			items[cursor.Get()].done = !items[cursor.Get()].done
			todos.Set(items)
		}
	})

	items := todos.Get()

	return ll.Box(ll.FlexColumn, ll.BorderRounded, ll.PadXY(2, 1), ll.Gap(1),
		ll.Text("Todo List", ll.Bold, ll.TextColor(ll.Yellow)),
		Timer(ctx),
		ll.Map(items, func(item todo, i int) *ll.Node {
			return TodoItem(ctx, item.text, item.done, i == cursor.Get())
		}),
		ll.Text("↑/↓ to move · ENTER to toggle · ESC to quit", ll.MarginTop(1), ll.TextColor(ll.Gray)),
	)
}

func main() {
	if err := ll.New(App).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
