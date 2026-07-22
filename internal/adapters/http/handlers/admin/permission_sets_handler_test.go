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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/permissionset"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockPermissionSetRepository is a mock implementation of ports.PermissionSetRepository
type MockPermissionSetRepository struct {
	mock.Mock
}

func (m *MockPermissionSetRepository) Create(ctx context.Context, ps *storage.PermissionSet) error {
	args := m.Called(ctx, ps)
	return args.Error(0)
}

func (m *MockPermissionSetRepository) Get(ctx context.Context, id id.PermissionSetID) (*storage.PermissionSet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.PermissionSet), args.Error(1)
}

func (m *MockPermissionSetRepository) GetByIDs(ctx context.Context, ids []id.PermissionSetID) ([]*storage.PermissionSet, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.PermissionSet), args.Error(1)
}

func (m *MockPermissionSetRepository) Update(ctx context.Context, ps *storage.PermissionSet) error {
	args := m.Called(ctx, ps)
	return args.Error(0)
}

func (m *MockPermissionSetRepository) Delete(ctx context.Context, id id.PermissionSetID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPermissionSetRepository) List(ctx context.Context, serviceID id.ServiceID) ([]*storage.PermissionSet, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.PermissionSet), args.Error(1)
}

func (m *MockPermissionSetRepository) CountAgentsReferencingPermissionSet(ctx context.Context, id id.PermissionSetID) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *MockPermissionSetRepository) CountGrantsReferencingPermissionSet(ctx context.Context, id id.PermissionSetID) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *MockPermissionSetRepository) CountPermissionSetsForService(ctx context.Context, serviceID id.ServiceID) (int, error) {
	args := m.Called(ctx, serviceID)
	return args.Int(0), args.Error(1)
}

// nopGrantRepo satisfies ports.UserGrantRepository with no-op implementations.
// Used in tests where grant-based deletion protection is not under test.
type nopGrantRepo struct {
	countGrantsReferencingFunc func(ctx context.Context, id id.PermissionSetID) (int, error)
}

func (n *nopGrantRepo) Create(_ context.Context, _ *storage.UserGrant) error { return nil }
func (n *nopGrantRepo) Get(_ context.Context, _ id.GrantID) (*storage.UserGrant, error) {
	return nil, nil
}
func (n *nopGrantRepo) Update(_ context.Context, _ *storage.UserGrant) error { return nil }
func (n *nopGrantRepo) Delete(_ context.Context, _ id.GrantID) error         { return nil }
func (n *nopGrantRepo) ListByPrincipalAndAgent(_ context.Context, _ id.Principal, _ id.AgentID) ([]*storage.UserGrant, error) {
	return nil, nil
}
func (n *nopGrantRepo) FindByPrincipalAndAgent(_ context.Context, _ id.Principal, _ id.AgentID) (*storage.UserGrant, error) {
	return nil, nil
}
func (n *nopGrantRepo) DeleteByAgent(_ context.Context, _ id.AgentID) error { return nil }
func (n *nopGrantRepo) DeleteByPrincipalAndAgentID(_ context.Context, _ id.Principal, _ id.AgentID) error {
	return nil
}
func (n *nopGrantRepo) ListByPrincipal(_ context.Context, _ id.Principal) ([]storage.UserGrant, error) {
	return nil, nil
}
func (n *nopGrantRepo) CountAgentsByServiceID(_ context.Context, _ id.ServiceID) (int, error) {
	return 0, nil
}
func (n *nopGrantRepo) ListByServiceID(_ context.Context, _ id.ServiceID) ([]id.AgentID, error) {
	return nil, nil
}
func (n *nopGrantRepo) CountGrantsReferencingPermissionSet(ctx context.Context, psID id.PermissionSetID) (int, error) {
	if n.countGrantsReferencingFunc != nil {
		return n.countGrantsReferencingFunc(ctx, psID)
	}
	return 0, nil
}

type providerScopeValidatorFunc func(context.Context, id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error)

func (f providerScopeValidatorFunc) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	return f(ctx, serviceID)
}

// newPermissionSetsHandlerForTest creates a PermissionSetsHandler backed by a real service
// wrapping a mock repository. This ensures the architecture invariant holds in tests:
// the handler always goes through the domain service, never raw storage.
func newPermissionSetsHandlerForTest(mockRepo *MockPermissionSetRepository, logger *slog.Logger) *PermissionSetsHandler {
	svc := permissionset.NewPermissionSetService(mockRepo, &nopGrantRepo{}, logger)
	return NewPermissionSetsHandler(svc, providerScopeValidatorFunc(func(context.Context, id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
		return &model.ThirdpartyOAuth2ProviderEntity{Scopes: []model.OAuthScope{{ScopeValue: "read"}}}, nil
	}), logger)
}

func newPermissionSetsHandlerWithGrantRepo(mockRepo *MockPermissionSetRepository, grantRepo *nopGrantRepo, logger *slog.Logger) *PermissionSetsHandler {
	svc := permissionset.NewPermissionSetService(mockRepo, grantRepo, logger)
	return NewPermissionSetsHandler(svc, providerScopeValidatorFunc(func(context.Context, id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
		return &model.ThirdpartyOAuth2ProviderEntity{Scopes: []model.OAuthScope{{ScopeValue: "read"}}}, nil
	}), logger)
}

func newPermissionSetsHandlerWithProviderForTest(mockRepo *MockPermissionSetRepository, providerScopeValidator ProviderScopeValidator, logger *slog.Logger) *PermissionSetsHandler {
	svc := permissionset.NewPermissionSetService(mockRepo, &nopGrantRepo{}, logger)
	return NewPermissionSetsHandler(svc, providerScopeValidator, logger)
}

func TestPermissionSetsHandler_Create(t *testing.T) {
	logger := slog.Default()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		reqBody := CreatePermissionSetRequest{
			Name:        "GitHub Read Access",
			Description: "Read repository contents",
			ServiceScopes: []ServiceScopeRequest{
				{
					ServiceID:       serviceID.String(),
					Scopes:          []string{"repo:read"},
					RequirementType: "optional",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(ps *storage.PermissionSet) bool {
			return ps.Name == "GitHub Read Access" && len(ps.ServiceScopes) == 1
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp PermissionSetResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		assert.Equal(t, "GitHub Read Access", resp.Name)
		assert.Equal(t, "Read repository contents", resp.Description)
		assert.Len(t, resp.ServiceScopes, 1)

		mockRepo.AssertExpectations(t)
	})

	t.Run("creates a permission set with a scope-less service", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerWithProviderForTest(mockRepo, providerScopeValidatorFunc(func(context.Context, id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			return &model.ThirdpartyOAuth2ProviderEntity{}, nil
		}), logger)
		serviceID := id.NewServiceID()
		reqBody := CreatePermissionSetRequest{
			Name:        "Scope-less service",
			Description: "Requires a session without OAuth scopes",
			ServiceScopes: []ServiceScopeRequest{{
				ServiceID:       serviceID.String(),
				RequirementType: "mandatory",
			}},
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(ps *storage.PermissionSet) bool {
			return len(ps.ServiceScopes) == 1 && len(ps.ServiceScopes[0].Scopes) == 0 && ps.ServiceScopes[0].RequirementType == storage.RequirementTypeMandatory
		})).Return(nil)
		w := httptest.NewRecorder()

		handler.Create(w, httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(body)))

		require.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("rejects an empty scope list for a scoped service", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		serviceID := id.NewServiceID()
		handler := newPermissionSetsHandlerWithProviderForTest(mockRepo, providerScopeValidatorFunc(func(_ context.Context, requestedID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			require.Equal(t, serviceID, requestedID)
			return &model.ThirdpartyOAuth2ProviderEntity{Scopes: []model.OAuthScope{{ScopeValue: "read"}}}, nil
		}), logger)
		body, err := json.Marshal(CreatePermissionSetRequest{
			Name:        "Invalid scope-less declaration",
			Description: "Scoped service requires scopes",
			ServiceScopes: []ServiceScopeRequest{{
				ServiceID:       serviceID.String(),
				RequirementType: "mandatory",
			}},
		})
		require.NoError(t, err)
		w := httptest.NewRecorder()

		handler.Create(w, httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(body)))

		require.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("name conflict returns 409", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		reqBody := CreatePermissionSetRequest{
			Name:        "GitHub Read Access",
			Description: "Read repository contents",
			ServiceScopes: []ServiceScopeRequest{
				{
					ServiceID:       serviceID.String(),
					Scopes:          []string{"repo:read"},
					RequirementType: "optional",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("Create", storage.ErrorKindConflict, nil, "permission set name already exists"),
		)

		req := httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error returns 400", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		// Empty service_scopes
		reqBody := CreatePermissionSetRequest{
			Name:          "GitHub Read Access",
			Description:   "Read repository contents",
			ServiceScopes: []ServiceScopeRequest{},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("blank supplied scope returns 400", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)
		reqBody := CreatePermissionSetRequest{
			Name:        "Invalid scope",
			Description: "Contains a blank scope",
			ServiceScopes: []ServiceScopeRequest{{
				ServiceID: id.NewServiceID().String(),
				Scopes:    []string{""},
			}},
		}
		body, err := json.Marshal(reqBody)
		require.NoError(t, err)
		w := httptest.NewRecorder()

		handler.Create(w, httptest.NewRequest(http.MethodPost, "/api/permission-sets", bytes.NewReader(body)))

		require.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertNotCalled(t, "Create")
	})
}

func TestPermissionSetsHandler_Get(t *testing.T) {
	logger := slog.Default()

	t.Run("successful get", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		ps := &storage.PermissionSet{
			ID:          psID,
			Name:        "GitHub Read Access",
			Description: "Read repository contents",
			ServiceScopes: []storage.ServiceScope{
				{
					ServiceID:       serviceID,
					Scopes:          []string{"repo:read"},
					RequirementType: storage.RequirementTypeOptional,
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		mockRepo.On("Get", mock.Anything, psID).Return(ps, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Get(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp PermissionSetResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", resp.ID)
		assert.Equal(t, "GitHub Read Access", resp.Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("not found returns 404", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

		mockRepo.On("Get", mock.Anything, psID).Return(
			nil,
			storage.NewStorageError("Get", storage.ErrorKindNotFound, nil, "permission set not found"),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Get(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPermissionSetsHandler_List(t *testing.T) {
	logger := slog.Default()

	t.Run("list all permission sets", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")
		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

		ps := &storage.PermissionSet{
			ID:          psID,
			Name:        "GitHub Read Access",
			Description: "Read repository contents",
			ServiceScopes: []storage.ServiceScope{
				{
					ServiceID:       serviceID,
					Scopes:          []string{"repo:read"},
					RequirementType: storage.RequirementTypeOptional,
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		mockRepo.On("List", mock.Anything, id.ServiceID{}).Return([]*storage.PermissionSet{ps}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/permission-sets", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp PermissionSetListResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp.Items, 1)
		assert.Equal(t, "GitHub Read Access", resp.Items[0].Name)

		mockRepo.AssertExpectations(t)
	})

	t.Run("list with service_id filter", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		mockRepo.On("List", mock.Anything, serviceID).Return([]*storage.PermissionSet{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/permission-sets?service_id=550e8400-e29b-41d4-a716-446655440000", nil)
		w := httptest.NewRecorder()

		handler.List(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPermissionSetsHandler_Update(t *testing.T) {
	logger := slog.Default()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		existingPS := &storage.PermissionSet{
			ID:          psID,
			Name:        "Old Name",
			Description: "Old description",
			ServiceScopes: []storage.ServiceScope{
				{
					ServiceID:       serviceID,
					Scopes:          []string{"repo:read"},
					RequirementType: storage.RequirementTypeOptional,
				},
			},
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
			UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
		}

		mockRepo.On("Get", mock.Anything, psID).Return(existingPS, nil)

		reqBody := CreatePermissionSetRequest{
			Name:        "GitHub Read Access Updated",
			Description: "Updated description",
			ServiceScopes: []ServiceScopeRequest{
				{
					ServiceID:       serviceID.String(),
					Scopes:          []string{"repo:read", "user:email"},
					RequirementType: "optional",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(ps *storage.PermissionSet) bool {
			return ps.ID == psID && ps.Name == "GitHub Read Access Updated"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("updates with a scope-less service", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		psID := id.NewPermissionSetID()
		serviceID := id.NewServiceID()
		handler := newPermissionSetsHandlerWithProviderForTest(mockRepo, providerScopeValidatorFunc(func(_ context.Context, requestedID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
			require.Equal(t, serviceID, requestedID)
			return &model.ThirdpartyOAuth2ProviderEntity{}, nil
		}), logger)
		mockRepo.On("Get", mock.Anything, psID).Return(&storage.PermissionSet{ID: psID, CreatedAt: time.Now().UTC()}, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(ps *storage.PermissionSet) bool {
			return len(ps.ServiceScopes) == 1 && len(ps.ServiceScopes[0].Scopes) == 0
		})).Return(nil)
		body, err := json.Marshal(CreatePermissionSetRequest{
			Name:        "Scope-less service",
			Description: "Requires a session without OAuth scopes",
			ServiceScopes: []ServiceScopeRequest{{
				ServiceID:       serviceID.String(),
				RequirementType: "mandatory",
			}},
		})
		require.NoError(t, err)
		req := httptest.NewRequest(http.MethodPut, "/api/permission-sets/"+psID.String(), bytes.NewReader(body))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", psID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.Update(w, req)

		require.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("name conflict returns 409", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		serviceID := id.MustParseServiceID("550e8400-e29b-41d4-a716-446655440000")

		existingPS := &storage.PermissionSet{
			ID:          psID,
			Name:        "Old Name",
			Description: "Old description",
			ServiceScopes: []storage.ServiceScope{
				{
					ServiceID:       serviceID,
					Scopes:          []string{"repo:read"},
					RequirementType: storage.RequirementTypeOptional,
				},
			},
			CreatedAt: time.Now().UTC().Add(-1 * time.Hour),
			UpdatedAt: time.Now().UTC().Add(-1 * time.Hour),
		}

		mockRepo.On("Get", mock.Anything, psID).Return(existingPS, nil)

		reqBody := CreatePermissionSetRequest{
			Name:        "GitHub Read Access",
			Description: "Updated description",
			ServiceScopes: []ServiceScopeRequest{
				{
					ServiceID:       serviceID.String(),
					Scopes:          []string{"repo:read"},
					RequirementType: "optional",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Update", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("Update", storage.ErrorKindConflict, nil, "permission set name already exists"),
		)

		req := httptest.NewRequest(http.MethodPut, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Update(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPermissionSetsHandler_Delete(t *testing.T) {
	logger := slog.Default()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

		mockRepo.On("CountAgentsReferencingPermissionSet", mock.Anything, psID).Return(0, nil)
		mockRepo.On("Delete", mock.Anything, psID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("deletion protection returns 409", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		handler := newPermissionSetsHandlerForTest(mockRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

		mockRepo.On("CountAgentsReferencingPermissionSet", mock.Anything, psID).Return(2, nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("active grant reference returns 409", func(t *testing.T) {
		mockRepo := new(MockPermissionSetRepository)
		grantRepo := &nopGrantRepo{
			countGrantsReferencingFunc: func(_ context.Context, _ id.PermissionSetID) (int, error) {
				return 1, nil
			},
		}
		handler := newPermissionSetsHandlerWithGrantRepo(mockRepo, grantRepo, logger)

		psID := id.MustParsePermissionSetID("a1b2c3d4-e5f6-7890-abcd-ef1234567890")

		mockRepo.On("CountAgentsReferencingPermissionSet", mock.Anything, psID).Return(0, nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/permission-sets/a1b2c3d4-e5f6-7890-abcd-ef1234567890", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "a1b2c3d4-e5f6-7890-abcd-ef1234567890")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.Delete(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Contains(t, body["message"], "grant")
		mockRepo.AssertExpectations(t)
	})

}
