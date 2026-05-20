package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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

func (m *MockRepository) Get(ctx context.Context, serviceID id.ServiceID) (*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx, serviceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, entity *model.ThirdpartyOAuth2ProviderEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, serviceID id.ServiceID) error {
	args := m.Called(ctx, serviceID)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context) ([]*model.ThirdpartyOAuth2ProviderEntity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ThirdpartyOAuth2ProviderEntity), args.Error(1)
}

func (m *MockRepository) CountGrantsReferencingService(ctx context.Context, serviceID id.ServiceID) (int, error) {
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

func (m *MockBranchKeyManager) Create(ctx context.Context, serviceID id.ServiceID) (string, error) {
	args := m.Called(ctx, serviceID)
	return args.String(0), args.Error(1)
}

// minimalValidEntity returns the smallest ThirdpartyOAuth2ProviderEntity that passes
// ValidateForCreate. Use this as the base for tests focused on service behavior
// (encryption, branch keys, storage) rather than validation logic.
func minimalValidEntity(svcID id.ServiceID, secret model.Secret) *model.ThirdpartyOAuth2ProviderEntity {
	return &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "Test Provider",
		ClientID:    id.ClientID("test-client-id"),
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	// Entity missing required DisplayName — validation must reject before ID is assigned
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     id.ServiceID{}, // deliberately zero to observe ID generation behavior
		Secret: model.NewPlaintextSecret("secret"),
	}

	err := svc.Create(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	assert.True(t, entity.ID.IsZero(), "ID must NOT be generated when validation fails")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_ValidationRejectsBeforeBranchKey(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, nil, false, slog.Default())

	ctx := context.Background()
	// Invalid entity (missing DisplayName) — branch key must NOT be provisioned
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     id.NewServiceID(),
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity(id.NewServiceID(), model.NewPlaintextSecret("secret"))
	entity.IssuerURI = "http://issuer.example.com" // HTTP non-localhost

	err := svc.Create(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	mockEnc.AssertNotCalled(t, "Encrypt")
}

func TestThirdpartyOAuth2ProviderService_Create_HTTPIssuerAllowedWithSkipHTTPS(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, true, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity(id.NewServiceID(), model.NewPlaintextSecret("secret"))
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	// Entity missing required DisplayName — encryption must NOT be attempted
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     id.NewServiceID(),
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity(id.NewServiceID(), model.NewPlaintextSecret("supersecret"))

	expectedEncContext := map[string]string{"service_id": entity.ID.String()}
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity(id.ServiceID{}, model.NewPlaintextSecret("my-secret"))

	mockEnc.On("Encrypt", ctx, []byte("my-secret"), mock.MatchedBy(func(ec map[string]string) bool {
		sid, ok := ec["service_id"]
		return ok && sid != ""
	})).Return([]byte("enc"), nil)
	mockRepo.On("Create", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		return !e.ID.IsZero()
	})).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	assert.False(t, entity.ID.IsZero(), "ID should be generated")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Create_WithBranchKeyManager(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("secret"))

	callOrder := make([]string, 0)
	mockBKM.On("Create", ctx, svcID).Return("bk-1", nil).Run(func(_ mock.Arguments) {
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("secret"))

	mockBKM.On("Create", ctx, svcID).Return("", errors.New("KMS unavailable"))

	err := svc.Create(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch key provisioning failed")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_EncryptionFailure(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	entity := minimalValidEntity(id.NewServiceID(), model.NewPlaintextSecret("secret"))

	mockEnc.On("Encrypt", ctx, mock.Anything, mock.Anything).Return(nil, errors.New("KMS error"))

	err := svc.Create(ctx, entity)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to encrypt")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestThirdpartyOAuth2ProviderService_Create_EncryptedSecretFails(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	// Entity with encrypted secret — ValidateForCreate requires plaintext state
	entity := minimalValidEntity(id.NewServiceID(), model.NewEncryptedSecret([]byte("already-encrypted")))

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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "GitHub",
		Secret:      model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("Get", ctx, svcID).Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("ciphertext"), map[string]string{"service_id": svcID.String()}).
		Return([]byte("plaintext-secret"), nil)

	result, err := svc.Get(ctx, svcID)

	require.NoError(t, err)
	assert.Equal(t, svcID, result.ID)
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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	nonexistentID := id.NewServiceID()
	mockRepo.On("Get", ctx, nonexistentID).Return(nil, errors.New("not found"))

	result, err := svc.Get(ctx, nonexistentID)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockEnc.AssertNotCalled(t, "Decrypt")
}

// CRITICAL SECURITY TEST: cross-service token swap prevention via context binding
// Even though Get is graceful on decryption failure (returns entity with encrypted secret),
// the security property is maintained: the entity's Secret is NOT in plaintext state,
// so any caller attempting to use the secret (e.g. OAuth2SessionService) will fail
// because Secret.GetPlaintext() returns an error for encrypted secrets.
func TestThirdpartyOAuth2ProviderService_CrossServiceProtection(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()

	// Service A's data stored with "service-a" binding
	svcAID := id.NewServiceID()
	storedEntityA := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     svcAID,
		Secret: model.NewEncryptedSecret([]byte("encrypted-for-a")),
	}
	mockRepo.On("Get", ctx, svcAID).Return(storedEntityA, nil)

	// Decryption fails because Service A's context is bound to its ID
	contextMismatchErr := errors.New("encryption context mismatch")
	mockEnc.On("Decrypt", ctx, []byte("encrypted-for-a"), map[string]string{"service_id": svcAID.String()}).
		Return(nil, contextMismatchErr)

	result, err := svc.Get(ctx, svcAID)

	// Get returns the entity gracefully (no error) but secret remains encrypted
	require.NoError(t, err)
	require.NotNil(t, result)
	// Security assertion: secret is NOT in plaintext state — callers cannot extract it
	assert.True(t, result.Secret.IsEncrypted(), "secret must remain encrypted when decryption fails")
	assert.False(t, result.Secret.IsPlaintext(), "secret must NOT be plaintext when decryption fails")
	_, ptErr := result.Secret.GetPlaintext()
	assert.Error(t, ptErr, "GetPlaintext must fail for encrypted secret — prevents cross-service token swap")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Get_DecryptionFailure_ReturnsEncryptedEntity(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "GitHub",
		Secret:      model.NewEncryptedSecret([]byte("old-encryption-ciphertext")),
	}
	mockRepo.On("Get", ctx, svcID).Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("old-encryption-ciphertext"), map[string]string{"service_id": svcID.String()}).
		Return(nil, errors.New("decryption failed: wrong encryption backend"))

	result, err := svc.Get(ctx, svcID)

	// Get succeeds gracefully — entity returned with encrypted secret
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, svcID, result.ID)
	assert.Equal(t, "GitHub", result.DisplayName)
	assert.True(t, result.Secret.IsEncrypted(), "secret should remain encrypted on decryption failure")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Update tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Update_WithNewSecret(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("new-secret"))

	mockEnc.On("Encrypt", ctx, []byte("new-secret"), map[string]string{"service_id": svcID.String()}).
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

func TestThirdpartyOAuth2ProviderService_Update_ProvisionsBranchKey(t *testing.T) {
	// Verifies that Update provisions the branch key BEFORE encrypting. This is the
	// migration path: a service created with a different encryption backend (e.g. raw AES)
	// has no branch key in the KMS key store; Update must provision it so encryption succeeds.
	// Call-order is asserted via an atomic sequence counter: Create must increment it before
	// Encrypt reads it, so the value seen by Encrypt is always > 0.
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("new-secret"))

	var callSeq atomic.Int32 // incremented by each call; used to verify ordering

	var createSeq int32
	mockBKM.On("Create", ctx, svcID).
		Return("sentinel-branch-key-id", nil).
		Run(func(args mock.Arguments) {
			createSeq = callSeq.Add(1)
		})

	var encryptSeq int32
	mockEnc.On("Encrypt", ctx, []byte("new-secret"), map[string]string{"service_id": svcID.String()}).
		Return([]byte("new-encrypted"), nil).
		Run(func(args mock.Arguments) {
			encryptSeq = callSeq.Add(1)
		})

	mockRepo.On("Update", ctx, mock.MatchedBy(func(e *model.ThirdpartyOAuth2ProviderEntity) bool {
		return e.Secret.IsEncrypted()
	})).Return(nil)

	err := svc.Update(ctx, entity)

	require.NoError(t, err)
	assert.True(t, entity.Secret.IsEncrypted(), "secret must be encrypted after update")
	assert.Less(t, createSeq, encryptSeq, "branch key Create must be called before Encrypt")
	mockBKM.AssertExpectations(t)
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Update_BranchKeyProvisioningFailure_AbortUpdate(t *testing.T) {
	// If branch key provisioning fails (e.g. DynamoDB unavailable), Update must fail
	// before attempting encryption so no partial state is written.
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	mockBKM := new(MockBranchKeyManager)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, mockBKM, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("new-secret"))

	mockBKM.On("Create", ctx, svcID).Return("", fmt.Errorf("DynamoDB unavailable"))

	err := svc.Update(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch key provisioning failed")
	mockBKM.AssertExpectations(t)
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestThirdpartyOAuth2ProviderService_Update_EncryptedSecretFails(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	// Passing an already-encrypted secret must be rejected — callers must always supply
	// plaintext so re-encryption always runs (prevents silent bypass during key rotation).
	entity := minimalValidEntity(id.NewServiceID(), model.NewEncryptedSecret([]byte("existing-ciphertext")))

	err := svc.Update(ctx, entity)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "provider validation failed")
	assert.Contains(t, err.Error(), "client_secret is required for update")
	mockEnc.AssertNotCalled(t, "Encrypt")
	mockRepo.AssertNotCalled(t, "Update")
}

// =============================================================================
// List tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_List_DecryptsAll(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svc1ID := id.NewServiceID()
	svc2ID := id.NewServiceID()
	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{ID: svc1ID, Secret: model.NewEncryptedSecret([]byte("enc-1"))},
		{ID: svc2ID, Secret: model.NewEncryptedSecret([]byte("enc-2"))},
	}
	mockRepo.On("List", ctx).Return(entities, nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-1"), map[string]string{"service_id": svc1ID.String()}).
		Return([]byte("secret-1"), nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-2"), map[string]string{"service_id": svc2ID.String()}).
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

func TestThirdpartyOAuth2ProviderService_List_GracefulDecryptionFailure(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svc1ID := id.NewServiceID()
	svc2ID := id.NewServiceID()
	entities := []*model.ThirdpartyOAuth2ProviderEntity{
		{ID: svc1ID, Secret: model.NewEncryptedSecret([]byte("enc-1"))},
		{ID: svc2ID, Secret: model.NewEncryptedSecret([]byte("corrupted"))},
	}
	mockRepo.On("List", ctx).Return(entities, nil)
	mockEnc.On("Decrypt", ctx, []byte("enc-1"), map[string]string{"service_id": svc1ID.String()}).
		Return([]byte("secret-1"), nil)
	mockEnc.On("Decrypt", ctx, []byte("corrupted"), map[string]string{"service_id": svc2ID.String()}).
		Return(nil, errors.New("decryption failed"))

	results, err := svc.List(ctx)

	// List succeeds even when individual decryption fails
	require.NoError(t, err)
	require.Len(t, results, 2)

	// First entity is decrypted successfully
	pt1, err := results[0].Secret.GetPlaintext()
	require.NoError(t, err)
	assert.Equal(t, "secret-1", pt1)

	// Second entity is returned with encrypted secret (decryption failed gracefully)
	assert.True(t, results[1].Secret.IsEncrypted(), "failed entity should retain encrypted secret")
	assert.Equal(t, svc2ID, results[1].ID)
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// Delete tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_Delete(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	mockRepo.On("Delete", ctx, svcID).Return(nil)

	err := svc.Delete(ctx, svcID)

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
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     svcID,
		Secret: model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("FindByProtectedResource", ctx, "https://api.example.com").Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("ciphertext"), map[string]string{"service_id": svcID.String()}).
		Return([]byte("plaintext"), nil)

	result, err := svc.FindByProtectedResource(ctx, "https://api.example.com")

	require.NoError(t, err)
	assert.Equal(t, svcID, result.ID)
	assert.True(t, result.Secret.IsPlaintext())
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_FindByProtectedResource_DecryptionFailure_ReturnsEncryptedEntity(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	storedEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:     svcID,
		Secret: model.NewEncryptedSecret([]byte("old-ciphertext")),
	}
	mockRepo.On("FindByProtectedResource", ctx, "https://api.example.com").Return(storedEntity, nil)
	mockEnc.On("Decrypt", ctx, []byte("old-ciphertext"), map[string]string{"service_id": svcID.String()}).
		Return(nil, errors.New("decryption failed: wrong encryption backend"))

	result, err := svc.FindByProtectedResource(ctx, "https://api.example.com")

	// Returns entity with encrypted secret instead of failing
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, svcID, result.ID)
	assert.True(t, result.Secret.IsEncrypted(), "secret should remain encrypted on decryption failure")
	mockEnc.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_Create_ServiceIDOnlyContext(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := minimalValidEntity(svcID, model.NewPlaintextSecret("secret"))

	// ADR 008: only service_id in context, no principal
	expectedContext := map[string]string{"service_id": svcID.String()}
	mockEnc.On("Encrypt", ctx, []byte("secret"), expectedContext).Return([]byte("enc"), nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil)

	err := svc.Create(ctx, entity)

	require.NoError(t, err)
	mockEnc.AssertExpectations(t)
}

// =============================================================================
// ValidateServiceRequirements tests
// =============================================================================

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_Empty(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	err := svc.ValidateServiceRequirements(context.Background(), nil)
	assert.NoError(t, err)

	err = svc.ValidateServiceRequirements(context.Background(), []storage.ServiceRequirement{})
	assert.NoError(t, err)

	mockRepo.AssertNotCalled(t, "Get")
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_ValidScopes(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "GitHub",
		Scopes: []model.OAuthScope{
			{ScopeValue: "repo"},
			{ScopeValue: "user:email"},
		},
		Secret: model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("Get", ctx, svcID).Return(entity, nil)

	serviceReqs := []storage.ServiceRequirement{
		{
			ServiceID:       svcID,
			RequirementType: storage.RequirementTypeMandatory,
			RequiredScopes:  []string{"repo", "user:email"},
		},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_MultipleServices(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	githubID := id.NewServiceID()
	gitlabID := id.NewServiceID()
	entity1 := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          githubID,
		DisplayName: "GitHub",
		Scopes:      []model.OAuthScope{{ScopeValue: "repo"}, {ScopeValue: "user:email"}},
		Secret:      model.NewEncryptedSecret([]byte("enc1")),
	}
	entity2 := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          gitlabID,
		DisplayName: "GitLab",
		Scopes:      []model.OAuthScope{{ScopeValue: "api"}, {ScopeValue: "read_user"}},
		Secret:      model.NewEncryptedSecret([]byte("enc2")),
	}
	mockRepo.On("Get", ctx, githubID).Return(entity1, nil)
	mockRepo.On("Get", ctx, gitlabID).Return(entity2, nil)

	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: githubID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"repo"}},
		{ServiceID: gitlabID, RequirementType: storage.RequirementTypeOptional, RequiredScopes: []string{"api"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_ServiceNotFound(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	missingSvcID := id.NewServiceID()
	notFoundErr := storage.NewStorageError("Get", storage.ErrorKindNotFound, nil, "not found")
	mockRepo.On("Get", ctx, missingSvcID).Return(nil, notFoundErr)

	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: missingSvcID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"read"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	assert.Contains(t, storageErr.Message, missingSvcID.String())
	assert.Contains(t, storageErr.Message, "not found")
	assert.Contains(t, storageErr.Message, "index 0")
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_InvalidScope(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "GitHub",
		Scopes:      []model.OAuthScope{{ScopeValue: "repo"}, {ScopeValue: "user:email"}},
		Secret:      model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("Get", ctx, svcID).Return(entity, nil)

	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: svcID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"repo", "invalid:scope"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	assert.Contains(t, storageErr.Message, "invalid:scope")
	assert.Contains(t, storageErr.Message, "GitHub")
	assert.Contains(t, storageErr.Message, "index 0")
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_SecondIndexError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svc1ID := id.NewServiceID()
	missingSvcID := id.NewServiceID()
	entity1 := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svc1ID,
		DisplayName: "GitHub",
		Scopes:      []model.OAuthScope{{ScopeValue: "repo"}},
		Secret:      model.NewEncryptedSecret([]byte("enc1")),
	}
	notFoundErr := storage.NewStorageError("Get", storage.ErrorKindNotFound, nil, "not found")
	mockRepo.On("Get", ctx, svc1ID).Return(entity1, nil)
	mockRepo.On("Get", ctx, missingSvcID).Return(nil, notFoundErr)

	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: svc1ID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"repo"}},
		{ServiceID: missingSvcID, RequirementType: storage.RequirementTypeOptional, RequiredScopes: []string{"read"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	assert.Contains(t, storageErr.Message, "index 1")
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_StorageError(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svc1ID := id.NewServiceID()
	connErr := storage.NewStorageError("Get", storage.ErrorKindConnection, nil, "database connection failed")
	mockRepo.On("Get", ctx, svc1ID).Return(nil, connErr)

	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: svc1ID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"repo"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindConnection, storageErr.Kind)
	assert.Contains(t, storageErr.Message, "database connection failed")
	mockRepo.AssertExpectations(t)
}

func TestThirdpartyOAuth2ProviderService_ValidateServiceRequirements_CaseSensitiveScopes(t *testing.T) {
	mockRepo := new(MockRepository)
	mockEnc := new(MockEncryption)
	svc := NewThirdpartyOAuth2ProviderService(mockRepo, mockEnc, nil, nil, false, slog.Default())

	ctx := context.Background()
	svcID := id.NewServiceID()
	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          svcID,
		DisplayName: "GitHub",
		Scopes:      []model.OAuthScope{{ScopeValue: "repo"}},
		Secret:      model.NewEncryptedSecret([]byte("ciphertext")),
	}
	mockRepo.On("Get", ctx, svcID).Return(entity, nil)

	// "REPO" must not match "repo" — OAuth2 scopes are case-sensitive
	serviceReqs := []storage.ServiceRequirement{
		{ServiceID: svcID, RequirementType: storage.RequirementTypeMandatory, RequiredScopes: []string{"REPO"}},
	}

	err := svc.ValidateServiceRequirements(ctx, serviceReqs)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	assert.Contains(t, storageErr.Message, "REPO")
	mockRepo.AssertExpectations(t)
}
