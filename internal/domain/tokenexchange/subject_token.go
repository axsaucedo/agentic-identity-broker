// Package tokenexchange provides domain types and errors for RFC 8693 OAuth 2.0 Token Exchange.
package tokenexchange

import (
	"fmt"
	"time"
)

// SubjectToken represents a parsed JWT used as the subject token in token exchange.
// It is an immutable value object containing the claims extracted from a subject_token JWT.
//
// Per RFC 8693 Section 2.1:
// - The subject_token is the token being exchanged (usually from upstream OAuth2 server)
// - Contains user principal and agent identifier extracted via CEL expressions
// - Is validated via JWKS signature verification (per SR-001, SR-006)
//
// Domain invariants:
// - Principal must not be empty (end-user identifier, extracted via CEL)
// - AgentClientID must not be empty (agent identifier, extracted via CEL)
// - ExpiresAt should be in the future (not expired)
// - Token string itself is never stored (per SR-005, only claims)
//
// Security note: No token string is stored to prevent accidental exposure in logs/responses.
// The validation happens during JWT parsing; this object only holds the validated claims.
//
// The principal and agent_client_id are extracted via CEL expressions during JWT validation.
// Per spec, default expressions are:
// - Principal: "subject_token.sub" (sub claim)
// - AgentClientID: "subject_token.azp" (azp/authorized party claim)
type SubjectToken struct {
	// Principal is the end-user identifier extracted from the subject_token.
	// Extracted via configurable CEL expression (default: subject_token.sub).
	// REQUIRED - identifies which user is requesting the token exchange.
	// Format: email, username, UUID, or other unique identifier per upstream server.
	// Example: "user@example.com", "user-123"
	Principal string

	// AgentClientID is the agent identifier extracted from the subject_token.
	// Extracted via configurable CEL expression (default: subject_token.azp).
	// REQUIRED - identifies which agent is requesting the token on behalf of the user.
	// Format: client_id or agent identifier per upstream server.
	// Example: "agent-123", "my-ai-agent"
	AgentClientID string

	// Issuer is the 'iss' claim identifying the upstream OAuth2 server.
	// Used to verify the token came from the expected issuer.
	// Expected to match the configured upstream_oauth2.issuer.
	Issuer string

	// Subject is the 'sub' claim from the JWT.
	// May differ from Principal if CEL expression extracts from different claim.
	// Stored for debugging and troubleshooting token validation.
	Subject string

	// AuthorizedParty is the 'azp' claim from the JWT.
	// May differ from AgentClientID if CEL expression extracts from different claim.
	// Stored for debugging and troubleshooting token validation.
	AuthorizedParty string

	// ExpiresAt is the 'exp' claim as a Unix timestamp.
	// Indicates when the token becomes invalid.
	// Validated before this object is created to ensure token is not expired.
	ExpiresAt int64

	// IssuedAt is the 'iat' claim as a Unix timestamp.
	// Indicates when the token was issued.
	IssuedAt int64

	// Scope is the 'scope' claim if present.
	// May be empty or space-separated list.
	// Represents what scopes the token is valid for.
	Scope string

	// CustomClaims stores any additional custom claims from the JWT.
	// Allows access to non-standard claims without changing fixed fields.
	// Keyed by claim name, values are arbitrary JSON-compatible types.
	// Examples: "org": "acme", "team": "platform"
	// Per spec, these are available for CEL evaluation.
	CustomClaims map[string]interface{}
}

// NewSubjectToken creates a new SubjectToken from parsed JWT claims.
// The principal and agentClientID are extracted via CEL during JWT validation.
// All parameters are immutable after creation.
// The function does NOT validate the values - validation happens at JWT parsing time.
func NewSubjectToken(
	principal, agentClientID string,
	issuer, subject, authorizedParty string,
	expiresAt, issuedAt int64,
	scope string,
	customClaims map[string]interface{},
) *SubjectToken {
	if customClaims == nil {
		customClaims = make(map[string]interface{})
	}
	return &SubjectToken{
		Principal:       principal,
		AgentClientID:   agentClientID,
		Issuer:          issuer,
		Subject:         subject,
		AuthorizedParty: authorizedParty,
		ExpiresAt:       expiresAt,
		IssuedAt:        issuedAt,
		Scope:           scope,
		CustomClaims:    customClaims,
	}
}

// Validate checks if the subject token claims are valid.
// Returns error if required fields are empty or invalid.
//
// Validation rules:
// 1. Principal must not be empty (end-user identifier required)
// 2. AgentClientID must not be empty (agent identifier required)
// 3. ExpiresAt should be greater than zero
// 4. IssuedAt should be greater than zero
//
// Note: Expiration checking against current time is NOT done here.
// That's handled by JWT validation layer before this object is created.
func (st *SubjectToken) Validate() error {
	if st.Principal == "" {
		return fmt.Errorf("principal is required for subject token")
	}

	if st.AgentClientID == "" {
		return fmt.Errorf("agent_client_id is required for subject token")
	}

	if st.ExpiresAt <= 0 {
		return fmt.Errorf("expiresAt (exp claim) must be a valid Unix timestamp")
	}

	if st.IssuedAt <= 0 {
		return fmt.Errorf("issuedAt (iat claim) must be a valid Unix timestamp")
	}

	return nil
}

// String returns a string representation suitable for logging.
// Does NOT include any token content or sensitive claims.
// Shows only metadata: principal, agent, issuer.
// No token values are exposed per SR-005.
func (st *SubjectToken) String() string {
	return fmt.Sprintf("SubjectToken{principal=%q, agentClientID=%q, issuer=%q, expiresAt=%d}",
		st.Principal, st.AgentClientID, st.Issuer, st.ExpiresAt)
}

// IsExpired returns true if the token has expired.
// Compares ExpiresAt with the provided current time (allows for testing with fixed time).
// Returns true if ExpiresAt <= now.
func (st *SubjectToken) IsExpired(now time.Time) bool {
	return st.ExpiresAt <= now.Unix()
}

// GetCustomClaim retrieves a custom claim value by name.
// Returns nil if the claim doesn't exist.
// Useful for CEL evaluation accessing custom claims.
func (st *SubjectToken) GetCustomClaim(name string) interface{} {
	if st.CustomClaims == nil {
		return nil
	}
	return st.CustomClaims[name]
}

// HasCustomClaim returns true if a custom claim exists.
func (st *SubjectToken) HasCustomClaim(name string) bool {
	if st.CustomClaims == nil {
		return false
	}
	_, ok := st.CustomClaims[name]
	return ok
}

// GetAllClaims returns a map of all claims as if they were a complete JWT claims object.
// Useful for CEL evaluation which may access claims by field name.
// Returns a new map, so modifications don't affect the object.
// Includes both standard JWT claims and custom claims.
func (st *SubjectToken) GetAllClaims() map[string]interface{} {
	claims := map[string]interface{}{
		"sub":   st.Subject,
		"azp":   st.AuthorizedParty,
		"iss":   st.Issuer,
		"exp":   st.ExpiresAt,
		"iat":   st.IssuedAt,
		"scope": st.Scope,
	}

	// Add custom claims
	for k, v := range st.CustomClaims {
		claims[k] = v
	}

	return claims
}

// GetToken returns a representation of the token suitable for CEL context.
// This is used by CEL evaluator to provide subject_token claims in expression evaluation.
// Returns the claims as a map with string keys.
func (st *SubjectToken) GetToken() map[string]interface{} {
	return st.GetAllClaims()
}

// GetPrincipal returns the principal (end-user identifier).
func (st *SubjectToken) GetPrincipal() string {
	return st.Principal
}

// GetAgentClientID returns the agent client ID.
func (st *SubjectToken) GetAgentClientID() string {
	return st.AgentClientID
}

// GetIssuer returns the issuer.
func (st *SubjectToken) GetIssuer() string {
	return st.Issuer
}

// GetSubject returns the subject claim (raw 'sub' claim).
func (st *SubjectToken) GetSubject() string {
	return st.Subject
}

// GetAuthorizedParty returns the authorized party claim (raw 'azp' claim).
func (st *SubjectToken) GetAuthorizedParty() string {
	return st.AuthorizedParty
}

// GetScope returns the scope claim.
func (st *SubjectToken) GetScope() string {
	return st.Scope
}

// GetExpiresAt returns the expiration time as Unix timestamp.
func (st *SubjectToken) GetExpiresAt() int64 {
	return st.ExpiresAt
}

// GetIssuedAt returns the issuance time as Unix timestamp.
func (st *SubjectToken) GetIssuedAt() int64 {
	return st.IssuedAt
}
