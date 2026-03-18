package quill

// Component is a function that renders UI. Use hooks (UseState, UseEffect,
// OnKey, etc.) inside it to manage state and side effects.
//
// For struct-based components, pass a method value:
//
//	quill.New(myForm.Render).Run()
type Component func(ctx *Context) *Node

// Context provides hooks and commands for components.
type Context struct {
	cmd         Cmd
	msgs        chan<- Msg
	msg         Msg
	hooks       []any
	hookIdx     int
	handled     bool                // true if a handler has consumed the current event
	batch       *renderBatch        // coalesces Set() calls during render
	subContexts map[string]*Context // keyed sub-contexts for component isolation
	provided    map[any]any         // context values set via ProvideContext
	parent      *Context            // parent context for UseContext lookup chain
}

// StopPropagation marks the current event as handled, preventing
// subsequent OnKey/OnMouse handlers from receiving it.
func (c *Context) StopPropagation() { c.handled = true }

// Handled reports whether the current event has been consumed by
// a prior handler via StopPropagation.
func (c *Context) Handled() bool { return c.handled }

// Quit tells the app to exit after this render.
func (c *Context) Quit() { c.cmd = quitFn() }

// runCleanups calls all UseEffectWithCleanup cleanup functions,
// stops all timer goroutines, and recurses into sub-contexts.
func (c *Context) runCleanups() {
	for _, h := range c.hooks {
		switch v := h.(type) {
		case *effectHook:
			if v.cleanup != nil {
				v.cleanup()
			}
		case *timerHook:
			close(v.stop)
		}
	}
	for _, sub := range c.subContexts {
		sub.runCleanups()
	}
}

// SubContext returns a child context with an isolated hook scope, keyed
// by the given string. Use this when rendering multiple instances of the
// same component (e.g. in a dynamic list) so each instance's hooks are
// preserved independently of call order.
//
//	for _, item := range items {
//	    sub := ctx.SubContext(item.ID)
//	    children = append(children, ItemComponent(sub, item))
//	}
func (c *Context) SubContext(key string) *Context {
	if c.subContexts == nil {
		c.subContexts = make(map[string]*Context)
	}
	sub, exists := c.subContexts[key]
	if !exists {
		sub = &Context{
			msgs:  c.msgs,
			batch: c.batch,
		}
		c.subContexts[key] = sub
	}
	sub.msg = c.msg
	sub.hookIdx = 0
	sub.handled = c.handled
	sub.cmd = c.cmd
	sub.parent = c
	return sub
}

// Render creates a keyed sub-context, calls the component, and propagates
// commands and event handling back to the parent context.
func (c *Context) Render(key string, comp Component) *Node {
	sub := c.SubContext(key)
	result := comp(sub)
	if sub.cmd != nil {
		c.cmd = sub.cmd
	}
	if sub.handled {
		c.handled = true
	}
	return result
}

// Exec schedules a command to run after this render.
func (c *Context) Exec(cmd Cmd) { c.cmd = cmd }

// Send pushes a message into the event loop, triggering a re-render.
func (c *Context) Send(msg Msg) {
	go func() { c.msgs <- msg }()
}

// Cmd is a function that performs an IO operation and returns a Msg.
// The app runs each Cmd in a goroutine. A nil Cmd means "do nothing".
type Cmd func() Msg

func quitFn() Cmd {
	return func() Msg { return quitMsg{} }
}

// Batch combines multiple commands into one. The app will run all of
// them concurrently and deliver their results as separate messages.
func Batch(cmds ...Cmd) Cmd {
	var valid []Cmd
	for _, c := range cmds {
		if c != nil {
			valid = append(valid, c)
		}
	}
	switch len(valid) {
	case 0:
		return nil
	case 1:
		return valid[0]
	default:
		return func() Msg { return batchMsg(valid) }
	}
}
