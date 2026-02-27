package consent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// newTestEncryption creates a real encryption adapter using a deterministic test key.
// Panics on error since the key is hardcoded and always valid.
func newTestEncryption() ports.EncryptionPort {
	adapter, _, err := awsencryption.NewAWSEncryption("ASNFZ4mrze/+3LqYdlQyEAEjRWeJq83v/ty6mHZUMhA=", "", 0)
	if err != nil {
		panic("newTestEncryption: failed to create test encryption adapter: " + err.Error())
	}
	return adapter
}

// newTestProviderService wraps a ThirdpartyOAuth2ProviderRepository in a domain service
// with test encryption. Used in tests across the consent handler package.
func newTestProviderService(repo ports.ThirdpartyOAuth2ProviderRepository) *thirdparty.ThirdpartyOAuth2ProviderService {
	return thirdparty.NewThirdpartyOAuth2ProviderService(repo, newTestEncryption(), nil, slog.Default())
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
type mockAgentDetailService struct {
	getAgentDetailFunc func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
}

func (m *mockAgentDetailService) GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
	if m.getAgentDetailFunc != nil {
		return m.getAgentDetailFunc(ctx, agentID)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentConsentInfo(ctx context.Context, agentID string) (*consent.AgentConsentInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) RevokeConsent(ctx context.Context, principal, agentID string) error {
	return errors.New("not implemented")
}

func (m *mockAgentDetailService) GetActiveGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentDelegations(ctx context.Context, principal string) ([]consent.AgentDelegation, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetUserGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func TestGetAgentDetail_Success(t *testing.T) {
	// Setup
	agentID := "agent-123"
	principalID := "user@example.com"
	governanceURL := "https://example.com/governance"
	userDocsURL := "https://example.com/docs"
	agentInterfaceURL := "https://example.com/interface"

	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			if agID != agentID {
				t.Errorf("expected agentID %s, got %s", agentID, agID)
			}

			agentDetail := &consent.AgentDetail{
				AgentID:              agentID,
				DisplayName:          "Test Agent",
				Description:          "A test agent for testing purposes",
				LogoURL:              nil,
				GovernanceURL:        &governanceURL,
				UserDocumentationURL: &userDocsURL,
				AgentInterfaceURL:    &agentInterfaceURL,
			}

			services := []consent.ThirdpartyService{
				{
					ServiceID:   "github",
					DisplayName: "GitHub",
					LogoURL:     nil,
					Scopes: []consent.ServiceScope{
						{Value: "read:user", Description: "Read user profile"},
						{Value: "repo", Description: "Full control of repositories"},
					},
				},
				{
					ServiceID:   "google",
					DisplayName: "Google",
					LogoURL:     nil,
					Scopes: []consent.ServiceScope{
						{Value: "email", Description: "View email address"},
					},
				},
			}

			return agentDetail, services, nil
		},
	}

	// Create agent with service requirements
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "test-client-id",
		DisplayName: "Test Agent",
		Description: "A test agent for testing purposes",
		ServiceRequirements: []storage.ServiceRequirement{
			{
				ServiceID:       "github",
				RequirementType: storage.RequirementTypeMandatory,
				RequiredScopes:  []string{"read:user", "repo"},
			},
			{
				ServiceID:       "google",
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
		ID:          "github",
		DisplayName: "GitHub",
		ClientID:    "github-client-id",
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
		ID:          "google",
		DisplayName: "Google",
		ClientID:    "google-client-id",
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
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)

	// Add chi route context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
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
	if response.Data.Services[0].ServiceID != "github" {
		t.Errorf("expected first service to be 'github', got %s", response.Data.Services[0].ServiceID)
	}
	if len(response.Data.Services[0].RequiredScopes) != 2 {
		t.Errorf("expected 2 scopes for GitHub, got %d", len(response.Data.Services[0].RequiredScopes))
	}
}

func TestGetAgentDetail_AgentNotFound(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			return nil, nil, consent.ErrAgentNotFound
		},
	}

	agentRepo := memory.NewAgentRepository()
	// Don't create any agents - agent should not be found

	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "nonexistent")
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
	// Setup
	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			return nil, nil, errors.New("database connection failed")
		},
	}

	agentRepo := memory.NewAgentRepository()
	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewInMemoryThirdpartyOAuth2ProviderRepository()

	handler := NewAgentDetailHandler(mockService, nil).
		WithAgentRepository(agentRepo).
		WithSessionRepository(sessionRepo).
		WithProviderService(newTestProviderService(serviceRepo))

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
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

func TestGetAgentDetail_EmptyServicesList(t *testing.T) {
	// Setup
	agentID := "agent-456"

	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			agentDetail := &consent.AgentDetail{
				AgentID:     agentID,
				DisplayName: "Test Agent",
				Description: "A test agent",
			}

			// Return empty services list
			services := []consent.ThirdpartyService{}

			return agentDetail, services, nil
		},
	}

	ctx := context.Background()
	agentRepo := memory.NewAgentRepository()
	// Create agent with no service requirements
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "test-client-id",
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
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID)
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

// Ensure mockAgentDetailService implements the required interface methods
var _ interface {
	GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
} = (*mockAgentDetailService)(nil)

// mockSessionRepository is a mock implementation of UserSessionRepository for testing.
type mockSessionRepository struct {
	findByPrincipalAndServiceFunc func(ctx context.Context, principal, serviceID string) (*storage.UserSession, error)
}

func (m *mockSessionRepository) FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error) {
	if m.findByPrincipalAndServiceFunc != nil {
		return m.findByPrincipalAndServiceFunc(ctx, principal, serviceID)
	}
	return nil, nil
}

func (m *mockSessionRepository) Create(ctx context.Context, session *storage.UserSession) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) Get(ctx context.Context, id string) (*storage.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error) {
	return nil, errors.New("not implemented")
}

func (m *mockSessionRepository) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error {
	return errors.New("not implemented")
}

func (m *mockSessionRepository) CountByService(ctx context.Context, serviceID string) (int, error) {
	return 0, errors.New("not implemented")
}

// mockServiceRepository is a mock implementation of ThirdpartyOAuth2ProviderRepository for testing.
type mockServiceRepository struct {
	getFunc func(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error)
}

func (m *mockServiceRepository) Get(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockServiceRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockServiceRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, errors.New("not implemented")
}

func (m *mockServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	return 0, errors.New("not implemented")
}

func (m *mockServiceRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return nil, errors.New("not implemented")
}

// Tests for buildServiceRequirementsForUser method

func TestBuildServiceRequirementsForUser_NoRequirements(t *testing.T) {
	// Setup: agent with no service requirements
	agentWithoutReqs := &storage.Agent{
		ID:          "agent-123",
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
	results, err := handler.buildServiceRequirementsForUser(ctx, "user@example.com", agentWithoutReqs)

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
	userID := "user@example.com"
	githubServiceID := "github-service-uuid"

	agentWithReqs := &storage.Agent{
		ID:          "agent-123",
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
		findByPrincipalAndServiceFunc: func(ctx context.Context, principal, serviceID string) (*storage.UserSession, error) {
			if principal == userID && serviceID == githubServiceID {
				return &storage.UserSession{
					ID:        "session-uuid",
					Principal: userID,
					ServiceID: githubServiceID,
					Scope:     []string{"read:user", "repo"}, // Has required scopes
				}, nil
			}
			return nil, nil
		},
	}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			if id == githubServiceID {
				return &model.ThirdpartyOAuth2ProviderEntity{
					ID:          githubServiceID,
					DisplayName: "GitHub",
					Scopes: []model.OAuthScope{
						{ScopeValue: "read:user", Description: "Read user profile"},
						{ScopeValue: "repo", Description: "Full control of repositories"},
					},
					Secret: model.NewEncryptedSecret(encryptSecretForTest(githubServiceID, "test-client-secret")),
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
	results, err := handler.buildServiceRequirementsForUser(ctx, userID, agentWithReqs)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	result := results[0]
	if result.ServiceID != githubServiceID {
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
	userID := "user@example.com"
	googleServiceID := "google-service-uuid"

	agentWithReqs := &storage.Agent{
		ID:          "agent-123",
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
		findByPrincipalAndServiceFunc: func(ctx context.Context, principal, serviceID string) (*storage.UserSession, error) {
			return nil, nil // User has no session
		},
	}

	mockServices := &mockServiceRepository{
		getFunc: func(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			if id == googleServiceID {
				return &model.ThirdpartyOAuth2ProviderEntity{
					ID:          googleServiceID,
					DisplayName: "Google",
					Scopes: []model.OAuthScope{
						{ScopeValue: "email", Description: "View email address"},
					},
					Secret: model.NewEncryptedSecret(encryptSecretForTest(googleServiceID, "test-client-secret")),
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
	results, err := handler.buildServiceRequirementsForUser(ctx, userID, agentWithReqs)

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
	userID := "user@example.com"
	nonexistentServiceID := "nonexistent-uuid"

	agentWithReqs := &storage.Agent{
		ID:          "agent-123",
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
		getFunc: func(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return nil, errors.New("service not found")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)
	handler.sessionRepository = mockSessions
	handler.providerService = newTestProviderService(mockServices)

	// Execute
	ctx := context.Background()
	results, err := handler.buildServiceRequirementsForUser(ctx, userID, agentWithReqs)

	// Assert - should skip missing service (fail-open for fetch, per spec)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Missing service should be skipped from results
	if len(results) != 0 {
		t.Errorf("expected 0 results (missing service skipped), got %d", len(results))
	}
}
