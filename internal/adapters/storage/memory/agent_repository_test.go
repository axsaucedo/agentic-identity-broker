package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestAgentRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success with generated ID", func(t *testing.T) {
		repo := NewAgentRepository()
		agent := &storage.Agent{
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)
		assert.False(t, agent.ID.IsZero()) // ID should be generated

		// Verify agent was stored
		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, agent.ClientID, retrieved.ClientID)
	})

	t.Run("success with provided ID", func(t *testing.T) {
		repo := NewAgentRepository()
		customID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a01")
		agent := &storage.Agent{
			ID:          customID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)
		assert.Equal(t, customID, agent.ID)
	})

	t.Run("duplicate ID conflict", func(t *testing.T) {
		repo := NewAgentRepository()
		dupID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a02")
		agent1 := &storage.Agent{
			ID:          dupID,
			ClientID:    id.ClientID("client-1"),
			DisplayName: "Agent 1",
			Description: "First agent",
		}
		agent2 := &storage.Agent{
			ID:          dupID,
			ClientID:    id.ClientID("client-2"),
			DisplayName: "Agent 2",
			Description: "Second agent",
		}

		err := repo.Create(ctx, agent1)
		require.NoError(t, err)

		err = repo.Create(ctx, agent2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("duplicate client_id allowed (Feature 021: multiple agents share one upstream client_id)", func(t *testing.T) {
		repo := NewAgentRepository()
		agent1 := &storage.Agent{
			ClientID:    id.ClientID("shared-upstream-client"),
			DisplayName: "Agent 1",
			Description: "First agent",
		}
		agent2 := &storage.Agent{
			ClientID:    id.ClientID("shared-upstream-client"),
			DisplayName: "Agent 2",
			Description: "Second agent",
		}

		err := repo.Create(ctx, agent1)
		require.NoError(t, err)

		// Feature 021: duplicate client_id must NOT return an error
		err = repo.Create(ctx, agent2)
		require.NoError(t, err, "multiple agents may share the same upstream client_id")
	})

	t.Run("validation failure", func(t *testing.T) {
		repo := NewAgentRepository()
		agent := &storage.Agent{
			ClientID: id.ClientID("test-client"),
			// Missing required DisplayName
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestAgentRepository_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, testAgentID)
		require.NoError(t, err)
		assert.Equal(t, testAgentID, retrieved.ID)
		assert.Equal(t, id.ClientID("test-client"), retrieved.ClientID)
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewAgentRepository()

		retrieved, err := repo.Get(ctx, id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"))
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("returns copy prevents external mutation", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Original Name",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Get and modify
		retrieved, err := repo.Get(ctx, testAgentID)
		require.NoError(t, err)
		retrieved.DisplayName = "Modified Name"

		// Original should be unchanged
		original, err := repo.Get(ctx, testAgentID)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", original.DisplayName)
	})
}

func TestAgentRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Original Name",
			Description: "Original description",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Update
		agent.DisplayName = "Updated Name"
		agent.Description = "Updated description"
		err = repo.Update(ctx, agent)
		require.NoError(t, err)

		// Verify update
		updated, err := repo.Get(ctx, testAgentID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.DisplayName)
		assert.Equal(t, "Updated description", updated.Description)
	})

	t.Run("not found", func(t *testing.T) {
		repo := NewAgentRepository()
		agent := &storage.Agent{
			ID:          id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"),
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Update(ctx, agent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("client_id conflict on update", func(t *testing.T) {
		repo := NewAgentRepository()

		agent1 := &storage.Agent{
			ID:          id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a21"),
			ClientID:    id.ClientID("client-1"),
			DisplayName: "Agent 1",
			Description: "First agent",
		}
		agent2 := &storage.Agent{
			ID:          id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22"),
			ClientID:    id.ClientID("client-2"),
			DisplayName: "Agent 2",
			Description: "Second agent",
		}

		err := repo.Create(ctx, agent1)
		require.NoError(t, err)
		err = repo.Create(ctx, agent2)
		require.NoError(t, err)

		// Try to update agent2 with agent1's client_id
		agent2.ClientID = id.ClientID("client-1")
		err = repo.Update(ctx, agent2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "client_id already exists")
	})

	t.Run("validation failure", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Invalid update (empty description)
		agent.Description = ""
		err = repo.Update(ctx, agent)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})
}

func TestAgentRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		err = repo.Delete(ctx, testAgentID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.Get(ctx, testAgentID)
		require.Error(t, err)
	})

	t.Run("idempotent - nonexistent agent", func(t *testing.T) {
		repo := NewAgentRepository()

		err := repo.Delete(ctx, id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a99"))
		require.NoError(t, err) // Should not error
	})

	t.Run("cleans up client_id index", func(t *testing.T) {
		repo := NewAgentRepository()
		testAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a10")
		agent := &storage.Agent{
			ID:          testAgentID,
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Test Agent",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		err = repo.Delete(ctx, testAgentID)
		require.NoError(t, err)

		// Should be able to reuse client_id
		newAgent := &storage.Agent{
			ClientID:    id.ClientID("test-client"),
			DisplayName: "New Agent",
			Description: "A new agent",
		}
		err = repo.Create(ctx, newAgent)
		require.NoError(t, err)
	})
}

func TestAgentRepository_List(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all agents", func(t *testing.T) {
		repo := NewAgentRepository()

		agent1 := &storage.Agent{
			ClientID:    id.ClientID("client-1"),
			DisplayName: "Agent 1",
			Description: "First agent",
		}
		agent2 := &storage.Agent{
			ClientID:    id.ClientID("client-2"),
			DisplayName: "Agent 2",
			Description: "Second agent",
		}

		err := repo.Create(ctx, agent1)
		require.NoError(t, err)
		err = repo.Create(ctx, agent2)
		require.NoError(t, err)

		agents, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 2)
	})

	t.Run("returns empty slice when no agents", func(t *testing.T) {
		repo := NewAgentRepository()

		agents, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, agents)
		assert.NotNil(t, agents) // Should be empty slice, not nil
	})

	t.Run("returns copies prevent external mutation", func(t *testing.T) {
		repo := NewAgentRepository()
		agent := &storage.Agent{
			ClientID:    id.ClientID("test-client"),
			DisplayName: "Original Name",
			Description: "A test agent",
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		agents, err := repo.List(ctx)
		require.NoError(t, err)
		require.Len(t, agents, 1)

		// Modify returned agent
		agents[0].DisplayName = "Modified Name"

		// Original should be unchanged
		original, err := repo.Get(ctx, agents[0].ID)
		require.NoError(t, err)
		assert.Equal(t, "Original Name", original.DisplayName)
	})
}
