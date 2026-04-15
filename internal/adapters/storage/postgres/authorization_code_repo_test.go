//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthCodeTestDB(t *testing.T) (*Adapter, func()) {
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

func TestAuthorizationCodeRepo_Create(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	// Create agent first (FK constraint)
	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)

	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      "sha256hashvalue1234567890abcdef",
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read write",
		ExpiresAt:     time.Now().UTC().Add(60 * time.Second),
		CreatedAt:     time.Now().UTC(),
	}

	err := repo.Create(context.Background(), code)
	require.NoError(t, err)
}

func TestAuthorizationCodeRepo_FindByCodeHash(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)
	ctx := context.Background()

	codeHash := "findbyhash" + id.NewAuthorizationCodeID().String()[:10]
	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      codeHash,
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().UTC().Add(60 * time.Second),
		CreatedAt:     time.Now().UTC(),
	}
	err := repo.Create(ctx, code)
	require.NoError(t, err)

	got, err := repo.FindByCodeHash(ctx, codeHash)
	require.NoError(t, err)
	assert.Equal(t, codeHash, got.CodeHash)
	assert.Equal(t, agent.ID, got.AgentID)
	assert.Equal(t, "user@example.com", got.Principal.String())
}

func TestAuthorizationCodeRepo_MarkUsed(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)
	ctx := context.Background()

	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      "markused" + id.NewAuthorizationCodeID().String()[:10],
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().UTC().Add(60 * time.Second),
		CreatedAt:     time.Now().UTC(),
	}
	err := repo.Create(ctx, code)
	require.NoError(t, err)

	err = repo.MarkUsed(ctx, code.ID)
	require.NoError(t, err)

	// Verify it's marked as used
	got, err := repo.FindByCodeHash(ctx, code.CodeHash)
	require.NoError(t, err)
	assert.NotNil(t, got.UsedAt)
}

func TestAuthorizationCodeRepo_UniqueCodeHash(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)
	ctx := context.Background()

	codeHash := "uniquehash" + id.NewAuthorizationCodeID().String()[:10]
	code1 := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      codeHash,
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().UTC().Add(60 * time.Second),
		CreatedAt:     time.Now().UTC(),
	}
	err := repo.Create(ctx, code1)
	require.NoError(t, err)

	code2 := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      codeHash, // same hash
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user2@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge2",
		Scope:         "write",
		ExpiresAt:     time.Now().UTC().Add(60 * time.Second),
		CreatedAt:     time.Now().UTC(),
	}
	err = repo.Create(ctx, code2)
	assert.Error(t, err, "duplicate code_hash should fail unique constraint")
}

func TestAuthorizationCodeRepo_FindByCodeHash_ExpiredUnused(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)
	ctx := context.Background()

	codeHash := "expired-unused-" + id.NewAuthorizationCodeID().String()[:10]
	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      codeHash,
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().UTC().Add(-10 * time.Minute), // expired
		CreatedAt:     time.Now().UTC().Add(-15 * time.Minute),
	}
	err := repo.Create(ctx, code)
	require.NoError(t, err)

	got, err := repo.FindByCodeHash(ctx, codeHash)
	require.NoError(t, err)
	assert.Equal(t, code.ID, got.ID)
	assert.Nil(t, got.UsedAt, "repo must not filter by expiry — domain layer owns that check")
}

func TestAuthorizationCodeRepo_FindByCodeHash_NotFound(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	repo := NewAuthorizationCodeRepo(adapter)
	_, err := repo.FindByCodeHash(context.Background(), "nonexistentcodehash")
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindNotFound, se.Kind)
}

func TestAuthorizationCodeRepo_FindByCodeHash_Timeout(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	repo := NewAuthorizationCodeRepo(adapter)

	// Use an already-expired deadline so the query immediately gets context.DeadlineExceeded.
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := repo.FindByCodeHash(ctx, "somehash")
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindTimeout, se.Kind)
}

func TestAuthorizationCodeRepo_DeleteExpired(t *testing.T) {
	adapter, cleanup := setupAuthCodeTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewAuthorizationCodeRepo(adapter)
	ctx := context.Background()

	// Create an expired code
	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      "expired" + id.NewAuthorizationCodeID().String()[:10],
		AgentID:       agent.ID,
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost:9999/callback",
		CodeChallenge: "S256challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().UTC().Add(-10 * time.Second), // already expired
		CreatedAt:     time.Now().UTC().Add(-70 * time.Second),
	}
	err := repo.Create(ctx, code)
	require.NoError(t, err)

	count, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
}
