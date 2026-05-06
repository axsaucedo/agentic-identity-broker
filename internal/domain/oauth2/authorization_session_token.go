package oauth2

import (
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server/cimd"
)

// authorizationSessionTokenTTL is the lifetime of an AuthorizationSessionClaims JWE token.
const authorizationSessionTokenTTL = 10 * time.Minute

// AuthorizationSessionClaims carries the full authorization context in a JWE token.
// It replaces the DB-backed AuthorizationSession: the same tamper-proof, expiring,
// principal-bound properties are achieved by sealing the claims in a JWE.
type AuthorizationSessionClaims struct {
	AgentID             id.AgentID             `json:"agent_id"`
	Principal           id.Principal           `json:"principal"`
	ClientID            string                 `json:"client_id"`
	OriginalURL         string                 `json:"original_url"`
	RedirectURI         string                 `json:"redirect_uri"`
	Scope               string                 `json:"scope"`
	State               string                 `json:"state"`
	CodeChallenge       string                 `json:"code_challenge"`
	CodeChallengeMethod string                 `json:"code_challenge_method"`
	CIMDMetadata        *cimd.MetadataSnapshot `json:"cimd_metadata,omitempty"`
	IssuedAt            time.Time              `json:"iat"`
	ExpiresAt           time.Time              `json:"exp"`
}

// IsExpired reports whether the session token TTL has elapsed.
func (c *AuthorizationSessionClaims) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// NewAuthorizationSessionClaims initialises claims with IssuedAt and ExpiresAt set from now.
func NewAuthorizationSessionClaims(
	agentID id.AgentID,
	principal id.Principal,
	clientID, originalURL, redirectURI, scope, state, codeChallenge, codeChallengeMethod string,
	cimdMetadata *cimd.MetadataSnapshot,
) *AuthorizationSessionClaims {
	now := time.Now()
	return &AuthorizationSessionClaims{
		AgentID:             agentID,
		Principal:           principal,
		ClientID:            clientID,
		OriginalURL:         originalURL,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CIMDMetadata:        cimdMetadata,
		IssuedAt:            now,
		ExpiresAt:           now.Add(authorizationSessionTokenTTL),
	}
}

// CreateAuthorizationSessionToken seals claims into a compact JWE string.
func (s *Service) CreateAuthorizationSessionToken(claims *AuthorizationSessionClaims) (string, error) {
	if s.jweTokenService == nil {
		return "", fmt.Errorf("jweTokenService not configured")
	}
	return s.jweTokenService.Encrypt(claims)
}

// ValidateAuthorizationSessionToken decrypts and validates an AuthorizationSessionClaims token.
func (s *Service) ValidateAuthorizationSessionToken(token string) (*AuthorizationSessionClaims, error) {
	if s.jweTokenService == nil {
		return nil, fmt.Errorf("jweTokenService not configured")
	}
	var claims AuthorizationSessionClaims
	if err := s.jweTokenService.DecryptAndValidate(token, &claims); err != nil {
		return nil, err
	}
	return &claims, nil
}
