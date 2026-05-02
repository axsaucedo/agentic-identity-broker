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
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, false, slog.Default())

	// Seed test data
	ctx := context.Background()

	testAgentID := id.NewAgentID()
	githubServiceID := id.NewServiceID()
	googleServiceID := id.NewServiceID()

	// Create agent
	agent := &storage.Agent{
		ID:          testAgentID,
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
		ID:          githubServiceID,
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
		ID:          googleServiceID,
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
	consentService := consent.NewService(agentRepo, providerService, grantRepo, slog.Default())

	// Create handler
	handler := NewGrantsHandler(consentService, nil)

	// Test 1: Create initial grant
	t.Run("create_grant", func(t *testing.T) {
		futureTime := time.Now().Add(30 * 24 * time.Hour) // 30 days
		reqBody := GrantRequest{
			ValidUntil: &futureTime,
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{
				{
					ThirdpartyOAuth2ServiceID: githubServiceID.String(),
					Scopes:                    []string{"repo", "user:email"},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", bytes.NewBuffer(jsonBody))
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", testAgentID.String())
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
		assert.Equal(t, testAgentID.String(), response.AgentID)
		assert.NotNil(t, response.ValidUntil)
		assert.Len(t, response.DelegatedOAuth2Tokens, 1)
		assert.Equal(t, githubServiceID.String(), response.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
		assert.ElementsMatch(t, []string{"repo", "user:email"}, response.DelegatedOAuth2Tokens[0].Scopes)
	})

	// Test 2: Update grant (upsert semantics)
	t.Run("update_grant", func(t *testing.T) {
		futureTime := time.Now().Add(60 * 24 * time.Hour) // 60 days
		reqBody := GrantRequest{
			ValidUntil: &futureTime,
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{
				{
					ThirdpartyOAuth2ServiceID: githubServiceID.String(),
					Scopes:                    []string{"repo", "user:email", "read:user"},
				},
				{
					ThirdpartyOAuth2ServiceID: googleServiceID.String(),
					Scopes:                    []string{"openid", "email"},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", bytes.NewBuffer(jsonBody))
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", testAgentID.String())
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
		assert.Equal(t, testAgentID.String(), response.AgentID)
		assert.Len(t, response.DelegatedOAuth2Tokens, 2)
	})

	// Test 3: Retrieve grants
	t.Run("get_grants", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/consent/agent/"+testAgentID.String()+"/grants", nil)
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", testAgentID.String())
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
		assert.Equal(t, testAgentID.String(), response[0].AgentID)
		assert.Len(t, response[0].DelegatedOAuth2Tokens, 2)
	})

	// Test 4: Revoke grant via DELETE /grants
	t.Run("revoke_grant", func(t *testing.T) {
		revokeHandler := NewRevokeGrantHandler(consentService, nil)

		req := httptest.NewRequest("DELETE", "/api/consent/agent/"+testAgentID.String()+"/grants", nil)
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", testAgentID.String())
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		revokeHandler.RevokeGrant(rr, req)

		// Verify response
		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	// Test 5: Verify grant is gone
	t.Run("verify_revoked", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/consent/agent/"+testAgentID.String()+"/grants", nil)
		//nolint:staticcheck // Using string key for test simplicity
		ctx := principal.WithPrincipal(req.Context(), "alice@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", testAgentID.String())
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

func TestGrantsIntegration_SessionReplayDoesNotMutateStoredGrant(t *testing.T) {
	agentRepo := memory.NewAgentRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	grantRepo := memory.NewUserGrantRepository()

	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, false, slog.Default())

	ctx := context.Background()
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()
	principalValue := id.Principal("alice@example.com")
	now := time.Now()

	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-test",
		DisplayName: "Test Agent",
		Description: "Integration test agent",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	require.NoError(t, agentRepo.Create(ctx, agent))

	service := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          serviceID,
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
			{ScopeValue: "read:user", Description: "Read user profile data"},
		},
	}
	require.NoError(t, providerService.Create(ctx, service))

	session := &storage.AuthorizationSession{
		SessionID:   "replay-race-session",
		AgentID:     agentID,
		Principal:   principalValue,
		OriginalURL: "/callback",
		CreatedAt:   now,
		ExpiresAt:   now.Add(10 * time.Minute),
	}
	consumeCalls := 0
	authSessionRepo := &mockAuthSessionRepo{
		getBySessionIDFunc: func(_ context.Context, _ string) (*storage.AuthorizationSession, error) {
			return session, nil
		},
		consumeFunc: func(_ context.Context, _ string) error {
			consumeCalls++
			if consumeCalls == 1 {
				return nil
			}
			return storage.NewStorageError("Consume", storage.ErrorKindConflict, nil, "authorization session has already been consumed")
		},
	}

	consentService := consent.NewService(agentRepo, providerService, grantRepo, slog.Default())
	handler := NewGrantsHandler(consentService, nil).WithAuthorizationSessionRepository(authSessionRepo)

	firstValidUntil := time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)
	secondValidUntil := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)

	newRequest := func(validUntil *time.Time, scopes []string) *http.Request {
		reqBody := GrantRequest{
			ValidUntil: validUntil,
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{
				{
					ThirdpartyOAuth2ServiceID: serviceID.String(),
					Scopes:                    scopes,
				},
			},
		}

		jsonBody, marshalErr := json.Marshal(reqBody)
		require.NoError(t, marshalErr)

		req := httptest.NewRequest("POST", "/api/consent/agent/"+agentID.String()+"/grants?session_id="+session.SessionID, bytes.NewBuffer(jsonBody))
		req = req.WithContext(principal.WithPrincipal(req.Context(), principalValue.String()))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	}

	firstResponse := httptest.NewRecorder()
	handler.CreateGrant(firstResponse, newRequest(&firstValidUntil, []string{"repo"}))
	require.Equal(t, http.StatusCreated, firstResponse.Code, firstResponse.Body.String())

	storedGrant, err := grantRepo.FindByPrincipalAndAgent(ctx, principalValue, agentID)
	require.NoError(t, err)
	require.NotNil(t, storedGrant)
	assert.Equal(t, &firstValidUntil, storedGrant.ValidUntil)
	require.Len(t, storedGrant.DelegatedOAuth2Tokens, 1)
	assert.Equal(t, []string{"repo"}, storedGrant.DelegatedOAuth2Tokens[0].Scopes)

	replayResponse := httptest.NewRecorder()
	handler.CreateGrant(replayResponse, newRequest(&secondValidUntil, []string{"repo", "read:user"}))
	assert.Equal(t, http.StatusBadRequest, replayResponse.Code, replayResponse.Body.String())

	storedGrantAfterReplay, err := grantRepo.FindByPrincipalAndAgent(ctx, principalValue, agentID)
	require.NoError(t, err)
	require.NotNil(t, storedGrantAfterReplay)
	assert.Equal(t, storedGrant.ID, storedGrantAfterReplay.ID)
	assert.Equal(t, &firstValidUntil, storedGrantAfterReplay.ValidUntil)
	require.Len(t, storedGrantAfterReplay.DelegatedOAuth2Tokens, 1)
	assert.Equal(t, []string{"repo"}, storedGrantAfterReplay.DelegatedOAuth2Tokens[0].Scopes)
}

// TestGrantsIntegration_OptionalOnlyAgent verifies that agents with only optional service
// requirements accept approval with no delegated tokens. Empty tokens create a grant with
// no delegations (201 Created) rather than triggering a revoke.
func TestGrantsIntegration_OptionalOnlyAgent(t *testing.T) {
	agentRepo := memory.NewAgentRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	grantRepo := memory.NewUserGrantRepository()

	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, false, slog.Default())
	ctx := context.Background()

	// Agent with only optional service requirements
	optionalAgentID := id.NewAgentID()
	optionalServiceID := id.NewServiceID()

	optionalAgent := &storage.Agent{
		ID:          optionalAgentID,
		ClientID:    "client-optional",
		DisplayName: "Optional-Only Agent",
		Description: "Agent with only optional service requirements",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       optionalServiceID,
				RequirementType: storage.RequirementTypeOptional,
				RequiredScopes:  []string{"read"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := agentRepo.Create(ctx, optionalAgent)
	require.NoError(t, err)

	optionalService := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          optionalServiceID,
		DisplayName: "Optional Service",
		ClientID:    "opt-client",
		Secret:      model.NewPlaintextSecret("opt-secret"),
		IssuerURI:   "https://optional.example.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://optional.example.com/token",
			AuthorizeEndpoint: "https://optional.example.com/auth",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
	}
	err = providerService.Create(ctx, optionalService)
	require.NoError(t, err)

	consentService := consent.NewService(agentRepo, providerService, grantRepo, slog.Default())
	handler := NewGrantsHandler(consentService, nil)

	// Approval with no selected services creates a grant with empty delegations (201).
	t.Run("approve_with_no_services_optional_only_agent", func(t *testing.T) {
		reqBody := GrantRequest{
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/"+optionalAgentID.String()+"/grants", bytes.NewBuffer(jsonBody))
		ctx := principal.WithPrincipal(req.Context(), "bob@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", optionalAgentID.String())
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.CreateGrant(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, optionalAgentID.String(), data["agent_id"])
		assert.Empty(t, data["delegated_oauth2_tokens"])
	})

	// Empty tokens with redirect_uri: creates grant and honours the redirect.
	t.Run("empty_tokens_with_redirect_uri_creates_grant_and_redirects", func(t *testing.T) {
		reqBody := GrantRequest{
			DelegatedOAuth2Tokens: []DelegatedTokenRequest{},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/consent/agent/"+optionalAgentID.String()+"/grants?redirect_uri=%2Fcallback", bytes.NewBuffer(jsonBody))
		ctx := principal.WithPrincipal(req.Context(), "bob@example.com")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", optionalAgentID.String())
		req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.CreateGrant(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp map[string]interface{}
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
		assert.Equal(t, "/callback", resp["redirect_url"])
	})
}

// TestGrantsIntegration_Validation tests validation logic (T077)
func TestGrantsIntegration_Validation(t *testing.T) {
	// Setup repositories
	agentRepo := memory.NewAgentRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	grantRepo := memory.NewUserGrantRepository()

	// Create providerService to handle encryption context binding (simulates domain layer)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(serviceRepo, newTestEncryption(), nil, false, slog.Default())

	ctx := context.Background()

	testAgentID := id.NewAgentID()
	testServiceID := id.NewServiceID()

	// Create agent
	agent := &storage.Agent{
		ID:          testAgentID,
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
		ID:          testServiceID,
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

	consentService := consent.NewService(agentRepo, providerService, grantRepo, slog.Default())
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
			agentID: testAgentID.String(),
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: testServiceID.String(),
						Scopes:                    []string{"invalid-scope"},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid scopes",
		},
		{
			name:    "nonexistent_service",
			agentID: testAgentID.String(),
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
						Scopes:                    []string{"read"},
					},
				},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal server error",
		},
		{
			name:    "nonexistent_agent",
			agentID: id.NewAgentID().String(),
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: testServiceID.String(),
						Scopes:                    []string{"read"},
					},
				},
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "agent not found",
		},
		{
			name:    "valid_request",
			agentID: testAgentID.String(),
			reqBody: GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: testServiceID.String(),
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
