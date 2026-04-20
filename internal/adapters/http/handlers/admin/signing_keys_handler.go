package admin

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// SigningKeysHandler handles admin API requests for signing key management.
type SigningKeysHandler struct {
	signingKeyRepo    ports.SigningKeyRepository
	signingKeyService *oauth2server.SigningKeyService
	logger            *slog.Logger
}

// NewSigningKeysHandler creates a new SigningKeysHandler.
func NewSigningKeysHandler(
	signingKeyRepo ports.SigningKeyRepository,
	signingKeyService *oauth2server.SigningKeyService,
	logger *slog.Logger,
) *SigningKeysHandler {
	return &SigningKeysHandler{
		signingKeyRepo:    signingKeyRepo,
		signingKeyService: signingKeyService,
		logger:            logger,
	}
}

type signingKeyResponse struct {
	KID         string  `json:"kid"`
	Algorithm   string  `json:"algorithm"`
	IsCurrent   bool    `json:"is_current"`
	ActivatesAt string  `json:"activates_at"`
	CreatedAt   string  `json:"created_at"`
	RemovedAt   *string `json:"removed_at,omitempty"`
}

type signingKeyAddRequest struct {
	Algorithm string `json:"algorithm"`
}

// Add generates and stores a new signing key.
// POST /api/oauth2-server/signing-keys
func (h *SigningKeysHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req signingKeyAddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if !errors.Is(err, io.EOF) {
			h.writeError(w, http.StatusBadRequest, "invalid request body", "")
			return
		}
		req.Algorithm = "ES256"
	}
	if req.Algorithm == "" {
		req.Algorithm = "ES256"
	}
	if req.Algorithm != "ES256" {
		h.writeError(w, http.StatusBadRequest, "unsupported algorithm", "algorithm must be ES256")
		return
	}

	key, err := h.signingKeyService.GenerateAndStoreKey(r.Context(), req.Algorithm, true)
	if err != nil {
		h.logger.Error("failed to add signing key", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	resp := signingKeyResponse{
		KID:         key.KID.String(),
		Algorithm:   key.Algorithm,
		IsCurrent:   key.IsCurrent,
		ActivatesAt: key.ActivatesAt.Format(time.RFC3339),
		CreatedAt:   key.CreatedAt.Format(time.RFC3339),
	}

	h.logger.Info("signing key added", "kid", key.KID, "algorithm", key.Algorithm, "activates_at", key.ActivatesAt)
	h.writeJSON(w, http.StatusCreated, resp)
}

// List returns all active signing keys (without private material).
// GET /api/oauth2-server/signing-keys
func (h *SigningKeysHandler) List(w http.ResponseWriter, r *http.Request) {
	keys, err := h.signingKeyRepo.ListActive(r.Context())
	if err != nil {
		h.logger.Error("failed to list signing keys", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	items := make([]signingKeyResponse, 0, len(keys))
	for _, key := range keys {
		items = append(items, signingKeyResponse{
			KID:         key.KID.String(),
			Algorithm:   key.Algorithm,
			IsCurrent:   key.IsCurrent,
			ActivatesAt: key.ActivatesAt.Format(time.RFC3339),
			CreatedAt:   key.CreatedAt.Format(time.RFC3339),
		})
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

// SetCurrent promotes a signing key to be the current key.
// PUT /api/oauth2-server/signing-keys/{kid}/current
func (h *SigningKeysHandler) SetCurrent(w http.ResponseWriter, r *http.Request) {
	kid := chi.URLParam(r, "kid")
	if kid == "" {
		h.writeError(w, http.StatusBadRequest, "kid is required", "")
		return
	}

	if err := h.signingKeyRepo.SetCurrent(r.Context(), id.NewKeyID(kid)); err != nil {
		if isNotFoundErr(err) {
			h.writeError(w, http.StatusNotFound, "signing key not found", "")
			return
		}
		h.logger.Error("failed to set current key", "kid", kid, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	key, err := h.signingKeyRepo.GetByKID(r.Context(), id.NewKeyID(kid))
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	resp := signingKeyResponse{
		KID:         key.KID.String(),
		Algorithm:   key.Algorithm,
		IsCurrent:   key.IsCurrent,
		ActivatesAt: key.ActivatesAt.Format(time.RFC3339),
		CreatedAt:   key.CreatedAt.Format(time.RFC3339),
	}

	h.logger.Info("signing key promoted to current", "kid", kid)
	h.writeJSON(w, http.StatusOK, resp)
}

// Remove soft-deletes a signing key.
// DELETE /api/oauth2-server/signing-keys/{kid}
func (h *SigningKeysHandler) Remove(w http.ResponseWriter, r *http.Request) {
	kid := chi.URLParam(r, "kid")
	if kid == "" {
		h.writeError(w, http.StatusBadRequest, "kid is required", "")
		return
	}

	// Check if this is the last active key
	count, err := h.signingKeyRepo.CountActive(r.Context())
	if err != nil {
		h.logger.Error("failed to count active keys", "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	// Check if the key is current
	key, err := h.signingKeyRepo.GetByKID(r.Context(), id.NewKeyID(kid))
	if err != nil {
		if isNotFoundErr(err) {
			h.writeError(w, http.StatusNotFound, "signing key not found", "")
			return
		}
		h.logger.Error("failed to get signing key", "kid", kid, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	if count <= 1 {
		h.writeError(w, http.StatusConflict, "cannot remove the last signing key", "")
		return
	}
	if key.IsCurrent {
		h.writeError(w, http.StatusConflict, "cannot remove the current signing key; promote another key first", "")
		return
	}

	if err := h.signingKeyRepo.Delete(r.Context(), id.NewKeyID(kid)); err != nil {
		if isNotFoundErr(err) {
			h.writeError(w, http.StatusNotFound, "signing key not found", "")
			return
		}
		h.logger.Error("failed to remove signing key", "kid", kid, "error", err)
		h.writeError(w, http.StatusInternalServerError, "internal server error", "")
		return
	}

	h.logger.Info("signing key removed", "kid", kid)
	w.WriteHeader(http.StatusNoContent)
}

func (h *SigningKeysHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *SigningKeysHandler) writeError(w http.ResponseWriter, statusCode int, errMsg string, message string) {
	resp := ErrorResponse{
		Error:   errMsg,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
