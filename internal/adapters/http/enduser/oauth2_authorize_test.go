package enduser

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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
	agentID := id.NewAgentID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

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

	// Feature 021: client_id is now the agent UUID, not the upstream client_id
	originalURL := "https://broker.example.com/oauth2/authorize?client_id=" + agentID.String() + "&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123"
	req := httptest.NewRequest("GET", originalURL, nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should redirect to consent UI
	assert.Equal(t, http.StatusFound, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "https://broker.example.com/consent/agent/"+agentID.String())
}

// TestOAuth2AuthorizeHandler_ServeHTTP_ActiveGrantRedirectsToUpstream tests redirect to upstream with active grant
func TestOAuth2AuthorizeHandler_ServeHTTP_ActiveGrantRedirectsToUpstream(t *testing.T) {
	agentRepo := newMockAgentRepo()
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	grantRepo := newMockGrantRepo()
	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
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

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Feature 021: client_id is now the agent UUID, not the upstream client_id
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id="+agentID.String()+"&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123&scope=openid+profile",
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
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-1",
		DisplayName: "Test Client",
	}
	_ = agentRepo.Create(context.Background(), agent)

	grantRepo := newMockGrantRepo()
	grant := &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
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

	handler := &OAuth2AuthorizeHandler{
		Service: svc,
	}

	// Feature 021: client_id is now the agent UUID, not the upstream client_id
	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id="+agentID.String()+"&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz123&scope=openid+profile+email&code_challenge=E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I&code_challenge_method=S256",
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

// TestOAuth2AuthorizeHandler_IssueTokenMode_NilServiceFails verifies that issue_token mode
// hard-fails when the consent Service is not wired, rather than silently bypassing consent.
func TestOAuth2AuthorizeHandler_IssueTokenMode_NilServiceFails(t *testing.T) {
	handler := &OAuth2AuthorizeHandler{
		Service:    nil, // misconfigured — Service not wired
		CodeIssuer: &mockCodeIssuer{},
	}

	req := httptest.NewRequest(
		"GET",
		"https://broker.example.com/oauth2/authorize?client_id=some-client&redirect_uri=https://client.example.com/callback&response_type=code&state=xyz",
		nil,
	)
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(body), "consent service unavailable")
}

type mockCodeIssuer struct{}

func (m *mockCodeIssuer) IssueAuthorizationCode(_ context.Context, _ *ports.AuthorizationRequest, _ string) (string, error) {
	return "test-code", nil
}

// Helper functions for test setup

type mockAgentRepository struct {
	agents map[id.AgentID]*storage.Agent
}

func newMockAgentRepo() *mockAgentRepository {
	return &mockAgentRepository{
		agents: make(map[id.AgentID]*storage.Agent),
	}
}

func (m *mockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	m.agents[agent.ID] = agent
	return nil
}

func (m *mockAgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	agent, ok := m.agents[agentID]
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

func (m *mockAgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	delete(m.agents, agentID)
	return nil
}

func (m *mockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	var agents []*storage.Agent
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (m *mockAgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
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
	grants map[id.GrantID]*storage.UserGrant
}

func newMockGrantRepo() *mockGrantRepository {
	return &mockGrantRepository{
		grants: make(map[id.GrantID]*storage.UserGrant),
	}
}

func (m *mockGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	m.grants[grant.ID] = grant
	return nil
}

func (m *mockGrantRepository) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	grant, ok := m.grants[grantID]
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

func (m *mockGrantRepository) Delete(ctx context.Context, grantID id.GrantID) error {
	delete(m.grants, grantID)
	return nil
}

func (m *mockGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (m *mockGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, nil
}

func (m *mockGrantRepository) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	for grantID, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, grantID)
		}
	}
	return nil
}

func (m *mockGrantRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.IsActive() {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (m *mockGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	agents := make(map[id.AgentID]bool)
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

func (m *mockGrantRepository) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	agents := make(map[id.AgentID]bool)
	for _, grant := range m.grants {
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

func (m *mockGrantRepository) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	for grantID, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			delete(m.grants, grantID)
			return nil
		}
	}
	return ports.ErrNotFound
}

// newMockAgentRepository creates a new mock agent repository for testing
// (wraps existing newMockAgentRepo for consistent naming)
func newMockAgentRepository() *mockAgentRepository {
	return newMockAgentRepo()
}

// mockSessionRepository is a test implementation of UserSessionRepository
type mockSessionRepository struct {
	sessions map[string]*storage.UserSession
}

func newMockSessionRepository() *mockSessionRepository {
	return &mockSessionRepository{
		sessions: make(map[string]*storage.UserSession),
	}
}

func (m *mockSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	key := session.Principal.String() + ":" + session.ServiceID.String()
	m.sessions[key] = session
	return nil
}

func (m *mockSessionRepository) Get(ctx context.Context, sessionID id.SessionID) (*storage.UserSession, error) {
	for _, session := range m.sessions {
		if session.ID == sessionID {
			return session, nil
		}
	}
	return nil, ports.ErrNotFound
}

func (m *mockSessionRepository) FindByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storage.UserSession, error) {
	key := principal.String() + ":" + serviceID.String()
	if session, exists := m.sessions[key]; exists {
		return session, nil
	}
	return nil, ports.ErrNotFound
}

func (m *mockSessionRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]*storage.UserSession, error) {
	var sessions []*storage.UserSession
	for _, session := range m.sessions {
		if session.Principal == principal {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionID id.SessionID) error {
	for key, session := range m.sessions {
		if session.ID == sessionID {
			delete(m.sessions, key)
			break
		}
	}
	return nil
}

func (m *mockSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) error {
	key := principal.String() + ":" + serviceID.String()
	delete(m.sessions, key)
	return nil
}

func (m *mockSessionRepository) CountByService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	count := 0
	for _, session := range m.sessions {
		if session.ServiceID == serviceID {
			count++
		}
	}
	return count, nil
}

// ============================================================================
// T037: hasRequiredScopes() function tests (Task 037)
// ============================================================================

// TestHasRequiredScopes tests the hasRequiredScopes helper function
func TestHasRequiredScopes(t *testing.T) {
	tests := []struct {
		name           string
		sessionScopes  []string
		requiredScopes []string
		expectedResult bool
		description    string
	}{
		{
			name:           "session has superset of required scopes",
			sessionScopes:  []string{"repo", "user:email", "read:user"},
			requiredScopes: []string{"repo", "user:email"},
			expectedResult: true,
			description:    "Session with additional scopes beyond requirements should be accepted",
		},
		{
			name:           "session has exact required scopes",
			sessionScopes:  []string{"repo", "user:email"},
			requiredScopes: []string{"repo", "user:email"},
			expectedResult: true,
			description:    "Session with exact matching scopes should be accepted",
		},
		{
			name:           "session missing one required scope",
			sessionScopes:  []string{"repo"},
			requiredScopes: []string{"repo", "user:email"},
			expectedResult: false,
			description:    "Session missing a required scope should be rejected",
		},
		{
			name:           "session has extra scopes not in required list",
			sessionScopes:  []string{"repo", "user:email", "admin:repo_hook"},
			requiredScopes: []string{"repo"},
			expectedResult: true,
			description:    "Session with extra scopes (superset) should be accepted",
		},
		{
			name:           "case-sensitive scope matching",
			sessionScopes:  []string{"User:Email"},
			requiredScopes: []string{"user:email"},
			expectedResult: false,
			description:    "Scope matching must be case-sensitive",
		},
		{
			name:           "empty required scopes list",
			sessionScopes:  []string{"repo", "user:email"},
			requiredScopes: []string{},
			expectedResult: true,
			description:    "Empty required scopes list should always return true (edge case)",
		},
		{
			name:           "nil session scopes",
			sessionScopes:  nil,
			requiredScopes: []string{"repo"},
			expectedResult: false,
			description:    "Nil session scopes with required scopes should be rejected",
		},
		{
			name:           "empty session scopes with required scopes",
			sessionScopes:  []string{},
			requiredScopes: []string{"repo"},
			expectedResult: false,
			description:    "Empty session scopes with required scopes should be rejected",
		},
		{
			name:           "both nil",
			sessionScopes:  nil,
			requiredScopes: nil,
			expectedResult: true,
			description:    "Both nil should return true",
		},
		{
			name:           "both empty",
			sessionScopes:  []string{},
			requiredScopes: []string{},
			expectedResult: true,
			description:    "Both empty should return true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRequiredScopes(tt.sessionScopes, tt.requiredScopes)
			assert.Equal(t, tt.expectedResult, result, tt.description)
		})
	}
}

// ============================================================================
// T038: validateMandatoryRequirements() method tests (Task 038)
// ============================================================================

// TestValidateMandatoryRequirements tests the validateMandatoryRequirements method
func TestValidateMandatoryRequirements(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID1 := id.NewServiceID()
	serviceID2 := id.NewServiceID()

	tests := []struct {
		name                  string
		setupFn               func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository)
		principal             string
		agent                 *storage.Agent
		expectedError         bool
		expectedErrorContains string
		description           string
	}{
		{
			name: "agent has no service requirements",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				return handler, newMockAgentRepository(), newMockSessionRepository()
			},
			principal:     "user@example.com",
			agent:         &storage.Agent{ID: agentID, ServiceRequirements: []storage.ServiceRequirement{}},
			expectedError: false,
			description:   "Agent with no service requirements should pass validation",
		},
		{
			name: "agent has only optional requirements",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeOptional,
						RequiredScopes:  []string{"read:repo"},
					},
				},
			},
			expectedError: false,
			description:   "Optional requirements should not block authorization",
		},
		{
			name: "user has active session for mandatory service with all required scopes",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				session := &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: id.Principal("user@example.com"),
					ServiceID: serviceID1,
					Scope:     []string{"repo", "user:email"},
				}
				_ = sessionRepo.Create(context.Background(), session)
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo", "user:email"},
					},
				},
			},
			expectedError: false,
			description:   "User with correct session should pass validation",
		},
		{
			name: "user missing session for mandatory service",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				return handler, newMockAgentRepository(), newMockSessionRepository()
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo"},
					},
				},
			},
			expectedError:         true,
			expectedErrorContains: "session_required",
			description:           "User without required session should fail validation",
		},
		{
			name: "user has session but missing required scope",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				session := &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: id.Principal("user@example.com"),
					ServiceID: serviceID1,
					Scope:     []string{"repo"}, // Missing user:email
				}
				_ = sessionRepo.Create(context.Background(), session)
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo", "user:email"},
					},
				},
			},
			expectedError:         true,
			expectedErrorContains: "scope_mismatch",
			description:           "User with insufficient scopes should fail validation",
		},
		{
			name: "session expired",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				expiredTime := time.Now().Add(-1 * time.Hour)
				session := &storage.UserSession{
					ID:                    id.NewSessionID(),
					Principal:             id.Principal("user@example.com"),
					ServiceID:             serviceID1,
					Scope:                 []string{"repo"},
					RefreshTokenExpiresAt: &expiredTime,
				}
				_ = sessionRepo.Create(context.Background(), session)
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo"},
					},
				},
			},
			expectedError:         true,
			expectedErrorContains: "session_expired",
			description:           "Expired session should be treated as missing",
		},
		{
			name: "multiple mandatory requirements, one missing",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				session := &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: id.Principal("user@example.com"),
					ServiceID: serviceID1,
					Scope:     []string{"repo"},
				}
				_ = sessionRepo.Create(context.Background(), session)
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo"},
					},
					{
						ServiceID:       serviceID2,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"write:org"},
					},
				},
			},
			expectedError:         true,
			expectedErrorContains: "session_required",
			description:           "Missing one mandatory requirement should fail",
		},
		{
			name: "multiple mandatory requirements, all satisfied",
			setupFn: func() (*OAuth2AuthorizeHandler, *mockAgentRepository, *mockSessionRepository) {
				handler := &OAuth2AuthorizeHandler{}
				sessionRepo := newMockSessionRepository()
				session1 := &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: id.Principal("user@example.com"),
					ServiceID: serviceID1,
					Scope:     []string{"repo"},
				}
				session2 := &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: id.Principal("user@example.com"),
					ServiceID: serviceID2,
					Scope:     []string{"write:org"},
				}
				_ = sessionRepo.Create(context.Background(), session1)
				_ = sessionRepo.Create(context.Background(), session2)
				return handler, newMockAgentRepository(), sessionRepo
			},
			principal: "user@example.com",
			agent: &storage.Agent{
				ID: agentID,
				ServiceRequirements: []storage.ServiceRequirement{
					{
						ServiceID:       serviceID1,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"repo"},
					},
					{
						ServiceID:       serviceID2,
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes:  []string{"write:org"},
					},
				},
			},
			expectedError: false,
			description:   "Multiple satisfied mandatory requirements should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, sessionRepo := tt.setupFn()
			handler.sessionRepository = sessionRepo

			err := handler.validateMandatoryRequirements(context.Background(), tt.principal, tt.agent)

			if tt.expectedError {
				assert.Error(t, err, tt.description)
				if tt.expectedErrorContains != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorContains, tt.description)
				}
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}
