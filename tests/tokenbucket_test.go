package tests

import (
	"go-rate-limiter/internal/limiter"
	"sync"
	"testing"
)

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	tb := limiter.NewTokenBucket(5, 1)

	for i := range 5 {
		if !tb.Allow() {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
	// bucket now empty
	if tb.Allow() {
		t.Fatal("expected deny when bucket empty")
	}
}

func TestTokenBucket_StartsFullByDefault(t *testing.T) {
	tb := limiter.NewTokenBucket(10, 1)
	if got := tb.Tokens(); got != 10 {
		t.Fatalf("expected 10 tokens, got %d", got)
	}
}

func TestTokenBucket_RejectWhenEmpty(t *testing.T) {
	tb := limiter.NewTokenBucket(1, 0) // refillRate=0 means no automatic refill in this window
	tb.Allow()                         // consume the 1 token
	if tb.Allow() {
		t.Fatal("expected deny on empty bucket")
	}
}

func TestTokenBucket_ConcurrentAllow(t *testing.T) {
	const capacity = 100
	tb := limiter.NewTokenBucket(capacity, 0)

	var (
		wg      sync.WaitGroup
		allowed int
		mu      sync.Mutex
	)

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed > capacity {
		t.Fatalf("allowed %d > capacity %d: token bucket over-issued", allowed, capacity)
	}
}

func TestTokenBucket_NoPanicOnZeroCapacity(t *testing.T) {
	tb := limiter.NewTokenBucket(0, 0)
	if tb.Allow() {
		t.Fatal("expected deny on zero-capacity bucket")
	}
}
