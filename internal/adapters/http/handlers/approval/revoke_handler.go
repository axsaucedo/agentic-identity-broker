package approval

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	domainapproval "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/approval"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// RevokeHandler handles POST /api/approvals/{id}/revoke.
type RevokeHandler struct {
	service *domainapproval.Service
}

// NewRevokeHandler creates a handler for revoking permanent approvals.
func NewRevokeHandler(service *domainapproval.Service) *RevokeHandler {
	return &RevokeHandler{service: service}
}

func (h *RevokeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract principal from X-Remote-User
	actingPrincipal, ok := principal.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing principal")
		return
	}

	// Parse approval ID
	rawID := chi.URLParam(r, "id")
	approvalID, err := id.ParseApprovalID(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid approval ID format")
		return
	}

	result, err := h.service.RevokePermanentApproval(r.Context(), approvalID, id.Principal(actingPrincipal))
	if err != nil {
		if err == domainapproval.ErrApprovalNotRevocable {
			writeError(w, http.StatusUnprocessableEntity, "not_revocable", "only permanent approvals can be revoked")
			return
		}
		mapDomainError(w, err)
		return
	}

	respData := map[string]any{
		"id":        result.ID,
		"status":    result.Status,
		"denied_at": result.DeniedAt,
	}
	if result.Persistence != nil {
		respData["persistence"] = *result.Persistence
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": respData})
}
