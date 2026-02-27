package consent

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGrantsIntegration_CreateUpdateRevoke tests the full lifecycle of a grant
// with real in-memory repositories. This verifies:
// - Grant creation (T075, T076, T079)
// - Grant update (upsert semantics)
// - Grant revocation (empty tokens)
// - Validation (T077)
func TestGrantsIntegration_CreateUpdateRevoke(t *testing.T) {
	// Setup real repositories
	agentRepo := memory.NewAgentRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	grantRepo := memory.NewUserGrantRepository()

	// Create providerService to handle encryption context binding (simulates domain layer)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, slog.Default())

	// Seed test data
	ctx := context.Background()

	// Create agent
	agent := &storage.Agent{
		ID:          "agent-test-123",
		ClientID:    "client-test",
		DisplayName: "Test Agent",
		Description: "Integration test agent",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	// Create GitHub service via providerService
	githubService := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "github",
		DisplayName: "GitHub",
		ClientID:    "github-client",
		Secret:      model.NewPlaintextSecret("github-secret"),
		IssuerURI:   "https://github.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "repo", Description: "Full control of private repositories"},
			{ScopeValue: "user:email", Description: "Access user email addresses"},
			{ScopeValue: "read:user", Description: "Read user profile data"},
		},
	}
	err = providerService.Create(ctx, githubService)
	require.NoError(t, err)

	// Create Google service via providerService
	googleService := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "google",
		DisplayName: "Google",
		ClientID:    "google-client",
		Secret:      model.NewPlaintextSecret("google-secret"),
		IssuerURI:   "https://accounts.google.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth2.googleapis.com/token",
			AuthorizeEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "openid", Description: "OpenID Connect"},
			{ScopeValue: "email", Description: "View email address"},
			{ScopeValue: "profile", Description: "View basic profile info"},
		},
	}
	err = providerService.Create(ctx, googleService)
	require.NoError(t, err)

	// Create consent service
	consentService := consent.NewService(agentRepo, providerService, grantRepo)

	// Create handler
	handler := NewGrantsHandler(consentService, nil)

	// Test 1: Create initial grant
	t.Run("create_grant", func(t *testing.T) {
		futureTime := time.Now().Add(30 * 24 * time.Hour) // 30 days
		reqBody := GrantRequest{
			ValidUntil: &futureTime,
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{
				{
					ThirdpartyOAuth2ServiceID: "github",
					Scopes:                    []string{"repo", "user:email"},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/agent-test-123/grants", bytes.NewBuffer(jsonBody))
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-test-123")
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.CreateGrant(rr, req)

		// Verify response
		assert.Equal(t, http.StatusCreated, rr.Code)

		var envelope map[string]GrantResponse
		err := json.NewDecoder(rr.Body).Decode(&envelope)
		require.NoError(t, err)

		response, ok := envelope["data"]
		require.True(t, ok, "expected 'data' field in response")

		assert.NotEmpty(t, response.ID)
		assert.Equal(t, "alice@example.com", response.Principal)
		assert.Equal(t, "agent-test-123", response.AgentID)
		assert.NotNil(t, response.ValidUntil)
		assert.Len(t, response.DelegatedOAuth2Tokens, 1)
		assert.Equal(t, "github", response.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
		assert.ElementsMatch(t, []string{"repo", "user:email"}, response.DelegatedOAuth2Tokens[0].Scopes)
	})

	// Test 2: Update grant (upsert semantics)
	t.Run("update_grant", func(t *testing.T) {
		futureTime := time.Now().Add(60 * 24 * time.Hour) // 60 days
		reqBody := GrantRequest{
			ValidUntil: &futureTime,
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{
				{
					ThirdpartyOAuth2ServiceID: "github",
					Scopes:                    []string{"repo", "user:email", "read:user"},
				},
				{
					ThirdpartyOAuth2ServiceID: "google",
					Scopes:                    []string{"openid", "email"},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/agent-test-123/grants", bytes.NewBuffer(jsonBody))
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-test-123")
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.CreateGrant(rr, req)

		// Verify response
		assert.Equal(t, http.StatusCreated, rr.Code)

		var envelope map[string]GrantResponse
		err := json.NewDecoder(rr.Body).Decode(&envelope)
		require.NoError(t, err)

		response, ok := envelope["data"]
		require.True(t, ok, "expected 'data' field in response")

		// Grant was updated, not created (same principal+agent)
		assert.Equal(t, "alice@example.com", response.Principal)
		assert.Equal(t, "agent-test-123", response.AgentID)
		assert.Len(t, response.DelegatedOAuth2Tokens, 2)
	})

	// Test 3: Retrieve grants
	t.Run("get_grants", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/consent/agent/agent-test-123/grants", nil)
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-test-123")
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.GetGrants(rr, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rr.Code)

		var envelope map[string][]GrantResponse
		err := json.NewDecoder(rr.Body).Decode(&envelope)
		require.NoError(t, err)

		response, ok := envelope["data"]
		require.True(t, ok, "expected 'data' field in response")

		assert.Len(t, response, 1) // Only one grant (upserted)
		assert.Equal(t, "alice@example.com", response[0].Principal)
		assert.Equal(t, "agent-test-123", response[0].AgentID)
		assert.Len(t, response[0].DelegatedOAuth2Tokens, 2)
	})

	// Test 4: Revoke grant (empty tokens)
	t.Run("revoke_grant", func(t *testing.T) {
		reqBody := GrantRequest{
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{}, // Empty = revoke
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/agent-test-123/grants", bytes.NewBuffer(jsonBody))
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-test-123")
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.CreateGrant(rr, req)

		// Verify response
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	// Test 5: Verify grant is gone
	t.Run("verify_revoked", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/consent/agent/agent-test-123/grants", nil)
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-test-123")
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.GetGrants(rr, req)

		// Verify response
		assert.Equal(t, http.StatusOK, rr.Code)

		var envelope map[string][]GrantResponse
		err := json.NewDecoder(rr.Body).Decode(&envelope)
		require.NoError(t, err)

		response, ok := envelope["data"]
		require.True(t, ok, "expected 'data' field in response")

		assert.Empty(t, response) // No grants after revocation
	})
}

// TestGrantsIntegration_Validation tests validation logic (T077)
func TestGrantsIntegration_Validation(t *testing.T) {
	// Setup repositories
	agentRepo := memory.NewAgentRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	grantRepo := memory.NewUserGrantRepository()

	// Create providerService to handle encryption context binding (simulates domain layer)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, slog.Default())

	ctx := context.Background()

	// Create agent
	agent := &storage.Agent{
		ID:          "agent-validate",
		ClientID:    "client-validate",
		DisplayName: "Validation Test Agent",
		Description: "Test validation",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	// Create service with limited scopes via providerService
	service := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service",
		DisplayName: "Test Service",
		ClientID:    "test-client",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://example.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://example.com/token",
			AuthorizeEndpoint: "https://example.com/auth",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
			{ScopeValue: "write", Description: "Write access"},
		},
	}
	err = providerService.Create(ctx, service)
	require.NoError(t, err)

	consentService := consent.NewService(agentRepo, providerService, grantRepo)
	handler := NewGrantsHandler(consentService, nil)

	tests := []struct {
		name           string
		agentID        string
		reqBody        GrantRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "invalid_scope",
			agentID: "agent-validate",
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: "test-service",
						Scopes:                    []string{"invalid-scope"},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid scopes",
		},
		{
			name:    "nonexistent_service",
			agentID: "agent-validate",
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: "nonexistent-service",
						Scopes:                    []string{"read"},
					},
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal server error",
		},
		{
			name:    "nonexistent_agent",
			agentID: "nonexistent-agent",
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: "test-service",
						Scopes:                    []string{"read"},
					},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "agent not found",
		},
		{
			name:    "valid_request",
			agentID: "agent-validate",
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: "test-service",
						Scopes:                    []string{"read", "write"},
					},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest("POST", "/api/consent/agent/"+tt.agentID+"/grants", bytes.NewBuffer(jsonBody))
			//nolint:staticcheck // Using string key for test simplicity
			ctx := principal.WithPrincipal(req.Context(), "test@example.com")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("agent-id", tt.agentID)
			req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.CreateGrant(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				var errResp ErrorResponse
				err := json.NewDecoder(rr.Body).Decode(&errResp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedError, errResp.Error)
			}
		})
	}
}
