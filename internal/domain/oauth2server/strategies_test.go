package oauth2server

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestJWXAccessTokenStrategy_SubClaimNotOverridable(t *testing.T) {
	svc, repo := newTestSigningKeyService()
	_, err := svc.GenerateAndStoreKey(context.Background(), "ES256", true)
	require.NoError(t, err)

	eval, err := NewTokenClaimsEvaluator(`{"sub": "attacker@evil.com", "extra": "ok"}`)
	require.NoError(t, err)

	strategy := NewJWXAccessTokenStrategy(svc, repo, "https://broker.example.com", time.Hour, eval, testSlogger())
	req := buildTestRequest("my-agent", "legitimate@example.com", []string{"read"})

	tokenStr, _, err := strategy.GenerateAccessToken(context.Background(), req)
	require.NoError(t, err)

	tok, err := jwt.ParseInsecure([]byte(tokenStr))
	require.NoError(t, err)
	sub, ok := tok.Subject()
	require.True(t, ok)
	assert.Equal(t, "legitimate@example.com", sub)

	var extra string
	require.NoError(t, tok.Get("extra", &extra), "non-reserved CEL claim should be present")
	assert.Equal(t, "ok", extra)
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
