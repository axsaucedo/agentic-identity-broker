//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
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
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent first
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		validUntil := time.Now().Add(24 * time.Hour)
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			ValidUntil:            &validUntil,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)
		assert.NotEmpty(t, grant.ID)

		// Verify retrieval
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.Equal(t, grant.Principal, retrieved.Principal)
		assert.Equal(t, grant.AgentID, retrieved.AgentID)
		assert.Len(t, retrieved.GrantedPermissionSets, 1)
	})

	t.Run("upsert semantics", func(t *testing.T) {
		principal := id.Principal("user2@example.com")
		validUntil1 := time.Now().Add(24 * time.Hour)

		psID1 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID1)
		grant1 := &storage.UserGrant{
			Principal:             principal,
			AgentID:               agent.ID,
			ValidUntil:            &validUntil1,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID1, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant1)
		require.NoError(t, err)
		firstID := grant1.ID

		// Create second grant for same principal+agent (should update)
		validUntil2 := time.Now().Add(48 * time.Hour)
		psID2 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID2)
		grant2 := &storage.UserGrant{
			Principal:             principal,
			AgentID:               agent.ID,
			ValidUntil:            &validUntil2,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID2, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err = grantRepo.Create(ctx, grant2)
		require.NoError(t, err)

		// Verify only one grant exists
		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Len(t, grants, 1)

		// Verify the grant was updated with the new permission set ID
		assert.Len(t, grants[0].GrantedPermissionSets, 1)
		assert.Equal(t, psID2, grants[0].GrantedPermissionSets[0].PermissionSetID)

		// The returned ID should be the first one (ON CONFLICT UPDATE keeps original ID)
		assert.Equal(t, firstID, grants[0].ID)
	})
}

func TestUserGrantRepository_Get(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
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
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("update existing grant", func(t *testing.T) {
		psID1 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID1)
		grant := &storage.UserGrant{
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID1, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Update grant
		validUntil := time.Now().Add(48 * time.Hour)
		grant.ValidUntil = &validUntil
		psID2 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID2)
		grant.GrantedPermissionSets = []storage.GrantedPermissionSetEntry{{PermissionSetID: psID2, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}}
		grant.UpdatedAt = time.Now().UTC()

		err = grantRepo.Update(ctx, grant)
		require.NoError(t, err)

		// Verify update
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.ValidUntil)
		assert.Len(t, retrieved.GrantedPermissionSets, 1)
		assert.Equal(t, psID2, retrieved.GrantedPermissionSets[0].PermissionSetID)
	})

	// This subtest also verifies error-ordering: the grant references a nonexistent PS,
	// so if verifyPermissionSetExistenceInTx ran before the UPDATE it would return a PS
	// conflict. ErrorKindNotFound must win, confirming the UPDATE runs first.
	t.Run("update non-existent grant returns NotFound before PS conflict", func(t *testing.T) {
		grant := &storage.UserGrant{
			ID:                    testGrantNotFound,
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: id.NewPermissionSetID(), IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Update(ctx, grant)
		require.Error(t, err)
		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})
}

func TestUserGrantRepository_Delete(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("delete existing grant", func(t *testing.T) {
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
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
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             principal,
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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

		// Insert an already-expired grant directly, bypassing domain validation
		// which rejects past valid_until values.
		pastTime := time.Now().Add(-time.Hour).UTC()
		expiredGrantID := id.NewGrantID()
		psID := id.NewPermissionSetID()
		grantedPS := fmt.Sprintf(`[{"permission_set_id":"%s","included_service_ids":[]}]`, psID.String())

		_, err := adapter.db.ExecContext(ctx,
			`INSERT INTO user_grants (id, principal, agent_id, valid_until, granted_permission_sets, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`,
			expiredGrantID.String(), string(principal2), agent.ID.String(),
			pastTime, grantedPS, pastTime, pastTime,
		)
		require.NoError(t, err)

		// Expired grant must still appear in ListByPrincipalAndAgent
		grants, err := grantRepo.ListByPrincipalAndAgent(ctx, principal2, agent.ID)
		require.NoError(t, err)
		assert.Len(t, grants, 1)
	})
}

func TestUserGrantRepository_FindByPrincipalAndAgent(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
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
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             principal,
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		found, err := grantRepo.FindByPrincipalAndAgent(ctx, principal, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, grant.ID, found.ID)
	})
}

func TestUserGrantRepository_DeleteByAgent(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agents
	agent1 := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent-1")),
		DisplayName: "Test Agent 1",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent1)
	require.NoError(t, err)

	agent2 := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent-2")),
		DisplayName: "Test Agent 2",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err = agentRepo.Create(ctx, agent2)
	require.NoError(t, err)

	t.Run("cascade delete all grants for agent", func(t *testing.T) {
		// Create multiple grants for agent1
		psID1 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID1)
		grant1 := &storage.UserGrant{
			Principal:             id.Principal("user1@example.com"),
			AgentID:               agent1.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID1, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}
		err := grantRepo.Create(ctx, grant1)
		require.NoError(t, err)

		psID2 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID2)
		grant2 := &storage.UserGrant{
			Principal:             id.Principal("user2@example.com"),
			AgentID:               agent1.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID2, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}
		err = grantRepo.Create(ctx, grant2)
		require.NoError(t, err)

		// Create grant for agent2
		psID3 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID3)
		grant3 := &storage.UserGrant{
			Principal:             id.Principal("user1@example.com"),
			AgentID:               agent2.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID3, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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

func TestUserGrantRepository_PermissionSetIDMarshaling(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	t.Run("marshal and unmarshal multiple permission set IDs", func(t *testing.T) {
		psID1 := id.NewPermissionSetID()
		psID2 := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID1)
		seedPermissionSet(t, adapter, psID2)
		grant := &storage.UserGrant{
			Principal:             id.Principal("user@example.com"),
			AgentID:               agent.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID1, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}, {PermissionSetID: psID2, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}

		err := grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := grantRepo.Get(ctx, grant.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.GrantedPermissionSets, 2)
	})
}

func TestUserGrantRepository_DeepCopy(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create test agent
	agent := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("test-agent")),
		DisplayName: "Test Agent",
		Description: "Test description",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	err := agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	psID := id.NewPermissionSetID()
	seedPermissionSet(t, adapter, psID)
	grant := &storage.UserGrant{
		Principal:             id.Principal("user@example.com"),
		AgentID:               agent.ID,
		GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}

	err = grantRepo.Create(ctx, grant)
	require.NoError(t, err)

	// Get grant
	retrieved, err := grantRepo.Get(ctx, grant.ID)
	require.NoError(t, err)

	// Modify retrieved grant (in-memory only — no Create/Update, so no seed needed)
	retrieved.GrantedPermissionSets = append(retrieved.GrantedPermissionSets, storage.GrantedPermissionSetEntry{PermissionSetID: id.NewPermissionSetID(), IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}})

	// Get again and verify original wasn't mutated
	retrieved2, err := grantRepo.Get(ctx, grant.ID)
	require.NoError(t, err)
	assert.Len(t, retrieved2.GrantedPermissionSets, 1)
}

func TestUserGrantRepository_DeleteByPrincipalAndAgentID(t *testing.T) {
	adapter, agentRepo, grantRepo, cleanup := setupUserGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	// Create two agents
	agent1 := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("agent-revoke-1")),
		DisplayName: "Revoke Test Agent 1",
		Description: "Used for DeleteByPrincipalAndAgentID tests",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	agent2 := &storage.Agent{
		ClientID:    ptr.To(id.ClientID("agent-revoke-2")),
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
		psID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)
		grant := &storage.UserGrant{
			Principal:             principal,
			AgentID:               agent1.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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

	t.Run("CountGrantsReferencingPermissionSet returns correct count", func(t *testing.T) {
		psID := id.NewPermissionSetID()
		otherPSID := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psID)

		// No grants yet — count should be 0
		count, err := grantRepo.CountGrantsReferencingPermissionSet(ctx, psID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Create a grant referencing psID
		grant := &storage.UserGrant{
			Principal: id.Principal("ps-count-user@example.com"),
			AgentID:   agent2.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{
				{PermissionSetID: psID, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}},
			},
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err = grantRepo.Create(ctx, grant)
		require.NoError(t, err)

		// Count for psID should now be 1
		count, err = grantRepo.CountGrantsReferencingPermissionSet(ctx, psID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// otherPSID has no grants
		count, err = grantRepo.CountGrantsReferencingPermissionSet(ctx, otherPSID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)

		// Delete the grant — count should drop back to 0
		err = grantRepo.Delete(ctx, grant.ID)
		require.NoError(t, err)

		count, err = grantRepo.CountGrantsReferencingPermissionSet(ctx, psID)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("expired grant is excluded from count", func(t *testing.T) {
		expiredPSID := id.NewPermissionSetID()
		pastTime := time.Now().Add(-time.Hour).UTC()
		expiredGrantID := id.NewGrantID()
		grantedPS := fmt.Sprintf(`[{"permission_set_id":"%s","included_service_ids":[]}]`, expiredPSID.String())

		// Insert an expired grant directly, bypassing domain validation which rejects past valid_until.
		_, execErr := adapter.db.ExecContext(ctx,
			`INSERT INTO user_grants (id, principal, agent_id, valid_until, granted_permission_sets, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`,
			expiredGrantID.String(), "expired-ps-user@example.com", agent2.ID.String(),
			pastTime, grantedPS, pastTime, pastTime,
		)
		require.NoError(t, execErr)

		count, err := grantRepo.CountGrantsReferencingPermissionSet(ctx, expiredPSID)
		require.NoError(t, err)
		assert.Equal(t, 0, count, "expired grant must not be counted")
	})

	t.Run("future valid_until grant is included in count", func(t *testing.T) {
		futurePSID := id.NewPermissionSetID()
		futureTime := time.Now().Add(time.Hour).UTC()
		futureGrantID := id.NewGrantID()
		grantedPS := fmt.Sprintf(`[{"permission_set_id":"%s","included_service_ids":[]}]`, futurePSID.String())
		now := time.Now().UTC()

		_, execErr := adapter.db.ExecContext(ctx,
			`INSERT INTO user_grants (id, principal, agent_id, valid_until, granted_permission_sets, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)`,
			futureGrantID.String(), "future-ps-user@example.com", agent2.ID.String(),
			futureTime, grantedPS, now, now,
		)
		require.NoError(t, execErr)

		count, err := grantRepo.CountGrantsReferencingPermissionSet(ctx, futurePSID)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "grant with future valid_until must be counted")
	})

	t.Run("cross-principal isolation: only deletes the specified principal's grant (SR-001)", func(t *testing.T) {
		principalA := id.Principal("isolation-user-a@example.com")
		principalB := id.Principal("isolation-user-b@example.com")

		psIDA := id.NewPermissionSetID()
		psIDB := id.NewPermissionSetID()
		seedPermissionSet(t, adapter, psIDA)
		seedPermissionSet(t, adapter, psIDB)
		grantA := &storage.UserGrant{
			Principal:             principalA,
			AgentID:               agent2.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psIDA, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
		}
		grantB := &storage.UserGrant{
			Principal:             principalB,
			AgentID:               agent2.ID,
			GrantedPermissionSets: []storage.GrantedPermissionSetEntry{{PermissionSetID: psIDB, IncludedServiceIDs: []id.ServiceID{id.NewServiceID()}}},
			CreatedAt:             time.Now().UTC(),
			UpdatedAt:             time.Now().UTC(),
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
