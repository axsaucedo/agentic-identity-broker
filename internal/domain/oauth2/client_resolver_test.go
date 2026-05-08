package oauth2

import (
	"context"
	"errors"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpaqueClientResolver_URLFormat_Rejected(t *testing.T) {
	resolver := NewOpaqueClientResolver(NewMockAgentRepository())

	for _, clientID := range []string{
		"https://agent.example.com/client",
		"http://agent.example.com/client",
	} {
		_, err := resolver.ResolveClient(context.Background(), id.ClientID(clientID))
		require.Error(t, err, "expected error for %s", clientID)

		var clientErr *ports.ClientIDError
		require.True(t, errors.As(err, &clientErr))
		assert.Equal(t, "invalid_client", clientErr.Code)
	}
}

func TestOpaqueClientResolver_NonUUID_Rejected(t *testing.T) {
	resolver := NewOpaqueClientResolver(NewMockAgentRepository())

	_, err := resolver.ResolveClient(context.Background(), "not-a-uuid")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestOpaqueClientResolver_AgentNotFound(t *testing.T) {
	resolver := NewOpaqueClientResolver(NewMockAgentRepository())

	_, err := resolver.ResolveClient(context.Background(), id.ClientID("00000000-0000-0000-0000-000000000099"))
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestOpaqueClientResolver_Success(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	repo := NewMockAgentRepository()
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.ClientID(agentID.String()),
		DisplayName: "Test Agent",
	}
	require.NoError(t, repo.Create(context.Background(), agent))

	resolver := NewOpaqueClientResolver(repo)
	resolution, err := resolver.ResolveClient(context.Background(), id.ClientID(agentID.String()))

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	assert.Nil(t, resolution.CIMDMetadata)
}
