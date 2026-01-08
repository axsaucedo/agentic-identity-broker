package enduser

import (
	"fmt"
	"net/http"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2AuthorizeHandler handles OAuth2 authorization endpoint requests
type OAuth2AuthorizeHandler struct {
	Service ports.OAuth2Service
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
