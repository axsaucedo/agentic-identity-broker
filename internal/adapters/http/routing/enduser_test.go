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

func TestSetupEnduserRoutes_ConsentCSRFTokenAllowsSamePrincipalAcrossRemoteAddrChanges(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)
	const principal = "user@example.com"
	csrfCookie := mintCSRFCookie(t, router, testAgentID, principal, "127.0.0.1:12345")

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", principal)
	postReq.Header.Set(httpmiddleware.CSRFTokenHeader, csrfCookie.Value)
	postReq.AddCookie(csrfCookie)
	postReq.RemoteAddr = "127.0.0.1:54321"

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusCreated, postResp.Code)
}

func TestSetupEnduserRoutes_ConsentCSRFTokenRejectsDifferentPrincipalEvenWithSameRemoteAddr(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)
	csrfCookie := mintCSRFCookie(t, router, testAgentID, "user@example.com", "127.0.0.1:12345")

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "other@example.com")
	postReq.Header.Set(httpmiddleware.CSRFTokenHeader, csrfCookie.Value)
	postReq.AddCookie(csrfCookie)
	postReq.RemoteAddr = "127.0.0.1:12345"

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusForbidden, postResp.Code)
}

// T045a: JWKS route is absent (404) when JWKS handler is nil (proxy mode).
// When h.JWKS != nil the route is registered — verified by local/hybrid E2E tests.
func TestSetupEnduserRoutes_JWKSRouteAbsentWhenHandlerNil(t *testing.T) {
	router := chi.NewRouter()
	routing.SetupEnduserRoutes(router, &app.EnduserHandlers{}, routing.EnduserRouteConfig{})

	req := httptest.NewRequest(http.MethodGet, "/oauth2/jwks.json", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func newEnduserConsentRouter(t *testing.T) (http.Handler, string) {
	t.Helper()

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
		if application.Shutdown != nil {
			require.NoError(t, application.Shutdown(context.Background()))
		}
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

	return router, testAgent.ID.String()
}

func mintCSRFCookie(t *testing.T, router http.Handler, agentID, principal, remoteAddr string) *http.Cookie {
	t.Helper()

	getReq := httptest.NewRequest(http.MethodGet, "/api/consent/agents/"+agentID+"/grants", nil)
	getReq.Header.Set("X-Remote-User", principal)
	getReq.RemoteAddr = remoteAddr

	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	require.Equal(t, http.StatusOK, getResp.Code)

	for _, cookie := range getResp.Result().Cookies() {
		if cookie.Name == httpmiddleware.CSRFCookieName {
			require.NotEmpty(t, cookie.Value)
			return cookie
		}
	}

	t.Fatal("expected CSRF cookie to be minted")
	return nil
}

func newGrantRequest(t *testing.T, agentID string) *http.Request {
	t.Helper()

	postBody, err := json.Marshal(map[string]any{
		"granted_permission_sets": map[string][]string{},
	})
	require.NoError(t, err)

	return httptest.NewRequest(
		http.MethodPost,
		"/api/consent/agents/"+agentID+"/grants",
		bytes.NewReader(postBody),
	)
}
