package oauth2server

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// verifyPKCE verifies a PKCE S256 code challenge against a code verifier.
func verifyPKCE(codeChallenge, codeVerifier string) error {
	if codeVerifier == "" {
		return fmt.Errorf("code_verifier is required")
	}

	// S256: code_challenge = BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	expected := base64.RawURLEncoding.EncodeToString(hash[:])

	if subtle.ConstantTimeCompare([]byte(codeChallenge), []byte(expected)) != 1 {
		return fmt.Errorf("code_verifier does not match code_challenge")
	}
	return nil
}
