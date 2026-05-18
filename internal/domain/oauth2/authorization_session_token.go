package oauth2

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
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
	if u, err := url.Parse(originalURL); err != nil || (u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https") {
		return nil, errors.New("originalURL must be a valid relative or http(s) URL")
	}
	now := time.Now()
	return &AuthorizationSessionClaims{
		AgentID:      agentID,
		Principal:    principal,
		OriginalURL:  originalURL,
		CIMDMetadata: cimdMetadata,
		IssuedAt:     now,
		ExpiresAt:    now.Add(authorizationSessionTokenTTL),
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

// ValidateAuthorizationSessionToken decrypts a JWE session token and validates
// expiry, agent binding, and principal binding. Returns the validated claims or
// a typed error indicating the failure reason.
func (s *AuthorizationService) ValidateAuthorizationSessionToken(token string, agentID id.AgentID, principal id.Principal) (*AuthorizationSessionClaims, error) {
	if s.jweTokenService == nil {
		return nil, fmt.Errorf("jweTokenService not configured")
	}
	var claims AuthorizationSessionClaims
	if err := s.jweTokenService.DecryptAndValidate(token, &claims); err != nil {
		if errors.Is(err, jwe.ErrExpired) {
			return nil, ErrSessionExpired
		}
		return nil, ErrSessionInvalidToken
	}
	if claims.AgentID != agentID {
		return nil, ErrSessionAgentMismatch
	}
	if claims.Principal != principal {
		return nil, ErrSessionPrincipalMismatch
	}
	return &claims, nil
}

// CreateAuthorizationSessionToken seals claims into a compact JWE string.
func (s *AuthorizationService) CreateAuthorizationSessionToken(claims *AuthorizationSessionClaims) (string, error) {
	if s.jweTokenService == nil {
		return "", fmt.Errorf("jweTokenService not configured")
	}
	return s.jweTokenService.Encrypt(claims)
}
