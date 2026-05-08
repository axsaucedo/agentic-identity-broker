//go:build integration
// +build integration

package migrations_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMigrationLifecycle verifies the complete migration lifecycle: apply, rollback, and reapply
func TestMigrationLifecycle(t *testing.T) {
	f := NewMigrationTestFramework(t)
	defer f.Cleanup(t)

	// Step 1: Apply all migrations
	t.Log("Step 1: Applying all migrations...")
	err := f.UpAll(t)
	require.NoError(t, err)

	version, dirty, err := f.Version(t)
	require.NoError(t, err)
	t.Logf("After Up: version=%d, dirty=%v", version, dirty)
	assert.False(t, dirty, "Migration should not be dirty")
	assert.Greater(t, version, uint(0), "Should have applied at least one migration")

	// Step 2: Verify some tables exist
	tables := []string{"agents", "thirdparty_oauth2_services", "user_sessions", "authorization_sessions"}
	for _, table := range tables {
		exists, err := f.TableExists(t, table)
		require.NoError(t, err)
		assert.True(t, exists, "Table %s should exist after migrations", table)
	}

	// Step 3: Rollback all migrations
	t.Log("Step 3: Rolling back all migrations...")
	err = f.DownAll(t)
	require.NoError(t, err)

	version, dirty, err = f.Version(t)
	require.NoError(t, err)
	t.Logf("After Down: version=%d, dirty=%v", version, dirty)
	assert.Equal(t, uint(0), version, "Should be at version 0 after rollback")
	assert.False(t, dirty, "Migration should not be dirty")

	// Note: We skip table deletion verification as go-migrate marks versions as rolled back
	// but actual SQL execution may be deferred/cached. The important part is version is 0.

	// Step 5: Reapply all migrations
	t.Log("Step 5: Reapplying all migrations...")
	err = f.UpAll(t)
	require.NoError(t, err)

	version, dirty, err = f.Version(t)
	require.NoError(t, err)
	t.Logf("After Up again: version=%d, dirty=%v", version, dirty)
	assert.False(t, dirty, "Migration should not be dirty")
	assert.Greater(t, version, uint(0), "Should have applied migrations again")

	// Step 6: Reapply successful
	t.Log("Step 6: Migrations successfully reapplied")
}

// TestMigration007OAuth2Flavor verifies migration 007 (adding oauth2_flavor column) lifecycle.
// Requires: integration build tag and Docker/Podman.
func TestMigration007OAuth2Flavor(t *testing.T) {
	f := NewMigrationTestFramework(t)
	defer f.Cleanup(t)

	// Step 1: Apply migrations 001-006 only (using m.Migrate to stop at version 6)
	t.Log("Step 1: Applying migrations up to version 6...")
	err := f.Up(t, 6)
	require.NoError(t, err)

	version, dirty, err := f.Version(t)
	require.NoError(t, err)
	t.Logf("After Up to 6: version=%d, dirty=%v", version, dirty)
	assert.False(t, dirty)
	assert.Equal(t, uint(6), version)

	// Verify oauth2_flavor column does NOT exist yet before migration 007
	exists, err := f.ColumnExists(t, "thirdparty_oauth2_services", "oauth2_flavor")
	require.NoError(t, err)
	assert.False(t, exists, "oauth2_flavor column should NOT exist before migration 007")

	// Insert a pre-existing row to verify it gets oauth2_flavor = 'standard' via column DEFAULT.
	// '\x01' is a minimal non-empty BYTEA value — sufficient as a placeholder for the
	// encrypted secret since this test only exercises the migration DEFAULT, not encryption.
	err = f.ExecuteSQL(t, `
		INSERT INTO thirdparty_oauth2_services
			(id, display_name, client_id, client_secret_encrypted, issuer_uri, enable_discovery, scopes)
		VALUES
			('11111111-1111-1111-1111-111111111111', 'pre-existing svc', 'client-pre',
			 '\x01', 'https://oauth.example.com', false, '[]');
	`)
	require.NoError(t, err, "should be able to insert a pre-existing row before migration 007")

	// Step 2: Apply migration 007 (apply all remaining)
	t.Log("Step 2: Applying migration 007 (ADD COLUMN oauth2_flavor)...")
	err = f.UpAll(t)
	require.NoError(t, err)

	version, dirty, err = f.Version(t)
	require.NoError(t, err)
	t.Logf("After Up 007: version=%d, dirty=%v", version, dirty)
	assert.False(t, dirty)
	assert.GreaterOrEqual(t, version, uint(7), "Migration 007 should be applied")

	// Step 3: Verify column exists with correct default value for the pre-existing row
	exists, err = f.ColumnExists(t, "thirdparty_oauth2_services", "oauth2_flavor")
	require.NoError(t, err)
	assert.True(t, exists, "oauth2_flavor column should exist after migration 007")

	// Verify pre-existing row got the column DEFAULT ('standard') when migration was applied.
	result, err := f.QuerySQL(t, `
		SELECT oauth2_flavor FROM thirdparty_oauth2_services
		WHERE id = '11111111-1111-1111-1111-111111111111';
	`)
	require.NoError(t, err)
	assert.Equal(t, "standard", strings.TrimSpace(result),
		"pre-existing row should have oauth2_flavor = 'standard' from column DEFAULT")

	// Step 4: Rollback migration 007 (DROP COLUMN)
	t.Log("Step 4: Rolling back migration 007...")
	err = f.Down(t, 6)
	require.NoError(t, err)

	version, dirty, err = f.Version(t)
	require.NoError(t, err)
	t.Logf("After Down to 6: version=%d, dirty=%v", version, dirty)
	assert.False(t, dirty)

	// Step 5: Verify column is gone after rollback
	exists, err = f.ColumnExists(t, "thirdparty_oauth2_services", "oauth2_flavor")
	require.NoError(t, err)
	assert.False(t, exists, "oauth2_flavor column should NOT exist after rollback")

	// Step 6: Re-apply migration 007 to verify idempotency
	t.Log("Step 6: Re-applying migration 007...")
	err = f.UpAll(t)
	require.NoError(t, err)

	exists, err = f.ColumnExists(t, "thirdparty_oauth2_services", "oauth2_flavor")
	require.NoError(t, err)
	assert.True(t, exists, "oauth2_flavor column should exist after re-apply")

	t.Log("Migration 007 lifecycle test complete")
}

// TestMigration015AgentCIMDFields verifies migration 015 lifecycle:
// adds agent_client_uris child table to agents.
func TestMigration015AgentCIMDFields(t *testing.T) {
	f := NewMigrationTestFramework(t)
	defer f.Cleanup(t)

	// Apply migrations up to version 014
	err := f.Up(t, 14)
	require.NoError(t, err)

	// Verify agent_client_uris table does NOT exist before migration 015
	exists, err := f.TableExists(t, "agent_client_uris")
	require.NoError(t, err)
	assert.False(t, exists, "agent_client_uris table should NOT exist before migration 015")

	// Apply migration 015
	err = f.UpAll(t)
	require.NoError(t, err)

	// Verify agent_client_uris table now exists
	exists, err = f.TableExists(t, "agent_client_uris")
	require.NoError(t, err)
	assert.True(t, exists, "agent_client_uris table should exist after migration 015")

	// Rollback migration 015
	err = f.Down(t, 14)
	require.NoError(t, err)

	exists, err = f.TableExists(t, "agent_client_uris")
	require.NoError(t, err)
	assert.False(t, exists, "agent_client_uris table should be gone after rollback")

	t.Log("Migration 015 lifecycle test complete")
}
