package oauth2server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	dstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// newTestProvider creates a Provider with in-memory storage and test encryption
// for unit testing purposes.
func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	codeRepo := memory.NewAuthorizationCodeStore()
	credRepo := memory.NewBrokerClientCredentialStore()
	agentRepo := memory.NewAgentRepository()
	signingKeyRepo := memory.NewSigningKeyStore()
	enc := &testEncryptor{}
	logger := testSlogger()

	provider, err := NewProvider(
		codeRepo,
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
	_, err = signingKeySvc.GenerateAndStoreKey(context.Background(), "ES256", true)
	require.NoError(t, err)

	return provider
}

// setupTestCredentials creates an agent with broker credentials and returns the agent, credential, and plaintext secret.
func setupTestCredentials(t *testing.T, provider *Provider) (*dstorage.Agent, *dstorage.BrokerClientCredential, string) {
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
		_, cred, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(
			context.Background(),
			cred.BrokerClientID,
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
		_, cred, _ := setupTestCredentials(t, provider)

		_, err := provider.HandleClientCredentials(
			context.Background(),
			cred.BrokerClientID,
			"wrong-secret",
			"read",
		)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("unknown client_id returns error", func(t *testing.T) {
		provider := newTestProvider(t)

		_, err := provider.HandleClientCredentials(
			context.Background(),
			id.NewBrokerClientID("broker_unknown"),
			"secret",
			"read",
		)
		assert.ErrorIs(t, err, ErrInvalidClient)
	})

	t.Run("empty scope works", func(t *testing.T) {
		provider := newTestProvider(t)
		_, cred, plaintext := setupTestCredentials(t, provider)

		resp, err := provider.HandleClientCredentials(
			context.Background(),
			cred.BrokerClientID,
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
		agent, cred, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
		_, cred, _ := setupTestCredentials(t, provider)
		// Agent created by setupTestCredentials has no RedirectURIs (empty slice)

		_, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
		agent, cred, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		// Update agent with redirect URIs
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		code, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
		agent, cred, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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

	t.Run("unregistered redirect_uri rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, cred, _ := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		_, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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

	t.Run("unknown client_id rejected", func(t *testing.T) {
		provider := newTestProvider(t)

		_, err := provider.HandleAuthorize(
			context.Background(),
			id.NewBrokerClientID("broker_unknown"),
			"http://localhost:8080/callback",
			"code",
			"read",
			"state",
			"challenge123",
			"S256",
			id.NewPrincipal("user@example.com"),
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUnknownClient)
	})
}

func TestProvider_HandleAuthorizationCodeExchange(t *testing.T) {
	t.Run("valid code exchange returns token", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, cred, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		// First, get an authorization code with a known verifier
		verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
			cred.BrokerClientID,
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
		agent, cred, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "test-verifier-for-replay"
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
			cred.BrokerClientID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			verifier,
		)
		require.NoError(t, err)

		// Second exchange (replay) should fail
		_, err = provider.HandleAuthorizationCodeExchange(
			context.Background(),
			cred.BrokerClientID,
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
		agent, cred, plaintext := setupTestCredentials(t, provider)
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
			cred.BrokerClientID,
			plaintext,
			rawCode,
			"http://localhost:8080/callback",
			verifier,
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})

	t.Run("PKCE mismatch rejected", func(t *testing.T) {
		provider := newTestProvider(t)
		agent, cred, plaintext := setupTestCredentials(t, provider)
		agent.RedirectURIs = []string{"http://localhost:8080/callback"}
		_ = provider.fositeStorage.agentRepo.Update(context.Background(), agent)

		verifier := "correct-verifier"
		challenge := generateS256Challenge(verifier)

		code, err := provider.HandleAuthorize(
			context.Background(),
			cred.BrokerClientID,
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
			cred.BrokerClientID,
			plaintext,
			code,
			"http://localhost:8080/callback",
			"wrong-verifier",
		)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidGrant)
	})
}

// generateS256Challenge generates a PKCE S256 challenge from a verifier.
func generateS256Challenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
