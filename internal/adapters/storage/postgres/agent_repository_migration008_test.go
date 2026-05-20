//go:build integration
// +build integration

package postgres

import (
	"context"
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
// After applying up to upToMigration, it always applies remaining schema migrations
// (009+) that the application-layer repositories require for their SQL queries.
func setupAgentTestDBWithMigrations(t *testing.T, upToMigration int) (*Adapter, func()) {
	t.Helper()

	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)

	ctx := context.Background()
	projectRoot, err := findProjectRoot()
	require.NoError(t, err)
	migrationsDir := filepath.Join(projectRoot, "migrations")

	createSchemaMigrationsTable(t, ctx, container)

	// Core migrations: applied only up to upToMigration (to test DB constraint behavior).
	coreMigrations := []struct {
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

	for _, m := range coreMigrations {
		if int(m.version) > upToMigration {
			break
		}
		applyOneMigration(t, ctx, container, migrationsDir, m.file, m.version)
	}

	// Always apply additional migrations so application-layer repo queries work.
	// Includes all migrations through the current schema so repo INSERT/SELECT
	// statements referencing columns added in later migrations (e.g. permission_sets
	// on agents from migration 017) do not fail with "column does not exist".
	additionalMigrations := []struct {
		file    string
		version int64
	}{
		{"009_add_agent_redirect_uris.up.sql", 9},
		{"010_create_client_credentials.up.sql", 10},
		{"011_create_signing_keys.up.sql", 11},
		{"012_create_authorization_codes.up.sql", 12},
		{"013_add_client_id_to_auth_codes.up.sql", 13},
		{"014_create_pkce_sessions.up.sql", 14},
		{"015_add_cimd_support.up.sql", 15},
		{"017_add_permission_sets.up.sql", 17},
		{"018_add_agent_permission_sets.up.sql", 18},
		{"019_migrate_user_grants_to_permission_sets.up.sql", 19},
		{"020_add_service_scope_requirement_type.up.sql", 20},
	}
	for _, m := range additionalMigrations {
		applyOneMigration(t, ctx, container, migrationsDir, m.file, m.version)
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
			ClientID:    &sharedClientID,
			DisplayName: "Agent One",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err, "first agent with shared client_id should be created")

		agent2 := &storage.Agent{
			ClientID:    &sharedClientID,
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
			ClientID:    &sharedClientID,
			DisplayName: "Agent One",
			Description: "First agent",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err := repo.Create(ctx, agent1)
		require.NoError(t, err, "first agent should be created")

		agent2 := &storage.Agent{
			ClientID:    &sharedClientID,
			DisplayName: "Agent Two",
			Description: "Duplicate",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		err = repo.Create(ctx, agent2)
		assert.Error(t, err, "second agent with duplicate client_id must fail before migration 008 (DB UNIQUE constraint)")
	})
}
