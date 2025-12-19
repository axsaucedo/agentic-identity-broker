package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRFProtection_GetRequest(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	// Add principal to context
	req = req.WithContext(req.Context())
	req.RemoteAddr = "127.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// GET requests should pass through
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check if cookie was set
	cookies := rr.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == CSRFCookieName {
			found = true
			if cookie.Value == "" {
				t.Error("CSRF cookie value should not be empty")
			}
		}
	}
	if !found {
		t.Error("CSRF cookie should be set on GET request")
	}
}

func TestCSRFProtection_PostWithoutToken(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST without CSRF token should fail
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCSRFProtection_PostWithValidToken(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "test-session"
	token := "test-csrf-token"

	// Pre-store the token
	store.Set(sessionID, token)

	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = sessionID
	req.Header.Set(CSRFTokenHeader, token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST with valid CSRF token should succeed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestCSRFProtection_PostWithInvalidToken(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "test-session"
	correctToken := "correct-token"
	wrongToken := "wrong-token"

	// Pre-store the correct token
	store.Set(sessionID, correctToken)

	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = sessionID
	req.Header.Set(CSRFTokenHeader, wrongToken)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST with invalid CSRF token should fail
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCSRFStore_SetAndGet(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "test-session"
	token := "test-token"

	// Set token
	store.Set(sessionID, token)

	// Get token
	retrievedToken, exists := store.Get(sessionID)
	if !exists {
		t.Error("token should exist")
	}
	if retrievedToken != token {
		t.Errorf("expected token %s, got %s", token, retrievedToken)
	}
}

func TestCSRFStore_Validate(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "test-session"
	token := "test-token"

	// Set token
	store.Set(sessionID, token)

	// Validate correct token
	if !store.Validate(sessionID, token) {
		t.Error("validation should succeed for correct token")
	}

	// Validate incorrect token
	if store.Validate(sessionID, "wrong-token") {
		t.Error("validation should fail for incorrect token")
	}

	// Validate non-existent session
	if store.Validate("non-existent", token) {
		t.Error("validation should fail for non-existent session")
	}
}

func TestGenerateToken(t *testing.T) {
	token1, err := generateToken()
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	token2, err := generateToken()
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	// Tokens should be non-empty
	if token1 == "" || token2 == "" {
		t.Error("generated tokens should not be empty")
	}

	// Tokens should be unique
	if token1 == token2 {
		t.Error("generated tokens should be unique")
	}

	// Token length should be reasonable (base64 encoded 32 bytes)
	if len(token1) < 40 {
		t.Errorf("token length seems too short: %d", len(token1))
	}
}
