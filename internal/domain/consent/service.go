package consent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

var (
	// ErrAgentNotFound is returned when the specified agent does not exist.
	ErrAgentNotFound = errors.New("agent not found")
	// ErrInvalidScopes is returned when requested scopes don't exist in the service configuration.
	ErrInvalidScopes = errors.New("invalid scopes requested")
	// ErrServiceNotFound is returned when a referenced service does not exist.
	ErrServiceNotFound = errors.New("third-party service not found")
	// ErrAgentAccessDenied is returned when a user has not granted an agent access.
	ErrAgentAccessDenied = errors.New("agent access denied")
	// ErrGrantExpired is returned when a user grant has expired.
	ErrGrantExpired = errors.New("grant expired")
	// ErrGrantNotFound is returned by RevokeConsentForPrincipal when no active grant exists
	// for the (principal, agent) pair. The handler maps this to HTTP 404.
	ErrGrantNotFound = errors.New("grant not found")
	// ErrGrantValidation is returned when a UserGrant fails domain validation (e.g. empty
	// scope list, duplicate scopes). The handler maps this to HTTP 400.
	ErrGrantValidation = errors.New("grant validation failed")
)

// Service provides consent management business logic.
// This service orchestrates between agent, service, and grant repositories
// to implement consent workflows following FR-009 through FR-020.
type Service struct {
	agentRepo       ports.AgentRepository
	providerService *thirdparty.ThirdpartyOAuth2ProviderService
	grantRepo       ports.UserGrantRepository
	sessionRepo     ports.UserSessionRepository
	logger          *slog.Logger
}

// NewService creates a new ConsentService.
func NewService(
	agentRepo ports.AgentRepository,
	providerService *thirdparty.ThirdpartyOAuth2ProviderService,
	grantRepo ports.UserGrantRepository,
	sessionRepo ports.UserSessionRepository,
	logger *slog.Logger,
) *Service {
	return &Service{
		agentRepo:       agentRepo,
		providerService: providerService,
		grantRepo:       grantRepo,
		sessionRepo:     sessionRepo,
		logger:          logger,
	}
}

// ServiceScopeInfo is an OAuth2 scope with its human-readable description.
type ServiceScopeInfo struct {
	Name        string
	Description string
}

// ServiceRequirementStatus is an agent service requirement enriched with the user's session status.
type ServiceRequirementStatus struct {
	ServiceID       id.ServiceID
	DisplayName     string
	RequirementType storage.RequirementType
	RequiredScopes  []ServiceScopeInfo
	IsConnected     bool
}

// GetAgentWithServiceRequirements retrieves the agent and its service requirements enriched
// with the authenticated user's session status for each required service.
// Returns ErrAgentNotFound if the agent does not exist.
func (s *Service) GetAgentWithServiceRequirements(
	ctx context.Context,
	userPrincipal id.Principal,
	agentID id.AgentID,
) (*storage.Agent, []ServiceRequirementStatus, error) {
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, nil, ErrAgentNotFound
		}
		return nil, nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if len(agent.ServiceRequirements) == 0 {
		return agent, []ServiceRequirementStatus{}, nil
	}

	// Deduplicate service IDs before loading to avoid redundant decryptions.
	serviceIDs := make(map[id.ServiceID]bool, len(agent.ServiceRequirements))
	for _, req := range agent.ServiceRequirements {
		serviceIDs[req.ServiceID] = true
	}

	serviceMap := make(map[id.ServiceID]*model.ThirdpartyOAuth2ProviderEntity, len(serviceIDs))
	for serviceID := range serviceIDs {
		svc, err := s.providerService.Get(ctx, serviceID)
		if err != nil {
			if errors.Is(err, ports.ErrNotFound) {
				s.logger.Warn("service not found for agent requirement",
					"service_id", serviceID,
					"agent_id", agentID)
				continue
			}
			return nil, nil, fmt.Errorf("loading service %s: %w", serviceID, err)
		}
		serviceMap[serviceID] = svc
	}

	requirements := make([]ServiceRequirementStatus, 0, len(agent.ServiceRequirements))
	for _, req := range agent.ServiceRequirements {
		svc, ok := serviceMap[req.ServiceID]
		if !ok {
			continue
		}

		session, err := s.sessionRepo.FindByPrincipalAndService(ctx, userPrincipal, req.ServiceID)
		if err != nil {
			return nil, nil, fmt.Errorf("checking session status for service %s: %w", req.ServiceID, err)
		}

		scopeDesc := make(map[string]string, len(svc.Scopes))
		for _, scope := range svc.Scopes {
			scopeDesc[scope.ScopeValue] = scope.Description
		}

		scopes := make([]ServiceScopeInfo, len(req.RequiredScopes))
		for i, name := range req.RequiredScopes {
			scopes[i] = ServiceScopeInfo{Name: name, Description: scopeDesc[name]}
		}

		requirements = append(requirements, ServiceRequirementStatus{
			ServiceID:       req.ServiceID,
			DisplayName:     svc.DisplayName,
			RequirementType: req.RequirementType,
			RequiredScopes:  scopes,
			IsConnected:     session != nil && !session.IsExpired(),
		})
	}

	return agent, requirements, nil
}

// AgentConsentInfo contains all information needed for a user to make a consent decision.
type AgentConsentInfo struct {
	Agent                       *storage.Agent
	AvailableThirdpartyServices []*model.ThirdpartyOAuth2ProviderEntity
}

// GetAgentConsentInfo retrieves agent metadata and all available third-party services.
// This provides the information a user needs to make an informed consent decision (FR-009, FR-010, FR-025).
// Returns ErrAgentNotFound if the agent doesn't exist.
func (s *Service) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*AgentConsentInfo, error) {
	// Fetch agent
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	// Fetch all available third-party services (FR-025: all services available to all agents)
	services, err := s.providerService.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list third-party services: %w", err)
	}

	// Redact client secrets in response (SR-003)
	redactedServices := make([]*model.ThirdpartyOAuth2ProviderEntity, len(services))
	for i, svc := range services {
		redactedServices[i] = svc.RedactedCopy()
	}

	return &AgentConsentInfo{
		Agent:                       agent.Copy(),
		AvailableThirdpartyServices: redactedServices,
	}, nil
}

// GrantRequest represents a request to grant or update permissions.
type GrantRequest struct {
	Principal             id.Principal
	AgentID               id.AgentID
	ValidUntil            *time.Time
	DelegatedOAuth2Tokens []storage.DelegatedToken
}

// ValidateGrantRequest validates a grant request without persisting anything.
// Exposed separately from GrantConsent so callers can fail fast before
// consuming irreversible state (e.g., a one-shot authorization session).
func (s *Service) ValidateGrantRequest(ctx context.Context, req *GrantRequest) error {
	// Validate agent exists
	agent, err := s.agentRepo.Get(ctx, req.AgentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return fmt.Errorf("%w: agent_id=%s", ErrAgentNotFound, req.AgentID)
		}
		return fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return fmt.Errorf("%w: agent_id=%s", ErrAgentNotFound, req.AgentID)
	}

	// Validate all scopes exist in their respective services (FR-018)
	if err := s.validateScopes(ctx, req.DelegatedOAuth2Tokens); err != nil {
		return err
	}

	return nil
}

// GrantConsent creates or updates a user grant (upsert semantics per FR-013, FR-015).
// Validates that:
// - Agent exists (FR-020)
// - All requested scopes exist in their respective service configurations (FR-018)
// - ValidUntil is in the future if provided (FR-016)
// Returns the created/updated grant or an error.
func (s *Service) GrantConsent(ctx context.Context, req *GrantRequest) (*storage.UserGrant, error) {
	if err := s.ValidateGrantRequest(ctx, req); err != nil {
		return nil, err
	}

	// Check for existing grant (upsert semantics)
	existingGrant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, req.Principal, req.AgentID)
	if err != nil && !errors.Is(err, ports.ErrNotFound) {
		return nil, fmt.Errorf("failed to find existing grant: %w", err)
	}

	var grant *storage.UserGrant
	if existingGrant != nil {
		if grantMatchesRequest(existingGrant, req) {
			return existingGrant.Copy(), nil
		}

		// Update existing grant (FR-013)
		existingGrant.ValidUntil = req.ValidUntil
		existingGrant.DelegatedOAuth2Tokens = req.DelegatedOAuth2Tokens
		existingGrant.UpdatedAt = time.Now()

		if err := existingGrant.Validate(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrGrantValidation, err)
		}

		if err := s.grantRepo.Update(ctx, existingGrant); err != nil {
			return nil, fmt.Errorf("failed to update grant: %w", err)
		}
		grant = existingGrant
	} else {
		// Create new grant (FR-011)
		grant = &storage.UserGrant{
			ID:                    id.NewGrantID(),
			Principal:             req.Principal,
			AgentID:               req.AgentID,
			ValidUntil:            req.ValidUntil,
			DelegatedOAuth2Tokens: req.DelegatedOAuth2Tokens,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}

		if err := grant.ValidateForCreate(); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrGrantValidation, err)
		}

		if err := s.grantRepo.Create(ctx, grant); err != nil {
			return nil, fmt.Errorf("failed to create grant: %w", err)
		}
	}

	return grant.Copy(), nil
}

func grantMatchesRequest(grant *storage.UserGrant, req *GrantRequest) bool {
	if grant == nil || req == nil {
		return false
	}

	return validUntilMatches(grant.ValidUntil, req.ValidUntil) &&
		delegatedTokensMatch(grant.DelegatedOAuth2Tokens, req.DelegatedOAuth2Tokens)
}

func validUntilMatches(left, right *time.Time) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.Equal(*right)
	}
}

func delegatedTokensMatch(left, right []storage.DelegatedToken) bool {
	return slices.EqualFunc(left, right, func(leftToken, rightToken storage.DelegatedToken) bool {
		return leftToken.ThirdpartyOAuth2ServiceID == rightToken.ThirdpartyOAuth2ServiceID &&
			slices.Equal(leftToken.Scopes, rightToken.Scopes)
	})
}

// RevokeConsent deletes a user grant (FR-014).
// Idempotent: returns nil if the grant doesn't exist (absence is not an error).
// Used by the POST /grants path with empty tokens.
func (s *Service) RevokeConsent(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	if err := s.grantRepo.DeleteByPrincipalAndAgentID(ctx, principal, agentID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil // idempotent: absence is not an error
		}
		return fmt.Errorf("failed to revoke consent: %w", err)
	}
	return nil
}

// RevokeConsentForPrincipal revokes the authenticated user's grant for the given agent (FR-014).
// This is the user-facing revocation entry point for DELETE /api/consent/agent/{agent-id}/grants.
//
// Unlike RevokeConsent, this method:
// - Is NOT idempotent: absence of grant returns ErrGrantNotFound (handler maps to 404)
// - Emits a structured audit log on success with action, principal, agent_id, and grant_id
func (s *Service) RevokeConsentForPrincipal(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	// Phase 1: look up the grant to capture the ID for the audit log.
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, principal, agentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return fmt.Errorf("%w", ErrGrantNotFound)
		}
		return fmt.Errorf("failed to find grant: %w", err)
	}

	// Phase 2: delete the grant.
	if err := s.grantRepo.DeleteByPrincipalAndAgentID(ctx, principal, agentID); err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			// Concurrent revocation raced us — treat as not found.
			return fmt.Errorf("%w", ErrGrantNotFound)
		}
		return fmt.Errorf("failed to revoke consent: %w", err)
	}

	s.logger.Info("grant revoked",
		"action", "grant_revoked",
		"principal", principal,
		"agent_id", agentID,
		"grant_id", grant.ID)

	return nil
}

// GetActiveGrants retrieves all active grants for a principal and agent.
// Filters expired grants per FR-019.
// Returns empty slice if no active grants exist (not an error per FR-012).
func (s *Service) GetActiveGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
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

// VerifyAgentAccess verifies that a user has granted an agent access.
// This method checks for grant existence, expiration, and revocation status.
// Returns the active grant if valid, or an error if missing, expired, or revoked.
//
// Error handling:
// - ErrAgentAccessDenied: User has not granted the agent any access
// - ErrGrantExpired: User grant has expired
// - Other errors: Repository or system errors
//
// This method is used by token exchange flows to verify authorization before
// issuing delegated tokens. Per Constitution Principle I (Security-First),
// fails closed with access denied for any ambiguous state.
func (s *Service) VerifyAgentAccess(ctx context.Context, principal id.Principal, agentID id.AgentID) (*storage.UserGrant, error) {
	// Look up grant by principal and agent
	grant, err := s.grantRepo.FindByPrincipalAndAgent(ctx, principal, agentID)
	if err != nil {
		// Check if it's a NotFound error
		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("%w: user has not granted permission for agent (principal: %s, agent: %s)",
				ErrAgentAccessDenied, principal, agentID)
		}
		return nil, fmt.Errorf("failed to verify user grant: %w", err)
	}

	// Defensive check: ensure grant is not nil
	if grant == nil {
		return nil, fmt.Errorf("%w: user has not granted permission for agent (principal: %s, agent: %s)",
			ErrAgentAccessDenied, principal, agentID)
	}

	// Check grant is active (not expired)
	if !grant.IsActive() {
		return nil, fmt.Errorf("%w: user grant expired at %s (principal: %s, agent: %s)",
			ErrGrantExpired, grant.ValidUntil.Format(time.RFC3339), principal, agentID)
	}

	// Hard deletion is the revocation mechanism: a missing grant (ports.ErrNotFound) is treated
	// as access denied (fail closed). No revoked status field is needed.

	// Return copy to prevent external mutation
	return grant.Copy(), nil
}

// validateScopes validates that all requested scopes exist in their respective service configurations.
// Returns ErrInvalidScopes with details if any scope doesn't exist (FR-018).
func (s *Service) validateScopes(ctx context.Context, delegations []storage.DelegatedToken) error {
	for i, delegation := range delegations {
		// Fetch service
		service, err := s.providerService.Get(ctx, delegation.ThirdpartyOAuth2ServiceID)
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

// AgentDelegation represents aggregated information about grants for a specific agent.
// This is used for the consent management UI to display active delegations.
type AgentDelegation struct {
	AgentID          id.AgentID `json:"agentId"`
	DisplayName      string     `json:"displayName"`
	LogoURL          *string    `json:"logoUrl,omitempty"`
	ActiveGrantCount int        `json:"activeGrantCount"`
	LastModifiedAt   time.Time  `json:"lastModifiedAt"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
}

// AgentDetail represents detailed information about an agent for User Story 2.
// This provides all metadata needed for the agent-specific grants view.
type AgentDetail struct {
	AgentID              id.AgentID `json:"agentId"`
	DisplayName          string     `json:"displayName"`
	Description          string     `json:"description"`
	LogoURL              *string    `json:"logoUrl,omitempty"`
	GovernanceURL        *string    `json:"governanceUrl,omitempty"`
	UserDocumentationURL *string    `json:"userDocumentationUrl,omitempty"`
	AgentInterfaceURL    *string    `json:"agentInterfaceUrl,omitempty"`
}

// ServiceScope represents a permission scope within a third-party service.
type ServiceScope struct {
	Value       string `json:"value"`
	Description string `json:"description"`
}

// ThirdpartyService represents a third-party service with its available scopes.
// This is used in the agent detail view to show what services an agent can request.
type ThirdpartyService struct {
	ServiceID   id.ServiceID   `json:"serviceId"`
	DisplayName string         `json:"displayName"`
	LogoURL     *string        `json:"logoUrl,omitempty"`
	Scopes      []ServiceScope `json:"scopes"`
}

// GetAgentDetail retrieves detailed information about an agent and its available services.
// This is used for User Story 2: Review Agent-Specific Grants.
// Returns agent metadata and all available third-party services with their scopes.
// Returns ErrAgentNotFound if the agent doesn't exist.
func (s *Service) GetAgentDetail(ctx context.Context, agentID id.AgentID) (*AgentDetail, []ThirdpartyService, error) {
	// Fetch agent
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, nil, ErrAgentNotFound
		}
		return nil, nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, nil, ErrAgentNotFound
	}

	// Fetch all available third-party services (FR-025: all services available to all agents)
	services, err := s.providerService.List(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list third-party services: %w", err)
	}

	// Convert agent to AgentDetail
	agentDetail := &AgentDetail{
		AgentID:              agent.ID,
		DisplayName:          agent.DisplayName,
		Description:          agent.Description,
		LogoURL:              nil, // TODO: add logo_url field to Agent entity
		GovernanceURL:        agent.GovernanceURL,
		UserDocumentationURL: agent.UserDocumentationURL,
		AgentInterfaceURL:    agent.AgentInterfaceURL,
	}

	// Convert services to ThirdpartyService DTOs
	thirdpartyServices := make([]ThirdpartyService, len(services))
	for i, svc := range services {
		scopes := make([]ServiceScope, len(svc.Scopes))
		for j, scope := range svc.Scopes {
			scopes[j] = ServiceScope{
				Value:       scope.ScopeValue,
				Description: scope.Description,
			}
		}

		thirdpartyServices[i] = ThirdpartyService{
			ServiceID:   svc.ID,
			DisplayName: svc.DisplayName,
			LogoURL:     nil, // TODO: add logo_url field to ThirdpartyOAuth2ProviderEntity
			Scopes:      scopes,
		}
	}

	return agentDetail, thirdpartyServices, nil
}

// GetUserGrants retrieves all grants for a specific principal and agent.
// This is used for User Story 2 to display what permissions the user has already granted to an agent.
// Returns empty slice if no grants exist (not an error).
// Returns ErrAgentNotFound if the agent doesn't exist.
func (s *Service) GetUserGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	// Verify agent exists first
	agent, err := s.agentRepo.Get(ctx, agentID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, ErrAgentNotFound
		}
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil {
		return nil, ErrAgentNotFound
	}

	// Fetch grants for this principal and agent
	grants, err := s.grantRepo.ListByPrincipalAndAgent(ctx, principal, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}

	// Return copies to prevent external mutation
	result := make([]*storage.UserGrant, len(grants))
	for i, grant := range grants {
		result[i] = grant.Copy()
	}

	return result, nil
}

// GetAgentDelegations retrieves all agent delegations for a principal.
// Groups grants by agent_id and returns summary information for each agent.
// Returns empty slice if no grants exist (not an error).
func (s *Service) GetAgentDelegations(ctx context.Context, principal id.Principal) ([]AgentDelegation, error) {
	// Fetch all active grants for this principal
	grants, err := s.grantRepo.ListByPrincipal(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("failed to list grants: %w", err)
	}

	// Group grants by agent_id
	agentMap := make(map[id.AgentID]*AgentDelegation)

	for _, grant := range grants {
		delegation, exists := agentMap[grant.AgentID]
		if !exists {
			// Fetch agent information
			agent, err := s.agentRepo.Get(ctx, grant.AgentID)
			if err != nil {
				// If agent not found, skip this grant (defensive: should not happen)
				if errors.Is(err, ports.ErrNotFound) {
					continue
				}
				return nil, fmt.Errorf("failed to get agent %s: %w", grant.AgentID, err)
			}

			delegation = &AgentDelegation{
				AgentID:          grant.AgentID,
				DisplayName:      agent.DisplayName,
				LogoURL:          nil, // TODO: add logo_url field to Agent entity
				ActiveGrantCount: 0,
				LastModifiedAt:   grant.UpdatedAt,
				ExpiresAt:        grant.ValidUntil,
			}
			agentMap[grant.AgentID] = delegation
		}

		// Update delegation stats
		delegation.ActiveGrantCount++
		if grant.UpdatedAt.After(delegation.LastModifiedAt) {
			delegation.LastModifiedAt = grant.UpdatedAt
		}
		// Update ExpiresAt to the earliest expiration if multiple grants exist
		if grant.ValidUntil != nil {
			if delegation.ExpiresAt == nil || grant.ValidUntil.Before(*delegation.ExpiresAt) {
				delegation.ExpiresAt = grant.ValidUntil
			}
		}
	}

	// Convert map to slice
	delegations := make([]AgentDelegation, 0, len(agentMap))
	for _, delegation := range agentMap {
		delegations = append(delegations, *delegation)
	}

	return delegations, nil
}
