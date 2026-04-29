package storage

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestSession(t *testing.T) *AuthorizationSession {
	t.Helper()
	s, err := NewAuthorizationSession(
		id.MustParseAgentID("00000000-0000-0000-0000-000000000001"),
		"user@example.com",
		"client-1",
		"https://example.com/original",
		"https://example.com/cb",
		"openid",
		"state-abc",
		"challenge",
		"S256",
		nil,
	)
	require.NoError(t, err)
	return s
}

func TestAuthorizationSession_IsExpired(t *testing.T) {
	t.Run("fresh session is not expired", func(t *testing.T) {
		s := newTestSession(t)
		assert.False(t, s.IsExpired())
	})

	t.Run("session past ExpiresAt is expired", func(t *testing.T) {
		s := newTestSession(t)
		s.ExpiresAt = time.Now().Add(-time.Second)
		assert.True(t, s.IsExpired())
	})

	t.Run("exact boundary — ExpiresAt equal to now is not expired", func(t *testing.T) {
		s := newTestSession(t)
		s.ExpiresAt = time.Now().Add(time.Millisecond)
		assert.False(t, s.IsExpired())
	})
}

func TestAuthorizationSession_IsConsumed(t *testing.T) {
	t.Run("fresh session is not consumed", func(t *testing.T) {
		s := newTestSession(t)
		assert.False(t, s.IsConsumed())
	})

	t.Run("consumed session reports true", func(t *testing.T) {
		s := newTestSession(t)
		s.Consume()
		assert.True(t, s.IsConsumed())
	})
}

func TestAuthorizationSession_Consume(t *testing.T) {
	t.Run("sets ConsumedAt on first call", func(t *testing.T) {
		s := newTestSession(t)
		require.Nil(t, s.ConsumedAt)

		s.Consume()
		require.NotNil(t, s.ConsumedAt)
	})

	t.Run("idempotent — second call does not overwrite ConsumedAt", func(t *testing.T) {
		s := newTestSession(t)
		s.Consume()
		first := *s.ConsumedAt

		time.Sleep(time.Millisecond)
		s.Consume()
		assert.Equal(t, first, *s.ConsumedAt)
	})
}
