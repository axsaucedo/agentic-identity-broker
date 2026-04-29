package consent

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents an error response.
// This is used across all consent handlers for consistent error formatting.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// writeBufferedJSON marshals data, then writes headers and body atomically.
// If marshalling fails it sends a 500; the caller's WriteHeader is never flushed
// with an empty body.
func writeBufferedJSON(w http.ResponseWriter, statusCode int, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		http.Error(w, `{"error":"internal_error","message":"response encoding failed"}`, http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(body)
	return err
}
