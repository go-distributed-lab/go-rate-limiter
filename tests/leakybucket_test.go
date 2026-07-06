package tests

import (
	"sync"
	"testing"
	"time"

	"go-rate-limiter/internal/limiter"
)

func TestLeakyBucket_AllowsUpToCapacity(t *testing.T) {
	lb := limiter.NewLeakyBucket(5, 1)

	for i := range 5 {
		if !lb.Allow() {
			t.Fatalf("expected allow on request %d", i+1)
		}
	}
	if lb.Allow() {
		t.Fatal("expected deny when queue full")
	}
}

func TestLeakyBucket_StartsEmpty(t *testing.T) {
	lb := limiter.NewLeakyBucket(10, 1)
	if got := lb.Queue(); got != 0 {
		t.Fatalf("expected empty queue, got %d", got)
	}
}

func TestLeakyBucket_DrainOverTime(t *testing.T) {
	// drainRate=10/sec, fill to capacity=5, wait 1s → should drain fully
	lb := limiter.NewLeakyBucket(5, 10)

	for range 5 {
		lb.Allow()
	}
	if lb.Queue() != 5 {
		t.Fatal("expected full queue after 5 requests")
	}

	time.Sleep(600 * time.Millisecond)

	if lb.Queue() != 0 {
		t.Fatalf("expected empty queue after drain, got %d", lb.Queue())
	}
}

func TestLeakyBucket_RejectWhenFull(t *testing.T) {
	lb := limiter.NewLeakyBucket(1, 0)
	lb.Allow() // fill the single slot
	if lb.Allow() {
		t.Fatal("expected deny on full queue")
	}
}

func TestLeakyBucket_ConcurrentAllow(t *testing.T) {
	const capacity = 100
	lb := limiter.NewLeakyBucket(capacity, 0)

	var (
		wg      sync.WaitGroup
		allowed int
		mu      sync.Mutex
	)

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if lb.Allow() {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowed > capacity {
		t.Fatalf("allowed %d > capacity %d: leaky bucket over-accepted", allowed, capacity)
	}
}

func TestLeakyBucket_NoPanicOnZeroCapacity(t *testing.T) {
	lb := limiter.NewLeakyBucket(0, 0)
	if lb.Allow() {
		t.Fatal("expected deny on zero-capacity bucket")
	}
}

func TestLeakyBucket_AcceptsAfterDrain(t *testing.T) {
	// fill completely, wait for drain, confirm accepts again
	lb := limiter.NewLeakyBucket(3, 10)

	for range 3 {
		lb.Allow()
	}
	if lb.Allow() {
		t.Fatal("expected deny when full")
	}

	time.Sleep(400 * time.Millisecond)

	if !lb.Allow() {
		t.Fatal("expected allow after drain")
	}
}
