package jwtauth

import "errors"

// Domain errors for JWT authentication failures.
// These sentinel errors allow the middleware to distinguish between different
// failure modes for appropriate HTTP status codes and audit logging.
var (
	// ErrInvalidSignature indicates that the JWT signature does not match any key
	// in the configured JWKS. The token may have been tampered with or signed by
	// an unknown issuer.
	ErrInvalidSignature = errors.New("jwt: invalid signature")

	// ErrTokenExpired indicates that the JWT exp claim is in the past.
	// Expiry is always enforced regardless of verification mode (signed or unsigned).
	ErrTokenExpired = errors.New("jwt: token expired")

	// ErrAudienceMismatch indicates that the JWT aud claim does not contain the
	// expected audience configured in authentication.jwt.expected_audience.
	ErrAudienceMismatch = errors.New("jwt: audience mismatch")

	// ErrIssuerMismatch indicates that the JWT iss claim does not match the
	// expected issuer configured in authentication.jwt.expected_issuer.
	ErrIssuerMismatch = errors.New("jwt: issuer mismatch")

	// ErrClaimExtraction indicates that the CEL expression for claim extraction
	// failed to produce a valid result. For the principal expression, this is a
	// hard failure. For optional profile expressions, this results in nil values.
	ErrClaimExtraction = errors.New("jwt: claim extraction failed")

	// ErrMalformedToken indicates that the JWT value is not a valid JWT format
	// (not a properly encoded three-part token).
	ErrMalformedToken = errors.New("jwt: malformed token")

	// ErrMissingExpiry indicates that the JWT has no exp claim. Expiry is mandatory
	// for security — JWTs without an expiration are rejected.
	ErrMissingExpiry = errors.New("jwt: missing exp claim")

	// ErrJWKSUnavailable indicates that the JWKS endpoint could not be reached or
	// returned an error. This is distinct from signature verification failure and
	// maps to HTTP 503 Service Unavailable rather than 401 Unauthorized.
	ErrJWKSUnavailable = errors.New("jwt: JWKS endpoint unavailable")
)
