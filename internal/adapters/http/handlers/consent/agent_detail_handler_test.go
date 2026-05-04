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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/testutil"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestEncryption returns a real encryption adapter backed by the shared deterministic test key.
func newTestEncryption() ports.EncryptionPort {
	return testutil.NewPanicTestEncryptionAdapter()
}

// newTestProviderService wraps a ThirdpartyOAuth2ProviderRepository in a domain service
// with test encryption. Used in tests across the consent handler package.
func newTestProviderService(repo ports.ThirdpartyOAuth2ProviderRepository) *thirdparty.ThirdpartyOAuth2ProviderService {
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, newTestEncryption(), nil, false, slog.Default())
}

// encryptSecretForTest encrypts a plaintext secret using the test encryption adapter.
// The serviceID is used as the encryption context binding.
func encryptSecretForTest(serviceID, secret string) []byte {
	enc := newTestEncryption()
	ciphertext, err := enc.Encrypt(context.Background(), []byte(secret), map[string]string{"service_id": serviceID})
	if err != nil {
		panic("encryptSecretForTest: " + err.Error())
	}
	return ciphertext
}

// mockAgentDetailService is a mock implementation of consent.Service for testing.
type mockAgentDetailService struct{}

func (m *mockAgentDetailService) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) ValidateGrantRequest(ctx context.Context, req *consent.GrantRequest) error {
	return nil
}

func (m *mockAgentDetailService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) RevokeConsent(ctx context.Context, p id.Principal, agentID id.AgentID) error {
	return errors.New("not implemented")
}

func (m *mockAgentDetailService) GetActiveGrants(ctx context.Context, p id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentDelegations(ctx context.Context, p id.Principal) ([]consent.AgentDelegation, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetUserGrants(ctx context.Context, p id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) RevokeConsentForPrincipal(ctx context.Context, p id.Principal, agentID id.AgentID) error {
	return nil
}

func TestGetAgentDetail_Success(t *testing.T) {
	// Setup
	agentID := id.NewAgentID()
	principalID := "user@example.com"
	governanceURL := "https://example.com/governance"
	userDocsURL := "https://example.com/docs"
	agentInterfaceURL := "https://example.com/interface"
	githubServiceID := id.NewServiceID()
	googleServiceID := id.NewServiceID()

	mockService := &mockAgentDetailService{}

	// Create agent with service requirements
	agent := &storage.Agent{
		ID:                   agentID,
		ClientID:             id.NewClientID("test-client-id"),
		DisplayName:          "Test Agent",
		Description:          "A test agent for testing purposes",
		GovernanceURL:        &governanceURL,
		UserDocumentationURL: &userDocsURL,
		AgentInterfaceURL:    &agentInterfaceURL,
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       githubServiceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"read:user", "repo"},
			},
			{
				ServiceID:       googleServiceID,
				RequirementType: storage.RequirementTypeOptional,
				RequiredScopes:  []string{"email"},
			},
		},
	}

	// Create in-memory repositories and populate with test data
	ctx := context.Background()
	agentRepo := memory.NewAgentRepository()
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()
	providerSvc := newTestProviderService(serviceRepo)

	// Add test services through the provider service so secrets are properly encrypted
	github := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          githubServiceID,
		DisplayName: "GitHub",
		ClientID:    id.NewClientID("github-client-id"),
		IssuerURI:   "https://github.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read:user", Description: "Read user profile"},
			{ScopeValue: "repo", Description: "Full control of repositories"},
		},
		Secret: model.NewPlaintextSecret("github-client-secret"),
	}
	google := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          googleServiceID,
		DisplayName: "Google",
		ClientID:    id.NewClientID("google-client-id"),
		IssuerURI:   "https://accounts.google.com",
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth2.googleapis.com/token",
			AuthorizeEndpoint: "https://accounts.google.com/o/oauth2/v2/auth",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "email", Description: "View email address"},
		},
		Secret: model.NewPlaintextSecret("google-client-secret"),
	}
	if err := providerSvc.Create(ctx, github); err != nil {
		t.Fatalf("failed to create github service: %v", err)
	}
	if err := providerSvc.Create(ctx, google); err != nil {
		t.Fatalf("failed to create google service: %v", err)
	}

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(providerSvc)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)

	// Add chi route context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add principal to context
	req = req.WithContext(principal.WithPrincipal(req.Context(), principalID))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var response GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify agent details
	if response.Data.Agent.AgentID != agentID {
		t.Errorf("expected agent ID %s, got %s", agentID, response.Data.Agent.AgentID)
	}
	if response.Data.Agent.DisplayName != "Test Agent" {
		t.Errorf("expected display name 'Test Agent', got %s", response.Data.Agent.DisplayName)
	}
	if response.Data.Agent.GovernanceURL == nil || *response.Data.Agent.GovernanceURL != governanceURL {
		t.Errorf("governance URL mismatch")
	}

	// Verify services
	if len(response.Data.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(response.Data.Services))
	}
	if response.Data.Services[0].ServiceID != githubServiceID.String() {
		t.Errorf("expected first service to be '%s', got %s", githubServiceID, response.Data.Services[0].ServiceID)
	}
	if len(response.Data.Services[0].RequiredScopes) != 2 {
		t.Errorf("expected 2 scopes for GitHub, got %d", len(response.Data.Services[0].RequiredScopes))
	}
}

func TestGetAgentDetail_AgentNotFound(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{}

	agentRepo := memory.NewAgentRepository()
	// Don't create any agents - agent should not be found

	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	// Create request with a valid UUID that doesn't exist
	nonexistentID := id.NewAgentID()
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+nonexistentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", nonexistentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add principal to context
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "not found" {
		t.Errorf("expected error 'not found', got %s", response.Error)
	}
}

func TestGetAgentDetail_MissingAgentID(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{}
	handler := NewAgentDetailHandler(mockService, nil)

	// Create request without agent ID
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/", nil)
	rctx := chi.NewRouteContext()
	// Don't add agentId parameter
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "bad request" {
		t.Errorf("expected error 'bad request', got %s", response.Error)
	}
}

func TestGetAgentDetail_ServiceError(t *testing.T) {
	mockService := &mockAgentDetailService{}

	agentRepo := &mockAgentRepository{
		getFunc: func(_ context.Context, _ id.AgentID) (*storage.Agent, error) {
			return nil, storage.NewStorageError("GetAgent", storage.ErrorKindConnection, errors.New("database connection failed"), "db unavailable")
		},
	}
	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	// Create request
	testAgentID := id.NewAgentID()
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+testAgentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add principal to context
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "internal server error" {
		t.Errorf("expected error 'internal server error', got %s", response.Error)
	}
}

func TestGetAgentDetail_ServiceRequirementSessionLookupError(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	mockService := &mockAgentDetailService{}

	agentRepo := memory.NewAgentRepository()
	require.NoError(t, agentRepo.Create(context.Background(), &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("test-client-id"),
		DisplayName: "Test Agent",
		Description: "Test agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       serviceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"repo"},
			},
		},
	}))

	sessionRepo := &mockSessionRepository{
		findByPrincipalAndServiceFunc: func(ctx context.Context, p id.Principal, svcID id.ServiceID) (*storage.UserSession, error) {
			return nil, storage.NewStorageError("FindByPrincipalAndService", storage.ErrorKindConnection, nil, "database unavailable")
		},
	}
	serviceRepo := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return &model.ThirdpartyOAuth2ProviderEntity{
				ID:          serviceID,
				DisplayName: "GitHub",
				Scopes: []model.OAuthScope{
					{ScopeValue: "repo", Description: "Repository access"},
				},
				Secret: model.NewEncryptedSecret(encryptSecretForTest(serviceID.String(), "test-client-secret")),
			}, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Equal(t, "internal server error", response.Error)
}

func TestGetAgentDetail_ServiceRequirementProviderLookupError(t *testing.T) {
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()

	mockService := &mockAgentDetailService{}

	agentRepo := memory.NewAgentRepository()
	require.NoError(t, agentRepo.Create(context.Background(), &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("test-client-id"),
		DisplayName: "Test Agent",
		Description: "Test agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       serviceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"repo"},
			},
		},
	}))

	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return nil, storage.NewStorageError("Get", storage.ErrorKindConnection, nil, "database unavailable")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Equal(t, "internal server error", response.Error)
}

func TestGetAgentDetail_EmptyServicesList(t *testing.T) {
	// Setup
	agentID := id.NewAgentID()

	mockService := &mockAgentDetailService{}

	ctx := context.Background()
	agentRepo := memory.NewAgentRepository()
	// Create agent with no service requirements
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.NewClientID("test-client-id"),
		DisplayName: "Test Agent",
		Description: "A test agent",
		// No service requirements
	}
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Add principal to context
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify empty services list
	if len(response.Data.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(response.Data.Services))
	}
}


// mockSessionRepository is a mock implementation of UserSessionRepository for testing.
type mockSessionRepository struct {
	findByPrincipalAndServiceFunc func(ctx context.Context, p id.Principal, serviceID id.ServiceID) (*storage.UserSession, error)
}

func (m *mockSessionRepository) FindByPrincipalAndService(ctx context.Context, p id.Principal, serviceID id.ServiceID) (*storage.UserSession, error) {
	if m.findByPrincipalAndServiceFunc != nil {
		return m.findByPrincipalAndServiceFunc(ctx, p, serviceID)
	}
	return nil, nil
}

func (m *mockSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) Get(ctx context.Context, sessionID id.SessionID) (*storage.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) ListByPrincipal(ctx context.Context, p id.Principal) ([]*storage.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) Delete(ctx context.Context, sessionID id.SessionID) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) DeleteByPrincipalAndService(ctx context.Context, p id.Principal, serviceID id.ServiceID) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) CountByService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, errors.New("not implemented")
}

// mockServiceRepository is a mock implementation of ThirdpartyOAuth2ProviderRepository for testing.
type mockServiceRepository struct {
	getFunc func(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error)
}

func (m *mockServiceRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, serviceID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockServiceRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, errors.New("not implemented")
}

func (m *mockServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	return 0, errors.New("not implemented")
}

func (m *mockServiceRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, errors.New("not implemented")
}

type mockAgentRepository struct {
	getFunc func(ctx context.Context, agentID id.AgentID) (*storage.Agent, error)
}

func (m *mockAgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, agentID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	return errors.New("not implemented")
}

func (m *mockAgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	return errors.New("not implemented")
}

func (m *mockAgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	return errors.New("not implemented")
}

func (m *mockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentRepository) GetByClientURI(ctx context.Context, uri string) (*storage.Agent, error) {
	return nil, errors.New("not implemented")
}

// Tests for buildServiceRequirementsForUser method

func TestBuildServiceRequirementsForUser_NoRequirements(t *testing.T) {
	// Setup: agent with no service requirements
	agentWithoutReqs := &storage.Agent{
		ID:          id.NewAgentID(),
		DisplayName: "Test Agent",
		// ServiceRequirements is nil
	}

	mockService := &mockAgentDetailService{}
	mockSessions := &mockSessionRepository{}
	mockServices := &mockServiceRepository{}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	// Execute
	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, id.NewPrincipal("user@example.com"), agentWithoutReqs)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty array, got %d items", len(results))
	}
}

func TestBuildServiceRequirementsForUser_WithRequirementsUserConnected(t *testing.T) {
	// Setup: agent with mandatory GitHub requirement
	userPrincipal := id.NewPrincipal("user@example.com")
	githubServiceID := id.NewServiceID()

	agentWithReqs := &storage.Agent{
		ID:          id.NewAgentID(),
		DisplayName: "Test Agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       githubServiceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"read:user", "repo"},
			},
		},
	}

	mockService := &mockAgentDetailService{}

	mockSessions := &mockSessionRepository{
		findByPrincipalAndServiceFunc: func(ctx context.Context, p id.Principal, svcID id.ServiceID) (*storage.UserSession, error) {
			if p == userPrincipal && svcID == githubServiceID {
				return &storage.UserSession{
					ID:        id.NewSessionID(),
					Principal: userPrincipal,
					ServiceID: githubServiceID,
					Scope:     []string{"read:user", "repo"}, // Has required scopes
				}, nil
			}
			return nil, nil
		},
	}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			if svcID == githubServiceID {
				return &model.ThirdpartyOAuth2ProviderEntity{
					ID:          githubServiceID,
					DisplayName: "GitHub",
					Scopes: []model.OAuthScope{
						{ScopeValue: "read:user", Description: "Read user profile"},
						{ScopeValue: "repo", Description: "Full control of repositories"},
					},
					Secret: model.NewEncryptedSecret(encryptSecretForTest(githubServiceID.String(), "test-client-secret")),
				}, nil
			}
			return nil, errors.New("service not found")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	// Execute
	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, userPrincipal, agentWithReqs)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.ServiceID != githubServiceID.String() {
		t.Errorf("expected service ID %s, got %s", githubServiceID, result.ServiceID)
	}
	if result.ServiceName != "GitHub" {
		t.Errorf("expected service name 'GitHub', got %s", result.ServiceName)
	}
	if result.RequirementType != "mandatory" {
		t.Errorf("expected requirement type 'mandatory', got %s", result.RequirementType)
	}
	if result.ConnectionStatus != "connected" {
		t.Errorf("expected connection status 'connected', got %s", result.ConnectionStatus)
	}
	if len(result.RequiredScopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(result.RequiredScopes))
	}
	if result.RequiredScopes[0].Name != "read:user" {
		t.Errorf("expected first scope name 'read:user', got %s", result.RequiredScopes[0].Name)
	}
	if result.RequiredScopes[0].Description != "Read user profile" {
		t.Errorf("expected first scope description 'Read user profile', got %s", result.RequiredScopes[0].Description)
	}
}

func TestBuildServiceRequirementsForUser_WithRequirementsUserNotConnected(t *testing.T) {
	// Setup: agent with optional Google requirement, user not connected
	userPrincipal := id.NewPrincipal("user@example.com")
	googleServiceID := id.NewServiceID()

	agentWithReqs := &storage.Agent{
		ID:          id.NewAgentID(),
		DisplayName: "Test Agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       googleServiceID,
				RequirementType: storage.RequirementTypeOptional,
				RequiredScopes:  []string{"email"},
			},
		},
	}

	mockService := &mockAgentDetailService{}

	mockSessions := &mockSessionRepository{
		findByPrincipalAndServiceFunc: func(ctx context.Context, p id.Principal, svcID id.ServiceID) (*storage.UserSession, error) {
			return nil, nil // User has no session
		},
	}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			if svcID == googleServiceID {
				return &model.ThirdpartyOAuth2ProviderEntity{
					ID:          googleServiceID,
					DisplayName: "Google",
					Scopes: []model.OAuthScope{
						{ScopeValue: "email", Description: "View email address"},
					},
					Secret: model.NewEncryptedSecret(encryptSecretForTest(googleServiceID.String(), "test-client-secret")),
				}, nil
			}
			return nil, errors.New("service not found")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	// Execute
	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, userPrincipal, agentWithReqs)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.ConnectionStatus != "not_connected" {
		t.Errorf("expected connection status 'not_connected', got %s", result.ConnectionStatus)
	}
	if result.RequirementType != "optional" {
		t.Errorf("expected requirement type 'optional', got %s", result.RequirementType)
	}
}

func TestBuildServiceRequirementsForUser_ServiceNotFound(t *testing.T) {
	// Setup: agent requirement references non-existent service
	userPrincipal := id.NewPrincipal("user@example.com")
	nonexistentServiceID := id.NewServiceID()

	agentWithReqs := &storage.Agent{
		ID:          id.NewAgentID(),
		DisplayName: "Test Agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       nonexistentServiceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"scope1"},
			},
		},
	}

	mockService := &mockAgentDetailService{}

	mockSessions := &mockSessionRepository{}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return nil, ports.ErrNotFound
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	// Execute
	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, userPrincipal, agentWithReqs)

	// Assert - should skip missing service (fail-open for fetch, per spec)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Missing service should be skipped from results
	if len(results) != 0 {
		t.Errorf("expected 0 results (missing service skipped), got %d", len(results))
	}
}

func TestBuildServiceRequirementsForUser_ExpiredSessionShowsNotConnected(t *testing.T) {
	// Expired sessions must surface as "not_connected" so the Login button is shown
	// on the consent screen and users can re-authenticate.
	userPrincipal := id.NewPrincipal("user@example.com")
	githubServiceID := id.NewServiceID()

	agentWithReqs := &storage.Agent{
		ID:          id.NewAgentID(),
		DisplayName: "Test Agent",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       githubServiceID,
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"read:user"},
			},
		},
	}

	mockService := &mockAgentDetailService{}

	expiredAt := time.Now().Add(-1 * time.Hour)
	mockSessions := &mockSessionRepository{
		findByPrincipalAndServiceFunc: func(ctx context.Context, p id.Principal, svcID id.ServiceID) (*storage.UserSession, error) {
			return &storage.UserSession{
				ID:                    id.NewSessionID(),
				Principal:             userPrincipal,
				ServiceID:             githubServiceID,
				TokenType:             "Bearer",
				RefreshTokenExpiresAt: &expiredAt, // fully expired
			}, nil
		},
	}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, svcID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return &model.ThirdpartyOAuth2ProviderEntity{
				ID:          githubServiceID,
				DisplayName: "GitHub",
				Scopes: []model.OAuthScope{
					{ScopeValue: "read:user", Description: "Read user profile"},
				},
				Secret: model.NewEncryptedSecret(encryptSecretForTest(githubServiceID.String(), "test-client-secret")),
			}, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, userPrincipal, agentWithReqs)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ConnectionStatus != "not_connected" {
		t.Errorf("expected connection status 'not_connected' for expired session, got %s", results[0].ConnectionStatus)
	}
}

func TestResolveCIMDMetadata_SessionAgentMismatch(t *testing.T) {
	agentA := id.NewAgentID()
	agentB := id.NewAgentID()
	principalID := "user@example.com"

	mockService := &mockAgentDetailService{}

	agentRepo := memory.NewAgentRepository()
	agentObjA := &storage.Agent{ID: agentA, ClientID: id.NewClientID("a"), DisplayName: "Agent A", Description: "Agent A desc"}
	agentObjB := &storage.Agent{ID: agentB, ClientID: id.NewClientID("b"), DisplayName: "Agent B", Description: "Agent B desc"}
	ctx := context.Background()
	if err := agentRepo.Create(ctx, agentObjA); err != nil {
		t.Fatalf("create agent A: %v", err)
	}
	if err := agentRepo.Create(ctx, agentObjB); err != nil {
		t.Fatalf("create agent B: %v", err)
	}

	authSessionRepo := memory.NewAuthorizationSessionRepository()
	session, err := storage.NewAuthorizationSession(
		agentA, id.Principal(principalID),
		"client-a", "https://example.com/original", "https://example.com/cb",
		"openid", "state123", "challenge", "S256", nil,
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := authSessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("persist session: %v", err)
	}

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(memory.NewInMemoryUserSessionRepository()).
		WithProviderService(newTestProviderService(memory.NewInMemoryThirdpartyOAuth2ProviderRepository())).
		WithAuthorizationSessionRepository(authSessionRepo)

	// Request agent B's detail with a session that belongs to agent A
	req := httptest.NewRequest(http.MethodGet,
		"/api/consent/agent/"+agentB.String()+"?session_id="+session.SessionID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentB.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), principalID))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for session-agent mismatch, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestResolveCIMDMetadata_SessionPrincipalMismatch(t *testing.T) {
	agentID := id.NewAgentID()

	mockService := &mockAgentDetailService{}

	agentRepo := memory.NewAgentRepository()
	agent := &storage.Agent{ID: agentID, ClientID: id.NewClientID("a"), DisplayName: "Agent", Description: "Agent desc"}
	ctx := context.Background()
	if err := agentRepo.Create(ctx, agent); err != nil {
		t.Fatalf("create agent: %v", err)
	}

	authSessionRepo := memory.NewAuthorizationSessionRepository()
	session, err := storage.NewAuthorizationSession(
		agentID, id.Principal("userA@example.com"),
		"client-a", "https://example.com/original", "https://example.com/cb",
		"openid", "state123", "challenge", "S256", nil,
	)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if err := authSessionRepo.Create(ctx, session); err != nil {
		t.Fatalf("persist session: %v", err)
	}

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(memory.NewInMemoryUserSessionRepository()).
		WithProviderService(newTestProviderService(memory.NewInMemoryThirdpartyOAuth2ProviderRepository())).
		WithAuthorizationSessionRepository(authSessionRepo)

	// User B tries to use user A's session
	req := httptest.NewRequest(http.MethodGet,
		"/api/consent/agent/"+agentID.String()+"?session_id="+session.SessionID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "userB@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for session-principal mismatch, got %d: %s", rr.Code, rr.Body.String())
	}
}
