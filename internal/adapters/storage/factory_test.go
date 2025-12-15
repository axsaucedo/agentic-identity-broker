package storage

import (
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
