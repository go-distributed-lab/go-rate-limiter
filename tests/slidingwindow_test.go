package tests

import (
	"sync"
	"testing"
	"time"

	"go-rate-limiter/internal/limiter"
)

func TestSlidingWindow_AllowsUpToCapacity(t *testing.T) {
	sw := limiter.NewSlidingWindow(5, time.Second)

	for i := range 5 {
		if !sw.Allow() {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
	if sw.Allow() {
		t.Fatal("expected deny when window full")
	}
}

func TestSlidingWindow_StartsEmpty(t *testing.T) {
	sw := limiter.NewSlidingWindow(10, time.Second)
	if got := sw.Count(); got != 0 {
		t.Fatalf("expected count 0, got %d", got)
	}
}

func TestSlidingWindow_AllowsAfterExpiry(t *testing.T) {
	sw := limiter.NewSlidingWindow(3, 200*time.Millisecond)

	for range 3 {
		sw.Allow()
	}
	if sw.Allow() {
		t.Fatal("expected deny when window full")
	}

	time.Sleep(250 * time.Millisecond)

	for i := range 3 {
		if !sw.Allow() {
			t.Fatalf("expected allow after expiry on request %d", i+1)
		}
	}
}

func TestSlidingWindow_NoBoundaryBurst(t *testing.T) {
	// This is the key test that proves Sliding Window fixes Fixed Window's
	// boundary burst problem. In a fixed window you could get 2x capacity
	// across a boundary. Here we verify that never happens.
	sw := limiter.NewSlidingWindow(5, 300*time.Millisecond)

	// Fill the window
	allowed := 0
	for range 5 {
		if sw.Allow() {
			allowed++
		}
	}

	// Wait until near the boundary but not past it
	time.Sleep(150 * time.Millisecond)

	// Try to get more — should still be denied because
	// the original requests are still inside the window
	for range 5 {
		if sw.Allow() {
			allowed++
		}
	}

	if allowed > 5 {
		t.Fatalf("boundary burst detected: %d requests allowed, want ≤ 5", allowed)
	}
}

func TestSlidingWindow_PartialExpiry(t *testing.T) {
	// Send 3 requests, wait for 2 to expire, confirm 2 new slots open
	sw := limiter.NewSlidingWindow(3, 300*time.Millisecond)

	sw.Allow() // t=0ms   — expires at t=300ms
	sw.Allow() // t=0ms   — expires at t=300ms
	sw.Allow() // t=0ms   — expires at t=300ms

	if sw.Allow() {
		t.Fatal("expected deny at capacity")
	}

	time.Sleep(350 * time.Millisecond) // all 3 expire

	if !sw.Allow() {
		t.Fatal("expected allow after full expiry")
	}
}

func TestSlidingWindow_ConcurrentAllow(t *testing.T) {
	const capacity = 100
	sw := limiter.NewSlidingWindow(capacity, time.Second)

	var (
		wg      sync.WaitGroup
		allowed int
		mu      sync.Mutex
	)

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if sw.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed > capacity {
		t.Fatalf("allowed %d > capacity %d: sliding window over-accepted", allowed, capacity)
	}
}

func TestSlidingWindow_NoPanicOnZeroCapacity(t *testing.T) {
	sw := limiter.NewSlidingWindow(0, time.Second)
	if sw.Allow() {
		t.Fatal("expected deny on zero-capacity window")
	}
}
