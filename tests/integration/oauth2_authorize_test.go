package integration

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// TestOAuth2AuthorizeEndpoint_InvalidClientError tests complete flow with invalid client
func TestOAuth2AuthorizeEndpoint_InvalidClientError(t *testing.T) {
	// Setup in-memory repositories
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	// Create service
	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
			SupportedResponseTypes:    []string{"code"},
			SupportedGrantTypes:       []string{"authorization_code"},
		},
	)

	// Create handler
	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Make request with unknown client
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=unknown&redirect_uri=https://client.example.com/callback&response_type=code&state=abc123",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL, err := url.Parse(w.Header().Get("Location"))
	require.NoError(t, err)

	// Verify error redirect
	query := redirectURL.Query()
	assert.Equal(t, "invalid_client", query.Get("error"))
	assert.Equal(t, "abc123", query.Get("state"))
	assert.NotEmpty(t, query.Get("error_description"))
}

// TestOAuth2AuthorizeEndpoint_MissingParameterError tests validation of required parameters
func TestOAuth2AuthorizeEndpoint_MissingParameterError(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	tests := []struct {
		name        string
		queryString string
	}{
		{"missing client_id", "?redirect_uri=https://client.example.com/callback&response_type=code"},
		{"missing redirect_uri", "?client_id=client-1&response_type=code"},
		{"missing response_type", "?client_id=client-1&redirect_uri=https://client.example.com/callback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				"GET",
				"https://broker.example.com/oauth2/authorize"+tt.queryString,
				nil,
			)
			req.Header.Set("X-Remote-User", "user@example.com")
			ctx := principal.WithPrincipal(req.Context(), "user@example.com")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			body, _ := io.ReadAll(w.Body)
			assert.True(t, len(body) > 0, "error message should be present")
		})
	}
}

// TestOAuth2AuthorizeEndpoint_NoGrantRedirectsToConsent tests redirect to consent UI
func TestOAuth2AuthorizeEndpoint_NoGrantRedirectsToConsent(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	// Register agent
	agent := &storage.Agent{
		ID:          id.NewAgentID(),
		ClientID:    id.NewClientID("client-1"),
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	originalURL := "https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123"
	req := httptest.NewRequest("GET", originalURL, nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify redirect to consent
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://broker.example.com/consent/agent/"+agent.ID.String())
}

// TestOAuth2AuthorizeEndpoint_ActiveGrantRedirectsToUpstream tests redirect to upstream with active grant
func TestOAuth2AuthorizeEndpoint_ActiveGrantRedirectsToUpstream(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	// Register agent
	agentID := id.NewAgentID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("client-1"),
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	// Create active grant for user
	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil, // Indefinite grant
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: id.NewServiceID(), Scopes: []string{"openid"}},
		},
	}
	_ = grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
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

	// Verify redirect to upstream
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://auth.example.com/authorize")

	// Verify parameters preserved
	parsedURL, _ := url.Parse(redirectURL)
	query := parsedURL.Query()
	assert.Equal(t, "client-1", query.Get("client_id"))
	assert.Equal(t, "https://client.example.com/callback", query.Get("redirect_uri"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "xyz123", query.Get("state"))
}

// TestOAuth2AuthorizeEndpoint_ExpiredGrantRedirectsToConsent tests redirect to consent with expired grant
func TestOAuth2AuthorizeEndpoint_ExpiredGrantRedirectsToConsent(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	// Register agent
	agentID := id.NewAgentID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("client-1"),
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	// Create expired grant
	expiredTime := time.Now().Add(-1 * time.Hour)
	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: &expiredTime,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: id.NewServiceID(), Scopes: []string{"openid"}},
		},
	}
	_ = grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify redirect to consent (not upstream)
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://broker.example.com/consent/agent/"+agent.ID.String())
	assert.NotContains(t, redirectURL, "https://auth.example.com/authorize")
}

// TestOAuth2AuthorizeEndpoint_WithMiddleware tests complete flow with audit middleware
func TestOAuth2AuthorizeEndpoint_WithMiddleware(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	agentID := id.NewAgentID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("client-1"),
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: id.NewServiceID(), Scopes: []string{"openid"}},
		},
	}
	_ = grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Wrap with audit middleware
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	auditMiddleware := middleware.OAuth2AuditMiddleware(logger)
	wrappedHandler := auditMiddleware(handler)

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	req.Header.Set("X-Request-ID", "req-123")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(w, req)

	// Verify successful redirect
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "https://auth.example.com/authorize")
}

// TestOAuth2AuthorizeEndpoint_PKCEParametersPreserved tests PKCE parameters preserved
func TestOAuth2AuthorizeEndpoint_PKCEParametersPreserved(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	agentID := id.NewAgentID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("client-1"),
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: id.NewServiceID(), Scopes: []string{"openid"}},
		},
	}
	_ = grantRepo.Create(context.Background(), grant)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=client-1&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123&code_challenge=E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I&code_challenge_method=S256",
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify PKCE parameters preserved in upstream URL
	redirectURL := w.Header().Get("Location")
	parsedURL, _ := url.Parse(redirectURL)
	query := parsedURL.Query()

	assert.Equal(t, "E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I", query.Get("code_challenge"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
}

// In-memory repository implementations for testing

type inMemoryAgentRepo struct {
	agents map[id.AgentID]*storage.Agent
}

func newInMemoryAgentRepo() *inMemoryAgentRepo {
	return &inMemoryAgentRepo{
		agents: make(map[id.AgentID]*storage.Agent),
	}
}

func (r *inMemoryAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	r.agents[agent.ID] = agent
	return nil
}

func (r *inMemoryAgentRepo) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	agent, ok := r.agents[agentID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return agent, nil
}

func (r *inMemoryAgentRepo) Update(ctx context.Context, agent *storage.Agent) error {
	if _, ok := r.agents[agent.ID]; !ok {
		return ports.ErrNotFound
	}
	r.agents[agent.ID] = agent
	return nil
}

func (r *inMemoryAgentRepo) Delete(ctx context.Context, agentID id.AgentID) error {
	delete(r.agents, agentID)
	return nil
}

func (r *inMemoryAgentRepo) List(ctx context.Context) ([]*storage.Agent, error) {
	var agents []*storage.Agent
	for _, agent := range r.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (r *inMemoryAgentRepo) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	for _, agent := range r.agents {
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

type inMemoryGrantRepo struct {
	grants map[id.GrantID]*storage.UserGrant
}

func newInMemoryGrantRepo() *inMemoryGrantRepo {
	return &inMemoryGrantRepo{
		grants: make(map[id.GrantID]*storage.UserGrant),
	}
}

func (r *inMemoryGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	r.grants[grant.ID] = grant
	return nil
}

func (r *inMemoryGrantRepo) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	grant, ok := r.grants[grantID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return grant, nil
}

func (r *inMemoryGrantRepo) Update(ctx context.Context, grant *storage.UserGrant) error {
	if _, ok := r.grants[grant.ID]; !ok {
		return ports.ErrNotFound
	}
	r.grants[grant.ID] = grant
	return nil
}

func (r *inMemoryGrantRepo) Delete(ctx context.Context, grantID id.GrantID) error {
	delete(r.grants, grantID)
	return nil
}

func (r *inMemoryGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (r *inMemoryGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, nil
}

func (r *inMemoryGrantRepo) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	for grantID, grant := range r.grants {
		if grant.AgentID == agentID {
			delete(r.grants, grantID)
		}
	}
	return nil
}

func (r *inMemoryGrantRepo) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.IsActive() {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (r *inMemoryGrantRepo) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	agents := make(map[id.AgentID]bool)
	for _, grant := range r.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				agents[grant.AgentID] = true
				break
			}
		}
	}
	return len(agents), nil
}

func (r *inMemoryGrantRepo) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	agents := make(map[id.AgentID]bool)
	for _, grant := range r.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				agents[grant.AgentID] = true
				break
			}
		}
	}
	var agentIDs []id.AgentID
	for agentID := range agents {
		agentIDs = append(agentIDs, agentID)
	}
	return agentIDs, nil
}
