// Package tokenexchange provides domain types and services for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"context"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// JWKSProvider defines the interface for JSON Web Key Set (JWKS) operations.
// This interface is defined locally in the domain to avoid circular imports with the ports package.
// Any object implementing GetKeySet and GetKey methods can be used (structural typing in Go).
//
// Typically implemented by: internal/adapters/jwks/adapter.Adapter (JWKSPort)
type JWKSProvider interface {
	// GetKeySet returns the cached JWKS key set from upstream OAuth2 server.
	GetKeySet(ctx context.Context) (jwk.Set, error)

	// GetKey retrieves a specific key from the JWKS by key ID (kid).
	GetKey(ctx context.Context, kid string) (jwk.Key, error)
}

// JWTValidator provides JWT validation for token exchange.
// It validates token signatures, issuer, audience, and expiration using the JWKSProvider adapter.
//
// This service implements claim validation using the lestrrat-go/jwx/v3 library (Principle III - Library-First Security).
// No custom cryptography is used; all validation relies on vetted libraries.
//
// The validator is stateless and thread-safe for concurrent calls.
type JWTValidator struct {
	jwksProvider     JWKSProvider
	expectedIssuer   string
	brokerAudience   string
	clockSkewSeconds int64
}

// NewJWTValidator creates a new JWT validator.
//
// Parameters:
//   - jwksProvider: Provider for fetching JWKS from upstream server (e.g., JWKSAdapter)
//     Can be any object implementing the JWKSProvider interface
//   - expectedIssuer: The expected issuer in token claims (e.g., "https://auth.example.com")
//   - brokerAudience: The broker's identifier that should be in token audience claims
//   - clockSkewSeconds: Clock skew tolerance in seconds for expiration checking (e.g., 60)
//     Must be between 0 and MaxClockSkewTolerance (300)
//
// Returns error if parameters are invalid.
func NewJWTValidator(
	jwksProvider JWKSProvider,
	expectedIssuer string,
	brokerAudience string,
	clockSkewSeconds int64,
) (*JWTValidator, error) {
	if jwksProvider == nil {
		return nil, fmt.Errorf("jwks_provider cannot be nil")
	}
	if expectedIssuer == "" {
		return nil, fmt.Errorf("expected_issuer cannot be empty")
	}
	if brokerAudience == "" {
		return nil, fmt.Errorf("broker_audience cannot be empty")
	}
	if clockSkewSeconds < 0 {
		return nil, fmt.Errorf("clock_skew_seconds cannot be negative")
	}
	if clockSkewSeconds > MaxClockSkewTolerance {
		return nil, fmt.Errorf("clock_skew_seconds (%d) cannot exceed max tolerance (%d)", clockSkewSeconds, MaxClockSkewTolerance)
	}

	return &JWTValidator{
		jwksProvider:     jwksProvider,
		expectedIssuer:   expectedIssuer,
		brokerAudience:   brokerAudience,
		clockSkewSeconds: clockSkewSeconds,
	}, nil
}

// ValidateSubjectToken validates a subject_token JWT from RFC 8693 token exchange request.
//
// Validation steps:
// 1. Parse JWT structure (no signature check yet)
// 2. Fetch JWKS from upstream server via JWKSProvider
// 3. Find key by kid from token header
// 4. Validate signature against key
// 5. Verify issuer matches configured upstream_oauth2.issuer
// 6. Verify audience includes broker identifier
// 7. Verify token not expired (with configurable clock skew tolerance)
//
// Per spec SR-006: Validation failure always denies the request.
// Per spec SR-005: Error messages do NOT expose token content (only metadata like issuer).
//
// Returns:
//   - The parsed and validated JWT token on success
//   - InvalidClientError if signature validation fails (per spec SR-001)
//   - InvalidGrantError if token is expired or malformed
//   - InvalidGrantError if issuer/audience verification fails
//   - ServerError if JWKS fetch fails
func (v *JWTValidator) ValidateSubjectToken(ctx context.Context, tokenString string) (jwt.Token, error) {
	if tokenString == "" {
		return nil, NewInvalidGrantError("subject_token is empty or missing")
	}

	// Parse JWT without verification (just structure check)
	if _, err := jwt.ParseString(tokenString, jwt.WithVerify(false)); err != nil {
		return nil, NewInvalidRequestError("subject_token is malformed")
	}

	// Fetch JWKS and get the specific key
	keyset, err := v.jwksProvider.GetKeySet(ctx)
	if err != nil {
		return nil, NewServerErrorWithCause("failed to fetch JWKS for token validation", err)
	}

	// Verify signature against JWKS
	if err := v.VerifySignature(ctx, tokenString, keyset); err != nil {
		return nil, err
	}

	// Re-parse with verification now that we know signature is valid
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithVerify(true),
		jwt.WithKeySet(keyset),
	)
	if err != nil {
		return nil, NewInvalidRequestError("subject_token signature verification failed")
	}

	// Verify issuer
	if err := v.VerifyIssuer(token, v.expectedIssuer); err != nil {
		return nil, err
	}

	// Verify audience
	if err := v.VerifyAudience(token, v.brokerAudience); err != nil {
		return nil, err
	}

	// Verify expiration with clock skew
	if err := v.VerifyExpiration(token, time.Duration(v.clockSkewSeconds)*time.Second); err != nil {
		return nil, err
	}

	return token, nil
}

// ValidateClientAssertion validates a client_assertion JWT from RFC 8693 token exchange request.
//
// Similar to ValidateSubjectToken but used for client authentication (RFC 7523).
//
// Validation steps:
// 1. Parse JWT structure
// 2. Fetch JWKS from upstream server
// 3. Find key by kid
// 4. Validate signature against key
// 5. Verify issuer matches configured upstream_oauth2.issuer
// 6. Verify audience includes broker identifier
// 7. Verify token not expired
//
// Returns:
//   - The parsed and validated JWT token on success
//   - InvalidClientError if signature validation or claims verification fails
//   - ServerError if JWKS fetch fails
func (v *JWTValidator) ValidateClientAssertion(ctx context.Context, tokenString string) (jwt.Token, error) {
	if tokenString == "" {
		return nil, NewInvalidClientError("client_assertion is empty or missing")
	}

	// Parse JWT without verification (just structure check)
	if _, err := jwt.ParseString(tokenString, jwt.WithVerify(false)); err != nil {
		return nil, NewInvalidClientError("client_assertion is malformed")
	}

	// Fetch JWKS and get the specific key
	keyset, err := v.jwksProvider.GetKeySet(ctx)
	if err != nil {
		return nil, NewServerErrorWithCause("failed to fetch JWKS for client_assertion validation", err)
	}

	// Verify signature against JWKS
	if err := v.VerifySignature(ctx, tokenString, keyset); err != nil {
		return nil, err
	}

	// Re-parse with verification
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithVerify(true),
		jwt.WithKeySet(keyset),
	)
	if err != nil {
		return nil, NewInvalidClientError("client_assertion signature verification failed")
	}

	// Verify issuer
	if err := v.VerifyIssuer(token, v.expectedIssuer); err != nil {
		return nil, NewInvalidClientError("client_assertion issuer validation failed")
	}

	// Verify audience
	if err := v.VerifyAudience(token, v.brokerAudience); err != nil {
		return nil, NewInvalidClientError("client_assertion audience validation failed")
	}

	// Verify expiration with clock skew
	if err := v.VerifyExpiration(token, time.Duration(v.clockSkewSeconds)*time.Second); err != nil {
		return nil, NewInvalidClientError("client_assertion has expired")
	}

	return token, nil
}

// VerifySignature validates a JWT signature against a JWKS key set.
//
// This method performs cryptographic signature verification using the lestrrat-go/jwx/v3 library.
// The lestrrat-go/jwx library automatically finds the correct key by kid from the token header
// and validates the signature against the key set.
//
// Per Principle III (Library-First Security): Uses only lestrrat-go/jwx/v3 for cryptography.
//
// Returns error (with "InvalidClient" code) if:
//   - Token is not a valid JWT structure
//   - kid not found in key set
//   - Signature verification fails
//   - No public key component found
func (v *JWTValidator) VerifySignature(ctx context.Context, tokenString string, keyset interface{}) error {
	// Parse JWS to validate structure
	parsedJWS, err := jws.ParseString(tokenString)
	if err != nil {
		return NewInvalidClientError("token is not a valid JWT")
	}

	if len(parsedJWS.Signatures()) == 0 {
		return NewInvalidClientError("token has no signatures")
	}

	// Use lestrrat-go to verify signature
	// This performs the cryptographic verification using the JWKS
	// The library automatically finds the key by kid from the token header
	payload, err := jws.Verify([]byte(tokenString), jws.WithKeySet(keyset.(jwk.Set)))
	if err != nil {
		// Per spec SR-005: Don't expose token content in error message
		return NewInvalidClientError("signature verification failed against JWKS")
	}

	// Ensure we got a payload (sanity check)
	if len(payload) == 0 {
		return NewInvalidClientError("signature verification failed: empty payload")
	}

	return nil
}

// VerifyIssuer checks that a token's issuer matches the expected value.
//
// The token's "iss" claim must exactly match the expectedIssuer.
// This prevents accepting tokens from untrusted issuers.
//
// Returns error if issuer doesn't match (for subject_token: InvalidGrantError, for client_assertion: InvalidClientError).
func (v *JWTValidator) VerifyIssuer(token jwt.Token, expectedIssuer string) error {
	if token == nil {
		return NewInvalidGrantError("token is nil")
	}

	issuer, _ := token.Issuer()
	if issuer != expectedIssuer {
		// Per spec SR-005: Only include metadata (issuer value), not full token
		return NewInvalidGrantError(fmt.Sprintf("issuer mismatch: expected '%s', got '%s'", expectedIssuer, issuer))
	}

	return nil
}

// VerifyAudience checks that a token's audience includes the broker identifier.
//
// The token's "aud" claim may be a single string or array of strings.
// At least one entry must match brokerAudience.
// This ensures the token was intended for this broker.
//
// Returns error if broker audience not found in token.
func (v *JWTValidator) VerifyAudience(token jwt.Token, brokerAudience string) error {
	if token == nil {
		return NewInvalidGrantError("token is nil")
	}

	audiences, _ := token.Audience()
	if len(audiences) == 0 {
		return NewInvalidGrantError("token missing 'aud' (audience) claim")
	}

	// Check if broker audience is in the list
	for _, aud := range audiences {
		if aud == brokerAudience {
			return nil
		}
	}

	return NewInvalidGrantError(fmt.Sprintf("token audience does not include broker identifier '%s'", brokerAudience))
}

// VerifyExpiration checks that a token has not expired.
//
// The token's "exp" claim is compared against current time.
// A clockSkew tolerance is applied to allow for minor time synchronization differences.
// This prevents accepting tokens outside their validity period.
//
// Per spec FR-042: Clock skew tolerance is configurable, default 60 seconds, max 300 seconds.
//
// Returns error if token is expired (beyond the clock skew tolerance).
func (v *JWTValidator) VerifyExpiration(token jwt.Token, clockSkew time.Duration) error {
	if token == nil {
		return NewInvalidGrantError("token is nil")
	}

	expiration, _ := token.Expiration()
	if expiration.IsZero() {
		return NewInvalidGrantError("token missing 'exp' (expiration) claim")
	}

	now := time.Now().UTC()
	adjustedExpiration := expiration.Add(clockSkew)

	if now.After(adjustedExpiration) {
		return NewInvalidGrantError("token has expired")
	}

	return nil
}
