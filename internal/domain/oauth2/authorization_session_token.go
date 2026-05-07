package oauth2

import (
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
)

// authorizationSessionTokenTTL is the lifetime of an AuthorizationSessionClaims JWE token.
const authorizationSessionTokenTTL = 10 * time.Minute

// AuthorizationSessionClaims carries the full authorization context in a JWE token.
// It replaces the DB-backed AuthorizationSession: the same tamper-proof, expiring,
// principal-bound properties are achieved by sealing the claims in a JWE.
// OriginalURL is the raw authorize request URL; scope and redirect_uri are parsed from
// it at the consumption site rather than duplicated as flat fields.
type AuthorizationSessionClaims struct {
	AgentID      id.AgentID                     `json:"agent_id"`
	Principal    id.Principal                   `json:"principal"`
	OriginalURL  string                         `json:"original_url"`
	CIMDMetadata *cimd.ClientIDMetadataDocument `json:"cimd_metadata,omitempty"`
	IssuedAt     time.Time                      `json:"iat"`
	ExpiresAt    time.Time                      `json:"exp"`
}

// IsExpired reports whether the session token TTL has elapsed.
func (c *AuthorizationSessionClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// NewAuthorizationSessionClaims initialises claims with IssuedAt and ExpiresAt set from now.
func NewAuthorizationSessionClaims(
	agentID id.AgentID,
	principal id.Principal,
	originalURL string,
	cimdMetadata *cimd.ClientIDMetadataDocument,
) *AuthorizationSessionClaims {
	now := time.Now()
	return &AuthorizationSessionClaims{
		AgentID:      agentID,
		Principal:    principal,
		OriginalURL:  originalURL,
		CIMDMetadata: cimdMetadata,
		IssuedAt:     now,
		ExpiresAt:    now.Add(authorizationSessionTokenTTL),
	}
}

// CreateAuthorizationSessionToken seals claims into a compact JWE string.
func (s *Service) CreateAuthorizationSessionToken(claims *AuthorizationSessionClaims) (string, error) {
	if s.jweTokenService == nil {
		return "", fmt.Errorf("jweTokenService not configured")
	}
	return s.jweTokenService.Encrypt(claims)
}
