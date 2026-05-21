package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

const (
	// CSRFTokenHeader is the header name for CSRF token
	CSRFTokenHeader = "X-CSRF-Token"
	// CSRFCookieName is the cookie name for CSRF token
	CSRFCookieName = "csrf_token"
	// CSRFTokenLength is the length of generated CSRF tokens in bytes
	CSRFTokenLength = 32
	// CSRFTokenTTL is the time-to-live for CSRF tokens
	CSRFTokenTTL = 24 * time.Hour
)

// csrfToken represents a CSRF token with expiration
type csrfToken struct {
	Value     string
	ExpiresAt time.Time
}

// CSRFStore manages CSRF tokens in memory
type CSRFStore struct {
	mu     sync.RWMutex
	tokens map[string]*csrfToken
	logger *slog.Logger
}

// NewCSRFStore creates a new CSRF token store
func NewCSRFStore(logger *slog.Logger) *CSRFStore {
	store := &CSRFStore{
		tokens: make(map[string]*csrfToken),
		logger: logger,
	}

	// Start cleanup goroutine
	go store.cleanupExpiredTokens()

	return store
}

// generateToken generates a random CSRF token
func generateToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// Get retrieves a CSRF token for a session
func (s *CSRFStore) Get(sessionID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[sessionID]
	if !exists {
		return "", false
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		return "", false
	}

	return token.Value, true
}

// GetOrCreate atomically retrieves an existing token or generates a new one.
// This eliminates the race condition when parallel GET requests arrive simultaneously.
func (s *CSRFStore) GetOrCreate(sessionID string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if token, exists := s.tokens[sessionID]; exists && time.Now().Before(token.ExpiresAt) {
		return token.Value, nil
	}

	tokenValue, err := generateToken()
	if err != nil {
		return "", err
	}

	s.tokens[sessionID] = &csrfToken{
		Value:     tokenValue,
		ExpiresAt: time.Now().Add(CSRFTokenTTL),
	}

	return tokenValue, nil
}

// Set stores a CSRF token for a session
func (s *CSRFStore) Set(sessionID, tokenValue string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[sessionID] = &csrfToken{
		Value:     tokenValue,
		ExpiresAt: time.Now().Add(CSRFTokenTTL),
	}
}

// Validate checks if a token is valid for a session
func (s *CSRFStore) Validate(sessionID, tokenValue string) bool {
	storedToken, exists := s.Get(sessionID)
	if !exists {
		return false
	}

	return storedToken == tokenValue
}

// cleanupExpiredTokens periodically removes expired tokens
func (s *CSRFStore) cleanupExpiredTokens() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for sessionID, token := range s.tokens {
			if now.After(token.ExpiresAt) {
				delete(s.tokens, sessionID)
			}
		}
		s.mu.Unlock()
	}
}

// CSRFProtection returns a middleware that validates CSRF tokens
// For GET requests: generates and sets CSRF token in cookie
// For POST/PUT/DELETE: validates token from header matches cookie
func CSRFProtection(store *CSRFStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF for safe methods (GET, HEAD, OPTIONS)
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				// Ensure CSRF token exists and set cookie for future mutating requests
				sessionID := getSessionID(r)
				if sessionID != "" {
					token, err := store.GetOrCreate(sessionID)
					if err != nil {
						store.logger.Error("failed to generate CSRF token", "error", err)
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}

					http.SetCookie(w, &http.Cookie{
						Name:     CSRFCookieName,
						Value:    token,
						Path:     "/",
						HttpOnly: false, // JavaScript needs to read this
						Secure:   r.TLS != nil,
						SameSite: http.SameSiteStrictMode,
						MaxAge:   int(CSRFTokenTTL.Seconds()),
					})
				}

				next.ServeHTTP(w, r)
				return
			}

			// For mutating methods, validate CSRF token
			sessionID := getSessionID(r)
			if sessionID == "" {
				store.logger.Warn("CSRF validation failed: no session ID")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Get token from header
			headerToken := r.Header.Get(CSRFTokenHeader)
			if headerToken == "" {
				store.logger.Warn("CSRF validation failed: no token in header",
					"session_id", sessionID)
				http.Error(w, "Forbidden: CSRF token required", http.StatusForbidden)
				return
			}

			// Validate token
			if !store.Validate(sessionID, headerToken) {
				store.logger.Warn("CSRF validation failed: invalid token",
					"session_id", sessionID)
				http.Error(w, "Forbidden: Invalid CSRF token", http.StatusForbidden)
				return
			}

			// Token valid, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// getSessionID extracts session ID from request.
// Returns the authenticated principal, or empty string if none is present.
// Callers must treat empty string as unauthenticated; the CSRF middleware
// rejects mutating requests without a session ID.
func getSessionID(r *http.Request) string {
	if principalValue, ok := principal.FromContext(r.Context()); ok && principalValue != "" {
		return principalValue
	}
	return ""
}
