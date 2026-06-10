package http

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleHealth_IncludesComponentHealth(t *testing.T) {
	server := &Server{
		config: ServerConfig{
			HealthComponents: func() map[string]string {
				return map[string]string{"upstream_jwks": "degraded"}
			},
		},
		healthState: int32(HealthStateHealthy),
		startTime:   time.Now().Add(-2 * time.Minute),
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	server.handleHealth().ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.Equal(t, "healthy", response.Status)
	assert.Equal(t, "degraded", response.Components["upstream_jwks"])
}

func TestHandleHealth_UsesCurrentServerStartTime(t *testing.T) {
	server := &Server{
		healthState: int32(HealthStateHealthy),
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	handler := server.handleHealth()
	server.startTime = time.Now().Add(-2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response HealthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
	assert.GreaterOrEqual(t, response.UptimeSeconds, int64(1))
	assert.Less(t, response.UptimeSeconds, int64(10))
}
