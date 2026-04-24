//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// init disables Ryuk for Podman compatibility
func init() {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// canAccessContainerRuntime checks if Docker or Podman is available
func canAccessContainerRuntime() bool {
	cmd := exec.Command("docker", "ps")
	if err := cmd.Run(); err == nil {
		return true
	}
	cmd = exec.Command("podman", "ps")
	return cmd.Run() == nil
}

// setupTestContainer creates a PostgreSQL test container
func setupTestContainer(t *testing.T) (testcontainers.Container, string, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if !canAccessContainerRuntime() {
		t.Skip("Skipping test: No container runtime available")
	}

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
		Networks: []string{"podman"},
	}

	genericReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	if os.Getenv("DOCKER_HOST") != "" {
		genericReq.ProviderType = testcontainers.ProviderPodman
	}

	container, err := testcontainers.GenericContainer(ctx, genericReq)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	cleanup := func() {
		container.Terminate(ctx)
	}

	return container, connStr, cleanup
}

// applyMigrations applies all database migrations to the test container.
// Files are copied into the container and executed via `psql -f` to avoid
// any issues with passing multi-statement SQL as a command-line argument.
func applyMigrations(t *testing.T, container testcontainers.Container) {
	t.Helper()
	applyMigrationsUpTo(t, container, 16)
}

// applyMigrationsUpTo applies migrations sequentially from 001 up to and including
// the migration with the given version number.
func applyMigrationsUpTo(t *testing.T, container testcontainers.Container, upTo int) {
	t.Helper()

	ctx := context.Background()

	createSchemaMigrationsTable(t, ctx, container)

	// Find project root (where migrations folder is)
	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	migrationsDir := filepath.Join(projectRoot, "migrations")

	migrations := []struct {
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
		{"009_add_agent_redirect_uris.up.sql", 9},
		{"010_create_client_credentials.up.sql", 10},
		{"011_create_signing_keys.up.sql", 11},
		{"012_create_authorization_codes.up.sql", 12},
		{"013_add_client_id_to_auth_codes.up.sql", 13},
		{"014_create_pkce_sessions.up.sql", 14},
		{"015_add_agent_cimd_fields.up.sql", 15},
		{"016_add_cimd_redirect_uris.up.sql", 16},
	}

	for _, migration := range migrations {
		if int(migration.version) > upTo {
			break
		}
		applyOneMigration(t, ctx, container, migrationsDir, migration.file, migration.version)
	}
}

// createSchemaMigrationsTable creates the schema_migrations tracking table in the container.
func createSchemaMigrationsTable(t *testing.T, ctx context.Context, container testcontainers.Container) {
	t.Helper()
	schemaSQL := []byte(`CREATE TABLE IF NOT EXISTS schema_migrations (version BIGINT PRIMARY KEY, dirty BOOLEAN NOT NULL DEFAULT FALSE);`)
	if err := container.CopyToContainer(ctx, schemaSQL, "/tmp/schema_migrations.sql", 0644); err != nil {
		t.Logf("Warning: Failed to copy schema_migrations.sql: %v", err)
		return
	}
	exitCode, _, err := container.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-f", "/tmp/schema_migrations.sql"})
	if err != nil || exitCode != 0 {
		t.Logf("Warning: Failed to create schema_migrations table (exit %d): %v", exitCode, err)
	}
}

// applyOneMigration copies a migration file into the container and runs it via psql -f.
func applyOneMigration(t *testing.T, ctx context.Context, container testcontainers.Container, migrationsDir, file string, version int64) {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(migrationsDir, file))
	if err != nil {
		t.Logf("Warning: Could not read migration %s: %v", file, err)
		return
	}

	// Copy the SQL file into the container so psql can read it with -f (avoids
	// any quoting or argument-length issues with psql -c "<multiline SQL>").
	containerPath := fmt.Sprintf("/tmp/migration_%03d.sql", version)
	if err := container.CopyToContainer(ctx, data, containerPath, 0644); err != nil {
		t.Logf("Warning: Could not copy migration %s to container: %v", file, err)
		return
	}

	exitCode, _, err := container.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-f", containerPath})
	if err != nil || exitCode != 0 {
		t.Logf("Warning: Migration %s failed (exit %d): %v", file, exitCode, err)
		return
	}

	// Record migration version in schema_migrations
	versionSQL := []byte(fmt.Sprintf("INSERT INTO schema_migrations (version, dirty) VALUES (%d, FALSE) ON CONFLICT DO NOTHING;", version))
	versionPath := fmt.Sprintf("/tmp/migration_%03d_version.sql", version)
	if err := container.CopyToContainer(ctx, versionSQL, versionPath, 0644); err == nil {
		container.Exec(ctx, []string{"psql", "-U", "testuser", "-d", "testdb", "-f", versionPath}) //nolint:errcheck
	}
}

// findProjectRoot walks up the directory tree to find the project root
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

// setupAgentTestDBWithCIMD creates a test database with migrations applied up to and including
// migration 016 (agent_client_uris, CIMD fields, and cimd_redirect_uris). Use for tests that exercise ClientURIs.
func setupAgentTestDBWithCIMD(t *testing.T) (*Adapter, func()) {
	t.Helper()

	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)

	applyMigrationsUpTo(t, container, 16)

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

// testStorageConfig creates a StorageConfig for integration testing with the given connection string.
func testStorageConfig(connString string) *ports.StorageConfig {
	return &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connString,
		},
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}
}
