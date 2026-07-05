.PHONY: run test bench lint build docker docker-down clean

run:
	go run ./cmd/ratelimiter

test:
	go test -race ./internal/... ./tests/...

bench:
	cd benchmarks && go test -bench=. -benchmem -benchtime=3s .

lint:
	go vet ./...
	staticcheck ./...

build:
	docker build -t go-rate-limiter:latest .

docker:
	docker compose up --build

docker-down:
	docker compose down

clean:
	docker compose down --rmi local
	go clean ./...