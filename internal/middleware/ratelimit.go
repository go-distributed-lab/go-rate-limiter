package middleware

import (
	"go-rate-limiter/internal/limiter"
	"go-rate-limiter/internal/logger"
	"go-rate-limiter/internal/metrics"
	"net/http"
)

func RateLimit(limiter limiter.Limiter, m *metrics.Metrics, log *logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			m.RecordDenied()
			log.Warn("rate limited",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
			)
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}
		m.RecordAllowed()
		next.ServeHTTP(w, r)
	})
}
