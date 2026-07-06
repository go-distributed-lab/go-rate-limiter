package limiter

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	mu        sync.Mutex
	capacity  int
	queue     int
	drainRate int
	lastDrain time.Time
}

func NewLeakyBucket(capacity, drainRate int) *LeakyBucket {
	return &LeakyBucket{
		capacity:  capacity,
		drainRate: drainRate,
		queue:     0,
		lastDrain: time.Now(),
	}
}

func (lb *LeakyBucket) drain() {
	now := time.Now()
	elapsed := now.Sub(lb.lastDrain).Seconds()
	drained := int(elapsed * float64(lb.drainRate))
	if drained > 0 {
		lb.queue -= drained
		if lb.queue < 0 {
			lb.queue = 0
		}
		lb.lastDrain = now
	}
}

func (lb *LeakyBucket) Allow() bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.drain()

	if lb.queue >= lb.capacity {
		return false
	}
	lb.queue++
	return true
}

func (lb *LeakyBucket) Queue() int {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.drain()
	return lb.queue
}
