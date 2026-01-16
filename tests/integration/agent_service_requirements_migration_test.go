//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestAgentServiceRequirementsMigration tests migration 005
// This test verifies that migration 005 applies cleanly and creates the service_requirements column
func TestAgentServiceRequirementsMigration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !canAccessContainerRuntime() {
		t.Skip("Skipping test: No container runtime available")
	}

	ctx := context.Background()
	container, cleanup := setupTestContainer(t)
	defer cleanup()

	// Apply migrations 001-004 (baseline)
	applyBaseMigrations(t, container)

	t.Run("migration 005 applies cleanly", func(t *testing.T) {
		// Apply migration 005
		applyMigration(t, container, "005_add_agent_service_requirements.up.sql")

		// Verify column exists
		checkSQL := `
			SELECT column_name, data_type, is_nullable, column_default
			FROM information_schema.columns
			WHERE table_name = 'agents'
			  AND column_name = 'service_requirements';
		`
		exitCode, outputReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t", // Tuples only
			"-c", checkSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode, "Query should succeed")

		// Read output
		outputBytes, err := io.ReadAll(outputReader)
		require.NoError(t, err)
		output := string(outputBytes)

		assert.Contains(t, output, "service_requirements", "Column should exist")
		assert.Contains(t, output, "jsonb", "Column should be JSONB type")
		assert.Contains(t, output, "YES", "Column should be nullable")
	})

	t.Run("GIN index is created", func(t *testing.T) {
		// Verify index exists
		checkIndexSQL := `
			SELECT indexname, indexdef
			FROM pg_indexes
			WHERE tablename = 'agents'
			  AND indexname = 'idx_agents_service_requirements';
		`
		exitCode, outputReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", checkIndexSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		outputBytes, err := io.ReadAll(outputReader)
		require.NoError(t, err)
		output := string(outputBytes)

		assert.Contains(t, output, "idx_agents_service_requirements", "Index should exist")
		assert.Contains(t, output, "USING gin", "Index should be GIN type")
	})

	t.Run("existing agents have NULL service_requirements", func(t *testing.T) {
		// Create test agent
		insertSQL := `
			INSERT INTO agents (client_id, display_name, description, created_at, updated_at)
			VALUES ('test-client', 'Test Agent', 'Test Description', NOW(), NOW())
			RETURNING id;
		`
		exitCode, agentIDReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", insertSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		agentIDBytes, err := io.ReadAll(agentIDReader)
		require.NoError(t, err)
		require.NotEmpty(t, string(agentIDBytes))

		// Verify service_requirements is NULL
		checkSQL := `
			SELECT service_requirements IS NULL as is_null
			FROM agents
			WHERE client_id = 'test-client';
		`
		exitCode, outputReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", checkSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		outputBytes, err := io.ReadAll(outputReader)
		require.NoError(t, err)
		output := string(outputBytes)

		assert.Contains(t, output, "t", "service_requirements should be NULL for new agents")
	})
}

// TestAgentServiceRequirementsMigrationRollback tests migration 005 rollback
func TestAgentServiceRequirementsMigrationRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !canAccessContainerRuntime() {
		t.Skip("Skipping test: No container runtime available")
	}

	ctx := context.Background()
	container, cleanup := setupTestContainer(t)
	defer cleanup()

	// Apply migrations 001-005
	applyBaseMigrations(t, container)
	applyMigration(t, container, "005_add_agent_service_requirements.up.sql")

	t.Run("migration 005 rollback removes column and index", func(t *testing.T) {
		// Apply rollback
		applyMigration(t, container, "005_add_agent_service_requirements.down.sql")

		// Verify column is removed
		checkSQL := `
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_name = 'agents'
			  AND column_name = 'service_requirements';
		`
		exitCode, outputReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", checkSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		outputBytes, err := io.ReadAll(outputReader)
		require.NoError(t, err)
		output := string(outputBytes)

		assert.Contains(t, output, "0", "Column should not exist after rollback")

		// Verify index is removed
		checkIndexSQL := `
			SELECT COUNT(*)
			FROM pg_indexes
			WHERE tablename = 'agents'
			  AND indexname = 'idx_agents_service_requirements';
		`
		exitCode, outputReader2, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", checkIndexSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		output2Bytes, err := io.ReadAll(outputReader2)
		require.NoError(t, err)
		output2 := string(output2Bytes)

		assert.Contains(t, output2, "0", "Index should not exist after rollback")
	})

	t.Run("can reapply migration after rollback", func(t *testing.T) {
		// Reapply migration 005
		applyMigration(t, container, "005_add_agent_service_requirements.up.sql")

		// Verify column exists again
		checkSQL := `
			SELECT COUNT(*)
			FROM information_schema.columns
			WHERE table_name = 'agents'
			  AND column_name = 'service_requirements';
		`
		exitCode, outputReader, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-t",
			"-c", checkSQL,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		outputBytes, err := io.ReadAll(outputReader)
		require.NoError(t, err)
		output := string(outputBytes)

		assert.Contains(t, output, "1", "Column should exist after reapplying migration")
	})
}

// Helper functions

func canAccessContainerRuntime() bool {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err == nil {
		return true
	}
	cmd = exec.Command("podman", "ps")
	return cmd.Run() == nil
}

func setupTestContainer(t *testing.T) (testcontainers.Container, func()) {
	t.Helper()

	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	// Disable Ryuk for Podman compatibility
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}

	genericReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	container, err := testcontainers.GenericContainer(ctx, genericReq)
	require.NoError(t, err)

	cleanup := func() {
		container.Terminate(ctx)
	}

	return container, cleanup
}

func applyBaseMigrations(t *testing.T, container testcontainers.Container) {
	t.Helper()

	// Create schema_migrations table
	ctx := context.Background()
	schemaSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		);
	`
	exitCode, _, err := container.Exec(ctx, []string{
		"psql",
		"-U", "testuser",
		"-d", "testdb",
		"-c", schemaSQL,
	})
	if err != nil || exitCode != 0 {
		t.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Apply migrations 001-004
	migrations := []string{
		"001_create_agents.up.sql",
		"002_create_thirdparty_services.up.sql",
		"003_create_user_grants.up.sql",
		"004_create_user_sessions.up.sql",
	}

	for _, migrationFile := range migrations {
		applyMigration(t, container, migrationFile)
	}
}

func applyMigration(t *testing.T, container testcontainers.Container, filename string) {
	t.Helper()

	ctx := context.Background()

	// Find project root
	projectRoot, err := findProjectRoot()
	require.NoError(t, err, "Failed to find project root")

	migrationPath := filepath.Join(projectRoot, "migrations", filename)
	data, err := os.ReadFile(migrationPath)
	require.NoError(t, err, "Failed to read migration file %s", filename)

	// Execute migration SQL
	exitCode, outputReader, err := container.Exec(ctx, []string{
		"psql",
		"-U", "testuser",
		"-d", "testdb",
		"-c", string(data),
	})

	if err != nil || exitCode != 0 {
		outputBytes, _ := io.ReadAll(outputReader)
		t.Logf("Migration output: %s", string(outputBytes))
		t.Fatalf("Migration %s failed (exit %d): %v", filename, exitCode, err)
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root")
		}
		dir = parent
	}
}
