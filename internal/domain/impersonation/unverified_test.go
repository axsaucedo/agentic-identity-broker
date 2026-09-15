package impersonation

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsignedSubjectJWT builds an alg:none compact JWT with an empty signature segment.
func unsignedSubjectJWT(t *testing.T, payload map[string]any) string {
	t.Helper()
	enc := func(m map[string]any) string {
		b, err := json.Marshal(m)
		require.NoError(t, err)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	return enc(map[string]any{"alg": "none", "typ": "JWT"}) + "." + enc(payload) + "."
}

// FR-003d: the unverified subject parser reads caller-asserted claims without verification.
func TestParseUnverifiedSubject_ReadsClaims(t *testing.T) {
	token := unsignedSubjectJWT(t, map[string]any{"sub": "chat-user-1", "email": "user@example.com"})
	claims, err := parseUnverifiedSubject(token)
	require.NoError(t, err)
	assert.Equal(t, "chat-user-1", claims["sub"])
	assert.Equal(t, "user@example.com", claims["email"])
}

// FR-003d/R3: a signed JWS must never be accepted through the unverified route.
func TestParseUnverifiedSubject_RejectsSignedJWS(t *testing.T) {
	signed := compactJWT(t, map[string]any{"alg": "RS256"}, map[string]any{"sub": "x"})
	_, err := parseUnverifiedSubject(signed)
	require.Error(t, err)
}

// alg:none but carrying a non-empty signature segment is rejected (empty-signature assertion).
func TestParseUnverifiedSubject_RejectsNoneWithSignature(t *testing.T) {
	withSig := compactJWT(t, map[string]any{"alg": "none"}, map[string]any{"sub": "x"})
	_, err := parseUnverifiedSubject(withSig)
	require.Error(t, err)
}

func TestParseUnverifiedSubject_RejectsMalformed(t *testing.T) {
	_, err := parseUnverifiedSubject("not-a-jwt")
	require.Error(t, err)
}
