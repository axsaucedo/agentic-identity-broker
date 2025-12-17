package memory

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserGrantRepository_Create(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	validUntil := time.Now().Add(24 * time.Hour)
	grant := &storage.UserGrant{
		Principal:  "user@example.com",
		AgentID:    "agent-123",
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo", "user:email"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)
	assert.NotEmpty(t, grant.ID, "ID should be generated")

	// Verify grant was stored
	retrieved, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Equal(t, grant.Principal, retrieved.Principal)
	assert.Equal(t, grant.AgentID, retrieved.AgentID)
}

func TestUserGrantRepository_UpsertSemantics(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	principal := "user@example.com"
	agentID := "agent-123"
	validUntil1 := time.Now().Add(24 * time.Hour)

	// Create first grant
	grant1 := &storage.UserGrant{
		Principal:  principal,
		AgentID:    agentID,
		ValidUntil: &validUntil1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant1)
	require.NoError(t, err)
	firstID := grant1.ID

	// Create second grant for same principal+agent (should update, not create new)
	validUntil2 := time.Now().Add(48 * time.Hour)
	grant2 := &storage.UserGrant{
		Principal:  principal,
		AgentID:    agentID,
		ValidUntil: &validUntil2,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-google",
				Scopes:                    []string{"openid", "email"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant2)
	require.NoError(t, err)

	// Should reuse the same ID (upsert)
	assert.Equal(t, firstID, grant2.ID, "upsert should reuse existing ID")

	// Verify only one grant exists for this principal+agent pair
	grants, err := repo.ListByPrincipalAndAgent(ctx, principal, agentID)
	require.NoError(t, err)
	assert.Len(t, grants, 1, "should have exactly one grant after upsert")

	// Verify the grant was updated with new tokens
	assert.Len(t, grants[0].DelegatedOAuth2Tokens, 1)
	assert.Equal(t, "service-google", grants[0].DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
}

func TestUserGrantRepository_Get(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Get non-existent grant
	_, err := repo.Get(ctx, "nonexistent-id")
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Create and get grant
	grant := &storage.UserGrant{
		Principal: "user@example.com",
		AgentID:   "agent-123",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant)
	require.NoError(t, err)

	retrieved, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Equal(t, grant.ID, retrieved.ID)
	assert.Equal(t, grant.Principal, retrieved.Principal)
}

func TestUserGrantRepository_Update(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant
	grant := &storage.UserGrant{
		Principal: "user@example.com",
		AgentID:   "agent-123",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)

	// Update grant
	validUntil := time.Now().Add(48 * time.Hour)
	grant.ValidUntil = &validUntil
	grant.DelegatedOAuth2Tokens = []storage.DelegatedToken{
		{
			ThirdpartyOAuth2ServiceID: "service-google",
			Scopes:                    []string{"openid"},
		},
	}

	err = repo.Update(ctx, grant)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.ValidUntil)
	assert.Len(t, retrieved.DelegatedOAuth2Tokens, 1)
	assert.Equal(t, "service-google", retrieved.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
}

func TestUserGrantRepository_Update_NotFound(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	grant := &storage.UserGrant{
		ID:        "nonexistent-id",
		Principal: "user@example.com",
		AgentID:   "agent-123",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Update(ctx, grant)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestUserGrantRepository_Delete(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant
	grant := &storage.UserGrant{
		Principal: "user@example.com",
		AgentID:   "agent-123",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)

	// Delete grant
	err = repo.Delete(ctx, grant.ID)
	require.NoError(t, err)

	// Verify deletion
	_, err = repo.Get(ctx, grant.ID)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestUserGrantRepository_Delete_Idempotent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Delete non-existent grant (should not error)
	err := repo.Delete(ctx, "nonexistent-id")
	require.NoError(t, err)
}

func TestUserGrantRepository_ListByPrincipalAndAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	principal := "user@example.com"
	agentID := "agent-123"

	// List when no grants exist
	grants, err := repo.ListByPrincipalAndAgent(ctx, principal, agentID)
	require.NoError(t, err)
	assert.Empty(t, grants)

	// Create grant
	grant := &storage.UserGrant{
		Principal: principal,
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant)
	require.NoError(t, err)

	// List grants
	grants, err = repo.ListByPrincipalAndAgent(ctx, principal, agentID)
	require.NoError(t, err)
	assert.Len(t, grants, 1)
	assert.Equal(t, grant.ID, grants[0].ID)
}

func TestUserGrantRepository_ListByPrincipalAndAgent_IncludesExpired(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	principal := "user@example.com"
	agentID := "agent-123"

	// Create grant with very short validity (1 millisecond in future)
	// This will be valid at creation but may expire by the time we retrieve it
	validUntil := time.Now().Add(1 * time.Millisecond)
	grant := &storage.UserGrant{
		Principal:  principal,
		AgentID:    agentID,
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)

	// Wait for grant to expire
	time.Sleep(10 * time.Millisecond)

	// List should include expired grants (filtering happens in service layer)
	grants, err := repo.ListByPrincipalAndAgent(ctx, principal, agentID)
	require.NoError(t, err)
	assert.Len(t, grants, 1, "expired grants should be included in repository results")
	// Note: Grant may or may not be active depending on timing, but it should be returned
}

func TestUserGrantRepository_FindByPrincipalAndAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	principal := "user@example.com"
	agentID := "agent-123"

	// Find when no grant exists
	_, err := repo.FindByPrincipalAndAgent(ctx, principal, agentID)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Create grant
	grant := &storage.UserGrant{
		Principal: principal,
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant)
	require.NoError(t, err)

	// Find grant
	found, err := repo.FindByPrincipalAndAgent(ctx, principal, agentID)
	require.NoError(t, err)
	assert.Equal(t, grant.ID, found.ID)
	assert.Equal(t, principal, found.Principal)
	assert.Equal(t, agentID, found.AgentID)
}

func TestUserGrantRepository_DeleteByAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	agentID := "agent-123"

	// Create multiple grants for same agent with different principals
	grant1 := &storage.UserGrant{
		Principal: "user1@example.com",
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	grant2 := &storage.UserGrant{
		Principal: "user2@example.com",
		AgentID:   agentID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-google",
				Scopes:                    []string{"openid"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create grant for different agent
	grant3 := &storage.UserGrant{
		Principal: "user1@example.com",
		AgentID:   "agent-456",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant1)
	require.NoError(t, err)
	err = repo.Create(ctx, grant2)
	require.NoError(t, err)
	err = repo.Create(ctx, grant3)
	require.NoError(t, err)

	// Delete all grants for agent-123
	err = repo.DeleteByAgent(ctx, agentID)
	require.NoError(t, err)

	// Verify grants for agent-123 are deleted
	_, err = repo.Get(ctx, grant1.ID)
	require.Error(t, err)
	_, err = repo.Get(ctx, grant2.ID)
	require.Error(t, err)

	// Verify grant for agent-456 still exists
	_, err = repo.Get(ctx, grant3.ID)
	require.NoError(t, err)
}

func TestUserGrantRepository_DeleteByAgent_Idempotent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Delete grants for non-existent agent (should not error)
	err := repo.DeleteByAgent(ctx, "nonexistent-agent")
	require.NoError(t, err)
}

func TestUserGrantRepository_ConcurrentAccess(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Test concurrent creates
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			grant := &storage.UserGrant{
				Principal: "user@example.com",
				AgentID:   "agent-123",
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "service-github",
						Scopes:                    []string{"repo"},
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = repo.Create(ctx, grant)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify only one grant exists (upsert semantics should handle concurrency)
	grants, err := repo.ListByPrincipalAndAgent(ctx, "user@example.com", "agent-123")
	require.NoError(t, err)
	assert.Len(t, grants, 1, "upsert should result in one grant despite concurrent creates")
}

func TestUserGrantRepository_DeepCopy(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant
	grant := &storage.UserGrant{
		Principal: "user@example.com",
		AgentID:   "agent-123",
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "service-github",
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)

	// Get grant
	retrieved, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)

	// Modify retrieved grant
	retrieved.DelegatedOAuth2Tokens[0].Scopes = append(retrieved.DelegatedOAuth2Tokens[0].Scopes, "user:email")

	// Get again and verify original wasn't mutated
	retrieved2, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Len(t, retrieved2.DelegatedOAuth2Tokens[0].Scopes, 1, "deep copy should prevent mutation")
}
