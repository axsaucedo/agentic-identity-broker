//go:build integration
// +build integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAgentTestDBWithMigrations creates a test DB with migrations 001-N applied.
func setupAgentTestDBWithMigrations(t *testing.T, upToMigration int) (*Adapter, func()) {
	t.Helper()

	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)

	ctx := context.Background()
	projectRoot, err := findProjectRoot()
	require.NoError(t, err)
	migrationsDir := filepath.Join(projectRoot, "migrations")

	allMigrations := []struct {
		file    string
		version int64
	}{
		{"001_create_agents.up.sql", 1},
		{"002_create_thirdparty_services.up.sql", 2},
		{"003_create_user_grants.up.sql", 3},
		{"004_create_user_sessions.up.sql", 4},
		{"005_add_agent_service_requirements.up.sql", 5},
		{"006_add_service_protected_resources.up.sql", 6},
		{"007_add_oauth2_flavor.up.sql", 7},
		{"008_drop_agent_client_id_unique.up.sql", 8},
	}

	// Create schema_migrations table
	container.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-c",
		`CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL DEFAULT FALSE);`})

	for _, m := range allMigrations {
		if int(m.version) > upToMigration {
			break
		}
		data, readErr := os.ReadFile(filepath.Join(migrationsDir, m.file))
		if readErr != nil {
			t.Logf("Skipping migration %s: %v", m.file, readErr)
			continue
		}
		container.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-c", string(data)})
	}

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
	adapter, adapterErr := NewAdapter(config)
	require.NoError(t, adapterErr)
	require.NoError(t, adapter.Initialize(ctx))

	return adapter, func() {
		adapter.Close(ctx)
		cleanup()
	}
}

// T042: Integration test for migration 008 — drop/restore agents.client_id unique constraint.
func TestMigration008_AgentClientIDUniqueConstraint(t *testing.T) {
	t.Run("after migration 008: two agents can share the same client_id", func(t *testing.T) {
		adapter, cleanup := setupAgentTestDBWithMigrations(t, 8)
		defer cleanup()

		repo := NewAgentRepository(adapter)
		ctx := context.Background()
		now := time.Now().UTC()

		sharedClientID := id.ClientID("shared-upstream-client-008")

		agent1 := &storage.Agent{
			ClientID:    sharedClientID,
			DisplayName: "Agent One",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err, "first agent with shared client_id should be created")

		agent2 := &storage.Agent{
			ClientID:    sharedClientID,
			DisplayName: "Agent Two",
			Description: "Second agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err = repo.Create(ctx, agent2)
		assert.NoError(t, err, "second agent with same client_id must succeed after migration 008")
	})

	t.Run("before migration 008 (migrations 001-007): duplicate client_id is rejected by DB constraint", func(t *testing.T) {
		adapter, cleanup := setupAgentTestDBWithMigrations(t, 7)
		defer cleanup()

		repo := NewAgentRepository(adapter)
		ctx := context.Background()
		now := time.Now().UTC()

		sharedClientID := id.ClientID("unique-client-pre-008")

		agent1 := &storage.Agent{
			ClientID:    sharedClientID,
			DisplayName: "Agent One",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err, "first agent should be created")

		agent2 := &storage.Agent{
			ClientID:    sharedClientID,
			DisplayName: "Agent Two",
			Description: "Duplicate",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err = repo.Create(ctx, agent2)
		assert.Error(t, err, "second agent with duplicate client_id must fail before migration 008 (DB UNIQUE constraint)")
	})
}
