package approval

import (
	"net/http"

	domainapproval "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/approval"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// PermanentHandler handles GET /api/approvals/permanent.
type PermanentHandler struct {
	service *domainapproval.Service
}

// NewPermanentHandler creates a handler for listing permanent approvals.
func NewPermanentHandler(service *domainapproval.Service) *PermanentHandler {
	return &PermanentHandler{service: service}
}

func (h *PermanentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract principal from X-Remote-User
	actingPrincipal, ok := principal.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing principal")
		return
	}

	approvals, err := h.service.ListPermanentApprovals(r.Context(), id.Principal(actingPrincipal))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list permanent approvals")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"data": approvals,
	})
}
