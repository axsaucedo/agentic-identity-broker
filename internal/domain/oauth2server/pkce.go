package oauth2server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// verifyPKCE verifies a PKCE S256 code challenge against a code verifier.
// The verifier must be 43–128 characters per RFC 7636 §4.1.
func verifyPKCE(codeChallenge, codeVerifier string) error {
	n := len(codeVerifier)
	if n < 43 || n > 128 {
		return fmt.Errorf("code_verifier length %d out of range [43, 128]", n)
	}

	// S256: code_challenge = BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])

	if subtle.ConstantTimeCompare([]byte(codeChallenge), []byte(expected)) != 1 {
		return fmt.Errorf("code_verifier does not match code_challenge")
	}
	return nil
}
