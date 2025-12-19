// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// UserInfoHandler handles HTTP requests for retrieving user information.
// Implements GET /api/me endpoint.
type UserInfoHandler struct {
	logger *slog.Logger
}

// NewUserInfoHandler creates a new user info handler.
func NewUserInfoHandler(logger *slog.Logger) *UserInfoHandler {
	if logger == nil {
		logger = slog.Default()
	}
	return &UserInfoHandler{
		logger: logger,
	}
}

// GetUserInfoResponse represents the response for GET /api/me.
type GetUserInfoResponse struct {
	Data consent.UserInfo `json:"data"`
}

// GetUserInfo handles GET /api/me
// Returns information about the currently authenticated user.
//
// Response codes:
// - 200 OK: Returns user info
// - 401 Unauthorized: No principal in context
func (h *UserInfoHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract principal from request context
	principalValue, ok := principal.FromContext(ctx)
	if !ok || principalValue == "" {
		h.logger.Warn("principal not found in context")
		h.writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	// Construct UserInfo from principal
	// For now, we use the principal as the display name
	// In a future implementation, this could be enriched with user profile data
	userInfo := consent.UserInfo{
		Principal:   principalValue,
		DisplayName: principalValue, // Use principal as display name for now
		PictureURL:  nil,             // No picture URL available yet
	}

	h.logger.Info("user info retrieved",
		"principal", principalValue)

	response := GetUserInfoResponse{
		Data: userInfo,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// writeJSON writes a JSON response.
func (h *UserInfoHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := encodeJSON(w, data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

// writeError writes an error response.
func (h *UserInfoHandler) writeError(w http.ResponseWriter, statusCode int, error string, message string) {
	resp := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.writeJSON(w, statusCode, resp)
}
