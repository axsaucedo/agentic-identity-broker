//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSigningKeyTestDB(t *testing.T) (*Adapter, func()) {
	t.Helper()
	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)
	applyMigrations(t, container)
	config := testStorageConfig(connString)
	adapter, err := NewAdapter(config)
	require.NoError(t, err)
	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	return adapter, func() {
		adapter.Close(ctx)
		cleanup()
	}
}

func TestSigningKeyRepo_Create(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)

	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-create-test"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("encrypted-key-material"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}

	err := repo.Create(context.Background(), key)
	require.NoError(t, err)
}

func TestSigningKeyRepo_GetByKID(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	kid := id.NewKeyID("kid-get-test")
	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 kid,
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("encrypted-key-material"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}
	err := repo.Create(ctx, key)
	require.NoError(t, err)

	got, err := repo.GetByKID(ctx, kid)
	require.NoError(t, err)
	assert.Equal(t, kid, got.KID)
	assert.Equal(t, "ES256", got.Algorithm)
	assert.True(t, got.IsCurrent)
}

func TestSigningKeyRepo_GetCurrent(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-current"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("encrypted-key-material"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}
	err := repo.Create(ctx, key)
	require.NoError(t, err)

	got, err := repo.GetCurrent(ctx)
	require.NoError(t, err)
	assert.Equal(t, key.KID, got.KID)
	assert.True(t, got.IsCurrent)
}

func TestSigningKeyRepo_ListActive(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	// Create two active keys
	key1 := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-list-1"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("key1"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}
	key2 := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-list-2"),
		Algorithm:           "RS256",
		PrivateKeyEncrypted: []byte("key2"),
		IsCurrent:           false,
		CreatedAt:           time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, key1))
	require.NoError(t, repo.Create(ctx, key2))

	active, err := repo.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 2)
}

func TestSigningKeyRepo_SetCurrent(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	key1 := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-set-1"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("key1"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}
	key2 := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-set-2"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("key2"),
		IsCurrent:           false,
		CreatedAt:           time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, key1))
	require.NoError(t, repo.Create(ctx, key2))

	// Promote key2
	err := repo.SetCurrent(ctx, key2.KID)
	require.NoError(t, err)

	// Verify key2 is now current
	got, err := repo.GetCurrent(ctx)
	require.NoError(t, err)
	assert.Equal(t, key2.KID, got.KID)

	// Verify key1 is no longer current
	got1, err := repo.GetByKID(ctx, key1.KID)
	require.NoError(t, err)
	assert.False(t, got1.IsCurrent)
}

func TestSigningKeyRepo_Delete(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-delete"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("key"),
		IsCurrent:           false,
		CreatedAt:           time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, key))

	err := repo.Delete(ctx, key.KID)
	require.NoError(t, err)

	// Should not appear in active list
	active, err := repo.ListActive(ctx)
	require.NoError(t, err)
	for _, k := range active {
		assert.NotEqual(t, key.KID, k.KID)
	}
}

func TestSigningKeyRepo_CountActive(t *testing.T) {
	adapter, cleanup := setupSigningKeyTestDB(t)
	defer cleanup()

	repo := NewSigningKeyRepo(adapter)
	ctx := context.Background()

	key := &storage.SigningKey{
		ID:                  id.NewSigningKeyID(),
		KID:                 id.NewKeyID("kid-count"),
		Algorithm:           "ES256",
		PrivateKeyEncrypted: []byte("key"),
		IsCurrent:           true,
		CreatedAt:           time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, key))

	count, err := repo.CountActive(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
