package benchmarks

import (
	"testing"
	"time"

	"go-rate-limiter/internal/limiter"
)

func BenchmarkSlidingWindow_Allow_Sequential(b *testing.B) {
	sw := limiter.NewSlidingWindow(b.N+1, time.Hour) // window never expires
	b.ResetTimer()
	for range b.N {
		sw.Allow()
	}
}

func BenchmarkSlidingWindow_Allow_Parallel(b *testing.B) {
	sw := limiter.NewSlidingWindow(b.N+1, time.Hour)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sw.Allow()
		}
	})
}

func BenchmarkSlidingWindow_Allow_Contended(b *testing.B) {
	// small capacity, realistic contention
	sw := limiter.NewSlidingWindow(10, time.Hour)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sw.Allow()
		}
	})
}
