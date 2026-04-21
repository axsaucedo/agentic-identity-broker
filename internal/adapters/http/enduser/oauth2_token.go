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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2TokenHandler handles OAuth2 token endpoint requests
// Routes between token exchange (RFC 8693) and standard OAuth2 token requests.
// When TokenMinting is set (issue_token mode), client_credentials and authorization_code
// grants are handled locally instead of being proxied to upstream.
type OAuth2TokenHandler struct {
	UpstreamTokenURL   string
	Client             *http.Client
	TokenExchange      *tokenexchange.TokenExchangeService // RFC 8693 token exchange service
	Logger             *slog.Logger                        // For structured logging
	MultiAgentVerifier ports.MultiAgentVerifier            // nil = feature disabled; non-nil = verify agent ID claim
	AgentRepository    ports.AgentRepository               // resolves broker agent UUID → upstream client_id
	TokenMinting       ports.TokenMintingStrategy          // nil = proxy mode; non-nil = local minting (issue_token mode)
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

	// When TokenMinting strategy is set (issue_token mode), handle
	// client_credentials and authorization_code grants locally.
	if h.TokenMinting != nil {
		h.handleLocalMinting(w, r, grantType, formData)
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

// proxyToUpstream forwards requests to upstream OAuth2 server (for non-token-exchange flows).
//
// The client_id in the token request is always the agent ID (broker-internal UUID).
// This is validated unconditionally: a missing or malformed client_id is rejected before
// forwarding to upstream (fail-closed per SR-001).
//
// When MultiAgentVerifier is non-nil, the upstream response body is additionally buffered
// so the agent ID claim in the returned token can be verified before forwarding the
// response. On verification failure the response is withheld and an OAuth2 server_error is
// returned to the client (fail-closed per SR-001). When MultiAgentVerifier is nil the body
// is streamed unchanged.
func (h *OAuth2TokenHandler) proxyToUpstream(w http.ResponseWriter, r *http.Request, body string) {
	ctx, span := otel.Tracer("upstream").Start(r.Context(), "oauth2.token_proxy")
	defer span.End()
	span.SetAttributes(attribute.String("http.method", "POST"))

	// The client_id is always the agent UUID from the broker's perspective.
	// Validate it unconditionally so that only agents with a recognised UUID can
	// reach the upstream OAuth2 server.
	formData, err := url.ParseQuery(body)
	if err != nil {
		// Fail closed: malformed form body means we cannot reliably determine agent ID.
		if h.Logger != nil {
			h.Logger.Error("TokenRequestFormParseError",
				slog.String("error", err.Error()),
			)
		}
		h.writeOAuth2Error(w, http.StatusBadRequest, "invalid_request", "token request body is not valid form-encoded data")
		return
	}

	rawClientID := formData.Get("client_id")
	if rawClientID == "" {
		// Fail closed when client_id is missing: we cannot determine the expected agent.
		if h.Logger != nil {
			h.Logger.Error("MissingClientIDInTokenRequest")
		}
		h.writeOAuth2Error(w, http.StatusUnauthorized, "invalid_client", "client_id is required")
		return
	}

	agentID, parseErr := id.ParseAgentID(rawClientID)
	if parseErr != nil {
		// Fail closed per SR-001: reject immediately rather than forwarding
		// a non-agent client_id to upstream.
		if h.Logger != nil {
			h.Logger.Error("AgentIDParseError",
				"received_client_id", rawClientID,
				"error", parseErr,
			)
		}
		h.writeOAuth2Error(w, http.StatusUnauthorized, "invalid_client", "client_id is not a valid agent UUID")
		return
	}

	// Resolve the upstream client_id from the broker-internal agent UUID.
	// The broker exposes agent UUIDs as client_ids, but the upstream OAuth2 server
	// uses the agent's configured ClientID (agent.client_id) for authentication.
	// Per review comment r2995280734: the upstream does not know about agent UUIDs.
	//
	// The nil guard is defensive: builder.go always wires AgentRepository, but direct
	// handler construction in tests or future code may omit it. Fail closed (per SR-001)
	// rather than forwarding the agent UUID to an upstream that cannot interpret it.
	if h.AgentRepository == nil {
		if h.Logger != nil {
			h.Logger.Error("AgentRepositoryNotConfigured")
		}
		h.writeOAuth2Error(w, http.StatusInternalServerError, "server_error", "agent repository not configured")
		return
	}
	agent, agentErr := h.AgentRepository.Get(ctx, agentID)
	if agentErr != nil {
		if h.Logger != nil {
			h.Logger.Error("AgentLookupFailed",
				"agent_id", agentID.String(),
				"error", agentErr,
			)
		}
		h.writeOAuth2Error(w, http.StatusUnauthorized, "invalid_client", "agent not found")
		return
	}
	// Replace the broker-internal UUID with the upstream client_id before forwarding.
	formData.Set("client_id", string(agent.ClientID))
	body = formData.Encode()

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

	// Feature 021: when verifier is set, buffer the response body and verify the agent ID claim
	// before forwarding. This prevents a compromised upstream from returning tokens for the
	// wrong agent (fail-closed per SR-001).
	if h.MultiAgentVerifier != nil && upstreamResp.StatusCode == http.StatusOK {
		responseBody, readErr := io.ReadAll(upstreamResp.Body)
		if readErr != nil {
			if h.Logger != nil {
				h.Logger.Error("AgentIDClaimMissing",
					"agent_id", agentID.String(),
					"reason", "failed to read upstream response body",
					"error", readErr,
				)
			}
			w.Header().Del("Content-Length")
			w.Header().Del("Transfer-Encoding")
			h.writeOAuth2Error(w, http.StatusInternalServerError, "server_error", "failed to read upstream response")
			return
		}

		if verifyErr := h.MultiAgentVerifier.VerifyAgentIDClaim(r.Context(), responseBody, agentID); verifyErr != nil {
			if h.Logger != nil {
				if mismatch, ok := verifyErr.(*oauth2.AgentIDMismatchError); ok {
					// T027: AgentIDClaimMismatch audit log (Error) — structured fields for SIEM
					h.Logger.Error("AgentIDClaimMismatch",
						"expected_agent_id", mismatch.Expected,
						"received_agent_id", mismatch.Received,
						"claim_name", mismatch.ClaimName,
					)
				} else {
					// T027: AgentIDClaimMissing audit log (Error)
					h.Logger.Error("AgentIDClaimMissing",
						"agent_id", agentID.String(),
						"error", verifyErr.Error(),
					)
				}
			}
			// Clear upstream headers before writing error response.
			// Upstream Content-Length would mismatch the error JSON body size,
			// causing HTTP/1.1 connection hangs (blocking architect fix).
			w.Header().Del("Content-Length")
			w.Header().Del("Transfer-Encoding")
			h.writeOAuth2Error(w, http.StatusInternalServerError, "server_error", "agent ID claim verification failed")
			return
		}

		// T027: AgentIDClaimVerified audit log (Info)
		if h.Logger != nil {
			h.Logger.Info("AgentIDClaimVerified",
				"agent_id", agentID.String(),
			)
		}

		// Write buffered body
		w.WriteHeader(upstreamResp.StatusCode)
		_, _ = w.Write(responseBody)
		return
	}

	// No verification needed — copy response status code and stream body
	w.WriteHeader(upstreamResp.StatusCode)
	if _, err := io.Copy(w, upstreamResp.Body); err != nil {
		if h.Logger != nil {
			h.Logger.ErrorContext(r.Context(), "failed to stream upstream token response", "error", err)
		}
	}
}

// writeOAuth2Error writes an RFC 6749/8693-style JSON error response.
func (h *OAuth2TokenHandler) writeOAuth2Error(w http.ResponseWriter, status int, code, description string) {
	body, err := json.Marshal(map[string]string{
		"error":             code,
		"error_description": description,
	})
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("failed to encode OAuth2 error response", "error", err)
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		if h.Logger != nil {
			h.Logger.Error("failed to write OAuth2 error response body", "error", err)
		}
	}
}

// hopByHopHeaders is the set of hop-by-hop headers per RFC 7230 that must not be forwarded.
var hopByHopHeaders = map[string]bool{
	"connection":          true,
	"keep-alive":          true,
	"proxy-authenticate":  true,
	"proxy-authorization": true,
	"te":                  true,
	"trailers":            true,
	"transfer-encoding":   true,
	"upgrade":             true,
}

// isHopByHopHeader returns true if the header is a hop-by-hop header per RFC 7230.
func isHopByHopHeader(headerName string) bool {
	return hopByHopHeaders[strings.ToLower(headerName)]
}

// NewOAuth2TokenHandler creates a new token handler
func NewOAuth2TokenHandler(upstreamTokenURL string, client *http.Client) *OAuth2TokenHandler {
	return &OAuth2TokenHandler{
		UpstreamTokenURL: upstreamTokenURL,
		Client:           client,
	}
}

// handleLocalMinting dispatches client_credentials and authorization_code grants
// to the local TokenMintingStrategy (issue_token mode).
func (h *OAuth2TokenHandler) handleLocalMinting(w http.ResponseWriter, r *http.Request, grantType string, formData url.Values) {
	switch grantType {
	case "client_credentials":
		rawClientID := formData.Get("client_id")
		clientSecret := formData.Get("client_secret")
		scope := formData.Get("scope")

		if rawClientID == "" || clientSecret == "" {
			h.writeOAuth2Error(w, http.StatusBadRequest, "invalid_request", "client_id and client_secret are required")
			return
		}

		agentID, err := id.ParseAgentID(rawClientID)
		if err != nil {
			h.writeOAuth2Error(w, http.StatusUnauthorized, "invalid_client", "client_id is not a valid agent UUID")
			return
		}

		resp, err := h.TokenMinting.HandleClientCredentials(r.Context(), agentID, clientSecret, scope)
		if err != nil {
			if h.Logger != nil {
				h.Logger.Error("client_credentials grant failed", "error", err, "client_id", rawClientID)
			}
			h.handleMintingError(w, err, "client_credentials", rawClientID)
			return
		}

		if h.Logger != nil {
			h.Logger.Info("TokenIssued",
				"event", "TokenIssued",
				"grant_type", "client_credentials",
				"client_id", rawClientID,
				"scope", scope,
			)
		}

		h.writeTokenResponse(w, resp)

	case "authorization_code":
		rawClientID := formData.Get("client_id")
		clientSecret := formData.Get("client_secret")
		code := formData.Get("code")
		redirectURI := formData.Get("redirect_uri")
		codeVerifier := formData.Get("code_verifier")

		if rawClientID == "" || clientSecret == "" {
			h.writeOAuth2Error(w, http.StatusBadRequest, "invalid_request", "client_id and client_secret are required")
			return
		}
		if code == "" {
			h.writeOAuth2Error(w, http.StatusBadRequest, "invalid_request", "code is required")
			return
		}

		agentID, err := id.ParseAgentID(rawClientID)
		if err != nil {
			h.writeOAuth2Error(w, http.StatusUnauthorized, "invalid_client", "client_id is not a valid agent UUID")
			return
		}

		resp, err := h.TokenMinting.HandleAuthorizationCodeExchange(r.Context(), agentID, clientSecret, code, redirectURI, codeVerifier)
		if err != nil {
			if h.Logger != nil {
				h.Logger.Error("authorization_code exchange failed", "error", err, "client_id", rawClientID)
			}
			h.handleMintingError(w, err, "authorization_code", rawClientID)
			return
		}

		if h.Logger != nil {
			h.Logger.Info("TokenIssued",
				"event", "TokenIssued",
				"grant_type", "authorization_code",
				"client_id", rawClientID,
			)
		}

		h.writeTokenResponse(w, resp)

	default:
		h.writeOAuth2Error(w, http.StatusBadRequest, "unsupported_grant_type",
			"grant_type must be 'client_credentials' or 'authorization_code'")
	}
}

// handleMintingError writes an OAuth2 error response derived from the fosite error type.
func (h *OAuth2TokenHandler) handleMintingError(w http.ResponseWriter, err error, grantType, clientID string) {
	var errorCode, errorDesc string
	var statusCode int

	var rfc6749Err *oauth2server.RFC6749Error
	if errors.As(err, &rfc6749Err) {
		errorCode = rfc6749Err.ErrorCode
		errorDesc = rfc6749Err.Description
		statusCode = rfc6749Err.HTTPStatus
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}
	} else {
		errorCode = "server_error"
		errorDesc = "internal server error"
		statusCode = http.StatusInternalServerError
	}

	h.writeOAuth2Error(w, statusCode, errorCode, errorDesc)

	if h.Logger != nil {
		h.Logger.Warn("TokenRequestFailed",
			"event", "TokenRequestFailed",
			"grant_type", grantType,
			"error_code", errorCode,
			"error_description", errorDesc,
			"client_id", clientID,
		)
	}
}

// writeTokenResponse writes a successful OAuth2 token response from the minting strategy.
func (h *OAuth2TokenHandler) writeTokenResponse(w http.ResponseWriter, resp *ports.TokenResponse) {
	tokenResp := map[string]interface{}{
		"access_token": resp.AccessToken,
		"token_type":   resp.TokenType,
		"expires_in":   resp.ExpiresIn,
	}
	if resp.Scope != "" {
		tokenResp["scope"] = resp.Scope
	}

	body, err := json.Marshal(tokenResp)
	if err != nil {
		if h.Logger != nil {
			h.Logger.Error("failed to encode token response", "error", err)
		}
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
