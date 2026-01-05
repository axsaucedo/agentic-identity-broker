package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOAuth2SecurityE2E_UnauthorizedAccessBlocked tests that unauthorized access is blocked
// T043: Security validation - authorization endpoint requires principal
func TestOAuth2SecurityE2E_UnauthorizedAccessBlocked(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	// Create test agent
	agent := &storage.Agent{
		ID:       "agent-1",
		ClientID: "client-1",
		DisplayName: "Test Agent",
	}
	err := agentRepo.Create(context.Background(), agent)
	require.NoError(t, err)

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:             "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Test: Authorization endpoint requires principal in context
	// This test deliberately does NOT set up the principal in context
	// to verify the handler properly panics when principal is missing
	params := url.Values{
		"client_id":     []string{"agent-1"},
		"redirect_uri":  []string{"https://client.example.com/callback"},
		"response_type": []string{"code"},
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?"+params.Encode(),
		nil,
	)
	// Deliberately omit principal from context (would normally be set by RequirePrincipalMiddleware)
	w := httptest.NewRecorder()

	// This test now expects a panic since MustFromContext panics when principal is missing
	// In production, RequirePrincipalMiddleware prevents requests without principal
	assert.Panics(t, func() {
		handler.ServeHTTP(w, req)
	})
}

// TestOAuth2SecurityE2E_InvalidClientRejected tests invalid client IDs are rejected
// T043: Security validation - client validation
func TestOAuth2SecurityE2E_InvalidClientRejected(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:             "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Test: Invalid client_id should be rejected
	params := url.Values{
		"client_id":     []string{"nonexistent-agent"},
		"redirect_uri":  []string{"https://client.example.com/callback"},
		"response_type": []string{"code"},
		"state":         []string{"state-123"},
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?"+params.Encode(),
		nil,
	)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should redirect with error
	assert.Equal(t, http.StatusFound, w.Code)
	location := w.Header().Get("Location")
	redirectURL, err := url.Parse(location)
	require.NoError(t, err)

	// Verify error redirect (OAuth2 error response)
	query := redirectURL.Query()
	assert.Equal(t, "invalid_client", query.Get("error"))
	assert.Equal(t, "state-123", query.Get("state"))
}

// TestOAuth2SecurityE2E_MetadataPublic tests metadata endpoint is public
// T043: Security validation - metadata is publicly accessible
func TestOAuth2SecurityE2E_MetadataPublic(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			UpstreamTokenEndpoint:     "https://auth.example.com/token",
			PublicURL:             "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2MetadataHandler{
		Service: svc,
	}

	// Test: Metadata endpoint should be public (no auth required)
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/.well-known/oauth-authorization-server",
		nil,
	)
	// Deliberately omit authentication
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should succeed without authentication
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

// TestOAuth2SecurityE2E_TokenEndpointContentType tests Content-Type validation
// T043: Security validation - Content-Type enforcement
func TestOAuth2SecurityE2E_TokenEndpointContentType(t *testing.T) {
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "token123"}`))
	}))
	defer mockUpstream.Close()

	handler := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
	}

	// Test 1: Invalid Content-Type should be rejected
	req := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader(`{"grant_type":"authorization_code"}`),
	)
	req.Header.Set("Content-Type", "application/json") // Invalid - should be form-urlencoded
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should reject invalid content type
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 2: Valid Content-Type should be accepted
	reqValid := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader("grant_type=authorization_code&code=abc123"),
	)
	reqValid.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wValid := httptest.NewRecorder()

	handler.ServeHTTP(wValid, reqValid)

	// Should succeed
	assert.Equal(t, http.StatusOK, wValid.Code)
}

// TestOAuth2SecurityE2E_HopByHopHeaderFiltering tests hop-by-hop headers are filtered
// T044: Error handling and proxy security
func TestOAuth2SecurityE2E_HopByHopHeaderFiltering(t *testing.T) {
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify hop-by-hop headers were NOT forwarded
		if r.Header.Get("Connection") != "" || r.Header.Get("Transfer-Encoding") != "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "hop-by-hop headers detected"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "token123"}`))
	}))
	defer mockUpstream.Close()

	handler := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
	}

	// Test: Hop-by-hop headers should be filtered
	req := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader("grant_type=authorization_code&code=abc123"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add hop-by-hop headers that should be filtered
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Transfer-Encoding", "chunked")
	req.Header.Set("Keep-Alive", "timeout=5")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Should succeed (headers were filtered)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestOAuth2SecurityE2E_UpstreamErrorsProxied tests upstream errors are safely proxied
// T044: Error handling - upstream error propagation
func TestOAuth2SecurityE2E_UpstreamErrorsProxied(t *testing.T) {
	// Test 1: 401 Unauthorized from upstream
	mockUpstream401 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid_client"}`))
	}))
	defer mockUpstream401.Close()

	handler401 := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream401.URL,
	}

	req401 := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader("grant_type=authorization_code&code=test"),
	)
	req401.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w401 := httptest.NewRecorder()

	handler401.ServeHTTP(w401, req401)

	// Should proxy the 401 error
	assert.Equal(t, http.StatusUnauthorized, w401.Code)

	// Test 2: 500 Internal Server Error from upstream
	mockUpstream500 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server_error"}`))
	}))
	defer mockUpstream500.Close()

	handler500 := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream500.URL,
	}

	req500 := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader("grant_type=authorization_code&code=test"),
	)
	req500.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w500 := httptest.NewRecorder()

	handler500.ServeHTTP(w500, req500)

	// Should proxy the 500 error
	assert.Equal(t, http.StatusInternalServerError, w500.Code)
}

// TestOAuth2SecurityE2E_MetadataCacheable tests metadata response headers
// T044: Error handling - cache control headers
func TestOAuth2SecurityE2E_MetadataCacheable(t *testing.T) {
	agentRepo := newInMemoryAgentRepo()
	grantRepo := newInMemoryGrantRepo()

	svc := oauth2.NewService(
		agentRepo,
		grantRepo,
		&oauth2.OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			UpstreamTokenEndpoint:     "https://auth.example.com/token",
			PublicURL:             "https://broker.example.com",
		},
	)

	handler := &enduser.OAuth2MetadataHandler{
		Service: svc,
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/.well-known/oauth-authorization-server",
		nil,
	)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify cache control headers for metadata
	assert.Equal(t, http.StatusOK, w.Code)
	cacheControl := w.Header().Get("Cache-Control")
	assert.NotEmpty(t, cacheControl)
	assert.Contains(t, cacheControl, "max-age")
}

// TestOAuth2SecurityE2E_ResponseStreamingPreservesHeaders tests response headers are preserved
// T044: Error handling - response header preservation
func TestOAuth2SecurityE2E_ResponseStreamingPreservesHeaders(t *testing.T) {
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"access_token": "token", "token_type": "Bearer"}`))
	}))
	defer mockUpstream.Close()

	handler := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: mockUpstream.URL,
	}

	req := httptest.NewRequest(
		"POST",
		"https://broker.example.com/oauth2/token",
		strings.NewReader("grant_type=authorization_code&code=abc123"),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Verify standard headers are preserved
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
}

// Test helpers

func newInMemoryAgentRepo() ports.AgentRepository {
	return &inMemoryAgentRepo{agents: make(map[string]*storage.Agent)}
}

func newInMemoryGrantRepo() ports.UserGrantRepository {
	return &inMemoryGrantRepo{grants: make(map[string]*storage.UserGrant)}
}

type inMemoryAgentRepo struct {
	agents map[string]*storage.Agent
}

func (r *inMemoryAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	r.agents[agent.ID] = agent
	return nil
}

func (r *inMemoryAgentRepo) Get(ctx context.Context, id string) (*storage.Agent, error) {
	agent, ok := r.agents[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return agent, nil
}

func (r *inMemoryAgentRepo) GetByID(ctx context.Context, id string) (*storage.Agent, error) {
	agent, ok := r.agents[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return agent, nil
}

func (r *inMemoryAgentRepo) List(ctx context.Context) ([]*storage.Agent, error) {
	agents := make([]*storage.Agent, 0, len(r.agents))
	for _, a := range r.agents {
		agents = append(agents, a)
	}
	return agents, nil
}

func (r *inMemoryAgentRepo) GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error) {
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

func (r *inMemoryAgentRepo) Update(ctx context.Context, agent *storage.Agent) error {
	r.agents[agent.ID] = agent
	return nil
}

func (r *inMemoryAgentRepo) Delete(ctx context.Context, id string) error {
	delete(r.agents, id)
	return nil
}

type inMemoryGrantRepo struct {
	grants map[string]*storage.UserGrant
}

func (r *inMemoryGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	r.grants[grant.ID] = grant
	return nil
}

func (r *inMemoryGrantRepo) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	grant, ok := r.grants[id]
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

func (r *inMemoryGrantRepo) Delete(ctx context.Context, id string) error {
	delete(r.grants, id)
	return nil
}

func (r *inMemoryGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (r *inMemoryGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal, agentID string) (*storage.UserGrant, error) {
	for _, grant := range r.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, ports.ErrNotFound
}

// Stub methods for interface compliance
func (r *inMemoryGrantRepo) GetExpired(ctx context.Context, before time.Time) ([]*storage.UserGrant, error) {
	return []*storage.UserGrant{}, nil
}

func (r *inMemoryGrantRepo) GetByServiceAndPrincipal(ctx context.Context, serviceID, principal string) ([]*storage.UserGrant, error) {
	return []*storage.UserGrant{}, nil
}

func (r *inMemoryGrantRepo) DeleteByAgent(ctx context.Context, agentID string) error {
	return nil
}

func (r *inMemoryGrantRepo) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range r.grants {
		if grant.Principal == principal {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (r *inMemoryGrantRepo) CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error) {
	agents := make(map[string]bool)
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

func (r *inMemoryGrantRepo) ListByServiceID(ctx context.Context, serviceID string) ([]string, error) {
	agents := make(map[string]bool)
	for _, grant := range r.grants {
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
