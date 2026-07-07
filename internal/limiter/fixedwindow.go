package limiter

import (
	"sync"
	"time"
)

type FixedWindow struct {
	mu          sync.Mutex
	capacity    int
	count       int
	windowSize  time.Duration
	windowStart time.Time
}

func NewFixedWindow(capacity int, windowSize time.Duration) *FixedWindow {
	return &FixedWindow{
		capacity:    capacity,
		count:       0,
		windowSize:  windowSize,
		windowStart: time.Now(),
	}
}

func (fw *FixedWindow) resetIfExpired() {
	if time.Since(fw.windowStart) >= fw.windowSize {
		fw.count = 0
		fw.windowStart = time.Now()
	}
}

func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.resetIfExpired()

	if fw.count >= fw.capacity {
		return false
	}
	fw.count++
	return true
}

func (fw *FixedWindow) Count() int {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.resetIfExpired()
	return fw.count
}
