package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/testutil"
)

// testConfig creates a minimal config for tests with HTTPS validation enabled (strict mode).
func testConfig() *ports.Config {
	return &ports.Config{
		Security: ports.SecurityConfig{
			SkipThirdpartyHTTPSValidation: false,
		},
	}
}

// MockProviderRepository is a mock implementation of ports.ThirdpartyOAuth2ProviderRepository.
type MockProviderRepository struct {
	mock.Mock
}

func (m *MockProviderRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockProviderRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockProviderRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockProviderRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	args := m.Called(ctx, serviceID)
	return args.Error(0)
}

func (m *MockProviderRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockProviderRepository) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	args := m.Called(ctx, serviceID)
	return args.Int(0), args.Error(1)
}

func (m *MockProviderRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx, resourceURI)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

// newTestEncryption returns a real encryption adapter backed by the shared deterministic test key.
func newTestEncryption() ports.EncryptionPort {
	return testutil.NewPanicTestEncryptionAdapter()
}

// setupHandler creates a handler backed by a mock repository and test encryption.
func setupHandler(t *testing.T, mockRepo *MockProviderRepository) *ServicesHandler {
	t.Helper()
	svc := thirdparty.NewThirdpartyOAuth2ProviderService(mockRepo, newTestEncryption(), nil, false, slog.Default())
	return NewServicesHandler(svc, testConfig(), slog.Default())
}

// encryptSecretForTest encrypts a plaintext secret using the test encryption adapter.
// The service_id is used as the encryption context.
func encryptSecretForTest(serviceID, secret string) []byte {
	enc := newTestEncryption()
	ciphertext, err := enc.Encrypt(context.Background(), []byte(secret), map[string]string{"service_id": serviceID})
	if err != nil {
		panic("encryptSecretForTest: " + err.Error())
	}
	return ciphertext
}

// encryptedEntity creates an entity with encrypted secret (as it would come from the repository).
// The secret is properly encrypted with the test key so the domain service can decrypt it.
func encryptedEntity(serviceID id.ServiceID, displayName string, clientID id.ClientID, clientSecret, issuerURI string, scopes []model.OAuthScope) *model.ThirdpartyOAuth2ProviderEntity {
	now := time.Now()
	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          serviceID,
		DisplayName: displayName,
		ClientID:    clientID,
		Secret:      model.NewEncryptedSecret(encryptSecretForTest(serviceID.String(), clientSecret)),
		IssuerURI:   issuerURI,
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://" + strings.TrimPrefix(issuerURI, "https://") + "/token",
			AuthorizeEndpoint: "https://" + strings.TrimPrefix(issuerURI, "https://") + "/authorize",
		},
		Scopes:    scopes,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestServicesHandler_CreateService(t *testing.T) {
	t.Run("successful creation without discovery", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		reqBody := ServiceRequest{
			DisplayName:  "GitHub",
			ClientID:     "github-client-id",
			ClientSecret: "github-client-secret",
			IssuerURI:    "https://github.com",
			Discovery: DiscoveryConfigRequest{
				EnableDiscovery: false,
			},
			Endpoints: &OAuth2EndpointsRequest{
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			},
			Scopes: []OAuthScopeRequest{
				{ScopeValue: "repo", Description: "Repository access"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// The domain service encrypts and then calls repo.Create with encrypted entity.
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
			return e.DisplayName == "GitHub" && e.ClientID == "github-client-id" && e.Secret.IsEncrypted()
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/third-party/oauth2/clients", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateService(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "GitHub", resp.DisplayName)
		assert.Equal(t, "github-client-id", resp.ClientID)
		assert.Equal(t, "REDACTED", resp.ClientSecret) // Secret must be redacted
		assert.False(t, resp.Discovery.EnableDiscovery)
		assert.Len(t, resp.Scopes, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		req := httptest.NewRequest(http.MethodPost, "/api/third-party/oauth2/clients", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateService(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid request body", resp.Error)
	})

	t.Run("validation error - missing display name", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		reqBody := ServiceRequest{
			DisplayName: "", // Invalid: empty display name
			ClientID:    "test-client",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/third-party/oauth2/clients", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateService(w, req)

		// Validation happens before any repo call
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, "display_name is required")

		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestServicesHandler_GetService(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID := id.NewServiceID()
		entity := encryptedEntity(serviceID, "GitHub", "github-client-id", "secret", "https://github.com",
			[]model.OAuthScope{{ScopeValue: "repo", Description: "Repository access"}})
		mockRepo.On("Get", mock.Anything, serviceID).Return(entity, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients/"+serviceID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetService(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, serviceID.String(), resp.ID)
		assert.Equal(t, "GitHub", resp.DisplayName)
		assert.Equal(t, "REDACTED", resp.ClientSecret) // Secret must be redacted
	})

	t.Run("service not found", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		notFoundID := id.NewServiceID()
		mockRepo.On("Get", mock.Anything, notFoundID).Return(nil,
			storage.NewStorageError("GetService", storage.ErrorKindNotFound, nil, "service not found"))

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients/"+notFoundID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", notFoundID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetService(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "service not found", resp.Error)
	})
}

func TestServicesHandler_UpdateService(t *testing.T) {
	t.Run("successful update with new secret", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID := id.NewServiceID()
		reqBody := ServiceRequest{
			DisplayName:  "GitHub Updated",
			ClientID:     "github-client-id-new",
			ClientSecret: "new-secret",
			IssuerURI:    "https://github.com",
			Discovery: DiscoveryConfigRequest{
				EnableDiscovery: false,
			},
			Endpoints: &OAuth2EndpointsRequest{
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			},
			Scopes: []OAuthScopeRequest{
				{ScopeValue: "repo", Description: "Repository access"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
			return e.ID == serviceID && e.DisplayName == "GitHub Updated" && e.Secret.IsEncrypted()
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/third-party/oauth2/clients/"+serviceID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.UpdateService(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "GitHub Updated", resp.DisplayName)
		assert.Equal(t, "REDACTED", resp.ClientSecret)

		mockRepo.AssertExpectations(t)
	})

	t.Run("missing client_secret returns 400", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID := id.NewServiceID()
		reqBody := ServiceRequest{
			DisplayName:  "GitHub Updated",
			ClientID:     "github-client-id",
			ClientSecret: "", // Missing — must be rejected per OpenAPI contract
			IssuerURI:    "https://github.com",
			Discovery:    DiscoveryConfigRequest{EnableDiscovery: false},
			Scopes:       []OAuthScopeRequest{{ScopeValue: "repo", Description: "Repository access"}},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/third-party/oauth2/clients/"+serviceID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.UpdateService(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, "client_secret is required")

		// No repo calls must have been made
		mockRepo.AssertNotCalled(t, "Get")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("service not found returns 404", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		nonexistentID := id.NewServiceID()
		reqBody := ServiceRequest{
			DisplayName:  "GitHub Updated",
			ClientID:     "github-client-id",
			ClientSecret: "new-secret",
			IssuerURI:    "https://github.com",
			Discovery:    DiscoveryConfigRequest{EnableDiscovery: false},
			Endpoints: &OAuth2EndpointsRequest{
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			},
			Scopes: []OAuthScopeRequest{
				{ScopeValue: "repo", Description: "Repository access"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
			return e.ID == nonexistentID && e.Secret.IsEncrypted()
		})).Return(storage.NewStorageError("UpdateService", storage.ErrorKindNotFound, nil, "service not found"))

		req := httptest.NewRequest(http.MethodPut, "/api/third-party/oauth2/clients/"+nonexistentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", nonexistentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.UpdateService(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "service not found", resp.Error)

		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error after valid client_secret - no Update called", func(t *testing.T) {
		// Regression guard for KMS ordering: ValidateForUpdate must fire before
		// providerService.Update (which triggers encryption). If Update is called
		// despite a validation failure, the ordering is broken.
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		reqBody := ServiceRequest{
			DisplayName:  "", // Invalid: triggers ValidateForUpdate error
			ClientID:     "github-client-id",
			ClientSecret: "valid-secret", // Valid: passes the early client_secret guard
			IssuerURI:    "https://github.com",
			Discovery:    DiscoveryConfigRequest{EnableDiscovery: false},
			Endpoints: &OAuth2EndpointsRequest{
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			},
			Scopes: []OAuthScopeRequest{
				{ScopeValue: "repo", Description: "Repository access"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		serviceID := id.NewServiceID()
		req := httptest.NewRequest(http.MethodPut, "/api/third-party/oauth2/clients/"+serviceID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.UpdateService(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, "display_name is required")

		// Validation must fire before any storage or encryption call.
		mockRepo.AssertNotCalled(t, "Update")
	})
}

func TestServicesHandler_DeleteService(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID := id.NewServiceID()
		mockRepo.On("Delete", mock.Anything, serviceID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/"+serviceID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.DeleteService(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion blocked by grants", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID := id.NewServiceID()
		mockRepo.On("Delete", mock.Anything, serviceID).Return(
			storage.NewStorageError("DeleteService", storage.ErrorKindConflict, nil, "cannot delete service: 5 grants reference it"))

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/"+serviceID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", serviceID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.DeleteService(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "conflict", resp.Error)
		assert.Contains(t, resp.Message, "5 grants reference it")

		mockRepo.AssertExpectations(t)
	})

	t.Run("service not found", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		notFoundID := id.NewServiceID()
		mockRepo.On("Delete", mock.Anything, notFoundID).Return(
			storage.NewStorageError("DeleteService", storage.ErrorKindNotFound, nil, "service not found"))

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/"+notFoundID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", notFoundID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.DeleteService(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestServicesHandler_ListServices(t *testing.T) {
	t.Run("successful list", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		serviceID1 := id.NewServiceID()
		serviceID2 := id.NewServiceID()
		entities := []*model.ThirdpartyOAuth2ProviderEntity{
			encryptedEntity(serviceID1, "GitHub", "github-client", "secret1", "https://github.com",
				[]model.OAuthScope{{ScopeValue: "repo", Description: "Repository access"}}),
			encryptedEntity(serviceID2, "Google", "google-client", "secret2", "https://accounts.google.com",
				[]model.OAuthScope{{ScopeValue: "email", Description: "Email access"}}),
		}

		mockRepo.On("List", mock.Anything).Return(entities, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients", nil)
		w := httptest.NewRecorder()

		handler.ListServices(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "GitHub", resp[0].DisplayName)
		assert.Equal(t, "Google", resp[1].DisplayName)
		assert.Equal(t, "REDACTED", resp[0].ClientSecret)
		assert.Equal(t, "REDACTED", resp[1].ClientSecret)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		mockRepo.On("List", mock.Anything).Return([]*model.ThirdpartyOAuth2ProviderEntity{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients", nil)
		w := httptest.NewRecorder()

		handler.ListServices(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 0)

		mockRepo.AssertExpectations(t)
	})
}

func TestServicesHandler_SecretRedaction(t *testing.T) {
	t.Run("secret redacted in all responses", func(t *testing.T) {
		mockRepo := new(MockProviderRepository)
		handler := setupHandler(t, mockRepo)

		svcID := id.NewServiceID()
		entity := encryptedEntity(svcID, "Test Service", id.ClientID("test-client"), "super-secret-value", "https://example.com",
			[]model.OAuthScope{{ScopeValue: "read", Description: "Read access"}})

		// Test Get endpoint
		mockRepo.On("Get", mock.Anything, svcID).Return(entity, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/services/"+svcID.String(), nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("service-id", svcID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetService(w, req)

		var resp ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		// Verify secret is redacted, not the actual value
		assert.Equal(t, "REDACTED", resp.ClientSecret)
		assert.NotEqual(t, "super-secret-value", resp.ClientSecret)
	})
}
