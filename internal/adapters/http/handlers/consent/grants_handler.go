// Package consent provides HTTP handlers for consent management APIs.
package consent

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
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

	parsedAgentID, err := id.ParseAgentID(agentID)
	if err != nil {
		h.logger.Warn("invalid agent ID format", "agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID must be a valid UUID")
		return
	}

	ctx := r.Context()
	grants, err := h.consentService.GetActiveGrants(ctx, id.Principal(principalValue), parsedAgentID)
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

	parsedAgentID, err := id.ParseAgentID(agentID)
	if err != nil {
		h.logger.Warn("invalid agent ID format", "agent_id", agentID)
		h.writeError(w, http.StatusBadRequest, "bad request", "agent ID must be a valid UUID")
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
		err := h.consentService.RevokeConsent(r.Context(), id.Principal(principalValue), parsedAgentID)
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

	// Check for redirect_uri parameter early - validate before processing grant (T051-T054)
	redirectURI := r.URL.Query().Get("redirect_uri")
	if redirectURI != "" {
		// Validate redirect_uri early to prevent unnecessary processing
		valid, err := validateRedirectURI(redirectURI, r)
		if err != nil {
			h.logger.Warn("malformed redirect_uri",
				"redirect_uri", redirectURI,
				"principal", principalValue,
				"agent_id", agentID,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "invalid redirect_uri", fmt.Sprintf("redirect_uri format is invalid: %v", err))
			return
		}

		if !valid {
			// External domain - reject (T054)
			h.logger.Warn("redirect_uri to external domain rejected",
				"redirect_uri", redirectURI,
				"request_host", r.Host,
				"principal", principalValue,
				"agent_id", agentID)
			h.writeError(w, http.StatusBadRequest, "invalid redirect_uri", "redirect_uri must be same-origin or relative")
			return
		}
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
		parsedServiceID, err := id.ParseServiceID(token.ThirdpartyOAuth2ServiceID)
		if err != nil {
			h.logger.Warn("invalid service ID format in grant request",
				"service_id", token.ThirdpartyOAuth2ServiceID,
				"principal", principalValue,
				"agent_id", agentID)
			h.writeError(w, http.StatusBadRequest, "invalid request", fmt.Sprintf("service ID %q must be a valid UUID", token.ThirdpartyOAuth2ServiceID))
			return
		}
		tokens[i] = storage.DelegatedToken{
			ThirdpartyOAuth2ServiceID: parsedServiceID,
			Scopes:                    token.Scopes,
		}
	}

	// Create grant request
	grantReq := &consent.GrantRequest{
		Principal:             id.Principal(principalValue),
		AgentID:               parsedAgentID,
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

	// If redirect_uri was provided and already validated, return redirect URL in response body
	// instead of HTTP 303 redirect (T056, T057). This avoids CORS issues when the redirect chain
	// includes cross-origin redirects (e.g., to upstream OAuth2 server).
	if redirectURI != "" {
		// Defensive re-validation of redirect_uri at the sink to prevent open redirects.
		// Normalize backslashes to forward slashes before parsing to avoid browser quirks.
		normalizedRedirectURI := strings.ReplaceAll(redirectURI, "\\", "/")

		target, err := url.Parse(normalizedRedirectURI)
		if err != nil {
			h.logger.Warn("malformed redirect_uri at redirect time",
				"redirect_uri", redirectURI,
				"principal", principalValue,
				"agent_id", agentID,
				"error", err)
			h.writeError(w, http.StatusBadRequest, "invalid redirect_uri", fmt.Sprintf("redirect_uri format is invalid: %v", err))
			return
		}

		// Allow only relative URLs or same-origin absolute URLs.
		// Derive the request host from the HTTP Host header, not from r.URL, which may be empty.
		var requestHost string
		if r.Host != "" {
			// Prepend a dummy scheme so we can reliably parse the host.
			if u, parseErr := url.Parse("http://" + r.Host); parseErr == nil {
				requestHost = u.Hostname()
			}
		}
		targetHost := target.Hostname()
		if targetHost != "" && targetHost != requestHost {
			h.logger.Warn("redirect_uri to external domain rejected at redirect time",
				"redirect_uri", redirectURI,
				"target_host", targetHost,
				"request_host", requestHost,
				"principal", principalValue,
				"agent_id", agentID)
			h.writeError(w, http.StatusBadRequest, "invalid redirect_uri", "redirect_uri must be same-origin or relative")
			return
		}

		// Return redirect URL in response body instead of HTTP redirect.
		// The frontend will use window.location.href to navigate, which properly handles
		// cross-origin redirects that would otherwise cause CORS errors with XMLHttpRequest.
		h.logger.Info("returning redirect URL after grant approval",
			"redirect_uri", normalizedRedirectURI,
			"principal", principalValue,
			"agent_id", agentID)
		response := h.toGrantResponse(grant)
		h.writeJSON(w, http.StatusCreated, map[string]interface{}{
			"data":         response,
			"redirect_url": normalizedRedirectURI,
		})
		return
	}

	// No redirect_uri: return success response (T057: Display success confirmation)
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
			ThirdpartyOAuth2ServiceID: token.ThirdpartyOAuth2ServiceID.String(),
			Scopes:                    token.Scopes,
		}
	}

	return GrantResponse{
		ID:                    grant.ID.String(),
		Principal:             grant.Principal.String(),
		AgentID:               grant.AgentID.String(),
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

// =========================================================================
// Helper Functions for Redirect URI Validation (User Story 6)
// =========================================================================

// validateRedirectURI validates that a redirect_uri is safe for redirection.
// It enforces same-origin policy: allows relative URLs and same-origin absolute URLs.
// Returns (valid, error):
// - (true, nil): redirect_uri is valid (relative or same-origin absolute)
// - (false, nil): redirect_uri is not valid (external domain)
// - (false, error): redirect_uri format is invalid (malformed URL)
//
// Per FR-026, FR-027: System MUST validate redirect_uri is same-origin or relative before redirecting
func validateRedirectURI(redirectURI string, r *http.Request) (bool, error) {
	// T049: Test case 8 - Empty redirect_uri is allowed
	if redirectURI == "" {
		return true, nil
	}

	// T049: Test case 1, 2 - Relative URLs (no scheme) are always allowed (T053)
	parsedURL, err := url.Parse(redirectURI)
	if err != nil {
		// T049: Test case 11 - Malformed URL returns error
		return false, fmt.Errorf("failed to parse redirect_uri: %w", err)
	}

	// If no scheme, it's relative - always allowed (T053)
	if parsedURL.Scheme == "" {
		return true, nil
	}

	// Absolute URL - perform same-origin check (T052)
	return isSameOrigin(parsedURL, r), nil
}

// isSameOrigin checks if a parsed URL has the same origin as the current request.
// Compares scheme, host, and port.
// Handles port normalization: http default 80, https default 443.
// SECURITY: Uses r.TLS to detect scheme (not hostname heuristics).
func isSameOrigin(u *url.URL, r *http.Request) bool {
	// Determine request scheme from TLS connection or X-Forwarded-Proto header
	// This is more reliable than trying to infer from hostname (prevents open redirect)
	requestScheme := "http"
	if r.TLS != nil {
		requestScheme = "https"
	}
	// Fallback for reverse proxy scenarios where TLS is terminated upstream
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		requestScheme = proto
	}

	// Parse request origin with correct scheme
	requestURL, err := url.Parse(fmt.Sprintf("%s://%s", requestScheme, r.Host))
	if err != nil {
		// If we can't parse the request host, it's not same-origin
		return false
	}

	// Compare scheme (T049: Test case 6 - Different scheme is not same-origin)
	if u.Scheme != requestURL.Scheme {
		return false
	}

	// Get normalized hosts and ports
	uHost := u.Hostname()
	reqHost := requestURL.Hostname()
	if uHost != reqHost {
		// T049: Test case 5 - Different domain is not same-origin
		return false
	}

	// Get ports with normalization
	uPort := normalizePort(u.Port(), u.Scheme)
	reqPort := normalizePort(requestURL.Port(), requestURL.Scheme)
	if uPort != reqPort {
		// T049: Test case 7 - Different port is not same-origin
		return false
	}

	// T049: Test case 3, 4, 9, 10 - Same-origin is valid
	return true
}

// normalizePort returns the port number, applying defaults for well-known schemes.
// http defaults to 80, https defaults to 443.
func normalizePort(port string, scheme string) string {
	if port != "" {
		return port
	}

	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}
