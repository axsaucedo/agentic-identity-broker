package memory

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationCodeStore_FindByCodeHash_ExpiredUnused(t *testing.T) {
	store := NewAuthorizationCodeStore()
	ctx := context.Background()

	code := &storage.AuthorizationCode{
		ID:            id.NewAuthorizationCodeID(),
		CodeHash:      "expiredhash",
		AgentID:       id.NewAgentID(),
		Principal:     id.NewPrincipal("user@example.com"),
		RedirectURI:   "http://localhost/callback",
		CodeChallenge: "challenge",
		Scope:         "read",
		ExpiresAt:     time.Now().Add(-10 * time.Minute), // expired
		CreatedAt:     time.Now().Add(-15 * time.Minute),
	}
	require.NoError(t, store.Create(ctx, code))

	got, err := store.FindByCodeHash(ctx, "expiredhash")
	require.NoError(t, err)
	assert.Equal(t, code.ID, got.ID)
	assert.Nil(t, got.UsedAt, "repo must not filter by expiry — domain layer owns that check")
}
