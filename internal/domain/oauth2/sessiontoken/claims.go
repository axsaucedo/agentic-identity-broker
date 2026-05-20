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

// CIMDMetadata carries the CIMD-resolved client metadata embedded in a session token.
type CIMDMetadata struct {
	ClientID     string   `json:"client_id"`
	ClientName   string   `json:"client_name,omitempty"`
	LogoURI      string   `json:"logo_uri,omitempty"`
	RedirectURIs []string `json:"redirect_uris"`
}

// AuthorizationSessionClaims carries the full authorization context in a JWE token.
// It replaces the DB-backed AuthorizationSession: the same tamper-proof, expiring,
// principal-bound properties are achieved by sealing the claims in a JWE.
// OriginalURL is the raw authorize request URL; scope and redirect_uri are parsed from
// it at the consumption site rather than duplicated as flat fields.
type AuthorizationSessionClaims struct {
	AgentID      id.AgentID    `json:"agent_id"`
	Principal    id.Principal  `json:"principal"`
	OriginalURL  string        `json:"original_url"`
	CIMDMetadata *CIMDMetadata `json:"cimd_metadata,omitempty"`
	IssuedAt     time.Time     `json:"iat"`
	ExpiresAt    time.Time     `json:"exp"`
}

// IsExpired reports whether the session token TTL has elapsed.
func (c *AuthorizationSessionClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// ToResult maps internal claims to the port-local DTO.
func (c *AuthorizationSessionClaims) ToResult() *ports.AuthorizationSession {
	r := &ports.AuthorizationSession{
		AgentID:     c.AgentID,
		Principal:   c.Principal,
		OriginalURL: c.OriginalURL,
	}
	if c.CIMDMetadata != nil {
		r.CIMDMetadata = &ports.SessionCIMDMetadata{
			ClientID:     c.CIMDMetadata.ClientID,
			ClientName:   c.CIMDMetadata.ClientName,
			LogoURI:      c.CIMDMetadata.LogoURI,
			RedirectURIs: c.CIMDMetadata.RedirectURIs,
		}
	}
	return r
}

// NewAuthorizationSessionClaims initialises claims with IssuedAt and ExpiresAt set from now.
func NewAuthorizationSessionClaims(
	agentID id.AgentID,
	principal id.Principal,
	originalURL string,
	cimdMetadata *CIMDMetadata,
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
		OriginalURL:  originalURL,
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

// ErrSessionServiceNotConfigured indicates the session token service is misconfigured
// (nil JWE token service). This is a server-side error, not a client token error.
var ErrSessionServiceNotConfigured = errors.New("authorization session service not configured")
