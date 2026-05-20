package sessiontoken

import (
	"errors"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
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
		return "", ErrSessionServiceNotConfigured
	}
	return s.jweTokenService.Encrypt(claims)
}

// ValidateAuthorizationSessionToken decrypts a JWE session token and validates
// expiry, agent binding, and principal binding. Returns a port-local DTO.
func (s *Service) ValidateAuthorizationSessionToken(token string, agentID id.AgentID, principal id.Principal) (*ports.AuthorizationSession, error) {
	if s.jweTokenService == nil {
		return nil, ErrSessionServiceNotConfigured
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
	return claims.ToResult(), nil
}
