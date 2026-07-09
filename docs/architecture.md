# go-rate-limiter Architecture

A production-grade rate limiting library written in Go implementing four classic algorithms behind a common interface.

---

# Table of Contents

- [System Overview](#system-overview)
- [Request Flow](#request-flow)
- [Limiter Interface](#limiter-interface)
- [Algorithms](#algorithms)
  - Token Bucket
  - Leaky Bucket
  - Fixed Window
  - Sliding Window
- [Concurrency Model](#concurrency-model)
- [Metrics & Logging](#metrics--logging)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Design Decisions](#design-decisions)
- [CI Pipeline](#ci-pipeline)
- [Benchmark Summary](#benchmark-summary)

---

# System Overview

```
                    HTTP Client
                         │
                         ▼
                  net/http Server
                         │
                         ▼
       middleware.RateLimit(...)
                         │
        ┌────────────────┴────────────────┐
        │                                 │
        ▼                                 ▼
 limiter.Allow() == true        limiter.Allow() == false
        │                                 │
        ▼                                 ▼
metrics.RecordAllowed()       metrics.RecordDenied()
        │                                 │
        ▼                                 ▼
    Next Handler                  HTTP 429 Response
```

Three HTTP endpoints are exposed.

| Endpoint | Description | Rate Limited |
|----------|-------------|-------------|
| `/` | Demo endpoint | ✅ |
| `/health` | Health check | ✅ |
| `/metrics` | Metrics endpoint | ✅ |

---

# Request Flow

```
Incoming Request
        │
        ▼
RateLimit Middleware
        │
        ▼
Limiter.Allow()
        │
 ┌──────┴──────┐
 │             │
 ▼             ▼
Allowed      Denied
 │             │
 ▼             ▼
200 OK      429 Too Many Requests
```

The middleware never knows which algorithm is being used.

It only depends on the common `Limiter` interface.

---

# Limiter Interface

Every algorithm implements the same interface.

```go
type Limiter interface {
    Allow() bool
}
```

Because of this, the middleware remains completely algorithm-independent.

Changing algorithms only requires changing the constructor in `main.go` (or using the configured environment variable).

---

# Algorithms

---

## 1. Token Bucket

### Principle

A bucket contains tokens.

Tokens refill continuously until reaching the bucket capacity.

Every request consumes one token.

```
       refill
         │
         ▼
 +----------------+
 |                |
 |    Tokens      |
 |                |
 +----------------+
         │
 consume │
         ▼
      Allow()
```

### Complexity

| Operation | Complexity |
|------------|-----------|
| Allow() | O(1) |
| Memory | O(1) |

### Advantages

- Supports bursts
- Excellent throughput
- Very common API limiter

### Limitation

A full bucket at startup allows an initial burst.

---

## 2. Leaky Bucket

### Principle

Instead of accumulating tokens, requests fill a queue.

The queue drains at a constant speed.

```
Requests
    │
    ▼

+-------------+
|             |
|   Queue     |
|             |
+-------------+
      │
      ▼
 Constant Drain
      │
      ▼
 Allow()
```

### Complexity

| Operation | Complexity |
|------------|-----------|
| Allow() | O(1) |
| Memory | O(1) |

### Advantages

- Constant output rate
- Smooth traffic
- Excellent for downstream services

### Limitation

Traditional implementations introduce latency.

This project instead fails fast with HTTP 429 instead of queuing goroutines.

---

## 3. Fixed Window

### Principle

Time is divided into equal windows.

Each window has its own request counter.

```
Window 1
+-----------+
| Counter=5 |
+-----------+

Reset

Window 2
+-----------+
| Counter=0 |
+-----------+
```

### Complexity

| Operation | Complexity |
|------------|-----------|
| Allow() | O(1) |
| Memory | O(1) |

### Advantages

- Extremely simple
- Fastest implementation
- Small memory footprint

### Limitation

Boundary Burst Problem.

A client may send:

```
Capacity requests
at end of Window A

+

Capacity requests
at beginning of Window B
```

Result:

```
2 × Capacity requests
within milliseconds
```

---

## 4. Sliding Window

### Principle

Instead of counters, timestamps are stored.

Older timestamps are evicted every request.

```
Old Requests

[ x ][ x ][ x ][ x ][ x ]

          Window

Now
```

Only timestamps inside the current window remain.

### Complexity

| Operation | Complexity |
|------------|-----------|
| Allow() | O(k) (expired entries) |
| Memory | O(capacity) |

### Advantages

- Accurate
- Fair
- Eliminates boundary burst problem

### Limitation

Requires storing timestamps.

Uses more memory than other algorithms.

---

# Algorithm Comparison

| Algorithm | Burst | Memory | Time | Typical Use |
|------------|-------|--------|------|-------------|
| Token Bucket | ✅ | O(1) | O(1) | Public APIs |
| Leaky Bucket | ❌ | O(1) | O(1) | Traffic shaping |
| Fixed Window | Edge Burst | O(1) | O(1) | Simple quotas |
| Sliding Window | ❌ | O(capacity) | O(k) | Strict fairness |

---

# Concurrency Model

Every limiter protects internal state using a single mutex.

```
          Request 1
              │
              ▼

          sync.Mutex

              ▲
              │

          Request 2
```

No background goroutines are spawned.

State updates are computed lazily using `time.Now()`.

Advantages:

- No ticker cleanup
- No goroutine leaks
- Simpler lifecycle
- Deterministic behavior

Metrics use atomic counters instead of mutexes.

```
Allowed Counter

LOCK XADD

Denied Counter

LOCK XADD
```

This avoids lock contention on every request.

---

# Metrics & Logging

## Metrics

Atomic counters track:

- Allowed requests
- Denied requests

Metrics endpoint returns:

```json
{
    "algorithm":"tokenbucket",
    "allowed":100,
    "denied":5
}
```

---

## Logger

Structured logging is used throughout the application.

Example:

```
level=INFO
algorithm=tokenbucket
allowed=123
denied=12
```

---

# Configuration

Configuration is loaded from environment variables.

Typical parameters include:

| Variable | Description |
|-----------|-------------|
| ALGORITHM | Rate limiting algorithm |
| CAPACITY | Maximum bucket/window size |
| REFILL_RATE | Token refill rate |
| DRAIN_RATE | Queue drain rate |
| WINDOW | Window duration |
| PORT | HTTP server port |

---

# Project Structure

```
go-rate-limiter/
│
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── benchmarks/
│   ├── fixedwindow_bench_test.go
│   ├── leakybucket_bench_test.go
│   ├── slidingwindow_bench_test.go
│   └── tokenbucket_bench_test.go
│
├── cmd/
│   └── ratelimiter/
│       └── main.go
│
├── docs/
│   ├── architecture.md
│   └── diagrams/
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── limiter/
│   │   ├── limiter.go
│   │   ├── tokenbucket.go
│   │   ├── leakybucket.go
│   │   ├── fixedwindow.go
│   │   └── slidingwindow.go
│   │
│   ├── logger/
│   │   ├── logger.go
│   │   └── logger_test.go
│   │
│   ├── metrics/
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   │
│   └── middleware/
│       └── ratelimit.go
│
├── pkg/
│
├── scripts/
│
├── tests/
│   ├── fixedwindow_test.go
│   ├── leakybucket_test.go
│   ├── slidingwindow_test.go
│   └── tokenbucket_test.go
│
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── README.md
└── ratelimiter.exe
```

---

# Design Decisions

## Common Interface

All algorithms expose:

```go
Allow() bool
```

This enables complete decoupling between middleware and implementation.

---

## Lazy Evaluation

Instead of background refill goroutines:

- Token refill
- Queue drain
- Window reset
- Timestamp eviction

are all computed during `Allow()`.

Benefits:

- Zero goroutines
- No shutdown complexity
- Lower memory usage

---

## Atomic Metrics

Metrics are updated on every request.

Using `sync/atomic` avoids mutex contention.

---

## Zero Dependencies

The project only relies on Go's standard library.

Benefits:

- Small Docker image
- Faster builds
- No dependency conflicts
- Easier maintenance

---

# CI Pipeline

The GitHub Actions workflow performs:

| Stage | Command |
|--------|---------|
| Format | `gofmt` |
| Vet | `go vet ./...` |
| Tests | `go test ./...` |
| Race Detection | `go test -race ./...` |
| Benchmarks | `go test -bench=. -benchmem ./benchmarks/...` |
| Docker Build | `docker build .` |

---

# Benchmark Summary

Hardware

- AMD Ryzen 5 5600H
- 12 Logical CPUs
- Go 1.24
- Windows amd64

| Algorithm | Sequential | Parallel | Contended | Allocations |
|------------|-----------|----------|------------|-------------|
| Fixed Window | 10.96 ns/op | 37.38 ns/op | 38.35 ns/op | 0 |
| Token Bucket | 13.14 ns/op | 40.95 ns/op | 41.58 ns/op | 0 |
| Leaky Bucket | 13.22 ns/op | 44.69 ns/op | 45.39 ns/op | 0 |
| Sliding Window | 38.09 ns/op | 80.47 ns/op | 63.75 ns/op | 0 |

All algorithms perform **zero heap allocations** during `Allow()`.

---

# Summary

This project demonstrates four classic rate limiting algorithms implemented behind a common interface with:

- Zero external dependencies
- Thread-safe implementations
- Atomic metrics
- Structured logging
- Graceful shutdown
- Docker support
- GitHub Actions CI
- Benchmarks
- Unit tests

The architecture is intentionally modular, making it easy to add new algorithms without modifying the middleware or HTTP server.