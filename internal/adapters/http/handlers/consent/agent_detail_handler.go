// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/go-chi/chi/v5"
)

// AgentDetailHandler handles HTTP requests for retrieving detailed agent information.
// Implements User Story 2: Review Agent-Specific Grants (GET /api/consent/agent/:agentId).
type AgentDetailHandler struct {
	consentService ConsentService
	logger         *slog.Logger
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

// GetAgentDetailResponse represents the response for GET /api/consent/agent/:agentId.
type GetAgentDetailResponse struct {
	Data AgentDetailData `json:"data"`
}

// AgentDetailData contains the agent detail and associated services.
type AgentDetailData struct {
	Agent    consent.AgentDetail         `json:"agent"`
	Services []consent.ThirdpartyService `json:"services"`
}

// GetAgentDetail handles GET /api/consent/agent/:agentId
// Returns detailed information about an agent and all available third-party services.
//
// Response codes:
// - 200 OK: Returns agent details and services
// - 400 Bad Request: Invalid agent ID format
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

	// Call consent service to get agent detail
	agentDetail, services, err := h.consentService.GetAgentDetail(ctx, agentID)
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

	// Build response
	response := GetAgentDetailResponse{
		Data: AgentDetailData{
			Agent:    *agentDetail,
			Services: services,
		},
	}

	h.logger.Info("agent detail retrieved",
		"agent_id", agentID,
		"services_count", len(services))

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
