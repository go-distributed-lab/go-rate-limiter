# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /rate-limiter ./cmd/ratelimiter

FROM scratch
COPY --from=builder /rate-limiter /rate-limiter
EXPOSE 8080
ENTRYPOINT ["/rate-limiter"]
