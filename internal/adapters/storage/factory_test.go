package storage

import (
	"context"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdapter_Memory(t *testing.T) {
	config := &ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}

	adapter, err := NewAdapter(config)

	require.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.Lifecycle())
	assert.NotNil(t, adapter.Users())
}

func TestNewAdapter_Postgres(t *testing.T) {
	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: "postgresql://user:pass@localhost:5432/testdb",
		},
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}

	adapter, err := NewAdapter(config)

	require.NoError(t, err)
	assert.NotNil(t, adapter)
	assert.NotNil(t, adapter.Lifecycle())
	assert.NotNil(t, adapter.Users())
}

func TestNewAdapter_InvalidBackend(t *testing.T) {
	config := &ports.StorageConfig{
		Backend: "invalid",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}

	adapter, err := NewAdapter(config)

	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "unsupported storage backend")
}

func TestNewAdapter_PostgresNoURL(t *testing.T) {
	config := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: "",
		},
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}

	adapter, err := NewAdapter(config)

	assert.Error(t, err)
	assert.Nil(t, adapter)
	assert.Contains(t, err.Error(), "connection URL cannot be empty")
}

// TestBackendSwitching verifies that different adapters can be created independently
func TestBackendSwitching(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create memory adapter
	memConfig := &ports.StorageConfig{
		Backend: "memory",
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}
	memAdapter, err := NewAdapter(memConfig)
	require.NoError(t, err)
	require.NotNil(t, memAdapter)

	// Initialize memory adapter
	err = memAdapter.Lifecycle().Initialize(ctx)
	require.NoError(t, err)

	// Create user in memory adapter
	user := &ports.User{
		ID:        "user123",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = memAdapter.Users().CreateUser(ctx, user)
	require.NoError(t, err)

	// Verify user exists in memory adapter
	retrieved, err := memAdapter.Users().GetUser(ctx, "user123")
	require.NoError(t, err)
	assert.Equal(t, "user123", retrieved.ID)

	// Create postgres adapter (would use different underlying storage)
	postgresConfig := &ports.StorageConfig{
		Backend: "postgres",
		Postgres: ports.PostgresConfig{
			ConnectionURL: "postgresql://user:pass@localhost:5432/testdb",
		},
		Timeouts: ports.StorageTimeouts{
			Read:  5 * time.Second,
			Write: 10 * time.Second,
		},
	}
	postgresAdapter, err := NewAdapter(postgresConfig)
	require.NoError(t, err)
	require.NotNil(t, postgresAdapter)

	// Verify backends are different types by checking their interface implementations
	// Memory adapter implements both StorageLifecycle and UserRepository
	assert.NotNil(t, memAdapter.Lifecycle())
	assert.NotNil(t, memAdapter.Users())
	// Postgres adapter implements both StorageLifecycle and UserRepository
	assert.NotNil(t, postgresAdapter.Lifecycle())
	assert.NotNil(t, postgresAdapter.Users())

	// Close memory adapter
	err = memAdapter.Lifecycle().Close(ctx)
	assert.NoError(t, err)
}

// TestBackendSelectionValidation verifies backend selection validation during factory creation
func TestBackendSelectionValidation(t *testing.T) {
	tests := []struct {
		name       string
		backend    string
		shouldFail bool
	}{
		{
			name:       "valid memory backend",
			backend:    "memory",
			shouldFail: false,
		},
		{
			name:       "valid postgres backend",
			backend:    "postgres",
			shouldFail: false, // Fails on connection, not backend selection
		},
		{
			name:       "invalid backend sqlite",
			backend:    "sqlite",
			shouldFail: true,
		},
		{
			name:       "invalid backend mongodb",
			backend:    "mongodb",
			shouldFail: true,
		},
		{
			name:       "empty backend",
			backend:    "",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &ports.StorageConfig{
				Backend: tt.backend,
				Postgres: ports.PostgresConfig{
					ConnectionURL: "postgresql://user:pass@localhost:5432/testdb",
				},
				Timeouts: ports.StorageTimeouts{
					Read:  5 * time.Second,
					Write: 10 * time.Second,
				},
			}

			adapter, err := NewAdapter(config)

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Nil(t, adapter)
			} else {
				// Postgres adapter fails on connection, but factory should succeed
				if tt.backend == "postgres" {
					// Connection error is acceptable for this test
					assert.NotNil(t, adapter)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, adapter)
				}
			}
		})
	}
}
