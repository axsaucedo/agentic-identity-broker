package consent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Mock implementations for testing

type mockAgentRepo struct {
	agents map[string]*storage.Agent
	err    error
}

func (m *mockAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	if m.err != nil {
		return m.err
	}
	m.agents[agent.ID] = agent.Copy()
	return nil
}

func (m *mockAgentRepo) Get(ctx context.Context, id string) (*storage.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	agent, exists := m.agents[id]
	if !exists {
		return nil, ports.ErrNotFound
	}
	return agent.Copy(), nil
}

func (m *mockAgentRepo) Update(ctx context.Context, agent *storage.Agent) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.agents[agent.ID]; !exists {
		return ports.ErrNotFound
	}
	m.agents[agent.ID] = agent.Copy()
	return nil
}

func (m *mockAgentRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.agents, id)
	return nil
}

func (m *mockAgentRepo) List(ctx context.Context) ([]*storage.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*storage.Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		result = append(result, agent.Copy())
	}
	return result, nil
}

type mockServiceRepo struct {
	services map[string]*storage.ThirdpartyOAuth2Service
	err      error
}

func (m *mockServiceRepo) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	if m.err != nil {
		return m.err
	}
	m.services[service.ID] = service.Copy()
	return nil
}

func (m *mockServiceRepo) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	if m.err != nil {
		return nil, m.err
	}
	service, exists := m.services[id]
	if !exists {
		return nil, ports.ErrNotFound
	}
	return service.Copy(), nil
}

func (m *mockServiceRepo) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.services[service.ID]; !exists {
		return ports.ErrNotFound
	}
	m.services[service.ID] = service.Copy()
	return nil
}

func (m *mockServiceRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.services, id)
	return nil
}

func (m *mockServiceRepo) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*storage.ThirdpartyOAuth2Service, 0, len(m.services))
	for _, service := range m.services {
		result = append(result, service.Copy())
	}
	return result, nil
}

func (m *mockServiceRepo) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 0, nil
}

type mockGrantRepo struct {
	grants map[string]*storage.UserGrant
	err    error
}

func (m *mockGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	if m.err != nil {
		return m.err
	}
	// Generate ID if not set
	if grant.ID == "" {
		grant.ID = "generated-grant-id"
	}
	m.grants[grant.ID] = grant.Copy()
	return nil
}

func (m *mockGrantRepo) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	if m.err != nil {
		return nil, m.err
	}
	grant, exists := m.grants[id]
	if !exists {
		return nil, ports.ErrNotFound
	}
	return grant.Copy(), nil
}

func (m *mockGrantRepo) Update(ctx context.Context, grant *storage.UserGrant) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.grants[grant.ID]; !exists {
		return ports.ErrNotFound
	}
	m.grants[grant.ID] = grant.Copy()
	return nil
}

func (m *mockGrantRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.grants, id)
	return nil
}

func (m *mockGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := []*storage.UserGrant{}
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			result = append(result, grant.Copy())
		}
	}
	return result, nil
}

func (m *mockGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal string, agentID string) (*storage.UserGrant, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			return grant.Copy(), nil
		}
	}
	return nil, ports.ErrNotFound
}

func (m *mockGrantRepo) DeleteByAgent(ctx context.Context, agentID string) error {
	if m.err != nil {
		return m.err
	}
	for id, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, id)
		}
	}
	return nil
}

// Test cases

func TestService_GetAgentConsentInfo(t *testing.T) {
	ctx := context.Background()

	agent := &storage.Agent{
		ID:          "agent-1",
		ClientID:    "test-client",
		DisplayName: "Test Agent",
		Description: "A test agent",
	}

	service1 := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub",
		ClientID:     "github-client",
		ClientSecret: "secret123",
		IssuerURI:    "https://github.com",
		Discovery:    storage.DiscoveryConfig{EnableDiscovery: true},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Repository access"},
		},
	}

	t.Run("success", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{"agent-1": agent}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{"service-1": service1}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		info, err := svc.GetAgentConsentInfo(ctx, "agent-1")
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "agent-1", info.Agent.ID)
		assert.Len(t, info.AvailableThirdpartyServices, 1)
		// Client secret should be redacted
		assert.Equal(t, "REDACTED", info.AvailableThirdpartyServices[0].ClientSecret)
	})

	t.Run("agent not found", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		info, err := svc.GetAgentConsentInfo(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestService_GrantConsent(t *testing.T) {
	ctx := context.Background()

	agent := &storage.Agent{
		ID:          "agent-1",
		ClientID:    "test-client",
		DisplayName: "Test Agent",
		Description: "A test agent",
	}

	service1 := &storage.ThirdpartyOAuth2Service{
		ID:           "service-1",
		DisplayName:  "GitHub",
		ClientID:     "github-client",
		ClientSecret: "secret123",
		IssuerURI:    "https://github.com",
		Discovery:    storage.DiscoveryConfig{EnableDiscovery: true},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Repository access"},
			{ScopeValue: "user:email", Description: "Email access"},
		},
	}

	t.Run("create new grant", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{"agent-1": agent}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{"service-1": service1}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		future := time.Now().Add(24 * time.Hour)
		req := &GrantRequest{
			Principal:  "user@example.com",
			AgentID:    "agent-1",
			ValidUntil: &future,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: "service-1",
					Scopes:                    []string{"repo", "user:email"},
				},
			},
		}

		grant, err := svc.GrantConsent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, grant)
		assert.Equal(t, "user@example.com", grant.Principal)
		assert.Equal(t, "agent-1", grant.AgentID)
		assert.Len(t, grant.DelegatedOAuth2Tokens, 1)
	})

	t.Run("invalid scopes", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{"agent-1": agent}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{"service-1": service1}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		req := &GrantRequest{
			Principal: "user@example.com",
			AgentID:   "agent-1",
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: "service-1",
					Scopes:                    []string{"invalid-scope"},
				},
			},
		}

		grant, err := svc.GrantConsent(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, grant)
		assert.ErrorIs(t, err, ErrInvalidScopes)
	})

	t.Run("agent not found", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		req := &GrantRequest{
			Principal: "user@example.com",
			AgentID:   "nonexistent",
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: "service-1",
					Scopes:                    []string{"repo"},
				},
			},
		}

		grant, err := svc.GrantConsent(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, grant)
		assert.ErrorIs(t, err, ErrAgentNotFound)
	})
}

func TestService_RevokeConsent(t *testing.T) {
	ctx := context.Background()

	existingGrant := &storage.UserGrant{
		ID:        "grant-1",
		Principal: "user@example.com",
		AgentID:   "agent-1",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-1",
				Scopes:                    []string{"repo"},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		grantRepo := &mockGrantRepo{grants: map[string]*storage.UserGrant{"grant-1": existingGrant}}
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			grantRepo,
		)

		err := svc.RevokeConsent(ctx, "user@example.com", "agent-1")
		require.NoError(t, err)
		assert.Empty(t, grantRepo.grants)
	})

	t.Run("grant not found - idempotent", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		err := svc.RevokeConsent(ctx, "user@example.com", "agent-1")
		require.NoError(t, err) // Should not error
	})
}

func TestService_GetActiveGrants(t *testing.T) {
	ctx := context.Background()

	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	activeGrant := &storage.UserGrant{
		ID:         "grant-1",
		Principal:  "user@example.com",
		AgentID:    "agent-1",
		ValidUntil: &future,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"repo"}},
		},
	}

	expiredGrant := &storage.UserGrant{
		ID:         "grant-2",
		Principal:  "user@example.com",
		AgentID:    "agent-1",
		ValidUntil: &past,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"repo"}},
		},
	}

	indefiniteGrant := &storage.UserGrant{
		ID:         "grant-3",
		Principal:  "user@example.com",
		AgentID:    "agent-1",
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: "service-1", Scopes: []string{"repo"}},
		},
	}

	t.Run("filters expired grants", func(t *testing.T) {
		grantRepo := &mockGrantRepo{grants: map[string]*storage.UserGrant{
			"grant-1": activeGrant,
			"grant-2": expiredGrant,
			"grant-3": indefiniteGrant,
		}}
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			grantRepo,
		)

		grants, err := svc.GetActiveGrants(ctx, "user@example.com", "agent-1")
		require.NoError(t, err)
		// Should return only active and indefinite grants (not expired)
		assert.Len(t, grants, 2)
	})

	t.Run("no grants - returns empty slice", func(t *testing.T) {
		svc := NewService(
			&mockAgentRepo{agents: map[string]*storage.Agent{}},
			&mockServiceRepo{services: map[string]*storage.ThirdpartyOAuth2Service{}},
			&mockGrantRepo{grants: map[string]*storage.UserGrant{}},
		)

		grants, err := svc.GetActiveGrants(ctx, "user@example.com", "agent-1")
		require.NoError(t, err)
		assert.Empty(t, grants)
	})
}
