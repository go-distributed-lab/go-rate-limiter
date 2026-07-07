package benchmarks

import (
	"testing"
	"time"

	"go-rate-limiter/internal/limiter"
)

func BenchmarkFixedWindow_Allow_Sequential(b *testing.B) {
	fw := limiter.NewFixedWindow(b.N+1, time.Hour) // window never expires
	b.ResetTimer()
	for range b.N {
		fw.Allow()
	}
}

func BenchmarkFixedWindow_Allow_Parallel(b *testing.B) {
	fw := limiter.NewFixedWindow(b.N+1, time.Hour)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fw.Allow()
		}
	})
}

func BenchmarkFixedWindow_Allow_Contended(b *testing.B) {
	// small capacity, window never expires — pure mutex contention
	fw := limiter.NewFixedWindow(10, time.Hour)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fw.Allow()
		}
	})
}

func BenchmarkFixedWindow_Allow_WithReset(b *testing.B) {
	// tiny window — resets frequently, tests reset path performance
	fw := limiter.NewFixedWindow(b.N+1, time.Nanosecond)
	b.ResetTimer()
	for range b.N {
		fw.Allow()
	}
}
