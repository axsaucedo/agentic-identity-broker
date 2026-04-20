package enduser

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ory/fosite"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// AuthorizationProceedStrategy handles the "proceed" action from an AuthorizationDecision.
// This is the only part that differs between proxy mode and issue_token mode; all other
// decision actions (redirect_to_consent, error) are handled by the shared ServeHTTP flow.
type AuthorizationProceedStrategy interface {
	HandleProceed(w http.ResponseWriter, r *http.Request, decision *ports.AuthorizationDecision, req *ports.AuthorizationRequest, principal id.Principal)
}

// proxyProceedStrategy redirects to the upstream OAuth2 server URL from the decision.
type proxyProceedStrategy struct{}

// NewProxyProceedStrategy returns a ProceedStrategy that redirects to the upstream OAuth2 server.
func NewProxyProceedStrategy() AuthorizationProceedStrategy {
	return &proxyProceedStrategy{}
}

func (s *proxyProceedStrategy) HandleProceed(w http.ResponseWriter, r *http.Request, decision *ports.AuthorizationDecision, _ *ports.AuthorizationRequest, _ id.Principal) {
	http.Redirect(w, r, decision.RedirectURL, http.StatusFound)
}

// issueTokenProceedStrategy issues a local authorization code and redirects back to the client.
type issueTokenProceedStrategy struct {
	issuer ports.AuthorizationCodeIssuer
	logger *slog.Logger
}

// NewIssueTokenProceedStrategy returns a ProceedStrategy that issues authorization codes locally.
func NewIssueTokenProceedStrategy(issuer ports.AuthorizationCodeIssuer, logger *slog.Logger) AuthorizationProceedStrategy {
	return &issueTokenProceedStrategy{issuer: issuer, logger: logger}
}

func (s *issueTokenProceedStrategy) HandleProceed(w http.ResponseWriter, r *http.Request, _ *ports.AuthorizationDecision, req *ports.AuthorizationRequest, principal id.Principal) {
	code, err := s.issuer.IssueAuthorizationCode(r.Context(), req, principal)
	if err != nil {
		// Per RFC 6749 §4.1.2.1: never redirect when the client or redirect_uri is invalid/unverified.
		// ErrInvalidClient covers unknown client; ErrInvalidRedirectURI covers redirect URI failures.
		if errors.Is(err, fosite.ErrInvalidClient) || errors.Is(err, oauth2server.ErrInvalidRedirectURI) {
			var fositeErr *fosite.RFC6749Error
			if errors.As(err, &fositeErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(fositeErr.CodeField)
				_, _ = fmt.Fprintf(w, `{"error":%q,"error_description":%q}`, fositeErr.ErrorField, fositeErr.DescriptionField)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, `{"error":"invalid_request","error_description":"request validation failed"}`)
			return
		}

		// All other OAuth2 errors: redirect with error params (redirect_uri has been validated).
		var fositeErr *fosite.RFC6749Error
		if errors.As(err, &fositeErr) {
			redirectWithError(w, r, req.RedirectURI, req.State, fositeErr.ErrorField, fositeErr.DescriptionField)
			return
		}

		if s.logger != nil {
			s.logger.Error("unexpected error issuing authorization code", "error", err)
		}
		redirectWithError(w, r, req.RedirectURI, req.State, "server_error", "authorization failed")
		return
	}

	redirectURL, err := url.Parse(req.RedirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}

	q := redirectURL.Query()
	q.Set("code", code)
	if req.State != "" {
		q.Set("state", req.State)
	}
	redirectURL.RawQuery = q.Encode()

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}
