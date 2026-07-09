.PHONY: run run-token run-leaky run-fixed run-sliding \
        test bench lint build \
        docker-token docker-leaky docker-fixed docker-sliding \
        docker-down clean help

# ── Local run ────────────────────────────────────────────────────────────────

run:
	ALGORITHM=token_bucket CAPACITY=10 REFILL_RATE=5 go run ./cmd/ratelimiter

run-token:
	ALGORITHM=token_bucket CAPACITY=10 REFILL_RATE=5 go run ./cmd/ratelimiter

run-leaky:
	ALGORITHM=leaky_bucket CAPACITY=10 REFILL_RATE=5 go run ./cmd/ratelimiter

run-fixed:
	ALGORITHM=fixed_window CAPACITY=10 go run ./cmd/ratelimiter

run-sliding:
	ALGORITHM=sliding_window CAPACITY=10 go run ./cmd/ratelimiter

# ── Test & lint ──────────────────────────────────────────────────────────────

test:
	go test -race ./internal/... ./tests/...

bench:
	cd benchmarks && go test -bench=. -benchmem -benchtime=3s .

lint:
	go vet ./...
	staticcheck ./...

# ── Docker ───────────────────────────────────────────────────────────────────

build:
	docker build -t go-rate-limiter:latest .

docker-token:
	docker compose --profile token-bucket up --build

docker-leaky:
	docker compose --profile leaky-bucket up --build

docker-fixed:
	docker compose --profile fixed-window up --build

docker-sliding:
	docker compose --profile sliding-window up --build

docker-down:
	docker compose down

# ── Clean ────────────────────────────────────────────────────────────────────

clean:
	docker compose down --rmi local
	go clean ./...

# ── Help ─────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "  go-rate-limiter"
	@echo ""
	@echo "  Local:"
	@echo "    make run-token     run with Token Bucket"
	@echo "    make run-leaky     run with Leaky Bucket"
	@echo "    make run-fixed     run with Fixed Window"
	@echo "    make run-sliding   run with Sliding Window"
	@echo ""
	@echo "  Test:"
	@echo "    make test          go test -race"
	@echo "    make bench         benchmarks (3s per case)"
	@echo "    make lint          go vet + staticcheck"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-token  docker compose token bucket"
	@echo "    make docker-leaky  docker compose leaky bucket"
	@echo "    make docker-fixed  docker compose fixed window"
	@echo "    make docker-sliding docker compose sliding window"
	@echo "    make docker-down   stop all containers"
	@echo ""
	@echo "    make clean         remove containers + images"
	@echo ""