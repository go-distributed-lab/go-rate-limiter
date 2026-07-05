package config

import (
	"os"
	"strconv"
)

// Config holds all tunable parameters for the rate limiter server.
type Config struct {
	Algorithm  string // ALGORITHM: token_bucket | leaky_bucket | sliding_window | fixed_window
	Capacity   int    // CAPACITY:  max tokens / window size
	RefillRate int    // REFILL_RATE: tokens per second (token bucket) or drain rate (leaky)
	Port       string // PORT: HTTP server port
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getEnvStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads configuration from environment variables with sensible defaults.
func Load() Config {
	return Config{
		Algorithm:  getEnvStr("ALGORITHM", "token_bucket"),
		Capacity:   getEnvInt("CAPACITY", 10),
		RefillRate: getEnvInt("REFILL_RATE", 5),
		Port:       getEnvStr("PORT", "8080"),
	}
}
