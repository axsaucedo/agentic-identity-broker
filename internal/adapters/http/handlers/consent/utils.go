package consent

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

// ErrorResponse represents an error response.
// This is used across all consent handlers for consistent error formatting.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// validateRedirectURI returns nil only for relative URIs (no scheme or host).
// Absolute URIs are always rejected — callers should silently ignore the
// redirect_uri rather than returning an error to the client.
func validateRedirectURI(rawURI string) error {
	u, err := url.Parse(rawURI)
	if err != nil {
		return errors.New("redirect_uri is not parseable")
	}
	if u.Scheme != "" || u.Host != "" {
		return errors.New("redirect_uri must be a relative path, not an absolute URL")
	}
	return nil
}

// writeBufferedJSON marshals data, then writes headers and body atomically.
// If marshalling fails it sends a 500; the caller's WriteHeader is never flushed
// with an empty body.
func writeBufferedJSON(w http.ResponseWriter, statusCode int, data any) error {
	body, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal_error","message":"response encoding failed"}`))
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(body)
	return err
}
