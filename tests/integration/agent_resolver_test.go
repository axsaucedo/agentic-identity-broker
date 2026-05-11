package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	memrepo "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resolveAgentIDByClientID mirrors the closure registered in builder.go when
// multi_agent_client is disabled. Kept here to test the guard logic directly.
func buildResolver(repo *memrepo.AgentRepository) func(string) (string, error) {
	return func(clientID string) (string, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		agent, err := repo.GetByClientID(ctx, id.ClientID(clientID))
		if err != nil {
			return "", fmt.Errorf("resolveAgentIdByClientId: %w", err)
		}
		dup, dupErr := repo.ExistsOtherWithClientID(ctx, id.ClientID(clientID), &agent.ID)
		if dupErr != nil {
			return "", fmt.Errorf("resolveAgentIdByClientId: duplicate check failed: %w", dupErr)
		}
		if dup {
			return "", fmt.Errorf("resolveAgentIdByClientId: ambiguous client_id %q matches multiple agents; deduplicate before disabling multi_agent_client", clientID)
		}
		return agent.ID.String(), nil
	}
}

// TestResolveAgentIDByClientID_AmbiguousClientID verifies that the single-agent
// resolver fails loudly when two agents share the same client_id (a state reachable
// when multi_agent_client was previously enabled).
func TestResolveAgentIDByClientID_AmbiguousClientID(t *testing.T) {
	repo := memrepo.NewAgentRepository()
	ctx := context.Background()

	shared := id.ClientID("shared-client-id")
	agent1 := &storage.Agent{ID: id.NewAgentID(), ClientID: shared, DisplayName: "A1", Description: "d"}
	agent2 := &storage.Agent{ID: id.NewAgentID(), ClientID: shared, DisplayName: "A2", Description: "d"}
	require.NoError(t, repo.Create(ctx, agent1))
	require.NoError(t, repo.Create(ctx, agent2))

	resolve := buildResolver(repo)

	_, err := resolve(string(shared))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ambiguous")
}

// TestResolveAgentIDByClientID_UnambiguousClientID verifies the happy path:
// a unique client_id resolves to the registered agent's UUID.
func TestResolveAgentIDByClientID_UnambiguousClientID(t *testing.T) {
	repo := memrepo.NewAgentRepository()
	ctx := context.Background()

	agentID := id.NewAgentID()
	agent := &storage.Agent{ID: agentID, ClientID: "unique-client", DisplayName: "A", Description: "d"}
	require.NoError(t, repo.Create(ctx, agent))

	resolve := buildResolver(repo)

	got, err := resolve("unique-client")
	require.NoError(t, err)
	assert.Equal(t, agentID.String(), got)
}
