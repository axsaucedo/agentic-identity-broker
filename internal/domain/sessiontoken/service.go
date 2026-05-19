package sessiontoken

import (
	"errors"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
)

// Service handles creation and validation of JWE-sealed authorization session tokens.
type Service struct {
	jweTokenService *jwe.TokenService
}

// NewService creates a SessionTokenService backed by the given JWE token service.
func NewService(jweTokenService *jwe.TokenService) *Service {
	return &Service{jweTokenService: jweTokenService}
}

// Create seals claims into a compact JWE string.
func (s *Service) Create(claims *AuthorizationSessionClaims) (string, error) {
	if s.jweTokenService == nil {
		return "", fmt.Errorf("jweTokenService not configured")
	}
	return s.jweTokenService.Encrypt(claims)
}

// ValidateAuthorizationSessionToken decrypts a JWE session token and validates
// expiry, agent binding, and principal binding.
func (s *Service) ValidateAuthorizationSessionToken(token string, agentID id.AgentID, principal id.Principal) (*AuthorizationSessionClaims, error) {
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
