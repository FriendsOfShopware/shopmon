// NOTE: This rate limiter is in-memory and resets on restart.
// For multi-instance deployments, replace with a Redis-backed rate limiter.
// The X-Real-IP header should only be trusted when behind a reverse proxy.
package auth

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

type rateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	window   time.Duration
	max      int
}

func newRateLimiter(ctx context.Context, window time.Duration, max int) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		window:   window,
		max:      max,
	}
	// Cleanup old entries periodically
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				rl.cleanup()
			}
		}
	}()
	return rl
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	cutoff := time.Now().Add(-rl.window)
	for key, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = valid
		}
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter to recent requests
	var recent []time.Time
	for _, t := range rl.requests[key] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.max {
		rl.requests[key] = recent
		return false
	}

	rl.requests[key] = append(recent, now)
	return true
}

// RateLimitMiddleware returns middleware that rate-limits by client IP.
func RateLimitMiddleware(rl *rateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			// Use X-Real-IP if available (set by RealIP middleware)
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				ip = realIP
			}

			if !rl.allow(ip) {
				w.Header().Set("Retry-After", "60")
				httputil.WriteError(w, http.StatusTooManyRequests, "too many requests, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
