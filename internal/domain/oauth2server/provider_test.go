package oauth2server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	dstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// newTestProvider creates a Provider with in-memory storage and test encryption
// for unit testing purposes.
func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	codeRepo := memory.NewAuthorizationCodeStore()
	credRepo := memory.NewClientCredentialStore()
	agentRepo := memory.NewAgentRepository()
	signingKeyRepo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := testSlogger()

	provider, err := NewProvider(
		codeRepo,
		memory.NewPKCESessionStore(),
		credRepo,
		agentRepo,
		signingKeyRepo,
		enc,
		"https://broker.example.com",
		time.Hour, // 1h TTL
		"",        // no CEL expression
		logger,
	)
	require.NoError(t, err)

	// Generate a signing key
	signingKeySvc := NewSigningKeyService(signingKeyRepo, enc, logger)
	_, err = signingKeySvc.generateAndStore(context.Background(), "ES256", true, time.Now())
	require.NoError(t, err)

	return provider
}

// setupTestCredentials creates an agent with broker credentials and returns the agent, credential, and plaintext secret.
func setupTestCredentials(t *testing.T, provider *Provider) (*dstorage.Agent, *dstorage.ClientCredential, string) {
	t.Helper()
	ctx := context.Background()

	// Create agent
	agent := &dstorage.Agent{
		ID:          id.NewAgentID(),
		ClientID:    id.ClientID("test-oauth2-client"),
		DisplayName: "Test OAuth2 Agent",
		Description: "Agent for provider test",
	}
	err := provider.fositeStorage.agentRepo.Create(ctx, agent)
	require.NoError(t, err)

	// Generate credentials
	cred, plaintext, err := provider.clientAuth.GenerateCredentials(agent.ID)
	require.NoError(t, err)

	err = provider.fositeStorage.credRepo.Create(ctx, cred)
	require.NoError(t, err)

	return agent, cred, plaintext
}

func TestProvider_HandleClientCredentials(t *testing.T) {
	t.Run("valid credentials return signed JWT", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(
			context.Background(),
			agent.ID,
			plaintext,
			"read write",
		)
		require.NoError(t, err)
		require.NotNil(t, resp)

		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "Bearer", resp.TokenType)
		assert.True(t, resp.ExpiresIn > 0)
		assert.Equal(t, "read write", resp.Scope)
	})

	t.Run("invalid credentials return error", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)

		_, err := provider.HandleClientCredentials(
			context.Background(),
			agent.ID,
			"wrong-secret",
			"read",
		)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("unknown agent_id returns error", func(t *testing.T) {
		provider := newTestProvider(t)

		_, err := provider.HandleClientCredentials(
			context.Background(),
			id.NewAgentID(),
			"secret",
			"read",
		)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("empty scope works", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(
			context.Background(),
			agent.ID,
			plaintext,
			"",
		)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.AccessToken)
	})
}

func TestProvider_HandleAuthorize(t *testing.T) {
	t.Run("redirect_uri exact match required", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost/callback/extra", // has extra path segment
			"code",
			"read",
			"state",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRedirectURI)
	})

	t.Run("empty redirect_uris list rejects", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		// Agent created by setupTestCredentials has no RedirectURIs (empty slice)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRedirectURI)
	})

	t.Run("valid request returns authorization code", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		// Update agent with redirect URIs
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"test-state-123",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)
		assert.NotEmpty(t, code)
	})

	t.Run("missing code_challenge rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"", // empty code_challenge
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequest)
	})

	t.Run("empty code_challenge_method rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"challenge123",
			"", // missing method — must not default to plain
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequest)
	})

	t.Run("plain code_challenge_method rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"challenge123",
			"plain",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequest)
	})

	t.Run("unregistered redirect_uri rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://evil.example.com/callback", // not registered
			"code",
			"read",
			"state",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRedirectURI)
	})

	t.Run("unknown agent_id rejected", func(t *testing.T) {
		provider := newTestProvider(t)

		_, err := provider.HandleAuthorize(
			context.Background(),
			id.NewAgentID(),
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})
}

func TestProvider_HandleAuthorizationCodeExchange(t *testing.T) {
	t.Run("valid code exchange returns token", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		// First, get an authorization code with a known verifier
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			challenge,
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)
		require.NotEmpty(t, code)

		// Exchange the code
		resp, err := provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			verifier,
		)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.AccessToken)
		assert.Equal(t, "Bearer", resp.TokenType)
	})

	t.Run("code replay rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "replay-code-verifier-xxxxxxxxxxxxxxxxxxxxxxxxxxx" // 43 chars (RFC 7636 minimum)
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			challenge,
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)

		// First exchange succeeds
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			verifier,
		)
		require.NoError(t, err)

		// Second exchange (replay) should fail
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			verifier,
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})

	t.Run("expired code rejects", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "test-verifier-for-expired"
		challenge := generateS256Challenge(verifier)

		// Directly create an authorization code that is already expired
		rawCode := "test-expired-code-value"
		codeHash := sha256Hex(rawCode)
		expiredCode := &dstorage.AuthorizationCode{
			ID:            id.NewAuthorizationCodeID(),
			CodeHash:      codeHash,
			AgentID:       agent.ID,
			Principal:     id.NewPrincipal("user@example.com"),
			RedirectURI:   "http://localhost:8080/callback",
			CodeChallenge: challenge,
			Scope:         "read",
			ExpiresAt:     time.Now().Add(-10 * time.Minute), // expired
			CreatedAt:     time.Now().Add(-15 * time.Minute),
		}
		err := provider.fositeStorage.codeRepo.Create(context.Background(), expiredCode)
		require.NoError(t, err)

		// Attempt exchange — should fail because code is expired
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			rawCode,
			"http://localhost:8080/callback",
			verifier,
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})

	t.Run("redirect_uri substitution rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback", "https://attacker.example.com/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "test-verifier-for-uri-substitution"
		challenge := generateS256Challenge(verifier)

		// Authorization was issued to the legitimate URI
		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			challenge,
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)

		// Attacker substitutes a different registered URI at token exchange (RFC 6749 §4.1.3)
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"https://attacker.example.com/callback",
			verifier,
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})

	t.Run("PKCE mismatch rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "correct-verifier"
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			challenge,
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)

		// Exchange with wrong verifier
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			"wrong-verifier",
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
}

// TestProvider_HandleAuthorizationCodeExchange_ConcurrentReplay verifies that when MarkUsed
// returns ErrorKindNotFound (the DB-level race guard fired — a concurrent request won and
// set used_at before ours could), the exchange returns ErrInvalidGrant rather than a 500.
func TestProvider_HandleAuthorizationCodeExchange_ConcurrentReplay(t *testing.T) {
	provider := newTestProvider(t)
	agent, _, plaintext := setupTestCredentials(t, provider)
	agent.RedirectURIs = []string{"http://localhost:8080/callback"}
	_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

	verifier := "concurrent-replay-verifier"
	challenge := generateS256Challenge(verifier)

	code, err := provider.HandleAuthorize(
		context.Background(),
		agent.ID,
		"http://localhost:8080/callback",
		"code",
		"read",
		"state",
		challenge,
		"S256",
		id.NewPrincipal("user@example.com"),
	)
	require.NoError(t, err)

	// Replace codeRepo with a wrapper that simulates the race: FindByCodeHash returns the
	// unused code (used_at IS NULL), but MarkUsed returns ErrorKindNotFound (zero rows
	// affected because a concurrent request already set used_at in the database).
	provider.fositeStorage.codeRepo = &raceCodeRepo{
		delegate:      provider.fositeStorage.codeRepo,
		markUsedError: dstorage.NewStorageError("AuthorizationCodeRepo.MarkUsed", dstorage.ErrorKindNotFound, nil, "authorization code not found or already used"),
	}

	_, err = provider.HandleAuthorizationCodeExchange(
		context.Background(),
		agent.ID,
		plaintext,
		code,
		"http://localhost:8080/callback",
		verifier,
	)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidGrant)
}

// raceCodeRepo wraps an AuthorizationCodeRepository to inject a fixed error on MarkUsed,
// simulating the concurrent replay scenario where the DB guard fires first.
type raceCodeRepo struct {
	delegate      ports.AuthorizationCodeRepository
	markUsedError error
}

func (r *raceCodeRepo) Create(ctx context.Context, code *dstorage.AuthorizationCode) error {
	return r.delegate.Create(ctx, code)
}

func (r *raceCodeRepo) FindByCodeHash(ctx context.Context, codeHash string) (*dstorage.AuthorizationCode, error) {
	return r.delegate.FindByCodeHash(ctx, codeHash)
}

func (r *raceCodeRepo) MarkUsed(_ context.Context, _ id.AuthorizationCodeID) error {
	return r.markUsedError
}

func (r *raceCodeRepo) DeleteExpired(ctx context.Context) (int, error) {
	return r.delegate.DeleteExpired(ctx)
}

// TestProvider_AccessToken_ClaimsAndSignature verifies that tokens issued via both
// grant flows carry all required claims and a verifiable ECDSA signature.
func TestProvider_AccessToken_ClaimsAndSignature(t *testing.T) {
	const issuer = "https://broker.example.com"

	verifyToken := func(t *testing.T, provider *Provider, tokenStr, wantSub, wantAgentID, wantScope string) {
		t.Helper()
		ctx := context.Background()

		jwks, err := provider.signingKeyService.BuildJWKS(ctx)
		require.NoError(t, err)

		tok, err := jwt.Parse([]byte(tokenStr), jwt.WithKeySet(jwks))
		require.NoError(t, err, "JWT signature must verify against the JWKS public key")

		iss, ok := tok.Issuer()
		require.True(t, ok, "iss must be present")
		assert.Equal(t, issuer, iss)

		sub, ok := tok.Subject()
		require.True(t, ok, "sub must be present")
		assert.Equal(t, wantSub, sub)

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
		assert.Equal(t, wantAgentID, gotAgentID)

		var scope string
		require.NoError(t, tok.Get("scope", &scope), "scope must be present")
		assert.Equal(t, wantScope, scope)
	}

	t.Run("client_credentials flow", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(context.Background(), agent.ID, plaintext, "read write")
		require.NoError(t, err)

		// In client_credentials, sub and agent_id are both the agent UUID (brokerClient.GetID() returns agent.ID).
		verifyToken(t, provider, resp.AccessToken, agent.ID.String(), agent.ID.String(), "read write")
	})

	t.Run("authorization_code flow", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "pkce-verifier-for-claims-test-abcdefghijklm" // 43 chars (RFC 7636 minimum)
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			agent.ID,
			"http://localhost:8080/callback",
			"code",
			"read write",
			"state",
			challenge,
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)

		resp, err := provider.HandleAuthorizationCodeExchange(
			context.Background(),
			agent.ID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			verifier,
		)
		require.NoError(t, err)

		// agent_id is the agent UUID; sub is the authenticated principal.
		verifyToken(t, provider, resp.AccessToken, "user@example.com", agent.ID.String(), "read write")
	})
}

// TestProvider_CEL_RequestGrantType verifies that request.grant_type is correctly
// populated for both client_credentials and authorization_code flows end-to-end.
func TestProvider_CEL_RequestGrantType(t *testing.T) {
	newProviderWithCEL := func(t *testing.T, expr string) *Provider {
		t.Helper()
		codeRepo := memory.NewAuthorizationCodeStore()
		credRepo := memory.NewClientCredentialStore()
		agentRepo := memory.NewAgentRepository()
		signingKeyRepo := memory.NewSigningKeyStore()
		enc := &testEncryptor{}
		logger := testSlogger()

		p, err := NewProvider(codeRepo, memory.NewPKCESessionStore(), credRepo, agentRepo, signingKeyRepo, enc,
			"https://broker.example.com", time.Hour, expr, logger)
		require.NoError(t, err)

		svc := NewSigningKeyService(signingKeyRepo, enc, logger)
		_, err = svc.generateAndStore(context.Background(), "ES256", true, time.Now())
		require.NoError(t, err)
		return p
	}

	getCustomClaim := func(t *testing.T, tokenStr, claim string) string {
		t.Helper()
		tok, err := jwt.Parse([]byte(tokenStr), jwt.WithValidate(false), jwt.WithVerify(false))
		require.NoError(t, err)
		var v string
		require.NoError(t, tok.Get(claim, &v))
		return v
	}

	t.Run("client_credentials grant populates request.grant_type", func(t *testing.T) {
		provider := newProviderWithCEL(t, `{"grant": request.grant_type}`)
		agent, _, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(context.Background(), agent.ID, plaintext, "read")
		require.NoError(t, err)

		assert.Equal(t, "client_credentials", getCustomClaim(t, resp.AccessToken, "grant"))
	})

	t.Run("authorization_code grant populates request.grant_type", func(t *testing.T) {
		provider := newProviderWithCEL(t, `{"grant": request.grant_type}`)
		agent, _, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "pkce-verifier-for-grant-type-test-abcdefghi" // 43 chars
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(), agent.ID,
			"http://localhost:8080/callback", "code", "read", "state",
			challenge, "S256", id.NewPrincipal("user@example.com"),
		)
		require.NoError(t, err)

		resp, err := provider.HandleAuthorizationCodeExchange(
			context.Background(), agent.ID, plaintext,
			code, "http://localhost:8080/callback", verifier,
		)
		require.NoError(t, err)

		assert.Equal(t, "authorization_code", getCustomClaim(t, resp.AccessToken, "grant"))
	})
}

// generateS256Challenge generates a PKCE S256 challenge from a verifier.
func generateS256Challenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
