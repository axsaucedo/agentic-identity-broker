package oauth2

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockAgentRepository is a test double for AgentRepository
type MockAgentRepository struct {
	agents map[string]*storage.Agent
}

func NewMockAgentRepository() *MockAgentRepository {
	return &MockAgentRepository{
		agents: make(map[string]*storage.Agent),
	}
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	m.agents[agent.ID] = agent
	return nil
}

func (m *MockAgentRepository) Get(ctx context.Context, id string) (*storage.Agent, error) {
	agent, ok := m.agents[id]
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

func (m *MockAgentRepository) Delete(ctx context.Context, id string) error {
	delete(m.agents, id)
	return nil
}

func (m *MockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	var agents []*storage.Agent
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents, nil
}

func (m *MockAgentRepository) GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error) {
	for _, agent := range m.agents {
		if agent.ClientID == clientID {
			return agent, nil
		}
	}
	return nil, ports.ErrNotFound
}

// MockGrantRepository is a test double for UserGrantRepository
type MockGrantRepository struct {
	grants map[string]*storage.UserGrant
}

func NewMockGrantRepository() *MockGrantRepository {
	return &MockGrantRepository{
		grants: make(map[string]*storage.UserGrant),
	}
}

func (m *MockGrantRepository) Create(ctx context.Context, grant *storage.UserGrant) error {
	m.grants[grant.ID] = grant
	return nil
}

func (m *MockGrantRepository) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	grant, ok := m.grants[id]
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

func (m *MockGrantRepository) Delete(ctx context.Context, id string) error {
	delete(m.grants, id)
	return nil
}

func (m *MockGrantRepository) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	var grants []*storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID && grant.IsActive() {
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (m *MockGrantRepository) FindByPrincipalAndAgent(ctx context.Context, principal, agentID string) (*storage.UserGrant, error) {
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant, nil
		}
	}
	return nil, nil
}

func (m *MockGrantRepository) DeleteByAgent(ctx context.Context, agentID string) error {
	for id, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, id)
		}
	}
	return nil
}

func (m *MockGrantRepository) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	var grants []storage.UserGrant
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.IsActive() {
			grants = append(grants, *grant)
		}
	}
	return grants, nil
}

func (m *MockGrantRepository) CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error) {
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

func (m *MockGrantRepository) ListByServiceID(ctx context.Context, serviceID string) ([]string, error) {
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

// TestService_HandleAuthorization tests the HandleAuthorization method with table-driven tests
func TestService_HandleAuthorization(t *testing.T) {
	tests := []struct {
		name       string
		setupAgent func(*MockAgentRepository)
		setupGrant func(*MockGrantRepository)
		authReq    *ports.AuthorizationRequest
		principal  string
		wantAction string
	}{
		{
			name:       "invalid client_id returns error redirect",
			setupAgent: func(r *MockAgentRepository) {},
			setupGrant: func(r *MockGrantRepository) {},
			authReq: &ports.AuthorizationRequest{
				ClientID:     "unknown-client",
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
			},
			principal:  "user@example.com",
			wantAction: "error",
		},
		{
			name: "valid client_id with no grant redirects to consent UI",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:          "agent-1",
					ClientID:    "client-1",
					DisplayName: "Test Client",
				}
				r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {},
			authReq: &ports.AuthorizationRequest{
				ClientID:     "client-1",
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=client-1",
			},
			principal:  "user@example.com",
			wantAction: "redirect_to_consent",
		},
		{
			name: "valid client_id with active grant redirects to upstream",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:          "agent-1",
					ClientID:    "client-1",
					DisplayName: "Test Client",
				}
				r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {
				grant := &storage.UserGrant{
					ID:         "grant-1",
					Principal:  "user@example.com",
					AgentID:    "agent-1",
					ValidUntil: nil,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"openid"}},
					},
				}
				r.Create(context.Background(), grant)
			},
			authReq: &ports.AuthorizationRequest{
				ClientID:     "client-1",
				RedirectURI:  "https://client.example.com/callback",
				Scope:        "openid profile",
				State:        "xyz123",
				ResponseType: "code",
			},
			principal:  "user@example.com",
			wantAction: "redirect_to_upstream",
		},
		{
			name: "expired grant redirects to consent UI",
			setupAgent: func(r *MockAgentRepository) {
				agent := &storage.Agent{
					ID:          "agent-1",
					ClientID:    "client-1",
					DisplayName: "Test Client",
				}
				r.Create(context.Background(), agent)
			},
			setupGrant: func(r *MockGrantRepository) {
				expiredTime := time.Now().Add(-1 * time.Hour)
				grant := &storage.UserGrant{
					ID:         "grant-1",
					Principal:  "user@example.com",
					AgentID:    "agent-1",
					ValidUntil: &expiredTime,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"openid"}},
					},
				}
				r.Create(context.Background(), grant)
			},
			authReq: &ports.AuthorizationRequest{
				ClientID:     "client-1",
				RedirectURI:  "https://client.example.com/callback",
				State:        "xyz123",
				ResponseType: "code",
				OriginalURL:  "https://broker.example.com/oauth2/authorize?client_id=client-1",
			},
			principal:  "user@example.com",
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

// TestService_HandleAuthorization_PreservesParameters tests that OAuth2 parameters are preserved
func TestService_HandleAuthorization_PreservesParameters(t *testing.T) {
	agentRepo := NewMockAgentRepository()
	grantRepo := NewMockGrantRepository()

	// Add agent
	agent := &storage.Agent{ID: "agent-1", ClientID: "client-1"}
	agentRepo.Create(context.Background(), agent)

	// Add active grant
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

	svc := NewService(agentRepo, grantRepo, &OAuth2Config{
		UpstreamAuthorizeEndpoint: "https://auth.example.com/authorize",
		PublicURL:                 "https://broker.example.com",
	})

	authReq := &ports.AuthorizationRequest{
		ClientID:            "client-1",
		RedirectURI:         "https://client.example.com/callback",
		Scope:               "openid profile email",
		State:               "state123",
		ResponseType:        "code",
		CodeChallenge:       "E9Mrozoa2owQB2dSBnnNBvjrNqtPTUAwY5uQp41VN-I",
		CodeChallengeMethod: "S256",
	}

	decision, err := svc.HandleAuthorization(context.Background(), authReq, "user@example.com")

	require.NoError(t, err)
	require.Equal(t, "redirect_to_upstream", decision.Action)

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
