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

var (
	testAgentID1    = id.MustParseAgentID("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01")
	testAgentID2    = id.MustParseAgentID("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02")
	testPrincipal1  = id.Principal("user@example.com")
	testPrincipal2  = id.Principal("user1@example.com")
	testPrincipal3  = id.Principal("user2@example.com")
	testServiceGH   = id.MustParseServiceID("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01")
	testServiceGoog = id.MustParseServiceID("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02")
)

func TestUserGrantRepository_Create(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	validUntil := time.Now().Add(24 * time.Hour)
	grant := &storage.UserGrant{
		Principal:  testPrincipal1,
		AgentID:    testAgentID1,
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
				Scopes:                    []string{"repo", "user:email"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)
	assert.False(t, grant.ID.IsZero(), "ID should be generated")

	// Verify grant was stored
	retrieved, err := repo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Equal(t, grant.Principal, retrieved.Principal)
	assert.Equal(t, grant.AgentID, retrieved.AgentID)
}

func TestUserGrantRepository_UpsertSemantics(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	validUntil1 := time.Now().Add(24 * time.Hour)

	// Create first grant
	grant1 := &storage.UserGrant{
		Principal:  testPrincipal1,
		AgentID:    testAgentID1,
		ValidUntil: &validUntil1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
		Principal:  testPrincipal1,
		AgentID:    testAgentID1,
		ValidUntil: &validUntil2,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGoog,
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
	grants, err := repo.ListByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Len(t, grants, 1, "should have exactly one grant after upsert")

	// Verify the grant was updated with new tokens
	assert.Len(t, grants[0].DelegatedOAuth2Tokens, 1)
	assert.Equal(t, testServiceGoog, grants[0].DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
}

func TestUserGrantRepository_Get(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Get non-existent grant
	_, err := repo.Get(ctx, id.MustParseGrantID("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"))
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Create and get grant
	grant := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
			ThirdpartyOAuth2ServiceID: testServiceGoog,
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
	assert.Equal(t, testServiceGoog, retrieved.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
}

func TestUserGrantRepository_Update_NotFound(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	grant := &storage.UserGrant{
		ID:        id.MustParseGrantID("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"),
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
	err := repo.Delete(ctx, id.MustParseGrantID("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"))
	require.NoError(t, err)
}

func TestUserGrantRepository_ListByPrincipalAndAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// List when no grants exist
	grants, err := repo.ListByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Empty(t, grants)

	// Create grant
	grant := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant)
	require.NoError(t, err)

	// List grants
	grants, err = repo.ListByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Len(t, grants, 1)
	assert.Equal(t, grant.ID, grants[0].ID)
}

func TestUserGrantRepository_ListByPrincipalAndAgent_IncludesExpired(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant with very short validity (1 millisecond in future)
	validUntil := time.Now().Add(1 * time.Millisecond)
	grant := &storage.UserGrant{
		Principal:  testPrincipal1,
		AgentID:    testAgentID1,
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
	grants, err := repo.ListByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Len(t, grants, 1, "expired grants should be included in repository results")
}

func TestUserGrantRepository_FindByPrincipalAndAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Find when no grant exists
	_, err := repo.FindByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Create grant
	grant := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = repo.Create(ctx, grant)
	require.NoError(t, err)

	// Find grant
	found, err := repo.FindByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Equal(t, grant.ID, found.ID)
	assert.Equal(t, testPrincipal1, found.Principal)
	assert.Equal(t, testAgentID1, found.AgentID)
}

func TestUserGrantRepository_DeleteByAgent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create multiple grants for same agent with different principals
	grant1 := &storage.UserGrant{
		Principal: testPrincipal2,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	grant2 := &storage.UserGrant{
		Principal: testPrincipal3,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGoog,
				Scopes:                    []string{"openid"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create grant for different agent
	grant3 := &storage.UserGrant{
		Principal: testPrincipal2,
		AgentID:   testAgentID2,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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

	// Delete all grants for testAgentID1
	err = repo.DeleteByAgent(ctx, testAgentID1)
	require.NoError(t, err)

	// Verify grants for testAgentID1 are deleted
	_, err = repo.Get(ctx, grant1.ID)
	require.Error(t, err)
	_, err = repo.Get(ctx, grant2.ID)
	require.Error(t, err)

	// Verify grant for testAgentID2 still exists
	_, err = repo.Get(ctx, grant3.ID)
	require.NoError(t, err)
}

func TestUserGrantRepository_DeleteByAgent_Idempotent(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Delete grants for non-existent agent (should not error)
	err := repo.DeleteByAgent(ctx, id.MustParseAgentID("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"))
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
				Principal: testPrincipal1,
				AgentID:   testAgentID1,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: testServiceGH,
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
	grants, err := repo.ListByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)
	assert.Len(t, grants, 1, "upsert should result in one grant despite concurrent creates")
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create a grant for testPrincipal1 + testAgentID1
	grant := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant)
	require.NoError(t, err)

	// Delete it
	err = repo.DeleteByPrincipalAndAgentID(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)

	// Verify the grant is gone by ID
	_, err = repo.Get(ctx, grant.ID)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// Verify the principal+agent index is cleaned up
	_, err = repo.FindByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.Error(t, err)
	storageErr, ok = err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID_NotFound(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Delete when no grant exists — must return NotFound (NOT idempotent, unlike DeleteByAgent)
	err := repo.DeleteByPrincipalAndAgentID(ctx, testPrincipal1, testAgentID1)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID_CrossPrincipalIsolation(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grants for two different principals, same agent
	grant1 := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: testServiceGH, Scopes: []string{"repo"}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	grant2 := &storage.UserGrant{
		Principal: testPrincipal2,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: testServiceGH, Scopes: []string{"repo"}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grant1)
	require.NoError(t, err)
	err = repo.Create(ctx, grant2)
	require.NoError(t, err)

	// Delete only principal1's grant
	err = repo.DeleteByPrincipalAndAgentID(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)

	// principal1's grant is gone
	_, err = repo.FindByPrincipalAndAgent(ctx, testPrincipal1, testAgentID1)
	require.Error(t, err)
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok)
	assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

	// principal2's grant remains untouched (SR-001: cross-principal isolation)
	found, err := repo.FindByPrincipalAndAgent(ctx, testPrincipal2, testAgentID1)
	require.NoError(t, err)
	assert.Equal(t, grant2.ID, found.ID)
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID_AgentIndexCleanup(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant for agent1 and agent2 under same principal
	grantA1 := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: testServiceGH, Scopes: []string{"repo"}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	grantA2 := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID2,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{ThirdpartyOAuth2ServiceID: testServiceGoog, Scopes: []string{"openid"}},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(ctx, grantA1)
	require.NoError(t, err)
	err = repo.Create(ctx, grantA2)
	require.NoError(t, err)

	// Delete agent1's grant
	err = repo.DeleteByPrincipalAndAgentID(ctx, testPrincipal1, testAgentID1)
	require.NoError(t, err)

	// Cascade delete by agent1 should now be a no-op (index cleaned up)
	err = repo.DeleteByAgent(ctx, testAgentID1)
	require.NoError(t, err)

	// agent2's grant is still intact
	found, err := repo.FindByPrincipalAndAgent(ctx, testPrincipal1, testAgentID2)
	require.NoError(t, err)
	assert.Equal(t, grantA2.ID, found.ID)
}

func TestUserGrantRepository_DeepCopy(t *testing.T) {
	repo := NewUserGrantRepository()
	ctx := context.Background()

	// Create grant
	grant := &storage.UserGrant{
		Principal: testPrincipal1,
		AgentID:   testAgentID1,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGH,
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
