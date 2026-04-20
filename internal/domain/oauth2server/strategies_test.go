package oauth2server

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jws"
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

func TestJWXAccessTokenStrategy_GenerateAccessToken(t *testing.T) {
	t.Run("JWT header contains the current signing key kid", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		ctx := context.Background()

		key, err := svc.generateAndStore(ctx, "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, repo, "https://issuer.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)
		tokenStr, sig, err := strategy.GenerateAccessToken(ctx, buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)
		require.NotEmpty(t, sig)

		// Regression guard for privKey.Set(kid): kid must survive into the JWS header.
		msg, err := jws.Parse([]byte(tokenStr))
		require.NoError(t, err)
		require.Len(t, msg.Signatures(), 1)
		kid, ok := msg.Signatures()[0].ProtectedHeaders().KeyID()
		require.True(t, ok, "JWS protected header must carry a kid")
		assert.Equal(t, key.KID.String(), kid)
	})

	t.Run("signature equals SHA-256 of the token string", func(t *testing.T) {
		svc, repo := newTestSigningKeyService()
		_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, repo, "https://issuer.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)
		tokenStr, sig, err := strategy.GenerateAccessToken(context.Background(), buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)

		assert.Equal(t, sha256Hex(tokenStr), sig)
	})

	t.Run("signature verifies against JWKS and all required claims are present", func(t *testing.T) {
		const issuer = "https://issuer.example.com"
		const agentID = "my-agent"
		const subject = "user@example.com"

		svc, repo := newTestSigningKeyService()
		ctx := context.Background()
		_, err := svc.generateAndStore(ctx, "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, repo, issuer, time.Hour, nil, testSlogger())
		require.NoError(t, err)

		req := buildTestRequest(agentID, subject, []string{"read", "write"})
		tokenStr, _, err := strategy.GenerateAccessToken(ctx, req)
		require.NoError(t, err)

		jwks, err := svc.BuildJWKS(ctx)
		require.NoError(t, err)

		tok, err := jwt.Parse([]byte(tokenStr), jwt.WithKeySet(jwks))
		require.NoError(t, err, "JWT signature must verify against the JWKS public key")

		iss, ok := tok.Issuer()
		require.True(t, ok, "iss must be present")
		assert.Equal(t, issuer, iss)

		sub, ok := tok.Subject()
		require.True(t, ok, "sub must be present")
		assert.Equal(t, subject, sub)

		iat, ok := tok.IssuedAt()
		require.True(t, ok, "iat must be present")
		assert.False(t, iat.IsZero())

		exp, ok := tok.Expiration()
		require.True(t, ok, "exp must be present")
		assert.True(t, exp.After(time.Now()), "exp must be in the future")

		jti, ok := tok.JwtID()
		require.True(t, ok, "jti must be present")
		assert.NotEmpty(t, jti)

		var gotAgentID string
		require.NoError(t, tok.Get("agent_id", &gotAgentID), "agent_id must be present")
		assert.Equal(t, req.GetClient().GetID(), gotAgentID)

		var scope string
		require.NoError(t, tok.Get("scope", &scope), "scope must be present")
		assert.Equal(t, "read write", scope)
	})
}

func TestNewJWXAccessTokenStrategy_Validation(t *testing.T) {
	svc, repo := newTestSigningKeyService()
	validIssuer := "https://issuer.example.com"

	issuerCases := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{"empty string rejected", "", true},
		{"whitespace-only rejected", "   ", true},
		{"tab-only rejected", "\t", true},
		{"relative URI rejected", "/oauth2", true},
		{"no scheme rejected", "issuer.example.com", true},
		{"unsupported scheme rejected", "ftp://issuer.example.com", true},
		{"fragment rejected (RFC 8414)", "https://issuer.example.com#frag", true},
		{"query component rejected (RFC 8414)", "https://issuer.example.com?tenant=a", true},
		{"host-less https rejected", "https:///path", true},
		{"single-slash https rejected", "https:/issuer.example.com", true},
		{"port-only host rejected", "https://:443", true},
		{"empty query component rejected", "https://issuer.example.com?", true},
		{"whitespace-padded accepted after trim", " https://issuer.example.com ", false},
		{"valid https accepted", "https://issuer.example.com", false},
		{"valid http accepted", "http://localhost:8080", false},
		{"https with path accepted", "https://issuer.example.com/realms/myrealm", false},
	}
	for _, tc := range issuerCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewJWXAccessTokenStrategy(svc, repo, tc.uri, time.Hour, nil, testSlogger())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("zero tokenTTL rejected", func(t *testing.T) {
		_, err := NewJWXAccessTokenStrategy(svc, repo, validIssuer, 0, nil, testSlogger())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tokenTTL")
	})

	t.Run("negative tokenTTL rejected", func(t *testing.T) {
		_, err := NewJWXAccessTokenStrategy(svc, repo, validIssuer, -time.Second, nil, testSlogger())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tokenTTL")
	})

	t.Run("valid inputs accepted", func(t *testing.T) {
		strategy, err := NewJWXAccessTokenStrategy(svc, repo, validIssuer, time.Hour, nil, testSlogger())
		require.NoError(t, err)
		assert.NotNil(t, strategy)
	})

	t.Run("whitespace-padded issuer is stored trimmed", func(t *testing.T) {
		_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)
		strategy, err := NewJWXAccessTokenStrategy(svc, repo, "  "+validIssuer+"  ", time.Hour, nil, testSlogger())
		require.NoError(t, err)

		tokenStr, _, err := strategy.GenerateAccessToken(context.Background(), buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)

		tok, err := jwt.ParseInsecure([]byte(tokenStr))
		require.NoError(t, err)
		iss, ok := tok.Issuer()
		require.True(t, ok)
		assert.Equal(t, validIssuer, iss, "stored issuer must be trimmed, not padded")
	})
}

func TestJWXAccessTokenStrategy_SubClaimNotOverridable(t *testing.T) {
	svc, repo := newTestSigningKeyService()
	_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
	require.NoError(t, err)

	eval, err := NewTokenClaimsEvaluator(`{"sub": "attacker@evil.com", "extra": "ok"}`)
	require.NoError(t, err)

	strategy, err := NewJWXAccessTokenStrategy(svc, repo, "https://broker.example.com", time.Hour, eval, testSlogger())
	require.NoError(t, err)
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
