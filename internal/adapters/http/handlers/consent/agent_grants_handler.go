// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// AgentGrantsHandler handles HTTP requests for retrieving user grants for a specific agent.
// Implements User Story 2: Review Agent-Specific Grants (GET /api/consent/agent/:agentId/grants).
type AgentGrantsHandler struct {
	consentService ConsentService
	logger         *slog.Logger
}

// NewAgentGrantsHandler creates a new agent grants handler.
func NewAgentGrantsHandler(consentService ConsentService, logger *slog.Logger) *AgentGrantsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &AgentGrantsHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// UserGrantDTO represents a user grant in the response.
type UserGrantDTO struct {
	ID                    string                `json:"id"`
	Principal             string                `json:"principal"`
	AgentID               string                `json:"agent_id"`
	ValidUntil            *time.Time            `json:"valid_until,omitempty"`
	DelegatedOAuth2Tokens []DelegatedTokenDTO   `json:"delegated_oauth2_tokens"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
}

// DelegatedTokenDTO represents a delegated token in the response.
type DelegatedTokenDTO struct {
	ThirdpartyOAuth2ServiceID string   `json:"thirdparty_oauth2_service_id"`
	Scopes                    []string `json:"scopes"`
}

// GetAgentGrantsResponse represents the response for GET /api/consent/agent/:agentId/grants.
type GetAgentGrantsResponse struct {
	Data []UserGrantDTO `json:"data"`
}

// GetAgentGrants handles GET /api/consent/agent/:agentId/grants
// Returns all grants the authenticated user has granted to the specified agent.
//
// Response codes:
// - 200 OK: Returns grants (may be empty array if no grants exist)
// - 400 Bad Request: Invalid agent ID format
// - 401 Unauthorized: No principal in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *AgentGrantsHandler) GetAgentGrants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	agentID := chi.URLParam(r, "agent-id")

	if agentID == "" {
		h.logger.Warn("agent ID is missing in request")
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID is required")
		return
	}

	// Extract principal from request context
	principalValue, ok := principal.FromContext(ctx)
	if !ok || principalValue == "" {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	// Call consent service to get user grants
	grants, err := h.consentService.GetUserGrants(ctx, principalValue, agentID)
	if err != nil {
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found",
				"agent_id", agentID,
				"principal", principalValue)
			h.writeError(w, http.StatusNotFound, "not found", "agent not found")
			return
		}

		h.logger.Error("failed to get user grants",
			"agent_id", agentID,
			"principal", principalValue,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Convert grants to DTOs
	grantDTOs := make([]UserGrantDTO, len(grants))
	for i, grant := range grants {
		grantDTOs[i] = h.toUserGrantDTO(grant)
	}

	// Return response (empty array if no grants)
	h.logger.Info("user grants retrieved",
		"agent_id", agentID,
		"principal", principalValue,
		"count", len(grantDTOs))

	response := GetAgentGrantsResponse{
		Data: grantDTOs,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// toUserGrantDTO converts a storage.UserGrant to UserGrantDTO.
func (h *AgentGrantsHandler) toUserGrantDTO(grant *storage.UserGrant) UserGrantDTO {
	tokens := make([]DelegatedTokenDTO, len(grant.DelegatedOAuth2Tokens))
	for i, token := range grant.DelegatedOAuth2Tokens {
		tokens[i] = DelegatedTokenDTO{
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID,
			Scopes:                    token.Scopes,
		}
	}

	return UserGrantDTO{
		ID:                    grant.ID,
		Principal:             grant.Principal,
		AgentID:               grant.AgentID,
		ValidUntil:            grant.ValidUntil,
		DelegatedOAuth2Tokens: tokens,
		CreatedAt:             grant.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             grant.UpdatedAt.Format(time.RFC3339),
	}
}

// writeJSON writes a JSON response.
func (h *AgentGrantsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := encodeJSON(w, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *AgentGrantsHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
