//go:build integration
// +build integration

package consent_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encrypti
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	memorystorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/go-chi/chi/v5"
)

const integrationTestKEK = "ASNFZ4mrze/+3LqYdlQyEAEjRWeJq83v/ty6mHZUMhA="

func newIntegrationProviderService(t *testing.T) *thirdparty.ThirdpartyOAuth2ProviderService {
	t.Helper()
	enc, _, err := awsencryption.NewAWSEncryption(integrationTestKEK, "", 0)
	if err != nil {
		t.Fatalf("failed to create encryption adapter: %v", err)
	}
	repo := memorystorage.NewInMemoryThirdpartyOAuth2ProviderRepository()
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, enc, nil, slog.Default())
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

// TestIntegration_GetAgentDetail demonstrates the GET /api/consent/agent/:agentId endpoint.
func TestIntegration_GetAgentDetail(t *testing.T) {
	agentRepo := memorystorage.NewAgentRepository()
	providerService := newIntegrationProviderService(t)
	grantRepo := memorystorage.NewUserGrantRepository()

	ctx := context.Background()

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
	_ = agentRepo.Create(ctx, agent)

	if err := providerService.Create(ctx, newGitHubServiceEntity()); err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	consentSvc := consentservice.NewService(agentRepo, providerService, grantRepo)
	handler := consent.NewAgentDetailHandler(consentSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var response consent.GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Status: %d", rr.Code)
	t.Logf("Agent: %+v", response.Data.Agent)
	t.Logf("Services Count: %d", len(response.Data.Services))
}

// TestIntegration_GetAgentGrants demonstrates the GET /api/consent/agent/:agentId/grants endpoint.
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
	_ = agentRepo.Create(ctx, agent)

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
	_ = grantRepo.Create(ctx, grant)

	consentSvc := consentservice.NewService(agentRepo, providerService, grantRepo)
	handler := consent.NewAgentGrantsHandler(consentSvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-456/grants", nil)
	ctx = principal.WithPrincipal(ctx, principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-456")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetAgentGrants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var response consent.GetAgentGrantsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	t.Logf("Status: %d", rr.Code)
	t.Logf("Grants Count: %d", len(response.Data))
	if len(response.Data) > 0 {
		t.Logf("First Grant: %+v", response.Data[0])
	}
}

// TestIntegration_AgentDetailFlow tests the complete flow for User Story 2.
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
		handler := consent.NewAgentDetailHandler(consentSvc, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-789", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-789")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.GetAgentDetail(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response consent.GetAgentDetailResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response.Data.Agent.DisplayName != "MyAgent AI Assistant" {
			t.Errorf("unexpected display name: %s", response.Data.Agent.DisplayName)
		}
		if len(response.Data.Services) != 2 {
			t.Errorf("expected 2 services, got %d", len(response.Data.Services))
		}

		t.Logf("Agent Detail Response: %+v", response)
	})

	t.Run("GetAgentGrants", func(t *testing.T) {
		handler := consent.NewAgentGrantsHandler(consentSvc, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-789/grants", nil)
		grantCtx := principal.WithPrincipal(context.Background(), principalValue)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-789")
		grantCtx = context.WithValue(grantCtx, chi.RouteCtxKey, rctx)
		req = req.WithContext(grantCtx)

		rr := httptest.NewRecorder()
		handler.GetAgentGrants(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response consent.GetAgentGrantsResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(response.Data) != 1 {
			t.Errorf("expected 1 grant, got %d", len(response.Data))
		}
		if len(response.Data) > 0 {
			if response.Data[0].AgentID != "agent-789" {
				t.Errorf("unexpected agent ID: %s", response.Data[0].AgentID)
			}
			if len(response.Data[0].DelegatedOAuth2Tokens) != 1 {
				t.Errorf("expected 1 delegated token, got %d", len(response.Data[0].DelegatedOAuth2Tokens))
			}
		}

		t.Logf("Agent Grants Response: %+v", response)
	})
}
