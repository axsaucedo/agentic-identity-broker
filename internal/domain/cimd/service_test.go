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

// --- helpers ---

func cimdFetchResult(t *testing.T, clientID string, authMethod, jwksURI string) *ports.CIMDFetchResult {
	t.Helper()
	doc := fmt.Sprintf(
		`{"client_id":%q,"client_name":"Test Agent","redirect_uris":["https://agent.example.com/cb"],"token_endpoint_auth_method":%q,"jwks_uri":%q}`,
		clientID, authMethod, jwksURI,
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
	cache.Set("https://agent.example.com/client", cached, nil, time.Now())

	fetcher := &mockFetcher{}
	svc := NewService(fetcher, cache, newMockAgentRepo(agent), nil, slog.Default())

	doc, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Equal(t, cached, doc)
	assert.Equal(t, 0, fetcher.calls, "should not fetch when cache hit")
}

func TestService_Resolve_FirstFetch_PopulatesSnapshotSilently(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)
	// No AuthMethod or JwksURI yet — first fetch
	require.Nil(t, agent.AuthMethod)
	require.Nil(t, agent.JwksURI)

	fetchResult := cimdFetchResult(t, "https://agent.example.com/client", "private_key_jwt", "https://agent.example.com/.well-known/jwks.json")
	repo := newMockAgentRepo(agent)
	svc := NewService(
		&mockFetcher{result: fetchResult},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		repo,
		nil,
		slog.Default(),
	)

	doc, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Equal(t, "https://agent.example.com/client", doc.ClientID)

	// Snapshot should be updated on agent
	require.Len(t, repo.updated, 1)
	updated := repo.updated[0]
	require.NotNil(t, updated.AuthMethod)
	assert.Equal(t, "private_key_jwt", *updated.AuthMethod)
	require.NotNil(t, updated.JwksURI)
	assert.Equal(t, "https://agent.example.com/.well-known/jwks.json", *updated.JwksURI)
}

func TestService_Resolve_SubsequentFetch_ChangedAuthMethod_EmitsAuditAndUpdates(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	prev := "private_key_jwt"
	agent := &storage.Agent{
		ID:          agentID,
		ClientID:    "https://agent.example.com/client",
		DisplayName: "Test Agent",
		AuthMethod:  &prev, // already has a snapshot
	}

	fetchResult := cimdFetchResult(t, "https://agent.example.com/client", "none", "")
	repo := newMockAgentRepo(agent)
	svc := NewService(
		&mockFetcher{result: fetchResult},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		repo,
		nil,
		slog.Default(),
	)

	doc, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Equal(t, "https://agent.example.com/client", doc.ClientID)

	require.Len(t, repo.updated, 1)
	assert.Equal(t, "none", *repo.updated[0].AuthMethod)
}

func TestService_Resolve_SubsequentFetch_UnchangedFields_NoUpdate(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	method := "private_key_jwt"
	jwks := "https://agent.example.com/.well-known/jwks.json"
	clientName := "Test Agent"
	agent := &storage.Agent{
		ID:               agentID,
		ClientID:         "https://agent.example.com/client",
		DisplayName:      "Test Agent",
		AuthMethod:       &method,
		JwksURI:          &jwks,
		CIMDClientName:   &clientName,
		CIMDRedirectURIs: []string{"https://agent.example.com/cb"},
	}

	fetchResult := cimdFetchResult(t, "https://agent.example.com/client", "private_key_jwt", "https://agent.example.com/.well-known/jwks.json")
	repo := newMockAgentRepo(agent)
	svc := NewService(
		&mockFetcher{result: fetchResult},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		repo,
		nil,
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Empty(t, repo.updated, "should not update agent when fields are unchanged")
}

func TestService_Resolve_FetchError(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	svc := NewService(
		&mockFetcher{err: fmt.Errorf("connection refused")},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		newMockAgentRepo(agent),
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
		newMockAgentRepo(agent),
		nil,
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.Error(t, err)
}

func TestService_Resolve_UpdateFailureFails(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	fetchResult := cimdFetchResult(t, "https://agent.example.com/client", "private_key_jwt", "")
	repo := newMockAgentRepo(agent)
	repo.updateErr = fmt.Errorf("database unavailable")

	svc := NewService(
		&mockFetcher{result: fetchResult},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		repo,
		nil,
		slog.Default(),
	)

	// Snapshot update failure must propagate as an error to prevent stale baselines
	// from causing repeated false-positive cimd_security_field_changed events.
	// The error must be wrapped in SnapshotPersistenceError so callers can return server_error.
	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database unavailable")
	var snapErr *SnapshotPersistenceError
	require.ErrorAs(t, err, &snapErr)
	assert.Nil(t, agent.AuthMethod, "caller agent must remain unchanged when snapshot persistence fails")
	assert.Nil(t, agent.JwksURI, "caller agent must remain unchanged when snapshot persistence fails")
}

func TestService_Resolve_BareQueryDelimiterLogsWarning(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, nil))
	fetcher := &mockFetcher{result: cimdFetchResult(t, "https://agent.example.com/client?", "private_key_jwt", "")}

	svc := NewService(
		fetcher,
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		newMockAgentRepo(agent),
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
		newMockAgentRepo(agent),
		[]string{"blocked"}, // name blocklist
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
		newMockAgentRepo(agent),
		[]string{"blocked"},
		slog.Default(),
	)

	_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err, "partial substring match must not be rejected")
}

func TestService_Resolve_OmittedAuthMethodDefaultsToNone(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)
	require.Nil(t, agent.AuthMethod)

	// Document omits token_endpoint_auth_method entirely
	noMethod := `{"client_id":"https://agent.example.com/client","client_name":"Test Agent","redirect_uris":["https://agent.example.com/cb"]}`
	repo := newMockAgentRepo(agent)
	svc := NewService(
		&mockFetcher{result: &ports.CIMDFetchResult{Body: []byte(noMethod)}},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		repo,
		nil,
		slog.Default(),
	)

	doc, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
	require.NoError(t, err)
	assert.Equal(t, "none", doc.AuthMethod, "omitted auth method should default to none")

	require.Len(t, repo.updated, 1)
	require.NotNil(t, repo.updated[0].AuthMethod)
	assert.Equal(t, "none", *repo.updated[0].AuthMethod, "snapshot should record default none")
}

func TestService_Resolve_InvalidURL(t *testing.T) {
	agentID := id.MustParseAgentID("00000000-0000-0000-0000-000000000001")
	agent := testAgent(agentID)

	svc := NewService(
		&mockFetcher{},
		mustNewCIMDCache(t, 60*time.Second, time.Hour),
		newMockAgentRepo(agent),
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
				newMockAgentRepo(agent),
				nil, // empty operator list — built-ins must still apply
				slog.Default(),
			)

			_, err := svc.Resolve(context.Background(), "https://agent.example.com/client", agent)
			require.Error(t, err, "built-in reserved name %q must be rejected even with empty operator blocklist", reserved)
		})
	}
}
