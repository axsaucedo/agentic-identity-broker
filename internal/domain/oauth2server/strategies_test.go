package oauth2server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
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
	t.Run("no key provisioned returns actionable error", func(t *testing.T) {
		svc, _ := newTestSigningKeyService() // no key stored
		strategy, err := NewJWXAccessTokenStrategy(svc, "https://issuer.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)
		_, _, err = strategy.GenerateAccessToken(context.Background(), buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "POST /api/oauth2-server/signing-keys")
	})

	t.Run("JWT header contains the current signing key kid", func(t *testing.T) {
		svc, _ := newTestSigningKeyService()
		ctx := context.Background()

		key, err := svc.generateAndStore(ctx, "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, "https://issuer.example.com", time.Hour, nil, testSlogger())
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
		svc, _ := newTestSigningKeyService()
		_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, "https://issuer.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)
		tokenStr, sig, err := strategy.GenerateAccessToken(context.Background(), buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)

		assert.Equal(t, sha256Hex(tokenStr), sig)
	})

	t.Run("signature verifies against JWKS and all required claims are present", func(t *testing.T) {
		const issuer = "https://issuer.example.com"
		const agentID = "my-agent"
		const subject = "user@example.com"

		svc, _ := newTestSigningKeyService()
		ctx := context.Background()
		_, err := svc.generateAndStore(ctx, "ES256", true, time.Now())
		require.NoError(t, err)

		strategy, err := NewJWXAccessTokenStrategy(svc, issuer, time.Hour, nil, testSlogger())
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
	svc, _ := newTestSigningKeyService()
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
			_, err := NewJWXAccessTokenStrategy(svc, tc.uri, time.Hour, nil, testSlogger())
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("zero tokenTTL rejected", func(t *testing.T) {
		_, err := NewJWXAccessTokenStrategy(svc, validIssuer, 0, nil, testSlogger())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tokenTTL")
	})

	t.Run("negative tokenTTL rejected", func(t *testing.T) {
		_, err := NewJWXAccessTokenStrategy(svc, validIssuer, -time.Second, nil, testSlogger())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tokenTTL")
	})

	t.Run("valid inputs accepted", func(t *testing.T) {
		strategy, err := NewJWXAccessTokenStrategy(svc, validIssuer, time.Hour, nil, testSlogger())
		require.NoError(t, err)
		assert.NotNil(t, strategy)
	})

	t.Run("whitespace-padded issuer is stored trimmed", func(t *testing.T) {
		_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)
		strategy, err := NewJWXAccessTokenStrategy(svc, "  "+validIssuer+"  ", time.Hour, nil, testSlogger())
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
	svc, _ := newTestSigningKeyService()
	_, err := svc.generateAndStore(context.Background(), "ES256", true, time.Now())
	require.NoError(t, err)

	eval, err := NewTokenClaimsEvaluator(`{"sub": "attacker@evil.com", "extra": "ok"}`)
	require.NoError(t, err)

	strategy, err := NewJWXAccessTokenStrategy(svc, "https://broker.example.com", time.Hour, eval, testSlogger())
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

func TestJWXAccessTokenStrategy_ValidateAccessToken(t *testing.T) {
	const issuer = "https://broker.example.com"
	svc, repo := newTestSigningKeyService()
	ctx := context.Background()
	_, err := svc.generateAndStore(ctx, "ES256", true, time.Now())
	require.NoError(t, err)

	strategy, err := NewJWXAccessTokenStrategy(svc, issuer, time.Hour, nil, testSlogger())
	require.NoError(t, err)

	t.Run("valid token passes", func(t *testing.T) {
		tokenStr, _, err := strategy.GenerateAccessToken(ctx, buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)
		assert.NoError(t, strategy.ValidateAccessToken(ctx, nil, tokenStr))
	})

	t.Run("garbage string rejected", func(t *testing.T) {
		assert.Error(t, strategy.ValidateAccessToken(ctx, nil, "not-a-jwt"))
	})

	t.Run("tampered signature rejected", func(t *testing.T) {
		tokenStr, _, err := strategy.GenerateAccessToken(ctx, buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)

		parts := strings.SplitN(tokenStr, ".", 3)
		require.Len(t, parts, 3)
		tampered := parts[0] + "." + parts[1] + ".invalidsignature"

		assert.Error(t, strategy.ValidateAccessToken(ctx, nil, tampered))
	})

	t.Run("expired token rejected", func(t *testing.T) {
		key, err := repo.GetCurrent(ctx)
		require.NoError(t, err)
		privPEM, err := svc.DecryptPrivateKey(ctx, key)
		require.NoError(t, err)
		privKey, err := jwk.ParseKey(privPEM, jwk.WithPEM(true))
		require.NoError(t, err)
		_ = privKey.Set(jwk.KeyIDKey, key.KID.String())

		now := time.Now()
		expiredTok, err := jwt.NewBuilder().
			Issuer(issuer).
			Subject("user@example.com").
			IssuedAt(now.Add(-2 * time.Hour)).
			Expiration(now.Add(-time.Hour)).
			JwtID("expired-jti").
			Build()
		require.NoError(t, err)

		signed, err := jwt.Sign(expiredTok, jwt.WithKey(jwa.ES256(), privKey))
		require.NoError(t, err)

		assert.Error(t, strategy.ValidateAccessToken(ctx, nil, string(signed)))
	})

	t.Run("token from different issuer rejected", func(t *testing.T) {
		evilStrategy, err := NewJWXAccessTokenStrategy(svc, "https://evil.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)

		tokenStr, _, err := evilStrategy.GenerateAccessToken(ctx, buildTestRequest("agent", "user@example.com", []string{"read"}))
		require.NoError(t, err)

		assert.Error(t, strategy.ValidateAccessToken(ctx, nil, tokenStr))
	})
}

// connectionErrorSigningKeyRepo returns a connection-kind StorageError from GetCurrent
// (and GetByKID) to simulate a DB outage or timeout. All other methods delegate to
// an in-memory store so the rest of the service is functional.
type connectionErrorSigningKeyRepo struct {
	*memory.SigningKeyStore
}

var _ ports.SigningKeyRepository = (*connectionErrorSigningKeyRepo)(nil)

func (r *connectionErrorSigningKeyRepo) GetCurrent(_ context.Context) (*storage.SigningKey, error) {
	return nil, storage.NewStorageError(
		"SigningKeyRepo.GetCurrent",
		storage.ErrorKindConnection,
		nil,
		"connection refused",
	)
}

func (r *connectionErrorSigningKeyRepo) GetByKID(_ context.Context, _ id.KeyID) (*storage.SigningKey, error) {
	return nil, storage.NewStorageError(
		"SigningKeyRepo.GetByKID",
		storage.ErrorKindConnection,
		nil,
		"connection refused",
	)
}

func TestJWXAccessTokenStrategy_GetCurrent_NonNotFoundError(t *testing.T) {
	t.Run("connection error does not produce 'no signing key provisioned' message", func(t *testing.T) {
		repo := &connectionErrorSigningKeyRepo{SigningKeyStore: memory.NewSigningKeyStore()}
		enc := &testEncryptor{}
		svc := NewSigningKeyService(repo, enc, nil, testSlogger())
		strategy, err := NewJWXAccessTokenStrategy(svc, "https://issuer.example.com", time.Hour, nil, testSlogger())
		require.NoError(t, err)

		_, _, err = strategy.GenerateAccessToken(
			context.Background(),
			buildTestRequest("agent", "user@example.com", []string{"read"}),
		)
		require.Error(t, err)
		// Must NOT produce the misleading "no signing key provisioned" message.
		assert.NotContains(t, err.Error(), "no signing key provisioned",
			"connection errors must not be misreported as missing keys")
		// Must propagate the real error so operators see 'failed to get current signing key'.
		assert.Contains(t, err.Error(), "failed to get current signing key")
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
