// Package tokenexchange provides domain types and services for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
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
		return nil, fmt.Errorf("expected_audience cannot be empty")
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
// 1. Fetch JWKS from upstream server via JWKSProvider
// 2. Parse JWT with comprehensive validation using library options:
//   - Signature verification (validates structure and signature)
//   - Issuer verification (matches configured upstream_oauth2.issuer)
//   - Audience verification (includes broker identifier)
//   - Expiration verification (with configurable clock skew tolerance)
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

	// Fetch JWKS and get the specific key
	keyset, err := v.jwksProvider.GetKeySet(ctx)
	if err != nil {
		return nil, NewServerErrorWithCause("failed to fetch JWKS for token validation", err)
	}

	// Parse with comprehensive validation using library options
	// This combines signature, issuer, audience, and expiration verification
	clockSkew := time.Duration(v.clockSkewSeconds) * time.Second
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithVerify(true),
		jwt.WithKeySet(keyset),
		jwt.WithValidate(true),
		jwt.WithIssuer(v.expectedIssuer),
		jwt.WithAudience(v.brokerAudience),
		jwt.WithAcceptableSkew(clockSkew),
	)
	if err != nil {
		// Map library errors to appropriate domain errors
		return nil, v.mapParseError(err, "subject_token")
	}

	return token, nil
}

// ValidateClientAssertion validates a client_assertion JWT from RFC 8693 token exchange request.
//
// Similar to ValidateSubjectToken but used for client authentication (RFC 7523).
//
// Validation steps:
// 1. Fetch JWKS from upstream server
// 2. Parse JWT with comprehensive validation using library options:
//   - Signature verification (validates structure and signature)
//   - Issuer verification (matches configured upstream_oauth2.issuer)
//   - Audience verification (includes broker identifier)
//   - Expiration verification (with configurable clock skew tolerance)
//
// Returns:
//   - The parsed and validated JWT token on success
//   - InvalidClientError if signature validation or claims verification fails
//   - ServerError if JWKS fetch fails
func (v *JWTValidator) ValidateClientAssertion(ctx context.Context, tokenString string) (jwt.Token, error) {
	if tokenString == "" {
		return nil, NewInvalidClientError("client_assertion is empty or missing")
	}

	// Fetch JWKS and get the specific key
	keyset, err := v.jwksProvider.GetKeySet(ctx)
	if err != nil {
		return nil, NewServerErrorWithCause("failed to fetch JWKS for client_assertion validation", err)
	}

	// Parse with comprehensive validation using library options
	// This combines signature, issuer, audience, and expiration verification
	clockSkew := time.Duration(v.clockSkewSeconds) * time.Second
	token, err := jwt.ParseString(
		tokenString,
		jwt.WithVerify(true),
		jwt.WithKeySet(keyset),
		jwt.WithValidate(true),
		jwt.WithIssuer(v.expectedIssuer),
		jwt.WithAudience(v.brokerAudience),
		jwt.WithAcceptableSkew(clockSkew),
	)
	if err != nil {
		// Map library errors to appropriate domain errors for client assertions
		return nil, v.mapClientAssertionParseError(err)
	}

	return token, nil
}

// mapParseError maps jwt.ParseString errors to appropriate domain errors for subject tokens.
//
// The lestrrat-go/jwx library returns various error types for different validation failures.
// This method maps them to the correct RFC 8693 error codes while preserving security.
//
// Per spec SR-005: Error messages do NOT expose token content (only metadata like issuer).
// The underlying library error is attached as a cause for internal logging only.
func (v *JWTValidator) mapParseError(err error, tokenType string) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, jwt.ParseError()):
		return NewInvalidRequestError(tokenType + " is malformed or signature verification failed").WithCause(err)
	case errors.Is(err, jwt.InvalidIssuerError()):
		return NewInvalidGrantError(
			fmt.Sprintf("%s issuer validation failed: expected iss=%q", tokenType, v.expectedIssuer),
		).WithCause(err)
	case errors.Is(err, jwt.InvalidAudienceError()):
		return NewInvalidGrantError(
			fmt.Sprintf("%s audience validation failed: expected aud=%q", tokenType, v.brokerAudience),
		).WithCause(err)
	case errors.Is(err, jwt.TokenExpiredError()):
		return NewInvalidGrantError(tokenType + " has expired").WithCause(err)
	case errors.Is(err, jwt.TokenNotYetValidError()):
		return NewInvalidGrantError(tokenType + " is not yet valid (nbf)").WithCause(err)
	default:
		return NewInvalidGrantError(tokenType + " validation failed").WithCause(err)
	}
}

// mapClientAssertionParseError maps jwt.ParseString errors to InvalidClientError for client assertions.
//
// Client assertion failures always result in InvalidClientError per RFC 7523.
// The underlying library error is attached as a cause for internal logging only.
func (v *JWTValidator) mapClientAssertionParseError(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, jwt.ParseError()):
		return NewInvalidClientError("client_assertion is malformed or signature verification failed").WithCause(err)
	case errors.Is(err, jwt.InvalidIssuerError()):
		return NewInvalidClientError(
			fmt.Sprintf("client_assertion issuer validation failed: expected iss=%q", v.expectedIssuer),
		).WithCause(err)
	case errors.Is(err, jwt.InvalidAudienceError()):
		return NewInvalidClientError(
			fmt.Sprintf("client_assertion audience validation failed: expected aud=%q", v.brokerAudience),
		).WithCause(err)
	case errors.Is(err, jwt.TokenExpiredError()):
		return NewInvalidClientError("client_assertion has expired").WithCause(err)
	case errors.Is(err, jwt.TokenNotYetValidError()):
		return NewInvalidClientError("client_assertion is not yet valid (nbf)").WithCause(err)
	default:
		return NewInvalidClientError("client_assertion validation failed").WithCause(err)
	}
}
