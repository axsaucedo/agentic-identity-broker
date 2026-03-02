package consent_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	memorystorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/testutil"
	"github.com/go-chi/chi/v5"
)

func newIntegrationProviderService(t *testing.T) *thirdparty.ThirdpartyOAuth2ProviderService {
	t.Helper()
	repo := memorystorage.NewInMemoryThirdpartyOAuth2ProviderRepository()
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, testutil.NewTestEncryptionAdapter(t), nil, false, slog.Default())
}

func newGitHubServiceEntity() *model.ThirdpartyOAuth2ProviderEntity {
	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "github",
		DisplayName: "GitHub",
		ClientID:    "github-client-id",
		Secret:      model.NewPlaintextSecret("github-client-secret"),
		IssuerURI:   "https://github.com",
		Discovery:   model.DiscoveryConfig{EnableDiscovery: false},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/oauth/token",
			AuthorizeEndpoint: "https://github.com/oauth/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read:user", Description: "Read user profile"},
			{ScopeValue: "repo", Description: "Full control of repositories"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// TestIntegration_GetAgentDetail exercises the GET /api/consent/agent/:agentId endpoint
// with real in-memory storage and encryption. No service requirements → Services list is empty.
func TestIntegration_GetAgentDetail(t *testing.T) {
	agentRepo := memorystorage.NewAgentRepository()
	providerService := newIntegrationProviderService(t)
	grantRepo := memorystorage.NewUserGrantRepository()

	ctx := context.Background()
	principalValue := "user@example.com"

	govURL := "https://example.com/governance"
	docsURL := "https://example.com/docs"
	agent := &storage.Agent{
		ID:                   "agent-123",
		ClientID:             "client-123",
		DisplayName:          "Example AI Agent",
		Description:          "An example AI agent for demonstrations",
		GovernanceURL:        &govURL,
		UserDocumentationURL: &docsURL,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	if err := providerService.Create(ctx, newGitHubServiceEntity()); err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	consentSvc := consentservice.NewService(agentRepo, providerService, grantRepo)
	handler := consent.NewAgentDetailHandler(consentSvc, nil).
		WithAgentRepository(agentRepo)

	reqCtx := principal.WithPrincipal(ctx, principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	reqCtx = context.WithValue(reqCtx, chi.RouteCtxKey, rctx)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123", nil)
	req = req.WithContext(reqCtx)

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response consent.GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data.Agent.AgentID != "agent-123" {
		t.Errorf("expected agent ID %q, got %q", "agent-123", response.Data.Agent.AgentID)
	}
	if response.Data.Agent.DisplayName != "Example AI Agent" {
		t.Errorf("expected display name %q, got %q", "Example AI Agent", response.Data.Agent.DisplayName)
	}
	if response.Data.Agent.Description != "An example AI agent for demonstrations" {
		t.Errorf("expected description %q, got %q", "An example AI agent for demonstrations", response.Data.Agent.Description)
	}
	// Agent has no ServiceRequirements, so the services list must be empty.
	if len(response.Data.Services) != 0 {
		t.Errorf("expected 0 services (no requirements declared), got %d", len(response.Data.Services))
	}
}

// TestIntegration_GetAgentGrants exercises the GET /api/consent/agent/:agentId/grants endpoint.
func TestIntegration_GetAgentGrants(t *testing.T) {
	agentRepo := memorystorage.NewAgentRepository()
	providerService := newIntegrationProviderService(t)
	grantRepo := memorystorage.NewUserGrantRepository()

	ctx := context.Background()
	principalValue := "user@example.com"

	agent := &storage.Agent{
		ID:          "agent-456",
		ClientID:    "client-456",
		DisplayName: "Example Agent",
		Description: "An example agent",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	validUntil := time.Now().Add(30 * 24 * time.Hour)
	grant := &storage.UserGrant{
		ID:         "grant-001",
		Principal:  principalValue,
		AgentID:    "agent-456",
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"read:user", "repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := grantRepo.Create(ctx, grant); err != nil {
		t.Fatalf("failed to create grant: %v", err)
	}

	consentSvc := consentservice.NewService(agentRepo, providerService, grantRepo)
	handler := consent.NewAgentGrantsHandler(consentSvc, nil)

	reqCtx := principal.WithPrincipal(ctx, principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-456")
	reqCtx = context.WithValue(reqCtx, chi.RouteCtxKey, rctx)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-456/grants", nil)
	req = req.WithContext(reqCtx)

	rr := httptest.NewRecorder()
	handler.GetAgentGrants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var response consent.GetAgentGrantsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Data == nil {
		t.Fatal("expected a grant in response, got nil")
	}
	if response.Data.AgentID != "agent-456" {
		t.Errorf("expected agent ID %q, got %q", "agent-456", response.Data.AgentID)
	}
	if response.Data.Principal != principalValue {
		t.Errorf("expected principal %q, got %q", principalValue, response.Data.Principal)
	}
	if len(response.Data.DelegatedOAuth2Tokens) != 1 {
		t.Fatalf("expected 1 delegated token, got %d", len(response.Data.DelegatedOAuth2Tokens))
	}
	if response.Data.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID != "github" {
		t.Errorf("expected service ID %q, got %q", "github", response.Data.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
	}
}

// TestIntegration_AgentDetailFlow tests the complete consent detail + grants flow.
func TestIntegration_AgentDetailFlow(t *testing.T) {
	agentRepo := memorystorage.NewAgentRepository()
	providerService := newIntegrationProviderService(t)
	grantRepo := memorystorage.NewUserGrantRepository()

	ctx := context.Background()
	principalValue := "alice@example.com"

	govURL := "https://myagent.ai/governance"
	docsURL := "https://docs.myagent.ai"
	interfaceURL := "https://chat.myagent.ai"
	agent := &storage.Agent{
		ID:                   "agent-789",
		ClientID:             "client-789",
		DisplayName:          "MyAgent AI Assistant",
		Description:          "A helpful AI assistant that can access your data",
		GovernanceURL:        &govURL,
		UserDocumentationURL: &docsURL,
		AgentInterfaceURL:    &interfaceURL,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	services := []*model.ThirdpartyOAuth2ProviderEntity{
		{
			ID:          "github",
			DisplayName: "GitHub",
			ClientID:    "github-client",
			Secret:      model.NewPlaintextSecret("github-secret"),
			IssuerURI:   "https://github.com",
			Discovery:   model.DiscoveryConfig{EnableDiscovery: false},
			Endpoints: model.OAuth2Endpoints{
				TokenEndpoint:     "https://github.com/oauth/token",
				AuthorizeEndpoint: "https://github.com/oauth/authorize",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "read:user", Description: "Read user profile"},
				{ScopeValue: "repo", Description: "Full control of repositories"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "google",
			DisplayName: "Google",
			ClientID:    "google-client",
			Secret:      model.NewPlaintextSecret("google-secret"),
			IssuerURI:   "https://accounts.google.com",
			Discovery:   model.DiscoveryConfig{EnableDiscovery: false},
			Endpoints: model.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth2.googleapis.com/token",
				AuthorizeEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "email", Description: "View email address"},
				{ScopeValue: "profile", Description: "View basic profile info"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, svc := range services {
		if err := providerService.Create(ctx, svc); err != nil {
			t.Fatalf("failed to create service %s: %v", svc.ID, err)
		}
	}

	validUntil := time.Now().Add(30 * 24 * time.Hour)
	existingGrant := &storage.UserGrant{
		ID:         "grant-existing",
		Principal:  principalValue,
		AgentID:    "agent-789",
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"read:user"},
			},
		},
		CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		UpdatedAt: time.Now().Add(-7 * 24 * time.Hour),
	}
	if err := grantRepo.Create(ctx, existingGrant); err != nil {
		t.Fatalf("failed to create grant: %v", err)
	}

	consentSvc := consentservice.NewService(agentRepo, providerService, grantRepo)

	t.Run("GetAgentDetail", func(t *testing.T) {
		handler := consent.NewAgentDetailHandler(consentSvc, nil).
			WithAgentRepository(agentRepo)

		reqCtx := principal.WithPrincipal(context.Background(), principalValue)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-789")
		reqCtx = context.WithValue(reqCtx, chi.RouteCtxKey, rctx)

		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-789", nil)
		req = req.WithContext(reqCtx)

		rr := httptest.NewRecorder()
		handler.GetAgentDetail(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var response consent.GetAgentDetailResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Data.Agent.DisplayName != "MyAgent AI Assistant" {
			t.Errorf("unexpected display name: %s", response.Data.Agent.DisplayName)
		}
		if response.Data.Agent.AgentID != "agent-789" {
			t.Errorf("expected agent ID %q, got %q", "agent-789", response.Data.Agent.AgentID)
		}
		// Agent has no ServiceRequirements, so no services are returned.
		if len(response.Data.Services) != 0 {
			t.Errorf("expected 0 services, got %d", len(response.Data.Services))
		}
	})

	t.Run("GetAgentGrants", func(t *testing.T) {
		handler := consent.NewAgentGrantsHandler(consentSvc, nil)

		reqCtx := principal.WithPrincipal(context.Background(), principalValue)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-789")
		reqCtx = context.WithValue(reqCtx, chi.RouteCtxKey, rctx)

		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-789/grants", nil)
		req = req.WithContext(reqCtx)

		rr := httptest.NewRecorder()
		handler.GetAgentGrants(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", rr.Code, rr.Body.String())
		}

		var response consent.GetAgentGrantsResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Data == nil {
			t.Fatal("expected a grant, got nil")
		}
		if response.Data.AgentID != "agent-789" {
			t.Errorf("expected agent ID %q, got %q", "agent-789", response.Data.AgentID)
		}
		if len(response.Data.DelegatedOAuth2Tokens) != 1 {
			t.Fatalf("expected 1 delegated token, got %d", len(response.Data.DelegatedOAuth2Tokens))
		}
	})
}
