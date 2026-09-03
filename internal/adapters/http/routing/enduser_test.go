package routing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lestrrat-go/jwx/v4/jwk"

	httpadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	enduserhttp "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/routing"
	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/app"
	"github.com/agentic-identity-broker/agentic-identity-broker/tests/e2e/fixtures"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type mockJWKSPublisher struct{}

func (m *mockJWKSPublisher) PublishJWKS(_ context.Context) (jwk.Set, error) {
	return jwk.NewSet(), nil
}

func TestSetupEnduserRoutes_ConsentAllowsSameOriginPost(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "user@example.com")
	postReq.Header.Set("Sec-Fetch-Site", "same-origin")

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusCreated, postResp.Code)
}

func TestSetupEnduserRoutes_ConsentRejectsCrossSitePost(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "user@example.com")
	postReq.Header.Set("Sec-Fetch-Site", "cross-site")

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusForbidden, postResp.Code)
}

func TestSetupEnduserRoutes_PanicsWhenOAuth2RoutesConfiguredWithoutJWKSHandler(t *testing.T) {
	router := chi.NewRouter()

	require.PanicsWithValue(t,
		"BUG: JWKS handler required when OAuth2 routes are enabled",
		func() {
			routing.SetupEnduserRoutes(router, &app.EnduserHandlers{
				OAuth2Metadata: &enduserhttp.OAuth2MetadataHandler{},
			}, routing.EnduserRouteConfig{})
		},
	)
}

func TestSetupEnduserRoutes_ConsentRejectsCrossOriginPost(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "user@example.com")
	postReq.Header.Set("Origin", "https://evil.example.com")
	postReq.Host = "broker.example.com"

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusForbidden, postResp.Code)
}

func TestSetupEnduserRoutes_ConsentAllowsNonBrowserPost(t *testing.T) {
	t.Parallel()

	router, testAgentID := newEnduserConsentRouter(t)

	postReq := newGrantRequest(t, testAgentID)
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("X-Remote-User", "user@example.com")
	// No Sec-Fetch-Site or Origin headers — non-browser client

	postResp := httptest.NewRecorder()
	router.ServeHTTP(postResp, postReq)

	require.Equal(t, http.StatusCreated, postResp.Code)
}

func TestSetupEnduserRoutes_ApprovalBrowserRoutesRequirePrincipal(t *testing.T) {
	t.Parallel()

	router, _ := newEnduserConsentRouter(t)

	missingPrincipalReq := httptest.NewRequest(http.MethodGet, "/api/approvals/pending", nil)
	missingPrincipalResp := httptest.NewRecorder()
	router.ServeHTTP(missingPrincipalResp, missingPrincipalReq)
	require.Equal(t, http.StatusUnauthorized, missingPrincipalResp.Code)

	authenticatedReq := httptest.NewRequest(http.MethodGet, "/api/approvals/pending", nil)
	authenticatedReq.Header.Set("X-Remote-User", "user@example.com")
	authenticatedResp := httptest.NewRecorder()
	router.ServeHTTP(authenticatedResp, authenticatedReq)
	require.Equal(t, http.StatusOK, authenticatedResp.Code)
}

func TestSetupEnduserRoutes_ApprovalMutationsRejectCrossOriginPost(t *testing.T) {
	t.Parallel()

	router, _ := newEnduserConsentRouter(t)
	for _, action := range []string{"approve", "deny", "revoke"} {
		t.Run(action, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/approvals/00000000-0000-0000-0000-000000000000/"+action, nil)
			req.Header.Set("X-Remote-User", "user@example.com")
			req.Header.Set("Origin", "https://evil.example.com")
			req.Host = "broker.example.com"

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			require.Equal(t, http.StatusForbidden, resp.Code)
		})
	}
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
		WithJWKSPublisher(&mockJWKSPublisher{}).
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
				Authentication:               application.Config.Server.EndUser.Authentication,
				Logger:                       logger,
				ApprovalRequestAuthenticator: application.ApprovalRequestAuthenticator,
				CORS:                         application.Config.Server.EndUser.CORS,
				Telemetry:                    application.Config.Telemetry,
			})
		},
		logger,
	)

	return router, testAgent.ID.String()
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
