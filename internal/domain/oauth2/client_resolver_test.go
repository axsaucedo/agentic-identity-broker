package oauth2

import (
	"context"
	"errors"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
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
		ClientID:    ptr.To(id.ClientID(agentID.String())),
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

// T043: Universal client resolver format detection — opaque (UUID) path.
// UUID format → lookup by Agent.ID.
func TestOpaqueClientResolver_UUID_ResolvesById(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000002")
	repo := NewMockAgentRepository()
	agent := &storage.Agent{
		ID:          agentID,
		DisplayName: "Local Agent",
		Description: "Plain local agent resolved by UUID",
	}
	require.NoError(t, repo.Create(context.Background(), agent))

	resolver := NewOpaqueClientResolver(repo)
	resolution, err := resolver.ResolveClient(context.Background(), id.ClientID(agentID.String()))

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, agentID, resolution.Agent.ID)
	assert.Equal(t, storage.LocalClient, resolution.Agent.ClientMode())
}

// T044: CIMD agent UUID rejection (FR-005) via OpaqueClientResolver.
// OpaqueClientResolver rejects URL-format client IDs before any lookup.
func TestOpaqueClientResolver_CIMDAgent_URLRejectedBefore_Lookup(t *testing.T) {
	repo := NewMockAgentRepository()
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000003")
	// Store a CIMD agent (with ClientURIs, no ClientID)
	agent := &storage.Agent{
		ID:          agentID,
		ClientURIs:  []string{"https://agent.example.com/.well-known/openid-configuration"},
		DisplayName: "CIMD Agent",
		Description: "Agent identified by URL",
	}
	require.NoError(t, repo.Create(context.Background(), agent))

	// Attempting to access the CIMD agent via its UUID must be rejected by OpaqueClientResolver
	// because URL-format client_ids are rejected before lookup (CIMD gate).
	resolver := NewOpaqueClientResolver(repo)
	_, err := resolver.ResolveClient(context.Background(), id.ClientID("https://agent.example.com/.well-known/openid-configuration"))

	require.Error(t, err)
	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
	assert.Contains(t, clientErr.Desc, "CIMD")
}

// T045: Mode enforcement — proxy mode rejects local/CIMD agents.
func TestModeStrategy_ProxyMode_RejectsLocalAndCIMD(t *testing.T) {
	s := NewProxyModeStrategy()
	assert.False(t, s.AcceptsClientMode(storage.LocalClient), "proxy mode must reject LocalClient")
	assert.False(t, s.AcceptsClientMode(storage.CIMDClient), "proxy mode must reject CIMDClient")
}

// T045: Mode enforcement — local mode rejects proxy agents.
func TestModeStrategy_LocalMode_RejectsProxy(t *testing.T) {
	s := NewLocalModeStrategy()
	assert.False(t, s.AcceptsClientMode(storage.ProxyClient), "local mode must reject ProxyClient")
}
