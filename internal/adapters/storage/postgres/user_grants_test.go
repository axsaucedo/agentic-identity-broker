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

var (
	testServiceGitHub = id.MustParseServiceID("10000000-0000-0000-0000-000000000001")
	testServiceGoogle = id.MustParseServiceID("10000000-0000-0000-0000-000000000002")
	testGrantNotFound = id.MustParseGrantID("99000000-0000-0000-0000-000000000001")
	testAgentNotFound = id.MustParseAgentID("99000000-0000-0000-0000-000000000001")
)

// setupUserGrantTestDB creates a test database with migrations applied.
func setupUserGrantTestDB(t *testing.T) (*Adapter, *AgentRepository, *UserGrantRepository, func()) {
	t.Helper()

	adapter, cleanup := setupAgentTestDB(t)

	agentRepo := NewAgentRepository(adapter)
	grantRepo := NewUserGrantRepository(adapter)

	return adapter, agentRepo, grantRepo, cleanup
}

func TestUserGrantRepository_Create(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent first
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		validUntil := time.Now().Add(24 * time.Hour)
		grant := &storage.UserGrant{
			Principal:  id.Principal("user@example.com"),
			AgentID:    agent.ID,
			ValidUntil: &validUntil,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo", "user:email"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)
		assert.NotEmpty(t, grant.ID)

		// Verify retrieval
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.Equal(t, grant.Principal, retrieved.Principal)
		assert.Equal(t, grant.AgentID, retrieved.AgentID)
		assert.Len(t, retrieved.DelegatedOAuth2Tokens, 1)
	})

	t.Run("upsert semantics", func(t *testing.T) {
		principal := id.Principal("user2@example.com")
		validUntil1 := time.Now().Add(24 * time.Hour)

		grant1 := &storage.UserGrant{
			Principal:  principal,
			AgentID:    agent.ID,
			ValidUntil: &validUntil1,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant1)
		require.NoError(t, err)
		firstID := grant1.ID

		// Create second grant for same principal+agent (should update)
		validUntil2 := time.Now().Add(48 * time.Hour)
		grant2 := &storage.UserGrant{
			Principal:  principal,
			AgentID:    agent.ID,
			ValidUntil: &validUntil2,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGoogle,
					Scopes:                    []string{"openid"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err = grantRepo.Create(ctx, grant2)
		require.NoError(t, err)

		// Verify only one grant exists
		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Len(t, grants, 1)

		// Verify the grant was updated
		assert.Equal(t, testServiceGoogle, grants[0].DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)

		// The returned ID should be the first one (ON CONFLICT UPDATE keeps original ID)
		assert.Equal(t, firstID, grants[0].ID)
	})
}

func TestUserGrantRepository_Get(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("get non-existent grant", func(t *testing.T) {
		_, err := grantRepo.Get(ctx, testGrantNotFound)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("get existing grant", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: id.Principal("user@example.com"),
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.Equal(t, grant.ID, retrieved.ID)
		assert.Equal(t, grant.Principal, retrieved.Principal)
	})
}

func TestUserGrantRepository_Update(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("update existing grant", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: id.Principal("user@example.com"),
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Update grant
		validUntil := time.Now().Add(48 * time.Hour)
		grant.ValidUntil = &validUntil
		grant.DelegatedOAuth2Tokens = []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGoogle,
				Scopes:                    []string{"openid"},
			},
		}
		grant.UpdatedAt = time.Now().UTC()

		err = grantRepo.Update(ctx, grant)
		require.NoError(t, err)

		// Verify update
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.ValidUntil)
		assert.Len(t, retrieved.DelegatedOAuth2Tokens, 1)
		assert.Equal(t, testServiceGoogle, retrieved.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
	})

	t.Run("update non-existent grant", func(t *testing.T) {
		grant := &storage.UserGrant{
			ID:        testGrantNotFound,
			Principal: id.Principal("user@example.com"),
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Update(ctx, grant)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})
}

func TestUserGrantRepository_Delete(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("delete existing grant", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: id.Principal("user@example.com"),
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Delete grant
		err = grantRepo.Delete(ctx, grant.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = grantRepo.Get(ctx, grant.ID)
		require.Error(t, err)
	})

	t.Run("delete non-existent grant (idempotent)", func(t *testing.T) {
		err := grantRepo.Delete(ctx, testGrantNotFound)
		require.NoError(t, err)
	})
}

func TestUserGrantRepository_ListByPrincipalAndAgent(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	principal := id.Principal("user@example.com")

	t.Run("list when no grants exist", func(t *testing.T) {
		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Empty(t, grants)
	})

	t.Run("list existing grants", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: principal,
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Len(t, grants, 1)
		assert.Equal(t, grant.ID, grants[0].ID)
	})

	t.Run("includes expired grants", func(t *testing.T) {
		principal2 := id.Principal("user2@example.com")

		// Create grant with short validity
		validUntil := time.Now().Add(1 * time.Millisecond)
		grant := &storage.UserGrant{
			Principal:  principal2,
			AgentID:    agent.ID,
			ValidUntil: &validUntil,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Should still be included in list
		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal2, agent.ID)
		require.NoError(t, err)
		assert.Len(t, grants, 1)
	})
}

func TestUserGrantRepository_FindByPrincipalAndAgent(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	principal := id.Principal("user@example.com")

	t.Run("find when no grant exists", func(t *testing.T) {
		_, err := grantRepo.FindByPrincipalAndAgent(ctx, principal, agent.ID)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("find existing grant", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: principal,
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		found, err := grantRepo.FindByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, grant.ID, found.ID)
	})
}

func TestUserGrantRepository_DeleteByAgent(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agents
	agent1 := &storage.Agent{
		ClientID:    "test-agent-1",
		DisplayName: "Test Agent 1",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent1)
	require.NoError(t, err)

	agent2 := &storage.Agent{
		ClientID:    "test-agent-2",
		DisplayName: "Test Agent 2",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err = agentRepo.Create(ctx, agent2)
	require.NoError(t, err)

	t.Run("cascade delete all grants for agent", func(t *testing.T) {
		// Create multiple grants for agent1
		grant1 := &storage.UserGrant{
			Principal: id.Principal("user1@example.com"),
			AgentID:   agent1.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err := grantRepo.Create(ctx, grant1)
		require.NoError(t, err)

		grant2 := &storage.UserGrant{
			Principal: id.Principal("user2@example.com"),
			AgentID:   agent1.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGoogle,
					Scopes:                    []string{"openid"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = grantRepo.Create(ctx, grant2)
		require.NoError(t, err)

		// Create grant for agent2
		grant3 := &storage.UserGrant{
			Principal: id.Principal("user1@example.com"),
			AgentID:   agent2.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = grantRepo.Create(ctx, grant3)
		require.NoError(t, err)

		// Delete all grants for agent1
		err = grantRepo.DeleteByAgent(ctx, agent1.ID)
		require.NoError(t, err)

		// Verify agent1 grants are deleted
		_, err = grantRepo.Get(ctx, grant1.ID)
		require.Error(t, err)
		_, err = grantRepo.Get(ctx, grant2.ID)
		require.Error(t, err)

		// Verify agent2 grant still exists
		_, err = grantRepo.Get(ctx, grant3.ID)
		require.NoError(t, err)
	})

	t.Run("delete by non-existent agent (idempotent)", func(t *testing.T) {
		err := grantRepo.DeleteByAgent(ctx, testAgentNotFound)
		require.NoError(t, err)
	})
}

func TestUserGrantRepository_JSONBMarshaling(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("marshal and unmarshal complex tokens", func(t *testing.T) {
		grant := &storage.UserGrant{
			Principal: id.Principal("user@example.com"),
			AgentID:   agent.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{
					ThirdpartyOAuth2ServiceID: testServiceGitHub,
					Scopes:                    []string{"repo", "user:email", "read:org"},
				},
				{
					ThirdpartyOAuth2ServiceID: testServiceGoogle,
					Scopes:                    []string{"openid", "email", "profile"},
				},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.DelegatedOAuth2Tokens, 2)
		assert.Len(t, retrieved.DelegatedOAuth2Tokens[0].Scopes, 3)
		assert.Len(t, retrieved.DelegatedOAuth2Tokens[1].Scopes, 3)
		assert.Equal(t, "repo", retrieved.DelegatedOAuth2Tokens[0].Scopes[0])
	})
}

func TestUserGrantRepository_DeepCopy(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    "test-agent",
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	grant := &storage.UserGrant{
		Principal: id.Principal("user@example.com"),
		AgentID:   agent.ID,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceGitHub,
				Scopes:                    []string{"repo"},
			},
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err = grantRepo.Create(ctx, grant)
	require.NoError(t, err)

	// Get grant
	retrieved, err := grantRepo.Get(ctx, grant.ID)
	require.NoError(t, err)

	// Modify retrieved grant
	retrieved.DelegatedOAuth2Tokens[0].Scopes = append(retrieved.DelegatedOAuth2Tokens[0].Scopes, "user:email")

	// Get again and verify original wasn't mutated
	retrieved2, err := grantRepo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Len(t, retrieved2.DelegatedOAuth2Tokens[0].Scopes, 1)
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID(t *testing.T) {
	_, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two agents
	agent1 := &storage.Agent{
		ClientID:    "agent-revoke-1",
		DisplayName: "Revoke Test Agent 1",
		Description: "Used for DeleteByPrincipalAndAgentID tests",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	agent2 := &storage.Agent{
		ClientID:    "agent-revoke-2",
		DisplayName: "Revoke Test Agent 2",
		Description: "Used for DeleteByPrincipalAndAgentID tests",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent1)
	require.NoError(t, err)
	err = agentRepo.Create(ctx, agent2)
	require.NoError(t, err)

	t.Run("success: deletes grant and cleans up indexes", func(t *testing.T) {
		principal := id.Principal("revoke-user@example.com")
		grant := &storage.UserGrant{
			Principal: principal,
			AgentID:   agent1.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: testServiceGitHub, Scopes: []string{"repo"}},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		err = grantRepo.DeleteByPrincipalAndAgentID(ctx, principal, agent1.ID)
		require.NoError(t, err)

		// Grant must be gone by ID
		_, err = grantRepo.Get(ctx, grant.ID)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

		// principal+agent index must be cleaned up
		_, err = grantRepo.FindByPrincipalAndAgent(ctx, principal, agent1.ID)
		require.Error(t, err)
		storageErr, ok = err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("not found: no active grant exists — returns NotFound error", func(t *testing.T) {
		err := grantRepo.DeleteByPrincipalAndAgentID(ctx, id.Principal("no-grant@example.com"), testAgentNotFound)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("cross-principal isolation: only deletes the specified principal's grant (SR-001)", func(t *testing.T) {
		principalA := id.Principal("isolation-user-a@example.com")
		principalB := id.Principal("isolation-user-b@example.com")

		grantA := &storage.UserGrant{
			Principal: principalA,
			AgentID:   agent2.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: testServiceGitHub, Scopes: []string{"repo"}},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		grantB := &storage.UserGrant{
			Principal: principalB,
			AgentID:   agent2.ID,
			DelegatedOAuth2Tokens: []storage.DelegatedToken{
				{ThirdpartyOAuth2ServiceID: testServiceGoogle, Scopes: []string{"openid"}},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grantA)
		require.NoError(t, err)
		err = grantRepo.Create(ctx, grantB)
		require.NoError(t, err)

		// Revoke only principalA's grant
		err = grantRepo.DeleteByPrincipalAndAgentID(ctx, principalA, agent2.ID)
		require.NoError(t, err)

		// principalA's grant is gone
		_, err = grantRepo.FindByPrincipalAndAgent(ctx, principalA, agent2.ID)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)

		// principalB's grant is untouched
		found, err := grantRepo.FindByPrincipalAndAgent(ctx, principalB, agent2.ID)
		require.NoError(t, err)
		assert.Equal(t, grantB.ID, found.ID)
	})
}
