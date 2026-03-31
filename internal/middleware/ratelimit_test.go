package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiterAllowsBelowLimit(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First 5 requests should be allowed (burst = 5).
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRateLimiterRejectsOverLimit(t *testing.T) {
	rl := NewRateLimiter(1, 3)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Exhaust the burst.
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:5678"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Next request should be rate limited.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:5678"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rec.Code)
	}
}

func TestRateLimiterSetsHeader(t *testing.T) {
	rl := NewRateLimiter(10, 5)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Error("expected X-RateLimit-Remaining header to be set")
	}
}

func TestRateLimiterDifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1)
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First IP uses its burst.
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "10.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("IP1 first request: expected 200, got %d", rec1.Code)
	}

	// Second IP should still be allowed (separate bucket).
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "10.0.0.2:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("IP2 first request: expected 200, got %d", rec2.Code)
	}
}

func TestExtractIPFromXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	ip := extractIP(req)
	if ip != "203.0.113.50" {
		t.Errorf("expected 203.0.113.50, got %s", ip)
	}
}

func TestExtractIPFromXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "198.51.100.1")
	ip := extractIP(req)
	if ip != "198.51.100.1" {
		t.Errorf("expected 198.51.100.1, got %s", ip)
	}
}

func TestExtractIPFromRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"
	ip := extractIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %s", ip)
	}
}
