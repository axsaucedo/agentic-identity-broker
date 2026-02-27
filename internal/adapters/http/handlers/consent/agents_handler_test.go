package consent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetAgentDelegations_Success tests successful retrieval of agent delegations.
func TestGetAgentDelegations_Success(t *testing.T) {
	// Setup mock data
	principalValue := "user@example.com"
	expiresAt := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	mockDelegations := []consent.AgentDelegation{
		{
			AgentID:          "agent-1",
			DisplayName:      "Data Analysis Assistant",
			LogoURL:          nil,
			ActiveGrantCount: 3,
			LastModifiedAt:   time.Date(2025, 12, 17, 14, 30, 0, 0, time.UTC),
			ExpiresAt:        nil,
		},
		{
			AgentID:          "agent-2",
			DisplayName:      "Document Processor",
			LogoURL:          nil,
			ActiveGrantCount: 2,
			LastModifiedAt:   time.Date(2025, 12, 16, 9, 15, 0, 0, time.UTC),
			ExpiresAt:        &expiresAt,
		},
	}

	// Create mock service
	mockSvc := &mockAgentsService{
		delegations: mockDelegations,
		err:         nil,
	}

	// Create handler
	handler := NewAgentsHandler(mockSvc.asService(), nil)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agents", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	// Execute request
	rec := httptest.NewRecorder()
	handler.GetAgentDelegations(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp GetAgentDelegationsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify delegations (order is non-deterministic due to map iteration)
	assert.Len(t, resp.Data, 2)

	// Build a map for easier lookup
	delegationMap := make(map[string]consent.AgentDelegation)
	for _, d := range resp.Data {
		delegationMap[d.AgentID] = d
	}

	// Verify agent-1
	agent1, ok := delegationMap["agent-1"]
	require.True(t, ok, "agent-1 should be in response")
	assert.Equal(t, "Data Analysis Assistant", agent1.DisplayName)
	assert.Equal(t, 3, agent1.ActiveGrantCount)
	assert.Nil(t, agent1.ExpiresAt)

	// Verify agent-2
	agent2, ok := delegationMap["agent-2"]
	require.True(t, ok, "agent-2 should be in response")
	assert.Equal(t, "Document Processor", agent2.DisplayName)
	assert.Equal(t, 2, agent2.ActiveGrantCount)
	assert.NotNil(t, agent2.ExpiresAt)
	assert.Equal(t, expiresAt, *agent2.ExpiresAt)
}

// TestGetAgentDelegations_EmptyList tests successful retrieval with no delegations.
func TestGetAgentDelegations_EmptyList(t *testing.T) {
	principalValue := "user@example.com"

	mockSvc := &mockAgentsService{
		delegations: []consent.AgentDelegation{},
		err:         nil,
	}

	handler := NewAgentsHandler(mockSvc.asService(), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agents", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetAgentDelegations(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp GetAgentDelegationsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Empty(t, resp.Data)
}

// TestGetAgentDelegations_MissingPrincipal tests error when principal is not in context.
func TestGetAgentDelegations_MissingPrincipal(t *testing.T) {
	mockSvc := &mockAgentsService{}
	handler := NewAgentsHandler(mockSvc.asService(), nil)

	// Request without principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agents", nil)

	rec := httptest.NewRecorder()
	handler.GetAgentDelegations(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
}

// TestGetAgentDelegations_ServiceError tests error handling when service fails.
func TestGetAgentDelegations_ServiceError(t *testing.T) {
	principalValue := "user@example.com"

	mockSvc := &mockAgentsService{
		delegations: nil,
		err:         errors.New("database connection failed"),
	}

	handler := NewAgentsHandler(mockSvc.asService(), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agents", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetAgentDelegations(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", resp.Error)
}

// TestGetAgentDelegations_ContentType tests that response has correct content type.
func TestGetAgentDelegations_ContentType(t *testing.T) {
	principalValue := "user@example.com"

	mockSvc := &mockAgentsService{
		delegations: []consent.AgentDelegation{},
		err:         nil,
	}

	handler := NewAgentsHandler(mockSvc.asService(), nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agents", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetAgentDelegations(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

// mockAgentsService wraps mock data to implement consent.Service behavior for agent delegations.
type mockAgentsService struct {
	delegations []consent.AgentDelegation
	err         error
}

// asService returns a consent.Service that uses the mock data.
func (m *mockAgentsService) asService() *consent.Service {
	// Create mock agent repo that returns agents with proper display names
	agentMap := make(map[string]*storage.Agent)
	for _, delegation := range m.delegations {
		agentMap[delegation.AgentID] = &storage.Agent{
			ID:          delegation.AgentID,
			DisplayName: delegation.DisplayName,
		}
	}

	mockAgentRepo := &mockAgentRepoForAgents{
		agents: agentMap,
	}
	mockServiceRepo := &mockServiceRepoForAgents{}
	mockGrantRepo := &mockGrantRepoForAgents{
		delegations: m.delegations,
		err:         m.err,
	}

	return consent.NewService(mockAgentRepo, newTestProviderService(mockServiceRepo), mockGrantRepo)
}

// Mock repository implementations for agents handler tests
type mockAgentRepoForAgents struct {
	agents map[string]*storage.Agent
}

func (m *mockAgentRepoForAgents) Create(ctx context.Context, agent *storage.Agent) error {
	return nil
}

func (m *mockAgentRepoForAgents) Get(ctx context.Context, id string) (*storage.Agent, error) {
	// Return agent from map if exists, otherwise return a dummy
	if agent, ok := m.agents[id]; ok {
		return agent, nil
	}
	return &storage.Agent{
		ID:          id,
		DisplayName: "Agent " + id,
	}, nil
}

func (m *mockAgentRepoForAgents) Update(ctx context.Context, agent *storage.Agent) error {
	return nil
}

func (m *mockAgentRepoForAgents) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAgentRepoForAgents) List(ctx context.Context) ([]*storage.Agent, error) {
	return nil, nil
}

func (m *mockAgentRepoForAgents) GetByClientID(ctx context.Context, clientID string) (*storage.Agent, error) {
	for _, agent := range m.agents {
		if agent.ClientID == clientID {
			return agent, nil
		}
	}
	return nil, nil
}

type mockServiceRepoForAgents struct{}

func (m *mockServiceRepoForAgents) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *mockServiceRepoForAgents) Get(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

func (m *mockServiceRepoForAgents) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *mockServiceRepoForAgents) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockServiceRepoForAgents) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

func (m *mockServiceRepoForAgents) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	return 0, nil
}

func (m *mockServiceRepoForAgents) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

type mockGrantRepoForAgents struct {
	delegations []consent.AgentDelegation
	err         error
}

func (m *mockGrantRepoForAgents) Create(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepoForAgents) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepoForAgents) Update(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepoForAgents) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockGrantRepoForAgents) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepoForAgents) FindByPrincipalAndAgent(ctx context.Context, principal string, agentID string) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepoForAgents) DeleteByAgent(ctx context.Context, agentID string) error {
	return nil
}

func (m *mockGrantRepoForAgents) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	// This is the method that GetAgentDelegations calls
	// We need to return grants that will result in the expected delegations
	if m.err != nil {
		return nil, m.err
	}

	// Convert delegations back to grants for the mock
	// In a real implementation, the service would aggregate these
	var grants []storage.UserGrant
	for _, delegation := range m.delegations {
		grant := storage.UserGrant{
			ID:         "grant-" + delegation.AgentID,
			Principal:  principal,
			AgentID:    delegation.AgentID,
			ValidUntil: delegation.ExpiresAt,
			UpdatedAt:  delegation.LastModifiedAt,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: "service-1",
					Scopes:                    []string{"scope1"},
				},
			},
		}
		// Add multiple grants if ActiveGrantCount > 1 to simulate aggregation
		for i := 0; i < delegation.ActiveGrantCount; i++ {
			grants = append(grants, grant)
		}
	}

	return grants, nil
}

func (m *mockGrantRepoForAgents) CountAgentsByServiceID(ctx context.Context, serviceID string) (int, error) {
	return 0, nil
}

func (m *mockGrantRepoForAgents) ListByServiceID(ctx context.Context, serviceID string) ([]string, error) {
	return []string{}, nil
}
