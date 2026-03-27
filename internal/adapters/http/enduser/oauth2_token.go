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
)

// OAuth2TokenHandler handles OAuth2 token endpoint requests
// Routes between token exchange (RFC 8693) and standard OAuth2 token requests
type OAuth2TokenHandler struct {
	UpstreamTokenURL string
	Client           *http.Client
	TokenExchange    *tokenexchange.TokenExchangeService // RFC 8693 token exchange service
	Logger           *slog.Logger                        // For structured logging
}

// ServeHTTP implements http.Handler for the token endpoint
func (h *OAuth2TokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Verify request method
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate Content-Type (allow charset parameter)
	// OAuth 2.0 token endpoint must accept application/x-www-form-urlencoded per RFC 6749 Section 4.1.3
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		http.Error(w, "invalid Content-Type: expected application/x-www-form-urlencoded", http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
		return
	}
	defer func() { _ = r.Body.Close() }()

	// Validate body is not empty
	if len(body) == 0 {
		http.Error(w, "request body cannot be empty", http.StatusBadRequest)
		return
	}

	// Parse form data to detect request type
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

	// Detect token exchange request (RFC 8693) by grant_type parameter
	if grantType == tokenexchange.TokenExchangeGrantType {
		if h.Logger != nil {
			h.Logger.Info("Routing to token exchange handler")
		}
		h.handleTokenExchange(w, r, formData)
		return
	}

	// For other grant types, proxy to upstream (standard OAuth2 flow)
	h.proxyToUpstream(w, r, string(body))
}

// handleTokenExchange processes RFC 8693 token exchange requests
// Implements complete token exchange flow: validation, authorization, token retrieval
func (h *OAuth2TokenHandler) handleTokenExchange(w http.ResponseWriter, r *http.Request, formData url.Values) {
	// Validate service is available (should have been validated in builder, but defensive check)
	if h.TokenExchange == nil {
		if h.Logger != nil {
			h.Logger.Error("Token exchange service not configured")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":             "server_error",
			"error_description": "token exchange service not configured",
		})
		return
	}

	// Parse token exchange request parameters from form data
	// Per RFC 8693: all parameters are form-encoded in request body
	req := tokenexchange.NewTokenExchangeRequest(
		formData.Get("grant_type"),
		formData.Get("subject_token"),
		formData.Get("subject_token_type"),
		formData.Get("client_assertion"),
		formData.Get("client_assertion_type"),
		formData.Get("resource"),
		formData.Get("scope"),
	)

	// Step 1: Validate required parameter: resource
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

	// Step 2: Call token exchange service
	// Service handles: request validation, JWT validation, authorization, token retrieval
	// Per Constitution Principle I (Security-First): service validates all inputs and fails closed
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
		if h.Logger != nil {
			logAttrs := []any{
				"error", err.Error(),
				"error_type", fmt.Sprintf("%T", err),
			}
			// Include underlying cause and structured details for operator diagnostics.
			// The cause is internal only and is never forwarded to the client (per SR-005).
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
		// Handle token exchange errors with proper RFC 8693 error codes
		h.handleTokenExchangeError(w, err)
		return
	}

	// Step 3: Return successful RFC 8693 response
	// Response includes: access_token, token_type, issued_token_type, expires_in
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token":      response.AccessToken,
		"token_type":        response.TokenType,
		"issued_token_type": response.IssuedTokenType,
		"expires_in":        response.ExpiresIn,
	}); err != nil {
		// Response headers already sent, can only log the encoding error
		if h.Logger != nil {
			h.Logger.Error("failed to encode token exchange response", "error", err)
		}
		return
	}

	// Audit log successful token exchange (optional, for operational observability)
	// Per SR-058: caller is responsible for audit logging (done here for successful flows)
	if h.Logger != nil {
		h.Logger.InfoContext(r.Context(), "token_exchange_succeeded",
			"resource", req.Resource,
			"issued_token_type", response.IssuedTokenType,
		)
	}
}

// handleTokenExchangeError maps domain-layer token exchange errors to RFC 8693 error responses
// Per spec FR-043: all errors return application/json error responses with standardized codes
// Per Constitution Principle I (Security-First): never expose sensitive token values in errors
func (h *OAuth2TokenHandler) handleTokenExchangeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Check if error is a domain TokenExchangeError with RFC 8693 error codes
	if tokenExchangeErr, ok := err.(*tokenexchange.TokenExchangeError); ok {
		// RFC 8693 error: use error code and description from domain
		w.WriteHeader(tokenExchangeErr.HTTPStatus())
		errBody := map[string]string{
			"error":             tokenExchangeErr.Code(),
			"error_description": tokenExchangeErr.Description(),
		}
		if tokenExchangeErr.ErrorURI() != "" {
			errBody["error_uri"] = tokenExchangeErr.ErrorURI()
		}
		if err := json.NewEncoder(w).Encode(errBody); err != nil {
			// Response headers already sent, can only log the encoding error
			if h.Logger != nil {
				h.Logger.Error("failed to encode token exchange error response", "error", err)
			}
		}
		return
	}

	// Unexpected error: return generic server_error
	// Per Constitution Principle I (Security-First): don't expose implementation details
	w.WriteHeader(http.StatusInternalServerError)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":             "server_error",
		"error_description": "internal server error during token exchange",
	}); err != nil {
		// Response headers already sent, can only log the encoding error
		if h.Logger != nil {
			h.Logger.Error("failed to encode generic error response", "error", err)
		}
	}
}

// proxyToUpstream forwards requests to upstream OAuth2 server (for non-token-exchange flows)
func (h *OAuth2TokenHandler) proxyToUpstream(w http.ResponseWriter, r *http.Request, body string) {
	ctx, span := otel.Tracer("upstream").Start(r.Context(), "oauth2.token_proxy")
	defer span.End()
	span.SetAttributes(attribute.String("http.method", "POST"))

	// Create upstream request with traced context
	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", h.UpstreamTokenURL, strings.NewReader(body))
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create upstream request: %v", err), http.StatusInternalServerError)
		return
	}

	// Copy headers from client request to upstream request, filtering hop-by-hop headers
	for key, values := range r.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			upstreamReq.Header.Add(key, value)
		}
	}

	// Make upstream request
	client := h.Client
	if client == nil {
		client = http.DefaultClient
	}

	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to contact upstream server: %v", err), http.StatusBadGateway)
		return
	}
	defer func() { _ = upstreamResp.Body.Close() }()

	// Set http.status_code attribute now that response is available
	span.SetAttributes(attribute.Int("http.status_code", upstreamResp.StatusCode))

	// Copy response headers from upstream to client, filtering hop-by-hop headers
	for key, values := range upstreamResp.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy response status code
	w.WriteHeader(upstreamResp.StatusCode)

	// Stream response body from upstream
	if _, err := io.Copy(w, upstreamResp.Body); err != nil {
		_, _ = fmt.Fprintf(w, "error streaming response: %v", err)
	}
}

// isHopByHopHeader returns true if the header is a hop-by-hop header per RFC 7230
func isHopByHopHeader(headerName string) bool {
	// Normalize to lowercase for comparison
	header := strings.ToLower(headerName)

	hopByHopHeaders := map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailers":            true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}

	return hopByHopHeaders[header]
}

// NewOAuth2TokenHandler creates a new token handler
func NewOAuth2TokenHandler(upstreamTokenURL string, client *http.Client) *OAuth2TokenHandler {
	return &OAuth2TokenHandler{
		UpstreamTokenURL: upstreamTokenURL,
		Client:           client,
	}
}
