package approval

import (
	"net/http"

	domainapproval "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/approval"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// PendingHandler handles GET /api/approvals/pending.
type PendingHandler struct {
	service *domainapproval.Service
}

// NewPendingHandler creates a handler for listing pending approvals.
func NewPendingHandler(service *domainapproval.Service) *PendingHandler {
	return &PendingHandler{service: service}
}

func (h *PendingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actingPrincipal, ok := principal.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing principal")
		return
	}

	approvals, err := h.service.ListPendingApprovals(r.Context(), id.Principal(actingPrincipal))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list pending approvals")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": approvals,
	})
}
