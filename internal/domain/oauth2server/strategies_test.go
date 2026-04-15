package oauth2server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestJWXAccessTokenStrategy_GenerateAccessToken(t *testing.T) {
	t.Run("generated JWT header contains the current signing key kid", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		strategy := NewJWXAccessTokenStrategy(svc, repo, "https://issuer.example.com", time.Hour, nil, testSlogger())

		agent := testAgent()
		cred := &dstorage.BrokerClientCredential{
			AgentID:        agent.ID,
			BrokerClientID: "broker_test",
			SecretHash:     "hash",
		}
		requester := &fosite.Request{
			Client:       &brokerClient{agent: agent, credential: cred},
			Session:      &fosite.DefaultSession{Subject: "user@example.com"},
			GrantedScope: fosite.Arguments{"read"},
		}

		tokenStr, sig, err := strategy.GenerateAccessToken(ctx, requester)
		require.NoError(t, err)
		require.NotEmpty(t, tokenStr)
		require.NotEmpty(t, sig)

		// Decode the JWT header (first segment) and assert kid is set correctly.
		// This is the regression guard for the privKey.Set(kid) fix — if the Set
		// call is removed or its error is silently discarded, kid disappears.
		parts := strings.SplitN(tokenStr, ".", 3)
		require.Len(t, parts, 3, "JWT must have three dot-separated parts")

		headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
		require.NoError(t, err)

		var header map[string]any
		require.NoError(t, json.Unmarshal(headerJSON, &header))

		assert.Equal(t, key.KID.String(), header["kid"], "JWT header must contain the current signing key kid")
	})

	t.Run("signature is deterministic SHA-256 of the token", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		_, err := svc.GenerateAndStoreKey(ctx, "ES256", true)
		require.NoError(t, err)

		strategy := NewJWXAccessTokenStrategy(svc, repo, "https://issuer.example.com", time.Hour, nil, testSlogger())

		agent := testAgent()
		cred := &dstorage.BrokerClientCredential{
			AgentID:        agent.ID,
			BrokerClientID: "broker_test",
			SecretHash:     "hash",
		}
		requester := &fosite.Request{
			Client:       &brokerClient{agent: agent, credential: cred},
			Session:      &fosite.DefaultSession{Subject: "user@example.com"},
			GrantedScope: fosite.Arguments{"read"},
		}

		tokenStr, sig, err := strategy.GenerateAccessToken(ctx, requester)
		require.NoError(t, err)

		assert.Equal(t, sha256Hex(tokenStr), sig, "signature must equal SHA-256 of the token string")
	})
}

func TestRandomCodeStrategy_GenerateAuthorizeCode(t *testing.T) {
	strategy := &RandomCodeStrategy{}

	t.Run("generates unique codes", func(t *testing.T) {
		code1, sig1, err := strategy.GenerateAuthorizeCode(context.Background(), nil)
		require.NoError(t, err)
		code2, sig2, err := strategy.GenerateAuthorizeCode(context.Background(), nil)
		require.NoError(t, err)

		assert.NotEqual(t, code1, code2)
		assert.NotEqual(t, sig1, sig2)
		assert.NotEmpty(t, code1)
		assert.NotEmpty(t, sig1)
	})

	t.Run("signature is deterministic for same code", func(t *testing.T) {
		code, sig, err := strategy.GenerateAuthorizeCode(context.Background(), nil)
		require.NoError(t, err)

		computedSig := strategy.AuthorizeCodeSignature(context.Background(), code)
		assert.Equal(t, sig, computedSig)
	})

	t.Run("validate is no-op", func(t *testing.T) {
		err := strategy.ValidateAuthorizeCode(context.Background(), nil, "any-code")
		assert.NoError(t, err)
	})
}

func TestSHA256Hex(t *testing.T) {
	t.Run("produces consistent output", func(t *testing.T) {
		hash1 := sha256Hex("test-input")
		hash2 := sha256Hex("test-input")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("produces different output for different input", func(t *testing.T) {
		hash1 := sha256Hex("input-1")
		hash2 := sha256Hex("input-2")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("output is hex encoded", func(t *testing.T) {
		hash := sha256Hex("test")
		assert.Len(t, hash, 64) // SHA-256 produces 32 bytes = 64 hex chars
	})
}
