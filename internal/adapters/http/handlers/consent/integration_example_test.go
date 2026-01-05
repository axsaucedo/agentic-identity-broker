//go:build integration
// +build integration

package consent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	memorystorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// TestIntegration_GetAgentDetail demonstrates the GET /api/consent/agent/:agentId endpoint.
func TestIntegration_GetAgentDetail(t *testing.T) {
	// Setup in-memory repositories
	agentRepo := memorystorage.NewAgentRepository()
	serviceRepo := memorystorage.NewThirdpartyServiceRepository()
	grantRepo := memorystorage.NewUserGrantRepository()

	// Create test data
	ctx := context.Background()

	// Create an agent
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

	// Create a third-party service
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "github",
		DisplayName:  "GitHub",
		ClientID:     "github-client-id",
		ClientSecret: "github-client-secret",
		IssuerURI:    "https://github.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/oauth/token",
			AuthorizeEndpoint: "https://github.com/oauth/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read:user", Description: "Read user profile"},
			{ScopeValue: "repo", Description: "Full control of repositories"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = serviceRepo.Create(ctx, service)

	// Create consent service and handler
	consentSvc := consentservice.NewService(agentRepo, serviceRepo, grantRepo)
	handler := consent.NewAgentDetailHandler(consentSvc, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute request
	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	// Verify response
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
	// Setup in-memory repositories
	agentRepo := memorystorage.NewAgentRepository()
	serviceRepo := memorystorage.NewThirdpartyServiceRepository()
	grantRepo := memorystorage.NewUserGrantRepository()

	// Create test data
	ctx := context.Background()
	principalValue := "user@example.com"

	// Create an agent
	agent := &storage.Agent{
		ID:          "agent-456",
		ClientID:    "client-456",
		DisplayName: "Example Agent",
		Description: "An example agent",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	_ = agentRepo.Create(ctx, agent)

	// Create a grant
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

	// Create consent service and handler
	consentSvc := consentservice.NewService(agentRepo, serviceRepo, grantRepo)
	handler := consent.NewAgentGrantsHandler(consentSvc, nil)

	// Create request with principal
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-456/grants", nil)
	ctx = principal.WithPrincipal(ctx, principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-456")
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Execute request
	rr := httptest.NewRecorder()
	handler.GetAgentGrants(rr, req)

	// Verify response
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
	// Setup in-memory repositories
	agentRepo := memorystorage.NewAgentRepository()
	serviceRepo := memorystorage.NewThirdpartyServiceRepository()
	grantRepo := memorystorage.NewUserGrantRepository()

	ctx := context.Background()
	principalValue := "alice@example.com"

	// Create test agent
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

	// Create third-party services
	services := []*storage.ThirdpartyOAuth2Service{
		{
			ID:           "github",
			DisplayName:  "GitHub",
			ClientID:     "github-client",
			ClientSecret: "github-secret",
			IssuerURI:    "https://github.com",
			Discovery:    storage.DiscoveryConfig{EnableDiscovery: false},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://github.com/oauth/token",
				AuthorizeEndpoint: "https://github.com/oauth/authorize",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "read:user", Description: "Read user profile"},
				{ScopeValue: "repo", Description: "Full control of repositories"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:           "google",
			DisplayName:  "Google",
			ClientID:     "google-client",
			ClientSecret: "google-secret",
			IssuerURI:    "https://accounts.google.com",
			Discovery:    storage.DiscoveryConfig{EnableDiscovery: false},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth2.googleapis.com/token",
				AuthorizeEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "email", Description: "View email address"},
				{ScopeValue: "profile", Description: "View basic profile info"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, svc := range services {
		if err := serviceRepo.Create(ctx, svc); err != nil {
			t.Fatalf("failed to create service %s: %v", svc.ID, err)
		}
	}

	// Create an existing grant (user has already granted access to GitHub)
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

	// Create consent service
	consentSvc := consentservice.NewService(agentRepo, serviceRepo, grantRepo)

	// Test 1: Get agent detail
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

		// Verify agent details
		if response.Data.Agent.DisplayName != "MyAgent AI Assistant" {
			t.Errorf("unexpected display name: %s", response.Data.Agent.DisplayName)
		}
		if len(response.Data.Services) != 2 {
			t.Errorf("expected 2 services, got %d", len(response.Data.Services))
		}

		t.Logf("Agent Detail Response: %+v", response)
	})

	// Test 2: Get user grants for this agent
	t.Run("GetAgentGrants", func(t *testing.T) {
		handler := consent.NewAgentGrantsHandler(consentSvc, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-789/grants", nil)
		ctx := principal.WithPrincipal(context.Background(), principalValue)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-789")
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.GetAgentGrants(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		var response consent.GetAgentGrantsResponse
		if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Verify existing grant
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
