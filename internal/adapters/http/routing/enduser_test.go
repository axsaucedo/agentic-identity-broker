package routing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	httpmiddleware "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/routing"
	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestSetupEnduserRoutes_GetConsentMintsCSRFCookieForSubsequentGrantPost(t *testing.T) {
	t.Parallel()

	logger := slog.Default()
	cfg := fixtures.DefaultOAuth2Config()

	storage, err := storageadapter.NewAdapter(&cfg.Storage)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, storage.Close(context.Background()))
	})

	testAgent := fixtures.ValidAgent()
	require.NoError(t, storage.Agents().Create(context.Background(), testAgent))

	application, err := app.NewBuilder().
		WithConfig(cfg).
		WithStorage(storage).
		WithLogger(logger).
		Build()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, application.ShutdownTelemetry(context.Background()))
	})

	router := httpadapter.NewHandler(
		httpadapter.ServerConfig{
			Authentication: application.Config.Server.EndUser.Authentication,
		},
		func(r chi.Router) {
			routing.SetupEnduserRoutes(r, application.EnduserHandlers, routing.EnduserRouteConfig{
				Authentication: application.Config.Server.EndUser.Authentication,
				Logger:         logger,
				CORS:           application.Config.Server.EndUser.CORS,
				CSRFStore:      httpmiddleware.NewCSRFStore(logger),
				Telemetry:      application.Config.Telemetry,
			})
		},
		logger,
	)

	const principal = "user@example.com"
	const remoteAddr = "127.0.0.1:12345"

	getReq := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+testAgent.ID.String()+"/grants", nil)
	getReq.Header.Set("X-Remote-User", principal)
	getReq.RemoteAddr = remoteAddr

	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	require.Equal(t, http.StatusOK, getResp.Code)

	var csrfCookie *http.Cookie
	for _, cookie := range getResp.Result().Cookies() {
		if cookie.Name == httpmiddleware.CSRFCookieName {
			csrfCookie = cookie
			break
		}
	}
	require.NotNil(t, csrfCookie)
	require.NotEmpty(t, csrfCookie.Value)

	postBody, err := json.Marshal(map[string]any{
		"delegated_oauth2_tokens": []any{},
	})
	require.NoError(t, err)

	postReq := httptest.NewRequest(
		http.MethodPost,
		"/api/consent/agent/"+testAgent.ID.String()+"/grants",
		bytes.NewReader(postBody),
	)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", principal)
	postReq.Header.Set(httpmiddleware.CSRFTokenHeader, csrfCookie.Value)
	postReq.AddCookie(csrfCookie)
	postReq.RemoteAddr = remoteAddr

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusCreated, postResp.Code)
}
