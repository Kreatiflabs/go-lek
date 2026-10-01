package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimitConfig defines rate limiting configuration.
type RateLimitConfig struct {
	// RequestsPerSecond is the max requests per second per key
	RequestsPerSecond int
	// BurstSize is the max burst size (default: RequestsPerSecond)
	BurstSize int
	// KeyFunc extracts the rate limit key from the request (default: client IP)
	KeyFunc func(r *http.Request) string
	// CleanupInterval is how often to clean expired entries (default: 1 minute)
	CleanupInterval time.Duration
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimit returns a middleware that limits requests per second per client IP.
// Uses a token bucket algorithm with in-memory storage.
func RateLimit(requestsPerSecond int) func(http.Handler) http.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		RequestsPerSecond: requestsPerSecond,
	})
}

// RateLimitWithConfig returns a RateLimit middleware with custom config.
func RateLimitWithConfig(config RateLimitConfig) func(http.Handler) http.Handler {
	if config.BurstSize <= 0 {
		config.BurstSize = config.RequestsPerSecond
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = time.Minute
	}
	if config.KeyFunc == nil {
		config.KeyFunc = func(r *http.Request) string {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return r.RemoteAddr
			}
			return host
		}
	}

	var buckets sync.Map
	var once sync.Once

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			once.Do(func() {
				go func() {
					ticker := time.NewTicker(config.CleanupInterval)
					for range ticker.C {
						now := time.Now()
						buckets.Range(func(key, value any) bool {
							b := value.(*bucket)
							b.mu.Lock()
							if now.Sub(b.lastRefill) > config.CleanupInterval*2 {
								buckets.Delete(key)
							}
							b.mu.Unlock()
							return true
						})
					}
				}()
			})

			key := config.KeyFunc(r)
			val, _ := buckets.LoadOrStore(key, &bucket{
				tokens:     float64(config.BurstSize),
				lastRefill: time.Now(),
			})
			b := val.(*bucket)

			b.mu.Lock()
			now := time.Now()
			elapsed := now.Sub(b.lastRefill).Seconds()
			b.tokens += elapsed * float64(config.RequestsPerSecond)
			if b.tokens > float64(config.BurstSize) {
				b.tokens = float64(config.BurstSize)
			}
			b.lastRefill = now

			allowed := false
			if b.tokens >= 1.0 {
				b.tokens -= 1.0
				allowed = true
			}
			b.mu.Unlock()

			if !allowed {
				w.Header().Set("Retry-After", strconv.Itoa(1))
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
