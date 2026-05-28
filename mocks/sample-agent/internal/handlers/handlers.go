package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/agentic-identity-broker/sample-agent/internal/config"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/oauth2"
)

// UserInfo represents user information retrieved from OAuth2
type UserInfo struct {
	Sub string `json:"sub"`
}

// Session represents a user session
type Session struct {
	Token        *oauth2.Token
	UserInfo     *UserInfo
	CreatedAt    time.Time
	ExpiresAt    int64
	CSRFState    string
	ClientType   string         // "proxy", "local", or "cimd"
	OAuth2Config *oauth2.Config // per-flow config (correct client_id/secret)
}

// Handlers handles HTTP requests
type Handlers struct {
	oauth2Config *oauth2.Config
	cfg          *config.Config
	sessions     map[string]*Session
	sessionsMu   sync.RWMutex
}

// New creates a new Handlers instance
func New(oauth2Config *oauth2.Config, cfg *config.Config) *Handlers {
	return &Handlers{
		oauth2Config: oauth2Config,
		cfg:          cfg,
		sessions:     make(map[string]*Session),
	}
}

// Health returns a health check response
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "sample-oauth2-client",
	})
}

// Home renders the home page with login or user info
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionID(r)

	h.sessionsMu.RLock()
	session, exists := h.sessions[sessionID]
	h.sessionsMu.RUnlock()

	if exists && session.Token.Valid() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		slog.Info("Displaying user info",
			"sub", session.UserInfo.Sub,
			"expiresAt", session.ExpiresAt)
		fmt.Fprint(w, renderUserPage(session.UserInfo, session.ExpiresAt, session.ClientType))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, renderSelectorPage(h.cfg))
}

// Login initiates the OAuth2 authorization flow for a chosen client type.
// Accepts ?type=proxy|local|cimd; defaults to proxy.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	clientType := r.URL.Query().Get("type")
	var agentID string
	switch clientType {
	case "local":
		agentID = h.cfg.OAuth2.LocalAgentID
	case "cimd":
		// For CIMD, the client_id in the authorize URL is the metadata URI, not the broker UUID.
		agentID = h.cfg.OAuth2.CIMDClientURI
	default:
		clientType = "proxy"
		agentID = h.cfg.OAuth2.ProxyAgentID
		if agentID == "" {
			agentID = h.cfg.OAuth2.ClientID // fallback for non-compose environments
		}
	}

	if agentID == "" {
		http.Error(w, "agent ID not available for type "+clientType+" — seed data may still be loading", http.StatusServiceUnavailable)
		return
	}

	state := generateRandomString(32)
	sessionID := generateRandomString(32)

	// Build per-flow OAuth2 config with the selected agent's credentials.
	flowConfig := *h.oauth2Config
	flowConfig.ClientID = agentID
	// Local and CIMD agents are public clients (no client secret).
	if clientType == "local" || clientType == "cimd" {
		flowConfig.ClientSecret = ""
	}

	h.sessionsMu.Lock()
	h.sessions[sessionID] = &Session{
		CreatedAt:    time.Now(),
		CSRFState:    state,
		UserInfo:     &UserInfo{},
		ClientType:   clientType,
		OAuth2Config: &flowConfig,
	}
	h.sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	authURL := flowConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)

	slog.Info("Initiating OAuth2 authorization flow",
		"client_type", clientType,
		"client_id", agentID,
		"state", state[:8])

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Callback handles the OAuth2 callback from the authorization server
func (h *Handlers) Callback(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionID(r)
	if sessionID == "" {
		http.Error(w, "Session not found", http.StatusBadRequest)
		return
	}

	// Get authorization code and state
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errMsg := r.URL.Query().Get("error")

	if errMsg != "" {
		errDesc := r.URL.Query().Get("error_description")
		slog.Error("OAuth2 authorization error",
			"error", errMsg,
			"error_description", errDesc)
		http.Error(w, fmt.Sprintf("Authorization error: %s", errMsg), http.StatusBadRequest)
		return
	}

	if code == "" {
		slog.Error("Missing authorization code in callback")
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Validate CSRF state
	h.sessionsMu.RLock()
	storedState := h.sessions[sessionID].CSRFState
	h.sessionsMu.RUnlock()

	if state != storedState {
		slog.Error("CSRF state mismatch in callback",
			"expected_state_length", len(storedState),
			"received_state_length", len(state))
		http.Error(w, "Invalid CSRF state", http.StatusBadRequest)
		return
	}

	slog.Info("Received authorization code and valid CSRF state",
		"code_length", len(code),
		"state_length", len(state))

	// Exchange code for token using the per-flow config stored at login time.
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.sessionsMu.RLock()
	flowCfg := h.sessions[sessionID].OAuth2Config
	h.sessionsMu.RUnlock()
	if flowCfg == nil {
		flowCfg = h.oauth2Config // fallback for sessions started before this change
	}

	token, err := flowCfg.Exchange(ctx, code)
	if err != nil {
		slog.Error("Failed to exchange code for token",
			"error", err.Error())
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully exchanged code for token",
		"token_type", token.TokenType,
		"token_expiry", token.Expiry.Format(time.RFC3339))

	// Extract 'sub' claim from access token
	// The golang oauth2 library stores the access token in the AccessToken field, not in Extra
	accessToken := token.AccessToken
	if accessToken == "" {
		slog.Error("Failed to extract access token from response")
		http.Error(w, "Failed to extract access token", http.StatusInternalServerError)
		return
	}

	// Try to extract sub claim from JWT token
	// If token is not JWT format (opaque), fall back to "N/A"
	sub := "N/A"
	if decodedSub, err := extractSubFromToken(accessToken); err == nil {
		sub = decodedSub
		slog.Info("Extracted sub claim from JWT token", "sub", sub)
	} else {
		slog.Info("Token is not JWT format (opaque token), using placeholder", "error", err.Error())
	}

	// Store token and user info in session
	h.sessionsMu.Lock()
	if session, exists := h.sessions[sessionID]; exists {
		session.Token = token
		session.ExpiresAt = token.Expiry.Unix()
		session.UserInfo.Sub = sub
		slog.Info("Updated session in callback",
			"sub", session.UserInfo.Sub,
			"expiresAt", session.ExpiresAt)
	} else {
		slog.Error("Session not found in callback")
	}
	h.sessionsMu.Unlock()

	// Redirect to home
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout clears the session
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	sessionID := h.getSessionID(r)

	if sessionID != "" {
		h.sessionsMu.Lock()
		delete(h.sessions, sessionID)
		h.sessionsMu.Unlock()
	}

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// CallMCP handles MCP tool calls with transparent token exchange through agentgateway.
// It uses the mcp-go client library for full MCP protocol compliance.
func (h *Handlers) CallMCP(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "method_not_allowed",
		})
		return
	}

	// Extract session and validate token
	sessionID := h.getSessionID(r)
	if sessionID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "unauthorized",
			"error_description": "session not found",
		})
		return
	}

	h.sessionsMu.RLock()
	session, exists := h.sessions[sessionID]
	h.sessionsMu.RUnlock()

	if !exists || !session.Token.Valid() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "unauthorized",
			"error_description": "session token invalid or expired",
		})
		return
	}

	// Get agentgateway URL from config
	gatewayURL := h.cfg.AgentGateway.MCPURL
	if gatewayURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "gateway_not_configured",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Create mcp-go streamable HTTP client with Bearer token injection
	accessToken := session.Token.AccessToken
	mcpClient, err := mcpclient.NewStreamableHttpClient(gatewayURL,
		transport.WithHTTPHeaderFunc(func(_ context.Context) map[string]string {
			return map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", accessToken),
			}
		}),
		transport.WithHTTPTimeout(15*time.Second),
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "mcp_call_failed",
			"error_description": "failed to create MCP client",
		})
		slog.Error("Failed to create MCP client", "error", err.Error())
		return
	}
	defer mcpClient.Close()

	// Start the client transport
	if err := mcpClient.Start(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "mcp_call_failed",
			"error_description": "failed to start MCP client",
		})
		slog.Error("MCP client start failed", "error", err.Error())
		return
	}

	// Step 1: initialize MCP session
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "sample-agent",
		Version: "1.0",
	}

	_, err = mcpClient.Initialize(ctx, initReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "mcp_call_failed",
			"error_description": "initialize request failed",
		})
		slog.Error("MCP initialize call failed", "error", err.Error())
		return
	}

	// Step 2: call MCP tools - whoami tool to demonstrate token exchange
	toolReq := mcp.CallToolRequest{}
	toolReq.Params.Name = "whoami"

	toolResult, err := mcpClient.CallTool(ctx, toolReq)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]string{
			"error":             "mcp_call_failed",
			"error_description": "tools/call request failed",
		})
		slog.Error("MCP tools/call failed", "error", err.Error())
		return
	}

	// Extract result text and JWT claims from the tool response
	var toolResultText string
	var jwtClaims map[string]interface{}

	if toolResult != nil && len(toolResult.Content) > 0 {
		if tc, ok := mcp.AsTextContent(toolResult.Content[0]); ok {
			toolResultText = tc.Text
			slog.Info("MCP tool text content", "text", toolResultText)
			// Parse jwt_claims embedded in the text: "auth_scheme=Bearer\njwt_claims={...}"
			if idx := strings.Index(toolResultText, "\njwt_claims="); idx >= 0 {
				claimsJSON := toolResultText[idx+len("\njwt_claims="):]
				var claims map[string]interface{}
				if err := json.Unmarshal([]byte(claimsJSON), &claims); err == nil {
					jwtClaims = claims
					slog.Info("JWT claims parsed from MCP tool text",
						"claims_count", len(claims))
				}
			}
		}
	}

	slog.Info("MCP call successful",
		"tool", "whoami",
		"has_jwt_claims", jwtClaims != nil)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"tool_result": toolResultText,
		"gateway_url": gatewayURL,
		"jwt_claims":  jwtClaims,
	})
}

// getSessionID retrieves the session ID from cookies
func (h *Handlers) getSessionID(r *http.Request) string {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return ""
	}
	return cookie.Value
}

// extractSubFromToken extracts the 'sub' claim from a JWT access token without verification
func extractSubFromToken(accessToken string) (string, error) {
	// JWT format: header.payload.signature
	parts := strings.Split(accessToken, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid token format")
	}

	// Decode the payload (second part)
	payload := parts[1]

	// Add padding if needed for base64 decoding
	switch len(payload) % 4 {
	case 1:
		payload += "==="
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode token payload: %w", err)
	}

	// Parse the JSON payload
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", fmt.Errorf("failed to parse token claims: %w", err)
	}

	// Extract the 'sub' claim
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("'sub' claim not found or not a string")
	}

	return sub, nil
}

// generateRandomString generates a cryptographically secure random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		// This should never happen in practice, but if it does, panic to avoid using weak randomness
		panic(fmt.Sprintf("failed to read random bytes: %v", err))
	}
	for i := range b {
		b[i] = charset[randomBytes[i]%byte(len(charset))]
	}
	return string(b)
}

// renderSelectorPage renders the client-type selector home page.
func renderSelectorPage(cfg *config.Config) string {
	proxyAvail := cfg.OAuth2.ProxyAgentID != ""
	localAvail := cfg.OAuth2.LocalAgentID != ""
	cimdAvail := cfg.OAuth2.CIMDClientURI != ""

	btnProxy := `<a href="/login?type=proxy" class="btn btn-proxy">Login as Proxy Client<span class="badge">forwards to upstream OAuth2</span></a>`
	if !proxyAvail {
		btnProxy = `<span class="btn btn-disabled">Proxy Client<span class="badge">not seeded yet — reload in a moment</span></span>`
	}
	btnLocal := `<a href="/login?type=local" class="btn btn-local">Login as Local Client<span class="badge">broker issues JWT locally</span></a>`
	if !localAvail {
		btnLocal = `<span class="btn btn-disabled">Local Client<span class="badge">not seeded yet — reload in a moment</span></span>`
	}
	btnCIMD := `<a href="/login?type=cimd" class="btn btn-cimd">Login as CIMD Client<span class="badge">resolved via metadata doc (placeholder URL)</span></a>`
	if !cimdAvail {
		btnCIMD = `<span class="btn btn-disabled">CIMD Client<span class="badge">not seeded yet — reload in a moment</span></span>`
	}

	return `<!DOCTYPE html>
<html>
<head>
    <title>Sample OAuth2 Client</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 640px;
            margin: 80px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 8px 16px rgba(0,0,0,0.2); }
        h1 { color: #333; margin-bottom: 4px; }
        .subtitle { color: #666; margin-bottom: 32px; }
        .btn {
            display: flex; flex-direction: column; align-items: flex-start;
            width: 100%%; padding: 14px 20px; margin-bottom: 12px;
            border-radius: 6px; font-size: 16px; font-weight: 600;
            text-decoration: none; color: white; cursor: pointer;
            transition: opacity 0.15s; box-sizing: border-box;
        }
        .btn:hover { opacity: 0.88; }
        .btn-proxy  { background: #667eea; }
        .btn-local  { background: #27ae60; }
        .btn-cimd   { background: #e67e22; }
        .btn-disabled { background: #bdc3c7; cursor: default; }
        .btn-disabled:hover { opacity: 1; }
        .badge { font-size: 12px; font-weight: 400; opacity: 0.85; margin-top: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 Sample OAuth2 Client</h1>
        <p class="subtitle">Select a client type to start an authorization flow through the identity broker</p>
        ` + btnProxy + `
        ` + btnLocal + `
        ` + btnCIMD + `
    </div>
</body>
</html>`
}

// renderUserPage renders the post-login user info page.
func renderUserPage(userInfo *UserInfo, expiresAt int64, clientType string) string {
	expiresTime := time.Unix(expiresAt, 0).Format(time.RFC3339)
	slog.Info("renderUserPage called with",
		"sub", userInfo.Sub,
		"expiresTime", expiresTime,
		"clientType", clientType)

	var flowLabel string
	switch clientType {
	case "local":
		flowLabel = "Local Client — token issued directly by broker (no upstream)"
	case "cimd":
		flowLabel = "CIMD Client — agent resolved via Client ID Metadata Document"
	default:
		flowLabel = "Proxy Client — token forwarded from upstream OAuth2 server"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>User Info - Sample OAuth2 Client</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 600px; margin: 80px auto; padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
        }
        .container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 8px 16px rgba(0,0,0,0.2); }
        h1 { color: #333; margin-bottom: 10px; }
        .flow-badge {
            background: #eef0ff; border-left: 4px solid #667eea;
            padding: 12px 16px; border-radius: 4px; margin-bottom: 24px;
            color: #3d4db7; font-size: 14px;
        }
        .user-info {
            background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%);
            padding: 20px; border-radius: 4px; margin-bottom: 24px;
        }
        .info-row { margin: 10px 0; padding: 10px; background: white; border-radius: 4px; }
        .info-row strong { color: #667eea; min-width: 110px; display: inline-block; }
        .buttons { display: flex; gap: 10px; margin-top: 16px; }
        a, button {
            display: inline-block; padding: 10px 20px; text-decoration: none;
            border-radius: 4px; font-weight: bold; border: none; cursor: pointer;
            transition: background-color 0.2s;
        }
        .btn-home  { background: #667eea; color: white; flex: 1; }
        .btn-home:hover  { background: #764ba2; }
        .btn-logout { background: #e74c3c; color: white; flex: 1; }
        .btn-logout:hover { background: #c0392b; }
        .btn-mcp {
            background: #27ae60; color: white; width: 100%%;
            margin-bottom: 10px; font-size: 14px;
        }
        .btn-mcp:hover { background: #219a52; }
        .btn-mcp:disabled { background: #95a5a6; cursor: not-allowed; }
        .mcp-result { margin-top: 15px; padding: 15px; border-radius: 4px; display: none; }
        .mcp-result.success { background: #d4edda; border: 1px solid #c3e6cb; color: #155724; }
        .mcp-result.error   { background: #f8d7da; border: 1px solid #f5c6cb; color: #721c24; }
        .mcp-result pre { margin: 8px 0 0 0; font-size: 13px; white-space: pre-wrap; word-break: break-all; }
    </style>
</head>
<body>
    <div class="container">
        <h1>✓ OAuth2 Flow Successful</h1>
        <div class="flow-badge">%s</div>
        <div class="user-info">
            <div class="info-row"><strong>Subject:</strong> <span>%s</span></div>
            <div class="info-row"><strong>Token Expires:</strong> <span>%s</span></div>
        </div>
        <button id="mcp-btn" class="btn-mcp" onclick="callMCPTool()">Call MCP Tool (Token Exchange)</button>
        <div id="mcp-result" class="mcp-result"></div>
        <div class="buttons">
            <a href="/" class="btn-home">Switch Client</a>
            <a href="/logout" class="btn-logout">Logout</a>
        </div>
    </div>
    <script>
    function callMCPTool() {
        var btn = document.getElementById('mcp-btn');
        var result = document.getElementById('mcp-result');
        btn.disabled = true; btn.textContent = 'Calling...';
        result.style.display = 'none'; result.className = 'mcp-result';
        fetch('/call-mcp', {method: 'POST'})
            .then(function(r) { return r.json(); })
            .then(function(d) {
                result.style.display = 'block';
                if (d.success) {
                    result.classList.add('success');
                    var html = '<strong>Token Exchange Successful!</strong><pre>Tool: whoami\nResult: ' + esc(d.tool_result) + '\nGateway: ' + esc(d.gateway_url);
                    if (d.jwt_claims && Object.keys(d.jwt_claims).length > 0) { html += '\n\nJWT Claims:\n' + formatJSON(d.jwt_claims); }
                    html += '</pre>';
                    result.innerHTML = html;
                } else {
                    result.classList.add('error');
                    result.innerHTML = '<strong>MCP Call Failed</strong><pre>' + esc(d.error) + '\nGateway: ' + esc(d.gateway_url) + '</pre>';
                }
            })
            .catch(function(e) { result.style.display='block'; result.classList.add('error'); result.innerHTML='<strong>Request Failed</strong><pre>'+esc(String(e))+'</pre>'; })
            .finally(function() { btn.disabled=false; btn.textContent='Call MCP Tool (Token Exchange)'; });
    }
    function formatJSON(obj) { return esc(JSON.stringify(obj, null, 2)); }
    function esc(s) { return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
    </script>
</body>
</html>`, flowLabel, userInfo.Sub, expiresTime)
}
