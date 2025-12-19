// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// GrantsHandler handles HTTP requests for user grants management.
// Implements FR-011 through FR-014 (grant CRUD operations).
type GrantsHandler struct {
	consentService ConsentService
	logger         *slog.Logger
}

// NewGrantsHandler creates a new grants handler.
func NewGrantsHandler(consentService ConsentService, logger *slog.Logger) *GrantsHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &GrantsHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// DelegatedTokenRequest represents a delegated token in the request.
type DelegatedTokenRequest struct {
	ThirdpartyOAuth2ServiceID string   `json:"thirdparty_oauth2_service_id"`
	Scopes                    []string `json:"scopes"`
}

// GrantRequest represents the request body for creating/updating a grant.
type GrantRequest struct {
	ValidUntil            *time.Time              `json:"valid_until,omitempty"`
	DelegatedOAuth2Tokens []DelegatedTokenRequest `json:"delegated_oauth2_tokens"`
}

// GrantResponse represents a user grant in the response.
type GrantResponse struct {
	ID                    string                  `json:"id"`
	Principal             string                  `json:"principal"`
	AgentID               string                  `json:"agent_id"`
	ValidUntil            *time.Time              `json:"valid_until,omitempty"`
	DelegatedOAuth2Tokens []DelegatedTokenRequest `json:"delegated_oauth2_tokens"`
	CreatedAt             string                  `json:"created_at"`
	UpdatedAt             string                  `json:"updated_at"`
}

// GetGrants handles GET /api/consent/agent/:agent-id/grants
// Returns all active grants for the authenticated principal and specified agent.
//
// Response codes:
// - 200 OK: Returns grants (may be empty array)
// - 401 Unauthorized: No principal in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *GrantsHandler) GetGrants(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agent-id")

	// Extract principal from context
	principalValue, ok := principal.FromContext(r.Context())
	if !ok || principalValue == "" {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	ctx := r.Context()
	grants, err := h.consentService.GetActiveGrants(ctx, principalValue, agentID)
	if err != nil {
		// Check if it's an agent not found error
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found",
				"agent_id", agentID,
				"principal", principalValue)
			h.writeError(w, http.StatusNotFound, "agent not found", "")
			return
		}

		h.logger.Error("failed to get grants",
			"agent_id", agentID,
			"principal", principalValue,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Convert to response format
	response := make([]GrantResponse, len(grants))
	for i, grant := range grants {
		response[i] = h.toGrantResponse(grant)
	}

	h.logger.Info("grants retrieved",
		"agent_id", agentID,
		"principal", principalValue,
		"count", len(response))

	// Wrap in data envelope to match frontend expectations
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"data": response,
	})
}

// CreateGrant handles POST /api/consent/agent/:agent-id/grants
// Creates or updates a grant (upsert semantics).
// Special case: empty delegated_oauth2_tokens array = revoke.
//
// Response codes:
// - 201 Created: Grant created/updated
// - 204 No Content: Grant revoked (empty tokens)
// - 400 Bad Request: Invalid request body or validation error
// - 401 Unauthorized: No principal in context
// - 404 Not Found: Agent doesn't exist
// - 500 Internal Server Error: Service error
func (h *GrantsHandler) CreateGrant(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agent-id")

	// Extract principal from context
	principalValue, ok := principal.FromContext(r.Context())
	if !ok || principalValue == "" {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	// Parse request
	var req GrantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body",
			"error", err,
			"principal", principalValue,
			"agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "invalid request", "request body must be valid JSON")
		return
	}

	// Special case: empty tokens = revoke
	if len(req.DelegatedOAuth2Tokens) == 0 {
		err := h.consentService.RevokeConsent(r.Context(), principalValue, agentID)
		if err != nil {
			h.logger.Error("failed to revoke consent",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusInternalServerError, "internal server error", "")
			return
		}

		h.logger.Info("grant revoked",
			"principal", principalValue,
			"agent_id", agentID)

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Validate valid_until is in future
	if req.ValidUntil != nil && req.ValidUntil.Before(time.Now()) {
		h.logger.Warn("valid_until is in the past",
			"valid_until", req.ValidUntil,
			"principal", principalValue,
			"agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "invalid request", "valid_until must be in the future")
		return
	}

	// Convert request to domain tokens
	tokens := make([]storage.DelegatedToken, len(req.DelegatedOAuth2Tokens))
	for i, token := range req.DelegatedOAuth2Tokens {
		tokens[i] = storage.DelegatedToken{
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID,
			Scopes:                    token.Scopes,
		}
	}

	// Create grant request
	grantReq := &consent.GrantRequest{
		Principal:             principalValue,
		AgentID:               agentID,
		ValidUntil:            req.ValidUntil,
		DelegatedOAuth2Tokens: tokens,
	}

	// Call consent service
	grant, err := h.consentService.GrantConsent(r.Context(), grantReq)
	if err != nil {
		// Check for specific error types
		if errors.Is(err, consent.ErrAgentNotFound) {
			h.logger.Warn("agent not found",
				"agent_id", agentID,
				"principal", principalValue)
			h.writeError(w, http.StatusNotFound, "agent not found", "")
			return
		}

		if errors.Is(err, consent.ErrInvalidScopes) {
			h.logger.Warn("invalid scopes requested",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "invalid scopes", err.Error())
			return
		}

		if errors.Is(err, consent.ErrServiceNotFound) {
			h.logger.Warn("service not found",
				"agent_id", agentID,
				"principal", principalValue,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "service not found", err.Error())
			return
		}

		h.logger.Error("failed to grant consent",
			"agent_id", agentID,
			"principal", principalValue,
			"error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Audit logging
	h.logger.Info("grant created",
		"principal", principalValue,
		"agent_id", agentID,
		"grant_id", grant.ID)

	response := h.toGrantResponse(grant)
	// Wrap in data envelope to match frontend expectations
	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"data": response,
	})
}

// toGrantResponse converts a UserGrant to GrantResponse.
func (h *GrantsHandler) toGrantResponse(grant *storage.UserGrant) GrantResponse {
	tokens := make([]DelegatedTokenRequest, len(grant.DelegatedOAuth2Tokens))
	for i, token := range grant.DelegatedOAuth2Tokens {
		tokens[i] = DelegatedTokenRequest{
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID,
			Scopes:                    token.Scopes,
		}
	}

	return GrantResponse{
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
func (h *GrantsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *GrantsHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
