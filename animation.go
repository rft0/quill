package quill

import (
	"math"
	"time"
)

// SpringOpts configures the spring physics for UseSpring.
type SpringOpts struct {
	Stiffness float64 // spring constant (default 170)
	Damping   float64 // friction coefficient (default 26)
	Mass      float64 // mass (default 1)
}

type tweenState struct {
	current   float64
	start     float64
	target    float64
	startTime time.Time
	animating bool
}

// UseTween smoothly interpolates toward target over duration using
// ease-out cubic easing. Returns the current animated value.
// The animation restarts whenever target changes.
//
//	width := quill.UseTween(ctx, targetWidth, 300*time.Millisecond)
func UseTween(ctx *Context, target float64, duration time.Duration) float64 {
	state := UseRef(ctx, tweenState{current: target, target: target})

	if state.target != target {
		state.start = state.current
		state.target = target
		state.startTime = time.Now()
		state.animating = true
	}

	UseInterval(ctx, 16*time.Millisecond, func() {
		if !state.animating {
			return
		}
		elapsed := time.Since(state.startTime)
		t := float64(elapsed) / float64(duration)
		if t >= 1 {
			t = 1
			state.animating = false
		}
		// Ease-out cubic.
		ease := 1 - math.Pow(1-t, 3)
		state.current = state.start + (state.target-state.start)*ease
	})

	return state.current
}

type springState struct {
	current  float64
	velocity float64
	target   float64
}

// UseSpring animates toward target using spring physics. Returns the
// current animated value. The spring continuously tracks the target
// value with natural-feeling motion.
//
//	y := quill.UseSpring(ctx, targetY, quill.SpringOpts{Stiffness: 200})
func UseSpring(ctx *Context, target float64, opts ...SpringOpts) float64 {
	cfg := SpringOpts{Stiffness: 170, Damping: 26, Mass: 1}
	if len(opts) > 0 {
		cfg = opts[0]
		if cfg.Stiffness == 0 {
			cfg.Stiffness = 170
		}
		if cfg.Damping == 0 {
			cfg.Damping = 26
		}
		if cfg.Mass == 0 {
			cfg.Mass = 1
		}
	}

	state := UseRef(ctx, springState{current: target, target: target})
	state.target = target

	UseInterval(ctx, 16*time.Millisecond, func() {
		dt := 0.016 // 16ms in seconds
		displacement := state.current - state.target
		springForce := -cfg.Stiffness * displacement
		dampingForce := -cfg.Damping * state.velocity
		acceleration := (springForce + dampingForce) / cfg.Mass

		state.velocity += acceleration * dt
		state.current += state.velocity * dt

		// Settle when close enough.
		if math.Abs(state.velocity) < 0.01 && math.Abs(displacement) < 0.01 {
			state.current = state.target
			state.velocity = 0
		}
	})

	return state.current
}
