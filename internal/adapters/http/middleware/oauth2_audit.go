package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// OAuth2AuditMiddleware creates audit logging middleware for OAuth2 endpoints
func OAuth2AuditMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract request ID from header
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Extract principal from context (set by RequirePrincipalMiddleware)
			// Note: May be empty for public endpoints (metadata, token proxy)
			principalValue, _ := principal.FromContext(r.Context())

			// Record request start time
			startTime := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriterWrapper{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Add request_id to context
			ctx := context.WithValue(r.Context(), "request_id", requestID)

			// Call next handler with updated context
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Log audit information
			duration := time.Since(startTime).Milliseconds()
			logAuditEvent(logger, &auditEvent{
				RequestID:  requestID,
				Principal:  principalValue,
				Method:     r.Method,
				Path:       r.URL.Path,
				StatusCode: wrapped.statusCode,
				Duration:   duration,
				RemoteAddr: r.RemoteAddr,
				UserAgent:  r.Header.Get("User-Agent"),
			})
		})
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

// auditEvent contains audit information about a request
type auditEvent struct {
	RequestID  string
	Principal  string
	Method     string
	Path       string
	StatusCode int
	Duration   int64
	RemoteAddr string
	UserAgent  string
}

// logAuditEvent logs an audit event with structured logging
func logAuditEvent(logger *slog.Logger, event *auditEvent) {
	// Determine log level based on status code
	level := slog.LevelInfo
	if event.StatusCode >= 400 && event.StatusCode < 500 {
		level = slog.LevelWarn
	} else if event.StatusCode >= 500 {
		level = slog.LevelError
	}

	// Get the remote host without port
	host, _, _ := net.SplitHostPort(event.RemoteAddr)
	if host == "" {
		host = event.RemoteAddr
	}

	logger.Log(
		context.Background(),
		level,
		"OAuth2 Authorization Request",
		slog.String("request_id", event.RequestID),
		slog.String("principal", event.Principal),
		slog.String("method", event.Method),
		slog.String("path", event.Path),
		slog.Int("status", event.StatusCode),
		slog.Int64("duration_ms", event.Duration),
		slog.String("remote_host", host),
		slog.String("user_agent", event.UserAgent),
	)
}

// generateRequestID generates a simple request ID
// In production, this could use UUID or other ID generation
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(time.Now().UnixNano()%10000))
}
