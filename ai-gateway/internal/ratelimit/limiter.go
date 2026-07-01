package ratelimit

import (
	"sync"
	"time"
)

// Limiter decides whether a keyed client may proceed with a request.
type Limiter interface {
	Allow(key string) bool
}

// InMemoryLimiter enforces a fixed-window rate limit per key using a mutex-protected map.
type InMemoryLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	clients map[string]*clientWindow
}

type clientWindow struct {
	count       int
	windowStart time.Time
}

// New creates an in-memory limiter allowing limit requests per minute per key.
func New(limit int) *InMemoryLimiter {
	return &InMemoryLimiter{
		limit:   limit,
		window:  time.Minute,
		clients: make(map[string]*clientWindow),
	}
}

// Allow reports whether key may proceed. It uses a fixed one-minute window
// and lazily resets counters when a window expires.
func (l *InMemoryLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cw, ok := l.clients[key]
	if !ok || now.Sub(cw.windowStart) >= l.window {
		l.clients[key] = &clientWindow{count: 1, windowStart: now}
		l.sweepLocked(now)
		return true
	}

	if cw.count >= l.limit {
		return false
	}

	cw.count++
	return true
}

// sweepLocked removes stale entries whose window has expired.
// Must be called while holding l.mu.
func (l *InMemoryLimiter) sweepLocked(now time.Time) {
	for key, cw := range l.clients {
		if now.Sub(cw.windowStart) >= l.window {
			delete(l.clients, key)
		}
	}
}
