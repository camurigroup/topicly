package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/forum-framework/api/internal/shared/httpx"
	"github.com/forum-framework/api/internal/shared/response"
)

// RateLimiter is a simple in-memory fixed-window limiter keyed by client IP.
// It is the fallback used when no distributed store (Redis) is configured.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*window
	limit    int
	window   time.Duration
}

type window struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter builds a limiter allowing `limit` requests per `per` window.
func NewRateLimiter(limit int, per time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*window),
		limit:    limit,
		window:   per,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for k, v := range rl.visitors {
			if now.After(v.resetAt) {
				delete(rl.visitors, k)
			}
		}
		rl.mu.Unlock()
	}
}

// allow reports whether a key may proceed and, if not, the seconds to wait.
func (rl *RateLimiter) allow(key string) (bool, int) {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, ok := rl.visitors[key]
	if !ok || now.After(v.resetAt) {
		rl.visitors[key] = &window{count: 1, resetAt: now.Add(rl.window)}
		return true, 0
	}
	if v.count >= rl.limit {
		return false, int(time.Until(v.resetAt).Seconds()) + 1
	}
	v.count++
	return true, 0
}

// Limit is the middleware that enforces the rate limit.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, retry := rl.allow(httpx.ClientIP(r))
		if !ok {
			w.Header().Set("Retry-After", strconv.Itoa(retry))
			response.Error(w, http.StatusTooManyRequests, "rate_limited", "muitas requisições, tente novamente em instantes")
			return
		}
		next.ServeHTTP(w, r)
	})
}
