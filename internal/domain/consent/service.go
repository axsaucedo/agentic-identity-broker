package consent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

var (
	// ErrAgentNotFound is returned when the specified agent does not exist.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrInvalidScopes is returned when requested scopes don't exist in the service configuration.
	ErrInvalidScopes = errors.New("invalid scopes requested")
	// ErrServiceNotFound is returned when a referenced service does not exist.
	ErrServiceNotFound = errors.New("third-party service not found")
)

// Service provides consent management business logic.
// This service orchestrates between agent, service, and grant repositories
// to implement consent workflows following FR-009 through FR-020.
type Service struct {
	agentRepo   ports.AgentRepository
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository
	grantRepo   ports.UserGrantRepository
}

// NewService creates a new ConsentService.
func NewService(
	agentRepo ports.AgentRepository,
	serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
	grantRepo ports.UserGrantRepository,
) *Service {
	return &Service{
		agentRepo:   agentRepo,
		serviceRepo: serviceRepo,
		grantRepo:   grantRepo,
	}
}

// AgentConsentInfo contains all information needed for a user to make a consent decision.
type AgentConsentInfo struct {
	Agent                    *storage.Agent
	AvailableThirdpartyServices []*storage.ThirdpartyOAuth2Service
}

// GetAgentConsentInfo retrieves agent metadata and all available third-party services.
// This provides the information a user needs to make an informed consent decision (FR-009, FR-010, FR-025).
// Returns ErrAgentNotFound if the agent doesn't exist.
func (s *Service) GetAgentConsentInfo(ctx context.Context, agentID string) (*AgentConsentInfo, error) {
	// Fetch agent
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	// Fetch all available third-party services (FR-025: all services available to all agents)
	services, err := s.serviceRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list third-party services: %w", err)
	}

	// Redact client secrets in response (SR-003)
	redactedServices := make([]*storage.ThirdpartyOAuth2Service, len(services))
	for i, svc := range services {
		redactedServices[i] = svc.RedactedCopy()
	}

	return &AgentConsentInfo{
		Agent:                    agent.Copy(),
		AvailableThirdpartyServices: redactedServices,
	}, nil
}

// GrantRequest represents a request to grant or update permissions.
type GrantRequest struct {
	Principal              string
	AgentID                string
	ValidUntil             *time.Time
	DelegatedOAuth2Tokens  []storage.DelegatedToken
}

// GrantConsent creates or updates a user grant (upsert semantics per FR-013, FR-015).
// Validates that:
// - Agent exists (FR-020)
// - All requested scopes exist in their respective service configurations (FR-018)
// - ValidUntil is in the future if provided (FR-016)
// Returns the created/updated grant or an error.
func (s *Service) GrantConsent(ctx context.Context, req *GrantRequest) (*storage.UserGrant, error) {
	// Validate agent exists
	agent, err := s.agentRepo.Get(ctx, req.AgentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("%w: agent_id=%s", ErrAgentNotFound, req.AgentID)
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("%w: agent_id=%s", ErrAgentNotFound, req.AgentID)
	}

	// Validate all scopes exist in their respective services (FR-018)
	if err := s.validateScopes(ctx, req.DelegatedOAuth2Tokens); err != nil {
		return nil, err
	}

	// Check for existing grant (upsert semantics)
	existingGrant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, req.Principal, req.AgentID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return nil, fmt.Errorf("failed to find existing grant: %w", err)
	}

	var grant *storage.UserGrant
	if existingGrant != nil {
		// Update existing grant (FR-013)
		existingGrant.ValidUntil = req.ValidUntil
		existingGrant.DelegatedOAuth2Tokens = req.DelegatedOAuth2Tokens
		existingGrant.UpdatedAt = time.Now()

		if err := existingGrant.Validate(); err != nil {
			return nil, fmt.Errorf("grant validation failed: %w", err)
		}

		if err := s.grantRepo.Update(ctx, existingGrant); err != nil {
			return nil, fmt.Errorf("failed to update grant: %w", err)
		}
		grant = existingGrant
	} else {
		// Create new grant (FR-011)
		grant = &storage.UserGrant{
			ID:                    generateID(), // ID generation will be handled by repository
			Principal:             req.Principal,
			AgentID:               req.AgentID,
			ValidUntil:            req.ValidUntil,
			DelegatedOAuth2Tokens: req.DelegatedOAuth2Tokens,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		if err := grant.ValidateForCreate(); err != nil {
			return nil, fmt.Errorf("grant validation failed: %w", err)
		}

		if err := s.grantRepo.Create(ctx, grant); err != nil {
			return nil, fmt.Errorf("failed to create grant: %w", err)
		}
	}

	return grant.Copy(), nil
}

// RevokeConsent deletes a user grant (FR-014).
// Idempotent: returns nil if grant doesn't exist.
func (s *Service) RevokeConsent(ctx context.Context, principal string, agentID string) error {
	// Find grant
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, principal, agentID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return fmt.Errorf("failed to find grant: %w", err)
	}

	// If grant doesn't exist, operation is idempotent (already revoked)
	if grant == nil {
		return nil
	}

	// Delete grant
	if err := s.grantRepo.Delete(ctx, grant.ID); err != nil {
		return fmt.Errorf("failed to delete grant: %w", err)
	}

	return nil
}

// GetActiveGrants retrieves all active grants for a principal and agent.
// Filters expired grants per FR-019.
// Returns empty slice if no active grants exist (not an error per FR-012).
func (s *Service) GetActiveGrants(ctx context.Context, principal string, agentID string) ([]*storage.UserGrant, error) {
	grants, err := s.grantRepo.ListByPrincipalAndAgent(ctx, principal, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}

	// Filter to only active grants (FR-019)
	activeGrants := make([]*storage.UserGrant, 0, len(grants))
	for _, grant := range grants {
		if grant.IsActive() {
			activeGrants = append(activeGrants, grant.Copy())
		}
	}

	return activeGrants, nil
}

// validateScopes validates that all requested scopes exist in their respective service configurations.
// Returns ErrInvalidScopes with details if any scope doesn't exist (FR-018).
func (s *Service) validateScopes(ctx context.Context, delegations []storage.DelegatedToken) error {
	for i, delegation := range delegations {
		// Fetch service
		service, err := s.serviceRepo.Get(ctx, delegation.ThirdpartyOAuth2ServiceID)
		if err != nil {
			return fmt.Errorf("failed to get service %s: %w", delegation.ThirdpartyOAuth2ServiceID, err)
		}
		if service == nil {
			return fmt.Errorf("%w: service_id=%s", ErrServiceNotFound, delegation.ThirdpartyOAuth2ServiceID)
		}

		// Build map of valid scopes for this service
		validScopes := make(map[string]bool)
		for _, scope := range service.Scopes {
			validScopes[scope.ScopeValue] = true
		}

		// Validate each requested scope exists
		invalidScopes := []string{}
		for _, requestedScope := range delegation.Scopes {
			if !validScopes[requestedScope] {
				invalidScopes = append(invalidScopes, requestedScope)
			}
		}

		if len(invalidScopes) > 0 {
			return fmt.Errorf("%w: delegation %d (service=%s) has invalid scopes: %v",
				ErrInvalidScopes, i, delegation.ThirdpartyOAuth2ServiceID, invalidScopes)
		}
	}

	return nil
}

// generateID is a placeholder for ID generation.
// In production, this would use UUID v4 generation.
func generateID() string {
	// This will be replaced by proper UUID generation in repository implementations
	return ""
}
