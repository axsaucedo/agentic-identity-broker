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

	memorystorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockConsentService implements a mock consent service for testing.
// It is shared across all test files in this package (grants_handler_test.go uses getActiveGrantsFunc).
type mockConsentService struct {
	getAgentConsentInfoFunc       func(ctx context.Context, agentID id.AgentID, principal id.Principal) (*consent.AgentConsentInfo, error)
	getActiveGrantsFunc           func(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
	grantConsentFunc              func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error)
	revokeConsentFunc             func(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	revokeConsentForPrincipalFunc func(ctx context.Context, principal id.Principal, agentID id.AgentID) error
	getAgentDelegationsFunc       func(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error)
	getUserGrantsFunc             func(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
}

var _ ConsentService = (*mockConsentService)(nil)

func (m *mockConsentService) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID, principal id.Principal) (*consent.AgentConsentInfo, error) {
	if m.getAgentConsentInfoFunc != nil {
		return m.getAgentConsentInfoFunc(ctx, agentID, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockConsentService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	if m.grantConsentFunc != nil {
		return m.grantConsentFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockConsentService) RevokeConsent(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if m.revokeConsentFunc != nil {
		return m.revokeConsentFunc(ctx, principal, agentID)
	}
	return errors.New("not implemented")
}

func (m *mockConsentService) RevokeConsentForPrincipal(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if m.revokeConsentForPrincipalFunc != nil {
		return m.revokeConsentForPrincipalFunc(ctx, principal, agentID)
	}
	return errors.New("not implemented")
}

func (m *mockConsentService) GetAgentWithServiceRequirements(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
	return nil, nil, errors.New("not implemented")
}

func (m *mockConsentService) GetAgentDelegations(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error) {
	if m.getAgentDelegationsFunc != nil {
		return m.getAgentDelegationsFunc(ctx, principal)
	}
	return nil, errors.New("not implemented")
}

func (m *mockConsentService) GetUserGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	if m.getUserGrantsFunc != nil {
		return m.getUserGrantsFunc(ctx, principal, agentID)
	}
	if m.getActiveGrantsFunc != nil {
		return m.getActiveGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}

func encryptSecretForTest(serviceID, secret string) []byte {
	enc := newTestEncryption()
	ciphertext, err := enc.Encrypt(context.Background(), []byte(secret), map[string]string{"service_id": serviceID})
	if err != nil {
		panic("encryptSecretForTest: " + err.Error())
	}
	return ciphertext
}

func TestGetAgentConsentInfo_Success(t *testing.T) {
	// Setup mock data
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-aaa000000123")
	govURL := "https://example.com/governance"
	docURL := "https://example.com/docs"
	interfaceURL := "https://example.com/interface"

	mockAgent := &storage.Agent{
		ID:                   agentID,
		ClientID:             ptr.To(id.ClientID("client-github")),
		DisplayName:          "GitHub Assistant",
		Description:          "AI assistant for GitHub",
		GovernanceURL:        &govURL,
		UserDocumentationURL: &docURL,
		AgentInterfaceURL:    &interfaceURL,
		CreatedAt:            time.Date(2025, 12, 17, 10, 0, 0, 0, time.UTC),
		UpdatedAt:            time.Date(2025, 12, 17, 11, 0, 0, 0, time.UTC),
	}

	mockServices := []*model.ThirdpartyOAuth2ProviderEntity{
		{
			ID:          id.MustParseServiceID("00000000-0000-0000-0000-ccc000000001"),
			DisplayName: "GitHub",
			Scopes: []model.OAuthScope{
				{ScopeValue: "repo", Description: "Full control of private repositories"},
				{ScopeValue: "user:email", Description: "Access user emails"},
			},
			Secret: model.NewEncryptedSecret(encryptSecretForTest("00000000-0000-0000-0000-ccc000000001", "test-client-secret")),
		},
		{
			ID:          id.MustParseServiceID("00000000-0000-0000-0000-ccc000000002"),
			DisplayName: "Google",
			Scopes: []model.OAuthScope{
				{ScopeValue: "openid", Description: "OpenID Connect"},
				{ScopeValue: "email", Description: "Access email address"},
			},
			Secret: model.NewEncryptedSecret(encryptSecretForTest("00000000-0000-0000-0000-ccc000000002", "test-client-secret")),
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

	// Create request with principal in context
	ctx := context.Background()
	ctx = principal.WithPrincipal(ctx, "user@example.com")
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
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
	assert.Equal(t, agentID.String(), resp.Agent.ID)
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
	assert.Equal(t, "00000000-0000-0000-0000-ccc000000001", githubSvc.ID)
	assert.Equal(t, "GitHub", githubSvc.DisplayName)
	assert.Len(t, githubSvc.Scopes, 2)
	assert.Equal(t, "repo", githubSvc.Scopes[0].ScopeValue)
	assert.Equal(t, "Full control of private repositories", githubSvc.Scopes[0].Description)

	// Check Google service
	googleSvc := resp.RequestedServices[1]
	assert.Equal(t, "00000000-0000-0000-0000-ccc000000002", googleSvc.ID)
	assert.Equal(t, "Google", googleSvc.DisplayName)
	assert.Len(t, googleSvc.Scopes, 2)
}

func TestGetAgentConsentInfo_AgentNotFound(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-bbb000000000")

	mockSvc := &mockConsentServiceWrapper{
		info: nil,
		err:  consent.ErrAgentNotFound,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	ctx := principal.WithPrincipal(context.Background(), "user@example.com")
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
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
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-aaa000000123")

	mockAgent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID("client-test")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Empty services array
	mockInfo := &consent.AgentConsentInfo{
		Agent:                       mockAgent,
		AvailableThirdpartyServices: []*model.ThirdpartyOAuth2ProviderEntity{},
	}

	mockSvc := &mockConsentServiceWrapper{
		info: mockInfo,
		err:  nil,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	ctx := principal.WithPrincipal(context.Background(), "user@example.com")
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	handler.GetAgentConsentInfo(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp AgentConsentInfoResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, agentID.String(), resp.Agent.ID)
	assert.Empty(t, resp.RequestedServices)
}

func TestGetAgentConsentInfo_ServiceError(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-aaa000000123")

	mockSvc := &mockConsentServiceWrapper{
		info: nil,
		err:  errors.New("database connection failed"),
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	ctx := principal.WithPrincipal(context.Background(), "user@example.com")
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	req = req.WithContext(ctx)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
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
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-aaa000000123")

	mockAgent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID("client-test")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockInfo := &consent.AgentConsentInfo{
		Agent:                       mockAgent,
		AvailableThirdpartyServices: []*model.ThirdpartyOAuth2ProviderEntity{},
	}

	mockSvc := &mockConsentServiceWrapper{
		info: mockInfo,
		err:  nil,
	}

	logger := slog.Default()
	handler := NewAgentInfoHandler(mockSvc.asService(), logger)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
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
	var services []*model.ThirdpartyOAuth2ProviderEntity

	if m.info != nil {
		agent = m.info.Agent
		services = m.info.AvailableThirdpartyServices
	}

	mockAgentRepo := &mockAgentRepo{agent: agent, err: m.err}
	mockServiceRepo := &mockServiceRepo{services: services, err: m.err}
	mockGrantRepo := &mockGrantRepo{}

	return consent.NewService(mockAgentRepo, newTestProviderService(mockServiceRepo), mockGrantRepo, memorystorage.NewInMemoryUserSessionRepository(), nil, slog.Default())
}

// Mock repository implementations
type mockAgentRepo struct {
	agent *storage.Agent
	err   error
}

func (m *mockAgentRepo) Create(ctx context.Context, agent *storage.Agent) error {
	return nil
}

func (m *mockAgentRepo) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
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

func (m *mockAgentRepo) Delete(ctx context.Context, agentID id.AgentID) error {
	return nil
}

func (m *mockAgentRepo) List(ctx context.Context) ([]*storage.Agent, error) {
	return nil, nil
}

func (m *mockAgentRepo) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	if m.agent != nil && m.agent.ClientID != nil && *m.agent.ClientID == clientID {
		return m.agent, nil
	}
	return nil, m.err
}

func (m *mockAgentRepo) GetByClientURI(_ context.Context, _ string) (*storage.Agent, error) {
	return nil, storage.NewStorageError("GetAgentByClientURI", storage.ErrorKindNotFound, nil, "not found")
}

func (m *mockAgentRepo) ExistsOtherWithClientID(_ context.Context, _ id.ClientID, _ *id.AgentID) (bool, error) {
	return false, nil
}

type mockServiceRepo struct {
	services []*model.ThirdpartyOAuth2ProviderEntity
	err      error
}

func (m *mockServiceRepo) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *mockServiceRepo) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

func (m *mockServiceRepo) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return nil
}

func (m *mockServiceRepo) Delete(ctx context.Context, serviceID id.ServiceID) error {
	return nil
}

func (m *mockServiceRepo) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.services, nil
}

func (m *mockServiceRepo) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

func (m *mockServiceRepo) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, nil
}

type mockGrantRepo struct{}

func (m *mockGrantRepo) Create(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepo) Get(ctx context.Context, grantID id.GrantID) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) Update(ctx context.Context, grant *storage.UserGrant) error {
	return nil
}

func (m *mockGrantRepo) Delete(ctx context.Context, grantID id.GrantID) error {
	return nil
}

func (m *mockGrantRepo) ListByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) FindByPrincipalAndAgent(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) DeleteByAgent(ctx context.Context, agentID id.AgentID) error {
	return nil
}

func (m *mockGrantRepo) ListByPrincipal(ctx context.Context, principal id.Principal) ([]storage.UserGrant, error) {
	return nil, nil
}

func (m *mockGrantRepo) CountAgentsByServiceID(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, nil
}

func (m *mockGrantRepo) ListByServiceID(ctx context.Context, serviceID id.ServiceID) ([]id.AgentID, error) {
	return []id.AgentID{}, nil
}

func (m *mockGrantRepo) DeleteByPrincipalAndAgentID(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	return nil
}

func (m *mockGrantRepo) CountGrantsReferencingPermissionSet(_ context.Context, _ id.PermissionSetID) (int, error) {
	return 0, nil
}
