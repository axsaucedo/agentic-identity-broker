package enduser

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2AuthorizeHandler handles OAuth2 authorization endpoint requests.
// The ProceedHandler strategy determines the response for an active grant: proxy mode
// redirects to the upstream OAuth2 server, issue_token mode issues a local authorization code.
type OAuth2AuthorizeHandler struct {
	Service           ports.OAuth2Service
	sessionRepository ports.UserSessionRepository
	logger            *slog.Logger
	ProceedHandler    AuthorizationProceedStrategy
}

// ServeHTTP implements http.Handler for the authorization endpoint.
// The shared flow handles consent checks and error decisions; the ProceedHandler
// strategy determines the response when the user has an active grant.
func (h *OAuth2AuthorizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	principalValue := principal.MustFromContext(r.Context())

	query := r.URL.Query()

	rawClientID := query.Get("client_id")
	if rawClientID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `{"error":"invalid_request","error_description":"missing required parameter: client_id"}`)
		return
	}

	redirectURI := query.Get("redirect_uri")
	if redirectURI == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `{"error":"invalid_request","error_description":"missing required parameter: redirect_uri"}`)
		return
	}

	responseType := query.Get("response_type")
	if responseType == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `{"error":"invalid_request","error_description":"missing required parameter: response_type"}`)
		return
	}

	authReq := &ports.AuthorizationRequest{
		ClientID:            id.ClientID(rawClientID),
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               query.Get("scope"),
		State:               query.Get("state"),
		CodeChallenge:       query.Get("code_challenge"),
		CodeChallengeMethod: query.Get("code_challenge_method"),
		OriginalURL:         r.URL.String(),
	}

	decision, err := h.Service.HandleAuthorization(r.Context(), authReq, id.NewPrincipal(principalValue))
	if err != nil {
		http.Error(w, fmt.Sprintf("authorization error: %v", err), http.StatusInternalServerError)
		return
	}

	switch decision.Action {
	case "proceed":
		h.ProceedHandler.HandleProceed(w, r, decision, authReq, id.NewPrincipal(principalValue))

	case "redirect_to_consent":
		http.Redirect(w, r, decision.RedirectURL, http.StatusFound)

	case "error":
		respondWithDecisionError(w, r, decision)

	default:
		if h.logger != nil {
			h.logger.Error("unexpected authorization decision action", "action", decision.Action)
		}
		redirectWithError(w, r, authReq.RedirectURI, authReq.State, "server_error", "unexpected authorization decision")
	}
}

// respondWithDecisionError writes an OAuth2 error response for an "error" decision.
// Redirects with error params when a redirect_uri is available; otherwise returns JSON.
func respondWithDecisionError(w http.ResponseWriter, r *http.Request, decision *ports.AuthorizationDecision) {
	if decision.RedirectURL != "" {
		http.Redirect(w, r, decision.RedirectURL, http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorDecisionStatus(decision.ErrorCode))
		_, _ = fmt.Fprintf(w, `{"error":"%s","error_description":"%s"}`, decision.ErrorCode, decision.ErrorDesc)
	}
}

// NewOAuth2AuthorizeHandler creates a new authorization handler
func NewOAuth2AuthorizeHandler(service ports.OAuth2Service) *OAuth2AuthorizeHandler {
	return &OAuth2AuthorizeHandler{
		Service: service,
	}
}

// errorDecisionStatus maps an OAuth2 error code to an HTTP status for direct (non-redirect)
// error responses. Infrastructure errors map to 500; all others map to 400.
func errorDecisionStatus(errorCode string) int {
	if errorCode == "server_error" {
		return http.StatusInternalServerError
	}
	return http.StatusBadRequest
}

// redirectWithError performs an OAuth2 error redirect per RFC 6749 Section 4.1.2.1.
func redirectWithError(w http.ResponseWriter, r *http.Request, redirectURI, state, errorCode, errorDescription string) {
	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	q := redirectURL.Query()
	q.Set("error", errorCode)
	q.Set("error_description", errorDescription)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
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
		session, err := h.sessionRepository.FindByPrincipalAndService(ctx, id.Principal(principal), req.ServiceID)
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
				ServiceID: req.ServiceID.String(),
				UserID:    principal,
				AgentID:   agent.ID.String(),
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
				ServiceID: req.ServiceID.String(),
				UserID:    principal,
				AgentID:   agent.ID.String(),
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
				ServiceID: req.ServiceID.String(),
				UserID:    principal,
				AgentID:   agent.ID.String(),
			}
		}
	}

	// All mandatory requirements satisfied
	return nil
}
