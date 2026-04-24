package cimd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockAgentRepoForCR(agents ...*storage.Agent) *mockAgentRepo {
	return newMockAgentRepo(agents...)
}

// cimdServiceForTest creates a CIMDService backed by the given fetch result.
func cimdServiceForTest(fetchResult *ports.CIMDFetchResult, fetchErr error, repo ports.AgentRepository) *Service {
	fetcher := &mockFetcher{result: fetchResult, err: fetchErr}
	cache := NewCIMDCache(60*time.Second, time.Hour)
	return NewService(fetcher, cache, repo, nil, slog.Default())
}

// --- tests ---

func TestCIMDClientResolver_OpaqueUUID_Success(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    id.ClientID(agentID.String()),
		DisplayName: "Test Agent",
	}
	repo := newMockAgentRepoForCR(agent)
	svc := cimdServiceForTest(nil, nil, repo)
	resolver := NewCIMDClientResolver(repo, svc)

	resolution, err := resolver.ResolveClient(context.Background(), id.ClientID(agentID.String()))
	require.NoError(t, err)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	assert.Nil(t, resolution.CIMDMetadata)
}

func TestCIMDClientResolver_OpaqueUUID_NotFound(t *testing.T) {
	repo := newMockAgentRepoForCR()
	svc := cimdServiceForTest(nil, nil, repo)
	resolver := NewCIMDClientResolver(repo, svc)

	_, err := resolver.ResolveClient(context.Background(), id.ClientID("00000000-0000-0000-0000-000000000099"))
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestCIMDClientResolver_InvalidURL_Rejected(t *testing.T) {
	repo := newMockAgentRepoForCR()
	svc := cimdServiceForTest(nil, nil, repo)
	resolver := NewCIMDClientResolver(repo, svc)

	// http:// is not a valid CIMD URL
	_, err := resolver.ResolveClient(context.Background(), "http://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_request", clientErr.Code)
}

func TestCIMDClientResolver_URLNotRegistered(t *testing.T) {
	repo := newMockAgentRepoForCR()
	svc := cimdServiceForTest(nil, nil, repo)
	resolver := NewCIMDClientResolver(repo, svc)

	_, err := resolver.ResolveClient(context.Background(), "https://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
	assert.Equal(t, "Client not registered", clientErr.Desc)
}

func TestCIMDClientResolver_CIMDFetchFails(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "https://agent.example.com/client",
		DisplayName: "Test Agent",
	}
	repo := newMockAgentRepoForCR(agent)
	svc := cimdServiceForTest(nil, fmt.Errorf("connection refused"), repo)
	resolver := NewCIMDClientResolver(repo, svc)

	_, err := resolver.ResolveClient(context.Background(), "https://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestCIMDClientResolver_URLFormat_Success(t *testing.T) {
	const clientURL = "https://agent.example.com/client"
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    clientURL,
		DisplayName: "Test Agent",
	}
	repo := newMockAgentRepoForCR(agent)
	body := fmt.Sprintf(
		`{"client_id":%q,"client_name":"Test Agent","redirect_uris":["https://agent.example.com/cb"]}`,
		clientURL,
	)
	fetchResult := &ports.CIMDFetchResult{Body: []byte(body), CacheControl: "max-age=300"}
	svc := cimdServiceForTest(fetchResult, nil, repo)
	resolver := NewCIMDClientResolver(repo, svc)

	resolution, err := resolver.ResolveClient(context.Background(), clientURL)
	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	require.NotNil(t, resolution.CIMDMetadata)
	assert.Equal(t, clientURL, resolution.CIMDMetadata.ClientID)
	assert.Equal(t, []string{"https://agent.example.com/cb"}, resolution.CIMDMetadata.RedirectURIs)
}
