package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// visitor tracks the token bucket state for a single IP address.
type visitor struct {
	tokens    float64
	lastSeen  time.Time
	mu        sync.Mutex
}

// RateLimiter implements a per-IP token bucket rate limiter.
type RateLimiter struct {
	visitors sync.Map
	rate     float64
	burst    int
	stop     chan struct{}
}

// NewRateLimiter creates a new RateLimiter that allows rps requests per second
// with the given burst capacity. It starts a background goroutine to clean up
// stale visitor entries every minute.
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rate:  rps,
		burst: burst,
		stop:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// allow checks whether the given IP is allowed to make a request.
// It returns the number of remaining tokens (floored to int) and whether the request is allowed.
func (rl *RateLimiter) allow(ip string) (int, bool) {
	val, _ := rl.visitors.LoadOrStore(ip, &visitor{
		tokens:   float64(rl.burst),
		lastSeen: time.Now(),
	})
	v := val.(*visitor)

	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now

	// Refill tokens based on elapsed time.
	v.tokens += elapsed * rl.rate
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}

	if v.tokens < 1 {
		remaining := int(v.tokens)
		if remaining < 0 {
			remaining = 0
		}
		return remaining, false
	}

	v.tokens--
	return int(v.tokens), true
}

// Middleware returns an HTTP middleware that enforces the rate limit.
// It returns 429 Too Many Requests when the limit is exceeded and sets
// the X-RateLimit-Remaining header on every response.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)
		remaining, allowed := rl.allow(ip)
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Stop signals the cleanup goroutine to exit.
func (rl *RateLimiter) Stop() {
	close(rl.stop)
}

// cleanupLoop removes visitor entries that have not been seen in over 3 minutes.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.visitors.Range(func(key, value interface{}) bool {
				v := value.(*visitor)
				v.mu.Lock()
				age := time.Since(v.lastSeen)
				v.mu.Unlock()
				if age > 3*time.Minute {
					rl.visitors.Delete(key)
				}
				return true
			})
		case <-rl.stop:
			return
		}
	}
}

// extractIP returns the client IP from the request, preferring X-Forwarded-For
// and falling back to RemoteAddr.
func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xff := r.Header.Get("X-Real-IP"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
