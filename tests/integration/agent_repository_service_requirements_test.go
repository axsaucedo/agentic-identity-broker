//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	storageadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	domainStorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentRepositoryServiceRequirements_Memory tests service requirements with in-memory adapter
func TestAgentRepositoryServiceRequirements_Memory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create memory adapter using production factory
	config := &ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  0,
			Write: 0,
		},
	}
	adapter, err := storageadapter.NewAdapter(config)
	require.NoError(t, err)

	t.Run("create agent with service requirements", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		agent := &domainStorage.Agent{
			ID:          "agent-123",
			ClientID:    "client-sr-1",
			DisplayName: "Service Requirements Agent",
			Description: "Agent with service requirements",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			ServiceRequirements: []domainStorage.ServiceRequirement{
				{
					ServiceID:       "service-123",
					RequirementType: domainStorage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
			},
		}

		err := adapter.Agents().Create(ctx, agent)
		require.NoError(t, err)

		// Retrieve and verify
		retrieved, err := adapter.Agents().Get(ctx, agent.ID)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.Equal(t, agent.ClientID, retrieved.ClientID)
		assert.Len(t, retrieved.ServiceRequirements, 1)
		assert.Equal(t, "service-123", retrieved.ServiceRequirements[0].ServiceID)
		assert.Equal(t, domainStorage.RequirementTypeMandatory, retrieved.ServiceRequirements[0].RequirementType)
		assert.Equal(t, []string{"repo", "user:email"}, retrieved.ServiceRequirements[0].RequiredScopes)
	})

	t.Run("update agent service requirements", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create initial agent with one requirement
		agent := &domainStorage.Agent{
			ID:          "agent-456",
			ClientID:    "client-sr-2",
			DisplayName: "Update Test Agent",
			Description: "Agent for update testing",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			ServiceRequirements: []domainStorage.ServiceRequirement{
				{
					ServiceID:       "service-123",
					RequirementType: domainStorage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo"},
				},
			},
		}

		err := adapter.Agents().Create(ctx, agent)
		require.NoError(t, err)

		// Update with different requirements
		agent.ServiceRequirements = []domainStorage.ServiceRequirement{
			{
				ServiceID:       "service-456",
				RequirementType: domainStorage.RequirementTypeOptional,
				RequiredScopes:  []string{"read:user"},
			},
		}
		agent.UpdatedAt = time.Now().UTC()

		err = adapter.Agents().Update(ctx, agent)
		require.NoError(t, err)

		// Verify update
		retrieved, err := adapter.Agents().Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.ServiceRequirements, 1)
		assert.Equal(t, "service-456", retrieved.ServiceRequirements[0].ServiceID)
		assert.Equal(t, domainStorage.RequirementTypeOptional, retrieved.ServiceRequirements[0].RequirementType)
	})

	t.Run("agent with null service requirements (backward compatibility)", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Create agent without service requirements
		agent := &domainStorage.Agent{
			ID:          "agent-789",
			ClientID:    "client-sr-3",
			DisplayName: "Legacy Agent",
			Description: "Agent without service requirements",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			// ServiceRequirements is nil (NULL in DB)
		}

		err := adapter.Agents().Create(ctx, agent)
		require.NoError(t, err)

		// Retrieve and verify it's nil
		retrieved, err := adapter.Agents().Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Nil(t, retrieved.ServiceRequirements)
	})

	t.Run("agent with empty service requirements array", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		agent := &domainStorage.Agent{
			ID:                  "agent-empty",
			ClientID:            "client-sr-4",
			DisplayName:         "Empty Requirements Agent",
			Description:         "Agent with empty requirements array",
			CreatedAt:           time.Now().UTC(),
			UpdatedAt:           time.Now().UTC(),
			ServiceRequirements: []domainStorage.ServiceRequirement{},
		}

		err := adapter.Agents().Create(ctx, agent)
		require.NoError(t, err)

		retrieved, err := adapter.Agents().Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.NotNil(t, retrieved.ServiceRequirements)
		assert.Len(t, retrieved.ServiceRequirements, 0)
	})

	t.Run("agent with multiple service requirements", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		agent := &domainStorage.Agent{
			ID:          "agent-multi",
			ClientID:    "client-sr-5",
			DisplayName: "Multi Service Agent",
			Description: "Agent with multiple service requirements",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			ServiceRequirements: []domainStorage.ServiceRequirement{
				{
					ServiceID:       "github-service",
					RequirementType: domainStorage.RequirementTypeMandatory,
					RequiredScopes:  []string{"repo", "user:email"},
				},
				{
					ServiceID:       "gitlab-service",
					RequirementType: domainStorage.RequirementTypeOptional,
					RequiredScopes:  []string{"api", "read_user"},
				},
				{
					ServiceID:       "slack-service",
					RequirementType: domainStorage.RequirementTypeMandatory,
					RequiredScopes:  []string{"users:read"},
				},
			},
		}

		err := adapter.Agents().Create(ctx, agent)
		require.NoError(t, err)

		retrieved, err := adapter.Agents().Get(ctx, agent.ID)
		require.NoError(t, err)
		assert.Len(t, retrieved.ServiceRequirements, 3)

		// Verify each requirement
		for i, req := range retrieved.ServiceRequirements {
			assert.Equal(t, agent.ServiceRequirements[i].ServiceID, req.ServiceID)
			assert.Equal(t, agent.ServiceRequirements[i].RequirementType, req.RequirementType)
			assert.Equal(t, agent.ServiceRequirements[i].RequiredScopes, req.RequiredScopes)
		}
	})
}

// TestAgentRepositoryServiceRequirements_PostgreSQL tests service requirements with PostgreSQL adapter
// This requires container setup and will be skipped if no Docker available
// NOTE: PostgreSQL integration tests are tested separately in agent_service_requirements_migration_test.go
func TestAgentRepositoryServiceRequirements_PostgreSQL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Skip("PostgreSQL integration tests covered in agent_service_requirements_migration_test.go")
}
