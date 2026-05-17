package oauth2

import (
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/stretchr/testify/assert"
)

func TestModeStrategy_Proxy(t *testing.T) {
	s := NewProxyModeStrategy()
	assert.Equal(t, servermode.Proxy, s.Mode())
	assert.True(t, s.AcceptsClientMode(storage.ProxyClient))
	assert.False(t, s.AcceptsClientMode(storage.CIMDClient))
	assert.False(t, s.AcceptsClientMode(storage.LocalClient))
}

func TestModeStrategy_Local(t *testing.T) {
	s := NewLocalModeStrategy()
	assert.Equal(t, servermode.Local, s.Mode())
	assert.False(t, s.AcceptsClientMode(storage.ProxyClient))
	assert.True(t, s.AcceptsClientMode(storage.CIMDClient))
	assert.True(t, s.AcceptsClientMode(storage.LocalClient))
}

func TestModeStrategy_Hybrid(t *testing.T) {
	s := NewHybridModeStrategy()
	assert.Equal(t, servermode.Hybrid, s.Mode())
	assert.True(t, s.AcceptsClientMode(storage.ProxyClient))
	assert.True(t, s.AcceptsClientMode(storage.CIMDClient))
	assert.True(t, s.AcceptsClientMode(storage.LocalClient))
}

func TestAgent_ClientMode(t *testing.T) {
	clientID := id.ClientID("upstream-client-id")

	cases := []struct {
		name  string
		agent storage.Agent
		want  storage.ClientMode
	}{
		{
			name:  "agent with ClientID is ProxyClient",
			agent: storage.Agent{ClientID: &clientID},
			want:  storage.ProxyClient,
		},
		{
			name:  "agent with ClientURIs and no ClientID is CIMDClient",
			agent: storage.Agent{ClientURIs: []string{"https://example.com/.well-known/openid-configuration"}},
			want:  storage.CIMDClient,
		},
		{
			name:  "agent with neither is LocalClient",
			agent: storage.Agent{},
			want:  storage.LocalClient,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.agent.ClientMode())
		})
	}
}
