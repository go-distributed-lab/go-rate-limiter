package benchmarks

import (
	"go-rate-limiter/internal/limiter"
	"io"
	"log"
	"os"
	"testing"
)

// TestMain silences log output during benchmarks.
func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func BenchmarkTokenBucket_Allow_Sequential(b *testing.B) {
	tb := limiter.NewTokenBucket(b.N+1, b.N+1) // always has tokens
	b.ResetTimer()
	for range b.N {
		tb.Allow()
	}
}

func BenchmarkTokenBucket_Allow_Parallel(b *testing.B) {
	tb := limiter.NewTokenBucket(b.N+1, b.N+1)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}

func BenchmarkTokenBucket_Allow_Contended(b *testing.B) {
	// Fixed small capacity — realistic contention scenario (many goroutines, few tokens)
	tb := limiter.NewTokenBucket(10, 1000000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}
