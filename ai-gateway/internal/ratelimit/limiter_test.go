package ratelimit_test

import (
	"testing"

	"github.com/kayodeayelegun/ai-gateway/internal/ratelimit"
)

func TestInMemoryLimiter_allowsUpToLimit(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(3)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("client-a") {
			t.Fatalf("request %d: expected allow, got deny", i+1)
		}
	}
}

func TestInMemoryLimiter_blocksOverLimit(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(3)

	for i := 0; i < 3; i++ {
		if !limiter.Allow("client-a") {
			t.Fatalf("request %d: expected allow, got deny", i+1)
		}
	}

	if limiter.Allow("client-a") {
		t.Fatal("expected deny after limit exceeded")
	}
}

func TestInMemoryLimiter_isolatesKeys(t *testing.T) {
	t.Parallel()

	limiter := ratelimit.New(1)

	if !limiter.Allow("client-a") {
		t.Fatal("expected client-a first request to be allowed")
	}
	if limiter.Allow("client-a") {
		t.Fatal("expected client-a second request to be denied")
	}
	if !limiter.Allow("client-b") {
		t.Fatal("expected client-b to have its own limit bucket")
	}
}
