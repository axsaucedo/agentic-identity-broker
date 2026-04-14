// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// LoggingMiddleware returns a middleware that logs HTTP requests with structured logging.
// Logs method, path, response status, and request duration.
//
// logPrefixes defines a whitelist of path prefixes that should be logged (e.g. "/api/").
// Requests whose path does not start with any of the given prefixes are silently passed
// through without producing a log entry, which avoids noise from infrastructure probes
// (e.g. /health) and static-asset serving.
// When no prefixes are provided every request is logged.
func LoggingMiddleware(logger *slog.Logger, logPrefixes ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(logPrefixes) > 0 {
				matched := false
				for _, prefix := range logPrefixes {
					if strings.HasPrefix(r.URL.Path, prefix) {
						matched = true
						break
					}
				}
				if !matched {
					next.ServeHTTP(w, r)
					return
				}
			}

			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Call the next handler
			next.ServeHTTP(wrapped, r)

			// Log after the request completes
			duration := time.Since(start)
			logger.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// RecoveryMiddleware returns a middleware that recovers from panics and logs them.
// Prevents the entire server from crashing due to a single request panic.
func RecoveryMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stack := debug.Stack()
					logger.Error("Panic recovered",
						"error", err,
						"method", r.Method,
						"path", r.URL.Path,
						"remote_addr", r.RemoteAddr,
						"stack", string(stack),
					)

					// Return 500 Internal Server Error
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
