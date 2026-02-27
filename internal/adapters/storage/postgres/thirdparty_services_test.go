//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/require"
)

// testKEKForServiceTests is the deterministic test KEK used for encryption in integration tests.
// This is the base64 encoding of known test bytes — NOT for production use.
const testKEKForServiceTests = "ASNFZ4mrze/+3LqYdlQyEAEjRWeJq83v/ty6mHZUMhA="

// newTestEncryptionForServices creates a real memory encryption adapter using the deterministic KEK.
func newTestEncryptionForServices(t *testing.T) ports.EncryptionPort {
	t.Helper()
	enc, _, err := awsencryption.NewAWSEncryption(testKEKForServiceTests, "", 0)
	require.NoError(t, err, "failed to create test encryption adapter")
	return enc
}

func TestThirdpartyServiceRepository_Create(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	// Create adapter
	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	// Verify service was stored and secret is decryptable
	retrieved, err := providerService.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Test Service", retrieved.DisplayName)
	p, err := retrieved.Secret.GetPlaintext()
	require.NoError(t, err)
	require.Equal(t, "test-secret", p)
	require.Len(t, retrieved.Scopes, 1)
	require.Equal(t, "read", retrieved.Scopes[0].ScopeValue)
}

func TestThirdpartyServiceRepository_Create_GeneratesID(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		// No ID provided — service should generate one
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)
	require.NotEmpty(t, entity.ID, "expected ID to be generated")
}

func TestThirdpartyServiceRepository_Create_DuplicateID(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create first service
	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	// Re-set secret to plaintext for second create attempt (first create changes it to encrypted)
	entity.Secret = model.NewPlaintextSecret("test-secret")

	// Try to create with same ID — must fail
	err = providerService.Create(ctx, entity)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr), "expected StorageError, got: %T: %v", err, err)
	require.Equal(t, storage.ErrorKindConflict, storageErr.Kind)
}

func TestThirdpartyServiceRepository_Get(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: true,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
			{ScopeValue: "write", Description: "Write access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	// Retrieve via service to properly decrypt secret
	retrieved, err := providerService.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "test-service-1", retrieved.ID)
	require.Equal(t, "Test Service", retrieved.DisplayName)
	p, err := retrieved.Secret.GetPlaintext()
	require.NoError(t, err)
	require.Equal(t, "test-secret", p)
	require.True(t, retrieved.Discovery.EnableDiscovery)
	require.Len(t, retrieved.Scopes, 2)
}

func TestThirdpartyServiceRepository_Get_NotFound(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)

	_, err = repo.Get(ctx, "non-existent")
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr), "expected StorageError")
	require.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestThirdpartyServiceRepository_Update(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	// Update service — set plaintext secret so service encrypts the new value
	entity.DisplayName = "Updated Service"
	entity.Secret = model.NewPlaintextSecret("new-secret")
	entity.Scopes = []model.OAuthScope{
		{ScopeValue: "read", Description: "Read access"},
		{ScopeValue: "write", Description: "Write access"},
	}

	err = providerService.Update(ctx, entity)
	require.NoError(t, err)

	// Verify update via service to properly decrypt secret
	retrieved, err := providerService.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Updated Service", retrieved.DisplayName)
	p, err := retrieved.Secret.GetPlaintext()
	require.NoError(t, err)
	require.Equal(t, "new-secret", p)
	require.Len(t, retrieved.Scopes, 2)
}

func TestThirdpartyServiceRepository_Update_NotFound(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "non-existent",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Update(ctx, entity)
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr), "expected StorageError")
	require.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestThirdpartyServiceRepository_Delete(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	err = providerService.Delete(ctx, "test-service-1")
	require.NoError(t, err)

	// Verify service was deleted
	_, err = providerService.Get(ctx, "test-service-1")
	require.Error(t, err)
}

func TestThirdpartyServiceRepository_Delete_Idempotent(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)

	// Delete non-existent service (should not error)
	err = repo.Delete(ctx, "non-existent")
	require.NoError(t, err)
}

func TestThirdpartyServiceRepository_List(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	// List empty repository
	services, err := providerService.List(ctx)
	require.NoError(t, err)
	require.Len(t, services, 0)

	// Add services
	for i := 1; i <= 3; i++ {
		entity := &model.ThirdpartyOAuth2ProviderEntity{
			DisplayName: "Test Service",
			ClientID:    "test-client-id",
			Secret:      model.NewPlaintextSecret("test-secret"),
			IssuerURI:   "https://oauth.example.com",
			Discovery: model.DiscoveryConfig{
				EnableDiscovery: false,
			},
			Endpoints: model.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth.example.com/token",
				AuthorizeEndpoint: "https://oauth.example.com/authorize",
			},
			Scopes: []model.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = providerService.Create(ctx, entity)
		require.NoError(t, err)
	}

	// List via service to get properly decrypted services
	services, err = providerService.List(ctx)
	require.NoError(t, err)
	require.Len(t, services, 3)
}

func TestThirdpartyServiceRepository_CountGrantsReferencingService(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)

	// Without grants, count should be 0
	count, err := repo.CountGrantsReferencingService(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func TestThirdpartyServiceRepository_DeepCopyProtection(t *testing.T) {
	container, connStr, cleanup := setupTestContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := newTestEncryptionForServices(t)
	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	providerService := thirdparty.NewThirdpartyOAuth2ProviderService(repo, encryption, nil, slog.Default())

	entity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          "test-service-1",
		DisplayName: "Test Service",
		ClientID:    "test-client-id",
		Secret:      model.NewPlaintextSecret("test-secret"),
		IssuerURI:   "https://oauth.example.com",
		Discovery: model.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: model.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []model.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = providerService.Create(ctx, entity)
	require.NoError(t, err)

	// Get service and modify it
	retrieved, err := providerService.Get(ctx, "test-service-1")
	require.NoError(t, err)

	retrieved.DisplayName = "Modified"
	retrieved.Scopes[0].ScopeValue = "write"

	// Verify original is unchanged (deep copy protection)
	original, err := providerService.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Test Service", original.DisplayName)
	require.Equal(t, "read", original.Scopes[0].ScopeValue)
}
