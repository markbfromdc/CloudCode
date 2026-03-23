// Package middleware provides HTTP middleware for the Cloud IDE backend,
// including authentication, rate limiting, and request logging.
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

type contextKey string

const (
	// ContextKeyUserID is the context key for the authenticated user's ID.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyEmail is the context key for the authenticated user's email.
	ContextKeyEmail contextKey = "email"
)

// JWTClaims represents the expected JWT payload claims.
type JWTClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Exp   int64  `json:"exp"`
	Iat   int64  `json:"iat"`
}

// Auth returns middleware that validates JWT tokens from the Authorization header.
// It extracts user identity and injects it into the request context.
func Auth(jwtSecret string, log *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := validateJWT(parts[1], jwtSecret)
			if err != nil {
				log.Warn("JWT validation failed: %v", err)
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.Sub)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateJWT performs a simplified HMAC-SHA256 JWT validation.
// In production, this should be replaced with a proper JWT library
// that supports key rotation and JWK sets.
func validateJWT(token, secret string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Verify signature.
	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, ErrInvalidSignature
	}

	// Decode payload.
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidPayload
	}

	var claims JWTClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidPayload
	}

	// Check expiration.
	if time.Now().Unix() > claims.Exp {
		return nil, ErrTokenExpired
	}

	return &claims, nil
}

// Sentinel errors for JWT validation.
var (
	ErrInvalidToken     = &AuthError{Message: "invalid token format"}
	ErrInvalidSignature = &AuthError{Message: "invalid signature"}
	ErrInvalidPayload   = &AuthError{Message: "invalid payload"}
	ErrTokenExpired     = &AuthError{Message: "token expired"}
)

// AuthError represents an authentication error.
type AuthError struct {
	Message string
}

// Error implements the error interface.
func (e *AuthError) Error() string {
	return e.Message
}
