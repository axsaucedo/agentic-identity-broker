package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/agentic-identity-broker/mock-upstream-oauth2-service/internal/config"
)

func TestHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	Health(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "upstream-oauth2-server") {
		t.Errorf("expected response to contain 'upstream-oauth2-server', got %s", w.Body.String())
	}
}

func TestConsentPageDistinctiveUI(t *testing.T) {
	cfg := &config.Config{
		OAuth2: config.OAuth2Config{
			Scopes: []config.ScopeConfig{
				{Name: "openid", Description: "OpenID Connect scope"},
				{Name: "profile", Description: "Access user profile"},
			},
		},
	}

	// Mock the user authorization handler
	handler := NewUserAuthorizationHandler(cfg)

	// Create a test request
	req := httptest.NewRequest("GET", "/oauth/authorize?client_id=test-client&response_type=code&redirect_uri=http://localhost/callback&scope=openid+profile&state=test123", nil)
	w := httptest.NewRecorder()

	// Call the handler
	_, err := handler(w, req)
	if err != nil {
		t.Errorf("handler returned error: %v", err)
	}

	body := w.Body.String()

	// Test 1: Verify distinctive teal background color
	if !strings.Contains(body, "#00d4aa") {
		t.Errorf("expected teal background color #00d4aa, not found in response")
	}

	// Test 2: Verify distinctive upstream badge
	if !strings.Contains(body, "UPSTREAM OAUTH2") {
		t.Errorf("expected 'UPSTREAM OAUTH2' badge, not found in response")
	}

	// Test 3: Verify server identifier
	if !strings.Contains(body, "Upstream OAuth2 Server") {
		t.Errorf("expected 'Upstream OAuth2 Server' identifier, not found in response")
	}

	// Test 4: Verify globe emoji badge
	if !strings.Contains(body, "🌐") {
		t.Errorf("expected globe emoji in badge, not found in response")
	}

	// Test 5: Verify form fields
	if !strings.Contains(body, `name="client_id"`) {
		t.Errorf("expected client_id form field, not found")
	}
	if !strings.Contains(body, `name="approval"`) {
		t.Errorf("expected approval form field, not found")
	}

	// Test 6: Verify approve/deny buttons with checkmarks
	if !strings.Contains(body, "✓ Approve") {
		t.Errorf("expected approve button with checkmark, not found")
	}
	if !strings.Contains(body, "✗ Deny") {
		t.Errorf("expected deny button with X mark, not found")
	}

	// Test 7: Verify HTTP status
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestConsentPageRequiredParams(t *testing.T) {
	cfg := &config.Config{
		OAuth2: config.OAuth2Config{
			Scopes: []config.ScopeConfig{},
		},
	}

	handler := NewUserAuthorizationHandler(cfg)

	// Test missing client_id
	req := httptest.NewRequest("GET", "/oauth/authorize?response_type=code&redirect_uri=http://localhost/callback", nil)
	w := httptest.NewRecorder()

	_, err := handler(w, req)

	if err == nil || !strings.Contains(err.Error(), "invalid_request") {
		t.Errorf("expected invalid_request error, got %v", err)
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestConsentFormSubmission(t *testing.T) {
	cfg := &config.Config{
		User: config.UserConfig{
			Sub:   "upstream-user@example.com",
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	handler := NewUserAuthorizationHandler(cfg)

	// Test approval submission
	body := strings.NewReader("approval=approve&client_id=test-client")
	req := httptest.NewRequest("POST", "/oauth/authorize", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	sub, err := handler(w, req)
	if err != nil {
		t.Errorf("handler returned error on approval: %v", err)
	}
	if sub != "upstream-user@example.com" {
		t.Errorf("expected user sub 'upstream-user@example.com', got %s", sub)
	}

	// Test denial submission
	body = strings.NewReader("approval=deny&client_id=test-client")
	req = httptest.NewRequest("POST", "/oauth/authorize", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()

	sub, err = handler(w, req)
	if err == nil || !strings.Contains(err.Error(), "access_denied") {
		t.Errorf("expected access_denied error, got %v", err)
	}
}
