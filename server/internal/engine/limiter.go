package engine

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// hostLimiter is a per-host token bucket that backs off when the host pushes back and
// recovers slowly afterwards. A tripped circuit pauses the host outright.
type hostLimiter struct {
	mu       sync.Mutex
	limiters map[string]*hostState
	ceiling  rate.Limit
	burst    int
	jitter   time.Duration
}

type hostState struct {
	lim       *rate.Limiter
	current   rate.Limit
	successes int
	failures  int
	pausedTil time.Time
}

const (
	recoverAfterSuccesses = 20
	recoverFactor         = 1.10
	breakerFailures       = 20
	breakerPause          = 60 * time.Second
)

func newHostLimiter(rps float64, burst int, jitter time.Duration) *hostLimiter {
	if burst < 1 {
		burst = 1
	}
	return &hostLimiter{
		limiters: map[string]*hostState{},
		ceiling:  rate.Limit(rps),
		burst:    burst,
		jitter:   jitter,
	}
}

func (h *hostLimiter) state(host string) *hostState {
	h.mu.Lock()
	defer h.mu.Unlock()
	s, ok := h.limiters[host]
	if !ok {
		s = &hostState{lim: rate.NewLimiter(h.ceiling, h.burst), current: h.ceiling}
		h.limiters[host] = s
	}
	return s
}

// Wait blocks until the host's budget allows a request, then adds jitter so concurrent
// workers do not fire in lockstep.
func (h *hostLimiter) Wait(ctx context.Context, host string) error {
	s := h.state(host)
	h.mu.Lock()
	pause := time.Until(s.pausedTil)
	h.mu.Unlock()
	if pause > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}
	if err := s.lim.Wait(ctx); err != nil {
		return err
	}
	if h.jitter > 0 {
		d := time.Duration(rand.Int64N(int64(h.jitter)))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d):
		}
	}
	return nil
}

// Throttle halves the host rate after push-back. retryAfter, when the server supplied one,
// also pauses the host for that long.
func (h *hostLimiter) Throttle(host string, retryAfter time.Duration) {
	s := h.state(host)
	h.mu.Lock()
	defer h.mu.Unlock()
	s.successes = 0
	s.current /= 2
	if s.current < rate.Limit(0.1) {
		s.current = rate.Limit(0.1)
	}
	s.lim.SetLimit(s.current)
	if retryAfter > 0 {
		s.pausedTil = time.Now().Add(retryAfter)
	}
}

// Succeed nudges the rate back towards the ceiling after a run of clean responses.
func (h *hostLimiter) Succeed(host string) {
	s := h.state(host)
	h.mu.Lock()
	defer h.mu.Unlock()
	s.failures = 0
	s.successes++
	if s.successes < recoverAfterSuccesses || s.current >= h.ceiling {
		return
	}
	s.successes = 0
	s.current = min(rate.Limit(float64(s.current)*recoverFactor), h.ceiling)
	s.lim.SetLimit(s.current)
}

// Fail trips the per-host circuit breaker after enough consecutive failures.
func (h *hostLimiter) Fail(host string) bool {
	s := h.state(host)
	h.mu.Lock()
	defer h.mu.Unlock()
	s.failures++
	if s.failures < breakerFailures {
		return false
	}
	s.failures = 0
	s.pausedTil = time.Now().Add(breakerPause)
	return true
}

// Rate reports the current effective limit for a host.
func (h *hostLimiter) Rate(host string) float64 {
	s := h.state(host)
	h.mu.Lock()
	defer h.mu.Unlock()
	return float64(s.current)
}
