package tokenexchange

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func b64url(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// TestDiagnoseTokenShape verifies the non-reversible structural diagnostics emitted
// for malformed subject tokens. Per SR-005 the output must never contain raw token
// bytes, the payload, or the signature — only structure and allowlisted JOSE header
// metadata.
func TestDiagnoseTokenShape(t *testing.T) {
	t.Parallel()

	const secretPayload = "SUPER_SECRET_PAYLOAD_CLAIMS"
	const secretSig = "SIGNATURE_BYTES_DO_NOT_LEAK"

	jweHeader := b64url(`{"alg":"RSA-OAEP-256","enc":"A256GCM"}`)
	jweToken := strings.Join([]string{jweHeader, b64url("enckey"), b64url("iv"), b64url(secretPayload), b64url(secretSig)}, ".")

	jwsHeader := b64url(`{"alg":"RS256","kid":"key-42","typ":"JWT"}`)
	jwsToken := strings.Join([]string{jwsHeader, b64url(secretPayload), b64url(secretSig)}, ".")

	tests := []struct {
		name         string
		token        string
		wantContains []string
		wantExcludes []string
		wantEmpty    bool
	}{
		{
			name:  "jwe five segment surfaces alg and enc",
			token: jweToken,
			wantContains: []string{
				"segments=5", "dot_count=4", "first_byte_class=base64url",
				`alg="RSA-OAEP-256"`, `enc="A256GCM"`,
			},
			// Header segment is decoded to metadata only; payload/signature and any
			// raw segment bytes must never appear.
			wantExcludes: []string{secretPayload, secretSig, b64url(secretPayload), b64url(secretSig)},
		},
		{
			name:  "jws three segment surfaces alg kid typ",
			token: jwsToken,
			wantContains: []string{
				"segments=3", "dot_count=2",
				`alg="RS256"`, `kid="key-42"`, `typ="JWT"`,
			},
			wantExcludes: []string{secretPayload, secretSig, b64url(secretPayload), b64url(secretSig)},
		},
		{
			name:         "opaque token has no header and does not echo the value",
			token:        "opaque-access-token-value-12345",
			wantContains: []string{"segments=1", "dot_count=0", "first_byte_class=base64url"},
			wantExcludes: []string{"header=", "opaque-access-token-value-12345"},
		},
		{
			name:         "json blob classified by first byte without header",
			token:        `{"keys":[{"kty":"RSA"}]}`,
			wantContains: []string{"first_byte_class=json_open_brace", "segments=1"},
			wantExcludes: []string{"header=", `"kty"`},
		},
		{
			name:         "leading whitespace classified as whitespace",
			token:        "  eyJ",
			wantContains: []string{"first_byte_class=whitespace"},
		},
		{
			name:      "empty token yields empty diagnostics",
			token:     "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := diagnoseTokenShape(tt.token)
			if tt.wantEmpty {
				assert.Empty(t, got)
				return
			}
			for _, want := range tt.wantContains {
				assert.Contains(t, got, want)
			}
			for _, exclude := range tt.wantExcludes {
				assert.NotContains(t, got, exclude, "diagnostics must not leak token content")
			}
		})
	}
}

// TestDiagnoseTokenShape_HostileHeaderBounded verifies oversized and control-character
// JOSE header fields are length-bounded and escaped so they cannot bloat or corrupt logs.
func TestDiagnoseTokenShape_HostileHeaderBounded(t *testing.T) {
	t.Parallel()

	longKid := strings.Repeat("A", 200)
	header := b64url(`{"alg":"HS256","kid":"` + longKid + `"}`)
	token := strings.Join([]string{header, b64url("p"), b64url("s")}, ".")

	got := diagnoseTokenShape(token)
	assert.Contains(t, got, "...", "oversized header value must be truncated")
	assert.NotContains(t, got, strings.Repeat("A", maxHeaderValueLen+1),
		"header value must be bounded to maxHeaderValueLen")

	ctrlHeader := b64url(`{"alg":"none","kid":"k\n1"}`)
	ctrlToken := strings.Join([]string{ctrlHeader, b64url("p"), b64url("s")}, ".")
	gotCtrl := diagnoseTokenShape(ctrlToken)
	assert.NotContains(t, gotCtrl, "\n", "control characters in header must be escaped, not emitted raw")
	assert.Contains(t, gotCtrl, `\n`, "control character must appear escaped")
}

// TestValidateSubjectToken_JWEShapeDiagnostics verifies end-to-end that a JWE-shaped
// subject token produces an invalid_request error whose internal Details() carry the
// shape diagnostics, while the client-facing Description() never does (SR-005 boundary).
func TestValidateSubjectToken_JWEShapeDiagnostics(t *testing.T) {
	t.Parallel()

	_, keySet := newTestKeyPair(t)
	validator, err := NewJWTValidator(&MockJWKSProvider{keySet: keySet}, "https://auth.example.com", "broker-id", 60)
	require.NoError(t, err)

	jweHeader := b64url(`{"alg":"RSA-OAEP-256","enc":"A256GCM"}`)
	jwe := strings.Join([]string{jweHeader, b64url("enckey"), b64url("iv"), b64url("ct"), b64url("tag")}, ".")

	_, valErr := validator.ValidateSubjectToken(context.Background(), jwe)
	require.Error(t, valErr)

	tokenErr, ok := valErr.(*TokenExchangeError)
	require.True(t, ok)

	assert.Equal(t, "invalid_request", tokenErr.Code())
	assert.Contains(t, tokenErr.Details(), "segments=5")
	assert.Contains(t, tokenErr.Details(), `enc="A256GCM"`)

	// Client boundary: the description returned to the caller must not carry any of
	// the structural diagnostics — those live only in Details() for internal logging.
	assert.NotContains(t, tokenErr.Description(), "segments=")
	assert.NotContains(t, tokenErr.Description(), "seg_b64url")
	assert.NotContains(t, tokenErr.Description(), "header=")
	assert.NotContains(t, tokenErr.Description(), "enc=")
}
