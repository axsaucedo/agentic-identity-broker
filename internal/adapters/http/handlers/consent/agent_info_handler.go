// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/go-chi/chi/v5"
)

// AgentInfoHandler handles HTTP requests for retrieving agent consent information.
// This handler implements FR-009 (display agent metadata) and FR-025 (all services available).
type AgentInfoHandler struct {
	consentService *consent.Service
	logger         *slog.Logger
}

// NewAgentInfoHandler creates a new agent info handler.
func NewAgentInfoHandler(consentService *consent.Service, logger *slog.Logger) *AgentInfoHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentInfoHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// AgentMetadata represents agent metadata in the response.
type AgentMetadata struct {
	ID                   string  `json:"id"`
	ClientID             string  `json:"client_id"`
	DisplayName          string  `json:"display_name"`
	Description          string  `json:"description"`
	GovernanceURL        *string `json:"governance_url,omitempty"`
	UserDocumentationURL *string `json:"user_documentation_url,omitempty"`
	AgentInterfaceURL    *string `json:"agent_interface_url,omitempty"`
	CreatedAt            string  `json:"created_at"`
	UpdatedAt            string  `json:"updated_at"`
}

// ServiceScope represents a scope in the response.
type ServiceScope struct {
	ScopeValue  string `json:"scope_value"`
	Description string `json:"description"`
}

// RequestedService represents a third-party service in the response.
type RequestedService struct {
	ID          string         `json:"id"`
	DisplayName string         `json:"display_name"`
	Scopes      []ServiceScope `json:"scopes"`
}

// ServiceScopeInfo represents one service within a permission set for the consent-info response.
// Raw scopes are intentionally omitted; the frontend uses requirement_type to determine lock status (FR-007).
type ServiceScopeInfo struct {
	ServiceID       string `json:"service_id"`
	RequirementType string `json:"requirement_type"`
}

// PermissionSetWithRequirement represents a permission set with its requirement type (Phase 4 US2).
type PermissionSetWithRequirement struct {
	PermissionSet struct {
		ID            string             `json:"id"`
		Name          string             `json:"name"`
		Description   string             `json:"description"`
		ServiceScopes []ServiceScopeInfo `json:"service_scopes"`
	} `json:"permission_set"`
	RequirementType string `json:"requirement_type"`
}

// Service represents a third-party service (redacted) (Phase 4 US2).
type Service struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// ServiceRequirementInfo represents a service requirement entry for the consent screen.
// Used by the frontend to determine which services within a PS are mandatory vs optional (FR-008).
type ServiceRequirementInfo struct {
	ServiceID       string `json:"service_id"`
	RequirementType string `json:"requirement_type"` // "mandatory" or "optional"
}

// AgentConsentInfoResponse represents the response body for agent consent info.
// Phase 4 US2: Extended to include permission sets, active sessions, and available services.
type AgentConsentInfoResponse struct {
	Agent                   AgentMetadata                  `json:"agent"`
	RequestedServices       []RequestedService             `json:"requested_services"`
	PermissionSets          []PermissionSetWithRequirement `json:"permission_sets"`            // NEW: Phase 4 US2
	ActiveSessionServiceIDs []string                       `json:"active_session_service_ids"` // NEW: Phase 4 US2
	AvailableServices       []Service                      `json:"available_services"`         // NEW: Phase 4 US2
	ServiceRequirements     []ServiceRequirementInfo       `json:"service_requirements"`       // NEW: per-service toggle support (FR-008)
}

// GetAgentConsentInfo handles GET /api/consent/agent/:agent-id
// Returns agent metadata and all available third-party OAuth2 services.
// Phase 4 US2: Extended to include resolved permission sets and active sessions.
//
// Response codes:
// - 200 OK: Agent found, returns metadata + services
// - 400 Bad Request: Invalid agent ID format
// - 401 Unauthorized: Principal not found in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *AgentInfoHandler) GetAgentConsentInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.logger.Warn("agent ID is missing in request")
		h.writeError(w, http.StatusBadRequest, "agent ID is required", "")
		return
	}

	parsedAgentID, parseErr := id.ParseAgentID(agentID)
	if parseErr != nil {
		h.writeError(w, http.StatusBadRequest, "invalid agent ID", parseErr.Error())
		return
	}

	// Phase 4 US2: Extract principal from context (middleware ensures it's present)
	userID, ok := principal.FromContext(ctx)
	if !ok {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "principal required", "")
		return
	}

	// Call consent service to get agent info
	info, err := h.consentService.GetAgentConsentInfo(ctx, parsedAgentID, id.Principal(userID))
	if err != nil {
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found", "agent_id", agentID)
			h.writeError(w, http.StatusNotFound, "agent not found", "")
			return
		}

		h.logger.Error("failed to get agent consent info",
			"agent_id", agentID,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Build response
	resp := h.toResponse(info)

	h.logger.Info("agent consent info retrieved",
		"agent_id", agentID,
		"services_count", len(resp.RequestedServices))

	h.writeJSON(w, http.StatusOK, resp)
}

// toResponse converts AgentConsentInfo to AgentConsentInfoResponse.
// Phase 4 US2: Extended to serialize permission sets, active sessions, and available services.
func (h *AgentInfoHandler) toResponse(info *consent.AgentConsentInfo) AgentConsentInfoResponse {
	// Convert agent metadata
	agentMeta := AgentMetadata{
		ID:                   info.Agent.ID.String(),
		ClientID:             info.Agent.ClientID.String(),
		DisplayName:          info.Agent.DisplayName,
		Description:          info.Agent.Description,
		GovernanceURL:        info.Agent.GovernanceURL,
		UserDocumentationURL: info.Agent.UserDocumentationURL,
		AgentInterfaceURL:    info.Agent.AgentInterfaceURL,
		CreatedAt:            info.Agent.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:            info.Agent.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Convert requested services
	requestedServices := make([]RequestedService, len(info.AvailableThirdpartyServices))
	for i, svc := range info.AvailableThirdpartyServices {
		scopes := make([]ServiceScope, len(svc.Scopes))
		for j, scope := range svc.Scopes {
			scopes[j] = ServiceScope{
				ScopeValue:  scope.ScopeValue,
				Description: scope.Description,
			}
		}

		requestedServices[i] = RequestedService{
			ID:          svc.ID.String(),
			DisplayName: svc.DisplayName,
			Scopes:      scopes,
		}
	}

	// Phase 4 US2: Convert resolved permission sets
	permissionSets := make([]PermissionSetWithRequirement, 0, len(info.ResolvedPermissionSets))
	for _, entry := range info.ResolvedPermissionSets {
		if entry.PermissionSet == nil {
			h.logger.Warn("resolved permission set entry has nil PermissionSet, omitting from response")
			continue
		}
		serviceScopes := make([]ServiceScopeInfo, 0, len(entry.PermissionSet.ServiceScopes))
		for _, ss := range entry.PermissionSet.ServiceScopes {
			serviceScopes = append(serviceScopes, ServiceScopeInfo{
				ServiceID:       ss.ServiceID.String(),
				RequirementType: string(ss.RequirementType),
			})
		}
		item := PermissionSetWithRequirement{
			RequirementType: string(entry.RequirementType),
		}
		item.PermissionSet.ID = entry.PermissionSet.ID.String()
		item.PermissionSet.Name = entry.PermissionSet.Name
		item.PermissionSet.Description = entry.PermissionSet.Description
		item.PermissionSet.ServiceScopes = serviceScopes
		permissionSets = append(permissionSets, item)
	}

	// Phase 4 US2: Convert active session service IDs
	activeSessionServiceIDs := make([]string, len(info.ActiveSessionServiceIDs))
	for i, serviceID := range info.ActiveSessionServiceIDs {
		activeSessionServiceIDs[i] = serviceID.String()
	}

	// Phase 4 US2: Convert available services (redacted) — same slice as AvailableThirdpartyServices
	availableServices := make([]Service, len(info.AvailableThirdpartyServices))
	for i, svc := range info.AvailableThirdpartyServices {
		availableServices[i] = Service{
			ID:          svc.ID.String(),
			DisplayName: svc.DisplayName,
		}
	}

	// Convert service requirements — used by frontend to render per-service toggles (FR-008)
	serviceRequirements := make([]ServiceRequirementInfo, len(info.Agent.ServiceRequirements))
	for i, sr := range info.Agent.ServiceRequirements {
		serviceRequirements[i] = ServiceRequirementInfo{
			ServiceID:       sr.ServiceID.String(),
			RequirementType: string(sr.RequirementType),
		}
	}

	return AgentConsentInfoResponse{
		Agent:                   agentMeta,
		RequestedServices:       requestedServices,
		PermissionSets:          permissionSets,
		ActiveSessionServiceIDs: activeSessionServiceIDs,
		AvailableServices:       availableServices,
		ServiceRequirements:     serviceRequirements,
	}
}

// writeJSON writes a JSON response.
func (h *AgentInfoHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	if err := writeBufferedJSON(w, statusCode, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *AgentInfoHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
