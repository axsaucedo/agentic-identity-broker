// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/telemetry"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func TestLoggingMiddlewareLogsWhitelistedPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := LoggingMiddleware(logger, "/api/")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(buf.String(), "HTTP request") {
		t.Errorf("expected log output for /api/ path, got: %s", buf.String())
	}
}

func TestLoggingMiddlewareSkipsNonWhitelistedPaths(t *testing.T) {
	paths := []string{"/health", "/consent/index.html", "/favicon.ico"}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			handler := LoggingMiddleware(logger, "/api/", "/oauth2/", "/.well-known/")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
			}
			if buf.Len() != 0 {
				t.Errorf("expected no log output for %q, got: %s", path, buf.String())
			}
		})
	}
}

func TestLoggingMiddlewareLogsAllPathsWhenNoPrefixesGiven(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	handler := LoggingMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(buf.String(), "HTTP request") {
		t.Errorf("expected log output when no prefixes given, got: %s", buf.String())
	}
}

func TestFirstValidTraceparent_IgnoresOtherPropagationFormats(t *testing.T) {
	previousPropagator := otel.GetTextMapPropagator()
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, b3.New()))
	t.Cleanup(func() { otel.SetTextMapPropagator(previousPropagator) })

	const validTraceparent = "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01"
	header := http.Header{}
	header.Add("traceparent", "not-a-valid-traceparent")
	header.Add("traceparent", validTraceparent)
	header.Set("b3", "cccccccccccccccccccccccccccccccc-dddddddddddddddd-1")

	got, ok := firstValidTraceparent(header, header.Values("traceparent"))
	require.True(t, ok)
	assert.Equal(t, validTraceparent, got)
}

type panicLogRecord struct {
	message string
	attrs   map[string]any
}

type panicLogCaptureHandler struct {
	records *[]panicLogRecord
	attrs   []slog.Attr
	groups  []string
}

func newPanicLogCaptureHandler() *panicLogCaptureHandler {
	records := make([]panicLogRecord, 0, 1)
	return &panicLogCaptureHandler{records: &records}
}

func (h *panicLogCaptureHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *panicLogCaptureHandler) Handle(_ context.Context, record slog.Record) error {
	attrs := make(map[string]any, record.NumAttrs()+len(h.attrs))
	for _, attr := range h.attrs {
		h.storeAttr(attrs, attr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		h.storeAttr(attrs, attr)
		return true
	})
	*h.records = append(*h.records, panicLogRecord{message: record.Message, attrs: attrs})
	return nil
}

func (h *panicLogCaptureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

func (h *panicLogCaptureHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string{}, h.groups...), name)
	return &clone
}

func (h *panicLogCaptureHandler) storeAttr(dst map[string]any, attr slog.Attr) {
	key := attr.Key
	if len(h.groups) > 0 {
		key = strings.Join(append(append([]string{}, h.groups...), attr.Key), ".")
	}
	dst[key] = attr.Value.Any()
}

func TestRecoveryMiddleware_PanicLogCarriesSecurityContextAndTraceHeader(t *testing.T) {
	t.Parallel()

	base := newPanicLogCaptureHandler()
	logger := slog.New(telemetry.NewContextHandler(base))
	handler := RecoveryMiddleware(logger, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/panic", nil).WithContext(security.WithSecurityContext(
		context.Background(),
		security.SecurityContext{
			TraceID:     "0123456789abcdef0123456789abcdef",
			Actor:       "alice@example.com",
			CallingPeer: "gateway-client-1",
		},
	))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Len(t, *base.records, 1)
	assert.Equal(t, "Panic recovered", (*base.records)[0].message)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", (*base.records)[0].attrs["trace_id"], "panic log must retain request trace correlation")
	assert.Equal(t, "alice@example.com", (*base.records)[0].attrs["actor"], "panic log must retain the finalized actor")
	assert.Equal(t, "gateway-client-1", (*base.records)[0].attrs["calling_peer"], "panic log must retain the finalized calling_peer")
	assert.NotEmpty(t, rr.Header().Get("traceresponse"), "recovered 500 responses must still carry traceresponse")
}

func TestRecoveryMiddleware_SuppressesTraceResponseHeaderWhenDisabled(t *testing.T) {
	t.Parallel()

	base := newPanicLogCaptureHandler()
	logger := slog.New(telemetry.NewContextHandler(base))
	handler := RecoveryMiddleware(logger, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/panic", nil).WithContext(security.WithSecurityContext(
		context.Background(),
		security.SecurityContext{
			TraceID: "0123456789abcdef0123456789abcdef",
			Actor:   "alice@example.com",
		},
	))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Len(t, *base.records, 1)
	assert.Equal(t, "0123456789abcdef0123456789abcdef", (*base.records)[0].attrs["trace_id"],
		"panic log must still retain trace correlation even when the response header is disabled")
	assert.Empty(t, rr.Header().Get("traceresponse"),
		"recovered 500 responses must honor trace.response_enabled=false and suppress the traceresponse header")
}
