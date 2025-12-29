package oauth2session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
)

// GeneratePKCE creates a PKCE code verifier and challenge per RFC 7636.
// verifierLength should be 32-128 bytes (recommended: 32 = 256 bits).
func GeneratePKCE(verifierLength int) (verifier, challenge string, err error) {
	if verifierLength < 32 || verifierLength > 128 {
		return "", "", errors.New("verifier length must be 32-128 bytes")
	}

	// Generate cryptographically random bytes
	randomBytes := make([]byte, verifierLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64url encode without padding
	verifier = base64.RawURLEncoding.EncodeToString(randomBytes)

	// SHA256 hash of verifier, base64url encoded
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])

	return verifier, challenge, nil
}
