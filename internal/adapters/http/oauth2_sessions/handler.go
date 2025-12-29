package oauth2_sessions

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Handler handles HTTP requests for OAuth2 sessions.
type Handler struct {
	sessionRepo ports.UserSessionRepository
	grantRepo   ports.UserGrantRepository
	service     *oauth2session.OAuth2SessionService
	logger      *slog.Logger
}

// NewHandler creates a new handler.
func NewHandler(
	sessionRepo ports.UserSessionRepository,
	grantRepo ports.UserGrantRepository,
	service *oauth2session.OAuth2SessionService,
) *Handler {
	return &Handler{
		sessionRepo: sessionRepo,
		grantRepo:   grantRepo,
		service:     service,
		logger:      slog.Default(),
	}
}

// ListSessions handles GET /api/third-party/sessions
func (h *Handler) ListSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract principal from header
	principal := r.Header.Get("X-Remote-User")
	if principal == "" {
		h.logger.Warn("missing X-Remote-User header")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "unauthorized",
			"message": "missing X-Remote-User header",
		})
		return
	}

	// Return empty sessions for now if service not initialized
	// TODO: Initialize OAuth2SessionService properly when encryption keys are available
	if h.service == nil {
		h.logger.Debug("OAuth2 session service not initialized", "principal", principal)
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"sessions": []interface{}{},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Fetch sessions
	summaries, err := h.service.ListUserSessions(ctx, principal)
	if err != nil {
		h.logger.Error("failed to list sessions", "principal", principal, "err", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "internal_error",
			"message": "failed to list sessions",
		})
		return
	}

	h.logger.Info("listed sessions", "principal", principal, "count", len(summaries))

	// Return response
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"sessions": summaries,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// RegisterRoutes registers all OAuth2 session routes with the router.
// Routes are relative to /api (e.g., "/third-party/sessions" becomes "/api/third-party/sessions")
func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/third-party/sessions", h.ListSessions)
}
