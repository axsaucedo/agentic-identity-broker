package enduser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2TokenHandler handles OAuth2 token endpoint requests.
// Routes RFC 8693 token exchange to handleTokenExchange; all other grants are
// resolved via OAuth2Service.ResolveForTokenGrant then delegated to GrantHandler.
type OAuth2TokenHandler struct {
	TokenExchange *tokenexchange.TokenExchangeService
	OAuth2Service ports.OAuth2Service
	Logger        *slog.Logger
	GrantHandler  TokenGrantStrategy // always non-nil: proxy, local, or hybrid
}

// ServeHTTP implements http.Handler for the token endpoint.
func (h *OAuth2TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// OAuth 2.0 token endpoint must accept application/x-www-form-urlencoded per RFC 6749 Section 4.1.3
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		http.Error(w, "invalid Content-Type: expected application/x-www-form-urlencoded", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	if len(body) == 0 {
		http.Error(w, "request body cannot be empty", http.StatusBadRequest)
		return
	}

	formData, err := url.ParseQuery(string(body))
	if err != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}

	grantType := formData.Get("grant_type")
	if h.Logger != nil {
		h.Logger.Info("Token endpoint request received",
			"grant_type", grantType,
			"expected_grant_type", tokenexchange.TokenExchangeGrantType,
		)
	}

	if grantType == tokenexchange.TokenExchangeGrantType {
		if h.Logger != nil {
			h.Logger.Info("Routing to token exchange handler")
		}
		h.handleTokenExchange(w, r, formData)
		return
	}

	rawClientID := formData.Get("client_id")
	if rawClientID == "" {
		writeOAuth2ErrorJSON(w, http.StatusBadRequest, "invalid_request", "client_id is required")
		return
	}

	if h.OAuth2Service == nil || h.GrantHandler == nil {
		writeOAuth2ErrorJSON(w, http.StatusServiceUnavailable, "server_error", "OAuth2 authorization server not configured")
		return
	}

	resolution, resolveErr := h.OAuth2Service.ResolveForTokenGrant(r.Context(), rawClientID)
	if resolveErr != nil {
		var clientErr *ports.ClientIDError
		if errors.As(resolveErr, &clientErr) {
			status := tokenEndpointStatus(clientErr.Code)
			writeOAuth2ErrorJSON(w, status, clientErr.Code, clientErr.Desc)
		} else {
			writeOAuth2ErrorJSON(w, http.StatusInternalServerError, "server_error", "client resolution failed")
		}
		return
	}

	h.GrantHandler.HandleTokenGrant(w, r, grantType, formData, resolution.Agent)
}

// handleTokenExchange processes RFC 8693 token exchange requests.
func (h *OAuth2TokenHandler) handleTokenExchange(w http.ResponseWriter, r *http.Request, formData url.Values) {
	if h.TokenExchange == nil {
		if h.Logger != nil {
			h.Logger.Info("token exchange not wired, returning unsupported_grant_type")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "unsupported_grant_type",
			"error_description": "token exchange is not available in this deployment mode",
		})
		return
	}

	req := tokenexchange.NewTokenExchangeRequest(
		formData.Get("grant_type"),
		formData.Get("subject_token"),
		formData.Get("subject_token_type"),
		formData.Get("client_assertion"),
		formData.Get("client_assertion_type"),
		formData.Get("resource"),
		formData.Get("scope"),
	)

	// Per FR-008: resource parameter is mandatory and validation occurs at HTTP layer
	if req.Resource == "" {
		if h.Logger != nil {
			h.Logger.Warn("Resource parameter missing")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "invalid_request",
			"error_description": "resource parameter is required",
		})
		return
	}

	if h.Logger != nil {
		h.Logger.Info("Calling TokenExchangeService.Exchange",
			"resource", req.Resource,
			"subject_token_present", req.SubjectToken != "",
			"client_assertion_present", req.ClientAssertion != "",
		)
	}

	ctx, span := otel.Tracer("tokenexchange").Start(r.Context(), "tokenexchange.exchange")
	defer span.End()
	span.SetAttributes(
		attribute.String("token_exchange.resource", req.Resource),
		attribute.String("token_exchange.grant_type", req.GrantType),
	)
	response, err := h.TokenExchange.Exchange(ctx, req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if tokenErr, ok := err.(*tokenexchange.TokenExchangeError); ok {
			if details := tokenErr.Details(); details != "" {
				span.SetAttributes(attribute.String("token_exchange.validation_details", truncateSpanAttribute(details, 512)))
			}
			span.SetAttributes(
				attribute.String("token_exchange.error_code", tokenErr.Code()),
				attribute.String("token_exchange.error_description", tokenErr.Description()),
			)
		}
		if h.Logger != nil {
			logAttrs := []any{
				"error", err.Error(),
				"error_type", fmt.Sprintf("%T", err),
			}
			if tokenErr, ok := err.(*tokenexchange.TokenExchangeError); ok {
				if cause := errors.Unwrap(tokenErr); cause != nil {
					logAttrs = append(logAttrs, "cause", cause.Error())
				}
				if details := tokenErr.Details(); details != "" {
					logAttrs = append(logAttrs, "details", details)
				}
			}
			h.Logger.Error("Token exchange failed", logAttrs...)
		}
		h.handleTokenExchangeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":      response.AccessToken,
		"token_type":        response.TokenType,
		"issued_token_type": response.IssuedTokenType,
		"expires_in":        response.ExpiresIn,
	}); err != nil {
		if h.Logger != nil {
			h.Logger.Error("failed to encode token exchange response", "error", err)
		}
		return
	}

	if h.Logger != nil {
		h.Logger.InfoContext(r.Context(), "token_exchange_succeeded",
			"resource", req.Resource,
			"issued_token_type", response.IssuedTokenType,
		)
	}
}

// handleTokenExchangeError maps domain-layer token exchange errors to RFC 8693 error responses.
func (h *OAuth2TokenHandler) handleTokenExchangeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	if tokenExchangeErr, ok := err.(*tokenexchange.TokenExchangeError); ok {
		w.WriteHeader(tokenExchangeErr.HTTPStatus())
		errBody := map[string]string{
			"error":             tokenExchangeErr.Code(),
			"error_description": tokenExchangeErr.Description(),
		}
		if tokenExchangeErr.ErrorURI() != "" {
			errBody["error_uri"] = tokenExchangeErr.ErrorURI()
		}
		if err := json.NewEncoder(w).Encode(errBody); err != nil {
			if h.Logger != nil {
				h.Logger.Error("failed to encode token exchange error response", "error", err)
			}
		}
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":             "server_error",
		"error_description": "internal server error during token exchange",
	}); err != nil {
		if h.Logger != nil {
			h.Logger.Error("failed to encode generic error response", "error", err)
		}
	}
}

// tokenEndpointStatus maps an OAuth2 error code to the appropriate HTTP status for the token endpoint.
// RFC 6749 §5.2: invalid_client → 401, unauthorized_client → 400, server_error → 500.
func tokenEndpointStatus(code string) int {
	switch code {
	case "unauthorized_client":
		return http.StatusBadRequest
	case "server_error":
		return http.StatusInternalServerError
	default:
		return http.StatusUnauthorized
	}
}

// truncateSpanAttribute trims s to at most maxRunes runes and replaces newlines
// with spaces, producing a single-line string safe to emit as an OTel span attribute.
func truncateSpanAttribute(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) > maxRunes {
		runes = runes[:maxRunes]
	}
	result := make([]rune, len(runes))
	for i, r := range runes {
		if r == '\n' || r == '\r' {
			result[i] = ' '
		} else {
			result[i] = r
		}
	}
	return string(result)
}
