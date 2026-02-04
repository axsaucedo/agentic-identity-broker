// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/go-chi/chi/v5"
)

// ServiceRequirementForUser represents a service requirement enriched with user session status.
// This is used in the consent screen to display which services the agent requires and user's connection status.
type ServiceRequirementForUser struct {
	ServiceID        string                 `json:"serviceId"`
	ServiceName      string                 `json:"serviceName"`
	RequirementType  string                 `json:"requirementType"` // "mandatory" or "optional"
	RequiredScopes   []ScopeWithDescription `json:"requiredScopes"`
	ConnectionStatus string                 `json:"connectionStatus"` // "connected" or "not_connected"
}

// ScopeWithDescription represents an OAuth2 scope with its description.
type ScopeWithDescription struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AgentDetailHandler handles HTTP requests for retrieving detailed agent information.
// Implements User Story 2: Review Agent-Specific Grants (GET /api/consent/agent/:agentId).
// Phase 6 extension: Includes service requirements with user connection status.
type AgentDetailHandler struct {
	consentService    ConsentService
	agentRepository   ports.AgentRepository
	sessionRepository ports.UserSessionRepository
	serviceRepository ports.ThirdpartyOAuth2ServiceRepository
	logger            *slog.Logger
}

// NewAgentDetailHandler creates a new agent detail handler.
func NewAgentDetailHandler(consentService ConsentService, logger *slog.Logger) *AgentDetailHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentDetailHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// WithAgentRepository sets the agent repository for this handler.
// Used to lookup the full agent entity including service requirements.
func (h *AgentDetailHandler) WithAgentRepository(repo ports.AgentRepository) *AgentDetailHandler {
	h.agentRepository = repo
	return h
}

// WithSessionRepository sets the session repository for this handler.
// Used to lookup user session status with third-party services.
func (h *AgentDetailHandler) WithSessionRepository(repo ports.UserSessionRepository) *AgentDetailHandler {
	h.sessionRepository = repo
	return h
}

// WithServiceRepository sets the service repository for this handler.
// Used to lookup service metadata including display names and scope descriptions.
func (h *AgentDetailHandler) WithServiceRepository(repo ports.ThirdpartyOAuth2ServiceRepository) *AgentDetailHandler {
	h.serviceRepository = repo
	return h
}

// GetAgentDetailResponse represents the response for GET /api/consent/agent/:agentId.
type GetAgentDetailResponse struct {
	Data AgentDetailData `json:"data"`
}

// AgentDetailData contains the agent detail and associated services.
type AgentDetailData struct {
	Agent    consent.AgentDetail         `json:"agent"`
	Services []ServiceRequirementForUser `json:"services"`
}

// GetAgentDetail handles GET /api/consent/agent/:agentId
// Returns detailed information about an agent and its required services with user session status.
// Only returns services that are configured as requirements for the agent (not all system services).
//
// Response codes:
// - 200 OK: Returns agent details and services with connection status
// - 400 Bad Request: Invalid agent ID format
// - 401 Unauthorized: Principal not found in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *AgentDetailHandler) GetAgentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.logger.Warn("agent ID is missing in request")
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID is required")
		return
	}

	// Extract principal from context (middleware ensures this is present)
	userID, ok := getPrincipalFromContext(ctx)
	if !ok {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "principal required")
		return
	}

	// Call consent service to get agent detail (used for basic agent info)
	agentDetail, _, err := h.consentService.GetAgentDetail(ctx, agentID)
	if err != nil {
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found", "agent_id", agentID)
			h.writeError(w, http.StatusNotFound, "not found", "agent not found")
			return
		}

		h.logger.Error("failed to get agent detail",
			"agent_id", agentID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Load full agent entity to access service requirements
	agent, err := h.getAgent(ctx, agentID)
	if err != nil {
		h.logger.Error("failed to load agent",
			"agent_id", agentID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Build service requirements enriched with user session status
	serviceRequirements, err := h.buildServiceRequirementsForUser(ctx, userID, agent)
	if err != nil {
		h.logger.Error("failed to build service requirements",
			"agent_id", agentID,
			"user_id", userID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Sort services: mandatory first, then optional
	sortServiceRequirements(serviceRequirements)

	// Build response
	response := GetAgentDetailResponse{
		Data: AgentDetailData{
			Agent:    *agentDetail,
			Services: serviceRequirements,
		},
	}

	h.logger.Info("agent detail retrieved",
		"agent_id", agentID,
		"user_id", userID,
		"services_count", len(serviceRequirements))

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response.
func (h *AgentDetailHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := encodeJSON(w, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *AgentDetailHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}

// buildServiceRequirementsForUser enriches agent service requirements with user-specific session status.
// Returns a list of ServiceRequirementForUser with connection status for each service.
// Requirements:
// - Query agent's ServiceRequirements array
// - For each requirement, lookup the ThirdPartyOAuth2Service by service_id
// - Check user's session status with that service (from session repository)
// - Include connection status: "connected" if session exists and is valid, "not_connected" otherwise
// - Resolve scope descriptions from service configuration
// - Return enriched ServiceRequirementForUser model
// Error handling:
// - If service not found: Log warning and skip (service may have been removed)
// - If session lookup fails: Treat as "not_connected"
// - Return partial results if some services unavailable (fail-open for fetch)
func (h *AgentDetailHandler) buildServiceRequirementsForUser(ctx context.Context, userID string, agent *storage.Agent) ([]ServiceRequirementForUser, error) {
	if len(agent.ServiceRequirements) == 0 {
		return []ServiceRequirementForUser{}, nil
	}

	var results []ServiceRequirementForUser

	for _, req := range agent.ServiceRequirements {
		// Lookup service by ID
		svc, err := h.serviceRepository.Get(ctx, req.ServiceID)
		if err != nil {
			// Service not found, skip or log
			h.logger.Warn("Service not found during requirement building",
				"service_id", req.ServiceID,
				"error", err)
			continue
		}

		// Check user's session status with this service
		session, err := h.sessionRepository.FindByPrincipalAndService(ctx, userID, req.ServiceID)
		if err != nil {
			h.logger.Warn("Error checking session status",
				"user_id", userID,
				"service_id", req.ServiceID,
				"error", err)
			// Treat error as no session
			session = nil
		}

		// Determine connection status
		connStatus := "not_connected"
		if session != nil {
			connStatus = "connected"
		}

		// Build scope list with descriptions
		var scopes []ScopeWithDescription
		for _, scopeName := range req.RequiredScopes {
			// Find scope description from service configuration
			scopeDesc := ""
			for _, svcScope := range svc.Scopes {
				if svcScope.ScopeValue == scopeName {
					scopeDesc = svcScope.Description
					break
				}
			}

			scope := ScopeWithDescription{
				Name:        scopeName,
				Description: scopeDesc,
			}
			scopes = append(scopes, scope)
		}

		result := ServiceRequirementForUser{
			ServiceID:        req.ServiceID,
			ServiceName:      svc.DisplayName,
			RequirementType:  string(req.RequirementType),
			RequiredScopes:   scopes,
			ConnectionStatus: connStatus,
		}
		results = append(results, result)
	}

	return results, nil
}

// getAgent loads a full agent entity from storage.
// This is used to access service requirements which are not available in AgentDetail DTO.
func (h *AgentDetailHandler) getAgent(ctx context.Context, agentID string) (*storage.Agent, error) {
	// We need an agent repository. For now, we'll use a workaround by checking if we have access to it
	// through the consentService. Since we don't have direct access, we need to add it to the handler.
	// This will be injected via builder or a new method.
	if h.agentRepository == nil {
		return nil, errors.New("agent repository not configured")
	}
	return h.agentRepository.Get(ctx, agentID)
}

// getPrincipalFromContext extracts the principal from the request context.
func getPrincipalFromContext(ctx context.Context) (string, bool) {
	return principal.FromContext(ctx)
}

// sortServiceRequirements sorts service requirements with mandatory services first, then optional.
func sortServiceRequirements(services []ServiceRequirementForUser) {
	// Sort so mandatory services appear first
	for i := 0; i < len(services); i++ {
		for j := i + 1; j < len(services); j++ {
			if services[i].RequirementType == "optional" && services[j].RequirementType == "mandatory" {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}
