//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/require"
)

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

	// Create repository with no-op encryption
	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Verify service was stored
	retrieved, err := repo.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Test Service", retrieved.DisplayName)
	require.Equal(t, "test-secret", retrieved.ClientSecret)
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		// No ID provided
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)
	require.NotEmpty(t, service.ID, "expected ID to be generated")
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Try to create with same ID
	err = repo.Create(ctx, service)
	require.Error(t, err)

	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok, "expected StorageError")
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: true,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
			{ScopeValue: "write", Description: "Write access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "test-service-1", retrieved.ID)
	require.Equal(t, "Test Service", retrieved.DisplayName)
	require.Equal(t, "test-secret", retrieved.ClientSecret)
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	_, err = repo.Get(ctx, "non-existent")
	require.Error(t, err)

	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok, "expected StorageError")
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Update service
	service.DisplayName = "Updated Service"
	service.ClientSecret = "new-secret"
	service.Scopes = []storage.OAuthScope{
		{ScopeValue: "read", Description: "Read access"},
		{ScopeValue: "write", Description: "Write access"},
	}

	err = repo.Update(ctx, service)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Updated Service", retrieved.DisplayName)
	require.Equal(t, "new-secret", retrieved.ClientSecret)
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "non-existent",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Update(ctx, service)
	require.Error(t, err)

	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok, "expected StorageError")
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	err = repo.Delete(ctx, "test-service-1")
	require.NoError(t, err)

	// Verify service was deleted
	_, err = repo.Get(ctx, "test-service-1")
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	// List empty repository
	services, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, services, 0)

	// Add services
	for i := 1; i <= 3; i++ {
		service := &storage.ThirdpartyOAuth2Service{
			DisplayName:  "Test Service",
			ClientID:     "test-client-id",
			ClientSecret: "test-secret",
			IssuerURI:    "https://oauth.example.com",
			Discovery: storage.DiscoveryConfig{
				EnableDiscovery: false,
			},
			Endpoints: storage.OAuth2Endpoints{
				TokenEndpoint:     "https://oauth.example.com/token",
				AuthorizeEndpoint: "https://oauth.example.com/authorize",
			},
			Scopes: []storage.OAuthScope{
				{ScopeValue: "read", Description: "Read access"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = repo.Create(ctx, service)
		require.NoError(t, err)
	}

	services, err = repo.List(ctx)
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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

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

	encryption := noop.NewNoOpEncryption()
	repo := NewThirdpartyServiceRepository(adapter, encryption)

	service := &storage.ThirdpartyOAuth2Service{
		ID:           "test-service-1",
		DisplayName:  "Test Service",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Get service and modify it
	retrieved, err := repo.Get(ctx, "test-service-1")
	require.NoError(t, err)

	retrieved.DisplayName = "Modified"
	retrieved.Scopes[0].ScopeValue = "write"

	// Verify original is unchanged
	original, err := repo.Get(ctx, "test-service-1")
	require.NoError(t, err)
	require.Equal(t, "Test Service", original.DisplayName)
	require.Equal(t, "read", original.Scopes[0].ScopeValue)
}
