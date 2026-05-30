//go:build integration
// +build integration

package integration

import "testing"

// TestAgentRepositoryServiceRequirements_PostgreSQL tests service requirements with PostgreSQL adapter.
// NOTE: PostgreSQL integration coverage is provided by dedicated migration and adapter suites.
func TestAgentRepositoryServiceRequirements_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("PostgreSQL integration tests covered in agent_service_requirements_migration_test.go")
}
