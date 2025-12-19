package consent

import (
	"encoding/json"
	"io"
)

// ErrorResponse represents an error response.
// This is used across all consent handlers for consistent error formatting.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// encodeJSON is a helper function to encode data as JSON to a writer.
// This provides a consistent JSON encoding pattern across all consent handlers.
func encodeJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)
}
