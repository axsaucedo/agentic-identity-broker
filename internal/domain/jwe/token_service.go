package jwe

import (
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwa"
	gojwe "github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

// TokenService encrypts/decrypts JSON-serializable claims into compact JWE strings.
// Algorithm: A256GCMKW key wrapping + A256GCM content encryption.
type TokenService struct {
	key jwk.Key
}

// New creates a TokenService backed by the given symmetric JWK.
func New(key jwk.Key) *TokenService {
	return &TokenService{key: key}
}

// Encrypt marshals v to JSON and returns a compact JWE string.
func (s *TokenService) Encrypt(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	encrypted, err := gojwe.Encrypt(payload,
		gojwe.WithKey(jwa.A256GCMKW(), s.key),
		gojwe.WithContentEncryption(jwa.A256GCM()),
		gojwe.WithCompact())
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	return string(encrypted), nil
}

// Decrypt decrypts a compact JWE string and unmarshals the payload into target.
func (s *TokenService) Decrypt(token string, target any) error {
	if token == "" {
		return fmt.Errorf("token is empty")
	}

	decrypted, err := gojwe.Decrypt(
		[]byte(token),
		gojwe.WithKey(jwa.A256GCMKW(), s.key),
	)
	if err != nil {
		return fmt.Errorf("failed to decrypt token: %w", err)
	}

	if err := json.Unmarshal(decrypted, target); err != nil {
		return fmt.Errorf("failed to unmarshal token payload: %w", err)
	}

	return nil
}
