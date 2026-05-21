package middleware

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

func TestCSRFProtection_GetRequest(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))
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

	if _, exists := store.Get("user@example.com"); !exists {
		t.Error("CSRF token should be stored for the authenticated principal")
	}
}

func TestCSRFProtection_PostWithoutToken(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req = req.WithContext(principal.WithPrincipal(req.Context(), "user@example.com"))
	req.RemoteAddr = "127.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST without CSRF token should fail
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCSRFProtection_PostWithValidTokenForPrincipalIgnoresRemoteAddr(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "user@example.com"
	token := "test-csrf-token"

	// Pre-store the token
	store.Set(sessionID, token)

	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req = req.WithContext(principal.WithPrincipal(req.Context(), sessionID))
	req.RemoteAddr = "127.0.0.1:54321"
	req.Header.Set(CSRFTokenHeader, token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST with valid CSRF token should succeed
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestCSRFProtection_PostWithDifferentPrincipalFailsEvenWhenRemoteAddrMatches(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	correctPrincipal := "user@example.com"
	wrongPrincipal := "other@example.com"
	correctToken := "correct-token"

	// Pre-store the correct token
	store.Set(correctPrincipal, correctToken)

	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req = req.WithContext(principal.WithPrincipal(req.Context(), wrongPrincipal))
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set(CSRFTokenHeader, correctToken)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// POST with the wrong principal should fail even if the address is unchanged.
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCSRFProtection_PostWithoutPrincipalForbidden(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	// Pre-store a token keyed by RemoteAddr to verify it is never consulted.
	store.Set("127.0.0.1:12345", "test-csrf-token")

	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set(CSRFTokenHeader, "test-csrf-token")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Must be 403: no principal means no session ID; RemoteAddr must not be used as fallback.
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestCSRFProtection_GetWithoutPrincipalPassesWithoutCookie(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// GET passes through even without a principal, but no CSRF cookie is set.
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			t.Error("CSRF cookie must not be set when no principal is present")
		}
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

func TestCSRFStore_GetOrCreate_Atomic(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	sessionID := "concurrent-session"

	// Launch multiple goroutines calling GetOrCreate simultaneously
	// to verify they all return the same token (no race condition).
	const goroutines = 50
	results := make(chan string, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			token, err := store.GetOrCreate(sessionID)
			if err != nil {
				t.Errorf("GetOrCreate failed: %v", err)
				results <- ""
				return
			}
			results <- token
		}()
	}

	// Collect all results
	var firstToken string
	for i := 0; i < goroutines; i++ {
		token := <-results
		if token == "" {
			continue
		}
		if firstToken == "" {
			firstToken = token
		}
		if token != firstToken {
			t.Errorf("GetOrCreate returned different tokens: got %q, want %q", token, firstToken)
		}
	}

	if firstToken == "" {
		t.Fatal("no tokens were generated")
	}
}

func TestCSRFProtection_ConcurrentGetsProduceSameCookie(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	sessionID := "user@concurrent.com"
	const requests = 10
	cookies := make(chan string, requests)

	// Simulate concurrent GET requests (like the frontend's Promise.all)
	for i := 0; i < requests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/consent/agents/123", nil)
			req = req.WithContext(principal.WithPrincipal(req.Context(), sessionID))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			for _, cookie := range rr.Result().Cookies() {
				if cookie.Name == CSRFCookieName {
					cookies <- cookie.Value
					return
				}
			}
			cookies <- ""
		}()
	}

	// All responses must have the same CSRF cookie value
	var firstCookie string
	for i := 0; i < requests; i++ {
		cookie := <-cookies
		if cookie == "" {
			t.Error("CSRF cookie should be set on GET response")
			continue
		}
		if firstCookie == "" {
			firstCookie = cookie
		}
		if cookie != firstCookie {
			t.Errorf("concurrent GET requests produced different CSRF cookies: %q vs %q", cookie, firstCookie)
		}
	}

	// Verify the stored token matches the cookie
	storedToken, exists := store.Get(sessionID)
	if !exists {
		t.Fatal("token should exist in store")
	}
	if storedToken != firstCookie {
		t.Errorf("stored token %q does not match cookie %q", storedToken, firstCookie)
	}
}

func TestCSRFProtection_GetThenPostFlow(t *testing.T) {
	store := NewCSRFStore(slog.Default())
	handler := CSRFProtection(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))

	sessionID := "flow-test@example.com"

	// Step 1: GET request sets the CSRF cookie
	getReq := httptest.NewRequest("GET", "/api/consent/agents/123", nil)
	getReq = getReq.WithContext(principal.WithPrincipal(getReq.Context(), sessionID))
	getRR := httptest.NewRecorder()
	handler.ServeHTTP(getRR, getReq)

	if getRR.Code != http.StatusOK {
		t.Fatalf("GET request failed: %d", getRR.Code)
	}

	// Extract CSRF cookie
	var csrfToken string
	for _, cookie := range getRR.Result().Cookies() {
		if cookie.Name == CSRFCookieName {
			csrfToken = cookie.Value
			break
		}
	}
	if csrfToken == "" {
		t.Fatal("CSRF cookie not set after GET request")
	}

	// Step 2: POST request with CSRF token header should succeed
	postReq := httptest.NewRequest("POST", "/api/consent/agents/123/grants", nil)
	postReq = postReq.WithContext(principal.WithPrincipal(postReq.Context(), sessionID))
	postReq.Header.Set(CSRFTokenHeader, csrfToken)
	postRR := httptest.NewRecorder()
	handler.ServeHTTP(postRR, postReq)

	if postRR.Code != http.StatusOK {
		t.Errorf("POST with valid CSRF token should succeed, got status %d", postRR.Code)
	}
}
