package quill

import (
	"sync"
	"time"
)

// stateMsg triggers a re-render when state changes from a goroutine.
type stateMsg struct{}

// tickMsg is sent by UseInterval/UseAfter timers.
type tickMsg struct{ id int }

// renderBatch coalesces multiple Set() calls during a synchronous render
// into a single re-render message.
type renderBatch struct {
	mu      sync.Mutex
	active  bool // true while inside the component function
	pending bool // true if Set() was called during render
}

// --- UseState ---

// State is a reactive state container. Calling Set from a goroutine
// (e.g. inside UseEffect) automatically triggers a re-render.
type State[T any] struct {
	mu    sync.Mutex
	val   T
	msgs  chan<- Msg
	batch *renderBatch
}

// Get returns the current value.
func (s *State[T]) Get() T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.val
}

// Set updates the value and triggers a re-render.
// Multiple Set() calls during a single render are batched into one re-render.
func (s *State[T]) Set(v T) {
	s.mu.Lock()
	s.val = v
	s.mu.Unlock()

	if s.batch != nil {
		s.batch.mu.Lock()
		if s.batch.active {
			s.batch.pending = true
			s.batch.mu.Unlock()
			return
		}
		s.batch.mu.Unlock()
	}

	select {
	case s.msgs <- stateMsg{}:
	default: // re-render already pending
	}
}

// UseState creates or retrieves a reactive state value.
// Returns the same *State on every render.
//
//	count := quill.UseState(ctx, 0)
//	count.Get()    // read
//	count.Set(5)   // write + re-render
func UseState[T any](ctx *Context, initial T) *State[T] {
	idx := ctx.hookIdx
	ctx.hookIdx++
	if idx < len(ctx.hooks) {
		return ctx.hooks[idx].(*State[T])
	}
	s := &State[T]{val: initial, msgs: ctx.msgs, batch: ctx.batch}
	ctx.hooks = append(ctx.hooks, s)
	return s
}

// --- UseRef ---

type ref[T any] struct{ val T }

// UseRef returns a stable pointer that persists across renders.
// Unlike UseState, mutating a ref does not trigger a re-render.
//
//	name := quill.UseRef(ctx, quill.InputState{})
//	name.Focus() // direct pointer access
func UseRef[T any](ctx *Context, initial T) *T {
	idx := ctx.hookIdx
	ctx.hookIdx++
	if idx < len(ctx.hooks) {
		return &ctx.hooks[idx].(*ref[T]).val
	}
	r := &ref[T]{val: initial}
	ctx.hooks = append(ctx.hooks, r)
	return &r.val
}

// --- UseMemo ---

type memo[T any] struct {
	val  T
	deps []any
}

// UseMemo returns a memoized value that is only recomputed when deps change.
//
//	label := quill.UseMemo(ctx, func() string {
//	    return fmt.Sprintf("Count: %d", count.Get())
//	}, count.Get())
func UseMemo[T any](ctx *Context, compute func() T, deps ...any) T {
	idx := ctx.hookIdx
	ctx.hookIdx++

	if idx < len(ctx.hooks) {
		m := ctx.hooks[idx].(*memo[T])
		if depsEqual(m.deps, deps) {
			return m.val
		}
		m.val = compute()
		m.deps = deps
		return m.val
	}

	m := &memo[T]{val: compute(), deps: deps}
	ctx.hooks = append(ctx.hooks, m)
	return m.val
}

func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- UseEffect ---

// effectHook stores UseEffect state including optional cleanup and deps.
type effectHook struct {
	cleanup func()
	deps    []any
	hasDeps bool // true if deps were provided (distinguishes nil deps from no deps)
}

// UseEffect runs fn on the first render and again whenever deps change.
// With no deps, fn runs only once (mount). With deps, fn re-runs when
// any dep value changes (compared with !=).
//
//	// Run once:
//	quill.UseEffect(ctx, func() { go fetchData() })
//
//	// Re-run when id changes:
//	quill.UseEffect(ctx, func() { go fetchData(id.Get()) }, id.Get())
func UseEffect(ctx *Context, fn func(), deps ...any) {
	idx := ctx.hookIdx
	ctx.hookIdx++
	if idx >= len(ctx.hooks) {
		ctx.hooks = append(ctx.hooks, &effectHook{deps: deps, hasDeps: len(deps) > 0})
		fn()
		return
	}
	eh := ctx.hooks[idx].(*effectHook)
	if eh.hasDeps && !depsEqual(eh.deps, deps) {
		eh.deps = deps
		fn()
	}
}

// UseEffectWithCleanup runs fn on the first render and again whenever deps
// change. fn returns a cleanup function that is called before re-running
// and when the app exits.
//
//	quill.UseEffectWithCleanup(ctx, func() func() {
//	    conn := connect()
//	    return func() { conn.Close() }
//	})
//
//	// With deps:
//	quill.UseEffectWithCleanup(ctx, func() func() {
//	    sub := subscribe(id.Get())
//	    return func() { sub.Close() }
//	}, id.Get())
func UseEffectWithCleanup(ctx *Context, fn func() func(), deps ...any) {
	idx := ctx.hookIdx
	ctx.hookIdx++
	if idx >= len(ctx.hooks) {
		cleanup := fn()
		ctx.hooks = append(ctx.hooks, &effectHook{cleanup: cleanup, deps: deps, hasDeps: len(deps) > 0})
		return
	}
	eh := ctx.hooks[idx].(*effectHook)
	if eh.hasDeps && !depsEqual(eh.deps, deps) {
		if eh.cleanup != nil {
			eh.cleanup()
		}
		eh.deps = deps
		eh.cleanup = fn()
	}
}

// --- UseInterval / UseAfter ---

// timerHook stores a timer's ID and stop channel for cleanup.
type timerHook struct {
	id   int
	stop chan struct{}
}

// UseInterval calls fn every d duration. The callback runs synchronously
// during render, so it is safe to modify state directly. The timer
// goroutine is stopped when the app exits.
//
//	quill.UseInterval(ctx, time.Second, func() {
//	    count.Set(count.Get() + 1)
//	})
func UseInterval(ctx *Context, d time.Duration, fn func()) {
	idx := ctx.hookIdx
	ctx.hookIdx++

	if idx >= len(ctx.hooks) {
		th := &timerHook{id: idx, stop: make(chan struct{})}
		ctx.hooks = append(ctx.hooks, th)
		msgs := ctx.msgs
		go func() {
			ticker := time.NewTicker(d)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					select {
					case msgs <- tickMsg{id: idx}:
					case <-th.stop:
						return
					}
				case <-th.stop:
					return
				}
			}
		}()
	}

	if tick, ok := ctx.msg.(tickMsg); ok && tick.id == idx {
		fn()
	}
}

// UseAfter calls fn once after d duration. The callback runs synchronously
// during render. The timer goroutine is stopped when the app exits.
func UseAfter(ctx *Context, d time.Duration, fn func()) {
	idx := ctx.hookIdx
	ctx.hookIdx++

	if idx >= len(ctx.hooks) {
		th := &timerHook{id: idx, stop: make(chan struct{})}
		ctx.hooks = append(ctx.hooks, th)
		msgs := ctx.msgs
		go func() {
			select {
			case <-time.After(d):
				select {
				case msgs <- tickMsg{id: idx}:
				case <-th.stop:
				}
			case <-th.stop:
			}
		}()
	}

	if tick, ok := ctx.msg.(tickMsg); ok && tick.id == idx {
		fn()
	}
}

// --- UseContext ---

// ContextKey identifies a context value. Create one per shared value type
// using [NewContextKey].
type ContextKey[T any] struct {
	defaultVal T
}

// NewContextKey creates a key for sharing a value of type T down the
// component tree. The default value is returned by [UseContext] when no
// ancestor has provided a value.
//
//	var ThemeKey = quill.NewContextKey("light")
func NewContextKey[T any](defaultVal T) *ContextKey[T] {
	return &ContextKey[T]{defaultVal: defaultVal}
}

// ProvideContext stores a value that descendant components can read with
// [UseContext]. Call this before ctx.Render() so children see the value.
//
//	quill.ProvideContext(ctx, ThemeKey, "dark")
func ProvideContext[T any](ctx *Context, key *ContextKey[T], value T) {
	if ctx.provided == nil {
		ctx.provided = make(map[any]any)
	}
	ctx.provided[key] = value
}

// UseContext reads a value provided by an ancestor via [ProvideContext].
// Returns the key's default value if no ancestor has provided one.
//
//	theme := quill.UseContext(ctx, ThemeKey)
func UseContext[T any](ctx *Context, key *ContextKey[T]) T {
	for c := ctx; c != nil; c = c.parent {
		if c.provided != nil {
			if val, ok := c.provided[key]; ok {
				return val.(T)
			}
		}
	}
	return key.defaultVal
}

// --- OnKey ---

// OnKey calls fn if the current event is a key press and the event
// has not been consumed by a prior handler (via ctx.StopPropagation).
//
//	quill.OnKey(ctx, func(key quill.KeyMsg) {
//	    if key.Type == quill.KeyEscape { ctx.Quit() }
//	})
func OnKey(ctx *Context, fn func(KeyMsg)) {
	if ctx.handled {
		return
	}
	if key, ok := ctx.msg.(KeyMsg); ok {
		fn(key)
	}
}

// OnMouse calls fn if the current event is a mouse event and the event
// has not been consumed by a prior handler (via ctx.StopPropagation).
//
//	quill.OnMouse(ctx, func(m quill.MouseMsg) {
//	    if m.Type == quill.MouseLeft { ... }
//	})
func OnMouse(ctx *Context, fn func(MouseMsg)) {
	if ctx.handled {
		return
	}
	if m, ok := ctx.msg.(MouseMsg); ok {
		fn(m)
	}
}

// OnResize calls fn if the current event is a terminal resize.
//
//	quill.OnResize(ctx, func(w, h int) { ... })
func OnResize(ctx *Context, fn func(width, height int)) {
	if r, ok := ctx.msg.(ResizeMsg); ok {
		fn(r.Width, r.Height)
	}
}
