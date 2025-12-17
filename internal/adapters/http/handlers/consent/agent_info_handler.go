// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
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

// AgentConsentInfoResponse represents the response body for agent consent info.
type AgentConsentInfoResponse struct {
	Agent             AgentMetadata      `json:"agent"`
	RequestedServices []RequestedService `json:"requested_services"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// GetAgentConsentInfo handles GET /api/consent/agent/:agent-id
// Returns agent metadata and all available third-party OAuth2 services.
//
// Response codes:
// - 200 OK: Agent found, returns metadata + services
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

	// Call consent service to get agent info
	info, err := h.consentService.GetAgentConsentInfo(ctx, agentID)
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
func (h *AgentInfoHandler) toResponse(info *consent.AgentConsentInfo) AgentConsentInfoResponse {
	// Convert agent metadata
	agentMeta := AgentMetadata{
		ID:                   info.Agent.ID,
		ClientID:             info.Agent.ClientID,
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
			ID:          svc.ID,
			DisplayName: svc.DisplayName,
			Scopes:      scopes,
		}
	}

	return AgentConsentInfoResponse{
		Agent:             agentMeta,
		RequestedServices: requestedServices,
	}
}

// writeJSON writes a JSON response.
func (h *AgentInfoHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
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
