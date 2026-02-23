//go:build integration
// +build integration

package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

// applyMigrations applies database migrations to the test container
func applyMigrations(t *testing.T, container testcontainers.Container) {
	t.Helper()

	ctx := context.Background()

	// First, create schema_migrations table
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
		t.Logf("Warning: Failed to create schema_migrations table (exit %d): %v", exitCode, err)
	}

	// Find project root (where migrations folder is)
	projectRoot, err := findProjectRoot()
	require.NoError(t, err)

	migrationsDir := filepath.Join(projectRoot, "migrations")

	// Read and apply each migration file
	migrations := []struct {
		file    string
		version int64
	}{
		{"001_create_agents.up.sql", 1},
		{"002_create_thirdparty_services.up.sql", 2},
		{"003_create_user_grants.up.sql", 3},
	}

	for _, migration := range migrations {
		migrationPath := filepath.Join(migrationsDir, migration.file)
		data, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Logf("Warning: Could not read migration %s: %v", migration.file, err)
			continue
		}

		// Execute migration SQL directly in container
		exitCode, _, err := container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-c", string(data),
		})

		if err != nil || exitCode != 0 {
			t.Logf("Warning: Migration %s failed (exit %d): %v", migration.file, exitCode, err)
			continue
		}

		// Record migration version
		versionSQL := fmt.Sprintf("INSERT INTO schema_migrations (version, dirty) VALUES (%d, FALSE) ON CONFLICT DO NOTHING;", migration.version)
		container.Exec(ctx, []string{
			"psql",
			"-U", "testuser",
			"-d", "testdb",
			"-c", versionSQL,
		})
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

// createServiceWithManager uses ServiceManager to create a service via encryption layer,
// ensuring proper encryption context binding. This simulates the domain layer behavior.
func createServiceWithManager(t *testing.T, ctx context.Context, serviceManager *thirdparty.ServiceManager, service *storage.ThirdpartyOAuth2Service) {
	t.Helper()
	_, err := serviceManager.Create(ctx, service)
	require.NoError(t, err)
}
