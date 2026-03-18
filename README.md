# Quill

A Go library for building terminal UIs with a React-like hooks API and CSS Flexbox layout engine. Renders inline (content-sized) by default, or fullscreen.

```
go get github.com/rft0/quill
```

## Quick Start

```go
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
```

## Layout

Build UIs with `Box` (flex container) and `Text` (styled text leaf). Props and children can be freely mixed in the argument list.

```go
ll.Box(ll.FlexColumn, ll.Gap(1), ll.BorderRounded, ll.PadXY(1, 1),
    ll.Text("Hello", ll.Bold, ll.TextColor(ll.Green)),
    ll.Box(ll.FlexRow, ll.JustifySpaceBetween,
        ll.Text("Left"),
        ll.Text("Right"),
    ),
)
```

**Direction:** `FlexRow`, `FlexColumn`

**Wrap:** `FlexWrapWrap` — children flow onto multiple lines when they exceed the main axis

**Alignment:** `JustifyCenter`, `JustifySpaceBetween`, `JustifySpaceAround`, `JustifyFlexStart`, `JustifyFlexEnd`, `AlignCenter`, `AlignStretch`, `AlignFlexStart`, `AlignFlexEnd`

**Sizing:** `Width()`, `Height()`, `MinWidth()`, `MaxWidth()`, `MinHeight()`, `MaxHeight()`, `Grow()`, `Shrink()`, `Basis()` — values via `Px()`, `Pct()`, or `AutoDim()`

**Spacing:** `Gap()`, `Padding()`, `PaddingX()`, `PaddingY()`, `PadXY()`, `Margin()`, `MarginTop()`, `MarginRight()`, `MarginBottom()`, `MarginLeft()`

**Borders:** `BorderSingle`, `BorderDouble`, `BorderRounded`, `BorderThick`, `BorderColor()`

**Text styling:** `Bold`, `Italic`, `Underline`, `Dim`, `Strikethrough`, `Reverse`, `TextColor()`, `BackgroundColor()`

**Colors:** `Black`, `Red`, `Green`, `Yellow`, `Blue`, `Magenta`, `Cyan`, `White`, `Gray`, `BrightRed`...`BrightWhite`, `RGBColor(r, g, b)`, `LerpColor(from, to, t)`, `Gradient(from, to, n)`

**Overflow:** `Ellipsis`, `ClipText`

**Positioning:** `Absolute`, `Left()`, `Top()`, `Right()`, `Bottom()`, `ZIndex()`

**Debug:** `Debug` — draws colored outlines around the node and all descendants

## Conditional Rendering

`If` and `IfElse` keep the declarative flow clean — nil nodes are safely ignored by `Box`.

```go
ll.Box(ll.FlexColumn,
    ll.If(showHeader, ll.Text("Header", ll.Bold)),
    ll.IfElse(loggedIn,
        ll.Text("Welcome back!", ll.TextColor(ll.Green)),
        ll.Text("Please log in", ll.TextColor(ll.Gray)),
    ),
    ll.Text("Always visible"),
)
```

## Hooks

Hooks must be called in the same order every render, just like React.

### UseState

Reactive state that triggers re-renders on change.

```go
count := ll.UseState(ctx, 0)
count.Get()                  // read
count.Set(count.Get() + 1)  // write — triggers re-render
```

### UseRef

Stable mutable pointer that survives re-renders without triggering them.

```go
input := ll.UseRef(ctx, ll.InputState{})
input.Value // direct field access
```

### UseMemo

Cached computation, recomputed when dependencies change.

```go
label := ll.UseMemo(ctx, func() string {
    return fmt.Sprintf("Count: %d", count.Get())
}, count.Get())
```

### UseEffect

Side effects that run once, or again when deps change. Return a cleanup function to run before re-execution and on app exit.

```go
// Without cleanup
ll.UseEffect(ctx, func() {
    go fetchData(id.Get())
}, id.Get())

// With cleanup
ll.UseEffectWithCleanup(ctx, func() func() {
    conn := connect()
    return func() { conn.Close() }
})
```

### UseInterval / UseAfter

Timer hooks for periodic or delayed callbacks.

```go
ll.UseInterval(ctx, time.Second, func() {
    count.Set(count.Get() + 1)
})

ll.UseAfter(ctx, 3*time.Second, func() {
    // runs once after delay
})
```

### UseContext

Share data down the component tree without prop drilling.

```go
// Define a key (typically at package level)
var ThemeKey = ll.NewContextKey("light")

// Provide in a parent component
func App(ctx *ll.Context) *ll.Node {
    ll.ProvideContext(ctx, ThemeKey, "dark")
    return ctx.Render("child", ChildComponent)
}

// Consume in any descendant
func ChildComponent(ctx *ll.Context) *ll.Node {
    theme := ll.UseContext(ctx, ThemeKey) // "dark"
    // ...
}
```

### Event Handlers

```go
ll.OnKey(ctx, func(key ll.KeyMsg) {
    switch key.Type {
    case ll.KeyEnter:
        // handle enter
    case ll.KeyCtrlC:
        ctx.Quit()
    }
})

ll.OnMouse(ctx, func(mouse ll.MouseMsg) {
    // mouse.Type, mouse.X, mouse.Y
})

ll.OnResize(ctx, func(resize ll.ResizeMsg) {
    // resize.Width, resize.Height
})
```

## Widgets

### Input

Single-line text input with cursor and full editing support.

```go
name := ll.UseRef(ctx, ll.InputState{})
ll.Input(name, ll.TextColor(ll.Yellow))
```

### Textarea

Multi-line text input.

### Select

List picker with j/k/arrow navigation.

```go
sel := ll.UseRef(ctx, ll.SelectState{
    Options: []string{"Option A", "Option B", "Option C"},
    Focused: true,
})
ll.Select(sel, ll.TextColor(ll.Cyan))
```

### Checkbox

Toggle with label.

```go
check := ll.UseRef(ctx, ll.CheckboxState{Checked: true})
ll.Checkbox(check, "Enable feature", ll.TextColor(ll.Green))
```

### ProgressBar

Horizontal progress bar (0.0 to 1.0), fills available width.

```go
ll.ProgressBar(0.75, ll.TextColor(ll.White))
```

### Spinner

Animated spinner with built-in frame sets (`SpinnerDots`, `SpinnerLine`, `SpinnerBlock`).

```go
frame := ll.UseState(ctx, 0)
ll.UseInterval(ctx, 80*time.Millisecond, func() {
    frame.Set((frame.Get() + 1) % len(ll.SpinnerDots))
})
ll.Spinner(frame.Get(), ll.SpinnerDots, ll.TextColor(ll.Cyan))
```

### ScrollView

Clipped scrollable container.

```go
scroll := ll.UseRef(ctx, ll.ScrollState{})
ll.ScrollView(scroll, ll.Height(ll.Px(10)),
    // children...
)
```

### Table

Column-aligned data table with borders.

```go
ll.Table(
    []string{"Name", "Age", "City"},
    [][]string{
        {"Alice", "30", "NYC"},
        {"Bob", "25", "LA"},
    },
)
```

### FocusGroup

Manages Tab/Shift+Tab focus cycling across multiple inputs.

```go
name := ll.UseRef(ctx, ll.InputState{})
email := ll.UseRef(ctx, ll.InputState{})
focus := ll.UseRef(ctx, ll.FocusGroup{})

ll.UseEffect(ctx, func() {
    *focus = ll.NewFocusGroup(name, email)
})

ll.OnKey(ctx, func(key ll.KeyMsg) {
    switch key.Type {
    case ll.KeyTab:
        focus.Next()
    default:
        focus.Update(key)
    }
})
```

### List

Virtualized scrollable list for large datasets — only renders visible items.

```go
items := []string{"Item 1", "Item 2", /* ... thousands ... */}
list := ll.UseRef(ctx, ll.ListState{Focused: true})

ll.OnKey(ctx, func(key ll.KeyMsg) { list.Update(key) })

ll.List(list, 10, len(items), func(i int, selected bool) *ll.Node {
    n := ll.Text(items[i], ll.TextColor(ll.White))
    if selected {
        n.Paint.Reverse = true
    }
    return n
})
```

### Notify

Absolutely positioned toast notification at the top-right. Control visibility with `If` and timing with `UseAfter`.

```go
show := ll.UseState(ctx, true)
ll.UseAfter(ctx, 3*time.Second, func() { show.Set(false) })

ll.If(show.Get(), ll.Notify(
    ll.Text("Saved!", ll.TextColor(ll.Green)),
    ll.BorderRounded, ll.PadXY(1, 0),
))
```

### Modal

Absolutely positioned overlay that centers content and renders above other children.

```go
ll.Box(ll.FlexColumn, ll.Width(60), ll.Height(20),
    ll.Text("Background content"),
    ll.If(showModal, ll.Modal(
        ll.Box(ll.BorderRounded, ll.PadXY(2, 1), ll.BackgroundColor(ll.Black),
            ll.Text("Are you sure?", ll.Bold),
        ),
        ll.BackgroundColor(ll.RGBColor(0, 0, 0)),
    )),
)
```

### FocusBorderColor

Returns the active or inactive border color based on focus state.

```go
ll.Box(ll.BorderRounded, ll.FocusBorderColor(name.Focused, ll.Cyan, ll.Gray),
    ll.Input(name, ll.TextColor(ll.Yellow)),
)
```

### Gradient Colors

Interpolate between colors for visual effects.

```go
// Single interpolation
mid := ll.LerpColor(ll.Red, ll.Blue, 0.5)

// Generate a gradient palette
colors := ll.Gradient(ll.RGBColor(255, 0, 0), ll.RGBColor(0, 0, 255), 20)
for i, char := range "Hello gradient world!" {
    children = append(children, ll.Text(string(char), ll.TextColor(colors[i])))
}
```

## App Configuration

```go
app := ll.New(MyComponent)
app.SetCursor(ll.CursorBar)   // cursor style: CursorBar, CursorBlock, CursorUnderline
app.EnableMouse()              // enable mouse event tracking
app.Fullscreen()               // use alternate screen buffer
if err := app.Run(); err != nil {
    log.Fatal(err)
}
```

## Examples

Run the included examples:

```bash
go run ./examples/counter/    # Counter with state and intervals
go run ./examples/form/       # Form inputs with focus management
go run ./examples/widgets/    # Spinner, progress bar, checkbox, select
go run ./examples/nested/     # Component composition
go run ./examples/scroll/     # Scrollable container
go run ./examples/mouse/      # Mouse event handling
go run ./examples/fullscreen/ # Fullscreen mode
```

## License

MIT
