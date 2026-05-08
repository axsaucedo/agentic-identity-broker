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

// RevokeGrantHandler handles DELETE /api/consent/agent/{agent-id}/grants (FR-014).
type RevokeGrantHandler struct {
	consentService ConsentService
	logger         *slog.Logger
}

// NewRevokeGrantHandler creates a new RevokeGrantHandler.
func NewRevokeGrantHandler(consentService ConsentService, logger *slog.Logger) *RevokeGrantHandler {
	return &RevokeGrantHandler{
		consentService: consentService,
		logger:         logger,
	}
}

// RevokeGrant handles DELETE /api/consent/agent/{agent-id}/grants.
// Security rules (SR-001): principal check (401) before UUID parse (400).
// Returns 204 No Content on success, 404 if grant not found, 500 on service error.
func (h *RevokeGrantHandler) RevokeGrant(w http.ResponseWriter, r *http.Request) {
	// SR-001: principal check BEFORE UUID format validation.
	principalValue, ok := principal.FromContext(r.Context())
	if !ok || principalValue == "" {
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	rawAgentID := chi.URLParam(r, "agent-id")
	agentID, err := id.ParseAgentID(rawAgentID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID must be a valid UUID")
		return
	}

	if err := h.consentService.RevokeConsentForPrincipal(r.Context(), id.Principal(principalValue), agentID); err != nil {
		if errors.Is(err, consent.ErrGrantNotFound) {
			h.writeError(w, http.StatusNotFound, "not found", "no active grant exists for this agent")
			return
		}
		// SR-003: fail closed — do not expose internal error details.
		h.writeError(w, http.StatusInternalServerError, "internal server error", "an unexpected error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RevokeGrantHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	if err := writeBufferedJSON(w, statusCode, data); err != nil && h.logger != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *RevokeGrantHandler) writeError(w http.ResponseWriter, statusCode int, errMsg string, message string) {
	h.writeJSON(w, statusCode, ErrorResponse{Error: errMsg, Message: message})
}
