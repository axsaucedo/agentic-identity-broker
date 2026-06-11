//go:build integration
// +build integration

package postgres

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAgentConstraintTestDB(t *testing.T, upToMigration int) (*sql.DB, func()) {
	t.Helper()

	connString, cleanup := setupDatabaseFromTemplate(t, fmt.Sprintf("agent_constraint_%d", upToMigration), func(t *testing.T, dbName string) {
		applyMigrationsUpToDatabase(t, requireSharedTestContainer(t), dbName, upToMigration)
	})

	db, err := sql.Open("pgx", connString)
	require.NoError(t, err)
	require.NoError(t, db.Ping())

	return db, func() {
		require.NoError(t, db.Close())
		cleanup()
	}
}

func insertAgentRow(t *testing.T, db *sql.DB, clientID, displayName string) error {
	t.Helper()

	_, err := db.Exec(`
		INSERT INTO agents (id, client_id, display_name, description, created_at, updated_at)
		VALUES (uuid_generate_v4(), $1, $2, 'test description', NOW(), NOW())
	`, clientID, displayName)
	return err
}

// T042: Integration test for migration 008 — drop/restore agents.client_id unique constraint.
func TestMigration008_AgentClientIDUniqueConstraint(t *testing.T) {
	t.Run("after migration 008: two agents can share the same client_id", func(t *testing.T) {
		db, cleanup := setupAgentConstraintTestDB(t, 8)
		defer cleanup()

		const sharedClientID = "shared-upstream-client-008"
		require.NoError(t, insertAgentRow(t, db, sharedClientID, "Agent One"))
		require.NoError(t, insertAgentRow(t, db, sharedClientID, "Agent Two"))

		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM agents WHERE client_id = $1`, sharedClientID).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 2, count, "migration 008 should allow duplicate client_id rows")
	})

	t.Run("before migration 008 (migrations 001-007): duplicate client_id is rejected by DB constraint", func(t *testing.T) {
		db, cleanup := setupAgentConstraintTestDB(t, 7)
		defer cleanup()

		const sharedClientID = "unique-client-pre-008"
		require.NoError(t, insertAgentRow(t, db, sharedClientID, "Agent One"))

		err := insertAgentRow(t, db, sharedClientID, "Agent Two")
		assert.Error(t, err, "second agent with duplicate client_id must fail before migration 008")
	})
}
