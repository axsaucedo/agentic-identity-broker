package sessiontoken

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

const ttl = 10 * time.Minute

// AuthorizationSessionClaims carries the full authorization context in a JWE token.
// OriginalURL is the raw authorize request URL; scope and redirect_uri are parsed from
// it at the consumption site rather than duplicated as flat fields.
type AuthorizationSessionClaims struct {
	AgentID      id.AgentID                 `json:"agent_id"`
	Principal    id.Principal               `json:"principal"`
	OriginalURL  string                     `json:"original_url"`
	CIMDMetadata *ports.SessionCIMDMetadata `json:"cimd_metadata,omitempty"`
	IssuedAt     time.Time                  `json:"iat"`
	ExpiresAt    time.Time                  `json:"exp"`
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
	cimdMetadata *ports.SessionCIMDMetadata,
) (*AuthorizationSessionClaims, error) {
	if agentID.IsZero() {
		return nil, errors.New("agentID must not be zero")
	}
	if principal.IsZero() {
		return nil, errors.New("principal must not be zero")
	}
	if originalURL == "" {
		return nil, errors.New("originalURL must not be empty")
	}
	normalized := strings.ReplaceAll(originalURL, "\\", "/")
	if u, err := url.Parse(normalized); err != nil ||
		(u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") ||
		(u.Scheme == "" && u.Host != "") {
		return nil, errors.New("originalURL must be a valid relative or http(s) URL")
	}
	now := time.Now()
	return &AuthorizationSessionClaims{
		AgentID:      agentID,
		Principal:    principal,
		OriginalURL:  normalized,
		CIMDMetadata: cimdMetadata,
		IssuedAt:     now,
		ExpiresAt:    now.Add(ttl),
	}, nil
}

// ErrSessionExpired indicates the authorization session token TTL has elapsed.
var ErrSessionExpired = errors.New("authorization session expired")

// ErrSessionInvalidToken indicates the token could not be decrypted or unmarshalled.
var ErrSessionInvalidToken = errors.New("authorization session token invalid")

// ErrSessionAgentMismatch indicates the token's agent_id does not match the requested agent.
var ErrSessionAgentMismatch = errors.New("authorization session does not match requested agent")

// ErrSessionPrincipalMismatch indicates the token's principal does not match the authenticated user.
var ErrSessionPrincipalMismatch = errors.New("authorization session does not belong to this user")
