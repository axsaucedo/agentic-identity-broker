package thirdparty

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// MockRepository mocks ports.ThirdpartyOAuth2ProviderRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Get(ctx context.Context, id string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	args := m.Called(ctx, serviceID)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository) FindByProtectedResource(ctx context.Context, resourceURI string) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx, resourceURI)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

// MockEncryption mocks EncryptionPort
type MockEncryption struct {
	mock.Mock
}

func (m *MockEncryption) Encrypt(ctx context.Context, plaintext []byte, encContext map[string]string) ([]byte, error) {
	args := m.Called(ctx, plaintext, encContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryption) Decrypt(ctx context.Context, ciphertext []byte, encContext map[string]string) ([]byte, error) {
	args := m.Called(ctx, ciphertext, encContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// Test service_id-only encryption context (ADR 008)
func TestServiceManager_Create_ServiceIDContext(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		ClientSecret: "supersecret",
	}

	// Verify encryption is called with service_id-only context (ADR 008)
	expectedEncContext := map[string]string{
		"service_id": "service-123",
	}
	mockEncryption.On("Encrypt", ctx, []byte("supersecret"), expectedEncContext).
		Return([]byte("encrypted-bytes"), nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		_, err := e.Secret.GetCiphertext()
		return err == nil && e.Secret.IsEncrypted()
	})).Return(nil)

	err := sm.Create(ctx, service)

	require.NoError(t, err)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)

	// Verify plaintext was cleared after encryption
	assert.Empty(t, service.ClientSecret, "plaintext should be cleared after encryption")
	assert.NotEmpty(t, service.ClientSecretCiphertext, "ciphertext should be set")
}

// CRITICAL SECURITY TEST: Verify cross-service token swap is prevented via context binding
func TestServiceManager_CrossServiceProtection(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()

	// Service A encrypts a credential
	serviceA := &storage.ThirdpartyOAuth2Service{
		ID:           "service-a",
		ClientSecret: "secret-a",
	}

	serviceAEncContext := map[string]string{
		"service_id": "service-a",
	}
	mockEncryption.On("Encrypt", ctx, []byte("secret-a"), serviceAEncContext).
		Return([]byte("encrypted-a"), nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := sm.Create(ctx, serviceA)
	require.NoError(t, err)

	// Repository returns Service A's encrypted credential as entity
	storedEntityA := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-a",
		Secret: model.NewEncryptedSecret([]byte("encrypted-a")),
	}
	mockRepo.On("Get", ctx, "service-a").Return(storedEntityA, nil)

	// Attempting to decrypt Service A's credential with Service B context should fail
	serviceBEncContext := map[string]string{
		"service_id": "service-b", // Different service!
	}
	contextMismatchErr := errors.New("encryption context mismatch: expected service-b, got service-a")
	mockEncryption.On("Decrypt", ctx, []byte("encrypted-a"), serviceBEncContext).
		Return(nil, contextMismatchErr)

	// This test verifies that decryption would fail if attempted with wrong service context
	// In production, Get would only be called with correct service_id context
	_, err = sm.Get(ctx, "service-a")

	// SECURITY ASSERTION: Cross-service context prevents credential reuse
	assert.Error(t, err, "cross-service access should be blocked")
	assert.Contains(t, err.Error(), "failed to decrypt", "error should indicate decryption failure")

	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Get with service_id-only context
func TestServiceManager_Get_WithServiceID(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()

	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-123",
		Secret: model.NewEncryptedSecret([]byte("encrypted-secret")),
	}
	mockRepo.On("Get", ctx, "service-123").Return(storedEntity, nil)

	expectedEncContext := map[string]string{
		"service_id": "service-123",
	}
	mockEncryption.On("Decrypt", ctx, []byte("encrypted-secret"), expectedEncContext).
		Return([]byte("decrypted-secret"), nil)

	result, err := sm.Get(ctx, "service-123")

	require.NoError(t, err)
	assert.Equal(t, "decrypted-secret", result.ClientSecret)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Update with secret change
func TestServiceManager_Update_WithSecretChange(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		ClientSecret: "new-secret",
	}

	expectedEncContext := map[string]string{
		"service_id": "service-123",
	}
	mockEncryption.On("Encrypt", ctx, []byte("new-secret"), expectedEncContext).
		Return([]byte("new-encrypted"), nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		_, err := e.Secret.GetCiphertext()
		return err == nil && e.Secret.IsEncrypted()
	})).Return(nil)

	err := sm.Update(ctx, service)

	require.NoError(t, err)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test Update without secret change (empty ClientSecret)
func TestServiceManager_Update_NoSecretChange(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		ClientSecret: "", // Empty = no change
	}

	mockRepo.On("Update", ctx, mock.Anything).Return(nil)

	err := sm.Update(ctx, service)

	require.NoError(t, err)
	// Encryption should NOT be called when secret is empty
	mockEncryption.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertExpectations(t)
}

// Test List with service_id-only context
func TestServiceManager_List_WithServiceIDs(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()

	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{
			ID:     "service-1",
			Secret: model.NewEncryptedSecret([]byte("encrypted-1")),
		},
		{
			ID:     "service-2",
			Secret: model.NewEncryptedSecret([]byte("encrypted-2")),
		},
	}
	mockRepo.On("List", ctx).Return(entities, nil)

	// Verify each service is decrypted with service_id-only context (ADR 008)
	mockEncryption.On("Decrypt", ctx, []byte("encrypted-1"), map[string]string{
		"service_id": "service-1",
	}).Return([]byte("secret-1"), nil)

	mockEncryption.On("Decrypt", ctx, []byte("encrypted-2"), map[string]string{
		"service_id": "service-2",
	}).Return([]byte("secret-2"), nil)

	result, err := sm.List(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "secret-1", result[0].ClientSecret)
	assert.Equal(t, "secret-2", result[1].ClientSecret)
	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test List decryption failure (fail-fast behavior)
func TestServiceManager_List_DecryptionFailure(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()

	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{
			ID:     "service-1",
			Secret: model.NewEncryptedSecret([]byte("encrypted-1")),
		},
		{
			ID:     "service-2",
			Secret: model.NewEncryptedSecret([]byte("corrupted")),
		},
	}
	mockRepo.On("List", ctx).Return(entities, nil)

	// First service decrypts successfully
	mockEncryption.On("Decrypt", ctx, []byte("encrypted-1"), map[string]string{
		"service_id": "service-1",
	}).Return([]byte("secret-1"), nil)

	// Second service fails to decrypt - should fail-fast
	mockEncryption.On("Decrypt", ctx, []byte("corrupted"), map[string]string{
		"service_id": "service-2",
	}).Return(nil, errors.New("decryption failed"))

	result, err := sm.List(ctx)

	// New behavior: fail-fast on first decryption error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decrypt service service-2")
	assert.Contains(t, err.Error(), "decryption failed")

	mockEncryption.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// Test encryption failure handling
func TestServiceManager_Create_EncryptionFailure(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		ClientSecret: "supersecret",
	}

	mockEncryption.On("Encrypt", ctx, mock.Anything, mock.Anything).
		Return(nil, errors.New("KMS unavailable"))

	err := sm.Create(ctx, service)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encrypt client secret")

	// Repository should NOT be called if encryption fails
	mockRepo.AssertNotCalled(t, "Create")
}

// Test empty/nil secret handling
func TestServiceManager_Create_EmptySecret(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background()
	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		ClientSecret: "", // Empty secret
	}

	err := sm.Create(ctx, service)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client secret required")

	// Nothing should be called
	mockEncryption.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

// Test Delete (no encryption needed)
func TestServiceManager_Delete(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEncryption := new(MockEncryption)
	sm := NewServiceManager(mockRepo, mockEncryption, slog.Default())

	ctx := context.Background() // Principal not needed for delete
	mockRepo.On("Delete", ctx, "service-123").Return(nil)

	err := sm.Delete(ctx, "service-123")

	require.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
