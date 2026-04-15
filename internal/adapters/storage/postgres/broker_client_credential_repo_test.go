//go:build integration
// +build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCredentialTestDB(t *testing.T) (*Adapter, func()) {
	t.Helper()
	container, connString, cleanup := setupTestContainer(t)
	t.Cleanup(cleanup)
	applyMigrations(t, container)
	config := testStorageConfig(connString)
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

func createTestAgent(t *testing.T, adapter *Adapter) *storage.Agent {
	t.Helper()
	repo := NewAgentRepository(adapter)
	now := time.Now().UTC()
	agent := &storage.Agent{
		ClientID:    id.ClientID("cred-test-" + id.NewAgentID().String()[:8]),
		DisplayName: "Credential Test Agent",
		Description: "Agent for credential testing",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := repo.Create(context.Background(), agent)
	require.NoError(t, err)
	return agent
}

func TestBrokerClientCredentialRepo_Create(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)

	cred := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: id.NewBrokerClientID("broker_test123456789012"),
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:      time.Now().UTC(),
	}

	err := repo.Create(context.Background(), cred)
	require.NoError(t, err)
}

func TestBrokerClientCredentialRepo_GetByAgentID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	cred := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: id.NewBrokerClientID("broker_get123456789012"),
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:      time.Now().UTC(),
	}
	err := repo.Create(ctx, cred)
	require.NoError(t, err)

	got, err := repo.GetByAgentID(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, cred.BrokerClientID, got.BrokerClientID)
	assert.Equal(t, cred.SecretHash, got.SecretHash)
}

func TestBrokerClientCredentialRepo_GetByBrokerClientID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	clientID := id.NewBrokerClientID("broker_bcid12345678901")
	cred := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: clientID,
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:      time.Now().UTC(),
	}
	err := repo.Create(ctx, cred)
	require.NoError(t, err)

	got, err := repo.GetByBrokerClientID(ctx, clientID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, got.AgentID)
}

func TestBrokerClientCredentialRepo_GetByAgentID_NotFound(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	repo := NewBrokerClientCredentialRepo(adapter)
	_, err := repo.GetByAgentID(context.Background(), id.NewAgentID())
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindNotFound, se.Kind)
}

func TestBrokerClientCredentialRepo_GetByBrokerClientID_NotFound(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	repo := NewBrokerClientCredentialRepo(adapter)
	_, err := repo.GetByBrokerClientID(context.Background(), id.NewBrokerClientID("broker_nonexistent12345"))
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindNotFound, se.Kind)
}

func TestBrokerClientCredentialRepo_Delete(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	cred := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: id.NewBrokerClientID("broker_del123456789012"),
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:      time.Now().UTC(),
	}
	err := repo.Create(ctx, cred)
	require.NoError(t, err)

	err = repo.Delete(ctx, agent.ID)
	require.NoError(t, err)

	_, err = repo.GetByAgentID(ctx, agent.ID)
	require.Error(t, err)
	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindNotFound, se.Kind)
}

func TestBrokerClientCredentialRepo_UniqueAgentID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	cred1 := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: id.NewBrokerClientID("broker_dup123456789012"),
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash1",
		CreatedAt:      time.Now().UTC(),
	}
	err := repo.Create(ctx, cred1)
	require.NoError(t, err)

	cred2 := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent.ID,
		BrokerClientID: id.NewBrokerClientID("broker_dup223456789012"),
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash2",
		CreatedAt:      time.Now().UTC(),
	}
	err = repo.Create(ctx, cred2)
	assert.Error(t, err, "second credential for same agent should fail unique constraint")
}

func TestBrokerClientCredentialRepo_UniqueBrokerClientID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent1 := createTestAgent(t, adapter)
	agent2 := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	clientID := id.NewBrokerClientID("broker_shared1234567890")
	cred1 := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent1.ID,
		BrokerClientID: clientID,
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash1",
		CreatedAt:      time.Now().UTC(),
	}
	err := repo.Create(ctx, cred1)
	require.NoError(t, err)

	cred2 := &storage.BrokerClientCredential{
		ID:             id.NewCredentialID(),
		AgentID:        agent2.ID,
		BrokerClientID: clientID,
		SecretHash:     "$argon2id$v=19$m=65536,t=3,p=4$salt$hash2",
		CreatedAt:      time.Now().UTC(),
	}
	err = repo.Create(ctx, cred2)
	assert.Error(t, err, "duplicate broker_client_id should fail unique constraint")
}
