//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresThirdpartyOAuth2ProviderRepository_ProtectedResources(t *testing.T) {
	adapter, cleanup := setupMigratedAdapter(t)
	defer cleanup()

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	ctx := context.Background()
	provider := newTestEntity()
	provider.ID = id.NewServiceID()
	provider.ProtectedResources = nil

	require.NoError(t, repo.Create(ctx, provider))
	assert.Equal(t, int64(1), provider.Version)

	resources, version, err := repo.ListProtectedResources(ctx, provider.ID)
	require.NoError(t, err)
	assert.Empty(t, resources)
	assert.Equal(t, int64(1), version)

	added, err := repo.AddProtectedResource(ctx, provider.ID, "https://api.example.com/added")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/added", added.Resource)
	assert.Equal(t, []string{"https://api.example.com/added"}, added.ProtectedResources)
	assert.Equal(t, int64(2), added.Version)
	assert.True(t, added.Changed)

	resources, version, err = repo.ListProtectedResources(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, added.ProtectedResources, resources)
	assert.Equal(t, added.Version, version)

	replay, err := repo.AddProtectedResource(ctx, provider.ID, "https://api.example.com/added")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/added", replay.Resource)
	assert.Equal(t, []string{"https://api.example.com/added"}, replay.ProtectedResources)
	assert.Equal(t, int64(2), replay.Version)
	assert.False(t, replay.Changed)

	renamed, err := repo.RenameProtectedResource(ctx, provider.ID, "https://api.example.com/added", "https://api.example.com/renamed")
	require.NoError(t, err)
	assert.Equal(t, "https://api.example.com/renamed", renamed.Resource)
	assert.Equal(t, []string{"https://api.example.com/renamed"}, renamed.ProtectedResources)
	assert.Equal(t, int64(3), renamed.Version)
	assert.True(t, renamed.Changed)

	other := newTestEntity()
	other.ID = id.NewServiceID()
	other.ProtectedResources = nil
	require.NoError(t, repo.Create(ctx, other))

	_, err = repo.AddProtectedResource(ctx, other.ID, "https://api.example.com/renamed")
	require.Error(t, err)
	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindConflict, storageErr.Kind)

	removed, err := repo.RemoveProtectedResource(ctx, provider.ID, "https://api.example.com/renamed")
	require.NoError(t, err)
	assert.Empty(t, removed.ProtectedResources)
	assert.Equal(t, int64(4), removed.Version)
	assert.True(t, removed.Changed)

	resources, version, err = repo.ListProtectedResources(ctx, provider.ID)
	require.NoError(t, err)
	assert.Empty(t, resources)
	assert.Equal(t, removed.Version, version)
}

func TestPostgresThirdpartyOAuth2ProviderRepository_UpdateCAS(t *testing.T) {
	adapter, cleanup := setupMigratedAdapter(t)
	defer cleanup()

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	ctx := context.Background()
	provider := newTestEntity()
	provider.ID = id.NewServiceID()
	provider.ProtectedResources = []string{"https://api.example.com/original"}
	require.NoError(t, repo.Create(ctx, provider))

	expectedVersion := provider.Version
	provider.DisplayName = "Updated Provider"
	provider.ProtectedResources = []string{"https://api.example.com/replaced"}
	require.NoError(t, repo.Update(ctx, provider, &expectedVersion))
	assert.Equal(t, int64(2), provider.Version)

	resources, version, err := repo.ListProtectedResources(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://api.example.com/replaced"}, resources)
	assert.Equal(t, provider.Version, version)

	staleVersion := expectedVersion
	provider.ProtectedResources = []string{"https://api.example.com/stale"}
	err = repo.Update(ctx, provider, &staleVersion)
	require.Error(t, err)
	var storageErr *storage.StorageError
	require.True(t, errors.As(err, &storageErr))
	assert.Equal(t, storage.ErrorKindConflict, storageErr.Kind)

	provider.ProtectedResources = []string{"https://api.example.com/ignored"}
	require.NoError(t, repo.Update(ctx, provider, nil))
	resources, version, err = repo.ListProtectedResources(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"https://api.example.com/replaced"}, resources)
	assert.Equal(t, int64(3), version)
}

func TestPostgresThirdpartyOAuth2ProviderRepository_GetUsesChildResourcesAndEncryptedSecret(t *testing.T) {
	adapter, cleanup := setupMigratedAdapter(t)
	defer cleanup()

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	ctx := context.Background()
	provider := newTestEntity()
	provider.ID = id.NewServiceID()
	provider.ProtectedResources = []string{"https://api.example.com/child"}
	require.NoError(t, repo.Create(ctx, provider))

	stored, err := repo.Get(ctx, provider.ID)
	require.NoError(t, err)
	assert.Equal(t, provider.ProtectedResources, stored.ProtectedResources)
	assert.Equal(t, int64(1), stored.Version)
	assert.True(t, stored.Secret.IsEncrypted())
	ciphertext, err := stored.Secret.GetCiphertext()
	require.NoError(t, err)
	assert.Equal(t, testCiphertext, ciphertext)
}

func TestPostgresThirdpartyOAuth2ProviderRepository_ChildResourceResolverLifecycle(t *testing.T) {
	adapter, cleanup := setupMigratedAdapter(t)
	defer cleanup()

	repo := NewPostgresThirdpartyOAuth2ProviderRepository(adapter)
	ctx := context.Background()
	provider := newTestEntity()
	provider.ID = id.NewServiceID()
	provider.ProtectedResources = nil
	require.NoError(t, repo.Create(ctx, provider))

	const resource = "https://api.example.com/resolver"
	_, err := repo.AddProtectedResource(ctx, provider.ID, resource)
	require.NoError(t, err)
	resolved, err := repo.FindByProtectedResource(ctx, resource)
	require.NoError(t, err)
	assert.Equal(t, provider.ID, resolved.ID)

	_, err = repo.RemoveProtectedResource(ctx, provider.ID, resource)
	require.NoError(t, err)
	_, err = repo.FindByProtectedResource(ctx, resource)
	require.Error(t, err)
	assert.True(t, tokenexchange.IsResourceNotConfigured(err))
}
