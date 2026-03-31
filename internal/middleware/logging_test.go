package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.INFO)

	handler := RequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected method in log, got: %s", output)
	}
	if !strings.Contains(output, "/test/path") {
		t.Errorf("expected path in log, got: %s", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("expected status code in log, got: %s", output)
	}
}

func TestRequestLoggerCapturesStatusCode(t *testing.T) {
	var buf bytes.Buffer
	log := logging.New(&buf, logging.INFO)

	handler := RequestLogger(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodPost, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "404") {
		t.Errorf("expected 404 in log, got: %s", output)
	}
}
