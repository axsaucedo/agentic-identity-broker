package enduser

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TokenGrantStrategy handles OAuth2 token grant requests at the HTTP transport layer.
// Parallel to AuthorizationProceedStrategy: proxy mode and local mode differ only
// in how they handle non-token-exchange grants.
type TokenGrantStrategy interface {
	HandleTokenGrant(w http.ResponseWriter, r *http.Request, grantType string, formData url.Values)
}

// proxyTokenGrantStrategy forwards token grant requests to an upstream OAuth2 server.
type proxyTokenGrantStrategy struct {
	upstreamTokenURL   string
	client             *http.Client
	agentRepository    ports.AgentRepository
	multiAgentVerifier ports.MultiAgentVerifier
	logger             *slog.Logger
}

// NewProxyTokenGrantStrategy returns a strategy that proxies token grants to an upstream server.
func NewProxyTokenGrantStrategy(
	upstreamTokenURL string,
	client *http.Client,
	agentRepository ports.AgentRepository,
	multiAgentVerifier ports.MultiAgentVerifier,
	logger *slog.Logger,
) *proxyTokenGrantStrategy {
	return &proxyTokenGrantStrategy{
		upstreamTokenURL:   upstreamTokenURL,
		client:             client,
		agentRepository:    agentRepository,
		multiAgentVerifier: multiAgentVerifier,
		logger:             logger,
	}
}

// HandleTokenGrant proxies the token request to the upstream OAuth2 server.
// The client_id is always the broker-internal agent UUID and is validated before
// forwarding (fail-closed per SR-001). When MultiAgentVerifier is set the upstream
// response body is buffered and the agent ID claim is verified before forwarding.
func (s *proxyTokenGrantStrategy) HandleTokenGrant(w http.ResponseWriter, r *http.Request, _ string, formData url.Values) {
	ctx, span := otel.Tracer("upstream").Start(r.Context(), "oauth2.token_proxy")
	defer span.End()
	span.SetAttributes(attribute.String("http.method", "POST"))

	rawClientID := formData.Get("client_id")
	if rawClientID == "" {
		if s.logger != nil {
			s.logger.Error("MissingClientIDInTokenRequest")
		}
		writeOAuth2ErrorJSON(w, http.StatusUnauthorized, "invalid_client", "client_id is required")
		return
	}

	agentID, parseErr := id.ParseAgentID(rawClientID)
	if parseErr != nil {
		if s.logger != nil {
			s.logger.Error("AgentIDParseError",
				"received_client_id", rawClientID,
				"error", parseErr,
			)
		}
		writeOAuth2ErrorJSON(w, http.StatusUnauthorized, "invalid_client", "client_id is not a valid agent UUID")
		return
	}

	// Defensive: builder.go always wires agentRepository, but direct construction in tests may omit it.
	if s.agentRepository == nil {
		if s.logger != nil {
			s.logger.Error("AgentRepositoryNotConfigured")
		}
		writeOAuth2ErrorJSON(w, http.StatusInternalServerError, "server_error", "agent repository not configured")
		return
	}

	agent, agentErr := s.agentRepository.Get(ctx, agentID)
	if agentErr != nil {
		if s.logger != nil {
			s.logger.Error("AgentLookupFailed",
				"agent_id", agentID.String(),
				"error", agentErr,
			)
		}
		writeOAuth2ErrorJSON(w, http.StatusUnauthorized, "invalid_client", "agent not found")
		return
	}

	// Replace the broker-internal UUID with the upstream client_id before forwarding.
	if agent.ClientID == nil {
		writeOAuth2ErrorJSON(w, http.StatusBadRequest, "invalid_client", "agent has no upstream client_id configured")
		return
	}
	formData.Set("client_id", agent.ClientID.String())
	body := formData.Encode()

	upstreamReq, err := http.NewRequestWithContext(ctx, "POST", s.upstreamTokenURL, strings.NewReader(body))
	if err != nil {
		http.Error(w, "failed to create upstream request", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			upstreamReq.Header.Add(key, value)
		}
	}

	client := s.client
	if client == nil {
		client = http.DefaultClient
	}

	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "failed to contact upstream server", http.StatusBadGateway)
		return
	}
	defer func() { _ = upstreamResp.Body.Close() }()

	span.SetAttributes(attribute.Int("http.status_code", upstreamResp.StatusCode))

	for key, values := range upstreamResp.Header {
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	if s.multiAgentVerifier != nil && upstreamResp.StatusCode == http.StatusOK {
		responseBody, readErr := io.ReadAll(upstreamResp.Body)
		if readErr != nil {
			if s.logger != nil {
				s.logger.Error("AgentIDClaimMissing",
					"agent_id", agentID.String(),
					"reason", "failed to read upstream response body",
					"error", readErr,
				)
			}
			w.Header().Del("Content-Length")
			w.Header().Del("Transfer-Encoding")
			writeOAuth2ErrorJSON(w, http.StatusInternalServerError, "server_error", "failed to read upstream response")
			return
		}

		if verifyErr := s.multiAgentVerifier.VerifyAgentIDClaim(r.Context(), responseBody, agentID); verifyErr != nil {
			if s.logger != nil {
				if mismatch, ok := verifyErr.(*oauth2.AgentIDMismatchError); ok {
					s.logger.Error("AgentIDClaimMismatch",
						"expected_agent_id", mismatch.Expected,
						"received_agent_id", mismatch.Received,
						"claim_name", mismatch.ClaimName,
					)
				} else {
					s.logger.Error("AgentIDClaimMissing",
						"agent_id", agentID.String(),
						"error", verifyErr.Error(),
					)
				}
			}
			w.Header().Del("Content-Length")
			w.Header().Del("Transfer-Encoding")
			writeOAuth2ErrorJSON(w, http.StatusInternalServerError, "server_error", "agent ID claim verification failed")
			return
		}

		if s.logger != nil {
			s.logger.Info("AgentIDClaimVerified",
				"agent_id", agentID.String(),
			)
		}

		w.WriteHeader(upstreamResp.StatusCode)
		_, _ = w.Write(responseBody)
		return
	}

	w.WriteHeader(upstreamResp.StatusCode)
	if _, err := io.Copy(w, upstreamResp.Body); err != nil {
		if s.logger != nil {
			s.logger.ErrorContext(r.Context(), "failed to stream upstream token response", "error", err)
		}
	}
}

// localGrantStrategy handles token grants locally using a TokenMintingStrategy.
type localGrantStrategy struct {
	minting         ports.TokenMintingStrategy
	agentRepository ports.AgentRepository
	logger          *slog.Logger
}

// NewLocalGrantStrategy returns a strategy that mints tokens locally.
// agentRepository is used to reject ProxyClient agents (those with an upstream client_id)
// from receiving locally minted tokens — ProxyClients must use the proxy path instead.
func NewLocalGrantStrategy(minting ports.TokenMintingStrategy, agentRepository ports.AgentRepository, logger *slog.Logger) *localGrantStrategy {
	return &localGrantStrategy{minting: minting, agentRepository: agentRepository, logger: logger}
}

// rejectIfProxyClient returns true (and writes an error) when the UUID client_id resolves
// to a ProxyClient agent. Non-UUID client_ids skip the check (CIMD URLs, opaque IDs).
func (s *localGrantStrategy) rejectIfProxyClient(w http.ResponseWriter, r *http.Request, rawClientID string) bool {
	if s.agentRepository == nil {
		return false
	}
	agentID, err := id.ParseAgentID(rawClientID)
	if err != nil {
		return false
	}
	agent, err := s.agentRepository.Get(r.Context(), agentID)
	if err != nil {
		return false
	}
	if agent.ClientMode() == storage.ProxyClient {
		writeOAuth2ErrorJSON(w, http.StatusUnauthorized, "invalid_client", "client is not eligible for local token issuance")
		return true
	}
	return false
}

// HandleTokenGrant dispatches client_credentials and authorization_code grants to the local minting strategy.
func (s *localGrantStrategy) HandleTokenGrant(w http.ResponseWriter, r *http.Request, grantType string, formData url.Values) {
	switch grantType {
	case "client_credentials":
		rawClientID := formData.Get("client_id")
		clientSecret := formData.Get("client_secret")
		scope := formData.Get("scope")

		if rawClientID == "" || clientSecret == "" {
			writeOAuth2ErrorJSON(w, http.StatusBadRequest, "invalid_request", "client_id and client_secret are required")
			return
		}

		if s.rejectIfProxyClient(w, r, rawClientID) {
			return
		}

		resp, err := s.minting.HandleClientCredentials(r.Context(), id.ClientID(rawClientID), clientSecret, scope)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("client_credentials grant failed", "error", err, "client_id", rawClientID)
			}
			s.handleMintingError(w, err, "client_credentials", rawClientID)
			return
		}

		if s.logger != nil {
			s.logger.Info("TokenIssued",
				"event", "TokenIssued",
				"grant_type", "client_credentials",
				"client_id", rawClientID,
				"scope", scope,
			)
		}

		s.writeTokenResponse(w, resp)

	case "authorization_code":
		rawClientID := formData.Get("client_id")
		clientSecret := formData.Get("client_secret")
		code := formData.Get("code")
		redirectURI := formData.Get("redirect_uri")
		codeVerifier := formData.Get("code_verifier")

		if rawClientID == "" {
			writeOAuth2ErrorJSON(w, http.StatusBadRequest, "invalid_request", "client_id is required")
			return
		}
		if code == "" {
			writeOAuth2ErrorJSON(w, http.StatusBadRequest, "invalid_request", "code is required")
			return
		}

		if s.rejectIfProxyClient(w, r, rawClientID) {
			return
		}

		resp, err := s.minting.HandleAuthorizationCodeExchange(r.Context(), id.ClientID(rawClientID), clientSecret, code, redirectURI, codeVerifier)
		if err != nil {
			if s.logger != nil {
				s.logger.Error("authorization_code exchange failed", "error", err, "client_id", rawClientID)
			}
			s.handleMintingError(w, err, "authorization_code", rawClientID)
			return
		}

		if s.logger != nil {
			s.logger.Info("TokenIssued",
				"event", "TokenIssued",
				"grant_type", "authorization_code",
				"client_id", rawClientID,
			)
		}

		s.writeTokenResponse(w, resp)

	default:
		writeOAuth2ErrorJSON(w, http.StatusBadRequest, "unsupported_grant_type",
			"grant_type must be 'client_credentials' or 'authorization_code'")
	}
}

func (s *localGrantStrategy) handleMintingError(w http.ResponseWriter, err error, grantType, clientID string) {
	var errorCode, errorDesc string
	var statusCode int

	var rfc6749Err *oauth2server.RFC6749Error
	if errors.As(err, &rfc6749Err) {
		errorCode = rfc6749Err.Code()
		errorDesc = rfc6749Err.Description()
		statusCode = rfc6749Err.HTTPStatus()
	} else {
		errorCode = "server_error"
		errorDesc = "internal server error"
		statusCode = http.StatusInternalServerError
	}

	writeOAuth2ErrorJSON(w, statusCode, errorCode, errorDesc)

	if s.logger != nil {
		s.logger.Warn("TokenRequestFailed",
			"event", "TokenRequestFailed",
			"grant_type", grantType,
			"error_code", errorCode,
			"error_description", errorDesc,
			"client_id", clientID,
		)
	}
}

func (s *localGrantStrategy) writeTokenResponse(w http.ResponseWriter, resp *ports.TokenResponse) {
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
		if s.logger != nil {
			s.logger.Error("failed to encode token response", "error", err)
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

// hybridTokenGrantStrategy dispatches token grants to proxy or local based on the agent's ClientMode.
// ProxyClient agents (with upstream ClientID) are forwarded; CIMDClient and LocalClient are minted locally.
type hybridTokenGrantStrategy struct {
	proxy           TokenGrantStrategy
	local           TokenGrantStrategy
	agentRepository ports.AgentRepository
	logger          *slog.Logger
}

// NewHybridTokenGrantStrategy returns a TokenGrantStrategy that dispatches based on client mode.
func NewHybridTokenGrantStrategy(
	proxy, local TokenGrantStrategy,
	agentRepository ports.AgentRepository,
	logger *slog.Logger,
) TokenGrantStrategy {
	return &hybridTokenGrantStrategy{
		proxy:           proxy,
		local:           local,
		agentRepository: agentRepository,
		logger:          logger,
	}
}

func (s *hybridTokenGrantStrategy) HandleTokenGrant(w http.ResponseWriter, r *http.Request, grantType string, formData url.Values) {
	rawClientID := formData.Get("client_id")

	var agent *storage.Agent
	var lookupErr error

	agentID, parseErr := id.ParseAgentID(rawClientID)
	if parseErr == nil {
		agent, lookupErr = s.agentRepository.Get(r.Context(), agentID)
	} else {
		// Non-UUID client_id: resolve via ClientURI (CIMD URL) first, then opaque ClientID.
		agent, lookupErr = s.agentRepository.GetByClientURI(r.Context(), rawClientID)
		if lookupErr != nil && ports.IsNotFoundErr(lookupErr) {
			agent, lookupErr = s.agentRepository.GetByClientID(r.Context(), id.ClientID(rawClientID))
		}
	}

	if lookupErr != nil {
		if s.logger != nil {
			s.logger.Error("HybridTokenGrant: agent lookup failed",
				"client_id", rawClientID,
				"error", lookupErr,
			)
		}
		writeOAuth2ErrorJSON(w, http.StatusUnauthorized, "invalid_client", "agent not found")
		return
	}

	if agent.ClientMode() == storage.ProxyClient {
		s.proxy.HandleTokenGrant(w, r, grantType, formData)
	} else {
		s.local.HandleTokenGrant(w, r, grantType, formData)
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

func isHopByHopHeader(headerName string) bool {
	return hopByHopHeaders[strings.ToLower(headerName)]
}
