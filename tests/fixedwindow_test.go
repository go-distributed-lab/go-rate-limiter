package tests

import (
	"sync"
	"testing"
	"time"

	"go-rate-limiter/internal/limiter"
)

func TestFixedWindow_AllowsUpToCapacity(t *testing.T) {
	fw := limiter.NewFixedWindow(5, time.Second)

	for i := range 5 {
		if !fw.Allow() {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
	if fw.Allow() {
		t.Fatal("expected deny when window full")
	}
}

func TestFixedWindow_StartsEmpty(t *testing.T) {
	fw := limiter.NewFixedWindow(10, time.Second)
	if got := fw.Count(); got != 0 {
		t.Fatalf("expected count 0, got %d", got)
	}
}

func TestFixedWindow_ResetsAfterWindow(t *testing.T) {
	fw := limiter.NewFixedWindow(3, 200*time.Millisecond)

	for range 3 {
		fw.Allow()
	}
	if fw.Allow() {
		t.Fatal("expected deny when window full")
	}

	time.Sleep(250 * time.Millisecond)

	for i := range 3 {
		if !fw.Allow() {
			t.Fatalf("expected allow after reset on request %d", i+1)
		}
	}
}

func TestFixedWindow_RejectWhenFull(t *testing.T) {
	fw := limiter.NewFixedWindow(1, time.Second)
	fw.Allow()
	if fw.Allow() {
		t.Fatal("expected deny on full window")
	}
}

func TestFixedWindow_ConcurrentAllow(t *testing.T) {
	const capacity = 100
	fw := limiter.NewFixedWindow(capacity, time.Second)

	var (
		wg      sync.WaitGroup
		allowed int
		mu      sync.Mutex
	)

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if fw.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed > capacity {
		t.Fatalf("allowed %d > capacity %d: fixed window over-accepted", allowed, capacity)
	}
}

func TestFixedWindow_NoPanicOnZeroCapacity(t *testing.T) {
	fw := limiter.NewFixedWindow(0, time.Second)
	if fw.Allow() {
		t.Fatal("expected deny on zero-capacity window")
	}
}

func TestFixedWindow_BoundaryBurst(t *testing.T) {
	// This test DOCUMENTS the known weakness of fixed window:
	// a client can get 2x capacity across a window boundary.
	fw := limiter.NewFixedWindow(5, 200*time.Millisecond)

	// Fill the first window completely
	firstWindow := 0
	for range 5 {
		if fw.Allow() {
			firstWindow++
		}
	}

	// Wait for window to reset
	time.Sleep(250 * time.Millisecond)

	// Immediately fill the second window
	secondWindow := 0
	for range 5 {
		if fw.Allow() {
			secondWindow++
		}
	}

	total := firstWindow + secondWindow
	// Both windows allow 5 each = 10 total in ~250ms
	// This is the boundary burst — document it, not fix it (Sliding Window does that)
	if total != 10 {
		t.Logf("boundary burst: %d requests allowed across window boundary", total)
	}
	t.Logf("KNOWN LIMITATION: boundary burst allows %d requests (2x capacity=%d)", total, 5)
}
