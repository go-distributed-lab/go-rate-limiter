package limiter

import (
	"sync"
	"time"
)

type TokenBucket struct {
	mu         sync.Mutex
	capacity   int
	tokens     int
	refillRate int
	lastRefill time.Time
}

func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	add := int(elapsed * float64(tb.refillRate))
	if add > 0 {
		tb.tokens += add
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	if tb.tokens <= 0 {
		return false
	}
	tb.tokens--
	return true
}

func (tb *TokenBucket) Tokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.refill()
	return tb.tokens
}
