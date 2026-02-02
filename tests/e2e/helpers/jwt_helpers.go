// Package helpers provides test utilities for E2E testing.
package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// GenerateTestRSAKeyPair generates a test RSA key pair suitable for JWT signing.
// Uses 2048-bit RSA keys for fast test execution without sacrificing security.
// Returns (privateKeyPEM, publicKeyPEM, error)
func GenerateTestRSAKeyPair() (string, string, error) {
	// Generate 2048-bit RSA key pair (smaller keys for faster test execution)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate RSA key pair: %w", err)
	}

	// Encode private key to PEM format
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	})
	if privateKeyPEM == nil {
		return "", "", fmt.Errorf("failed to encode private key to PEM")
	}

	// Encode public key to PEM format
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})
	if publicKeyPEM == nil {
		return "", "", fmt.Errorf("failed to encode public key to PEM")
	}

	return string(privateKeyPEM), string(publicKeyPEM), nil
}

// SignTestJWT creates and signs a JWT with the provided claims using RS256 algorithm.
// The private key must be in PEM format (output from GenerateTestRSAKeyPair).
// Returns the signed JWT string (header.payload.signature format).
func SignTestJWT(claims map[string]interface{}, privateKeyPEM string) (string, error) {
	// Parse private key from PEM format
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Build JWT with provided claims
	tok := jwt.New()
	for key, value := range claims {
		if err := tok.Set(key, value); err != nil {
			return "", fmt.Errorf("failed to set claim %q: %w", key, err)
		}
	}

	// Convert private key to JWK for signing
	jwkKey, err := jwk.Import(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to import private key as JWK: %w", err)
	}

	// Set kid (key ID) for JWKS lookup
	if err := jwkKey.Set(jwk.KeyIDKey, "test-key"); err != nil {
		return "", fmt.Errorf("failed to set kid: %w", err)
	}

	// Sign JWT with RS256 and JWK (includes kid in header)
	signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256(), jwkKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	return string(signed), nil
}

// GenerateJWKSFromPublicKey converts a public key in PEM format to JWKS Set format.
// Returns a map matching the standard JWKS Set structure: {"keys": [{...key data...}]}
// The returned map can be marshalled to JSON for serving from a JWKS endpoint.
func GenerateJWKSFromPublicKey(publicKeyPEM string) (map[string]interface{}, error) {
	// Parse public key from PEM format
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Convert raw public key to lestrrat-go/jwx JWK format using jwk.Import
	jwkKey, err := jwk.Import(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JWK: %w", err)
	}

	// Set algorithm for clarity
	if err := jwkKey.Set(jwk.AlgorithmKey, jwa.RS256()); err != nil {
		return nil, fmt.Errorf("failed to set algorithm: %w", err)
	}

	// Set use to sig (signing)
	if err := jwkKey.Set(jwk.KeyUsageKey, "sig"); err != nil {
		return nil, fmt.Errorf("failed to set key usage: %w", err)
	}

	// Set kid (key ID) to match the kid in signed test tokens
	if err := jwkKey.Set(jwk.KeyIDKey, "test-key"); err != nil {
		return nil, fmt.Errorf("failed to set kid: %w", err)
	}

	// Marshal the key to JSON to get the JWKS key format
	keyJSON, err := json.Marshal(jwkKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal key to JSON: %w", err)
	}

	// Unmarshal into a map for the single key
	var keyData map[string]interface{}
	if err := json.Unmarshal(keyJSON, &keyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal key data: %w", err)
	}

	// Wrap in JWKS Set structure
	jwksSet := map[string]interface{}{
		"keys": []map[string]interface{}{keyData},
	}

	return jwksSet, nil
}
