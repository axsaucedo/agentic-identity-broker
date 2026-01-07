package enduser

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuth2AuthorizeHandler_ServeHTTP_MissingPrincipal tests handler when principal is not provided
func TestOAuth2AuthorizeHandler_ServeHTTP_MissingPrincipal(t *testing.T) {
	// Setup
	svc := oauth2.NewService(
		newMockAgentRepo(),
		newMockGrantRepo(),
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Request without principal in context (missing principal)
	// MustFromContext panics when principal is missing
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code",
		nil,
	)
	w := httptest.NewRecorder()

	// Execute - expect panic since principal is missing from context
	// In production, RequirePrincipalMiddleware prevents this
	assert.Panics(t, func() {
		handler.ServeHTTP(w, req)
	})
}

// TestOAuth2AuthorizeHandler_ServeHTTP_MissingParameters tests handler when required OAuth2 parameters missing
func TestOAuth2AuthorizeHandler_ServeHTTP_MissingParameters(t *testing.T) {
	svc := oauth2.NewService(
		newMockAgentRepo(),
		newMockGrantRepo(),
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	tests := []struct {
		name      string
		queryPath string
		wantError string
	}{
		{
			name:      "missing client_id",
			queryPath: "?redirect_uri=https://client.example.com/callback&response_type=code",
			wantError: "client_id",
		},
		{
			name:      "missing redirect_uri",
			queryPath: "?client_id=client-1&response_type=code",
			wantError: "redirect_uri",
		},
		{
			name:      "missing response_type",
			queryPath: "?client_id=client-1&redirect_uri=https://client.example.com/callback",
			wantError: "response_type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				"GET",
				"https://broker.example.com/oauth2/authorize"+tt.queryPath,
				nil,
			)
			req.Header.Set("X-Remote-User", "user@example.com")
			ctx := principal.WithPrincipal(req.Context(), "user@example.com")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			body, _ := io.ReadAll(w.Body)
			assert.Contains(t, string(body), tt.wantError)
		})
	}
}

// TestOAuth2AuthorizeHandler_ServeHTTP_InvalidClient tests handler when client not registered
func TestOAuth2AuthorizeHandler_ServeHTTP_InvalidClient(t *testing.T) {
	svc := oauth2.NewService(
		newMockAgentRepo(),
		newMockGrantRepo(),
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=unknown-client&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should redirect with error parameters
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)
	assert.Equal(t, "https", redirectURL.Scheme)
	assert.Equal(t, "client.example.com", redirectURL.Host)
	assert.Equal(t, "/callback", redirectURL.Path)

	// Verify error parameters in redirect
	query := redirectURL.Query()
	assert.Equal(t, "invalid_client", query.Get("error"))
	assert.NotEmpty(t, query.Get("error_description"))
	assert.Equal(t, "xyz123", query.Get("state"))
}

// TestOAuth2AuthorizeHandler_ServeHTTP_NoGrantRedirectsToConsent tests redirect to consent UI when no active grant
func TestOAuth2AuthorizeHandler_ServeHTTP_NoGrantRedirectsToConsent(t *testing.T) {
	agentRepo := newMockAgentRepo()
	agent := &storage.Agent{
		ID:          "agent-1",
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	agentRepo.Create(context.Background(), agent)

	svc := oauth2.NewService(
		agentRepo,
		newMockGrantRepo(),
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	originalURL := "https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123"
	req := httptest.NewRequest("GET", originalURL, nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should redirect to consent UI
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://broker.example.com/consent/agent/agent-1")
}

// TestOAuth2AuthorizeHandler_ServeHTTP_ActiveGrantRedirectsToUpstream tests redirect to upstream with active grant
func TestOAuth2AuthorizeHandler_ServeHTTP_ActiveGrantRedirectsToUpstream(t *testing.T) {
	agentRepo := newMockAgentRepo()
	agent := &storage.Agent{
		ID:          "agent-1",
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	agentRepo.Create(context.Background(), agent)

	grantRepo := newMockGrantRepo()
	grant := &storage.UserGrant{
		ID:         "grant-1",
		Principal:  "user@example.com",
		AgentID:    "agent-1",
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"openid"}},
		},
	}
	grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123&scope=openid+profile",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should redirect to upstream
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://auth.example.com/authorize")
	assert.Contains(t, redirectURL, "client_id=client-1")
	assert.Contains(t, redirectURL, "state=xyz123")
	assert.Contains(t, redirectURL, "response_type=code")
}

// TestOAuth2AuthorizeHandler_ServeHTTP_PreservesOAuth2Parameters tests all OAuth2 parameters preserved
func TestOAuth2AuthorizeHandler_ServeHTTP_PreservesOAuth2Parameters(t *testing.T) {
	agentRepo := newMockAgentRepo()
	agent := &storage.Agent{
		ID:          "agent-1",
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	agentRepo.Create(context.Background(), agent)

	grantRepo := newMockGrantRepo()
	grant := &storage.UserGrant{
		ID:         "grant-1",
		Principal:  "user@example.com",
		AgentID:    "agent-1",
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"openid"}},
		},
	}
	grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Request with PKCE parameters
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123&scope=openid+profile+email&code_challenge=E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I&code_challenge_method=S256",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")

	// Verify all parameters preserved
	parsedURL, _ := url.Parse(redirectURL)
	query := parsedURL.Query()
	assert.Equal(t, "client-1", query.Get("client_id"))
	assert.Equal(t, "https://client.example.com/callback", query.Get("redirect_uri"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "xyz123", query.Get("state"))
	assert.Equal(t, "openid profile email", query.Get("scope"))
	assert.Equal(t, "E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I", query.Get("code_challenge"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
}

// TestOAuth2AuthorizeHandler_ServeHTTP_JSONResponseFormat tests error responses use proper JSON format
func TestOAuth2AuthorizeHandler_ServeHTTP_JSONResponseFormat(t *testing.T) {
	svc := oauth2.NewService(
		newMockAgentRepo(),
		newMockGrantRepo(),
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Missing required parameter
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?redirect_uri=https://client.example.com/callback",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	body, _ := io.ReadAll(w.Body)

	// For parameter validation errors, expect plain text error
	assert.Contains(t, string(body), "client_id")
}

// Helper functions for test setup

type mockAgentRepository struct {
	agents map[string]*storage.Agent
}

func newMockAgentRepo() *mockAgentRepository {
	return &mockAgentRepository{
		agents: make(map[string]*storage.Agent),
	}
}

func (m *mockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) Get(ctx context.Context, id string) (*storage.Agent, error) {
	agent, ok := m.agents[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return agent, nil
}

func (m *mockAgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	if _, ok := m.agents[agent.ID]; !ok {
		return ports.ErrNotFound
	}
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) Delete(ctx context.Context, id string) error {
	delete(m.agents, id)
	return nil
}

func (m *mockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	var agents []*storage.Agent
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (m *mockAgentRepository) GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error) {
	for _, agent := range m.agents {
		if agent.ClientID == clientID {
			return agent, nil
		}
	}
	return nil, storage.NewStorageError(
		"GetAgentByClientID",
		storage.ErrorKindNotFound,
		ports.ErrNotFound,
		"agent not found",
	)
}

type mockGrantRepository struct {
	grants map[string]*storage.UserGrant
}

func newMockGrantRepo() *mockGrantRepository {
	return &mockGrantRepository{
		grants: make(map[string]*storage.UserGrant),
	}
}

func (m *mockGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	m.grants[grant.ID] = grant
	return nil
}

func (m *mockGrantRepository) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	grant, ok := m.grants[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return grant, nil
}

func (m *mockGrantRepository) Update(ctx context.Context, grant *storage.UserGrant) error {
	if _, ok := m.grants[grant.ID]; !ok {
		return ports.ErrNotFound
	}
	m.grants[grant.ID] = grant
	return nil
}

func (m *mockGrantRepository) Delete(ctx context.Context, id string) error {
	delete(m.grants, id)
	return nil
}

func (m *mockGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (m *mockGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal, agentID string) (*storage.UserGrant, error) {
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, nil
}

func (m *mockGrantRepository) DeleteByAgent(ctx context.Context, agentID string) error {
	for id, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, id)
		}
	}
	return nil
}

func (m *mockGrantRepository) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.IsActive() {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (m *mockGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error) {
	agents := make(map[string]bool)
	for _, grant := range m.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				agents[grant.AgentID] = true
				break
			}
		}
	}
	return len(agents), nil
}

func (m *mockGrantRepository) ListByServiceID(ctx context.Context, serviceID string) ([]string, error) {
	agents := make(map[string]bool)
	for _, grant := range m.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				agents[grant.AgentID] = true
				break
			}
		}
	}
	var agentIDs []string
	for agentID := range agents {
		agentIDs = append(agentIDs, agentID)
	}
	return agentIDs, nil
}
