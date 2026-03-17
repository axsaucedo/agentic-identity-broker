//go:build integration
// +build integration

package migrations_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// init disables Ryuk for Podman/Docker compatibility
// Ryuk tries to use "bridge" network which may not be available in some Docker configurations
func init() {
	if os.Getenv("TESTCONTAINERS_RYUK_DISABLED") == "" {
		os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	}
}

// MigrationTestFramework provides reusable migration testing infrastructure
type MigrationTestFramework struct {
	container     testcontainers.Container
	connStr       string
	migrationsDir string
}

// NewMigrationTestFramework creates a test database and initializes the framework
func NewMigrationTestFramework(t *testing.T) *MigrationTestFramework {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Find migrations directory relative to this test file
	migrationsDir := filepath.Join("..", "..", "..", "migrations")

	// Convert to absolute path for reliable go-migrate resolution
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		t.Fatalf("Failed to resolve migrations directory: %v", err)
	}
	migrationsDir = absPath

	// Setup PostgreSQL container
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

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := fmt.Sprintf(
		"postgres://testuser:testpass@%s:%s/testdb?sslmode=disable",
		host, port.Port(),
	)

	return &MigrationTestFramework{
		container:     container,
		connStr:       connStr,
		migrationsDir: migrationsDir,
	}
}

// Cleanup terminates the test database
func (f *MigrationTestFramework) Cleanup(t *testing.T) {
	t.Helper()
	f.container.Terminate(context.Background())
}

// Up runs migrations up to the specified version
func (f *MigrationTestFramework) Up(t *testing.T, targetVersion uint) error {
	t.Helper()

	// Use file:/// (with 3 slashes) for absolute path
	migrationsURL := "file://" + f.migrationsDir
	t.Logf("Applying migrations up to version %d from: %s", targetVersion, migrationsURL)

	m, err := migrate.New(migrationsURL, f.connStr)
	require.NoErrorf(t, err, "failed to create migrate instance")

	defer m.Close()

	if err := m.Migrate(targetVersion); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration to version %d failed: %w", targetVersion, err)
	}

	return nil
}

// UpAll runs all available migrations
func (f *MigrationTestFramework) UpAll(t *testing.T) error {
	t.Helper()

	migrationsURL := "file://" + f.migrationsDir
	t.Logf("Migration source URL: %s", migrationsURL)
	t.Logf("Migrations directory exists: %s", f.migrationsDir)

	// List the files in the migrations directory
	files, err := filepath.Glob(filepath.Join(f.migrationsDir, "*.sql"))
	if err == nil {
		t.Logf("Found %d SQL files:", len(files))
		for _, file := range files {
			t.Logf("  - %s", filepath.Base(file))
		}
	}

	m, err := migrate.New(migrationsURL, f.connStr)
	require.NoErrorf(t, err, "failed to create migrate instance")

	defer m.Close()

	// Log available migrations
	s, d, err := m.Version()
	if err != migrate.ErrNilVersion {
		t.Logf("Current migration version: source=%d, dirty=%v", s, d)
	}

	// Apply all migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	// Log final state
	s, d, err = m.Version()
	if err != migrate.ErrNilVersion && err != nil {
		return err
	}
	t.Logf("Final state after Up: version=%d, dirty=%v", s, d)

	return nil
}

// Down rolls back migrations down to the specified version
func (f *MigrationTestFramework) Down(t *testing.T, targetVersion uint) error {
	t.Helper()

	m, err := migrate.New("file://"+f.migrationsDir, f.connStr)
	require.NoErrorf(t, err, "failed to create migrate instance")

	defer m.Close()

	// Rollback to specific version using Steps
	// Get current version to calculate steps
	currentVersion, _, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// Calculate steps needed to reach target version
	steps := int(currentVersion) - int(targetVersion)
	if steps > 0 {
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("migration down failed: %w", err)
		}
	}

	return nil
}

// DownAll rolls back all migrations
func (f *MigrationTestFramework) DownAll(t *testing.T) error {
	t.Helper()

	migrationsURL := "file://" + f.migrationsDir
	t.Logf("Rolling back all migrations from: %s", migrationsURL)

	m, err := migrate.New(migrationsURL, f.connStr)
	require.NoErrorf(t, err, "failed to create migrate instance")

	defer m.Close()

	// Log current state
	s, d, err := m.Version()
	if err != migrate.ErrNilVersion {
		t.Logf("Before Down: source=%d, dirty=%v", s, d)
	}

	// Roll back all migrations using Steps
	for i := 0; i < 10; i++ {
		currentVersion, _, err := m.Version()
		if err == migrate.ErrNilVersion {
			t.Logf("All migrations rolled back, at version 0")
			break
		}
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}

		t.Logf("Rolling back from version %d...", currentVersion)

		if err := m.Steps(-1); err != nil {
			if err == migrate.ErrNoChange {
				t.Logf("No more migrations to rollback")
				break
			}
			return fmt.Errorf("rollback step %d failed: %w", i+1, err)
		}

		nextVersion, _, err := m.Version()
		if err == migrate.ErrNilVersion {
			t.Logf("After rollback step %d: version=0 (no migrations)", i+1)
		} else if err != nil {
			return fmt.Errorf("failed to get version after rollback: %w", err)
		} else {
			t.Logf("After rollback step %d: version=%d", i+1, nextVersion)
		}
	}

	// Log final state
	s, d, err = m.Version()
	if err == migrate.ErrNilVersion {
		t.Logf("After Down: version=0 (no migrations), dirty=%v", d)
		return nil
	}
	if err != nil {
		return err
	}
	t.Logf("After Down: source=%d, dirty=%v", s, d)

	return nil
}

// Version returns the current migration version
func (f *MigrationTestFramework) Version(t *testing.T) (uint, bool, error) {
	t.Helper()

	m, err := migrate.New("file://"+f.migrationsDir, f.connStr)
	if err != nil {
		return 0, false, err
	}

	defer m.Close()

	version, dirty, err := m.Version()
	// If no migrations have been applied yet, m.Version() returns ErrNilVersion
	// In this case, we should return version 0
	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}
	return version, dirty, err
}

// QuerySQL executes a SQL query and returns results
func (f *MigrationTestFramework) QuerySQL(t *testing.T, query string) (string, error) {
	t.Helper()

	ctx := context.Background()
	exitCode, outReader, err := f.container.Exec(ctx, []string{
		"psql",
		"-U", "testuser",
		"-d", "testdb",
		"-t", // Tuples only (no headers)
		"-c", query,
	})

	if exitCode != 0 {
		return "", fmt.Errorf("psql exited with code %d: %v", exitCode, err)
	}

	output, err := io.ReadAll(outReader)
	if err != nil {
		return "", fmt.Errorf("failed to read psql output: %v", err)
	}

	return string(output), nil
}

// ExecuteSQL executes a SQL statement (no output)
func (f *MigrationTestFramework) ExecuteSQL(t *testing.T, sql string) error {
	t.Helper()

	ctx := context.Background()
	exitCode, _, err := f.container.Exec(ctx, []string{
		"psql",
		"-U", "testuser",
		"-d", "testdb",
		"-c", sql,
	})

	if exitCode != 0 {
		return fmt.Errorf("psql exited with code %d: %v", exitCode, err)
	}

	return nil
}

// ColumnExists checks if a column exists in a table
func (f *MigrationTestFramework) ColumnExists(t *testing.T, table, column string) (bool, error) {
	t.Helper()

	result, err := f.QuerySQL(t, fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.columns
		WHERE table_name = '%s' AND column_name = '%s';
	`, table, column))
	if err != nil {
		return false, err
	}

	return result != "0", nil
}

// IndexExists checks if an index exists
func (f *MigrationTestFramework) IndexExists(t *testing.T, indexName string) (bool, error) {
	t.Helper()

	result, err := f.QuerySQL(t, fmt.Sprintf(`
		SELECT COUNT(*) FROM pg_indexes WHERE indexname = '%s';
	`, indexName))
	if err != nil {
		return false, err
	}

	return result != "0", nil
}

// GetColumnType returns the data type of a column
func (f *MigrationTestFramework) GetColumnType(t *testing.T, table, column string) (string, error) {
	t.Helper()

	result, err := f.QuerySQL(t, fmt.Sprintf(`
		SELECT data_type FROM information_schema.columns
		WHERE table_name = '%s' AND column_name = '%s';
	`, table, column))
	if err != nil {
		return "", err
	}

	return result, nil
}

// CountRows returns the number of rows in a table
func (f *MigrationTestFramework) CountRows(t *testing.T, table string, where ...string) (int64, error) {
	t.Helper()

	whereClause := ""
	if len(where) > 0 {
		whereClause = fmt.Sprintf("WHERE %s", where[0])
	}

	result, err := f.QuerySQL(t, fmt.Sprintf(`
		SELECT COUNT(*) FROM %s %s;
	`, table, whereClause))
	if err != nil {
		return 0, err
	}

	var count int64
	fmt.Sscanf(result, "%d", &count)
	return count, nil
}

// TableExists checks if a table exists
func (f *MigrationTestFramework) TableExists(t *testing.T, tableName string) (bool, error) {
	t.Helper()

	result, err := f.QuerySQL(t, fmt.Sprintf(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_name = '%s';
	`, tableName))
	if err != nil {
		return false, err
	}

	return result != "0", nil
}
