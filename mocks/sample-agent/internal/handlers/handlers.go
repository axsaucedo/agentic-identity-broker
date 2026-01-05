package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/agentic-identity-broker/sample-agent/internal/config"
	"golang.org/x/oauth2"
)

// UserInfo represents user information retrieved from OAuth2
type UserInfo struct {
	Sub string `json:"sub"`
}

// Session represents a user session
type Session struct {
	Token     *oauth2.Token
	UserInfo  *UserInfo
	CreatedAt time.Time
	ExpiresAt int64
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
		// User is logged in, show user info
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		slog.Info("Displaying user info",
			"sub", session.UserInfo.Sub,
			"expiresAt", session.ExpiresAt,
			"sessionID", sessionID)
		fmt.Fprint(w, renderUserPage(session.UserInfo, session.ExpiresAt, sessionID))
		return
	}

	// User is not logged in, show login page
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, renderLoginPage())
}

// Login initiates the OAuth2 authorization flow
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	// Generate state for CSRF protection
	state := generateRandomString(32)

	// Create session
	sessionID := generateRandomString(32)
	h.sessionsMu.Lock()
	h.sessions[sessionID] = &Session{
		CreatedAt: time.Now(),
	}
	h.sessionsMu.Unlock()

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	// Store state in session
	h.sessionsMu.Lock()
	h.sessions[sessionID].UserInfo = &UserInfo{}
	h.sessionsMu.Unlock()

	// Generate authorization URL
	authURL := h.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOnline)

	slog.Info("Initiating OAuth2 authorization flow",
		"client_id", h.oauth2Config.ClientID,
		"state", state[:8],
		"auth_url_host", authURL[:50])

	// Redirect to authorization server
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

	slog.Info("Received authorization code",
		"code_length", len(code),
		"state_length", len(state))

	// Exchange code for token
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	token, err := h.oauth2Config.Exchange(ctx, code)
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
	accessToken, ok := token.Extra("access_token").(string)
	if !ok {
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
			"sessionID", sessionID,
			"sub", session.UserInfo.Sub,
			"expiresAt", session.ExpiresAt)
	} else {
		slog.Error("Session not found in callback", "sessionID", sessionID)
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

// generateRandomString generates a random string of given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// renderLoginPage renders the login page HTML
func renderLoginPage() string {
	return `
<!DOCTYPE html>
<html>
<head>
    <title>Sample OAuth2 Client</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 600px;
            margin: 100px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 8px 16px rgba(0,0,0,0.2);
            text-align: center;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
        }
        .info {
            background-color: #f5f5f5;
            padding: 20px;
            border-radius: 4px;
            margin-bottom: 30px;
            text-align: left;
            border-left: 4px solid #667eea;
        }
        .info p {
            margin: 8px 0;
            color: #555;
        }
        .info strong {
            color: #333;
        }
        a.button {
            display: inline-block;
            background-color: #667eea;
            color: white;
            padding: 12px 30px;
            text-decoration: none;
            border-radius: 4px;
            font-weight: bold;
            transition: background-color 0.2s;
        }
        a.button:hover {
            background-color: #764ba2;
        }
        .flow-diagram {
            margin-top: 30px;
            padding-top: 30px;
            border-top: 1px solid #eee;
            text-align: left;
        }
        .flow-step {
            margin: 10px 0;
            padding: 8px;
            background: #f9f9f9;
            border-left: 3px solid #667eea;
            padding-left: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 Sample OAuth2 Client</h1>
        <p class="subtitle">End-to-End OAuth2 Flow Testing</p>

        <div class="info">
            <p><strong>This Sample Agent demonstrates the complete OAuth2 flow:</strong></p>
            <div class="flow-diagram">
                <div class="flow-step">① Sample Agent (this app) ← OAuth2 Client</div>
                <div class="flow-step">② Identity Broker ← OAuth2 Authorization Server</div>
                <div class="flow-step">③ Upstream OAuth2 Mock ← Actual Authorization Server</div>
            </div>
        </div>

        <p style="margin-bottom: 30px; color: #666;">
            Click "Login" to start the OAuth2 authorization flow through the identity broker.
        </p>

        <a href="/login" class="button">Login with OAuth2</a>

        <div class="info" style="margin-top: 30px;">
            <p><strong>Flow Details:</strong></p>
            <p>• Client ID: upstream-oauth2-client</p>
            <p>• Authorization Server: http://localhost:8000 (Identity Broker)</p>
            <p>• Upstream Server: http://localhost:9001 (Mock OAuth2)</p>
            <p>• Scopes: openid, profile, email</p>
        </div>
    </div>
</body>
</html>
`
}

// renderUserPage renders the user information page
func renderUserPage(userInfo *UserInfo, expiresAt int64, sessionID string) string {
	expiresTime := time.Unix(expiresAt, 0).Format(time.RFC3339)
	slog.Info("renderUserPage called with",
		"sub", userInfo.Sub,
		"expiresTime", expiresTime,
		"sessionID", sessionID)
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>User Info - Sample OAuth2 Client</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            max-width: 600px;
            margin: 100px auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 8px;
            box-shadow: 0 8px 16px rgba(0,0,0,0.2);
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
        }
        .success {
            color: #27ae60;
            margin-bottom: 30px;
            font-size: 18px;
        }
        .user-info {
            background: linear-gradient(135deg, #f5f7fa 0%%, #c3cfe2 100%%);
            padding: 20px;
            border-radius: 4px;
            margin-bottom: 30px;
        }
        .info-row {
            margin: 12px 0;
            padding: 10px;
            background: white;
            border-radius: 4px;
        }
        .info-row strong {
            color: #667eea;
            min-width: 100px;
            display: inline-block;
        }
        .info-row span {
            color: #333;
        }
        .flow-success {
            background-color: #d4edda;
            border: 1px solid #c3e6cb;
            padding: 15px;
            border-radius: 4px;
            margin-bottom: 30px;
            color: #155724;
        }
        .buttons {
            display: flex;
            gap: 10px;
        }
        a, button {
            display: inline-block;
            padding: 10px 20px;
            text-decoration: none;
            border-radius: 4px;
            font-weight: bold;
            border: none;
            cursor: pointer;
            transition: background-color 0.2s;
        }
        .btn-home {
            background-color: #667eea;
            color: white;
            flex: 1;
        }
        .btn-home:hover {
            background-color: #764ba2;
        }
        .btn-logout {
            background-color: #e74c3c;
            color: white;
            flex: 1;
        }
        .btn-logout:hover {
            background-color: #c0392b;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>✓ OAuth2 Authentication Successful!</h1>
        <p class="success">You have successfully completed the OAuth2 flow.</p>

        <div class="flow-success">
            <strong>✓ End-to-End Flow Completed:</strong><br>
            Sample Agent → Identity Broker → Upstream OAuth2 Server
        </div>

        <div class="user-info">
            <h2 style="margin-top: 0; color: #333;">User Information</h2>
            <div class="info-row">
                <strong>Subject:</strong>
                <span>%s</span>
            </div>
            <div class="info-row">
                <strong>Token Expires:</strong>
                <span>%s</span>
            </div>
            <div class="info-row">
                <strong>Session ID:</strong>
                <span style="font-family: monospace; font-size: 12px;">%s</span>
            </div>
        </div>

        <div class="buttons">
            <a href="/" class="btn-home">Home</a>
            <a href="/logout" class="btn-logout">Logout</a>
        </div>
    </div>
</body>
</html>
`, userInfo.Sub, expiresTime, sessionID)
}
