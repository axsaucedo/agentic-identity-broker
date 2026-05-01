//go:build integration
// +build integration

package storage_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/postgres"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// applyAllMigrations applies all migrations using go-migrate.
func applyAllMigrations(t *testing.T, connStr string) {
	t.Helper()

	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	absDir, err := filepath.Abs(filepath.Join(projectRoot, "migrations"))
	require.NoError(t, err)

	m, err := migrate.New("file://"+absDir, connStr)
	require.NoError(t, err, "failed to create migrate instance")
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migration failed: %v", err)
	}
}

// newAuthSessionRepo creates a configured postgres AuthorizationSessionRepo for tests.
func newAuthSessionRepo(t *testing.T, connStr string, agentIDs ...id.AgentID) (ports.AuthorizationSessionRepository, func()) {
	t.Helper()

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, adapter.Initialize(ctx))

	agentRepo := postgres.NewAgentRepository(adapter)
	now := time.Now().UTC()
	for _, agentID := range agentIDs {
		require.NoError(t, agentRepo.Create(ctx, &storage.Agent{
			ID:          agentID,
			ClientID:    id.ClientID("auth-session-test-" + agentID.String()[:8]),
			DisplayName: "Authorization Session Test Agent",
			Description: "Agent for authorization session integration tests",
			CreatedAt:   now,
			UpdatedAt:   now,
		}))
	}

	return postgres.NewAuthorizationSessionRepo(adapter), func() { adapter.Close(ctx) }
}

// newTestSession builds a valid AuthorizationSession for test fixtures.
func newTestSession(t *testing.T, agentID id.AgentID) *storage.AuthorizationSession {
	t.Helper()
	meta := &storage.CIMDMetadataSnapshot{
		ClientID:     "https://agent.example.com/.well-known/client",
		ClientName:   "Test Agent",
		RedirectURIs: []string{"https://agent.example.com/callback"},
	}
	sess, err := storage.NewAuthorizationSession(
		agentID,
		"test@example.com",
		"https://agent.example.com/.well-known/client",
		"https://broker.example.com/oauth2/authorize?client_id=...",
		"https://agent.example.com/callback",
		"read write",
		"state-abc",
		"challenge-xyz",
		"S256",
		meta,
	)
	require.NoError(t, err)
	return sess
}

func TestAuthorizationSessionRepo_CreateAndGet(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	agentID := id.AgentID(uuid.New())
	repo, closeRepo := newAuthSessionRepo(t, connStr, agentID)
	defer closeRepo()

	ctx := context.Background()
	sess := newTestSession(t, agentID)

	require.NoError(t, repo.Create(ctx, sess))

	got, err := repo.GetBySessionID(ctx, sess.SessionID)
	require.NoError(t, err)
	assert.Equal(t, sess.SessionID, got.SessionID)
	assert.Equal(t, agentID, got.AgentID)
	assert.Equal(t, sess.ClientID, got.ClientID)
	assert.Equal(t, sess.RedirectURI, got.RedirectURI)
	assert.Equal(t, sess.Scope, got.Scope)
	assert.Equal(t, sess.State, got.State)
	assert.Equal(t, sess.CodeChallenge, got.CodeChallenge)
	assert.Equal(t, sess.CodeChallengeMethod, got.CodeChallengeMethod)
	assert.Nil(t, got.ConsumedAt, "new session must not be consumed")
	assert.False(t, got.IsExpired(), "new session must not be expired")
	require.NotNil(t, got.CIMDMetadata)
	assert.Equal(t, "Test Agent", got.CIMDMetadata.ClientName)
}

func TestAuthorizationSessionRepo_GetNotFound(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	repo, closeRepo := newAuthSessionRepo(t, connStr)
	defer closeRepo()

	ctx := context.Background()
	_, err := repo.GetBySessionID(ctx, "nonexistent-session-id")
	require.Error(t, err)

	var storageErr *storage.StorageError
	require.ErrorAs(t, err, &storageErr)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestAuthorizationSessionRepo_GetExpiredSession(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	agentID := id.AgentID(uuid.New())
	repo, closeRepo := newAuthSessionRepo(t, connStr, agentID)
	defer closeRepo()

	ctx := context.Background()
	sess := newTestSession(t, agentID)

	// Force expiry before inserting
	sess.ExpiresAt = time.Now().Add(-1 * time.Hour)
	require.NoError(t, repo.Create(ctx, sess))

	// Repo still returns the session — callers check IsExpired()
	got, err := repo.GetBySessionID(ctx, sess.SessionID)
	require.NoError(t, err)
	assert.True(t, got.IsExpired(), "session should report expired")
}

func TestAuthorizationSessionRepo_Consume(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	agentID := id.AgentID(uuid.New())
	repo, closeRepo := newAuthSessionRepo(t, connStr, agentID)
	defer closeRepo()

	ctx := context.Background()
	sess := newTestSession(t, agentID)
	require.NoError(t, repo.Create(ctx, sess))

	require.NoError(t, repo.Consume(ctx, sess.SessionID))

	got, err := repo.GetBySessionID(ctx, sess.SessionID)
	require.NoError(t, err)
	assert.True(t, got.IsConsumed(), "session should be consumed after Consume()")
	assert.NotNil(t, got.ConsumedAt)
}

func TestAuthorizationSessionRepo_ConsumeAlreadyConsumed(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	agentID := id.AgentID(uuid.New())
	repo, closeRepo := newAuthSessionRepo(t, connStr, agentID)
	defer closeRepo()

	ctx := context.Background()
	sess := newTestSession(t, agentID)
	require.NoError(t, repo.Create(ctx, sess))
	require.NoError(t, repo.Consume(ctx, sess.SessionID))

	// Second consume must fail — session is already consumed
	err := repo.Consume(ctx, sess.SessionID)
	require.Error(t, err, "consuming an already-consumed session must return error")
}

func TestAuthorizationSessionRepo_DeleteExpired(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()
	_ = container

	applyAllMigrations(t, connStr)

	agentID := id.AgentID(uuid.New())
	repo, closeRepo := newAuthSessionRepo(t, connStr, agentID)
	defer closeRepo()

	ctx := context.Background()

	// Create an expired session
	expired := newTestSession(t, agentID)
	expired.ExpiresAt = time.Now().Add(-1 * time.Hour)
	require.NoError(t, repo.Create(ctx, expired))

	// Create a valid session
	valid := newTestSession(t, agentID)
	require.NoError(t, repo.Create(ctx, valid))

	deleted, err := repo.DeleteExpired(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted, "only the expired session should be deleted")

	// Expired session is gone
	_, err = repo.GetBySessionID(ctx, expired.SessionID)
	require.Error(t, err)
	var storageErr *storage.StorageError
	require.ErrorAs(t, err, &storageErr)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Valid session still exists
	got, err := repo.GetBySessionID(ctx, valid.SessionID)
	require.NoError(t, err)
	assert.Equal(t, valid.SessionID, got.SessionID)
}
