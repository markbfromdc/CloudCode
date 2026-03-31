package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ContextKeyRequestID is the context key for the request ID.
const ContextKeyRequestID contextKey = "request_id"

// RequestID returns middleware that generates a unique request ID for each
// request, sets it in the response header X-Request-ID, and makes it
// available in the request context under ContextKeyRequestID.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ContextKeyRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the request context.
// Returns an empty string if no request ID is present.
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(ContextKeyRequestID).(string); ok {
		return id
	}
	return ""
}
