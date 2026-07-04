# go-rate-limiter — Architecture

## System overview

```
HTTP Client
    │
    ▼
 :8080 (net/http)
    │
    ▼
middleware.RateLimit(limiter.Limiter, next)
    │
    ├── l.Allow() == true  → 200 OK → next handler
    └── l.Allow() == false → 429 Too Many Requests
```

## Limiter interface

All four algorithms implement one interface:

```go
type Limiter interface {
    Allow() bool
}
```

The HTTP middleware is completely algorithm-agnostic. Swapping from Token Bucket
to Sliding Window requires zero middleware changes — only the constructor call
in main.go changes.

## Algorithm map

| Algorithm      | State              | Burst | Memory | Clock dependency |
|----------------|--------------------|-------|--------|-----------------|
| Token Bucket   | tokens + time      | Yes   | O(1)   | Yes             |
| Leaky Bucket   | queue + time       | No    | O(n)   | Yes             |
| Fixed Window   | counter + reset    | Edge  | O(1)   | Yes             |
| Sliding Window | ring of timestamps | No    | O(n)   | Yes             |

## Folder structure

```
go-rate-limiter/
├── cmd/
│   └── ratelimiter/
│       └── main.go          # HTTP server entry point
├── internal/
│   ├── limiter/
│   │   ├── limiter.go       # Limiter interface
│   │   └── tokenbucket.go   # Day 1 implementation
│   ├── middleware/
│   │   └── ratelimit.go     # HTTP middleware (algorithm-agnostic)
│   └── config/
│       └── config.go        # Env-driven config
├── tests/
│   └── tokenbucket_test.go  # Integration + race tests
├── benchmarks/
│   └── tokenbucket_bench_test.go
├── docs/
│   └── architecture.md
├── .github/
│   └── workflows/
│       └── ci.yml
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## Concurrency model

- `TokenBucket` uses a single `sync.Mutex` — one lock per `Allow()` call.
- No goroutines are spawned by the library itself; all concurrency comes from
  the HTTP server's goroutine-per-connection model.
- Refill is lazy (computed on each `Allow()` call from elapsed time) — no
  background ticker goroutine needed. This avoids the "goroutine leak on
  shutdown" class of bugs.

## Key design decisions

**Why lazy refill instead of a ticker goroutine?**
A background ticker requires a shutdown signal and risks leaking goroutines if
Stop() is never called. Lazy refill from time.Now() gives identical behaviour
with zero goroutine overhead and no lifecycle to manage.

**Why `Allow() bool` instead of `Allow() error`?**
Rate limiting is a hot path. A boolean is cheaper to check than an error, and
the semantic is binary — either a request is allowed or it isn't. Errors are
for unexpected failures, not for policy decisions.

**Why channel-close-driven shutdown (inherited from go-worker-pool)?**
Context cancellation can silently drop in-flight requests. Closing a channel
propagates the shutdown signal without losing work. The HTTP server uses
http.Server.Shutdown() for graceful drain — same principle.

## Algorithm deep-dives

### Token Bucket
Tokens accumulate at a fixed rate up to a maximum capacity. Each request
consumes one token. If no token is available, the request is rejected.
Allows controlled bursting up to the capacity limit.

- Best for: APIs that want to allow short bursts but enforce a long-run rate.
- Weakness: A full bucket at startup allows an immediate burst of `capacity`
  requests before the rate limit kicks in.

### Leaky Bucket (Day 2)
Requests enter a fixed-size queue and are processed at a constant rate.
If the queue is full, new requests are dropped. Produces perfectly smooth
output regardless of input burst shape.

- Best for: Outbound request shaping (e.g. calling a third-party API).
- Weakness: Adds latency — requests queue instead of failing fast.

### Fixed Window (Day 3)
A counter resets every fixed interval (e.g. every minute). Simple and
predictable, but vulnerable to boundary bursts: a client can send
2× the limit by hammering the window boundary.

- Best for: Simple quota enforcement (e.g. 1000 req/hour).
- Weakness: The boundary burst problem.

### Sliding Window (Day 4)
Tracks individual request timestamps in a ring buffer. The window slides
continuously rather than resetting at fixed points. Eliminates boundary
bursts entirely.

- Best for: Strict per-user rate limiting where fairness matters.
- Weakness: O(n) memory per client (n = max requests in window).

## CI pipeline

| Job        | Command                                      |
|------------|----------------------------------------------|
| lint       | go vet ./... + staticcheck ./...             |
| test       | go test -race ./internal/... ./tests/...     |
| bench      | go test -bench=. -benchmem -benchtime=3s     |
| docker     | docker build (scratch image, ~4MB)           |
| security   | govulncheck ./...                            |
