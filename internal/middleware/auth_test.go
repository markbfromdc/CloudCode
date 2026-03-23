package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

const testSecret = "test-jwt-secret-key"

// generateTestJWT creates a valid JWT token for testing purposes.
func generateTestJWT(sub, email string, exp time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := JWTClaims{
		Sub:   sub,
		Email: email,
		Exp:   exp.Unix(),
		Iat:   time.Now().Unix(),
	}
	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return header + "." + payload + "." + signature
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	token := generateTestJWT("user-123", "test@example.com", time.Now().Add(time.Hour))

	var capturedUserID, capturedEmail string
	handler := Auth(testSecret, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, _ = r.Context().Value(ContextKeyUserID).(string)
		capturedEmail, _ = r.Context().Value(ContextKeyEmail).(string)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("expected user ID 'user-123', got %q", capturedUserID)
	}
	if capturedEmail != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", capturedEmail)
	}
}

func TestAuthMiddlewareMissingHeader(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := Auth(testSecret, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	token := generateTestJWT("user-123", "test@example.com", time.Now().Add(-time.Hour))

	handler := Auth(testSecret, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for expired token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddlewareInvalidSignature(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	token := generateTestJWT("user-123", "test@example.com", time.Now().Add(time.Hour))
	// Corrupt the signature.
	token = token[:len(token)-4] + "XXXX"

	handler := Auth(testSecret, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for invalid signature")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddlewareInvalidFormat(t *testing.T) {
	log := logging.New(nil, logging.INFO)
	handler := Auth(testSecret, log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rec.Code)
	}
}
