package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// mockJWKSPort is a hand-rolled test double for ports.JWKSPort.
type mockJWKSPort struct {
	keySet jwk.Set
	err    error
}

func (m *mockJWKSPort) GetKeySet(_ context.Context) (jwk.Set, error) {
	return m.keySet, m.err
}

func (m *mockJWKSPort) GetKey(_ context.Context, kid string) (jwk.Key, error) {
	if m.err != nil {
		return nil, m.err
	}
	key, found := m.keySet.LookupKeyID(kid)
	if !found {
		return nil, fmt.Errorf("key %q not found", kid)
	}
	return key, nil
}

var _ ports.JWKSPort = (*mockJWKSPort)(nil)

// generateTestRSAKeyPair generates an RSA key pair for test JWT signing.
func generateTestRSAKeyPair(t *testing.T) (*rsa.PrivateKey, jwk.Set) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubJWK, err := jwk.Import(&privateKey.PublicKey)
	require.NoError(t, err)
	require.NoError(t, pubJWK.Set(jwk.KeyIDKey, "test-key-id"))
	require.NoError(t, pubJWK.Set(jwk.AlgorithmKey, jwa.RS256()))

	keySet := jwk.NewSet()
	require.NoError(t, keySet.AddKey(pubJWK))
	return privateKey, keySet
}

// signTestJWT signs a JWT with the given RSA private key and returns the compact serialization.
func signTestJWT(t *testing.T, privateKey *rsa.PrivateKey, claims map[string]interface{}) string {
	t.Helper()
	tok := jwt.New()
	for k, v := range claims {
		require.NoError(t, tok.Set(k, v))
	}
	privJWK, err := jwk.Import(privateKey)
	require.NoError(t, err)
	require.NoError(t, privJWK.Set(jwk.KeyIDKey, "test-key-id"))
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), privJWK))
	require.NoError(t, err)
	return string(signed)
}

// buildTokenResponse wraps a JWT string in a JSON access_token response body.
func buildTokenResponse(t *testing.T, accessToken string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]string{"access_token": accessToken})
	require.NoError(t, err)
	return body
}

func TestNewMultiAgentTokenVerifier(t *testing.T) {
	_, keySet := generateTestRSAKeyPair(t)
	jwksPort := &mockJWKSPort{keySet: keySet}

	t.Run("success", func(t *testing.T) {
		v, err := NewMultiAgentTokenVerifier("x_agent_id", jwksPort)
		require.NoError(t, err)
		assert.NotNil(t, v)
	})

	t.Run("empty claim name returns error", func(t *testing.T) {
		_, err := NewMultiAgentTokenVerifier("", jwksPort)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agentIDClaimName")
	})

	t.Run("nil jwksPort returns error", func(t *testing.T) {
		_, err := NewMultiAgentTokenVerifier("x_agent_id", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "jwksPort")
	})
}

func TestMultiAgentTokenVerifier_VerifyAgentIDClaim(t *testing.T) {
	ctx := context.Background()
	agentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	claimName := "x_agent_id"

	t.Run("success: valid signature and matching claim", func(t *testing.T) {
		privateKey, keySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: keySet}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		accessToken := signTestJWT(t, privateKey, map[string]interface{}{
			claimName: agentID.String(),
		})
		body := buildTokenResponse(t, accessToken)

		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.NoError(t, err)
	})

	t.Run("rejects token signed with unknown key (signature verification)", func(t *testing.T) {
		// Create two different key pairs: one used to sign the token, another in the JWKS.
		// This is the core security test: the verifier must reject tokens with invalid signatures.
		privateKey, _ := generateTestRSAKeyPair(t)
		_, differentKeySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: differentKeySet}

		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		accessToken := signTestJWT(t, privateKey, map[string]interface{}{
			claimName: agentID.String(),
		})
		body := buildTokenResponse(t, accessToken)

		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.Error(t, err, "must reject token signed with key not in JWKS")
	})

	t.Run("rejects when JWKS fetch fails", func(t *testing.T) {
		privateKey, _ := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{err: fmt.Errorf("JWKS endpoint unreachable")}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		accessToken := signTestJWT(t, privateKey, map[string]interface{}{
			claimName: agentID.String(),
		})
		body := buildTokenResponse(t, accessToken)

		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent ID claim absent")
	})

	t.Run("claim mismatch returns AgentIDMismatchError", func(t *testing.T) {
		privateKey, keySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: keySet}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		wrongAgentID := id.MustParseAgentID("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22")
		accessToken := signTestJWT(t, privateKey, map[string]interface{}{
			claimName: wrongAgentID.String(),
		})
		body := buildTokenResponse(t, accessToken)

		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.Error(t, err)
		var mismatch *AgentIDMismatchError
		require.ErrorAs(t, err, &mismatch)
		assert.Equal(t, agentID.String(), mismatch.Expected)
		assert.Equal(t, wrongAgentID.String(), mismatch.Received)
	})

	t.Run("absent claim returns descriptive error", func(t *testing.T) {
		privateKey, keySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: keySet}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		// Token has no x_agent_id claim
		accessToken := signTestJWT(t, privateKey, map[string]interface{}{
			"sub": "some-subject",
		})
		body := buildTokenResponse(t, accessToken)

		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent ID claim absent")
	})

	t.Run("malformed JSON response body", func(t *testing.T) {
		_, keySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: keySet}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		err = v.VerifyAgentIDClaim(ctx, []byte("not json"), agentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent ID claim absent")
	})

	t.Run("missing access_token in response", func(t *testing.T) {
		_, keySet := generateTestRSAKeyPair(t)
		jwksPort := &mockJWKSPort{keySet: keySet}
		v, err := NewMultiAgentTokenVerifier(claimName, jwksPort)
		require.NoError(t, err)

		body, _ := json.Marshal(map[string]string{"token_type": "bearer"})
		err = v.VerifyAgentIDClaim(ctx, body, agentID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent ID claim absent")
	})
}
