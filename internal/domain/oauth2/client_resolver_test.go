package oauth2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCIMDFetcher is a test double for ports.CIMDFetcher.
type mockCIMDFetcher struct {
	result *ports.CIMDFetchResult
	err    error
}

func (m *mockCIMDFetcher) Fetch(_ context.Context, _ string) (*ports.CIMDFetchResult, error) {
	return m.result, m.err
}

// cimdServiceForTest creates a cimd.Service backed by the given fetch result.
func cimdServiceForTest(fetchResult *ports.CIMDFetchResult, fetchErr error) *cimd.Service {
	fetcher := &mockCIMDFetcher{result: fetchResult, err: fetchErr}
	cache, _ := cimd.NewCIMDCache(60*time.Second, time.Hour, 1000)
	return cimd.NewService(fetcher, cache, nil, slog.Default())
}

// --- CIMD enabled tests ---

func TestAgentClientResolver_OpaqueUUID_Success(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID(agentID.String())),
		DisplayName: "Test Agent",
	}
	repo := NewMockAgentRepository()
	require.NoError(t, repo.Create(context.Background(), agent))
	svc := cimdServiceForTest(nil, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	resolution, err := resolver.ResolveClient(context.Background(), id.ClientID(agentID.String()))
	require.NoError(t, err)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	assert.Nil(t, resolution.CIMDMetadata)
}

func TestAgentClientResolver_OpaqueUUID_NotFound(t *testing.T) {
	repo := NewMockAgentRepository()
	svc := cimdServiceForTest(nil, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	_, err := resolver.ResolveClient(context.Background(), id.ClientID("00000000-0000-0000-0000-000000000099"))
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestAgentClientResolver_InvalidURL_Rejected(t *testing.T) {
	repo := NewMockAgentRepository()
	svc := cimdServiceForTest(nil, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	_, err := resolver.ResolveClient(context.Background(), "http://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestAgentClientResolver_URLNotRegistered(t *testing.T) {
	repo := NewMockAgentRepository()
	svc := cimdServiceForTest(nil, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	_, err := resolver.ResolveClient(context.Background(), "https://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
	assert.Equal(t, "Client not registered", clientErr.Desc)
}

func TestAgentClientResolver_CIMDFetchFails(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID("https://agent.example.com/client")),
		DisplayName: "Test Agent",
	}
	repo := NewMockAgentRepository()
	require.NoError(t, repo.Create(context.Background(), agent))
	repo.RegisterURI("https://agent.example.com/client", agent)
	svc := cimdServiceForTest(nil, fmt.Errorf("connection refused"))
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	_, err := resolver.ResolveClient(context.Background(), "https://agent.example.com/client")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestAgentClientResolver_URLFormat_Success(t *testing.T) {
	const clientURL = "https://agent.example.com/client"
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID(clientURL)),
		DisplayName: "Test Agent",
	}
	repo := NewMockAgentRepository()
	require.NoError(t, repo.Create(context.Background(), agent))
	repo.RegisterURI(clientURL, agent)
	body := fmt.Sprintf(
		`{"client_id":%q,"client_name":"Test Agent","redirect_uris":["https://agent.example.com/cb"]}`,
		clientURL,
	)
	fetchResult := &ports.CIMDFetchResult{Body: []byte(body), CacheControl: "max-age=300"}
	svc := cimdServiceForTest(fetchResult, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	resolution, err := resolver.ResolveClient(context.Background(), clientURL)
	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	require.NotNil(t, resolution.CIMDMetadata)
	assert.Equal(t, clientURL, resolution.CIMDMetadata.ClientID)
	assert.Equal(t, []string{"https://agent.example.com/cb"}, resolution.CIMDMetadata.RedirectURIs)
}

// TestAgentClientResolver_ResolvesViaCIMDURINotClientID proves the resolver uses
// GetByClientURI (pre-registered URI lookup) rather than GetByClientID. When the
// agent's ClientID is a UUID and the CIMD URL is only in ClientURIs, the resolver
// must still find the agent.
func TestAgentClientResolver_ResolvesViaCIMDURINotClientID(t *testing.T) {
	const cimdURI = "https://agent.example.com/cimd-endpoint"
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000099")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID(agentID.String())), // UUID-form ClientID, not the CIMD URL
		DisplayName: "URI-Only Agent",
	}
	repo := NewMockAgentRepository()
	require.NoError(t, repo.Create(context.Background(), agent))
	repo.RegisterURI(cimdURI, agent) // registered as ClientURI, not ClientID

	body := fmt.Sprintf(`{"client_id":%q,"client_name":"URI-Only Agent","redirect_uris":["https://agent.example.com/cb"]}`, cimdURI)
	fetchResult := &ports.CIMDFetchResult{Body: []byte(body), CacheControl: "max-age=300"}
	svc := cimdServiceForTest(fetchResult, nil)
	resolver := NewAgentClientResolverWithCIMD(repo, svc, slog.Default())

	resolution, err := resolver.ResolveClient(context.Background(), cimdURI)
	require.NoError(t, err, "should resolve via ClientURI even when ClientID is a different value")
	assert.Equal(t, agentID, resolution.Agent.ID)
}

// --- CIMD disabled tests ---

func TestAgentClientResolver_Disabled_URLFormat_Rejected(t *testing.T) {
	repo := NewMockAgentRepository()
	resolver := NewAgentClientResolver(repo, nil)

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

func TestAgentClientResolver_Disabled_NonUUID_Rejected(t *testing.T) {
	repo := NewMockAgentRepository()
	resolver := NewAgentClientResolver(repo, nil)

	_, err := resolver.ResolveClient(context.Background(), "not-a-uuid")
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}

func TestAgentClientResolver_Disabled_OpaqueUUID_Success(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    ptr.To(id.ClientID(agentID.String())),
		DisplayName: "Test Agent",
	}
	repo := NewMockAgentRepository()
	require.NoError(t, repo.Create(context.Background(), agent))
	resolver := NewAgentClientResolver(repo, nil)

	resolution, err := resolver.ResolveClient(context.Background(), id.ClientID(agentID.String()))
	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, agent.ID, resolution.Agent.ID)
	assert.Nil(t, resolution.CIMDMetadata)
}

func TestAgentClientResolver_Disabled_AgentNotFound(t *testing.T) {
	repo := NewMockAgentRepository()
	resolver := NewAgentClientResolver(repo, nil)

	_, err := resolver.ResolveClient(context.Background(), id.ClientID("00000000-0000-0000-0000-000000000099"))
	require.Error(t, err)

	var clientErr *ports.ClientIDError
	require.True(t, errors.As(err, &clientErr))
	assert.Equal(t, "invalid_client", clientErr.Code)
}
