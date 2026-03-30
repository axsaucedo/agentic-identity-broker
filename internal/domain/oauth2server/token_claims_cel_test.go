package oauth2server

import (
	"context"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		eval, err := NewTokenClaimsEvaluator(`{"agent_name": agent.client_id}`)
		require.NoError(t, err)

		req := buildTestRequest("my-agent-id", "user@example.com", []string{"read"})
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "my-agent-id", claims["agent_name"])
	})

	t.Run("expression with principal variable works", func(t *testing.T) {
		eval, err := NewTokenClaimsEvaluator(`{"user": principal}`)
		require.NoError(t, err)

		req := buildTestRequest("my-agent", "admin@example.com", []string{"read"})
		claims, err := eval.Evaluate(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "admin@example.com", claims["user"])
	})
}

// buildTestRequest creates a fosite.Requester for testing CEL evaluation.
func buildTestRequest(clientID, subject string, scopes []string) fosite.Requester {
	session := &fosite.DefaultSession{
		Subject: subject,
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(time.Hour),
		},
	}
	return &fosite.Request{
		Client: &fosite.DefaultClient{
			ID: clientID,
		},
		Session:      session,
		GrantedScope: scopes,
	}
}
