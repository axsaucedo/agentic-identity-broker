package cimd

import (
	"bytes"
	"context"
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

// --- hand-rolled mocks ---

type mockFetcher struct {
	result *ports.CIMDFetchResult
	err    error
	calls  int
}

func (m *mockFetcher) Fetch(_ context.Context, _ string) (*ports.CIMDFetchResult, error) {
	m.calls++
	return m.result, m.err
}

type mockAgentRepo struct {
	agents    map[id.AgentID]*storage.Agent
	updateErr error
	updated   []*storage.Agent
}

func newMockAgentRepo(agents ...*storage.Agent) *mockAgentRepo {
	r := &mockAgentRepo{agents: make(map[id.AgentID]*storage.Agent)}
	for _, a := range agents {
		r.agents[a.ID] = a
	}
	return r
}

func (m *mockAgentRepo) Create(_ context.Context, a *storage.Agent) error {
	m.agents[a.ID] = a
	return nil
}
func (m *mockAgentRepo) Get(_ context.Context, id id.AgentID) (*storage.Agent, error) {
	a, ok := m.agents[id]
	if !ok {
		return nil, ports.ErrNotFound
	}
	return a, nil
}
func (m *mockAgentRepo) Update(_ context.Context, a *storage.Agent) error {
	m.updated = append(m.updated, a)
	if m.updateErr != nil {
		return m.updateErr
	}
	m.agents[a.ID] = a
	return nil
}
func (m *mockAgentRepo) Delete(_ context.Context, id id.AgentID) error {
	delete(m.agents, id)
	return nil
}
func (m *mockAgentRepo) List(_ context.Context) ([]*storage.Agent, error) {
	var out []*storage.Agent
	for _, a := range m.agents {
		out = append(out, a)
	}
	return out, nil
}
func (m *mockAgentRepo) GetByClientID(_ context.Context, clientID id.ClientID) (*storage.Agent, error) {
	for _, a := range m.agents {
		if a.ClientID == clientID {
			return a, nil
		}
	}
	return nil, ports.ErrNotFound
}
func (m *mockAgentRepo) GetByClientURI(_ context.Context, _ string) (*storage.Agent, error) {
	return nil, ports.ErrNotFound
}

func (m *mockAgentRepo) ExistsOtherWithClientID(_ context.Context, _ id.ClientID, _ *id.AgentID) (bool, error) {
	return false, nil
}

// --- helpers ---

func cimdFetchResult(t *testing.T, clientID string, authMethod string) *ports.CIMDFetchResult {
	t.Helper()
	doc := fmt.Sprintf(
		`{"client_id":%q,"client_name":"Test Agent","redirect_uris":["https://agent.example.com/cb"],"token_endpoint_auth_method":%q}`,
		clientID, authMethod,
	)
	return &ports.CIMDFetchResult{
		Body:         []byte(doc),
		CacheControl: "max-age=300",
	}
}

func testAgent(id id.AgentID) *storage.Agent {
	return &storage.Agent{
		ID:          id,
		ClientID:    "https://agent.example.com/client",
		DisplayName: "Test Agent",
	}
}

// --- tests ---

func TestService_Resolve_CacheHit(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	cache := mustNewCIMDCache(t, 60*time.Second, time.Hour)
	cached := &ClientIDMetadataDocument{
		ClientID:     "https://agent.example.com/client",
		ClientName:   "Test Agent",
		RedirectURIs: []string{"https://agent.example.com/cb"},
	}
	cache.Set("https://agent.example.com/client", cached, CacheHeaders{}, time.Now())

	fetcher := &mockFetcher{}
	svc := NewService(fetcher, cache, nil, slog.Default())

	doc, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Equal(t, cached, doc)
	assert.Equal(t, 0, fetcher.calls, "should not fetch when cache hit")
}

func TestService_Resolve_FetchError(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	svc := NewService(
		&mockFetcher{err: fmt.Errorf("connection refused")},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		nil,
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestService_Resolve_InvalidDocument(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	// client_id in document doesn't match fetch URL — should fail validation
	badBody := `{"client_id":"https://other.example.com/client","client_name":"Bad Agent","redirect_uris":["https://other.example.com/cb"]}`
	svc := NewService(
		&mockFetcher{result: &ports.CIMDFetchResult{Body: []byte(badBody)}},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		nil,
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.Error(t, err)
}

func TestService_Resolve_BareQueryDelimiterLogsWarning(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	fetcher := &mockFetcher{result: cimdFetchResult(t, "https://agent.example.com/client?", "none")}

	svc := NewService(
		fetcher,
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		nil,
		logger,
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client?", agent)
	require.NoError(t, err)
	assert.Contains(t, logBuf.String(), "client_id URL contains query string")
}

func TestService_Resolve_NameBlocklist_Rejected(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	badName := `{"client_id":"https://agent.example.com/client","client_name":"Blocked","redirect_uris":["https://agent.example.com/cb"]}`
	svc := NewService(
		&mockFetcher{result: &ports.CIMDFetchResult{Body: []byte(badName)}},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		[]string{"blocked"},
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.Error(t, err)
}

func TestService_Resolve_NameBlocklist_PartialMatchNotRejected(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	// "Blocked Agent" contains "blocked" as a substring but is not exactly "blocked"
	doc := `{"client_id":"https://agent.example.com/client","client_name":"Blocked Agent","redirect_uris":["https://agent.example.com/cb"]}`
	svc := NewService(
		&mockFetcher{result: &ports.CIMDFetchResult{Body: []byte(doc)}},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		[]string{"blocked"},
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err, "partial substring match must not be rejected")
}

func TestService_Resolve_InvalidURL(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	svc := NewService(
		&mockFetcher{},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		nil,
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "http://agent.example.com/client", agent)
	require.Error(t, err, "http:// should be rejected")
}

func TestService_Resolve_BuiltinReservedNamesRejectedWithEmptyConfig(t *testing.T) {
	for _, reserved := range []string{"admin", "administrator", "system", "operator", "root", "superuser"} {
		t.Run(reserved, func(t *testing.T) {
			agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
			agent := testAgent(agentID)

			body := fmt.Sprintf(
				`{"client_id":"https://agent.example.com/client","client_name":%q,"redirect_uris":["https://agent.example.com/cb"]}`,
				reserved,
			)
			svc := NewService(
				&mockFetcher{result: &ports.CIMDFetchResult{Body: []byte(body)}},
				mustNewCIMDCache(t, 60*time.Second, time.Hour),
				nil,
				slog.Default(),
			)

			_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
			require.Error(t, err, "built-in reserved name %q must be rejected even with empty operator blocklist", reserved)
		})
	}
}
