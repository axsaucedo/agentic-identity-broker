package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T) *storage.AuthorizationSession {
	t.Helper()
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	session, err := storage.NewAuthorizationSession(agentID, "test@example.com", "https://example.com/client", "https://example.com/authorize", "", "", "", "", "", nil)
	require.NoError(t, err)
	return session
}

func TestAuthorizationSessionRepository_CreateValidation(t *testing.T) {
	t.Run("returns validation error for nil session", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()

		err := repo.Create(context.Background(), nil)
		require.Error(t, err)
		var storErr *storage.StorageError
		require.ErrorAs(t, err, &storErr)
		assert.Equal(t, storage.ErrorKindValidation, storErr.Kind)
	})

	t.Run("returns validation error for empty SessionID", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		session.SessionID = ""

		err := repo.Create(context.Background(), session)
		require.Error(t, err)
		var storErr *storage.StorageError
		require.ErrorAs(t, err, &storErr)
		assert.Equal(t, storage.ErrorKindValidation, storErr.Kind)
	})
}

func TestAuthorizationSessionRepository_Consume(t *testing.T) {
	t.Run("succeeds on first consume", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		require.NoError(t, repo.Create(context.Background(), session))

		err := repo.Consume(context.Background(), session.SessionID)
		require.NoError(t, err)
	})

	t.Run("returns conflict on second consume", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		require.NoError(t, repo.Create(context.Background(), session))
		require.NoError(t, repo.Consume(context.Background(), session.SessionID))

		err := repo.Consume(context.Background(), session.SessionID)
		require.Error(t, err)
		var storErr *storage.StorageError
		require.ErrorAs(t, err, &storErr)
		assert.Equal(t, storage.ErrorKindConflict, storErr.Kind)
	})

	t.Run("returns not found for unknown session", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()

		err := repo.Consume(context.Background(), "nonexistent")
		require.Error(t, err)
		var storErr *storage.StorageError
		require.ErrorAs(t, err, &storErr)
		assert.Equal(t, storage.ErrorKindNotFound, storErr.Kind)
	})

	t.Run("concurrent consume: exactly one succeeds", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		require.NoError(t, repo.Create(context.Background(), session))

		const goroutines = 20
		results := make([]error, goroutines)
		var wg sync.WaitGroup
		start := make(chan struct{})

		for i := 0; i < goroutines; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				results[i] = repo.Consume(context.Background(), session.SessionID)
			}()
		}

		// Release all goroutines simultaneously to maximize contention.
		time.Sleep(1 * time.Millisecond)
		close(start)
		wg.Wait()

		successes := 0
		for _, err := range results {
			if err == nil {
				successes++
			}
		}
		assert.Equal(t, 1, successes, "exactly one concurrent consume should succeed")
	})
}

func TestAuthorizationSessionRepository_ConsumeIf(t *testing.T) {
	t.Run("marks session consumed after successful callback", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		require.NoError(t, repo.Create(context.Background(), session))

		err := repo.ConsumeIf(context.Background(), session.SessionID, func(context.Context) error {
			return nil
		})
		require.NoError(t, err)

		stored, err := repo.GetBySessionID(context.Background(), session.SessionID)
		require.NoError(t, err)
		assert.True(t, stored.IsConsumed())
	})

	t.Run("leaves session reusable when callback fails", func(t *testing.T) {
		repo := NewAuthorizationSessionRepository()
		session := newTestSession(t)
		require.NoError(t, repo.Create(context.Background(), session))

		err := repo.ConsumeIf(context.Background(), session.SessionID, func(context.Context) error {
			return assert.AnError
		})
		require.ErrorIs(t, err, assert.AnError)

		stored, err := repo.GetBySessionID(context.Background(), session.SessionID)
		require.NoError(t, err)
		assert.False(t, stored.IsConsumed())

		require.NoError(t, repo.Consume(context.Background(), session.SessionID))
	})
}

func TestAuthorizationSessionRepository_CreateDeepCopiesCIMDMetadata(t *testing.T) {
	repo := NewAuthorizationSessionRepository()
	session := newTestSession(t)
	session.CIMDMetadata = &storage.CIMDMetadataSnapshot{
		ClientID:     "https://agent.example.com/client",
		RedirectURIs: []string{"https://agent.example.com/callback"},
	}

	require.NoError(t, repo.Create(context.Background(), session))

	session.CIMDMetadata.RedirectURIs[0] = "https://attacker.example.com/callback"

	stored, err := repo.GetBySessionID(context.Background(), session.SessionID)
	require.NoError(t, err)
	require.NotNil(t, stored.CIMDMetadata)
	assert.Equal(t, "https://agent.example.com/callback", stored.CIMDMetadata.RedirectURIs[0])
}

func TestAuthorizationSessionRepository_GetBySessionIDReturnsDeepCopyOfCIMDMetadata(t *testing.T) {
	repo := NewAuthorizationSessionRepository()
	session := newTestSession(t)
	session.CIMDMetadata = &storage.CIMDMetadataSnapshot{
		ClientID:     "https://agent.example.com/client",
		RedirectURIs: []string{"https://agent.example.com/callback"},
	}

	require.NoError(t, repo.Create(context.Background(), session))

	first, err := repo.GetBySessionID(context.Background(), session.SessionID)
	require.NoError(t, err)
	first.CIMDMetadata.RedirectURIs[0] = "https://attacker.example.com/callback"

	second, err := repo.GetBySessionID(context.Background(), session.SessionID)
	require.NoError(t, err)
	require.NotNil(t, second.CIMDMetadata)
	assert.Equal(t, "https://agent.example.com/callback", second.CIMDMetadata.RedirectURIs[0])
}
