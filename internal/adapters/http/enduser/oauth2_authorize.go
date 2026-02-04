package enduser

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2AuthorizeHandler handles OAuth2 authorization endpoint requests
type OAuth2AuthorizeHandler struct {
	Service           ports.OAuth2Service
	sessionRepository ports.UserSessionRepository
	logger            *slog.Logger
}

// ServeHTTP implements http.Handler for the authorization endpoint
func (h *OAuth2AuthorizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract principal from context (guaranteed by RequirePrincipalMiddleware)
	principalValue := principal.MustFromContext(r.Context())

	// Parse query parameters
	query := r.URL.Query()

	// Extract required OAuth2 parameters
	clientID := query.Get("client_id")
	if clientID == "" {
		http.Error(w, "missing required parameter: client_id", http.StatusBadRequest)
		return
	}

	redirectURI := query.Get("redirect_uri")
	if redirectURI == "" {
		http.Error(w, "missing required parameter: redirect_uri", http.StatusBadRequest)
		return
	}

	responseType := query.Get("response_type")
	if responseType == "" {
		http.Error(w, "missing required parameter: response_type", http.StatusBadRequest)
		return
	}

	// Extract optional OAuth2 parameters
	scope := query.Get("scope")
	state := query.Get("state")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")

	// Build authorization request
	authReq := &ports.AuthorizationRequest{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		OriginalURL:         r.URL.String(),
	}

	// Handle authorization
	decision, err := h.Service.HandleAuthorization(r.Context(), authReq, principalValue)
	if err != nil {
		http.Error(w, fmt.Sprintf("authorization error: %v", err), http.StatusInternalServerError)
		return
	}

	// Process decision
	switch decision.Action {
	case "redirect_to_upstream":
		http.Redirect(w, r, decision.RedirectURL, http.StatusFound)

	case "redirect_to_consent":
		http.Redirect(w, r, decision.RedirectURL, http.StatusFound)

	case "error":
		// Redirect with error if we have a redirect_uri, otherwise return error response
		if decision.RedirectURL != "" {
			http.Redirect(w, r, decision.RedirectURL, http.StatusFound)
		} else {
			// Return error response as JSON or plain text
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, `{"error":"%s","error_description":"%s"}`, decision.ErrorCode, decision.ErrorDesc)
		}

	default:
		http.Error(w, fmt.Sprintf("unknown authorization action: %s", decision.Action), http.StatusInternalServerError)
	}
}

// NewOAuth2AuthorizeHandler creates a new authorization handler
func NewOAuth2AuthorizeHandler(service ports.OAuth2Service) *OAuth2AuthorizeHandler {
	return &OAuth2AuthorizeHandler{
		Service: service,
	}
}

// ============================================================================
// T040: hasRequiredScopes helper function
// ============================================================================

// hasRequiredScopes checks if sessionScopes is a superset of requiredScopes (case-sensitive).
// Returns true if sessionScopes contains all scopes in requiredScopes, allowing for extra scopes.
// Returns true if requiredScopes is empty or nil (no requirements).
// Scope comparison is case-sensitive per SR-006.
func hasRequiredScopes(sessionScopes []string, requiredScopes []string) bool {
	// If no required scopes, always pass
	if len(requiredScopes) == 0 {
		return true
	}

	// If required scopes exist but session has none, fail
	if len(sessionScopes) == 0 {
		return false
	}

	// Build map of session scopes for efficient lookup
	sessionScopeMap := make(map[string]bool)
	for _, scope := range sessionScopes {
		sessionScopeMap[scope] = true
	}

	// Check that all required scopes exist in session scopes
	for _, required := range requiredScopes {
		if !sessionScopeMap[required] {
			return false
		}
	}

	return true
}

// ============================================================================
// T041: validateMandatoryRequirements method
// ============================================================================

// MandatoryRequirementError represents a validation error for mandatory requirements
type MandatoryRequirementError struct {
	Code      string
	Message   string
	ServiceID string
	UserID    string
	AgentID   string
}

func (e *MandatoryRequirementError) Error() string {
	if e.ServiceID != "" {
		return fmt.Sprintf("%s: %s (service: %s)", e.Code, e.Message, e.ServiceID)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// validateMandatoryRequirements checks that user has active sessions for all mandatory services
// with required scopes. Returns error if any mandatory requirement is not satisfied.
// Optional requirements are ignored and never block authorization.
func (h *OAuth2AuthorizeHandler) validateMandatoryRequirements(
	ctx context.Context,
	principal string,
	agent *storage.Agent,
) error {
	// If agent has no service requirements, nothing to validate
	if len(agent.ServiceRequirements) == 0 {
		return nil
	}

	// Check each mandatory requirement
	for _, req := range agent.ServiceRequirements {
		// Skip optional requirements
		if !req.IsMandatory() {
			continue
		}

		// Verify user has active session for this service
		session, err := h.sessionRepository.FindByPrincipalAndService(ctx, principal, req.ServiceID)
		if err != nil {
			// Session not found
			if h.logger != nil {
				h.logger.Warn(
					"MandatoryRequirementNotMet",
					"user_id", principal,
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "session_not_found",
				)
			}
			return &MandatoryRequirementError{
				Code:      "session_required",
				Message:   "User does not have an active session for required service",
				ServiceID: req.ServiceID,
				UserID:    principal,
				AgentID:   agent.ID,
			}
		}

		// Check if session is expired
		if session.IsExpired() {
			if h.logger != nil {
				h.logger.Warn(
					"MandatoryRequirementNotMet",
					"user_id", principal,
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "session_expired",
				)
			}
			return &MandatoryRequirementError{
				Code:      "session_expired",
				Message:   "User's session for required service has expired",
				ServiceID: req.ServiceID,
				UserID:    principal,
				AgentID:   agent.ID,
			}
		}

		// Check if session scopes include all required scopes (superset check)
		if !hasRequiredScopes(session.Scope, req.RequiredScopes) {
			if h.logger != nil {
				h.logger.Warn(
					"MandatoryRequirementNotMet",
					"user_id", principal,
					"agent_id", agent.ID,
					"service_id", req.ServiceID,
					"reason", "scope_mismatch",
					"required_scopes", req.RequiredScopes,
					"actual_scopes", session.Scope,
				)
			}
			return &MandatoryRequirementError{
				Code:      "scope_mismatch",
				Message:   "User's session lacks required scopes for service",
				ServiceID: req.ServiceID,
				UserID:    principal,
				AgentID:   agent.ID,
			}
		}
	}

	// All mandatory requirements satisfied
	return nil
}
