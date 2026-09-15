package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

type userDelegationVerifier struct {
	consentService *consent.Service
}

func newUserDelegationVerifier(consentService *consent.Service) ports.UserDelegationVerifier {
	return userDelegationVerifier{consentService: consentService}
}

func (v userDelegationVerifier) VerifyUserDelegation(ctx context.Context, principal id.Principal, agentID id.AgentID) (ports.UserDelegationStatus, error) {
	if v.consentService == nil {
		return "", fmt.Errorf("user delegation verifier is not configured")
	}

	_, err := v.consentService.VerifyAgentAccess(ctx, principal, agentID)
	if err == nil {
		return ports.UserDelegationActive, nil
	}
	if errors.Is(err, consent.ErrAgentAccessDenied) {
		return ports.UserDelegationMissing, nil
	}
	if errors.Is(err, consent.ErrGrantExpired) {
		return ports.UserDelegationExpired, nil
	}
	return "", fmt.Errorf("verify user delegation: %w", err)
}
