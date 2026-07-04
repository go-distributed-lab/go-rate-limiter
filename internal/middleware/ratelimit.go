package middleware

import (
	"go-rate-limiter/internal/limiter"
	"log"
	"net/http"
)

func RateLimit(limiter limiter.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			log.Printf("rate limited : %s %s", r.Method, r.URL.Path)
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
