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

// MockEncryption mocks ports.EncryptionPort
type MockEncryption struct {
	mock.Mock
}

func (m *MockEncryption) Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error) {
	args := m.Called(ctx, plaintext, encryptionContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncryption) Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error) {
	args := m.Called(ctx, ciphertext, encryptionContext)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

// MockBranchKeyManager mocks ports.BranchKeyManager
type MockBranchKeyManager struct {
	mock.Mock
}

func (m *MockBranchKeyManager) Create(ctx context.Context, serviceID string) (string, error) {
	args := m.Called(ctx, serviceID)
	return args.String(0), args.Error(1)
}

// minimalValidEntity returns the smallest ThirdpartyOAuth2ProviderEntity that passes
// ValidateForCreate. Use this as the base for tests focused on service behavior
// (encryption, branch keys, storage) rather than validation logic.
func minimalValidEntity(id string, secret model.Secret) *model.ThirdpartyOAuth2ProviderEntity {
	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          id,
		DisplayName: "Test Provider",
		ClientID:    "test-client-id",
		Secret:      secret,
		IssuerURI:   "https://issuer.example.com",
		Discovery:   model.DiscoveryConfig{EnableDiscovery: true},
		Scopes:      []model.OAuthScope{{ScopeValue: "read", Description: "Read access"}},
	}
}

// =============================================================================
// Validation tests (ValidateForCreate / ValidateForUpdate)
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Create_ValidationRejectsBeforeIDGeneration(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	// Entity missing required DisplayName — validation must reject before ID is assigned
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "", // deliberately empty to observe ID generation behavior
		Secret: model.NewPlaintextSecret("secret"),
	}

	err := svc.Create(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	assert.Empty(t, entity.ID, "ID must NOT be generated when validation fails")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_ValidationRejectsBeforeBranchKey(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, false, slog.Default())

	ctx := context.Background()
	// Invalid entity (missing DisplayName) — branch key must NOT be provisioned
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-123",
		Secret: model.NewPlaintextSecret("secret"),
	}

	err := svc.Create(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	mockBKM.AssertNotCalled(t, "Create")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_HTTPIssuerRejectedByDefault(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("svc-1", model.NewPlaintextSecret("secret"))
	entity.IssuerURI = "http://issuer.example.com" // HTTP non-localhost

	err := svc.Create(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	mockEnc.AssertNotCalled(t, "Encrypt")
}

func TestThirdpartyOAuth2ProviderService_Create_HTTPIssuerAllowedWithSkipHTTPS(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, true, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("svc-1", model.NewPlaintextSecret("secret"))
	entity.IssuerURI = "http://issuer.example.com" // HTTP non-localhost, allowed in dev

	mockEnc.On("Encrypt", ctx, mock.Anything, mock.Anything).Return([]byte("enc"), nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	mockEnc.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Update_ValidationRejectsBeforeEncryption(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	// Entity missing required DisplayName — encryption must NOT be attempted
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-123",
		Secret: model.NewPlaintextSecret("new-secret"),
	}

	err := svc.Update(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Update")
}

// =============================================================================
// Create tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Create_EncryptsAndStores(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewPlaintextSecret("supersecret"))

	expectedEncContext := map[string]string{"service_id": "service-123"}
	mockEnc.On("Encrypt", ctx, []byte("supersecret"), expectedEncContext).
		Return([]byte("encrypted-bytes"), nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		ct, err := e.Secret.GetCiphertext()
		return err == nil && e.Secret.IsEncrypted() && string(ct) == "encrypted-bytes"
	})).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	// Secret must be in encrypted state after Create
	assert.True(t, entity.Secret.IsEncrypted())
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Create_GeneratesIDIfEmpty(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("", model.NewPlaintextSecret("my-secret"))

	mockEnc.On("Encrypt", ctx, []byte("my-secret"), mock.MatchedBy(func(ec map[string]string) bool {
		id, ok := ec["service_id"]
		return ok && id != "" // ID must be set in encryption context
	})).Return([]byte("enc"), nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		return e.ID != "" // ID must be set
	})).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	assert.NotEmpty(t, entity.ID, "ID should be generated")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Create_WithBranchKeyManager(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewPlaintextSecret("secret"))

	callOrder := make([]string, 0)
	mockBKM.On("Create", ctx, "service-123").Return("bk-1", nil).Run(func(_ mock.Arguments) {
		callOrder = append(callOrder, "branch_key")
	})
	mockEnc.On("Encrypt", ctx, mock.Anything, mock.Anything).Return([]byte("enc"), nil).Run(func(_ mock.Arguments) {
		callOrder = append(callOrder, "encrypt")
	})
	mockRepo.On("Create", ctx, mock.Anything).Return(nil).Run(func(_ mock.Arguments) {
		callOrder = append(callOrder, "repo_create")
	})

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	assert.Equal(t, []string{"branch_key", "encrypt", "repo_create"}, callOrder,
		"branch key must be provisioned before encryption and repo storage")
	mockBKM.AssertExpectations(t)
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Create_BranchKeyFailure_AbortCreate(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewPlaintextSecret("secret"))

	mockBKM.On("Create", ctx, "service-123").Return("", errors.New("KMS unavailable"))

	err := svc.Create(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch key provisioning failed")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_EncryptionFailure(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewPlaintextSecret("secret"))

	mockEnc.On("Encrypt", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("KMS error"))

	err := svc.Create(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_EncryptedSecretFails(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	// Entity with encrypted secret — ValidateForCreate requires plaintext state
	entity := minimalValidEntity("service-123", model.NewEncryptedSecret([]byte("already-encrypted")))

	err := svc.Create(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client_secret is required for create")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

// =============================================================================
// Get tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Get_DecryptsSecret(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "service-123",
		DisplayName: "GitHub",
		Secret:      model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("Get", ctx, "service-123").Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("ciphertext"), map[string]string{"service_id": "service-123"}).
		Return([]byte("plaintext-secret"), nil)

	result, err := svc.Get(ctx, "service-123")

	require.NoError(t, err)
	assert.Equal(t, "service-123", result.ID)
	assert.Equal(t, "GitHub", result.DisplayName)
	assert.True(t, result.Secret.IsPlaintext())
	pt, err := result.Secret.GetPlaintext()
	require.NoError(t, err)
	assert.Equal(t, "plaintext-secret", pt)
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Get_NotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	mockRepo.On("Get", ctx, "nonexistent").Return(nil, errors.New("not found"))

	result, err := svc.Get(ctx, "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, result)
	mockEnc.AssertNotCalled(t, "Decrypt")
}

// CRITICAL SECURITY TEST: cross-service token swap prevention via context binding
func TestThirdpartyOAuth2ProviderService_CrossServiceProtection(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()

	// Service A's data stored with "service-a" binding
	storedEntityA := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-a",
		Secret: model.NewEncryptedSecret([]byte("encrypted-for-a")),
	}
	mockRepo.On("Get", ctx, "service-a").Return(storedEntityA, nil)

	// Decryption fails because Service A's context is {"service_id": "service-a"}
	contextMismatchErr := errors.New("encryption context mismatch")
	mockEnc.On("Decrypt", ctx, []byte("encrypted-for-a"), map[string]string{"service_id": "service-a"}).
		Return(nil, contextMismatchErr)

	_, err := svc.Get(ctx, "service-a")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
	mockEnc.AssertExpectations(t)
}

// =============================================================================
// Update tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Update_WithNewSecret(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewPlaintextSecret("new-secret"))

	mockEnc.On("Encrypt", ctx, []byte("new-secret"), map[string]string{"service_id": "service-123"}).
		Return([]byte("new-encrypted"), nil)
	mockRepo.On("Update", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		return e.Secret.IsEncrypted()
	})).Return(nil)

	err := svc.Update(ctx, entity)

	require.NoError(t, err)
	assert.True(t, entity.Secret.IsEncrypted(), "secret must be encrypted after update")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Update_NoSecretChange(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-123", model.NewEncryptedSecret([]byte("existing-ciphertext")))

	mockRepo.On("Update", ctx, entity).Return(nil)

	err := svc.Update(ctx, entity)

	require.NoError(t, err)
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// List tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_List_DecryptsAll(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{ID: "svc-1", Secret: model.NewEncryptedSecret([]byte("enc-1"))},
		{ID: "svc-2", Secret: model.NewEncryptedSecret([]byte("enc-2"))},
	}
	mockRepo.On("List", ctx).Return(entities, nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-1"), map[string]string{"service_id": "svc-1"}).
		Return([]byte("secret-1"), nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-2"), map[string]string{"service_id": "svc-2"}).
		Return([]byte("secret-2"), nil)

	results, err := svc.List(ctx)

	require.NoError(t, err)
	require.Len(t, results, 2)
	pt1, _ := results[0].Secret.GetPlaintext()
	pt2, _ := results[1].Secret.GetPlaintext()
	assert.Equal(t, "secret-1", pt1)
	assert.Equal(t, "secret-2", pt2)
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_List_FailFastOnDecryptionError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{ID: "svc-1", Secret: model.NewEncryptedSecret([]byte("enc-1"))},
		{ID: "svc-2", Secret: model.NewEncryptedSecret([]byte("corrupted"))},
	}
	mockRepo.On("List", ctx).Return(entities, nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-1"), map[string]string{"service_id": "svc-1"}).
		Return([]byte("secret-1"), nil)
	mockEnc.On("Decrypt", ctx, []byte("corrupted"), map[string]string{"service_id": "svc-2"}).
		Return(nil, errors.New("decryption failed"))

	results, err := svc.List(ctx)

	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "svc-2")
	mockEnc.AssertExpectations(t)
}

// =============================================================================
// Delete tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Delete(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	mockRepo.On("Delete", ctx, "service-123").Return(nil)

	err := svc.Delete(ctx, "service-123")

	require.NoError(t, err)
	mockEnc.AssertNotCalled(t, "Decrypt")
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// FindByProtectedResource tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_FindByProtectedResource_DecryptsSecret(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     "service-123",
		Secret: model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("FindByProtectedResource", ctx, "https://api.example.com").Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("ciphertext"), map[string]string{"service_id": "service-123"}).
		Return([]byte("plaintext"), nil)

	result, err := svc.FindByProtectedResource(ctx, "https://api.example.com")

	require.NoError(t, err)
	assert.Equal(t, "service-123", result.ID)
	assert.True(t, result.Secret.IsPlaintext())
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Service ID-only encryption context (ADR 008)
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Create_ServiceIDOnlyContext(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity("service-abc", model.NewPlaintextSecret("secret"))

	// ADR 008: only service_id in context, no principal
	expectedContext := map[string]string{"service_id": "service-abc"}
	mockEnc.On("Encrypt", ctx, []byte("secret"), expectedContext).Return([]byte("enc"), nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	mockEnc.AssertExpectations(t)
}
