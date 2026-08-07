package approval

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	domainapproval "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/approval"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// ConsumeHandler handles POST /api/approvals/{id}/consume.
type ConsumeHandler struct {
	service *domainapproval.Service
}

// NewConsumeHandler creates a handler for consuming approved approvals.
func NewConsumeHandler(service *domainapproval.Service) *ConsumeHandler {
	return &ConsumeHandler{service: service}
}

func (h *ConsumeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actingPrincipal, ok := principal.FromContext(r.Context())
	if !ok || actingPrincipal == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	rawID := chi.URLParam(r, "id")
	approvalID, err := id.ParseApprovalID(rawID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid approval ID format")
		return
	}

	result, err := h.service.ConsumeApproval(r.Context(), approvalID, id.Principal(actingPrincipal))
	if err != nil {
		if err == domainapproval.ErrApprovalNotConsumable {
			writeError(w, http.StatusUnprocessableEntity, "not_consumable", "only once-persistence approved approvals can be consumed")
			return
		}
		mapDomainError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": map[string]any{
			"id":          result.ID,
			"consumed":    result.Consumed,
			"consumed_at": result.ConsumedAt,
		},
	})
}
