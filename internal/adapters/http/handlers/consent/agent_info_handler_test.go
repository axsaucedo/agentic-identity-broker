package consent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConsentService implements a mock consent service for testing.
//
//nolint:unused // Used in tests
type mockConsentService struct {
	getAgentConsentInfoFunc func(ctx context.Context, agentID string) (*consent.AgentConsentInfo, error)
	getActiveGrantsFunc     func(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error)
	grantConsentFunc        func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error)
	revokeConsentFunc       func(ctx context.Context, principal, agentID string) error
	getAgentDelegationsFunc func(ctx context.Context, principal string) ([]consent.AgentDelegation, error)
	getAgentDetailFunc      func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
	getUserGrantsFunc       func(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error)
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentConsentInfo(ctx context.Context, agentID string) (*consent.AgentConsentInfo, error) {
	if m.getAgentConsentInfoFunc != nil {
		return m.getAgentConsentInfoFunc(ctx, agentID)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetActiveGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	if m.getActiveGrantsFunc != nil {
		return m.getActiveGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	if m.grantConsentFunc != nil {
		return m.grantConsentFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) RevokeConsent(ctx context.Context, principal, agentID string) error {
	if m.revokeConsentFunc != nil {
		return m.revokeConsentFunc(ctx, principal, agentID)
	}
	return errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentDelegations(ctx context.Context, principal string) ([]consent.AgentDelegation, error) {
	if m.getAgentDelegationsFunc != nil {
		return m.getAgentDelegationsFunc(ctx, principal)
	}
	return nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
	if m.getAgentDetailFunc != nil {
		return m.getAgentDetailFunc(ctx, agentID)
	}
	return nil, nil, errors.New("not implemented")
}

//nolint:unused // Used in tests
func (m *mockConsentService) GetUserGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	if m.getUserGrantsFunc != nil {
		return m.getUserGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}

func TestGetAgentConsentInfo_Success(t *testing.T) {
	// Setup mock data
	agentID := "agent-123"
	govURL := "https://example.com/governance"
	docURL := "https://example.com/docs"
	interfaceURL := "https://example.com/interface"

	mockAgent := &storage.Agent{
		ID:                   agentID,
		ClientID:             "client-github",
		DisplayName:          "GitHub Assistant",
		Description:          "AI assistant for GitHub",
		GovernanceURL:        &govURL,
		UserDocumentationURL: &docURL,
		AgentInterfaceURL:    &interfaceURL,
		CreatedAt:            time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC),
		UpdatedAt:            time.Date(2025, 12, 17, 11, 0, 0, 0, time.UTC),
	}

	mockServices := []*storage.ThirdpartyOAuth2Service{
		{
			ID:          "service-github",
			DisplayName: "GitHub",
			Scopes: []storage.OAuthScope{
				{ScopeValue: "repo", Description: "Full control of private repositories"},
				{ScopeValue: "user:email", Description: "Access user emails"},
			},
		},
		{
			ID:          "service-google",
			DisplayName: "Google",
			Scopes: []storage.OAuthScope{
				{ScopeValue: "openid", Description: "OpenID Connect"},
				{ScopeValue: "email", Description: "Access email address"},
			},
		},
	}

	mockInfo := &consent.AgentConsentInfo{
		Agent:                       mockAgent,
		AvailableThirdpartyServices: mockServices,
	}

	// Create mock service with a real consent.Service-like interface
	// We'll use a wrapper that implements the interface
	mockSvc := &mockConsentServiceWrapper{
		info: mockInfo,
		err:  nil,
	}

	// Create handler
	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Execute request
	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp AgentConsentInfoResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify agent metadata
	assert.Equal(t, agentID, resp.Agent.ID)
	assert.Equal(t, "client-github", resp.Agent.ClientID)
	assert.Equal(t, "GitHub Assistant", resp.Agent.DisplayName)
	assert.Equal(t, "AI assistant for GitHub", resp.Agent.Description)
	assert.NotNil(t, resp.Agent.GovernanceURL)
	assert.Equal(t, govURL, *resp.Agent.GovernanceURL)
	assert.NotNil(t, resp.Agent.UserDocumentationURL)
	assert.Equal(t, docURL, *resp.Agent.UserDocumentationURL)
	assert.NotNil(t, resp.Agent.AgentInterfaceURL)
	assert.Equal(t, interfaceURL, *resp.Agent.AgentInterfaceURL)

	// Verify services
	assert.Len(t, resp.RequestedServices, 2)

	// Check GitHub service
	githubSvc := resp.RequestedServices[0]
	assert.Equal(t, "service-github", githubSvc.ID)
	assert.Equal(t, "GitHub", githubSvc.DisplayName)
	assert.Len(t, githubSvc.Scopes, 2)
	assert.Equal(t, "repo", githubSvc.Scopes[0].ScopeValue)
	assert.Equal(t, "Full control of private repositories", githubSvc.Scopes[0].Description)

	// Check Google service
	googleSvc := resp.RequestedServices[1]
	assert.Equal(t, "service-google", googleSvc.ID)
	assert.Equal(t, "Google", googleSvc.DisplayName)
	assert.Len(t, googleSvc.Scopes, 2)
}

func TestGetAgentConsentInfo_AgentNotFound(t *testing.T) {
	agentID := "nonexistent-agent"

	mockSvc := &mockConsentServiceWrapper{
		info: nil,
		err:  consent.ErrAgentNotFound,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "agent not found", resp.Error)
}

func TestGetAgentConsentInfo_EmptyServices(t *testing.T) {
	agentID := "agent-123"

	mockAgent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-test",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Empty services array
	mockInfo := &consent.AgentConsentInfo{
		Agent:                       mockAgent,
		AvailableThirdpartyServices: []*storage.ThirdpartyOAuth2Service{},
	}

	mockSvc := &mockConsentServiceWrapper{
		info: mockInfo,
		err:  nil,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp AgentConsentInfoResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, agentID, resp.Agent.ID)
	assert.Empty(t, resp.RequestedServices)
}

func TestGetAgentConsentInfo_ServiceError(t *testing.T) {
	agentID := "agent-123"

	mockSvc := &mockConsentServiceWrapper{
		info: nil,
		err:  errors.New("database connection failed"),
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "internal server error", resp.Error)
}

func TestGetAgentConsentInfo_ContentType(t *testing.T) {
	agentID := "agent-123"

	mockAgent := &storage.Agent{
		ID:          agentID,
		ClientID:    "client-test",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockInfo := &consent.AgentConsentInfo{
		Agent:                       mockAgent,
		AvailableThirdpartyServices: []*storage.ThirdpartyOAuth2Service{},
	}

	mockSvc := &mockConsentServiceWrapper{
		info: mockInfo,
		err:  nil,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
}

// mockConsentServiceWrapper wraps mock data to implement consent.Service behavior
type mockConsentServiceWrapper struct {
	info *consent.AgentConsentInfo
	err  error
}

// asService returns a consent.Service that uses the mock data
func (m *mockConsentServiceWrapper) asService() *consent.Service {
	// We'll create a service with nil repositories since we won't actually call them
	// The handler only calls GetAgentConsentInfo, so we can use this approach
	// In a real implementation, we'd use proper mocks or create a ConsentService interface

	// For now, we'll use a different approach: create a real service with mock repos
	var agent *storage.Agent
	var services []*storage.ThirdpartyOAuth2Service

	if m.info != nil {
		agent = m.info.Agent
		services = m.info.AvailableThirdpartyServices
	}

	mockAgentRepo := &mockAgentRepo{agent: agent, err: m.err}
	mockServiceRepo := &mockServiceRepo{services: services, err: m.err}
	mockGrantRepo := &mockGrantRepo{}

	return consent.NewService(mockAgentRepo, mockServiceRepo, mockGrantRepo)
}

// Mock repository implementations
type mockAgentRepo struct {
	agent *storage.Agent
	err   error
}

func (m *mockAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	return nil
}

func (m *mockAgentRepo) Get(ctx context.Context, id string) (*storage.Agent, error) {
	if m.err != nil {
		// For ErrAgentNotFound, we return nil agent
		return nil, m.err
	}
	if m.agent == nil {
		return nil, nil
	}
	return m.agent, nil
}

func (m *mockAgentRepo) Update(ctx context.Context, agent *storage.Agent) error {
	return nil
}

func (m *mockAgentRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockAgentRepo) List(ctx context.Context) ([]*storage.Agent, error) {
	return nil, nil
}

type mockServiceRepo struct {
	services []*storage.ThirdpartyOAuth2Service
	err      error
}

func (m *mockServiceRepo) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	return nil
}

func (m *mockServiceRepo) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	return nil, nil
}

func (m *mockServiceRepo) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	return nil
}

func (m *mockServiceRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockServiceRepo) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.services, nil
}

func (m *mockServiceRepo) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	return 0, nil
}

type mockGrantRepo struct{}

func (m *mockGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepo) Get(ctx context.Context, id string) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) Update(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepo) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal string, agentID string) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) DeleteByAgent(ctx context.Context, agentID string) error {
	return nil
}

func (m *mockGrantRepo) ListByPrincipal(ctx context.Context, principal string) ([]storage.UserGrant, error) {
	return nil, nil
}
