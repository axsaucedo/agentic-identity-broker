//go:build integration
// +build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// TestAgentRepository_Get_EmitsSpan verifies that AgentRepository.Get emits an OTel span
// named "storage.get.agent" with db.system=postgresql attribute (T030).
// TDD Red Phase: this test must fail before T033 implements the child span.
func TestAgentRepository_Get_EmitsSpan(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	// Set up an in-memory SpanRecorder as the global TracerProvider
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(recorder),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	prevTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		require.NoError(t, tp.Shutdown(context.Background()))
		otel.SetTracerProvider(prevTP)
	})

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	// Create an agent so Get can find it
	now := time.Now().UTC()
	agent := &storage.Agent{
		ClientID:    "span-test-client",
		DisplayName: "Span Test Agent",
		Description: "Agent for span testing",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := repo.Create(ctx, agent)
	require.NoError(t, err)

	// Call Get — should emit a span named "storage.get.agent"
	_, err = repo.Get(ctx, agent.ID)
	require.NoError(t, err)

	// Flush spans
	require.NoError(t, tp.ForceFlush(ctx))

	// Assert span was emitted
	ended := recorder.Ended()
	require.NotEmpty(t, ended, "expected at least one span to be recorded")

	var foundSpan bool
	for _, span := range ended {
		if span.Name() == "storage.get.agent" {
			foundSpan = true
			// Verify db.system=postgresql attribute
			foundAttr := false
			for _, attr := range span.Attributes() {
				if string(attr.Key) == "db.system" && attr.Value.AsString() == "postgresql" {
					foundAttr = true
					break
				}
			}
			assert.True(t, foundAttr, "expected span to have attribute db.system=postgresql")
			break
		}
	}
	assert.True(t, foundSpan, "expected a span named 'storage.get.agent' to be recorded")
}

// setupAgentTestDB creates a test database with migrations applied.
func setupAgentTestDB(t *testing.T) (*Adapter, func()) {
	t.Helper()

	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)

	// Apply migrations
	applyMigrations(t, container)

	// Create adapter
	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connString,
		},
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}

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

func TestAgentRepository_Create(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "test-client-1",
			DisplayName: "Test Agent",
			Description: "Test agent description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)
		assert.NotEmpty(t, agent.ID, "ID should be generated")

		// Verify retrieval
		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, agent.ClientID, retrieved.ClientID)
		assert.Equal(t, agent.DisplayName, retrieved.DisplayName)
		assert.Equal(t, agent.Description, retrieved.Description)
	})

	t.Run("with optional fields", func(t *testing.T) {
		now := time.Now().UTC()
		externalID := id.ExternalID("ext-123")
		govURL := "https://example.com/governance"
		docURL := "https://example.com/docs"
		agentURL := "https://example.com/agent"

		agent := &storage.Agent{
			ClientID:             "test-client-2",
			ExternalID:           &externalID,
			DisplayName:          "Test Agent 2",
			Description:          "Test agent with optional fields",
			GovernanceURL:        &govURL,
			UserDocumentationURL: &docURL,
			AgentInterfaceURL:    &agentURL,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, id.ExternalID(externalID), *retrieved.ExternalID)
		assert.Equal(t, govURL, *retrieved.GovernanceURL)
		assert.Equal(t, docURL, *retrieved.UserDocumentationURL)
		assert.Equal(t, agentURL, *retrieved.AgentInterfaceURL)
	})

	t.Run("duplicate client_id", func(t *testing.T) {
		now := time.Now().UTC()
		agent1 := &storage.Agent{
			ClientID:    "duplicate-client",
			DisplayName: "Agent 1",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent1)
		require.NoError(t, err)

		agent2 := &storage.Agent{
			ClientID:    "duplicate-client",
			DisplayName: "Agent 2",
			Description: "Second agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Feature 021: multiple agents may share the same upstream client_id
		err = repo.Create(ctx, agent2)
		require.NoError(t, err)
		assert.NotEqual(t, agent1.ID, agent2.ID, "both agents must have distinct IDs")
	})

	t.Run("validation failure - empty client_id", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "",
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})

	t.Run("validation failure - empty display_name", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "test-client-3",
			DisplayName: "",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})

	t.Run("validation failure - invalid URL", func(t *testing.T) {
		now := time.Now().UTC()
		invalidURL := "not-a-url"
		agent := &storage.Agent{
			ClientID:      "test-client-4",
			DisplayName:   "Test Agent",
			Description:   "Test description",
			GovernanceURL: &invalidURL,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		err := repo.Create(ctx, agent)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})
}

func TestAgentRepository_Get(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("existing agent", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "get-test-client",
			DisplayName: "Get Test Agent",
			Description: "Agent for get testing",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, agent.ID, retrieved.ID)
		assert.Equal(t, agent.ClientID, retrieved.ClientID)
		assert.Equal(t, agent.DisplayName, retrieved.DisplayName)
	})

	t.Run("non-existent agent", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, id.MustParseAgentID("00000000-0000-0000-0000-000000000001"))
		require.Error(t, err)
		assert.Nil(t, retrieved)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("empty ID", func(t *testing.T) {
		retrieved, err := repo.Get(ctx, id.AgentID{})
		require.Error(t, err)
		assert.Nil(t, retrieved)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})

	t.Run("returns copy prevents mutation", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "mutation-test-client",
			DisplayName: "Mutation Test",
			Description: "Test mutation protection",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		retrieved1, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)

		// Mutate retrieved copy
		retrieved1.DisplayName = "Modified Name"

		// Get again and verify original is unchanged
		retrieved2, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, "Mutation Test", retrieved2.DisplayName)
		assert.NotEqual(t, retrieved1.DisplayName, retrieved2.DisplayName)
	})
}

func TestAgentRepository_Update(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "update-test-client",
			DisplayName: "Original Name",
			Description: "Original description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Update agent
		agent.DisplayName = "Updated Name"
		agent.Description = "Updated description"
		agent.UpdatedAt = time.Now().UTC()

		err = repo.Update(ctx, agent)
		require.NoError(t, err)

		// Verify update
		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.DisplayName)
		assert.Equal(t, "Updated description", retrieved.Description)
	})

	t.Run("update with client_id change", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "original-client-id",
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Update client_id
		agent.ClientID = "new-client-id"
		agent.UpdatedAt = time.Now().UTC()

		err = repo.Update(ctx, agent)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, id.ClientID("new-client-id"), retrieved.ClientID)
	})

	t.Run("update non-existent agent", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ID:          id.MustParseAgentID("00000000-0000-0000-0000-000000000002"),
			ClientID:    "test-client",
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Update(ctx, agent)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("update with duplicate client_id", func(t *testing.T) {
		now := time.Now().UTC()
		agent1 := &storage.Agent{
			ClientID:    "client-1",
			DisplayName: "Agent 1",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err)

		agent2 := &storage.Agent{
			ClientID:    "client-2",
			DisplayName: "Agent 2",
			Description: "Second agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err = repo.Create(ctx, agent2)
		require.NoError(t, err)

		// Feature 021: duplicate client_id is allowed — update must succeed
		agent2.ClientID = "client-1"
		err = repo.Update(ctx, agent2)
		require.NoError(t, err)

		retrieved, err := repo.Get(ctx, agent2.ID)
		require.NoError(t, err)
		assert.Equal(t, id.ClientID("client-1"), retrieved.ClientID)
	})

	t.Run("validation failure on update", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "validation-test-client",
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Update with invalid data
		agent.DisplayName = ""
		err = repo.Update(ctx, agent)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})
}

func TestAgentRepository_Delete(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "delete-test-client",
			DisplayName: "Delete Test Agent",
			Description: "Agent for delete testing",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		// Delete agent
		err = repo.Delete(ctx, agent.ID)
		require.NoError(t, err)

		// Verify deletion
		_, err = repo.Get(ctx, agent.ID)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("idempotent deletion", func(t *testing.T) {
		// Delete non-existent agent - should not error
		err := repo.Delete(ctx, id.MustParseAgentID("00000000-0000-0000-0000-000000000003"))
		require.NoError(t, err)
	})

	t.Run("empty ID validation", func(t *testing.T) {
		err := repo.Delete(ctx, id.AgentID{})
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
	})
}

func TestAgentRepository_ClientURIs(t *testing.T) {
	adapter, cleanup := setupAgentTestDBWithCIMD(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("round-trip through Create and Get", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "cimd-create-get-client",
			DisplayName: "CIMD Create/Get Agent",
			Description: "Tests ClientURIs round-trip",
			ClientURIs:  []string{"https://example.com/client1", "https://example.com/client2"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent))

		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.ElementsMatch(t, agent.ClientURIs, retrieved.ClientURIs)
	})

	t.Run("round-trip through Create and List", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "cimd-list-client",
			DisplayName: "CIMD List Agent",
			Description: "Tests ClientURIs in List",
			ClientURIs:  []string{"https://example.com/list-client1"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent))

		agents, err := repo.List(ctx)
		require.NoError(t, err)

		var found *storage.Agent
		for _, a := range agents {
			if a.ID == agent.ID {
				found = a
				break
			}
		}
		require.NotNil(t, found)
		assert.Equal(t, agent.ClientURIs, found.ClientURIs)
	})

	t.Run("round-trip through Update and Get", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "cimd-update-client",
			DisplayName: "CIMD Update Agent",
			Description: "Tests ClientURIs update",
			ClientURIs:  []string{"https://example.com/update-original"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent))

		agent.ClientURIs = []string{"https://example.com/update-new"}
		agent.UpdatedAt = time.Now().UTC()
		require.NoError(t, repo.Update(ctx, agent))

		retrieved, err := repo.Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Equal(t, []string{"https://example.com/update-new"}, retrieved.ClientURIs)
	})

	t.Run("GetByClientID returns ClientURIs populated", func(t *testing.T) {
		now := time.Now().UTC()
		clientID := id.ClientID("cimd-getclientid-client")
		agent := &storage.Agent{
			ClientID:    clientID,
			DisplayName: "CIMD GetByClientID Agent",
			Description: "Tests GetByClientID hydrates ClientURIs",
			ClientURIs:  []string{"https://example.com/getclientid-uri"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent))

		retrieved, err := repo.GetByClientID(ctx, clientID)
		require.NoError(t, err)
		assert.Equal(t, agent.ClientURIs, retrieved.ClientURIs)
	})
}

func TestAgentRepository_GetByClientURI(t *testing.T) {
	adapter, cleanup := setupAgentTestDBWithCIMD(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("returns agent with ClientURIs populated", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "cimd-getbyuri-client",
			DisplayName: "CIMD GetByClientURI Agent",
			Description: "Tests GetByClientURI",
			ClientURIs:  []string{"https://example.com/lookup-uri", "https://example.com/other-uri"},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent))

		retrieved, err := repo.GetByClientURI(ctx, "https://example.com/lookup-uri")
		require.NoError(t, err)
		assert.Equal(t, agent.ID, retrieved.ID)
		assert.ElementsMatch(t, agent.ClientURIs, retrieved.ClientURIs)
	})

	t.Run("returns not-found for unregistered URI", func(t *testing.T) {
		_, err := repo.GetByClientURI(ctx, "https://example.com/not-registered")
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, storageErr.Kind)
	})

	t.Run("duplicate URI on Create rolls back entire operation", func(t *testing.T) {
		now := time.Now().UTC()
		sharedURI := "https://example.com/conflict-uri"

		agent1 := &storage.Agent{
			ClientID:    "conflict-agent-1",
			DisplayName: "Conflict Agent 1",
			Description: "Agent with the URI that will conflict",
			ClientURIs:  []string{sharedURI},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agent1))

		agent2 := &storage.Agent{
			ClientID:    "conflict-agent-2",
			DisplayName: "Conflict Agent 2",
			Description: "Agent that conflicts on URI",
			ClientURIs:  []string{sharedURI},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent2)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindConflict, storageErr.Kind)

		// Verify agent2 row was rolled back (not persisted)
		_, getErr := repo.Get(ctx, agent2.ID)
		require.Error(t, getErr)
		conflictStorageErr, ok := getErr.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindNotFound, conflictStorageErr.Kind)
	})

	t.Run("duplicate URI on Update rolls back entire operation", func(t *testing.T) {
		now := time.Now().UTC()
		uri1 := "https://example.com/update-conflict-uri1"
		uri2 := "https://example.com/update-conflict-uri2"

		agentA := &storage.Agent{
			ClientID:    "update-conflict-agent-a",
			DisplayName: "Update Conflict Agent A",
			Description: "Holds URI1 permanently",
			ClientURIs:  []string{uri1},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agentA))

		agentB := &storage.Agent{
			ClientID:    "update-conflict-agent-b",
			DisplayName: "Update Conflict Agent B",
			Description: "Tries to steal URI1 on update",
			ClientURIs:  []string{uri2},
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		require.NoError(t, repo.Create(ctx, agentB))

		// Try to update agentB to use uri1 (which agentA already holds)
		agentB.ClientURIs = []string{uri1}
		agentB.UpdatedAt = time.Now().UTC()
		err := repo.Update(ctx, agentB)
		require.Error(t, err)

		storageErr, ok := err.(*storage.StorageError)
		require.True(t, ok)
		assert.Equal(t, storage.ErrorKindConflict, storageErr.Kind)

		// Verify agentB still has its original URI (update was rolled back)
		retrieved, getErr := repo.Get(ctx, agentB.ID)
		require.NoError(t, getErr)
		assert.Equal(t, []string{uri2}, retrieved.ClientURIs)
	})
}

func TestAgentRepository_List(t *testing.T) {
	adapter, cleanup := setupAgentTestDB(t)
	defer cleanup()

	repo := NewAgentRepository(adapter)
	ctx := context.Background()

	t.Run("empty list", func(t *testing.T) {
		agents, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, agents)
	})

	t.Run("list multiple agents", func(t *testing.T) {
		now := time.Now().UTC()

		// Create multiple agents
		agent1 := &storage.Agent{
			ClientID:    "list-client-1",
			DisplayName: "Agent 1",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err)

		agent2 := &storage.Agent{
			ClientID:    "list-client-2",
			DisplayName: "Agent 2",
			Description: "Second agent",
			CreatedAt:   now.Add(1 * time.Second),
			UpdatedAt:   now.Add(1 * time.Second),
		}
		err = repo.Create(ctx, agent2)
		require.NoError(t, err)

		agent3 := &storage.Agent{
			ClientID:    "list-client-3",
			DisplayName: "Agent 3",
			Description: "Third agent",
			CreatedAt:   now.Add(2 * time.Second),
			UpdatedAt:   now.Add(2 * time.Second),
		}
		err = repo.Create(ctx, agent3)
		require.NoError(t, err)

		// List all agents
		agents, err := repo.List(ctx)
		require.NoError(t, err)
		assert.Len(t, agents, 3)

		// Verify order (should be DESC by created_at)
		assert.Equal(t, agent3.ID, agents[0].ID)
		assert.Equal(t, agent2.ID, agents[1].ID)
		assert.Equal(t, agent1.ID, agents[2].ID)
	})

	t.Run("returns copies prevent mutation", func(t *testing.T) {
		now := time.Now().UTC()
		agent := &storage.Agent{
			ClientID:    "list-mutation-client",
			DisplayName: "Mutation Test",
			Description: "Test mutation protection",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		err := repo.Create(ctx, agent)
		require.NoError(t, err)

		agents1, err := repo.List(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, agents1)

		// Mutate list item
		for _, a := range agents1 {
			if a.ID == agent.ID {
				a.DisplayName = "Modified Name"
			}
		}

		// List again and verify original is unchanged
		agents2, err := repo.List(ctx)
		require.NoError(t, err)
		for _, a := range agents2 {
			if a.ID == agent.ID {
				assert.Equal(t, "Mutation Test", a.DisplayName)
			}
		}
	})
}
