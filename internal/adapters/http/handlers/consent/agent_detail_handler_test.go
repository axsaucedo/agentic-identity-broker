package consent

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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

// mockAgentDetailService is a configurable mock implementation of ConsentService for testing.
type mockAgentDetailService struct {
	getAgentWithServiceRequirementsFunc func(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error)
}

func (m *mockAgentDetailService) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentWithServiceRequirements(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
	if m.getAgentWithServiceRequirementsFunc != nil {
		return m.getAgentWithServiceRequirementsFunc(ctx, userPrincipal, agentID)
	}
	return nil, nil, errors.New("not implemented")
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
	agentID := id.NewAgentID()
	principalID := "user@example.com"
	governanceURL := "https://example.com/governance"
	userDocsURL := "https://example.com/docs"
	agentInterfaceURL := "https://example.com/interface"
	githubServiceID := id.NewServiceID()
	googleServiceID := id.NewServiceID()

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return &storage.Agent{
					ID:                   agentID,
					ClientID:             id.NewClientID("test-client-id"),
					DisplayName:          "Test Agent",
					Description:          "A test agent for testing purposes",
					GovernanceURL:        &governanceURL,
					UserDocumentationURL: &userDocsURL,
					AgentInterfaceURL:    &agentInterfaceURL,
				},
				[]consent.ServiceRequirementStatus{
					{
						ServiceID:       githubServiceID,
						DisplayName:     "GitHub",
						RequirementType: storage.RequirementTypeMandatory,
						RequiredScopes: []consent.ServiceScopeInfo{
							{Name: "read:user", Description: "Read user profile"},
							{Name: "repo", Description: "Full control of repositories"},
						},
						IsConnected: false,
					},
					{
						ServiceID:       googleServiceID,
						DisplayName:     "Google",
						RequirementType: storage.RequirementTypeOptional,
						RequiredScopes:  []consent.ServiceScopeInfo{{Name: "email", Description: "View email address"}},
						IsConnected:     false,
					},
				}, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), principalID))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusOK, rr.Code, rr.Body.String())

	var response GetAgentDetailResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))

	assert.Equal(t, agentID, response.Data.Agent.AgentID)
	assert.Equal(t, "Test Agent", response.Data.Agent.DisplayName)
	require.NotNil(t, response.Data.Agent.GovernanceURL)
	assert.Equal(t, governanceURL, *response.Data.Agent.GovernanceURL)

	require.Len(t, response.Data.Services, 2)
	assert.Equal(t, githubServiceID.String(), response.Data.Services[0].ServiceID)
	assert.Len(t, response.Data.Services[0].RequiredScopes, 2)
}

func TestGetAgentDetail_AgentNotFound(t *testing.T) {
	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return nil, nil, consent.ErrAgentNotFound
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	nonexistentID := id.NewAgentID()
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+nonexistentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", nonexistentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Equal(t, "not found", response.Error)
}

func TestGetAgentDetail_MissingAgentID(t *testing.T) {
	handler := NewAgentDetailHandler(&mockAgentDetailService{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/", nil)
	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Equal(t, "bad request", response.Error)
}

func TestGetAgentDetail_ServiceError(t *testing.T) {
	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return nil, nil, storage.NewStorageError("GetAgent", storage.ErrorKindConnection, errors.New("database connection failed"), "db unavailable")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	testAgentID := id.NewAgentID()
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+testAgentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)

	var response ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Equal(t, "internal server error", response.Error)
}

func TestGetAgentDetail_ServiceRequirementError(t *testing.T) {
	agentID := id.NewAgentID()

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return nil, nil, storage.NewStorageError("FindByPrincipalAndService", storage.ErrorKindConnection, nil, "database unavailable")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

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
	agentID := id.NewAgentID()

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return &storage.Agent{
				ID:          agentID,
				ClientID:    id.NewClientID("test-client-id"),
				DisplayName: "Test Agent",
				Description: "A test agent",
			}, []consent.ServiceRequirementStatus{}, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response GetAgentDetailResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	assert.Len(t, response.Data.Services, 0)
}

func TestResolveCIMDMetadata_SessionAgentMismatch(t *testing.T) {
	agentA := id.NewAgentID()
	agentB := id.NewAgentID()
	principalID := "user@example.com"

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return &storage.Agent{ID: agentID, ClientID: id.NewClientID("test"), DisplayName: "Agent"}, []consent.ServiceRequirementStatus{}, nil
		},
	}

	authSessionRepo := memory.NewAuthorizationSessionRepository()
	session, err := storage.NewAuthorizationSession(
		agentA, id.Principal(principalID),
		"client-a", "https://example.com/original", "https://example.com/cb",
		"openid", "state123", "challenge", "S256", nil,
	)
	require.NoError(t, err)
	require.NoError(t, authSessionRepo.Create(context.Background(), session))

	handler := NewAgentDetailHandler(mockService, nil).
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

	require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
}

func TestResolveCIMDMetadata_SessionPrincipalMismatch(t *testing.T) {
	agentID := id.NewAgentID()

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, aID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return &storage.Agent{ID: aID, ClientID: id.NewClientID("test"), DisplayName: "Agent"}, []consent.ServiceRequirementStatus{}, nil
		},
	}

	authSessionRepo := memory.NewAuthorizationSessionRepository()
	session, err := storage.NewAuthorizationSession(
		agentID, id.Principal("userA@example.com"),
		"client-a", "https://example.com/original", "https://example.com/cb",
		"openid", "state123", "challenge", "S256", nil,
	)
	require.NoError(t, err)
	require.NoError(t, authSessionRepo.Create(context.Background(), session))

	handler := NewAgentDetailHandler(mockService, nil).
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

	require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
}

func TestGetAgentDetail_SortsMandatoryFirst(t *testing.T) {
	agentID := id.NewAgentID()
	optionalServiceID := id.NewServiceID()
	mandatoryServiceID := id.NewServiceID()

	mockService := &mockAgentDetailService{
		getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
			return &storage.Agent{ID: agentID, ClientID: id.NewClientID("c"), DisplayName: "Agent"},
				[]consent.ServiceRequirementStatus{
					{ServiceID: optionalServiceID, DisplayName: "Optional", RequirementType: storage.RequirementTypeOptional},
					{ServiceID: mandatoryServiceID, DisplayName: "Mandatory", RequirementType: storage.RequirementTypeMandatory},
				}, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))

	rr := httptest.NewRecorder()
	handler.GetAgentDetail(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response GetAgentDetailResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&response))
	require.Len(t, response.Data.Services, 2)
	assert.Equal(t, "mandatory", response.Data.Services[0].RequirementType)
	assert.Equal(t, "optional", response.Data.Services[1].RequirementType)
}

func TestGetAgentDetail_ConnectionStatus(t *testing.T) {
	agentID := id.NewAgentID()
	svcID := id.NewServiceID()

	t.Run("connected", func(t *testing.T) {
		mockService := &mockAgentDetailService{
			getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
				return &storage.Agent{ID: agentID, ClientID: id.NewClientID("c"), DisplayName: "A"},
					[]consent.ServiceRequirementStatus{{ServiceID: svcID, IsConnected: true, RequirementType: storage.RequirementTypeMandatory}}, nil
			},
		}
		handler := NewAgentDetailHandler(mockService, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(principal.WithPrincipal(req.Context(), "u@example.com"))
		rr := httptest.NewRecorder()
		handler.GetAgentDetail(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var resp GetAgentDetailResponse
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
		require.Len(t, resp.Data.Services, 1)
		assert.Equal(t, "connected", resp.Data.Services[0].ConnectionStatus)
	})

	t.Run("not_connected", func(t *testing.T) {
		mockService := &mockAgentDetailService{
			getAgentWithServiceRequirementsFunc: func(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
				return &storage.Agent{ID: agentID, ClientID: id.NewClientID("c"), DisplayName: "A"},
					[]consent.ServiceRequirementStatus{{ServiceID: svcID, IsConnected: false, RequirementType: storage.RequirementTypeMandatory}}, nil
			},
		}
		handler := NewAgentDetailHandler(mockService, nil)
		req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(principal.WithPrincipal(req.Context(), "u@example.com"))
		rr := httptest.NewRecorder()
		handler.GetAgentDetail(rr, req)
		require.Equal(t, http.StatusOK, rr.Code)
		var resp GetAgentDetailResponse
		require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
		require.Len(t, resp.Data.Services, 1)
		assert.Equal(t, "not_connected", resp.Data.Services[0].ConnectionStatus)
	})
}

