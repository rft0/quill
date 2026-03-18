package quill

import "testing"

// mockContext creates a minimal Context for testing hooks.
func mockContext() *Context {
	return &Context{msgs: make(chan<- Msg, 64)}
}

func TestUseStateGetSet(t *testing.T) {
	ctx := mockContext()

	s := UseState(ctx, 42)
	if s.Get() != 42 {
		t.Errorf("initial = %d, want 42", s.Get())
	}

	s.Set(100)
	if s.Get() != 100 {
		t.Errorf("after Set = %d, want 100", s.Get())
	}
}

func TestUseStatePersistsAcrossRenders(t *testing.T) {
	ctx := mockContext()

	// First render.
	s1 := UseState(ctx, 0)
	s1.Set(5)

	// Simulate re-render.
	ctx.hookIdx = 0
	s2 := UseState(ctx, 0)
	if s2.Get() != 5 {
		t.Errorf("after re-render = %d, want 5", s2.Get())
	}
	if s1 != s2 {
		t.Error("UseState should return same pointer across renders")
	}
}

func TestUseRef(t *testing.T) {
	ctx := mockContext()

	ref := UseRef(ctx, "hello")
	if *ref != "hello" {
		t.Errorf("initial = %q, want %q", *ref, "hello")
	}

	*ref = "world"

	// Re-render: same pointer, mutated value.
	ctx.hookIdx = 0
	ref2 := UseRef(ctx, "hello")
	if *ref2 != "world" {
		t.Errorf("after mutation = %q, want %q", *ref2, "world")
	}
	if ref != ref2 {
		t.Error("UseRef should return same pointer across renders")
	}
}

func TestUseMemo(t *testing.T) {
	ctx := mockContext()
	calls := 0

	v1 := UseMemo(ctx, func() int {
		calls++
		return 42
	}, "dep1")
	if v1 != 42 || calls != 1 {
		t.Errorf("first call: v=%d, calls=%d", v1, calls)
	}

	// Re-render with same deps: should not recompute.
	ctx.hookIdx = 0
	v2 := UseMemo(ctx, func() int {
		calls++
		return 99
	}, "dep1")
	if v2 != 42 || calls != 1 {
		t.Errorf("same deps: v=%d, calls=%d", v2, calls)
	}

	// Re-render with different deps: should recompute.
	ctx.hookIdx = 0
	v3 := UseMemo(ctx, func() int {
		calls++
		return 99
	}, "dep2")
	if v3 != 99 || calls != 2 {
		t.Errorf("changed deps: v=%d, calls=%d", v3, calls)
	}
}

func TestUseEffectMountOnly(t *testing.T) {
	ctx := mockContext()
	calls := 0

	// No deps: runs once.
	UseEffect(ctx, func() { calls++ })
	if calls != 1 {
		t.Errorf("mount: calls = %d, want 1", calls)
	}

	// Re-render: should not run again.
	ctx.hookIdx = 0
	UseEffect(ctx, func() { calls++ })
	if calls != 1 {
		t.Errorf("re-render: calls = %d, want 1", calls)
	}
}

func TestUseEffectWithDeps(t *testing.T) {
	ctx := mockContext()
	calls := 0

	UseEffect(ctx, func() { calls++ }, "a")
	if calls != 1 {
		t.Fatalf("mount: calls = %d, want 1", calls)
	}

	// Same deps: no re-run.
	ctx.hookIdx = 0
	UseEffect(ctx, func() { calls++ }, "a")
	if calls != 1 {
		t.Errorf("same deps: calls = %d, want 1", calls)
	}

	// Changed deps: re-run.
	ctx.hookIdx = 0
	UseEffect(ctx, func() { calls++ }, "b")
	if calls != 2 {
		t.Errorf("changed deps: calls = %d, want 2", calls)
	}
}

func TestUseEffectWithCleanupDeps(t *testing.T) {
	ctx := mockContext()
	runs := 0
	cleanups := 0

	UseEffectWithCleanup(ctx, func() func() {
		runs++
		return func() { cleanups++ }
	}, "a")
	if runs != 1 || cleanups != 0 {
		t.Fatalf("mount: runs=%d, cleanups=%d", runs, cleanups)
	}

	// Same deps: no re-run, no cleanup.
	ctx.hookIdx = 0
	UseEffectWithCleanup(ctx, func() func() {
		runs++
		return func() { cleanups++ }
	}, "a")
	if runs != 1 || cleanups != 0 {
		t.Errorf("same deps: runs=%d, cleanups=%d", runs, cleanups)
	}

	// Changed deps: cleanup old, run new.
	ctx.hookIdx = 0
	UseEffectWithCleanup(ctx, func() func() {
		runs++
		return func() { cleanups++ }
	}, "b")
	if runs != 2 || cleanups != 1 {
		t.Errorf("changed deps: runs=%d, cleanups=%d", runs, cleanups)
	}
}

func TestStopPropagation(t *testing.T) {
	ctx := mockContext()
	ctx.msg = KeyMsg{Type: KeyRune, Rune: 'a'}

	childCalled := false
	parentCalled := false

	// Child handler: consumes event.
	OnKey(ctx, func(key KeyMsg) {
		childCalled = true
		ctx.StopPropagation()
	})

	// Parent handler: should not fire.
	OnKey(ctx, func(key KeyMsg) {
		parentCalled = true
	})

	if !childCalled {
		t.Error("child handler should have been called")
	}
	if parentCalled {
		t.Error("parent handler should NOT have been called after StopPropagation")
	}
	if !ctx.Handled() {
		t.Error("Handled() should return true")
	}
}

func TestStopPropagationMouse(t *testing.T) {
	ctx := mockContext()
	ctx.msg = MouseMsg{Type: MouseLeft, X: 5, Y: 10}

	childCalled := false
	parentCalled := false

	OnMouse(ctx, func(m MouseMsg) {
		childCalled = true
		ctx.StopPropagation()
	})

	OnMouse(ctx, func(m MouseMsg) {
		parentCalled = true
	})

	if !childCalled {
		t.Error("child mouse handler should have been called")
	}
	if parentCalled {
		t.Error("parent mouse handler should NOT have been called")
	}
}

func TestHandledResetBetweenRenders(t *testing.T) {
	ctx := mockContext()
	ctx.msg = KeyMsg{Type: KeyRune, Rune: 'x'}

	OnKey(ctx, func(key KeyMsg) {
		ctx.StopPropagation()
	})
	if !ctx.Handled() {
		t.Fatal("should be handled")
	}

	// Simulate new render cycle.
	ctx.handled = false
	ctx.msg = KeyMsg{Type: KeyRune, Rune: 'y'}

	called := false
	OnKey(ctx, func(key KeyMsg) {
		called = true
	})
	if !called {
		t.Error("handler should fire after handled reset")
	}
}
