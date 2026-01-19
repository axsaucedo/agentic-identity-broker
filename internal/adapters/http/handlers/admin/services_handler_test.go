package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// testConfig creates a minimal config for tests with HTTPS validation enabled (strict mode)
func testConfig() *ports.Config {
	return &ports.Config{
		Security: ports.SecurityConfig{
			SkipThirdpartyHTTPSValidation: false,
		},
	}
}

// MockServiceRepository is a mock implementation of ports.ThirdpartyOAuth2ServiceRepository
type MockServiceRepository struct {
	mock.Mock
}

func (m *MockServiceRepository) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockServiceRepository) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ThirdpartyOAuth2Service), args.Error(1)
}

func (m *MockServiceRepository) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockServiceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockServiceRepository) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.ThirdpartyOAuth2Service), args.Error(1)
}

func (m *MockServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	args := m.Called(ctx, serviceID)
	return args.Int(0), args.Error(1)
}

// MockBranchKeyRepository is a mock implementation of ports.BranchKeyManager
type MockBranchKeyRepository struct {
	mock.Mock
}

func (m *MockBranchKeyRepository) Create(ctx context.Context, serviceID string) (string, error) {
	args := m.Called(ctx, serviceID)
	return args.String(0), args.Error(1)
}

func (m *MockBranchKeyRepository) Get(ctx context.Context, serviceID string) (string, error) {
	args := m.Called(ctx, serviceID)
	return args.String(0), args.Error(1)
}

func (m *MockBranchKeyRepository) ResolveBranchKeyID(serviceID string) (string, error) {
	args := m.Called(serviceID)
	return args.String(0), args.Error(1)
}

func TestServicesHandler_CreateService(t *testing.T) {
	logger := slog.Default()

	t.Run("successful creation without discovery", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

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

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(s *storage.ThirdpartyOAuth2Service) bool {
			return s.DisplayName == "GitHub" && s.ClientID == "github-client-id" && !s.Discovery.EnableDiscovery
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
		assert.Equal(t, "REDACTED", resp.ClientSecret) // Secret should be redacted
		assert.False(t, resp.Discovery.EnableDiscovery)
		assert.Len(t, resp.Scopes, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

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

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		reqBody := ServiceRequest{
			DisplayName: "", // Invalid: empty display name
			ClientID:    "test-client",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateService", storage.ErrorKindValidation, nil, "validation failed"))

		req := httptest.NewRequest(http.MethodPost, "/api/third-party/oauth2/clients", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateService(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
	})
}

func TestServicesHandler_GetService(t *testing.T) {
	logger := slog.Default()

	t.Run("successful get", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		service := &storage.ThirdpartyOAuth2Service{
			ID:           "service-123",
			DisplayName:  "GitHub",
			ClientID:     "github-client-id",
			ClientSecret: "secret",
			IssuerURI:    "https://github.com",
			Discovery: storage.DiscoveryConfig{
				EnableDiscovery: false,
			},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://github.com/login/oauth/access_token",
				AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "repo", Description: "Repository access"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockRepo.On("Get", mock.Anything, "service-123").Return(service, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients/service-123", nil)
		w := httptest.NewRecorder()

		// Set up chi URL params
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "service-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.GetService(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp ServiceResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "service-123", resp.ID)
		assert.Equal(t, "GitHub", resp.DisplayName)
		assert.Equal(t, "REDACTED", resp.ClientSecret) // Secret should be redacted
	})

	t.Run("service not found", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		mockRepo.On("Get", mock.Anything, "nonexistent").Return(nil,
			storage.NewStorageError("GetService", storage.ErrorKindNotFound, nil, "service not found"))

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients/nonexistent", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "nonexistent")
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
	logger := slog.Default()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		existing := &storage.ThirdpartyOAuth2Service{
			ID:           "service-123",
			DisplayName:  "GitHub",
			ClientID:     "github-client-id",
			ClientSecret: "old-secret",
			IssuerURI:    "https://github.com",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now().Add(-24 * time.Hour),
		}

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

		mockRepo.On("Get", mock.Anything, "service-123").Return(existing, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(s *storage.ThirdpartyOAuth2Service) bool {
			return s.ID == "service-123" && s.DisplayName == "GitHub Updated"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/third-party/oauth2/clients/service-123", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "service-123")
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
}

func TestServicesHandler_DeleteService(t *testing.T) {
	logger := slog.Default()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		mockRepo.On("Delete", mock.Anything, "service-123").Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/service-123", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "service-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.DeleteService(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion blocked by grants", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		mockRepo.On("Delete", mock.Anything, "service-123").Return(
			storage.NewStorageError("DeleteService", storage.ErrorKindConflict, nil, "cannot delete service: 5 grants reference it"))

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/service-123", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "service-123")
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
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		mockRepo.On("Delete", mock.Anything, "nonexistent").Return(
			storage.NewStorageError("DeleteService", storage.ErrorKindNotFound, nil, "service not found"))

		req := httptest.NewRequest(http.MethodDelete, "/api/third-party/oauth2/clients/nonexistent", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "nonexistent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		handler.DeleteService(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestServicesHandler_ListServices(t *testing.T) {
	logger := slog.Default()

	t.Run("successful list", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		services := []*storage.ThirdpartyOAuth2Service{
			{
				ID:           "service-1",
				DisplayName:  "GitHub",
				ClientID:     "github-client",
				ClientSecret: "secret1",
				IssuerURI:    "https://github.com",
				Scopes: []storage.OAuthScope{
					{ScopeValue: "repo", Description: "Repository access"},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:           "service-2",
				DisplayName:  "Google",
				ClientID:     "google-client",
				ClientSecret: "secret2",
				IssuerURI:    "https://accounts.google.com",
				Scopes: []storage.OAuthScope{
					{ScopeValue: "email", Description: "Email access"},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		mockRepo.On("List", mock.Anything).Return(services, nil)

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
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		mockRepo.On("List", mock.Anything).Return([]*storage.ThirdpartyOAuth2Service{}, nil)

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
	logger := slog.Default()

	t.Run("secret redacted in all responses", func(t *testing.T) {
		mockRepo := new(MockServiceRepository)
		mockBranchKeyRepo := new(MockBranchKeyRepository)
		mockBranchKeyRepo.On("Create", mock.Anything, mock.Anything).Return("service_test_branch_key", nil)
		handler := NewServicesHandler(mockRepo, mockBranchKeyRepo, testConfig(), logger)

		service := &storage.ThirdpartyOAuth2Service{
			ID:           "service-123",
			DisplayName:  "Test Service",
			ClientID:     "test-client",
			ClientSecret: "super-secret-value",
			IssuerURI:    "https://example.com",
			Scopes: []storage.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Test Get endpoint
		mockRepo.On("Get", mock.Anything, "service-123").Return(service, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/third-party/oauth2/clients/service-123", nil)
		w := httptest.NewRecorder()

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("client-id", "service-123")
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
