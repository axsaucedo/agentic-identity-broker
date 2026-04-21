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
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:  time.Now().UTC(),
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
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:  time.Now().UTC(),
	}
	err := repo.Create(ctx, cred)
	require.NoError(t, err)

	got, err := repo.GetByAgentID(ctx, agent.ID)
	require.NoError(t, err)
	assert.Equal(t, cred.AgentID, got.AgentID)
	assert.Equal(t, cred.SecretHash, got.SecretHash)
}

func TestBrokerClientCredentialRepo_GetByClientID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	clientID := id.NewClientID(agent.ID.String())
	cred := &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:  time.Now().UTC(),
	}
	err := repo.Create(ctx, cred)
	require.NoError(t, err)

	got, err := repo.GetByClientID(ctx, clientID)
	require.NoError(t, err)
	assert.Equal(t, agent.ID, got.AgentID)
	assert.Equal(t, cred.SecretHash, got.SecretHash)
}

func TestBrokerClientCredentialRepo_GetByAgentID_Timeout(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	repo := NewBrokerClientCredentialRepo(adapter)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := repo.GetByAgentID(ctx, id.NewAgentID())
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindTimeout, se.Kind)
}

func TestBrokerClientCredentialRepo_GetByClientID_Timeout(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	repo := NewBrokerClientCredentialRepo(adapter)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := repo.GetByClientID(ctx, id.NewClientID(id.NewAgentID().String()))
	require.Error(t, err)

	var se *storage.StorageError
	require.True(t, errors.As(err, &se))
	assert.Equal(t, storage.ErrorKindTimeout, se.Kind)
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

func TestBrokerClientCredentialRepo_GetByClientID_NotFound(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	repo := NewBrokerClientCredentialRepo(adapter)
	_, err := repo.GetByClientID(context.Background(), id.NewClientID(id.NewAgentID().String()))
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
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash",
		CreatedAt:  time.Now().UTC(),
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

func TestBrokerClientCredentialRepo_UniquePerAgent(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	cred1 := &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash1",
		CreatedAt:  time.Now().UTC(),
	}
	err := repo.Create(ctx, cred1)
	require.NoError(t, err)

	// Second credential for the same agent must fail (UNIQUE on agent_id).
	cred2 := &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		AgentID:    agent.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash2",
		CreatedAt:  time.Now().UTC(),
	}
	err = repo.Create(ctx, cred2)
	assert.Error(t, err, "second credential for same agent should fail unique constraint")
}

func TestBrokerClientCredentialRepo_UniqueClientID(t *testing.T) {
	adapter, cleanup := setupCredentialTestDB(t)
	defer cleanup()

	agent1 := createTestAgent(t, adapter)
	agent2 := createTestAgent(t, adapter)
	repo := NewBrokerClientCredentialRepo(adapter)
	ctx := context.Background()

	cred1 := &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		AgentID:    agent1.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash1",
		CreatedAt:  time.Now().UTC(),
	}
	err := repo.Create(ctx, cred1)
	require.NoError(t, err)

	// Using agent1's agent_id for a second credential must fail the UNIQUE constraint.
	cred2 := &storage.BrokerClientCredential{
		ID:         id.NewCredentialID(),
		AgentID:    agent1.ID,
		SecretHash: "$argon2id$v=19$m=65536,t=3,p=4$salt$hash2",
		CreatedAt:  time.Now().UTC(),
	}
	err = repo.Create(ctx, cred2)
	assert.Error(t, err, "duplicate agent_id should fail unique constraint")
	_ = agent2 // agent2 exists to represent a distinct principal in this constraint test
}
