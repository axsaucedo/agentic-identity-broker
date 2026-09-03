// Package helpers provides test utilities for E2E testing.
package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
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
	jwkKey, err := jwk.Import[jwk.Key](privateKey)
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
	jwkKey, err := jwk.Import[jwk.Key](publicKey)
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

// CreateUnsignedJWT creates an unsigned JWT (alg: "none") with the provided claims.
// Used for testing unsigned JWT pre-auth in service mesh environments.
// Returns the JWT string in the format: base64(header).base64(payload).
func CreateUnsignedJWT(claims map[string]interface{}) (string, error) {
	encoding := base64.RawURLEncoding

	// Header: {"alg":"none","typ":"JWT"}
	header := map[string]string{"alg": "none", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}
	headerEncoded := encoding.EncodeToString(headerJSON)

	// Payload: claims
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}
	payloadEncoded := encoding.EncodeToString(payloadJSON)

	// Unsigned JWT: header.payload. (empty signature)
	return headerEncoded + "." + payloadEncoded + ".", nil
}

// JWTClaimsBuilder provides a fluent interface for building JWT claims maps.
// It generates standard claims and allows customization for various test scenarios.
type JWTClaimsBuilder struct {
	claims map[string]interface{}
}

// NewJWTClaims creates a JWTClaimsBuilder with standard claims for testing.
// Default claims: sub, iss, aud, exp (1 hour from now), iat (now).
func NewJWTClaims() *JWTClaimsBuilder {
	now := time.Now()
	return &JWTClaimsBuilder{
		claims: map[string]interface{}{
			"sub": "user@example.com",
			"iss": "https://auth.example.com",
			"aud": "agentic-identity-broker",
			"exp": now.Add(1 * time.Hour).Unix(),
			"iat": now.Unix(),
		},
	}
}

// WithSubject sets the sub claim.
func (b *JWTClaimsBuilder) WithSubject(sub string) *JWTClaimsBuilder {
	b.claims["sub"] = sub
	return b
}

// WithIssuer sets the iss claim.
func (b *JWTClaimsBuilder) WithIssuer(iss string) *JWTClaimsBuilder {
	b.claims["iss"] = iss
	return b
}

// WithAudience sets the aud claim.
func (b *JWTClaimsBuilder) WithAudience(aud string) *JWTClaimsBuilder {
	b.claims["aud"] = aud
	return b
}

// WithExpiry sets the exp claim to a specific time.
func (b *JWTClaimsBuilder) WithExpiry(exp time.Time) *JWTClaimsBuilder {
	b.claims["exp"] = exp.Unix()
	return b
}

// WithExpired sets the exp claim to 1 hour in the past (expired token).
func (b *JWTClaimsBuilder) WithExpired() *JWTClaimsBuilder {
	b.claims["exp"] = time.Now().Add(-1 * time.Hour).Unix()
	return b
}

// WithoutExpiry removes the exp claim (for testing missing expiry rejection).
func (b *JWTClaimsBuilder) WithoutExpiry() *JWTClaimsBuilder {
	delete(b.claims, "exp")
	return b
}

// WithName sets the name claim (display name).
func (b *JWTClaimsBuilder) WithName(name string) *JWTClaimsBuilder {
	b.claims["name"] = name
	return b
}

// WithEmail sets the email claim.
func (b *JWTClaimsBuilder) WithEmail(email string) *JWTClaimsBuilder {
	b.claims["email"] = email
	return b
}

// WithPicture sets the picture claim (profile picture URL).
func (b *JWTClaimsBuilder) WithPicture(pictureURL string) *JWTClaimsBuilder {
	b.claims["picture"] = pictureURL
	return b
}

// WithClaim sets an arbitrary claim.
func (b *JWTClaimsBuilder) WithClaim(key string, value interface{}) *JWTClaimsBuilder {
	b.claims[key] = value
	return b
}

// Build returns the claims map.
func (b *JWTClaimsBuilder) Build() map[string]interface{} {
	// Return a copy to prevent mutation
	result := make(map[string]interface{}, len(b.claims))
	for k, v := range b.claims {
		result[k] = v
	}
	return result
}
