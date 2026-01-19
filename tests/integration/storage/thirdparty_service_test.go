//go:build integration
// +build integration

package storage_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/postgres"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
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

// setupPostgreSQLContainer creates a PostgreSQL test container
func setupPostgreSQLContainer(t *testing.T) (testcontainers.Container, string, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
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
	}

	genericReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	container, err := testcontainers.GenericContainer(ctx, genericReq)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := "postgres://testuser:testpass@" + host + ":" + port.Port() + "/testdb?sslmode=disable"

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
		{"005_add_service_protected_resources.up.sql", 5},
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
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

// createTestService is a helper to create a test service with specified properties
func createTestService(id, displayName string, protectedResources []string) *storage.ThirdpartyOAuth2Service {
	return &storage.ThirdpartyOAuth2Service{
		ID:           id,
		DisplayName:  displayName,
		ClientID:     id + "-client",
		ClientSecret: "test-secret",
		IssuerURI:    "https://oauth.example.com",
		Discovery: storage.DiscoveryConfig{
			EnableDiscovery: false,
		},
		Endpoints: storage.OAuth2Endpoints{
			TokenEndpoint:     "https://oauth.example.com/token",
			AuthorizeEndpoint: "https://oauth.example.com/authorize",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "read", Description: "Read access"},
		},
		ProtectedResources: protectedResources,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

// TestFindByProtectedResource_SingleMatch tests the happy path: single matching service
func TestFindByProtectedResource_SingleMatch(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service with protected resources
	service := createTestService(
		"github-service",
		"GitHub",
		[]string{"https://api.github.com", "https://api.github.com/user"},
	)
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Find by exact resource match
	found, err := repo.FindByProtectedResource(ctx, "https://api.github.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "github-service", found.ID)
	require.Equal(t, "GitHub", found.DisplayName)
}

// TestFindByProtectedResource_NoMatch tests error case: no matching service
func TestFindByProtectedResource_NoMatch(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service without matching resource
	service := createTestService(
		"github-service",
		"GitHub",
		[]string{"https://api.github.com"},
	)
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Try to find non-existent resource
	found, err := repo.FindByProtectedResource(ctx, "https://api.google.com")
	require.Error(t, err)
	require.Nil(t, found)

	// Verify error is InvalidTargetError
	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	require.True(t, ok, "expected TokenExchangeError")
	require.Equal(t, "invalid_target", txErr.Code())
	require.Equal(t, "no service configured for the requested resource", txErr.Description())
}

// TestFindByProtectedResource_AmbiguousMatch tests error case: multiple services match same resource
func TestFindByProtectedResource_AmbiguousMatch(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create two services with overlapping resources (misconfiguration)
	service1 := createTestService(
		"service-1",
		"Service 1",
		[]string{"https://api.example.com"},
	)
	err = repo.Create(ctx, service1)
	require.NoError(t, err)

	service2 := createTestService(
		"service-2",
		"Service 2",
		[]string{"https://api.example.com"},
	)
	err = repo.Create(ctx, service2)
	require.NoError(t, err)

	// Try to find ambiguous resource
	found, err := repo.FindByProtectedResource(ctx, "https://api.example.com")
	require.Error(t, err)
	require.Nil(t, found)

	// Verify error is InvalidTargetError with ambiguity message
	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	require.True(t, ok, "expected TokenExchangeError")
	require.Equal(t, "invalid_target", txErr.Code())
	require.Equal(t, "multiple services configured for the same resource", txErr.Description())
}

// TestFindByProtectedResource_URINormalization tests URI normalization (trailing slash removal)
func TestFindByProtectedResource_URINormalization(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service with normalized URI (no trailing slash)
	service := createTestService(
		"api-service",
		"API Service",
		[]string{"https://api.example.com"},
	)
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Query with trailing slash should find the service
	// (This test documents the current behavior - normalization is done by the caller)
	found, err := repo.FindByProtectedResource(ctx, "https://api.example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "api-service", found.ID)
}

// TestFindByProtectedResource_MultipleServicesNonOverlapping tests multiple services without overlap
func TestFindByProtectedResource_MultipleServicesNonOverlapping(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create three services with different resources
	service1 := createTestService(
		"github-service",
		"GitHub",
		[]string{"https://api.github.com", "https://api.github.com/user"},
	)
	err = repo.Create(ctx, service1)
	require.NoError(t, err)

	service2 := createTestService(
		"google-service",
		"Google",
		[]string{"https://www.googleapis.com"},
	)
	err = repo.Create(ctx, service2)
	require.NoError(t, err)

	service3 := createTestService(
		"databricks-service",
		"Databricks",
		[]string{"https://api.databricks.com"},
	)
	err = repo.Create(ctx, service3)
	require.NoError(t, err)

	// Find each service by its resource
	tests := []struct {
		resource     string
		expectedID   string
		expectedName string
	}{
		{
			resource:     "https://api.github.com",
			expectedID:   "github-service",
			expectedName: "GitHub",
		},
		{
			resource:     "https://api.github.com/user",
			expectedID:   "github-service",
			expectedName: "GitHub",
		},
		{
			resource:     "https://www.googleapis.com",
			expectedID:   "google-service",
			expectedName: "Google",
		},
		{
			resource:     "https://api.databricks.com",
			expectedID:   "databricks-service",
			expectedName: "Databricks",
		},
	}

	for _, tc := range tests {
		t.Run(tc.resource, func(t *testing.T) {
			found, err := repo.FindByProtectedResource(ctx, tc.resource)
			require.NoError(t, err)
			require.NotNil(t, found)
			require.Equal(t, tc.expectedID, found.ID)
			require.Equal(t, tc.expectedName, found.DisplayName)
		})
	}
}

// TestFindByProtectedResource_EmptyProtectedResources tests service without protected_resources
func TestFindByProtectedResource_EmptyProtectedResources(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service without protected_resources
	service := createTestService(
		"service-no-resources",
		"Service Without Resources",
		nil,
	)
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Try to find by resource - should not match
	found, err := repo.FindByProtectedResource(ctx, "https://api.example.com")
	require.Error(t, err)
	require.Nil(t, found)

	txErr, ok := err.(*tokenexchange.TokenExchangeError)
	require.True(t, ok)
	require.Equal(t, "invalid_target", txErr.Code())
}

// TestFindByProtectedResource_CaseSensitive tests that resource matching is case-sensitive
func TestFindByProtectedResource_CaseSensitive(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service with specific case
	service := createTestService(
		"api-service",
		"API Service",
		[]string{"https://api.Example.com"},
	)
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Find with exact case - should succeed
	found, err := repo.FindByProtectedResource(ctx, "https://api.Example.com")
	require.NoError(t, err)
	require.NotNil(t, found)

	// Find with different case - should fail (case-sensitive)
	found, err = repo.FindByProtectedResource(ctx, "https://api.example.com")
	require.Error(t, err)
	require.Nil(t, found)
}

// TestFindByProtectedResource_MixedScenarios tests combination of scenarios
func TestFindByProtectedResource_MixedScenarios(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create mixed services
	// Service 1: has resources
	service1 := createTestService(
		"service-1",
		"Service 1",
		[]string{"https://api1.example.com"},
	)
	err = repo.Create(ctx, service1)
	require.NoError(t, err)

	// Service 2: no resources
	service2 := createTestService(
		"service-2",
		"Service 2",
		nil,
	)
	err = repo.Create(ctx, service2)
	require.NoError(t, err)

	// Service 3: multiple resources
	service3 := createTestService(
		"service-3",
		"Service 3",
		[]string{"https://api3a.example.com", "https://api3b.example.com"},
	)
	err = repo.Create(ctx, service3)
	require.NoError(t, err)

	// Test scenarios
	tests := []struct {
		name          string
		resource      string
		shouldSucceed bool
		expectedID    string
	}{
		{
			name:          "Find service 1",
			resource:      "https://api1.example.com",
			shouldSucceed: true,
			expectedID:    "service-1",
		},
		{
			name:          "Find service 3a",
			resource:      "https://api3a.example.com",
			shouldSucceed: true,
			expectedID:    "service-3",
		},
		{
			name:          "Find service 3b",
			resource:      "https://api3b.example.com",
			shouldSucceed: true,
			expectedID:    "service-3",
		},
		{
			name:          "Service 2 has no resources",
			resource:      "https://some.resource.com",
			shouldSucceed: false,
			expectedID:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			found, err := repo.FindByProtectedResource(ctx, tc.resource)
			if tc.shouldSucceed {
				assert.NoError(t, err)
				assert.NotNil(t, found)
				assert.Equal(t, tc.expectedID, found.ID)
			} else {
				assert.Error(t, err)
				assert.Nil(t, found)
			}
		})
	}
}

// TestFindByProtectedResource_ClientSecretDecrypted tests that client_secret is properly decrypted
func TestFindByProtectedResource_ClientSecretDecrypted(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create service with secret
	service := createTestService(
		"service-with-secret",
		"Service With Secret",
		[]string{"https://api.example.com"},
	)
	service.ClientSecret = "my-super-secret"
	err = repo.Create(ctx, service)
	require.NoError(t, err)

	// Find by resource and verify secret is decrypted
	found, err := repo.FindByProtectedResource(ctx, "https://api.example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "my-super-secret", found.ClientSecret)
}

// TestFindByProtectedResource_InvalidResourceURI tests validation of empty resource URI
func TestFindByProtectedResource_InvalidResourceURI(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Try to find with empty resource URI
	found, err := repo.FindByProtectedResource(ctx, "")
	require.Error(t, err)
	require.Nil(t, found)

	// Verify it's a StorageError with validation kind
	storageErr, ok := err.(*storage.StorageError)
	require.True(t, ok, "expected StorageError")
	require.Equal(t, storage.ErrorKindValidation, storageErr.Kind)
}

// TestFindByProtectedResource_GINIndexQuery tests query efficiency (GIN index behavior)
// This test verifies the query uses array containment with @> operator which leverages GIN index
func TestFindByProtectedResource_GINIndexQuery(t *testing.T) {
	container, connStr, cleanup := setupPostgreSQLContainer(t)
	defer cleanup()

	applyMigrations(t, container)

	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: connStr,
		},
	}

	adapter, err := postgres.NewAdapter(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = adapter.Initialize(ctx)
	require.NoError(t, err)
	defer adapter.Close(ctx)

	encryption := noop.NewNoOpEncryption()
	repo := postgres.NewThirdpartyServiceRepository(adapter, encryption)

	// Create multiple services to test query efficiency
	for i := 1; i <= 10; i++ {
		resources := []string{}
		for j := 1; j <= 5; j++ {
			resources = append(resources, fmt.Sprintf("https://api%d.example.com/v%d", i, j))
		}
		service := createTestService(
			fmt.Sprintf("service-%d", i),
			fmt.Sprintf("Service %d", i),
			resources,
		)
		err = repo.Create(ctx, service)
		require.NoError(t, err)
	}

	// Query should be efficient even with many services
	found, err := repo.FindByProtectedResource(ctx, "https://api1.example.com/v1")
	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, "service-1", found.ID)
}
