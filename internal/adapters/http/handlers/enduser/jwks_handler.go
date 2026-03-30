package enduser

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
)

// JWKSHandler serves the JWKS endpoint for public key discovery.
type JWKSHandler struct {
	signingKeyService *oauth2server.SigningKeyService
	logger            *slog.Logger
}

// NewJWKSHandler creates a new JWKSHandler.
func NewJWKSHandler(
	signingKeyService *oauth2server.SigningKeyService,
	logger *slog.Logger,
) *JWKSHandler {
	return &JWKSHandler{
		signingKeyService: signingKeyService,
		logger:            logger,
	}
}

// ServeJWKS returns the JSON Web Key Set with active public keys.
// GET /oauth2/jwks.json
func (h *JWKSHandler) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	jwks, err := h.signingKeyService.BuildJWKS(r.Context())
	if err != nil {
		h.logger.Error("failed to build JWKS", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		h.logger.Error("failed to encode JWKS response", "error", err)
	}
}
