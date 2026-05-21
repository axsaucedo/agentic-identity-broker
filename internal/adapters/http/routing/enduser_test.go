package routing_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/routing"
	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestSetupEnduserRoutes_ConsentCSRFTokenAllowsMutatingRequest(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)
	const principal = "user@example.com"
	internalCookie, maskedToken := mintCSRFToken(t, router, testAgentID, principal)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", principal)
	postReq.Header.Set("X-CSRF-Token", maskedToken)
	postReq.AddCookie(internalCookie)

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusCreated, postResp.Code)
}

func TestSetupEnduserRoutes_ConsentCSRFRejectsMissingToken(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "user@example.com")

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusForbidden, postResp.Code)
}

func TestSetupEnduserRoutes_ConsentCSRFRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)
	const principal = "user@example.com"
	internalCookie, _ := mintCSRFToken(t, router, testAgentID, principal)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", principal)
	postReq.Header.Set("X-CSRF-Token", "totally-invalid-token")
	postReq.AddCookie(internalCookie)

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusForbidden, postResp.Code)
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

	csrfKey := make([]byte, 32)
	_, err = rand.Read(csrfKey)
	require.NoError(t, err)

	router := httpadapter.NewHandler(
		httpadapter.ServerConfig{
			Authentication: application.Config.Server.EndUser.Authentication,
		},
		func(r chi.Router) {
			routing.SetupEnduserRoutes(r, application.EnduserHandlers, routing.EnduserRouteConfig{
				Authentication: application.Config.Server.EndUser.Authentication,
				Logger:         logger,
				CORS:           application.Config.Server.EndUser.CORS,
				CSRFKey:        csrfKey,
				CSRFSecure:     false,
				Telemetry:      application.Config.Telemetry,
			})
		},
		logger,
	)

	return router, testAgent.ID.String()
}

func mintCSRFToken(t *testing.T, router http.Handler, agentID, principal string) (*http.Cookie, string) {
	t.Helper()

	getReq := httptest.NewRequest(http.MethodGet, "/api/consent/agents/"+agentID+"/grants", nil)
	getReq.Header.Set("X-Remote-User", principal)

	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	require.Equal(t, http.StatusOK, getResp.Code)

	var internalCookie *http.Cookie
	var maskedToken string
	for _, cookie := range getResp.Result().Cookies() {
		switch cookie.Name {
		case "_csrf":
			internalCookie = cookie
		case "csrf_token":
			maskedToken = cookie.Value
		}
	}

	require.NotNil(t, internalCookie, "expected _csrf cookie to be set")
	require.NotEmpty(t, maskedToken, "expected csrf_token cookie to be set")

	return internalCookie, maskedToken
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
