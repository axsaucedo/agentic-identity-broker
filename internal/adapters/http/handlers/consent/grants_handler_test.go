package consent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domjwe "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	domotp2 "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestJWETokenService returns a real JWE token service backed by a deterministic test key.
func newTestJWETokenService() *domjwe.TokenService {
	keyBytes, err := base64.StdEncoding.DecodeString("ASNFZ4mrze/+3LqYdlQyEAEjRWeJq83v/ty6mHZUMhA=")
	if err != nil {
		panic("grants_handler_test: failed to decode test JWE key: " + err.Error())
	}
	jweKey, err := jwk.Import(keyBytes)
	if err != nil {
		panic("grants_handler_test: failed to import test JWE key: " + err.Error())
	}
	return domjwe.New(jweKey)
}

// newTestSessionToken creates a valid JWE session token for the given agent and principal.
func newTestSessionToken(ts *domjwe.TokenService, agentID id.AgentID, principalVal string, originalURL string) string {
	claims := domotp2.NewAuthorizationSessionClaims(
		agentID, id.Principal(principalVal),
		"https://agent.example.com/client",
		originalURL,
		"https://agent.example.com/callback",
		"read", "state-xyz", "challenge", "S256", nil,
	)
	token, err := ts.Encrypt(claims)
	if err != nil {
		panic("newTestSessionToken: failed to encrypt claims: " + err.Error())
	}
	return token
}

// Helper to create request with principal context
func newRequestWithPrincipal(method, path, principalValue string, body any) *http.Request {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req := httptest.NewRequest(method, path, reqBody)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	return req.WithContext(ctx)
}

func TestGetGrants_NoPrincipal(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil)
	testAgentID := id.NewAgentID()

	// Create request without principal
	req := httptest.NewRequest("GET", "/api/consent/agent/"+testAgentID.String()+"/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetGrants(rr, req)

	// Verify response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got '%s'", errResp.Error)
	}
}

func TestCreateGrant_NoPrincipal(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil)
	testAgentID := id.NewAgentID()

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", bytes.NewBuffer(jsonBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestCreateGrant_InvalidJSON(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil)
	testAgentID := id.NewAgentID()

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	req.Body = httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("invalid json"))).Body

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify response
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "invalid request" {
		t.Errorf("expected error 'invalid request', got '%s'", errResp.Error)
	}
}

func TestCreateGrant_ValidUntilInPast(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil)
	testAgentID := id.NewAgentID()

	pastTime := time.Now().Add(-1 * time.Hour)
	reqBody := GrantRequest{
		ValidUntil: &pastTime,
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", bytes.NewBuffer(jsonBody))
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(ctx, chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify response
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "invalid request" {
		t.Errorf("expected error 'invalid request', got '%s'", errResp.Error)
	}
	if errResp.Message != "valid_until must be in the future" {
		t.Errorf("expected message about valid_until, got '%s'", errResp.Message)
	}
}

func TestToGrantResponse(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil)

	testGrantID := id.NewGrantID()
	testAgentID := id.NewAgentID()
	testServiceID1 := id.NewServiceID()
	testServiceID2 := id.NewServiceID()

	validUntil := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	grant := &storage.UserGrant{
		ID:         testGrantID,
		Principal:  id.Principal("user@example.com"),
		AgentID:    testAgentID,
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: testServiceID1,
				Scopes:                    []string{"repo", "user"},
			},
			{
				ThirdpartyOAuth2ServiceID: testServiceID2,
				Scopes:                    []string{"email"},
			},
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	response := handler.toGrantResponse(grant)

	// Verify basic fields
	if response.ID != testGrantID.String() {
		t.Errorf("expected ID '%s', got '%s'", testGrantID.String(), response.ID)
	}
	if response.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got '%s'", response.Principal)
	}
	if response.AgentID != testAgentID.String() {
		t.Errorf("expected agentID '%s', got '%s'", testAgentID.String(), response.AgentID)
	}

	// Verify valid_until
	if response.ValidUntil == nil {
		t.Error("expected valid_until to be set")
	} else if !response.ValidUntil.Equal(validUntil) {
		t.Errorf("expected valid_until %v, got %v", validUntil, *response.ValidUntil)
	}

	// Verify tokens
	if len(response.DelegatedOAuth2Tokens) != 2 {
		t.Fatalf("expected 2 tokens, got %d", len(response.DelegatedOAuth2Tokens))
	}

	token1 := response.DelegatedOAuth2Tokens[0]
	if token1.ThirdpartyOAuth2ServiceID != testServiceID1.String() {
		t.Errorf("expected service '%s', got '%s'", testServiceID1.String(), token1.ThirdpartyOAuth2ServiceID)
	}
	if len(token1.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(token1.Scopes))
	}

	token2 := response.DelegatedOAuth2Tokens[1]
	if token2.ThirdpartyOAuth2ServiceID != testServiceID2.String() {
		t.Errorf("expected service '%s', got '%s'", testServiceID2.String(), token2.ThirdpartyOAuth2ServiceID)
	}
	if len(token2.Scopes) != 1 {
		t.Errorf("expected 1 scope, got %d", len(token2.Scopes))
	}

	// Verify timestamps are formatted correctly (RFC3339)
	if response.CreatedAt != "2025-01-01T00:00:00Z" {
		t.Errorf("expected created_at '2025-01-01T00:00:00Z', got '%s'", response.CreatedAt)
	}
	if response.UpdatedAt != "2025-06-01T00:00:00Z" {
		t.Errorf("expected updated_at '2025-06-01T00:00:00Z', got '%s'", response.UpdatedAt)
	}
}

// Integration-style test that verifies error handling for service errors
func TestCreateGrant_ServiceErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		serviceError   error
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "agent not found",
			serviceError:   consent.ErrAgentNotFound,
			expectedStatus: http.StatusNotFound,
			expectedError:  "agent not found",
		},
		{
			name:           "invalid scopes",
			serviceError:   consent.ErrInvalidScopes,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid scopes",
		},
		{
			name:           "service not found",
			serviceError:   consent.ErrServiceNotFound,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "service not found",
		},
		{
			name:           "grant validation failed",
			serviceError:   consent.ErrGrantValidation,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request",
		},
		{
			name:           "generic error",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testAgentID := id.NewAgentID()

			// Create mock service that returns the error
			mockService := &mockConsentService{
				grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
					return nil, tt.serviceError
				},
			}
			handler := NewGrantsHandler(mockService, nil)

			// Create valid request
			reqBody := GrantRequest{
				DelegatedOAuth2Tokens: []DelegatedTokenRequest{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
						Scopes:                    []string{"repo"},
					},
				},
			}

			req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", reqBody)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("agent-id", testAgentID.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.CreateGrant(rr, req)

			// Verify response code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Verify error message
			var errResp ErrorResponse
			if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if errResp.Error != tt.expectedError {
				t.Errorf("expected error '%s', got '%s'", tt.expectedError, errResp.Error)
			}
		})
	}
}

// TestCreateGrant_Success verifies successful grant creation
func TestCreateGrant_Success(t *testing.T) {
	t.Parallel()
	// Create mock service
	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	var capturedRequest *consent.GrantRequest
	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			capturedRequest = req
			return &storage.UserGrant{
				ID:         testGrantID,
				Principal:  id.Principal("user@example.com"),
				AgentID:    testAgentID,
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID(),
						Scopes:                    []string{"repo", "user"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	// Create request
	reqBody := GrantRequest{
		ValidUntil: &futureTime,
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo", "user"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify service was called and request was captured
	if capturedRequest == nil {
		t.Fatal("expected grant request to be passed to service")
	}
	if capturedRequest.Principal != id.Principal("user@example.com") {
		t.Errorf("expected principal 'user@example.com', got '%s'", capturedRequest.Principal)
	}
	if capturedRequest.AgentID != testAgentID {
		t.Errorf("expected agent_id '%s', got '%s'", testAgentID, capturedRequest.AgentID)
	}

	// Verify response
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var envelope map[string]GrantResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	response, ok := envelope["data"]
	if !ok {
		t.Fatal("expected 'data' field in response")
	}

	if response.ID != testGrantID.String() {
		t.Errorf("expected ID '%s', got '%s'", testGrantID.String(), response.ID)
	}
	if response.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got '%s'", response.Principal)
	}
	if response.AgentID != testAgentID.String() {
		t.Errorf("expected agent_id '%s', got '%s'", testAgentID.String(), response.AgentID)
	}
}

// TestGetGrants_Success verifies successful grant retrieval
func TestGetGrants_Success(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()
	testGrantID1 := id.NewGrantID()
	testGrantID2 := id.NewGrantID()
	now := time.Now()

	mockService := &mockConsentService{
		getActiveGrantsFunc: func(ctx context.Context, p id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
			return []*storage.UserGrant{
				{
					ID:        testGrantID1,
					Principal: id.Principal("user@example.com"),
					AgentID:   testAgentID,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: id.NewServiceID(),
							Scopes:                    []string{"repo"},
						},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        testGrantID2,
					Principal: id.Principal("user@example.com"),
					AgentID:   testAgentID,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: id.NewServiceID(),
							Scopes:                    []string{"email"},
						},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	req := newRequestWithPrincipal("GET", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetGrants(rr, req)

	// Verify response
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var envelope map[string][]GrantResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	response, ok := envelope["data"]
	if !ok {
		t.Fatal("expected 'data' field in response")
	}

	if len(response) != 2 {
		t.Errorf("expected 2 grants, got %d", len(response))
	}
}

// TestGetGrants_AgentNotFound verifies agent not found error
func TestGetGrants_AgentNotFound(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()
	mockService := &mockConsentService{
		getActiveGrantsFunc: func(ctx context.Context, p id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
			return nil, consent.ErrAgentNotFound
		},
	}
	handler := NewGrantsHandler(mockService, nil)

	req := newRequestWithPrincipal("GET", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.GetGrants(rr, req)

	// Verify response
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error != "agent not found" {
		t.Errorf("expected error 'agent not found', got '%s'", errResp.Error)
	}
}

// =========================================================================
// Tests for User Story 6: Redirect URL Validation (T049, T050)
// =========================================================================

// TestValidateRedirectURI tests the redirect URI validation function
func TestValidateRedirectURI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		redirectURI   string
		requestHost   string
		expectedValid bool
		expectError   bool
	}{
		// T049: Test case 1 - Relative URL without scheme/host
		{
			name:          "relative URL without scheme",
			redirectURI:   "/callback",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 2 - Relative URL with query params
		{
			name:          "relative URL with query params",
			redirectURI:   "/callback?code=abc",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 3 - Same-origin absolute URL (http)
		{
			name:          "same-origin absolute URL http",
			redirectURI:   "http://localhost:8000/callback",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 4 - Same-origin absolute URL (https)
		{
			name:          "same-origin absolute URL https",
			redirectURI:   "https://example.com/callback",
			requestHost:   "example.com:443", // Explicitly https
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 5 - Different origin/domain
		{
			name:          "different domain",
			redirectURI:   "https://evil.com/callback",
			requestHost:   "example.com",
			expectedValid: false,
			expectError:   false,
		},
		// T049: Test case 6 - Different scheme (http vs https)
		// Note: requestHost without port defaults to http://
		{
			name:          "different scheme",
			redirectURI:   "https://example.com/callback",
			requestHost:   "example.com:80", // Explicitly http
			expectedValid: false,
			expectError:   false,
		},
		// T049: Test case 7 - Different port
		{
			name:          "different port",
			redirectURI:   "http://localhost:8001/callback",
			requestHost:   "localhost:8000",
			expectedValid: false,
			expectError:   false,
		},
		// T049: Test case 8 - Empty redirect_uri
		{
			name:          "empty redirect_uri",
			redirectURI:   "",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 9 - URL with special characters encoded
		{
			name:          "URL with encoded special characters",
			redirectURI:   "http://localhost:8000/callback?state=%20test",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 10 - Fragment in URL
		{
			name:          "URL with fragment",
			redirectURI:   "http://localhost:8000/callback#section",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// T049: Test case 11 - Malformed URL
		{
			name:          "malformed URL",
			redirectURI:   "ht!tp://invalid",
			requestHost:   "localhost:8000",
			expectedValid: false,
			expectError:   true,
		},
		// T053: Relative paths with dots
		{
			name:          "relative URL with dot notation",
			redirectURI:   "../../callback",
			requestHost:   "localhost:8000",
			expectedValid: true,
			expectError:   false,
		},
		// Port normalization - default http port
		{
			name:          "http default port normalization",
			redirectURI:   "http://localhost:80/callback",
			requestHost:   "localhost",
			expectedValid: true,
			expectError:   false,
		},
		// Port normalization - default https port
		{
			name:          "https default port normalization",
			redirectURI:   "https://example.com:443/callback",
			requestHost:   "example.com:443", // Explicitly https
			expectedValid: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create *http.Request from requestHost string
			// Detect if HTTPS by checking for :443 port indicator
			isTLS := strings.Contains(tt.requestHost, ":443")

			req := httptest.NewRequest("GET", "http://"+tt.requestHost+"/", nil)
			req.Host = tt.requestHost
			if isTLS {
				req.TLS = &tls.ConnectionState{}
			}

			valid, err := validateRedirectURI(tt.redirectURI, req)

			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if valid != tt.expectedValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectedValid, valid)
			}
		})
	}
}

// TestCreateGrant_WithRedirectURI_Valid tests approval with valid redirect_uri
func TestCreateGrant_WithRedirectURI_Valid(t *testing.T) {
	t.Parallel()
	// T050: Test case 2 - Approval with valid redirect_uri should redirect
	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			return &storage.UserGrant{
				ID:         testGrantID,
				Principal:  id.Principal("user@example.com"),
				AgentID:    testAgentID,
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID(),
						Scopes:                    []string{"repo"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants?redirect_uri=%2Fcallback&code=xyz", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Should return 201 Created with redirect_url in response body (not HTTP redirect)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}

	// Check response body contains redirect_url
	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	redirectUrl, ok := response["redirect_url"].(string)
	if !ok {
		t.Error("expected redirect_url in response body")
	}
	if redirectUrl != "/callback" {
		t.Errorf("expected redirect_url '/callback', got '%s'", redirectUrl)
	}
}

// TestCreateGrant_WithRedirectURI_RelativeValid tests approval with relative redirect_uri
func TestCreateGrant_WithRedirectURI_RelativeValid(t *testing.T) {
	t.Parallel()
	// T050: Test case 3 - Approval with relative redirect_uri should redirect
	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			return &storage.UserGrant{
				ID:         testGrantID,
				Principal:  id.Principal("user@example.com"),
				AgentID:    testAgentID,
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID(),
						Scopes:                    []string{"repo"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants?redirect_uri=/auth/return", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Should return 201 Created with redirect_url in response body (not HTTP redirect)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}

	// Check response body contains redirect_url
	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	redirectUrl, ok := response["redirect_url"].(string)
	if !ok {
		t.Error("expected redirect_url in response body")
	}
	if redirectUrl != "/auth/return" {
		t.Errorf("expected redirect_url '/auth/return', got '%s'", redirectUrl)
	}
}

// TestCreateGrant_WithRedirectURI_InvalidDomain tests approval with external domain redirect_uri
func TestCreateGrant_WithRedirectURI_InvalidDomain(t *testing.T) {
	t.Parallel()
	// T050: Test case 4 - Approval with invalid redirect_uri (external domain) should return error
	testAgentID := id.NewAgentID()
	handler := NewGrantsHandler(nil, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants?redirect_uri=https://evil.com/callback", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Should return HTTP 400
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Error("expected error response")
	}
}

// TestCreateGrant_WithoutRedirectURI tests approval without redirect_uri (success page)
func TestCreateGrant_WithoutRedirectURI(t *testing.T) {
	t.Parallel()
	// T050: Test case 1 - Approval without redirect_uri should return success page
	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			return &storage.UserGrant{
				ID:         testGrantID,
				Principal:  id.Principal("user@example.com"),
				AgentID:    testAgentID,
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID(),
						Scopes:                    []string{"repo"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Should return 201 Created (not a redirect)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	var envelope map[string]GrantResponse
	if err := json.NewDecoder(rr.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := envelope["data"]; !ok {
		t.Error("expected 'data' field in success response")
	}
}

// TestCreateGrant_WithRedirectURI_PreservesQueryParams tests that query parameters are preserved
func TestCreateGrant_WithRedirectURI_PreservesQueryParams(t *testing.T) {
	t.Parallel()
	// T058: Test case - Approval preserves query parameters in redirect
	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			return &storage.UserGrant{
				ID:         testGrantID,
				Principal:  id.Principal("user@example.com"),
				AgentID:    testAgentID,
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: id.NewServiceID(),
						Scopes:                    []string{"repo"},
					},
				},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: id.NewServiceID().String(),
				Scopes:                    []string{"repo"},
			},
		},
	}

	// redirect_uri already has query params, and we have additional OAuth params
	req := newRequestWithPrincipal("POST", "/api/consent/agent/"+testAgentID.String()+"/grants?redirect_uri=/callback%3Fsession%3Dabc&state=xyz", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Should return 201 Created with redirect_url in response body (not HTTP redirect)
	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", rr.Code)
	}

	// Check response body contains redirect_url
	var response map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	redirectUrl, ok := response["redirect_url"].(string)
	if !ok {
		t.Error("expected redirect_url in response body")
	}
	// Verify that query parameters are preserved in the redirect URL
	if redirectUrl != "/callback?session=abc" {
		t.Errorf("expected redirect_url '/callback?session=abc', got '%s'", redirectUrl)
	}
}

// TestCreateGrant_SessionToken_ValidFlow verifies that a valid JWE session token
// produces a 201 with redirect_url from the token claims.
func TestCreateGrant_SessionToken_ValidFlow(t *testing.T) {
	t.Parallel()

	testAgentID := id.NewAgentID()
	testGrantID := id.NewGrantID()
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)
	principalVal := "user@example.com"
	originalURL := "https://agent.example.com/authorize"

	ts := newTestJWETokenService()
	sessionToken := newTestSessionToken(ts, testAgentID, principalVal, originalURL)

	mockService := &mockConsentService{
		grantConsentFunc: func(_ context.Context, _ *consent.GrantRequest) (*storage.UserGrant, error) {
			return &storage.UserGrant{
				ID:                    testGrantID,
				Principal:             id.Principal(principalVal),
				AgentID:               testAgentID,
				ValidUntil:            &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{},
				CreatedAt:             now,
				UpdatedAt:             now,
			}, nil
		},
	}

	handler := NewGrantsHandler(mockService, nil).WithJWETokenService(ts)

	req := newRequestWithPrincipal(
		"POST",
		"/api/consent/agent/"+testAgentID.String()+"/grants?session_token="+sessionToken,
		principalVal,
		GrantRequest{DelegatedOAuth2Tokens: []DelegatedTokenRequest{}},
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.CreateGrant(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code, rr.Body.String())
	var resp map[string]any
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(t, originalURL, resp["redirect_url"])
}

// TestCreateGrant_SessionToken_InvalidToken verifies that a tampered or invalid
// JWE session token produces 400.
func TestCreateGrant_SessionToken_InvalidToken(t *testing.T) {
	t.Parallel()

	testAgentID := id.NewAgentID()
	principalVal := "user@example.com"

	ts := newTestJWETokenService()
	handler := NewGrantsHandler(&mockConsentService{}, nil).WithJWETokenService(ts)

	req := newRequestWithPrincipal(
		"POST",
		"/api/consent/agent/"+testAgentID.String()+"/grants?session_token=notavalidjwetoken",
		principalVal,
		GrantRequest{DelegatedOAuth2Tokens: []DelegatedTokenRequest{}},
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.CreateGrant(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
}

// TestCreateGrant_SessionToken_AgentMismatch verifies that a session token issued for
// a different agent produces 400.
func TestCreateGrant_SessionToken_AgentMismatch(t *testing.T) {
	t.Parallel()

	agentA := id.NewAgentID()
	agentB := id.NewAgentID()
	principalVal := "user@example.com"

	ts := newTestJWETokenService()
	tokenForAgentA := newTestSessionToken(ts, agentA, principalVal, "/callback")

	handler := NewGrantsHandler(&mockConsentService{}, nil).WithJWETokenService(ts)

	req := newRequestWithPrincipal(
		"POST",
		"/api/consent/agent/"+agentB.String()+"/grants?session_token="+tokenForAgentA,
		principalVal,
		GrantRequest{DelegatedOAuth2Tokens: []DelegatedTokenRequest{}},
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentB.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.CreateGrant(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
}

// TestCreateGrant_SessionToken_PrincipalMismatch verifies that a session token issued
// for a different user produces 403.
func TestCreateGrant_SessionToken_PrincipalMismatch(t *testing.T) {
	t.Parallel()

	testAgentID := id.NewAgentID()

	ts := newTestJWETokenService()
	tokenForUserA := newTestSessionToken(ts, testAgentID, "userA@example.com", "/callback")

	handler := NewGrantsHandler(&mockConsentService{}, nil).WithJWETokenService(ts)

	req := newRequestWithPrincipal(
		"POST",
		"/api/consent/agent/"+testAgentID.String()+"/grants?session_token="+tokenForUserA,
		"userB@example.com",
		GrantRequest{DelegatedOAuth2Tokens: []DelegatedTokenRequest{}},
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.CreateGrant(rr, req)

	require.Equal(t, http.StatusForbidden, rr.Code, rr.Body.String())
}
