package oauth2

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockAgentRepository struct {
	agents map[id.AgentID]*storage.Agent
}

func NewMockAgentRepository() *MockAgentRepository {
	return &MockAgentRepository{
		agents: make(map[id.AgentID]*storage.Agent),
	}
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	m.agents[agent.ID] = agent
	return nil
}

func (m *MockAgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	agent, ok := m.agents[agentID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return agent, nil
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	if _, ok := m.agents[agent.ID]; !ok {
		return ports.ErrNotFound
	}
	m.agents[agent.ID] = agent
	return nil
}

func (m *MockAgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	delete(m.agents, agentID)
	return nil
}

func (m *MockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	var agents []*storage.Agent
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (m *MockAgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	for _, agent := range m.agents {
		if agent.ClientID != nil && *agent.ClientID == clientID {
			return agent, nil
		}
	}
	return nil, ports.ErrNotFound
}

func (m *MockAgentRepository) GetByClientURI(ctx context.Context, uri string) (*storage.Agent, error) {
	return nil, storage.NewStorageError("GetAgentByClientURI", storage.ErrorKindNotFound, ports.ErrNotFound, "not found")
}

func (m *MockAgentRepository) ExistsOtherWithClientID(_ context.Context, _ id.ClientID, _ *id.AgentID) (bool, error) {
	return false, nil
}

// MockGrantRepository is a test double for UserGrantRepository
type MockGrantRepository struct {
	grants map[id.GrantID]*storage.UserGrant
}

func NewMockGrantRepository() *MockGrantRepository {
	return &MockGrantRepository{
		grants: make(map[id.GrantID]*storage.UserGrant),
	}
}

func (m *MockGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	m.grants[grant.ID] = grant
	return nil
}

func (m *MockGrantRepository) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	grant, ok := m.grants[grantID]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return grant, nil
}

func (m *MockGrantRepository) Update(ctx context.Context, grant *storage.UserGrant) error {
	if _, ok := m.grants[grant.ID]; !ok {
		return ports.ErrNotFound
	}
	m.grants[grant.ID] = grant
	return nil
}

func (m *MockGrantRepository) Delete(ctx context.Context, grantID id.GrantID) error {
	delete(m.grants, grantID)
	return nil
}

func (m *MockGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (m *MockGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, nil
}

func (m *MockGrantRepository) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	for grantKey, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, grantKey)
		}
	}
	return nil
}

func (m *MockGrantRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.IsActive() {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (m *MockGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
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

func (m *MockGrantRepository) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
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

func (m *MockGrantRepository) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	for grantKey, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			delete(m.grants, grantKey)
			return nil
		}
	}
	return ports.ErrNotFound
}

// MockSessionRepository is a test double for UserSessionRepository.
// Uses a configurable findFunc so each test case can define its own behaviour for
// FindByPrincipalAndService — the only method exercised by the authorization flow.
// All other methods are intentionally no-op stubs.
type MockSessionRepository struct {
	findFunc func(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storage.UserSession, error)
}

func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{}
}

func (m *MockSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	return nil
}

// Get is not exercised by the authorization flow tests but must satisfy the interface.
func (m *MockSessionRepository) Get(ctx context.Context, sessionID id.SessionID) (*storage.UserSession, error) {
	return nil, &storage.StorageError{Kind: storage.ErrorKindNotFound}
}

func (m *MockSessionRepository) FindByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) (*storage.UserSession, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, principal, serviceID)
	}
	return nil, nil
}

// ListByPrincipal, Delete, DeleteByPrincipalAndService, and CountByService are not
// exercised by the authorization flow; they are intentionally no-op stubs.
func (m *MockSessionRepository) ListByPrincipal(ctx context.Context, principal id.Principal) ([]*storage.UserSession, error) {
	return nil, nil
}

func (m *MockSessionRepository) Delete(ctx context.Context, sessionID id.SessionID) error {
	return nil
}

func (m *MockSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal id.Principal, serviceID id.ServiceID) error {
	return nil
}

func (m *MockSessionRepository) CountByService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

// TestService_HandleAuthorization tests the HandleAuthorization method with table-driven tests
func TestService_HandleAuthorization(t *testing.T) {
	testAgentID := id.NewAgentID()
	testServiceID := id.NewServiceID()

	tests := []struct {
		name       string
		setupAgent func(*MockAgentRepository)
		setupGrant func(*MockGrantRepository)
		authReq    *ports.AuthorizationRequest
		principal  id.Principal
		wantAction string
	}{
		{
			name:       "unregistered agent UUID returns error redirect",
			setupAgent: func(r *MockAgentRepository) {},
			setupGrant: func(r *MockGrantRepository) {},
			authReq: &ports.AuthorizationRequest{
				ClientID:     id.ClientID(id.NewAgentID().String()), // valid UUID but not in repo
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
			},
			principal:  id.NewPrincipal("user@example.com"),
			wantAction: "error",
		},
		{
			name: "valid UUID client_id with no grant redirects to consent UI",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:           testAgentID,
					ClientID:     ptr.To(id.ClientID("client-1")),
					DisplayName:  "Test Client",
					RedirectURIs: []string{"https://client.example.com/callback"},
				}
				_ = r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {},
			authReq: &ports.AuthorizationRequest{
				ClientID:     id.ClientID(testAgentID.String()),
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + testAgentID.String(),
			},
			principal:  id.NewPrincipal("user@example.com"),
			wantAction: "redirect_to_consent",
		},
		{
			name: "valid UUID client_id with active grant redirects to upstream",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:           testAgentID,
					ClientID:     ptr.To(id.ClientID("client-1")),
					DisplayName:  "Test Client",
					RedirectURIs: []string{"https://client.example.com/callback"},
				}
				_ = r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {
				grant := &storage.UserGrant{
					ID:         id.NewGrantID(),
					Principal:  id.Principal("user@example.com"),
					AgentID:    testAgentID,
					ValidUntil: nil,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: testServiceID, Scopes: []string{"openid"}},
					},
				}
				_ = r.Create(context.Background(), grant)
			},
			authReq: &ports.AuthorizationRequest{
				ClientID:     id.ClientID(testAgentID.String()),
				RedirectURI:  "https://client.example.com/callback",
				Scope:        "openid profile",
				State:        "xyz123",
				ResponseType: "code",
			},
			principal:  id.NewPrincipal("user@example.com"),
			wantAction: "proceed",
		},
		{
			name: "expired grant redirects to consent UI",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:           testAgentID,
					ClientID:     ptr.To(id.ClientID("client-1")),
					DisplayName:  "Test Client",
					RedirectURIs: []string{"https://client.example.com/callback"},
				}
				_ = r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {
				expiredTime := time.Now().Add(-1 * time.Hour)
				grant := &storage.UserGrant{
					ID:         id.NewGrantID(),
					Principal:  id.Principal("user@example.com"),
					AgentID:    testAgentID,
					ValidUntil: &expiredTime,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: testServiceID, Scopes: []string{"openid"}},
					},
				}
				_ = r.Create(context.Background(), grant)
			},
			authReq: &ports.AuthorizationRequest{
				ClientID:     id.ClientID(testAgentID.String()),
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + testAgentID.String(),
			},
			principal:  id.NewPrincipal("user@example.com"),
			wantAction: "redirect_to_consent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()

			tt.setupAgent(agentRepo)
			tt.setupGrant(grantRepo)

			svc := NewService(agentRepo, grantRepo, &OAuth2Config{
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				PublicURL:                 "https://broker.example.com",
				SupportedResponseTypes:    []string{"code"},
				SupportedGrantTypes:       []string{"authorization_code"},
			})

			ctx := context.Background()
			decision, err := svc.HandleAuthorization(ctx, tt.authReq, tt.principal)

			require.NoError(t, err, "HandleAuthorization should not error")
			require.NotNil(t, decision, "Decision should not be nil")

			assert.Equal(t, tt.wantAction, decision.Action, "Action mismatch")
		})
	}
}

// TestService_HandleAuthorization_SessionExpiry tests that expired delegated sessions
// redirect back to consent even when the grant itself is still active.
func TestService_HandleAuthorization_SessionExpiry(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	activeGrant := func(r *MockGrantRepository) {
		grant := &storage.UserGrant{
			ID:        id.NewGrantID(),
			Principal: id.Principal("user@example.com"),
			AgentID:   agentID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
			},
		}
		_ = r.Create(context.Background(), grant)
	}

	cfg := &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		PublicURL:                 "https://broker.example.com",
	}

	authReq := &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + agentID.String(),
	}

	tests := []struct {
		name         string
		setupSession func(*MockSessionRepository)
		wantAction   string
	}{
		{
			name: "active grant with valid session redirects to upstream",
			setupSession: func(r *MockSessionRepository) {
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return &storage.UserSession{
						ID:        id.NewSessionID(),
						Principal: id.Principal("user@example.com"),
						ServiceID: serviceID,
						TokenType: "Bearer",
					}, nil
				}
			},
			wantAction: "proceed",
		},
		{
			name: "active grant with expired session redirects to consent",
			setupSession: func(r *MockSessionRepository) {
				expiredAt := time.Now().Add(-1 * time.Hour)
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return &storage.UserSession{
						ID:                    id.NewSessionID(),
						Principal:             id.Principal("user@example.com"),
						ServiceID:             serviceID,
						TokenType:             "Bearer",
						RefreshTokenExpiresAt: &expiredAt,
					}, nil
				}
			},
			wantAction: "redirect_to_consent",
		},
		{
			name: "active grant with no session yet still redirects to upstream",
			setupSession: func(r *MockSessionRepository) {
				// findFunc returns nil, nil — no session established yet.
				// Absence of a session is not an expiry; mandatory-requirements
				// validation (Step 5) handles that case separately.
			},
			wantAction: "proceed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()
			sessionRepo := NewMockSessionRepository()

			_ = agentRepo.Create(context.Background(), &storage.Agent{
				ID:           agentID,
				ClientID:     ptr.To(id.ClientID("client-1")),
				DisplayName:  "Test Client",
				RedirectURIs: []string{"https://client.example.com/callback"},
			})
			activeGrant(grantRepo)
			tt.setupSession(sessionRepo)

			svc := NewServiceWithSessions(agentRepo, grantRepo, sessionRepo, cfg, nil)

			decision, err := svc.HandleAuthorization(context.Background(), authReq, id.NewPrincipal("user@example.com"))

			require.NoError(t, err)
			require.NotNil(t, decision)
			assert.Equal(t, tt.wantAction, decision.Action)
		})
	}
}

// TestService_HandleAuthorization_PreservesParameters tests that OAuth2 parameters are preserved
func TestService_HandleAuthorization_PreservesParameters(t *testing.T) {
	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	// Add agent
	agent := &storage.Agent{
		ID:           agentID,
		ClientID:     ptr.To(id.ClientID("client-1")),
		RedirectURIs: []string{"https://client.example.com/callback"},
	}
	_ = agentRepo.Create(context.Background(), agent)

	// Add active grant
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

	svc := NewService(agentRepo, grantRepo, &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		PublicURL:                 "https://broker.example.com",
	})

	authReq := &ports.AuthorizationRequest{
		ClientID:            id.ClientID(agentID.String()),
		RedirectURI:         "https://client.example.com/callback",
		Scope:               "openid profile email",
		State:               "state123",
		ResponseType:        "code",
		CodeChallenge:       "E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I",
		CodeChallengeMethod: "S256",
	}

	decision, err := svc.HandleAuthorization(context.Background(), authReq, id.NewPrincipal("user@example.com"))

	require.NoError(t, err)
	require.Equal(t, "proceed", decision.Action)

	// Verify the redirect URL contains all parameters
	redirectURL := decision.RedirectURL
	assert.Contains(t, redirectURL, "client_id=client-1")
	assert.Contains(t, redirectURL, "redirect_uri=https%3A%2F%2Fclient.example.com%2Fcallback")
	assert.Contains(t, redirectURL, "scope=openid+profile+email")
	assert.Contains(t, redirectURL, "state=state123")
	assert.Contains(t, redirectURL, "response_type=code")
	assert.Contains(t, redirectURL, "code_challenge=E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I")
	assert.Contains(t, redirectURL, "code_challenge_method=S256")
}

// TestService_HandleAuthorization_UUIDResolution verifies agent resolution via the authorize endpoint.
// The client_id value is parsed as a UUID internally by the service to look up the agent.
func TestService_HandleAuthorization_UUIDResolution(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	setupAgent := func(r *MockAgentRepository) {
		agent := &storage.Agent{
			ID:           agentID,
			ClientID:     ptr.To(id.ClientID("upstream-client-1")), // upstream OAuth2 client ID
			DisplayName:  "Test Agent",
			RedirectURIs: []string{"https://client.example.com/callback"},
		}
		_ = r.Create(context.Background(), agent)
	}

	setupActiveGrant := func(r *MockGrantRepository) {
		grant := &storage.UserGrant{
			ID:        id.NewGrantID(),
			Principal: id.Principal("user@example.com"),
			AgentID:   agentID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
			},
		}
		_ = r.Create(context.Background(), grant)
	}

	unknownAgentID := id.NewAgentID()

	tests := []struct {
		name          string
		setupAgent    func(*MockAgentRepository)
		setupGrant    func(*MockGrantRepository)
		clientID      id.ClientID
		wantAction    string
		wantErrorCode string
	}{
		{
			name:          "valid agent UUID resolves agent and redirects to upstream",
			setupAgent:    setupAgent,
			setupGrant:    setupActiveGrant,
			clientID:      id.ClientID(agentID.String()),
			wantAction:    "proceed",
			wantErrorCode: "",
		},
		{
			name:          "well-formed UUID that is not a registered agent returns invalid_client",
			setupAgent:    func(r *MockAgentRepository) {}, // empty repo
			setupGrant:    func(r *MockGrantRepository) {},
			clientID:      id.ClientID(unknownAgentID.String()),
			wantAction:    "error",
			wantErrorCode: "invalid_client",
		},
		{
			name:          "malformed UUID string returns invalid_client",
			setupAgent:    func(r *MockAgentRepository) {},
			setupGrant:    func(r *MockGrantRepository) {},
			clientID:      id.ClientID("not-a-uuid"),
			wantAction:    "error",
			wantErrorCode: "invalid_client",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()

			tt.setupAgent(agentRepo)
			tt.setupGrant(grantRepo)

			svc := NewService(agentRepo, grantRepo, &OAuth2Config{
				UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
				PublicURL:                 "https://broker.example.com",
				SupportedResponseTypes:    []string{"code"},
			})

			req := &ports.AuthorizationRequest{
				ClientID:     tt.clientID,
				RedirectURI:  "https://client.example.com/callback",
				ResponseType: "code",
				State:        "state123",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + tt.clientID.String(),
			}

			decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))

			require.NoError(t, err)
			require.NotNil(t, decision)
			assert.Equal(t, tt.wantAction, decision.Action)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, decision.ErrorCode)
			}
		})
	}
}

// TestService_HandleAuthorization_UUIDResolution_UpstreamClientID verifies that the
// upstream authorize URL uses agent.ClientID (upstream OAuth2 client ID), NOT the
// broker's internal agent UUID.
func TestService_HandleAuthorization_UUIDResolution_UpstreamClientID(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	agent := &storage.Agent{
		ID:           agentID,
		ClientID:     ptr.To(id.ClientID("upstream-client-abc")), // this is what should appear in upstream URL
		DisplayName:  "Test Agent",
		RedirectURIs: []string{"https://client.example.com/callback"},
	}
	_ = agentRepo.Create(context.Background(), agent)

	grant := &storage.UserGrant{
		ID:        id.NewGrantID(),
		Principal: id.Principal("user@example.com"),
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
		},
	}
	_ = grantRepo.Create(context.Background(), grant)

	svc := NewService(agentRepo, grantRepo, &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		PublicURL:                 "https://broker.example.com",
	})

	req := &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		State:        "state123",
	}

	decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))

	require.NoError(t, err)
	assert.Equal(t, "proceed", decision.Action)

	// The upstream URL MUST use the agent's upstream ClientID, NOT the internal UUID
	assert.Contains(t, decision.RedirectURL, "client_id=upstream-client-abc",
		"upstream URL must use agent.ClientID (upstream OAuth2 client ID), not the broker UUID")
	assert.NotContains(t, decision.RedirectURL, agentID.String(),
		"upstream URL must NOT expose the broker's internal agent UUID as client_id")
}

// TestService_GenerateMetadata tests RFC 8414 metadata generation
func TestService_GenerateMetadata(t *testing.T) {
	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	config := &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		UpstreamTokenEndpoint:     "https://auth.example.com/token",
		PublicURL:                 "https://broker.example.com",
		SupportedResponseTypes:    []string{"code"},
		SupportedGrantTypes:       []string{"authorization_code", "refresh_token"},
	}

	svc := NewService(agentRepo, grantRepo, config)

	tests := []struct {
		name string
		want *ports.MetadataResponse
	}{
		{
			name: "valid metadata generation",
			want: &ports.MetadataResponse{
				Issuer:                            "https://broker.example.com",
				AuthorizationEndpoint:             "https://broker.example.com/oauth2/authorize",
				TokenEndpoint:                     "https://broker.example.com/oauth2/token",
				ResponseTypesSupported:            []string{"code"},
				GrantTypesSupported:               []string{"authorization_code", "refresh_token"},
				TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := svc.GenerateMetadata(context.Background())

			require.NoError(t, err)
			require.NotNil(t, metadata)

			assert.Equal(t, tt.want.Issuer, metadata.Issuer)
			assert.Equal(t, tt.want.AuthorizationEndpoint, metadata.AuthorizationEndpoint)
			assert.Equal(t, tt.want.TokenEndpoint, metadata.TokenEndpoint)
			assert.Equal(t, tt.want.ResponseTypesSupported, metadata.ResponseTypesSupported)
			assert.Equal(t, tt.want.GrantTypesSupported, metadata.GrantTypesSupported)
			assert.Equal(t, tt.want.TokenEndpointAuthMethodsSupported, metadata.TokenEndpointAuthMethodsSupported)
		})
	}
}

// TestService_HandleAuthorization_MultiAgentParamInjection tests that the agent UUID
// is appended to the upstream authorize URL when multi_agent_client is enabled,
// and is absent when the feature is disabled (Feature 021).
func TestService_HandleAuthorization_MultiAgentParamInjection(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	makeRepos := func() (*MockAgentRepository, *MockGrantRepository) {
		agentRepo := NewMockAgentRepository()
		grantRepo := NewMockGrantRepository()

		agent := &storage.Agent{
			ID:           agentID,
			ClientID:     ptr.To(id.ClientID("shared-upstream-client")),
			DisplayName:  "Test Agent",
			RedirectURIs: []string{"https://client.example.com/callback"},
		}
		_ = agentRepo.Create(context.Background(), agent)

		grant := &storage.UserGrant{
			ID:        id.NewGrantID(),
			Principal: id.Principal("user@example.com"),
			AgentID:   agentID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: serviceID, Scopes: []string{"openid"}},
			},
		}
		_ = grantRepo.Create(context.Background(), grant)
		return agentRepo, grantRepo
	}

	req := &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		State:        "state-xyz",
	}

	t.Run("param injected when multi_agent_client enabled", func(t *testing.T) {
		agentRepo, grantRepo := makeRepos()
		svc := NewService(agentRepo, grantRepo, &OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
			MultiAgentClient: ports.MultiAgentClientConfig{
				Enabled:          true,
				AgentIDParamName: "x_agent_id",
				AgentIDClaimName: "x_agent_id",
			},
		})

		decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))
		require.NoError(t, err)
		assert.Equal(t, "proceed", decision.Action)
		assert.Contains(t, decision.RedirectURL, "x_agent_id="+agentID.String(),
			"agent UUID must be injected as x_agent_id param when multi_agent_client is enabled")
	})

	t.Run("param absent when multi_agent_client disabled", func(t *testing.T) {
		agentRepo, grantRepo := makeRepos()
		svc := NewService(agentRepo, grantRepo, &OAuth2Config{
			UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
			PublicURL:                 "https://broker.example.com",
			MultiAgentClient:          ports.MultiAgentClientConfig{Enabled: false},
		})

		decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))
		require.NoError(t, err)
		assert.Equal(t, "proceed", decision.Action)
		assert.NotContains(t, decision.RedirectURL, "x_agent_id",
			"agent UUID param must not be present when multi_agent_client is disabled")
	})
}

// TestService_HandleAuthorization_RedirectURIValidation verifies that when an agent has
// registered redirect URIs, only those URIs are accepted. Agents with no registered URIs
// are rejected — fail closed per RFC 6749 §4.1.2.1.
func TestService_HandleAuthorization_RedirectURIValidation(t *testing.T) {
	agentID := id.NewAgentID()
	registeredURI := "https://client.example.com/callback"
	unregisteredURI := "https://evil.example.com/steal"

	makeAgent := func(uris []string) *storage.Agent {
		return &storage.Agent{
			ID:           agentID,
			ClientID:     ptr.To(id.ClientID("client-1")),
			DisplayName:  "Test Agent",
			RedirectURIs: uris,
		}
	}

	tests := []struct {
		name          string
		redirectURIs  []string // agent's registered URIs
		requestURI    string   // URI from the incoming request
		wantAction    string
		wantErrorCode string
	}{
		{
			name:          "agent with no registered URIs rejects any redirect_uri (fail closed)",
			redirectURIs:  nil,
			requestURI:    unregisteredURI,
			wantAction:    "error",
			wantErrorCode: "invalid_redirect_uri",
		},
		{
			name:         "matching registered URI is accepted",
			redirectURIs: []string{registeredURI},
			requestURI:   registeredURI,
			wantAction:   "redirect_to_consent",
		},
		{
			name:          "non-matching URI returns invalid_redirect_uri with no redirect",
			redirectURIs:  []string{registeredURI},
			requestURI:    unregisteredURI,
			wantAction:    "error",
			wantErrorCode: "invalid_redirect_uri",
		},
		{
			name:          "non-local http URI in registry rejected at runtime (legacy data guard)",
			redirectURIs:  []string{"http://legacy.example.com/callback"},
			requestURI:    "http://legacy.example.com/callback",
			wantAction:    "error",
			wantErrorCode: "invalid_redirect_uri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()
			_ = agentRepo.Create(context.Background(), makeAgent(tt.redirectURIs))

			svc := NewService(agentRepo, grantRepo, &OAuth2Config{
				PublicURL: "https://broker.example.com",
			})

			req := &ports.AuthorizationRequest{
				ClientID:     id.ClientID(agentID.String()),
				RedirectURI:  tt.requestURI,
				ResponseType: "code",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + agentID.String(),
			}

			decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))

			require.NoError(t, err)
			require.NotNil(t, decision)
			assert.Equal(t, tt.wantAction, decision.Action)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, decision.ErrorCode)
				assert.Empty(t, decision.RedirectURL, "invalid_redirect_uri must not include a redirect URL")
			}
		})
	}
}

// TestService_HandleAuthorization_ScopeValidation verifies that when an agent has
// registered allowed scopes, only those scopes may be requested. Agents with no
// allowed scopes accept any requested scope.
func TestService_HandleAuthorization_ScopeValidation(t *testing.T) {
	agentID := id.NewAgentID()

	makeAgent := func(allowed []string) *storage.Agent {
		return &storage.Agent{
			ID:            agentID,
			ClientID:      ptr.To(id.ClientID("client-1")),
			DisplayName:   "Test Agent",
			RedirectURIs:  []string{"https://client.example.com/callback"},
			AllowedScopes: allowed,
		}
	}

	tests := []struct {
		name          string
		allowedScopes []string
		requestScope  string
		wantAction    string
		wantErrorCode string
		wantRedirect  bool // true if error should carry a redirect URL
	}{
		{
			name:          "agent with no allowed scopes accepts any requested scope",
			allowedScopes: nil,
			requestScope:  "openid profile email",
			wantAction:    "redirect_to_consent",
		},
		{
			name:          "requesting only allowed scopes is accepted",
			allowedScopes: []string{"openid", "profile"},
			requestScope:  "openid profile",
			wantAction:    "redirect_to_consent",
		},
		{
			name:          "requesting an unauthorized scope returns invalid_scope redirect",
			allowedScopes: []string{"openid"},
			requestScope:  "openid admin",
			wantAction:    "error",
			wantErrorCode: "invalid_scope",
			wantRedirect:  true,
		},
		{
			name:          "empty scope with restricted agent is accepted",
			allowedScopes: []string{"openid"},
			requestScope:  "",
			wantAction:    "redirect_to_consent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()
			_ = agentRepo.Create(context.Background(), makeAgent(tt.allowedScopes))

			svc := NewService(agentRepo, grantRepo, &OAuth2Config{
				PublicURL: "https://broker.example.com",
			})

			req := &ports.AuthorizationRequest{
				ClientID:     id.ClientID(agentID.String()),
				RedirectURI:  "https://client.example.com/callback",
				ResponseType: "code",
				Scope:        tt.requestScope,
				State:        "state-abc",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=" + agentID.String(),
			}

			decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))

			require.NoError(t, err)
			require.NotNil(t, decision)
			assert.Equal(t, tt.wantAction, decision.Action)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, decision.ErrorCode)
			}
			if tt.wantRedirect {
				assert.NotEmpty(t, decision.RedirectURL, "invalid_scope should include a redirect URL")
				assert.Contains(t, decision.RedirectURL, "error=invalid_scope")
				assert.Contains(t, decision.RedirectURL, "state=state-abc")
			}
		})
	}
}

// TestService_GenerateMetadata_RFC8414Compliance tests RFC 8414 compliance
func TestService_GenerateMetadata_RFC8414Compliance(t *testing.T) {
	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	config := &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		UpstreamTokenEndpoint:     "https://auth.example.com/token",
		PublicURL:                 "https://broker.example.com",
		SupportedResponseTypes:    []string{"code"},
		SupportedGrantTypes:       []string{"authorization_code"},
	}

	svc := NewService(agentRepo, grantRepo, config)
	metadata, err := svc.GenerateMetadata(context.Background())

	require.NoError(t, err)

	// RFC 8414 Section 2 requires these fields
	assert.NotEmpty(t, metadata.Issuer, "issuer must not be empty")
	assert.NotEmpty(t, metadata.AuthorizationEndpoint, "authorization_endpoint must not be empty")
	assert.NotEmpty(t, metadata.TokenEndpoint, "token_endpoint must not be empty")
	assert.NotEmpty(t, metadata.ResponseTypesSupported, "response_types_supported must not be empty")
	assert.NotEmpty(t, metadata.GrantTypesSupported, "grant_types_supported must not be empty")

	// Verify issuer is HTTPS
	assert.True(t, strings.HasPrefix(metadata.Issuer, "https://"), "issuer must use HTTPS")

	// Verify endpoints are HTTPS
	assert.True(t, strings.HasPrefix(metadata.AuthorizationEndpoint, "https://"), "authorization_endpoint must use HTTPS")
	assert.True(t, strings.HasPrefix(metadata.TokenEndpoint, "https://"), "token_endpoint must use HTTPS")
}

// errorGrantRepository is a mock that returns a configurable error from FindByPrincipalAndAgent.
type errorGrantRepository struct {
	MockGrantRepository
	findErr error
}

func (r *errorGrantRepository) FindByPrincipalAndAgent(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.UserGrant, error) {
	return nil, r.findErr
}

// TestService_HandleAuthorization_GrantLookupError verifies that when the grant repository
// returns a non-not-found error (e.g., connection failure) after redirect_uri is validated,
// the service returns a server_error decision with a redirect URL (safe to redirect because
// redirect_uri has already been verified).
func TestService_HandleAuthorization_GrantLookupError(t *testing.T) {
	agentID := id.NewAgentID()
	connErr := storage.NewStorageError("FindByPrincipalAndAgent", storage.ErrorKindConnection, nil, "connection refused")

	agentRepo := NewMockAgentRepository()
	_ = agentRepo.Create(context.Background(), &storage.Agent{
		ID:           agentID,
		ClientID:     ptr.To(id.ClientID("client-1")),
		DisplayName:  "Test Agent",
		RedirectURIs: []string{"https://client.example.com/callback"},
	})

	grantRepo := &errorGrantRepository{
		MockGrantRepository: *NewMockGrantRepository(),
		findErr:             connErr,
	}

	svc := NewService(agentRepo, grantRepo, &OAuth2Config{
		PublicURL: "https://broker.example.com",
	})

	req := &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		State:        "abc123",
		OriginalURL:  "https://broker.example.com/oauth2/authorize",
	}

	decision, err := svc.HandleAuthorization(context.Background(), req, id.NewPrincipal("user@example.com"))

	require.NoError(t, err)
	require.NotNil(t, decision)
	assert.Equal(t, "error", decision.Action)
	assert.Equal(t, "server_error", decision.ErrorCode)
	assert.NotEmpty(t, decision.RedirectURL, "post-validation server_error should carry a redirect URL")
	assert.Contains(t, decision.RedirectURL, "https://client.example.com/callback")
	assert.Contains(t, decision.RedirectURL, "error=server_error")
}

// TestService_HandleAuthorization_MandatoryRequirements tests that step 5 of HandleAuthorization
// enforces mandatory service requirements on agents with active grants.
func TestService_HandleAuthorization_MandatoryRequirements(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	cfg := &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		PublicURL:                 "https://broker.example.com",
	}

	authReq := &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		OriginalURL:  "https://broker.example.com/oauth2/authorize",
	}

	// activeGrant sets up an agent with a mandatory service requirement and an active grant.
	// DelegatedOAuth2Tokens is empty so step 4 (session expiry) does not interfere.
	setupAgent := func(agentRepo *MockAgentRepository) {
		_ = agentRepo.Create(context.Background(), &storage.Agent{
			ID:           agentID,
			ClientID:     ptr.To(id.ClientID("client-1")),
			DisplayName:  "Test Agent",
			RedirectURIs: []string{"https://client.example.com/callback"},
			ServiceRequirements: []storage.ServiceRequirement{
				{
					ServiceID:       serviceID,
					RequirementType: storage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			},
		})
	}

	setupGrant := func(grantRepo *MockGrantRepository) {
		_ = grantRepo.Create(context.Background(), &storage.UserGrant{
			ID:                    id.NewGrantID(),
			Principal:             id.Principal("user@example.com"),
			AgentID:               agentID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{},
		})
	}

	tests := []struct {
		name          string
		setupSess     func(*MockSessionRepository)
		wantAction    string
		wantErrorCode string
	}{
		{
			name: "active session with required scopes proceeds",
			setupSess: func(r *MockSessionRepository) {
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return &storage.UserSession{
						ID:        id.NewSessionID(),
						Principal: id.Principal("user@example.com"),
						ServiceID: serviceID,
						Scope:     []string{"repo", "user:email", "read:user"},
					}, nil
				}
			},
			wantAction: "proceed",
		},
		{
			name: "storage error returns server error",
			setupSess: func(r *MockSessionRepository) {
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return nil, storage.NewStorageError(
						"FindByPrincipalAndService",
						storage.ErrorKindConnection,
						nil,
						"database unavailable",
					)
				}
			},
			wantAction:    "error",
			wantErrorCode: "server_error",
		},
		{
			name: "missing session redirects to consent",
			setupSess: func(r *MockSessionRepository) {
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return nil, nil
				}
			},
			wantAction: "redirect_to_consent",
		},
		{
			name: "session with insufficient scopes redirects to consent",
			setupSess: func(r *MockSessionRepository) {
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return &storage.UserSession{
						ID:        id.NewSessionID(),
						Principal: id.Principal("user@example.com"),
						ServiceID: serviceID,
						Scope:     []string{"repo"}, // missing user:email
					}, nil
				}
			},
			wantAction: "redirect_to_consent",
		},
		{
			name: "expired session redirects to consent",
			setupSess: func(r *MockSessionRepository) {
				expiredAt := time.Now().Add(-1 * time.Hour)
				r.findFunc = func(_ context.Context, _ id.Principal, _ id.ServiceID) (*storage.UserSession, error) {
					return &storage.UserSession{
						ID:                    id.NewSessionID(),
						Principal:             id.Principal("user@example.com"),
						ServiceID:             serviceID,
						Scope:                 []string{"repo", "user:email"},
						RefreshTokenExpiresAt: &expiredAt,
					}, nil
				}
			},
			wantAction: "redirect_to_consent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agentRepo := NewMockAgentRepository()
			grantRepo := NewMockGrantRepository()
			sessionRepo := NewMockSessionRepository()

			setupAgent(agentRepo)
			setupGrant(grantRepo)
			tt.setupSess(sessionRepo)

			svc := NewServiceWithSessions(agentRepo, grantRepo, sessionRepo, cfg, nil)
			decision, err := svc.HandleAuthorization(context.Background(), authReq, id.NewPrincipal("user@example.com"))

			require.NoError(t, err)
			require.NotNil(t, decision)
			assert.Equal(t, tt.wantAction, decision.Action)
			if tt.wantErrorCode != "" {
				assert.Equal(t, tt.wantErrorCode, decision.ErrorCode)
			}
		})
	}
}

func TestService_HandleAuthorization_InvalidUpstreamAuthorizeURL(t *testing.T) {
	agentID := id.NewAgentID()

	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	require.NoError(t, agentRepo.Create(context.Background(), &storage.Agent{
		ID:           agentID,
		ClientID:     ptr.To(id.ClientID("client-1")),
		DisplayName:  "Test Agent",
		RedirectURIs: []string{"https://client.example.com/callback"},
	}))

	require.NoError(t, grantRepo.Create(context.Background(), &storage.UserGrant{
		ID:                    id.NewGrantID(),
		Principal:             id.Principal("user@example.com"),
		AgentID:               agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{},
	}))

	svc := NewService(agentRepo, grantRepo, &OAuth2Config{
		UpstreamAuthorizeEndpoint: "%",
		PublicURL:                 "https://broker.example.com",
	}).(*Service)

	decision, err := svc.HandleAuthorization(context.Background(), &ports.AuthorizationRequest{
		ClientID:     id.ClientID(agentID.String()),
		RedirectURI:  "https://client.example.com/callback",
		ResponseType: "code",
		State:        "xyz",
		OriginalURL:  "https://broker.example.com/oauth2/authorize",
	}, id.NewPrincipal("user@example.com"))

	require.NoError(t, err)
	require.NotNil(t, decision)
	assert.Equal(t, "error", decision.Action)
	assert.Equal(t, "server_error", decision.ErrorCode)
	assert.Contains(t, decision.RedirectURL, "error=server_error")
}
