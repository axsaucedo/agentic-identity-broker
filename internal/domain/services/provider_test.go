package services

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockThirdpartyOAuth2ServiceRepository is a mock implementation of ports.ThirdpartyOAuth2ServiceRepository
type MockThirdpartyOAuth2ServiceRepository struct {
	mock.Mock
}

func (m *MockThirdpartyOAuth2ServiceRepository) Create(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockThirdpartyOAuth2ServiceRepository) Get(ctx context.Context, id string) (*storage.ThirdpartyOAuth2Service, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.ThirdpartyOAuth2Service), args.Error(1)
}

func (m *MockThirdpartyOAuth2ServiceRepository) Update(ctx context.Context, service *storage.ThirdpartyOAuth2Service) error {
	args := m.Called(ctx, service)
	return args.Error(0)
}

func (m *MockThirdpartyOAuth2ServiceRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockThirdpartyOAuth2ServiceRepository) List(ctx context.Context) ([]*storage.ThirdpartyOAuth2Service, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.ThirdpartyOAuth2Service), args.Error(1)
}

func (m *MockThirdpartyOAuth2ServiceRepository) CountGrantsReferencingService(ctx context.Context, serviceID string) (int, error) {
	args := m.Called(ctx, serviceID)
	return args.Int(0), args.Error(1)
}

// MockBranchKeyManager is a mock implementation of ports.BranchKeyManager
type MockBranchKeyManager struct {
	mock.Mock
}

func (m *MockBranchKeyManager) Create(ctx context.Context, serviceID string) (string, error) {
	args := m.Called(ctx, serviceID)
	return args.String(0), args.Error(1)
}

func TestNewAuthProvider(t *testing.T) {
	mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
	mockBranchKeyManager := new(MockBranchKeyManager)
	logger := slog.Default()

	t.Run("with all dependencies", func(t *testing.T) {
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)
		assert.NotNil(t, authProvider)
		assert.Equal(t, mockRepo, authProvider.serviceRepository)
		assert.Equal(t, mockBranchKeyManager, authProvider.branchKeyManager)
		assert.Equal(t, logger, authProvider.logger)
	})

	t.Run("with nil branch key manager", func(t *testing.T) {
		authProvider := NewAuthProvider(mockRepo, nil, logger)
		assert.NotNil(t, authProvider)
		assert.Equal(t, mockRepo, authProvider.serviceRepository)
		assert.Nil(t, authProvider.branchKeyManager)
		assert.Equal(t, logger, authProvider.logger)
	})

	t.Run("with nil logger", func(t *testing.T) {
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, nil)
		assert.NotNil(t, authProvider)
		assert.Equal(t, mockRepo, authProvider.serviceRepository)
		assert.Equal(t, mockBranchKeyManager, authProvider.branchKeyManager)
		assert.Equal(t, slog.Default(), authProvider.logger)
	})
}

func TestAuthProvider_Create(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		DisplayName:  "GitHub",
		ClientID:     "github-client-id",
		ClientSecret: "github-client-secret",
		IssuerURI:    "https://github.com",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("successful creation with branch key manager", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		// Mock branch key provisioning first
		mockBranchKeyManager.On("Create", ctx, "service-123").Return("service_service-123_branch_key", nil)

		// Then mock service creation
		mockRepo.On("Create", ctx, service).Return(nil)

		result, err := authProvider.Create(ctx, service)

		require.NoError(t, err)
		assert.Equal(t, service, result)

		// Verify branch key provisioning happened before service creation
		mockBranchKeyManager.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("successful creation without branch key manager", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		authProvider := NewAuthProvider(mockRepo, nil, logger)

		// Only mock service creation (no branch key provisioning)
		mockRepo.On("Create", ctx, service).Return(nil)

		result, err := authProvider.Create(ctx, service)

		require.NoError(t, err)
		assert.Equal(t, service, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("branch key provisioning fails", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		// Mock branch key provisioning failure
		branchKeyError := fmt.Errorf("failed to create branch key in DynamoDB")
		mockBranchKeyManager.On("Create", ctx, "service-123").Return("", branchKeyError)

		// Service creation should NOT be called if branch key provisioning fails
		// mockRepo.On("Create", ...) is NOT called

		result, err := authProvider.Create(ctx, service)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "branch key provisioning failed")
		assert.Contains(t, err.Error(), "failed to create branch key in DynamoDB")

		mockBranchKeyManager.AssertExpectations(t)
		// Service creation should not have been attempted
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("service creation fails after branch key provisioning succeeds", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		// Mock successful branch key provisioning
		mockBranchKeyManager.On("Create", ctx, "service-123").Return("service_service-123_branch_key", nil)

		// Mock service creation failure
		serviceError := storage.NewStorageError("CreateService", storage.ErrorKindValidation, nil, "validation failed")
		mockRepo.On("Create", ctx, service).Return(serviceError)

		result, err := authProvider.Create(ctx, service)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create service")
		assert.Contains(t, err.Error(), "validation failed")

		// Both operations should have been attempted
		mockBranchKeyManager.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthProvider_Get(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	service := &storage.ThirdpartyOAuth2Service{
		ID:          "service-123",
		DisplayName: "GitHub",
		ClientID:    "github-client-id",
	}

	t.Run("successful get", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		mockRepo.On("Get", ctx, "service-123").Return(service, nil)

		result, err := authProvider.Get(ctx, "service-123")

		require.NoError(t, err)
		assert.Equal(t, service, result)

		mockRepo.AssertExpectations(t)
	})

	t.Run("service not found", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		notFoundError := storage.NewStorageError("GetService", storage.ErrorKindNotFound, nil, "service not found")
		mockRepo.On("Get", ctx, "nonexistent").Return(nil, notFoundError)

		result, err := authProvider.Get(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &storage.StorageError{}, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthProvider_Update(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	service := &storage.ThirdpartyOAuth2Service{
		ID:          "service-123",
		DisplayName: "GitHub Updated",
		ClientID:    "github-client-id-updated",
	}

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		mockRepo.On("Update", ctx, service).Return(nil)

		err := authProvider.Update(ctx, service)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("update failure", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		updateError := storage.NewStorageError("UpdateService", storage.ErrorKindNotFound, nil, "service not found")
		mockRepo.On("Update", ctx, service).Return(updateError)

		err := authProvider.Update(ctx, service)

		assert.Error(t, err)
		assert.IsType(t, &storage.StorageError{}, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthProvider_Delete(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	t.Run("successful delete", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		mockRepo.On("Delete", ctx, "service-123").Return(nil)

		err := authProvider.Delete(ctx, "service-123")

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("delete blocked by grants", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		conflictError := storage.NewStorageError("DeleteService", storage.ErrorKindConflict, nil, "cannot delete service: 5 grants reference it")
		mockRepo.On("Delete", ctx, "service-123").Return(conflictError)

		err := authProvider.Delete(ctx, "service-123")

		assert.Error(t, err)
		assert.IsType(t, &storage.StorageError{}, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthProvider_List(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	services := []*storage.ThirdpartyOAuth2Service{
		{
			ID:          "service-1",
			DisplayName: "GitHub",
			ClientID:    "github-client",
		},
		{
			ID:          "service-2",
			DisplayName: "Google",
			ClientID:    "google-client",
		},
	}

	t.Run("successful list", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		mockRepo.On("List", ctx).Return(services, nil)

		result, err := authProvider.List(ctx)

		require.NoError(t, err)
		assert.Equal(t, services, result)
		assert.Len(t, result, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		emptyServices := []*storage.ThirdpartyOAuth2Service{}
		mockRepo.On("List", ctx).Return(emptyServices, nil)

		result, err := authProvider.List(ctx)

		require.NoError(t, err)
		assert.Equal(t, emptyServices, result)
		assert.Len(t, result, 0)

		mockRepo.AssertExpectations(t)
	})

	t.Run("list failure", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		listError := storage.NewStorageError("ListServices", storage.ErrorKindConnection, nil, "database connection failed")
		mockRepo.On("List", ctx).Return(nil, listError)

		result, err := authProvider.List(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.IsType(t, &storage.StorageError{}, err)

		mockRepo.AssertExpectations(t)
	})
}

// TestAuthProvider_AtomicBehavior tests the core architectural principle:
// branch key provisioning must happen before service creation, and if it fails,
// service creation must be skipped (fail-fast behavior).
func TestAuthProvider_AtomicBehavior(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "service-123",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	t.Run("atomic failure handling - branch key provisioning fails", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		// Simulate branch key provisioning failure
		branchKeyError := fmt.Errorf("KMS operation failed")
		mockBranchKeyManager.On("Create", ctx, "service-123").Return("", branchKeyError)

		// Service repository Create should NOT be called at all
		// This is the key behavior we're testing: atomic failure handling

		result, err := authProvider.Create(ctx, service)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "branch key provisioning failed")

		// Verify branch key manager was called
		mockBranchKeyManager.AssertExpectations(t)

		// Verify service repository was NOT called - this is the atomic behavior
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("proper orchestration order", func(t *testing.T) {
		mockRepo := new(MockThirdpartyOAuth2ServiceRepository)
		mockBranchKeyManager := new(MockBranchKeyManager)
		authProvider := NewAuthProvider(mockRepo, mockBranchKeyManager, logger)

		// Set up call order tracking
		callOrder := make([]string, 0)

		mockBranchKeyManager.On("Create", ctx, "service-123").Return("service_service-123_branch_key", nil).Run(func(args mock.Arguments) {
			callOrder = append(callOrder, "branch_key_create")
		})

		mockRepo.On("Create", ctx, service).Return(nil).Run(func(args mock.Arguments) {
			callOrder = append(callOrder, "service_create")
		})

		result, err := authProvider.Create(ctx, service)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify correct call order: branch key provisioning before service creation
		assert.Equal(t, []string{"branch_key_create", "service_create"}, callOrder)

		mockBranchKeyManager.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
