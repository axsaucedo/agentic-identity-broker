// Package http provides HTTP server adapters for the identity broker.
package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const (
	// maxPrincipalLength is the maximum allowed length for a principal value
	maxPrincipalLength = 200
)

// ErrorResponse represents a JSON error response structure.
type ErrorResponse struct {
	Error string `json:"error"`
}

// RequirePrincipalMiddleware returns middleware that requires a valid principal to be present.
// If the principal header is missing, empty, or invalid, the request is rejected with an error.
// If valid, the principal is extracted, trimmed, validated, and added to the request context.
//
// Returns:
// - 401 Unauthorized: When the principal header is missing or empty
// - 400 Bad Request: When the principal exceeds maximum length
// - Continues to next handler: When principal is valid
func RequirePrincipalMiddleware(authConfig ports.AuthenticationConfig, logger *slog.Logger) func(next http.Handler) http.Handler {
	headerName := authConfig.Preauth.PrincipalHeaderName

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the principal from the configured header
			principalValue := r.Header.Get(headerName)

			// Trim leading and trailing whitespace
			principalValue = strings.TrimSpace(principalValue)

			// Validate: principal must be present and non-empty
			if principalValue == "" {
				err := principal.NewMissingPrincipalError(headerName)
				logger.Warn("Missing or empty principal (401 REJECT)",
					"header", headerName,
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path)

				writeErrorJSON(w, http.StatusUnauthorized, err.Error())
				return
			}

			// Validate: principal must not exceed maximum length
			if len(principalValue) > maxPrincipalLength {
				err := principal.PrincipalTooLongError(len(principalValue), maxPrincipalLength)
				logger.Warn("Principal exceeds maximum length",
					"header", headerName,
					"length", len(principalValue),
					"max_length", maxPrincipalLength,
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path)

				writeErrorJSON(w, http.StatusBadRequest, err.Error())
				return
			}

			// Add principal to request context
			ctx := principal.WithPrincipal(r.Context(), principalValue)
			r = r.WithContext(ctx)

			logger.Debug("Principal extracted",
				"header", headerName,
				"principal", principalValue,
				"remote_addr", r.RemoteAddr,
				"path", r.URL.Path)

			// Continue to the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// OptionalPrincipalMiddleware returns middleware that optionally extracts a principal.
// If the principal header is present and valid, it is extracted and added to the request context.
// If the header is missing, empty, or invalid, the request continues without a principal.
// This middleware never rejects requests - it only adds principals if they are valid.
//
// Returns:
// - Continues to next handler: In all cases (principal may or may not be in context)
func OptionalPrincipalMiddleware(authConfig ports.AuthenticationConfig, logger *slog.Logger) func(next http.Handler) http.Handler {
	headerName := authConfig.Preauth.PrincipalHeaderName

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract the principal from the configured header
			principalValue := r.Header.Get(headerName)

			// Trim leading and trailing whitespace
			principalValue = strings.TrimSpace(principalValue)

			// If principal is present and valid, add it to context
			if principalValue != "" && len(principalValue) <= maxPrincipalLength {
				ctx := principal.WithPrincipal(r.Context(), principalValue)
				r = r.WithContext(ctx)

				logger.Debug("Optional principal extracted",
					"header", headerName,
					"principal", principalValue,
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path)
			} else if principalValue != "" && len(principalValue) > maxPrincipalLength {
				// Log but don't reject: just skip adding invalid principal to context
				logger.Debug("Optional principal ignored (exceeds max length)",
					"header", headerName,
					"length", len(principalValue),
					"max_length", maxPrincipalLength,
					"remote_addr", r.RemoteAddr,
					"path", r.URL.Path)
			}
			// If principalValue is empty after trim, we simply don't add anything to context
			// and continue without a principal (which is fine for optional routes)

			// Continue to the next handler (always)
			next.ServeHTTP(w, r)
		})
	}
}

// writeErrorJSON writes a JSON error response to the response writer.
// Sets Content-Type to application/json and writes the provided HTTP status code.
func writeErrorJSON(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{Error: errorMessage}
	_ = json.NewEncoder(w).Encode(response)
}
