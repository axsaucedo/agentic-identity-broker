package consent

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// testEncryption is a non-identity test double for EncryptionPort used in domain-layer
// unit tests. It applies an invertible XOR transformation (key byte 0x55) so that
// ciphertext ≠ plaintext, keeping the Secret value-object state machine honest without
// importing any adapter package.
//
// This is NOT a production noop. The runtime noop fallback was removed in Phase 6.
// Adapter-layer and integration tests should use testutil.NewTestEncryptionAdapter(t)
// for real AES-256-GCM roundtrips.
type testEncryption struct{}

func (e *testEncryption) Encrypt(_ context.Context, plaintext []byte, _ map[string]string) ([]byte, error) {
	out := make([]byte, len(plaintext))
	for i, b := range plaintext {
		out[i] = b ^ 0x55
	}
	return out, nil
}

func (e *testEncryption) Decrypt(_ context.Context, ciphertext []byte, _ map[string]string) ([]byte, error) {
	out := make([]byte, len(ciphertext))
	for i, b := range ciphertext {
		out[i] = b ^ 0x55
	}
	return out, nil
}

// newTestProviderService wraps a ThirdpartyOAuth2ProviderRepository in a domain service
// with a non-identity test double for encryption. Used in domain-layer tests that exercise
// consent business logic, not encryption correctness.
func newTestProviderService(repo ports.ThirdpartyOAuth2ProviderRepository) *thirdparty.ThirdpartyOAuth2ProviderService {
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, &testEncryption{}, nil, false, nil)
}

// Mock implementations for testing

type mockAgentRepo struct {
	agents map[id.AgentID]*storage.Agent
	err    error
}

func (m *mockAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	if m.err != nil {
		return m.err
	}
	m.agents[agent.ID] = agent.Copy()
	return nil
}

func (m *mockAgentRepo) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	agent, exists := m.agents[agentID]
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

func (m *mockAgentRepo) Delete(ctx context.Context, agentID id.AgentID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.agents, agentID)
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

func (m *mockAgentRepo) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, agent := range m.agents {
		if agent.ClientID == clientID {
			return agent.Copy(), nil
		}
	}
	return nil, ports.ErrNotFound
}

type mockServiceRepo struct {
	services map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity
	err      error
}

func (m *mockServiceRepo) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	if m.err != nil {
		return m.err
	}
	m.services[entity.ID] = entity.Copy()
	return nil
}

func (m *mockServiceRepo) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	entity, exists := m.services[serviceID]
	if !exists {
		return nil, ports.ErrNotFound
	}
	return entity.Copy(), nil
}

func (m *mockServiceRepo) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	if m.err != nil {
		return m.err
	}
	if _, exists := m.services[entity.ID]; !exists {
		return ports.ErrNotFound
	}
	m.services[entity.ID] = entity.Copy()
	return nil
}

func (m *mockServiceRepo) Delete(ctx context.Context, serviceID id.ServiceID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.services, serviceID)
	return nil
}

func (m *mockServiceRepo) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*model.ThirdpartyOAuth2ProviderEntity, 0, len(m.services))
	for _, entity := range m.services {
		result = append(result, entity.Copy())
	}
	return result, nil
}

func (m *mockServiceRepo) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	return 0, nil
}

func (m *mockServiceRepo) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, entity := range m.services {
		for _, resource := range entity.ProtectedResources {
			if resource == resourceURI {
				return entity.Copy(), nil
			}
		}
	}
	return nil, ports.ErrNotFound
}

type mockGrantRepo struct {
	grants map[id.GrantID]*storage.UserGrant
	err    error
}

func (m *mockGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	if m.err != nil {
		return m.err
	}
	// Generate ID if not set
	if grant.ID.IsZero() {
		grant.ID = id.NewGrantID()
	}
	m.grants[grant.ID] = grant.Copy()
	return nil
}

func (m *mockGrantRepo) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	if m.err != nil {
		return nil, m.err
	}
	grant, exists := m.grants[grantID]
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

func (m *mockGrantRepo) Delete(ctx context.Context, grantID id.GrantID) error {
	if m.err != nil {
		return m.err
	}
	delete(m.grants, grantID)
	return nil
}

func (m *mockGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
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

func (m *mockGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
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

func (m *mockGrantRepo) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	if m.err != nil {
		return m.err
	}
	for grantKey, grant := range m.grants {
		if grant.AgentID == agentID {
			delete(m.grants, grantKey)
		}
	}
	return nil
}

func (m *mockGrantRepo) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := []storage.UserGrant{}
	for _, grant := range m.grants {
		if grant.Principal == principal && grant.IsActive() {
			result = append(result, *grant.Copy())
		}
	}
	return result, nil
}

func (m *mockGrantRepo) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	uniqueAgents := make(map[id.AgentID]bool)
	for _, grant := range m.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				uniqueAgents[grant.AgentID] = true
				break
			}
		}
	}
	return len(uniqueAgents), nil
}

func (m *mockGrantRepo) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	if m.err != nil {
		return nil, m.err
	}
	uniqueAgents := make(map[id.AgentID]bool)
	for _, grant := range m.grants {
		for _, token := range grant.DelegatedOAuth2Tokens {
			if token.ThirdpartyOAuth2ServiceID == serviceID {
				uniqueAgents[grant.AgentID] = true
				break
			}
		}
	}
	agentIDs := make([]id.AgentID, 0, len(uniqueAgents))
	for agentID := range uniqueAgents {
		agentIDs = append(agentIDs, agentID)
	}
	return agentIDs, nil
}

func (m *mockGrantRepo) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if m.err != nil {
		return m.err
	}
	for grantID, grant := range m.grants {
		if grant.Principal == principal && grant.AgentID == agentID {
			delete(m.grants, grantID)
			return nil
		}
	}
	return ports.ErrNotFound
}

// Test cases

func TestService_GetAgentConsentInfo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentID := id.NewAgentID()
	serviceID1 := id.NewServiceID()

	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.ClientID("test-client"),
		DisplayName: "Test Agent",
		Description: "A test agent",
	}

	service1 := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          serviceID1,
		DisplayName: "GitHub",
		ClientID:    id.ClientID("github-client"),
		IssuerURI:   "https://github.com",
		Discovery:   model.DiscoveryConfig{EnableDiscovery: true},
		Scopes: []model.OAuthScope{
			{ScopeValue: "repo", Description: "Repository access"},
		},
		Secret: model.NewEncryptedSecret([]byte("test-ciphertext")),
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{agentID: agent}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{serviceID1: service1}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		info, err := svc.GetAgentConsentInfo(ctx, agentID)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, agentID, info.Agent.ID)
		assert.Len(t, info.AvailableThirdpartyServices, 1)
		// Client secret should be redacted
		plaintext, ptErr := info.AvailableThirdpartyServices[0].Secret.GetPlaintext()
		require.NoError(t, ptErr)
		assert.Equal(t, "REDACTED", plaintext)
	})

	t.Run("agent not found", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		info, err := svc.GetAgentConsentInfo(ctx, id.NewAgentID())
		assert.Error(t, err)
		assert.Nil(t, info)
	})
}

func TestService_GrantConsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentID := id.NewAgentID()
	serviceID1 := id.NewServiceID()

	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.ClientID("test-client"),
		DisplayName: "Test Agent",
		Description: "A test agent",
	}

	service1 := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          serviceID1,
		DisplayName: "GitHub",
		ClientID:    id.ClientID("github-client"),
		IssuerURI:   "https://github.com",
		Discovery:   model.DiscoveryConfig{EnableDiscovery: true},
		Scopes: []model.OAuthScope{
			{ScopeValue: "repo", Description: "Repository access"},
			{ScopeValue: "user:email", Description: "Email access"},
		},
		Secret: model.NewEncryptedSecret([]byte("test-ciphertext")),
	}

	t.Run("create new grant", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{agentID: agent}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{serviceID1: service1}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		future := time.Now().Add(24 * time.Hour)
		req := &GrantRequest{
			Principal:  id.Principal("user@example.com"),
			AgentID:    agentID,
			ValidUntil: &future,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: serviceID1,
					Scopes:                    []string{"repo", "user:email"},
				},
			},
		}

		grant, err := svc.GrantConsent(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, grant)
		assert.Equal(t, id.Principal("user@example.com"), grant.Principal)
		assert.Equal(t, agentID, grant.AgentID)
		assert.Len(t, grant.DelegatedOAuth2Tokens, 1)
	})

	t.Run("invalid scopes", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{agentID: agent}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{serviceID1: service1}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		req := &GrantRequest{
			Principal: id.Principal("user@example.com"),
			AgentID:   agentID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: serviceID1,
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
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		req := &GrantRequest{
			Principal: id.Principal("user@example.com"),
			AgentID:   id.NewAgentID(),
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: serviceID1,
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

// TestService_RevokeConsentForPrincipal tests the user-facing revoke method (FR-014).
// This is the dedicated method for DELETE /api/consent/agent/{agent-id}/grants — non-idempotent,
// maps storage not-found to ErrGrantNotFound so the handler can return 404.
func TestService_RevokeConsentForPrincipal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentID := id.NewAgentID()
	grantID := id.NewGrantID()
	serviceID1 := id.NewServiceID()

	existingGrant := &storage.UserGrant{
		ID:        grantID,
		Principal: id.Principal("user@example.com"),
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
		},
	}

	t.Run("success: grant is deleted", func(t *testing.T) {
		t.Parallel()
		grantRepo := &mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{grantID: existingGrant}}
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			grantRepo,
			slog.Default(),
		)

		err := svc.RevokeConsentForPrincipal(ctx, id.Principal("user@example.com"), agentID)
		require.NoError(t, err)
		assert.Empty(t, grantRepo.grants)
	})

	t.Run("grant not found: returns ErrGrantNotFound (not raw ErrNotFound)", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		err := svc.RevokeConsentForPrincipal(ctx, id.Principal("user@example.com"), agentID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrGrantNotFound, "must surface domain sentinel, not raw storage error")
	})

	t.Run("cross-principal: different principal cannot revoke another's grant (SR-001)", func(t *testing.T) {
		t.Parallel()
		grantRepo := &mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{grantID: existingGrant}}
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			grantRepo,
			slog.Default(),
		)

		// Different principal has no grant for this agent
		err := svc.RevokeConsentForPrincipal(ctx, id.Principal("other@example.com"), agentID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrGrantNotFound)

		// Original grant still exists
		assert.Len(t, grantRepo.grants, 1)
	})
}

func TestService_RevokeConsent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentID := id.NewAgentID()
	grantID := id.NewGrantID()
	serviceID1 := id.NewServiceID()

	existingGrant := &storage.UserGrant{
		ID:        grantID,
		Principal: id.Principal("user@example.com"),
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: serviceID1,
				Scopes:                    []string{"repo"},
			},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		grantRepo := &mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{grantID: existingGrant}}
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			grantRepo,
			slog.Default(),
		)

		err := svc.RevokeConsent(ctx, id.Principal("user@example.com"), agentID)
		require.NoError(t, err)
		assert.Empty(t, grantRepo.grants)
	})

	t.Run("grant not found - returns nil (idempotent: POST empty-tokens path)", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		err := svc.RevokeConsent(ctx, id.Principal("user@example.com"), agentID)
		require.NoError(t, err) // Idempotent: absence is not an error
	})
}

func TestService_GetActiveGrants(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agentID := id.NewAgentID()
	grantID1 := id.NewGrantID()
	grantID2 := id.NewGrantID()
	grantID3 := id.NewGrantID()
	serviceID1 := id.NewServiceID()

	past := time.Now().Add(-24 * time.Hour)
	future := time.Now().Add(24 * time.Hour)

	activeGrant := &storage.UserGrant{
		ID:         grantID1,
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: &future,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
		},
	}

	expiredGrant := &storage.UserGrant{
		ID:         grantID2,
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: &past,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
		},
	}

	indefiniteGrant := &storage.UserGrant{
		ID:         grantID3,
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: nil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
		},
	}

	t.Run("filters expired grants", func(t *testing.T) {
		t.Parallel()
		grantRepo := &mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{
			grantID1: activeGrant,
			grantID2: expiredGrant,
			grantID3: indefiniteGrant,
		}}
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			grantRepo,
			slog.Default(),
		)

		grants, err := svc.GetActiveGrants(ctx, id.Principal("user@example.com"), agentID)
		require.NoError(t, err)
		// Should return only active and indefinite grants (not expired)
		assert.Len(t, grants, 2)
	})

	t.Run("no grants - returns empty slice", func(t *testing.T) {
		t.Parallel()
		svc := NewService(
			&mockAgentRepo{agents: map[id.AgentID]*storage.Agent{}},
			newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
			&mockGrantRepo{grants: map[id.GrantID]*storage.UserGrant{}},
			slog.Default(),
		)

		grants, err := svc.GetActiveGrants(ctx, id.Principal("user@example.com"), agentID)
		require.NoError(t, err)
		assert.Empty(t, grants)
	})
}

func TestService_GetAgentDelegations(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	agent1ID := id.NewAgentID()
	agent2ID := id.NewAgentID()
	serviceID1 := id.NewServiceID()
	grantID1 := id.NewGrantID()
	grantID2 := id.NewGrantID()

	agent1 := &storage.Agent{
		ID:          agent1ID,
		ClientID:    id.ClientID("test-client-1"),
		DisplayName: "Test Agent 1",
		Description: "First test agent",
	}

	agent2 := &storage.Agent{
		ID:          agent2ID,
		ClientID:    id.ClientID("test-client-2"),
		DisplayName: "Test Agent 2",
		Description: "Second test agent",
	}

	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name          string
		principal     id.Principal
		grants        map[id.GrantID]*storage.UserGrant
		agents        map[id.AgentID]*storage.Agent
		expectedCount int
		expectError   bool
		validate      func(t *testing.T, delegations []AgentDelegation)
	}{
		{
			name:          "empty grants returns empty list",
			principal:     id.Principal("user@example.com"),
			grants:        map[id.GrantID]*storage.UserGrant{},
			agents:        map[id.AgentID]*storage.Agent{agent1ID: agent1},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:      "single grant returns one delegation",
			principal: id.Principal("user@example.com"),
			grants: map[id.GrantID]*storage.UserGrant{
				grantID1: {
					ID:         grantID1,
					Principal:  id.Principal("user@example.com"),
					AgentID:    agent1ID,
					ValidUntil: &future,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			agents:        map[id.AgentID]*storage.Agent{agent1ID: agent1},
			expectedCount: 1,
			expectError:   false,
			validate: func(t *testing.T, delegations []AgentDelegation) {
				require.Len(t, delegations, 1)
				assert.Equal(t, agent1ID, delegations[0].AgentID)
				assert.Equal(t, "Test Agent 1", delegations[0].DisplayName)
				assert.Equal(t, 1, delegations[0].ActiveGrantCount)
			},
		},
		{
			name:      "multiple grants for same agent groups correctly",
			principal: id.Principal("user@example.com"),
			grants: map[id.GrantID]*storage.UserGrant{
				grantID1: {
					ID:         grantID1,
					Principal:  id.Principal("user@example.com"),
					AgentID:    agent1ID,
					ValidUntil: &future,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			agents:        map[id.AgentID]*storage.Agent{agent1ID: agent1},
			expectedCount: 1,
			expectError:   false,
			validate: func(t *testing.T, delegations []AgentDelegation) {
				require.Len(t, delegations, 1)
				assert.Equal(t, agent1ID, delegations[0].AgentID)
				assert.Equal(t, 1, delegations[0].ActiveGrantCount)
			},
		},
		{
			name:      "multiple agents returns multiple delegations",
			principal: id.Principal("user@example.com"),
			grants: map[id.GrantID]*storage.UserGrant{
				grantID1: {
					ID:         grantID1,
					Principal:  id.Principal("user@example.com"),
					AgentID:    agent1ID,
					ValidUntil: &future,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
				grantID2: {
					ID:         grantID2,
					Principal:  id.Principal("user@example.com"),
					AgentID:    agent2ID,
					ValidUntil: &future,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"user"}},
					},
					CreatedAt: now,
					UpdatedAt: now.Add(1 * time.Hour),
				},
			},
			agents: map[id.AgentID]*storage.Agent{
				agent1ID: agent1,
				agent2ID: agent2,
			},
			expectedCount: 2,
			expectError:   false,
			validate: func(t *testing.T, delegations []AgentDelegation) {
				require.Len(t, delegations, 2)

				// Find each agent in results
				var agent1Delegation, agent2Delegation *AgentDelegation
				for i := range delegations {
					if delegations[i].AgentID == agent1ID {
						agent1Delegation = &delegations[i]
					}
					if delegations[i].AgentID == agent2ID {
						agent2Delegation = &delegations[i]
					}
				}

				require.NotNil(t, agent1Delegation)
				require.NotNil(t, agent2Delegation)
				assert.Equal(t, "Test Agent 1", agent1Delegation.DisplayName)
				assert.Equal(t, "Test Agent 2", agent2Delegation.DisplayName)
				assert.Equal(t, 1, agent1Delegation.ActiveGrantCount)
				assert.Equal(t, 1, agent2Delegation.ActiveGrantCount)
			},
		},
		{
			name:      "expired grants are filtered out",
			principal: id.Principal("user@example.com"),
			grants: map[id.GrantID]*storage.UserGrant{
				grantID1: {
					ID:         grantID1,
					Principal:  id.Principal("user@example.com"),
					AgentID:    agent1ID,
					ValidUntil: &past, // Expired
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
					},
					CreatedAt: now.Add(-48 * time.Hour),
					UpdatedAt: now.Add(-48 * time.Hour),
				},
			},
			agents:        map[id.AgentID]*storage.Agent{agent1ID: agent1},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:      "grants for different principal are not included",
			principal: id.Principal("user@example.com"),
			grants: map[id.GrantID]*storage.UserGrant{
				grantID1: {
					ID:         grantID1,
					Principal:  id.Principal("other@example.com"),
					AgentID:    agent1ID,
					ValidUntil: &future,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{ThirdpartyOAuth2ServiceID: serviceID1, Scopes: []string{"repo"}},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			agents:        map[id.AgentID]*storage.Agent{agent1ID: agent1},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewService(
				&mockAgentRepo{agents: tt.agents},
				newTestProviderService(&mockServiceRepo{services: map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity{}}),
				&mockGrantRepo{grants: tt.grants},
				slog.Default(),
			)

			delegations, err := svc.GetAgentDelegations(ctx, tt.principal)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, delegations, tt.expectedCount)

			if tt.validate != nil {
				tt.validate(t, delegations)
			}
		})
	}
}
