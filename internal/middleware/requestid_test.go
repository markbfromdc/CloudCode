package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDSetsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Request-ID")
	if id == "" {
		t.Error("expected X-Request-ID header to be set")
	}
	// UUID v4 is 36 characters (8-4-4-4-12).
	if len(id) != 36 {
		t.Errorf("expected UUID format (36 chars), got %q (%d chars)", id, len(id))
	}
}

func TestRequestIDAvailableInContext(t *testing.T) {
	var ctxID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = GetRequestID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	headerID := rec.Header().Get("X-Request-ID")
	if ctxID == "" {
		t.Error("expected request ID in context")
	}
	if ctxID != headerID {
		t.Errorf("context ID %q does not match header ID %q", ctxID, headerID)
	}
}

func TestRequestIDUniquePerRequest(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req1)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)

	id1 := rec1.Header().Get("X-Request-ID")
	id2 := rec2.Header().Get("X-Request-ID")

	if id1 == id2 {
		t.Errorf("expected unique IDs, both were %q", id1)
	}
}

func TestGetRequestIDNoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := GetRequestID(req)
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}
