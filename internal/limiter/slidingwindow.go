package limiter

import (
	"sync"
	"time"
)

// SlidingWindow implements the sliding window log algorithm.
//
// Instead of a fixed counter that resets at clock boundaries, it tracks
// the timestamp of every accepted request in a ring buffer. On each Allow()
// call it evicts timestamps older than windowSize and checks whether the
// remaining count is below capacity.
//
// This eliminates the boundary burst problem of Fixed Window: at any point
// in time, no more than capacity requests can have been accepted in the
// last windowSize duration.
//
// Memory: O(capacity) — one time.Time (24 bytes) per slot.
// Thread-safe: all fields protected by mu.
type SlidingWindow struct {
	mu         sync.Mutex
	capacity   int
	windowSize time.Duration
	timestamps []time.Time
}

// NewSlidingWindow creates a sliding window rate limiter.
//
//	capacity   – max requests allowed in any windowSize duration
//	windowSize – the rolling time window (e.g. time.Second)
func NewSlidingWindow(capacity int, windowSize time.Duration) *SlidingWindow {
	return &SlidingWindow{
		capacity:   capacity,
		windowSize: windowSize,
		timestamps: make([]time.Time, 0, capacity),
	}
}

// evict removes all timestamps that have fallen outside the current window.
// Must be called with mu held.
func (sw *SlidingWindow) evict() {
	cutoff := time.Now().Add(-sw.windowSize)
	i := 0
	for i < len(sw.timestamps) && sw.timestamps[i].Before(cutoff) {
		i++
	}
	sw.timestamps = sw.timestamps[i:]
}

// Allow returns true if fewer than capacity requests have been accepted
// in the last windowSize duration, false otherwise (→ HTTP 429).
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	sw.evict()

	if len(sw.timestamps) >= sw.capacity {
		return false
	}
	sw.timestamps = append(sw.timestamps, time.Now())
	return true
}

// Count returns the number of requests in the current window.
// Useful for testing and metrics.
func (sw *SlidingWindow) Count() int {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.evict()
	return len(sw.timestamps)
}
