package api

import (
	"encoding/json"
	"net/http"

	"github.com/markbfromdc/cloudcode/internal/middleware"
)

// APIError represents a standardized JSON error response.
type APIError struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// WriteError writes a standardized JSON error response without a request ID.
func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:    status,
		Message: message,
	})
}

// WriteErrorWithID writes a standardized JSON error response that includes
// the request ID from the request context (if available).
func WriteErrorWithID(w http.ResponseWriter, r *http.Request, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:      status,
		Message:   message,
		RequestID: middleware.GetRequestID(r),
	})
}
