package quill

import (
	"math"
	"strings"
)

// RenderToString renders a node tree to a plain text string for testing.
// It computes inline layout at the given width and returns the canvas
// content without ANSI escape codes.
func RenderToString(root *Node, width int) string {
	h := int(math.Ceil(ComputeInline(root, float64(width))))
	if h < 1 {
		h = 1
	}
	canvas := NewCanvas(width, h)
	Render(root, canvas)
	return canvasToString(canvas)
}

// RenderToCanvas renders a node tree to a Canvas for detailed inspection.
func RenderToCanvas(root *Node, width int) *Canvas {
	h := int(math.Ceil(ComputeInline(root, float64(width))))
	if h < 1 {
		h = 1
	}
	canvas := NewCanvas(width, h)
	Render(root, canvas)
	return canvas
}

// canvasToString converts a Canvas to plain text, trimming trailing spaces
// per line. Wide continuation cells are skipped.
func canvasToString(c *Canvas) string {
	var buf strings.Builder
	for y := 0; y < c.Height; y++ {
		line := strings.Builder{}
		for x := 0; x < c.Width; x++ {
			cell := c.Get(x, y)
			if cell.Wide {
				continue
			}
			r := cell.Rune
			if r == 0 {
				r = ' '
			}
			line.WriteRune(r)
		}
		row := strings.TrimRight(line.String(), " ")
		buf.WriteString(row)
		if y < c.Height-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// TestApp is a test harness for components. It renders without a terminal
// and allows programmatic input injection.
type TestApp struct {
	component Component
	ctx       *Context
	msgs      chan Msg
	width     int
	root      *Node
}

// NewTestApp creates a test harness that renders the component at the
// given width. The component is rendered once immediately.
func NewTestApp(c Component, width int) *TestApp {
	msgs := make(chan Msg, 64)
	batch := &renderBatch{}
	ctx := &Context{msgs: msgs, batch: batch}
	ta := &TestApp{
		component: c,
		ctx:       ctx,
		msgs:      msgs,
		width:     width,
	}
	ta.render(nil)
	return ta
}

func (ta *TestApp) render(msg Msg) {
	ta.ctx.hookIdx = 0
	ta.ctx.msg = msg
	ta.ctx.handled = false
	ta.ctx.cmd = nil

	ta.ctx.batch.mu.Lock()
	ta.ctx.batch.active = true
	ta.ctx.batch.pending = false
	ta.ctx.batch.mu.Unlock()

	ta.root = ta.component(ta.ctx)

	ta.ctx.batch.mu.Lock()
	pending := ta.ctx.batch.pending
	ta.ctx.batch.active = false
	ta.ctx.batch.mu.Unlock()

	// If Set() was called during render, re-render to pick up the new state.
	if pending {
		ta.render(stateMsg{})
	}
}

// SendKey simulates a key press and re-renders.
func (ta *TestApp) SendKey(key KeyMsg) {
	ta.render(key)
	ta.drainState()
}

// SendMouse simulates a mouse event and re-renders.
func (ta *TestApp) SendMouse(mouse MouseMsg) {
	ta.render(mouse)
	ta.drainState()
}

// Send sends any message and re-renders.
func (ta *TestApp) Send(msg Msg) {
	ta.render(msg)
	ta.drainState()
}

func (ta *TestApp) drainState() {
	for {
		select {
		case msg := <-ta.msgs:
			if _, ok := msg.(stateMsg); ok {
				ta.render(msg)
			}
		default:
			return
		}
	}
}

// Output returns the current render as a plain text string.
func (ta *TestApp) Output() string {
	if ta.root == nil {
		return ""
	}
	return RenderToString(ta.root, ta.width)
}

// Canvas returns the current render as a Canvas for cell-level inspection.
func (ta *TestApp) Canvas() *Canvas {
	if ta.root == nil {
		return NewCanvas(ta.width, 1)
	}
	return RenderToCanvas(ta.root, ta.width)
}
