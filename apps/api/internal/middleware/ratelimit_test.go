package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiterAllowsUpToLimitThenBlocks(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	do := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := do(); code != http.StatusOK {
		t.Fatalf("request 1: got %d, want 200", code)
	}
	if code := do(); code != http.StatusOK {
		t.Fatalf("request 2: got %d, want 200", code)
	}
	if code := do(); code != http.StatusTooManyRequests {
		t.Fatalf("request 3: got %d, want 429", code)
	}
}

func TestRateLimiterSeparatesByIP(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	handler := rl.Limit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	do := func(addr string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = addr
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := do("10.0.0.1:1"); code != http.StatusOK {
		t.Fatalf("ip1 first request got %d", code)
	}
	if code := do("10.0.0.2:1"); code != http.StatusOK {
		t.Fatalf("ip2 first request should pass, got %d", code)
	}
}
