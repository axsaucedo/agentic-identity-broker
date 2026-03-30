package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCEVerifier generates a random PKCE code verifier (43-128 chars, base64url).
func PKCEVerifier() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GenerateCodeChallenge generates a S256 PKCE code challenge from a verifier.
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
