package main

import (
	"fmt"
	"go-rate-limiter/internal/config"
	"go-rate-limiter/internal/limiter"
	"go-rate-limiter/internal/middleware"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := config.Load()

	log.Printf("algorithm=%s capacity=%d refillRate=%d port=%s",
		cfg.Algorithm, cfg.Capacity, cfg.RefillRate, cfg.Port)

	var l limiter.Limiter
	switch cfg.Algorithm {
	case "token_bucket":
		l = limiter.NewTokenBucket(cfg.Capacity, cfg.RefillRate)
	case "leaky_bucket":
		l = limiter.NewLeakyBucket(cfg.Capacity, cfg.RefillRate)
	case "fixed_window":
		l = limiter.NewFixedWindow(cfg.Capacity, time.Second)
	case "sliding_window":
		l = limiter.NewSlidingWindow(cfg.Capacity, time.Second)
	default:
		log.Fatalf("unknown algorithm: %s", cfg.Algorithm)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "healthy")
	})

	handler := middleware.RateLimit(l, mux)

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
