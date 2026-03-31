package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/markbfromdc/cloudcode/internal/config"
	"github.com/markbfromdc/cloudcode/internal/logging"
)

func TestNewHandler(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	if handler.hub != hub {
		t.Error("expected hub to be set")
	}
	if handler.cfg != cfg {
		t.Error("expected config to be set")
	}
}

func TestServeHTTPMissingSessionID(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	// Request with no session_id.
	req := httptest.NewRequest(http.MethodGet, "/ws/terminal", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestServeHTTPMissingUserContext(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	// Request with session_id but no user context.
	req := httptest.NewRequest(http.MethodGet, "/ws/terminal?session_id=abc", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without user context, got %d", rec.Code)
	}
}

func TestServeHTTPInvalidUserContext(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	// Request with session_id and wrong user context type.
	req := httptest.NewRequest(http.MethodGet, "/ws/terminal?session_id=abc", nil)
	ctx := context.WithValue(req.Context(), contextKeyUserID, 12345) // int, not string
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for invalid user context type, got %d", rec.Code)
	}
}

func TestServeHTTPNoWebSocketUpgrade(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	// Request with valid auth context but no WebSocket upgrade headers.
	req := httptest.NewRequest(http.MethodGet, "/ws/terminal?session_id=abc", nil)
	ctx := context.WithValue(req.Context(), contextKeyUserID, "user-1")
	req = req.WithContext(ctx)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Upgrader returns error because this isn't a real WebSocket request.
	// The handler logs the error and returns — no panic.
	if rec.Code == http.StatusUnauthorized {
		t.Error("should have passed auth check")
	}
}

func TestContextKeyUserID(t *testing.T) {
	if contextKeyUserID != "user_id" {
		t.Errorf("expected user_id, got %s", contextKeyUserID)
	}
}

func TestUpgraderOriginCheck(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	hub := NewHub(log)
	cfg := &config.Config{
		WSReadBufferSize:  4096,
		WSWriteBufferSize: 4096,
		AllowedOrigins:    []string{"http://localhost:3000", "https://app.cloudcode.dev"},
	}

	handler := NewHandler(hub, cfg, nil, log)

	// Test the upgrader's CheckOrigin function.
	allowedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	allowedReq.Header.Set("Origin", "http://localhost:3000")

	blockedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	blockedReq.Header.Set("Origin", "http://evil.com")

	if !handler.upgrader.CheckOrigin(allowedReq) {
		t.Error("expected localhost:3000 to be allowed")
	}

	if handler.upgrader.CheckOrigin(blockedReq) {
		t.Error("expected evil.com to be blocked")
	}

	// Test second allowed origin.
	allowed2 := httptest.NewRequest(http.MethodGet, "/", nil)
	allowed2.Header.Set("Origin", "https://app.cloudcode.dev")

	if !handler.upgrader.CheckOrigin(allowed2) {
		t.Error("expected app.cloudcode.dev to be allowed")
	}
}
