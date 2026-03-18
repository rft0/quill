// Package quill builds terminal user interfaces with a React-like hooks API
// and a CSS Flexbox layout engine. It renders inline (content-sized) by default,
// or fullscreen using the alternate screen buffer.
//
// # Quick start
//
// Define a component as a function that takes a [Context] and returns a [*Node]:
//
//	func Counter(ctx *quill.Context) *quill.Node {
//	    count := quill.UseState(ctx, 0)
//
//	    quill.OnKey(ctx, func(key quill.KeyMsg) {
//	        switch key.Type {
//	        case quill.KeyUp:
//	            count.Set(count.Get() + 1)
//	        case quill.KeyEscape:
//	            ctx.Quit()
//	        }
//	    })
//
//	    return quill.Box(quill.FlexRow,
//	        quill.Text("Count: ", quill.Bold),
//	        quill.Text(fmt.Sprintf("%d", count.Get()), quill.TextColor(quill.Cyan)),
//	    )
//	}
//
//	func main() {
//	    if err := quill.New(Counter).Run(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// # Layout
//
// [Box] creates a flex container. Pass [FlexRow] or [FlexColumn] for direction,
// plus layout props like [JustifyCenter], [AlignCenter], [Gap], [Padding], and
// [Margin]. Children can be [*Node] values or props — they are sorted out
// automatically.
//
// [Text] creates a styled text leaf. Props like [Bold], [Italic], [TextColor],
// and [BackgroundColor] control appearance.
//
// # Hooks
//
// Hooks must be called in the same order every render (like React):
//
//   - [UseState] — reactive state; calling Set triggers a re-render
//   - [UseRef] — stable mutable pointer that survives re-renders
//   - [UseMemo] — cached computation with dependency tracking
//   - [UseEffect] / [UseEffectWithCleanup] — side effects
//   - [UseInterval] / [UseAfter] — timer hooks
//   - [OnKey] / [OnMouse] / [OnResize] — event handlers
//
// # Widgets
//
//   - [Input] — single-line text input with cursor
//   - [Textarea] — multi-line text input
//   - [Select] — list picker with keyboard navigation
//   - [Checkbox] — toggle with label
//   - [ProgressBar] — horizontal bar (0.0–1.0)
//   - [Spinner] — animated spinner
//   - [ScrollView] — clipped scrollable container
//   - [Table] — column-aligned data table
//   - [FocusGroup] — Tab/Shift+Tab focus cycling
package quill
