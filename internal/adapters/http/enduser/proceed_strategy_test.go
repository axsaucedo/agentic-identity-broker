package enduser

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

func newProceedRequest(t *testing.T) *http.Request {
	t.Helper()
	req := httptest.NewRequest("GET", "https://broker.example.com/oauth2/authorize", nil)
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	return req.WithContext(ctx)
}

func TestProxyProceedStrategy_RedirectsToDecisionURL(t *testing.T) {
	strategy := NewProxyProceedStrategy()
	w := httptest.NewRecorder()
	r := newProceedRequest(t)

	decision := &ports.AuthorizationDecision{
		Action:      "proceed",
		RedirectURL: "https://auth.example.com/authorize?client_id=abc&response_type=code",
	}
	req := &ports.AuthorizationRequest{RedirectURI: "https://client.example.com/callback"}

	strategy.HandleProceed(w, r, decision, req, id.NewPrincipal("user@example.com"))

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, decision.RedirectURL, w.Header().Get("Location"))
}

func TestIssueTokenProceedStrategy_Success_RedirectsWithCode(t *testing.T) {
	strategy := NewIssueTokenProceedStrategy(&mockCodeIssuer{}, nil)
	w := httptest.NewRecorder()
	r := newProceedRequest(t)

	req := &ports.AuthorizationRequest{
		RedirectURI: "https://client.example.com/callback",
		State:       "xyz123",
	}

	strategy.HandleProceed(w, r, &ports.AuthorizationDecision{}, req, id.NewPrincipal("user@example.com"))

	assert.Equal(t, http.StatusFound, w.Code)
	loc := w.Header().Get("Location")
	parsed, err := url.Parse(loc)
	require.NoError(t, err)
	assert.Equal(t, "test-code", parsed.Query().Get("code"))
	assert.Equal(t, "xyz123", parsed.Query().Get("state"))
}

func TestIssueTokenProceedStrategy_Success_OmitsStateWhenEmpty(t *testing.T) {
	strategy := NewIssueTokenProceedStrategy(&mockCodeIssuer{}, nil)
	w := httptest.NewRecorder()
	r := newProceedRequest(t)

	req := &ports.AuthorizationRequest{
		RedirectURI: "https://client.example.com/callback",
		State:       "",
	}

	strategy.HandleProceed(w, r, &ports.AuthorizationDecision{}, req, id.NewPrincipal("user@example.com"))

	assert.Equal(t, http.StatusFound, w.Code)
	loc := w.Header().Get("Location")
	parsed, err := url.Parse(loc)
	require.NoError(t, err)
	assert.Equal(t, "test-code", parsed.Query().Get("code"))
	assert.Empty(t, parsed.Query().Get("state"))
}

func TestIssueTokenProceedStrategy_Errors(t *testing.T) {
	agentID := id.NewAgentID()
	cases := []struct {
		name        string
		err         error
		wantStatus  int
		wantErrCode string
		isRedirect  bool
	}{
		// fosite.ErrInvalidClient: HTTP 401 per fosite — never redirect (RFC 6749 §4.1.2.1)
		{"unknown client", fosite.ErrInvalidClient, http.StatusUnauthorized, "invalid_client", false},
		// ErrInvalidRedirectURI wraps fosite.ErrInvalidRequest: HTTP 400 — never redirect
		{"invalid redirect uri", oauth2server.ErrInvalidRedirectURI, http.StatusBadRequest, "invalid_request", false},
		// fosite.ErrServerError: HTTP 500 — redirect is safe (redirect_uri was validated)
		{"server error", fosite.ErrServerError, http.StatusFound, "server_error", true},
		// fosite.ErrInvalidRequest: HTTP 400 — redirect is safe
		{"invalid request", fosite.ErrInvalidRequest, http.StatusFound, "invalid_request", true},
		// fosite.ErrUnsupportedResponseType: HTTP 400 — redirect is safe
		{"unsupported response type", fosite.ErrUnsupportedResponseType, http.StatusFound, "unsupported_response_type", true},
		// fosite.ErrInvalidScope: HTTP 400 — redirect is safe
		{"invalid scope", fosite.ErrInvalidScope, http.StatusFound, "invalid_scope", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			strategy := NewIssueTokenProceedStrategy(&errCodeIssuer{err: tc.err}, nil)
			w := httptest.NewRecorder()
			r := newProceedRequest(t)

			req := &ports.AuthorizationRequest{
				ClientID:    id.ClientID(agentID.String()),
				RedirectURI: "https://client.example.com/callback",
				State:       "xyz",
			}

			strategy.HandleProceed(w, r, &ports.AuthorizationDecision{}, req, id.NewPrincipal("user@example.com"))

			assert.Equal(t, tc.wantStatus, w.Code)
			if tc.isRedirect {
				loc := w.Header().Get("Location")
				parsed, err := url.Parse(loc)
				require.NoError(t, err)
				assert.Equal(t, tc.wantErrCode, parsed.Query().Get("error"))
				assert.NotEmpty(t, parsed.Query().Get("error_description"))
				assert.Equal(t, "xyz", parsed.Query().Get("state"))
			} else {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
				assert.Contains(t, w.Body.String(), tc.wantErrCode)
			}
		})
	}
}
