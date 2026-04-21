package admin

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSigningKeyRepository struct {
	mock.Mock
}

func (m *MockSigningKeyRepository) Create(ctx context.Context, key *storage.SigningKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockSigningKeyRepository) CreateAndSetCurrent(ctx context.Context, key *storage.SigningKey) error {
	return m.Called(ctx, key).Error(0)
}

func (m *MockSigningKeyRepository) GetByKID(ctx context.Context, kid id.KeyID) (*storage.SigningKey, error) {
	args := m.Called(ctx, kid)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.SigningKey), args.Error(1)
}

func (m *MockSigningKeyRepository) GetCurrent(ctx context.Context) (*storage.SigningKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.SigningKey), args.Error(1)
}

func (m *MockSigningKeyRepository) ListActive(ctx context.Context) ([]*storage.SigningKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.SigningKey), args.Error(1)
}

func (m *MockSigningKeyRepository) SetCurrent(ctx context.Context, kid id.KeyID) error {
	return m.Called(ctx, kid).Error(0)
}

func (m *MockSigningKeyRepository) Delete(ctx context.Context, kid id.KeyID) error {
	return m.Called(ctx, kid).Error(0)
}

func (m *MockSigningKeyRepository) CountActive(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type MockSigningKeyManager struct {
	mock.Mock
}

func (m *MockSigningKeyManager) GenerateAndStoreKey(ctx context.Context, algorithm string, makeCurrent bool) (*storage.SigningKey, error) {
	args := m.Called(ctx, algorithm, makeCurrent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.SigningKey), args.Error(1)
}

func (m *MockSigningKeyManager) BuildJWKS(ctx context.Context) (jwk.Set, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwk.Set), args.Error(1)
}

func (m *MockSigningKeyManager) DeleteKey(ctx context.Context, kid id.KeyID) error {
	return m.Called(ctx, kid).Error(0)
}

func newTestSigningKeysHandler(repo *MockSigningKeyRepository, manager *MockSigningKeyManager) *SigningKeysHandler {
	return NewSigningKeysHandler(repo, manager, slog.Default())
}

func stubSigningKey() *storage.SigningKey {
	return &storage.SigningKey{
		KID:         id.NewKeyID("test-kid"),
		Algorithm:   "ES256",
		IsCurrent:   true,
		ActivatesAt: time.Now(),
		CreatedAt:   time.Now(),
	}
}

func TestSigningKeysHandler_Add(t *testing.T) {
	t.Run("empty body defaults to ES256", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}
		manager.On("GenerateAndStoreKey", mock.Anything, "ES256", true).Return(stubSigningKey(), nil)

		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", http.NoBody)
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		manager.AssertExpectations(t)
	})

	t.Run("valid JSON with ES256 creates key", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}
		manager.On("GenerateAndStoreKey", mock.Anything, "ES256", true).Return(stubSigningKey(), nil)

		body := `{"algorithm":"ES256"}`
		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", strings.NewReader(body))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		manager.AssertExpectations(t)
	})

	t.Run("omitted algorithm field defaults to ES256", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}
		manager.On("GenerateAndStoreKey", mock.Anything, "ES256", true).Return(stubSigningKey(), nil)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", strings.NewReader(body))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		manager.AssertExpectations(t)
	})

	t.Run("malformed JSON returns 400", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}

		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", strings.NewReader(`{bad json`))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "invalid request body", resp.Error)
		manager.AssertNotCalled(t, "GenerateAndStoreKey")
	})

	t.Run("unsupported algorithm returns 400", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}

		body := `{"algorithm":"RS256"}`
		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", strings.NewReader(body))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "unsupported algorithm", resp.Error)
		manager.AssertNotCalled(t, "GenerateAndStoreKey")
	})

	t.Run("service error returns 500", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}
		manager.On("GenerateAndStoreKey", mock.Anything, "ES256", true).
			Return(nil, storage.NewStorageError("test", storage.ErrorKindUnknown, nil, "db error"))

		body := `{"algorithm":"ES256"}`
		req := httptest.NewRequest(http.MethodPost, "/api/oauth2-server/signing-keys", strings.NewReader(body))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Add(w, req)

		require.Equal(t, http.StatusInternalServerError, w.Code)
		manager.AssertExpectations(t)
	})
}

func TestSigningKeysHandler_Remove(t *testing.T) {
	t.Run("last active key conflict returns 409", func(t *testing.T) {
		repo := &MockSigningKeyRepository{}
		manager := &MockSigningKeyManager{}
		manager.On("DeleteKey", mock.Anything, id.NewKeyID("test-kid")).
			Return(ports.ErrLastActiveKey)

		req := httptest.NewRequest(http.MethodDelete, "/api/oauth2-server/signing-keys/test-kid", http.NoBody)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("kid", "test-kid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		newTestSigningKeysHandler(repo, manager).Remove(w, req)

		require.Equal(t, http.StatusConflict, w.Code)
		manager.AssertExpectations(t)
	})
}
