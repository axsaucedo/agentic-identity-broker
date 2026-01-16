// Package tokenexchange provides domain types and errors for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"fmt"
	"time"
)

// ClientAssertion represents a parsed JWT used for client authentication in token exchange.
// It is an immutable value object containing the claims extracted from a client_assertion JWT.
//
// Per RFC 7523 and RFC 8693:
// - The client_assertion is a JWT signed by the gateway/client
// - Contains claims identifying the gateway and authorizing it for token exchange
// - Is validated via JWKS signature verification (per SR-001, SR-006)
//
// Domain invariants:
// - Subject (sub) must not be empty (identifies the gateway)
// - Audiences must contain at least one value
// - ExpiresAt should be in the future (not expired)
// - Token string itself is never stored (per SR-005, only claims)
//
// Security note: No token string is stored to prevent accidental exposure in logs/responses.
// The validation happens during JWT parsing; this object only holds the validated claims.
type ClientAssertion struct {
	// Subject is the 'sub' claim identifying the gateway/client.
	// REQUIRED - identifies which gateway made the token exchange request.
	// Format typically: gateway identifier, service account, client_id, etc.
	Subject string

	// Audiences is the 'aud' claim as a slice of audience values.
	// Per RFC 7523, indicates which resource the assertion is intended for.
	// At least one value must match this broker's identifier.
	// Typical values: ["https://auth.example.com", "token-exchange-broker"]
	Audiences []string

	// ExpiresAt is the 'exp' claim as a Unix timestamp.
	// Indicates when the assertion becomes invalid.
	// Typically 5-10 minutes after issuance.
	ExpiresAt int64

	// IssuedAt is the 'iat' claim as a Unix timestamp.
	// Indicates when the assertion was issued.
	// Used for clock skew validation.
	IssuedAt int64

	// Issuer is the 'iss' claim identifying the upstream OAuth2 server.
	// Expected to match the configured upstream issuer.
	Issuer string

	// Scopes is the 'scope' claim if present.
	// May be empty or space-separated list.
	// Represents what scopes the client is requesting.
	Scopes string

	// CustomClaims stores any additional custom claims from the JWT.
	// Allows extension without changing the fixed claim fields.
	// Keyed by claim name, values are arbitrary JSON-compatible types.
	// Examples: "custom_field": "value", "role": "admin"
	// Per spec, these are available for CEL evaluation.
	CustomClaims map[string]interface{}
}

// NewClientAssertion creates a new ClientAssertion from parsed JWT claims.
// All parameters are required and immutable after creation.
// The function does NOT validate the values - validation happens at JWT parsing time.
func NewClientAssertion(
	subject string,
	audiences []string,
	expiresAt, issuedAt int64,
	issuer, scopes string,
	customClaims map[string]interface{},
) *ClientAssertion {
	if customClaims == nil {
		customClaims = make(map[string]interface{})
	}
	return &ClientAssertion{
		Subject:      subject,
		Audiences:    audiences,
		ExpiresAt:    expiresAt,
		IssuedAt:     issuedAt,
		Issuer:       issuer,
		Scopes:       scopes,
		CustomClaims: customClaims,
	}
}

// Validate checks if the client assertion claims are valid.
// Returns error if required fields are empty or invalid.
//
// Validation rules:
// 1. Subject must not be empty (gateway identifier required)
// 2. Audiences must not be empty (at least one audience must be present)
// 3. ExpiresAt should be greater than zero
// 4. IssuedAt should be greater than zero
//
// Note: Expiration checking against current time is NOT done here.
// That's handled by JWT validation layer before this object is created.
func (ca *ClientAssertion) Validate() error {
	if ca.Subject == "" {
		return fmt.Errorf("subject (sub claim) is required for client assertion")
	}

	if len(ca.Audiences) == 0 {
		return fmt.Errorf("audiences (aud claim) must contain at least one value")
	}

	if ca.ExpiresAt <= 0 {
		return fmt.Errorf("expiresAt (exp claim) must be a valid Unix timestamp")
	}

	if ca.IssuedAt <= 0 {
		return fmt.Errorf("issuedAt (iat claim) must be a valid Unix timestamp")
	}

	return nil
}

// String returns a string representation suitable for logging.
// Does NOT include any token content or sensitive claims.
// Shows only metadata: subject, audience count, expiration.
func (ca *ClientAssertion) String() string {
	numAudiences := 0
	if ca.Audiences != nil {
		numAudiences = len(ca.Audiences)
	}
	return fmt.Sprintf("ClientAssertion{subject=%q, audienceCount=%d, issuer=%q, expiresAt=%d}",
		ca.Subject, numAudiences, ca.Issuer, ca.ExpiresAt)
}

// HasAudience returns true if the assertion includes the specified audience.
// Audience comparison is case-sensitive.
// Used during JWT validation to check if broker is in intended audience.
func (ca *ClientAssertion) HasAudience(audience string) bool {
	for _, aud := range ca.Audiences {
		if aud == audience {
			return true
		}
	}
	return false
}

// IsExpired returns true if the assertion has expired.
// Compares ExpiresAt with the provided current time (allows for testing with fixed time).
// Returns true if ExpiresAt <= now.
func (ca *ClientAssertion) IsExpired(now time.Time) bool {
	return ca.ExpiresAt <= now.Unix()
}

// GetCustomClaim retrieves a custom claim value by name.
// Returns nil if the claim doesn't exist.
// Useful for CEL evaluation accessing custom claims.
func (ca *ClientAssertion) GetCustomClaim(name string) interface{} {
	if ca.CustomClaims == nil {
		return nil
	}
	return ca.CustomClaims[name]
}

// HasCustomClaim returns true if a custom claim exists.
func (ca *ClientAssertion) HasCustomClaim(name string) bool {
	if ca.CustomClaims == nil {
		return false
	}
	_, ok := ca.CustomClaims[name]
	return ok
}

// GetAllClaims returns a map of all standard claims as if they were a complete JWT claims object.
// Useful for CEL evaluation which may access claims by field name.
// Returns a new map, so modifications don't affect the object.
func (ca *ClientAssertion) GetAllClaims() map[string]interface{} {
	claims := map[string]interface{}{
		"sub":   ca.Subject,
		"aud":   ca.Audiences,
		"exp":   ca.ExpiresAt,
		"iat":   ca.IssuedAt,
		"iss":   ca.Issuer,
		"scope": ca.Scopes,
	}

	// Add custom claims
	for k, v := range ca.CustomClaims {
		claims[k] = v
	}

	return claims
}

// GetAssertion returns a representation of the assertion suitable for CEL context.
// This is used by CEL evaluator to provide client_assertion claims in expression evaluation.
// Returns the claims as a map with string keys.
func (ca *ClientAssertion) GetAssertion() map[string]interface{} {
	return ca.GetAllClaims()
}

// GetSubject returns the subject (gateway identifier).
func (ca *ClientAssertion) GetSubject() string {
	return ca.Subject
}

// GetAudiences returns the audiences.
func (ca *ClientAssertion) GetAudiences() []string {
	return ca.Audiences
}

// GetIssuer returns the issuer.
func (ca *ClientAssertion) GetIssuer() string {
	return ca.Issuer
}

// GetScopes returns the scopes claim.
func (ca *ClientAssertion) GetScopes() string {
	return ca.Scopes
}

// GetExpiresAt returns the expiration time as Unix timestamp.
func (ca *ClientAssertion) GetExpiresAt() int64 {
	return ca.ExpiresAt
}

// GetIssuedAt returns the issuance time as Unix timestamp.
func (ca *ClientAssertion) GetIssuedAt() int64 {
	return ca.IssuedAt
}
