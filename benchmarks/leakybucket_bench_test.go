package benchmarks

import (
	"testing"

	"go-rate-limiter/internal/limiter"
)

func BenchmarkLeakyBucket_Allow_Sequential(b *testing.B) {
	lb := limiter.NewLeakyBucket(b.N+1, 0) // drainRate=0: no drain during bench
	b.ResetTimer()
	for range b.N {
		lb.Allow()
	}
}

func BenchmarkLeakyBucket_Allow_Parallel(b *testing.B) {
	lb := limiter.NewLeakyBucket(b.N+1, 0)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.Allow()
		}
	})
}

func BenchmarkLeakyBucket_Allow_Contended(b *testing.B) {
	// small queue, high concurrency — realistic contention
	lb := limiter.NewLeakyBucket(10, 1000000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			lb.Allow()
		}
	})
}
