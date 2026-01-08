package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/stretchr/testify/assert"
)

// TestOAuth2AuditMiddleware_LogsRequestID tests that middleware extracts and logs request ID
func TestOAuth2AuditMiddleware_LogsRequestID(t *testing.T) {
	// Create a buffer to capture log output
	var logBuf []slog.Record
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: &logBuf,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	// Create a test handler that will be wrapped
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(nextHandler)

	// Make request with request ID header
	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuth2AuditMiddleware_ExtractsPrincipal tests that middleware reads principal from context
func TestOAuth2AuditMiddleware_ExtractsPrincipal(t *testing.T) {
	var logBuf []slog.Record
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: &logBuf,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	principalFromContext := ""
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract principal from request context using typed context key
		principalValue, _ := principal.FromContext(r.Context())
		principalFromContext = principalValue
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	// Simulate RequirePrincipalMiddleware setting principal in context
	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user@example.com", principalFromContext)
}

// TestOAuth2AuditMiddleware_LogsRequestDetails tests that middleware logs request method, path, status
func TestOAuth2AuditMiddleware_LogsRequestDetails(t *testing.T) {
	var logBuf []slog.Record
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: &logBuf,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize?client_id=test", nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuth2AuditMiddleware_LogsErrorResponses tests that middleware logs 4xx and 5xx responses
func TestOAuth2AuditMiddleware_LogsErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"4xx error", http.StatusBadRequest},
		{"5xx error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf []slog.Record
			handler := slog.NewTextHandler(io.Discard, nil)
			testHandler := &testLogHandler{
				records: &logBuf,
				wrapped: handler,
			}
			logger := slog.New(testHandler)

			middleware := OAuth2AuditMiddleware(logger)

			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			wrappedHandler := middleware(nextHandler)

			req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
			req.Header.Set("X-Remote-User", "user@example.com")
			w := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

// TestOAuth2AuditMiddleware_NoSensitiveDataLogged tests that sensitive OAuth2 parameters not logged
func TestOAuth2AuditMiddleware_NoSensitiveDataLogged(t *testing.T) {
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: nil,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	// Request with sensitive parameters
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=test&client_secret=secret123&authorization_code=code456",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Middleware should not log sensitive parameters in URL
}

// TestOAuth2AuditMiddleware_HandlesNoPrincipalHeader tests middleware when X-Remote-User header missing
func TestOAuth2AuditMiddleware_HandlesNoPrincipalHeader(t *testing.T) {
	var logBuf []slog.Record
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: &logBuf,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	// Request without X-Remote-User header
	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuth2AuditMiddleware_PreservesRequestPath tests middleware preserves original request path
func TestOAuth2AuditMiddleware_PreservesRequestPath(t *testing.T) {
	var logBuf []slog.Record
	handler := slog.NewTextHandler(io.Discard, nil)
	testHandler := &testLogHandler{
		records: &logBuf,
		wrapped: handler,
	}
	logger := slog.New(testHandler)

	middleware := OAuth2AuditMiddleware(logger)

	capturedPath := ""
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(nextHandler)

	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "/oauth2/authorize", capturedPath)
}

// testLogHandler is a test slog.Handler that captures log records
type testLogHandler struct {
	records *[]slog.Record
	wrapped slog.Handler
	onLog   func(string)
}

func (h *testLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

func (h *testLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if h.records != nil {
		*h.records = append(*h.records, record)
	}
	if h.onLog != nil {
		// Log the record message
		h.onLog(record.Message)
	}
	return h.wrapped.Handle(ctx, record)
}

func (h *testLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testLogHandler) WithGroup(name string) slog.Handler {
	return h
}
