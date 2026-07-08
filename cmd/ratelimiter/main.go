package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-rate-limiter/internal/config"
	"go-rate-limiter/internal/limiter"
	"go-rate-limiter/internal/logger"
	"go-rate-limiter/internal/metrics"
	"go-rate-limiter/internal/middleware"
)

func main() {
	log := logger.Default()
	cfg := config.Load()

	log.Info("starting",
		"algorithm", cfg.Algorithm,
		"capacity", cfg.Capacity,
		"refillRate", cfg.RefillRate,
		"port", cfg.Port,
	)

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
		log.Error("unknown algorithm", "algorithm", cfg.Algorithm)
		os.Exit(1)
	}

	m := metrics.NewMetrics(cfg.Algorithm)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.Handle("/metrics", m.Handler())

	handler := middleware.RateLimit(l, m, log, mux)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	// Block until SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		log.Error("server error", "err", err)
		os.Exit(1)
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig)
	}

	// Graceful shutdown — drain in-flight requests, 10s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", "err", err)
		os.Exit(1)
	}

	log.Info("shutdown complete",
		"allowed", m.Allowed.Load(),
		"denied", m.Denied.Load(),
	)
}
