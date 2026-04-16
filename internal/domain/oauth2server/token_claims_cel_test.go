package oauth2server

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestNewTokenClaimsEvaluator(t *testing.T) {
	t.Run("empty expression returns nil evaluator", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator("")
		assert.NoError(t, err)
		assert.Nil(t, eval)
	})

	t.Run("valid expression compiles", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"team": "engineering"}`)
		require.NoError(t, err)
		require.NotNil(t, eval)
	})

	t.Run("invalid expression fails at startup", func(t *testing.T) {
		_, err := NewTokenClaimsEvaluator(`invalid syntax !!!`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token_claims_expression")
	})

	t.Run("expression with agent variable compiles", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"client": agent.client_id}`)
		require.NoError(t, err)
		require.NotNil(t, eval)
	})
}

func TestTokenClaimsEvaluator_Evaluate(t *testing.T) {
	t.Run("static claims evaluated correctly", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"team": "engineering", "env": "prod"}`)
		require.NoError(t, err)
		require.NotNil(t, eval)

		req := buildTestRequest("test-client", "user@example.com", []string{"read"})
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, claims)
		assert.Equal(t, "engineering", claims["team"])
		assert.Equal(t, "prod", claims["env"])
	})

	t.Run("nil evaluator produces no claims", func(t *testing.T) {
		var eval *TokenClaimsEvaluator
		req := buildTestRequest("test-client", "user@example.com", []string{"read"})
		claims, err := eval.Evaluate(context.Background(), req)
		assert.NoError(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expression with agent context works", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"agent_name": agent.client_id, "agent_uuid": agent.id}`)
		require.NoError(t, err)

		agentID := id.NewAgentID()
		req := &fosite.Request{
			Client: &brokerClient{
				agent: &storage.Agent{
					ID:       agentID,
					ClientID: id.ClientID("upstream-client-id"),
				},
				credential: &storage.BrokerClientCredential{},
			},
			Session: &fosite.DefaultSession{
				Subject: "user@example.com",
				ExpiresAt: map[fosite.TokenType]time.Time{
					fosite.AccessToken: time.Now().Add(time.Hour),
				},
			},
			GrantedScope: []string{"read"},
		}
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "upstream-client-id", claims["agent_name"])
		assert.Equal(t, agentID.String(), claims["agent_uuid"])
	})

	t.Run("expression with principal.id works", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"user": principal.id}`)
		require.NoError(t, err)

		req := buildTestRequest("my-agent", "admin@example.com", []string{"read"})
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "admin@example.com", claims["user"])
	})

	t.Run("expression with agent.display_name works", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"name": agent.display_name}`)
		require.NoError(t, err)

		req := &fosite.Request{
			Client: &brokerClient{
				agent:      &storage.Agent{ID: id.NewAgentID(), ClientID: "c", DisplayName: "My Agent"},
				credential: &storage.BrokerClientCredential{},
			},
			Session: &fosite.DefaultSession{
				Subject:   "user@example.com",
				ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(time.Hour)},
			},
			GrantedScope: []string{"read"},
		}
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "My Agent", claims["name"])
	})

	t.Run("expression with request.grant_type works", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"grant": request.grant_type}`)
		require.NoError(t, err)

		req := &fosite.Request{
			Client: &brokerClient{
				agent:      &storage.Agent{ID: id.NewAgentID(), ClientID: "c"},
				credential: &storage.BrokerClientCredential{},
			},
			Session: &fosite.DefaultSession{
				Subject:   "user@example.com",
				ExpiresAt: map[fosite.TokenType]time.Time{fosite.AccessToken: time.Now().Add(time.Hour)},
			},
			GrantedScope: []string{"read"},
			Form:         url.Values{"grant_type": {"client_credentials"}},
		}
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "client_credentials", claims["grant"])
	})
}

// buildTestRequest creates a fosite.Requester for testing CEL evaluation.
// clientID is set as Agent.ClientID (upstream OAuth2 client ID); agent.id is a fresh UUID.
func buildTestRequest(clientID, subject string, scopes []string) fosite.Requester {
	session := &fosite.DefaultSession{
		Subject: subject,
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(time.Hour),
		},
	}
	agent := &storage.Agent{
		ID:       id.NewAgentID(),
		ClientID: id.ClientID(clientID),
	}
	return &fosite.Request{
		Client:       &brokerClient{agent: agent, credential: &storage.BrokerClientCredential{}},
		Session:      session,
		GrantedScope: scopes,
	}
}
