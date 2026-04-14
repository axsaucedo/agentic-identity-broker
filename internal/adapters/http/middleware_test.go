// Package http provides HTTP server adapters for the identity broker.
package http

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
