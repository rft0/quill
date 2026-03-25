package quill

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

// App ties together a Component, the terminal, and the event loop.
// By default it renders inline (content-sized). Use [WithFullscreen]
// to take over the entire terminal instead.
type App struct {
	component    Component
	cursorStyle  CursorStyle
	mouse        bool
	fullscreen   bool
	exitOnCtrlC  bool
}

// Option configures an [App]. Pass options to [New].
type Option func(*App)

// WithFullscreen makes the app take over the entire terminal using the
// alternate screen buffer. Without this, the app renders inline.
func WithFullscreen() Option { return func(a *App) { a.fullscreen = true } }

// WithMouse enables mouse event tracking (SGR protocol).
func WithMouse() Option { return func(a *App) { a.mouse = true } }

// WithCursor sets the terminal cursor shape.
func WithCursor(style CursorStyle) Option { return func(a *App) { a.cursorStyle = style } }

// New creates an App that drives the given component.
//
//	quill.New(MyApp, quill.WithFullscreen(), quill.WithMouse()).Run()
func New(c Component, opts ...Option) *App {
	a := &App{component: c}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// ExitOnCtrlC adds a default Ctrl+C handler that quits the app.
func (a *App) ExitOnCtrlC() *App {
	a.exitOnCtrlC = true
	return a
}

// cursorSeq returns the ANSI escape sequence for the configured cursor.
func (a *App) cursorSeq() string {
	switch a.cursorStyle {
	case CursorHidden:
		return "\x1b[?25l"
	case CursorBlockBlink:
		return "\x1b[?25h\x1b[1 q"
	case CursorBlock:
		return "\x1b[?25h\x1b[2 q"
	case CursorUnderlineBlink:
		return "\x1b[?25h\x1b[3 q"
	case CursorUnderline:
		return "\x1b[?25h\x1b[4 q"
	case CursorBarBlink:
		return "\x1b[?25h\x1b[5 q"
	case CursorBar:
		return "\x1b[?25h\x1b[6 q"
	default:
		return "\x1b[?25l"
	}
}

// Run enters raw mode, renders the component, processes input,
// and restores the terminal on exit. In inline mode (default) it
// renders content-sized output. In fullscreen mode it uses the
// alternate screen buffer and takes over the entire terminal.
func (a *App) Run() error {
	inFd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(inFd)
	if err != nil {
		return err
	}
	defer term.Restore(inFd, oldState)

	outFd := int(os.Stdout.Fd())
	w, h, err := term.GetSize(outFd)
	if err != nil {
		return err
	}

	out := os.Stdout

	// Fullscreen: enter alternate screen buffer and clear it.
	if a.fullscreen {
		fmt.Fprint(out, "\x1b[?1049h\x1b[2J\x1b[H")
		defer fmt.Fprint(out, "\x1b[?1049l")
	}

	// Enable mouse tracking if requested.
	if a.mouse {
		fmt.Fprint(out, "\x1b[?1006h\x1b[?1003h")
		defer fmt.Fprint(out, "\x1b[?1006l\x1b[?1003l")
	}

	var prevCanvas *Canvas
	prevLines := 0
	prevWidth := 0

	msgs := make(chan Msg, 64)
	batch := &renderBatch{}
	ctx := &Context{msgs: msgs, batch: batch}

	// Poll for terminal resize.
	go func() {
		lastW, lastH := w, h
		for {
			time.Sleep(100 * time.Millisecond)
			if nw, nh, err := term.GetSize(outFd); err == nil && (nw != lastW || nh != lastH) {
				lastW, lastH = nw, nh
				msgs <- ResizeMsg{Width: nw, Height: nh}
			}
		}
	}()

	render := func(msg Msg) (cmd Cmd) {
		defer func() {
			if r := recover(); r != nil {
				// Show error in the UI instead of crashing.
				errText := fmt.Sprintf("PANIC: %v", r)
				root := Box(FlexColumn,
					Text(errText, TextColor(Red)),
					Text("Press Ctrl+C to exit", TextColor(Yellow)),
				)
				var canvas *Canvas
				if a.fullscreen {
					Compute(root, float64(w), float64(h))
					canvas = NewCanvas(w, h)
				} else {
					contentH := int(math.Ceil(ComputeInline(root, float64(w))))
					if contentH < 1 {
						contentH = 1
					}
					canvas = NewCanvas(w, contentH)
				}
				Render(root, canvas)
				if a.fullscreen {
					a.flushFullscreen(out, canvas, prevCanvas)
				} else {
					a.flushInline(out, canvas, prevCanvas, prevLines, prevWidth)
				}
				prevLines = canvas.Height
				prevWidth = canvas.Width
				prevCanvas = canvas
				cmd = nil
			}
		}()

		ctx.cmd = nil
		ctx.hookIdx = 0
		ctx.msg = msg
		ctx.handled = false

		// Update dimensions on resize.
		if r, ok := msg.(ResizeMsg); ok {
			w = r.Width
			if a.fullscreen {
				h = r.Height
			}
			prevCanvas = nil // force full redraw on resize
		}

		batch.mu.Lock()
		batch.active = true
		batch.pending = false
		batch.mu.Unlock()

		if a.exitOnCtrlC {
			OnKey(ctx, func(key KeyMsg) {
				if key.Type == KeyCtrlC {
					ctx.Quit()
				}
			})
		}
		root := a.component(ctx)
		cmd = ctx.cmd

		batch.mu.Lock()
		pending := batch.pending
		batch.active = false
		batch.mu.Unlock()

		// If Set() was called during render, schedule one re-render.
		if pending {
			go func() {
				select {
				case msgs <- stateMsg{}:
				default:
				}
			}()
		}

		if root == nil {
			return cmd
		}

		var canvas *Canvas
		if a.fullscreen {
			Compute(root, float64(w), float64(h))
			canvas = NewCanvas(w, h)
		} else {
			contentH := int(math.Ceil(ComputeInline(root, float64(w))))
			if contentH < 1 {
				contentH = 1
			}
			canvas = NewCanvas(w, contentH)
		}
		Render(root, canvas)

		if a.fullscreen {
			a.flushFullscreen(out, canvas, prevCanvas)
		} else {
			a.flushInline(out, canvas, prevCanvas, prevLines, prevWidth)
		}
		prevLines = canvas.Height
		prevWidth = canvas.Width
		prevCanvas = canvas
		return cmd
	}

	render(nil)

	go readInput(os.Stdin, msgs)

	for msg := range msgs {
		cmd := render(msg)

		if cmd != nil {
			result := cmd()
			switch v := result.(type) {
			case quitMsg:
				ctx.runCleanups()
				if a.fullscreen {
					// Alt screen buffer restore is handled by defer.
					fmt.Fprint(out, "\x1b[?25h\x1b[0 q")
				} else {
					fmt.Fprint(out, "\x1b[?25h\x1b[0 q\r\n")
				}
				return nil
			case batchMsg:
				for _, c := range v {
					go func(fn Cmd) { msgs <- fn() }(c)
				}
			default:
				go func() { msgs <- result }()
			}
		}
	}

	return nil
}

// flushFullscreen renders a frame to the terminal in fullscreen mode.
// Uses absolute cursor positioning and cell-level diffing against prev.
func (a *App) flushFullscreen(out io.Writer, curr, prev *Canvas) {
	w, h := curr.Width, curr.Height

	// Check if anything changed.
	if prev != nil && prev.Width == w && prev.Height == h {
		changed := false
		for i := range curr.cells {
			if curr.cells[i] != prev.cells[i] {
				changed = true
				break
			}
		}
		if !changed {
			return
		}
	}

	var buf strings.Builder
	buf.WriteString("\x1b[?25l") // hide cursor

	lastFG, lastBG, lastAttrs := ColorDefault, ColorDefault, CellAttrs{}
	needsSGR := true
	prevX, prevY := -2, -1

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := curr.Get(x, y)
			if cell.Wide {
				continue
			}
			if prev != nil && prev.Width == w && prev.Height == h {
				if cell == prev.Get(x, y) {
					needsSGR = true
					continue
				}
			}

			// Move cursor only when not consecutive.
			rw := 1
			if cell.Rune != 0 {
				rw = runeWidth(cell.Rune)
			}
			if x != prevX+1 || y != prevY {
				fmt.Fprintf(&buf, "\x1b[%d;%dH", y+1, x+1)
				needsSGR = true
			}
			prevX, prevY = x+rw-1, y

			if needsSGR || cell.FG != lastFG || cell.BG != lastBG || cell.Attrs != lastAttrs {
				buf.WriteString(sgr(cell.FG, cell.BG, cell.Attrs))
				lastFG, lastBG, lastAttrs = cell.FG, cell.BG, cell.Attrs
				needsSGR = false
			}

			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			buf.WriteRune(r)
		}
	}

	buf.WriteString("\x1b[0m")
	buf.WriteString(a.cursorSeq())

	io.WriteString(out, buf.String()) //nolint:errcheck
}

// flushInline renders a frame to the terminal in inline mode.
// Uses relative cursor movement and cell-level diffing against prev.
func (a *App) flushInline(out io.Writer, curr, prev *Canvas, prevLines, prevWidth int) {
	contentH := curr.Height
	w := curr.Width

	var buf strings.Builder

	// Diff path: same dimensions as previous frame.
	if prev != nil && prevLines == contentH && prev.Width == w {
		changed := false
		for i := range curr.cells {
			if curr.cells[i] != prev.cells[i] {
				changed = true
				break
			}
		}
		if !changed {
			return
		}

		buf.WriteString("\x1b[?25l")

		// Move cursor to top-left of frame.
		if prevLines > 1 {
			fmt.Fprintf(&buf, "\x1b[%dA", prevLines-1)
		}
		buf.WriteString("\r")

		curRow, curCol := 0, 0
		lastFG, lastBG, lastAttrs := ColorDefault, ColorDefault, CellAttrs{}
		needsSGR := true

		for y := 0; y < contentH; y++ {
			for x := 0; x < w; x++ {
				cell := curr.Get(x, y)
				if cell.Wide {
					continue
				}
				p := prev.Get(x, y)
				if cell == p {
					needsSGR = true
					continue
				}

				if y != curRow {
					dy := y - curRow
					if dy > 0 {
						fmt.Fprintf(&buf, "\x1b[%dB", dy)
					} else {
						fmt.Fprintf(&buf, "\x1b[%dA", -dy)
					}
					curRow = y
					needsSGR = true
				}
				if x != curCol {
					fmt.Fprintf(&buf, "\x1b[%dG", x+1)
					curCol = x
					needsSGR = true
				}

				if needsSGR || cell.FG != lastFG || cell.BG != lastBG || cell.Attrs != lastAttrs {
					buf.WriteString(sgr(cell.FG, cell.BG, cell.Attrs))
					lastFG, lastBG, lastAttrs = cell.FG, cell.BG, cell.Attrs
					needsSGR = false
				}

				r := cell.Rune
				if r == 0 {
					r = ' '
				}
				buf.WriteRune(r)
				curCol += runeWidth(r)
			}
		}

		if curRow != contentH-1 {
			fmt.Fprintf(&buf, "\x1b[%dB", contentH-1-curRow)
		}
		fmt.Fprintf(&buf, "\x1b[%dG", w)

		buf.WriteString("\x1b[0m")
		buf.WriteString(a.cursorSeq())

		io.WriteString(out, buf.String()) //nolint:errcheck
		return
	}

	// Full redraw: first frame or height changed.
	if prevLines > 0 {
		// When the terminal shrinks, old lines (rendered at prevWidth) get
		// wrapped by the terminal into ceil(prevWidth/newWidth) terminal lines
		// each. We must move up enough to reach the true start of the old frame.
		wrappedTotal := prevLines
		if prevWidth > 0 && w > 0 && prevWidth > w {
			wrappedTotal = 0
			for i := 0; i < prevLines; i++ {
				wrappedTotal += (prevWidth + w - 1) / w
			}
		}
		if wrappedTotal > 1 {
			fmt.Fprintf(&buf, "\x1b[%dF", wrappedTotal-1)
		}
		buf.WriteString("\r\x1b[J") // erase from cursor to end of screen
	}

	buf.WriteString("\x1b[?25l")

	lastFG, lastBG, lastAttrs := ColorDefault, ColorDefault, CellAttrs{}
	first := true

	for y := 0; y < contentH; y++ {
		for x := 0; x < w; x++ {
			cell := curr.Get(x, y)
			if cell.Wide {
				continue
			}
			if first || cell.FG != lastFG || cell.BG != lastBG || cell.Attrs != lastAttrs {
				buf.WriteString(sgr(cell.FG, cell.BG, cell.Attrs))
				lastFG, lastBG, lastAttrs = cell.FG, cell.BG, cell.Attrs
				first = false
			}
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			buf.WriteRune(r)
		}
		buf.WriteString("\x1b[0m")
		first = true
		if y < contentH-1 {
			buf.WriteString("\r\n")
		}
	}

	buf.WriteString(a.cursorSeq())

	io.WriteString(out, buf.String()) //nolint:errcheck
}
