# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go test ./...              # Run all tests
go test -v -run TestName . # Run a single test
go vet ./...               # Check for issues
go fmt ./...               # Format code

# Run examples
go run ./examples/counter/
go run ./examples/form/
go run ./examples/nested/
go run ./examples/widgets/
go run ./examples/scroll/
go run ./examples/mouse/
```

## Architecture

Quill is a Go library (`package quill`) for building terminal UIs with a React-like hooks API. It renders inline (content-sized, not full-screen). One external dependency: `golang.org/x/term` for raw mode.

### Layers (bottom-up)

1. **Layout engine** (`style.go`, `node.go`, `compute.go`) — CSS Flexbox. `Node` is the tree element, `Style` holds flex properties. `Compute()` for fixed-size, `ComputeInline()` for content-sized height. 5-phase algorithm: base sizing → flex grow/shrink → cross-axis → justify → absolute coords. Supports `FlexWrapWrap` for multi-line wrapping.

2. **Rendering** (`canvas.go`, `render.go`, `screen.go`) — `Canvas` is a 2D cell grid. `Render()` walks the tree painting text, borders, backgrounds. Supports clipping (`clipRect`) for scroll views. `sgr()` emits ANSI SGR sequences (bold, dim, italic, underline, strikethrough, reverse, 16/24-bit color).

3. **Declarative API** (`element.go`) — `Box(args...)` and `Text(content, args...)` constructors. Style enums (`FlexColumn`, `BorderRounded`, etc.) and prop functions (`TextColor()`, `Padding()`, etc.) implement the `prop` interface. `*Node` children and props can be freely mixed in args. `If()`/`IfElse()` for conditional rendering. `Debug` prop for layout visualization. `FocusBorderColor()` for focus-aware borders.

4. **Event system** (`event.go`, `reader.go`) — `Msg` interface with `KeyMsg`, `MouseMsg`, `ResizeMsg`. `reader.go` parses ANSI escape sequences including SGR mouse protocol (`\x1b[<btn;col;rowM`).

5. **Component model** (`component.go`, `hooks.go`, `app.go`) — React-like functional components:
   - `Component` is `func(ctx *Context) *Node`
   - Hooks follow React's call-order rules (same order every render)
   - `UseState[T]` — reactive state, `Set()` triggers re-render (thread-safe via mutex)
   - `UseRef[T]` — stable mutable pointer, no re-render on change
   - `UseMemo[T]` — cached computation with identity-based deps (`!=`)
   - `UseEffect` / `UseEffectWithCleanup` — side effects, cleanup runs on quit
   - `UseInterval` / `UseAfter` — timer hooks
   - `OnKey` / `OnMouse` / `OnResize` — event handlers (check `ctx.msg` type internally)
   - `UseContext[T]` / `ProvideContext[T]` — share data down tree without prop drilling
   - Context methods: `Quit()`, `Exec(cmd)`, `Send(msg)`
   - Cursor hidden by default; use `SetCursor()` for inputs

6. **Widgets** (`widgets.go`, `input.go`):
   - `Input` — text input with cursor, focus/blur, full editing (uses `UseRef` for `InputState`)
   - `Select` — list picker with j/k/↑/↓ navigation (`SelectState`)
   - `Checkbox` — toggle with label (`CheckboxState`)
   - `ProgressBar` — 0.0–1.0 value, fills available width (`FlexGrow=1`, sentinel `"__progress__"`)
   - `Spinner` — frame-based animation (dots/line/block frame sets)
   - `ScrollView` — clipped scrollable container (`ScrollState` with `ScrollUp`/`Down`/`PageUp`/`PageDown`)
   - `FocusGroup` — manages Tab/Shift+Tab focus cycling across inputs
   - `Modal` — absolutely positioned centered overlay with optional backdrop
   - `List` — virtualized scrollable list for large datasets (`ListState`)
   - `Notify` — absolutely positioned toast notification at top-right

### Key design decisions

- Hooks are package-level functions (`quill.UseState`, not `ctx.UseState`) because Go doesn't support generic methods on structs
- `UseState.Set()` sends a `stateMsg` to the message channel to trigger re-render; uses mutex for goroutine safety
- `UseInterval`/`UseAfter` use `tickMsg` with unique IDs (hook index) to route ticks to the correct hook
- Internal message types (`stateMsg`, `tickMsg`) are hidden from users — only `OnKey`/`OnMouse`/`OnResize` are exposed
- ProgressBar uses sentinel text `"__progress__"` detected at render time to generate the bar dynamically based on available width
