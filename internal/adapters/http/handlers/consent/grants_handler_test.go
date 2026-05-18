package consent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
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
	claims, err := domotp2.NewAuthorizationSessionClaims(agentID, id.Principal(principalVal), originalURL, nil)
	if err != nil {
		panic("newTestSessionToken: invalid claims: " + err.Error())
	}
	token, err := ts.Encrypt(claims)
	if err != nil {
		panic("newTestSessionToken: failed to encrypt claims: " + err.Error())
	}
	return token
}

// newExpiredTestSessionToken creates a JWE session token whose TTL has already elapsed.
func newExpiredTestSessionToken(ts *domjwe.TokenService, agentID id.AgentID, principalVal string) string {
	past := time.Now().Add(-time.Hour)
	claims := &domotp2.AuthorizationSessionClaims{
		AgentID:   agentID,
		Principal: id.Principal(principalVal),
		IssuedAt:  past,
		ExpiresAt: past,
	}
	token, err := ts.Encrypt(claims)
	if err != nil {
		panic("newExpiredTestSessionToken: failed to encrypt claims: " + err.Error())
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

func TestCreateGrant_NoPrincipal(t *testing.T) {
	t.Parallel()
	handler := NewGrantsHandler(nil, nil, newTestJWETokenService())
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
	handler := NewGrantsHandler(nil, nil, newTestJWETokenService())
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
	handler := NewGrantsHandler(nil, nil, newTestJWETokenService())
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
	handler := NewGrantsHandler(nil, nil, newTestJWETokenService())

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
			handler := NewGrantsHandler(mockService, nil, newTestJWETokenService())

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

	handler := NewGrantsHandler(mockService, nil, newTestJWETokenService())

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

	handler := NewGrantsHandler(mockService, nil, newTestJWETokenService())

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

	handler := NewGrantsHandler(mockService, nil, ts)

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
	handler := NewGrantsHandler(&mockConsentService{}, nil, ts)

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

	handler := NewGrantsHandler(&mockConsentService{}, nil, ts)

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

// TestCreateGrant_RedirectURIWithoutSessionToken verifies that providing redirect_uri
// without session_token is rejected with 400 — no insecure fallback (US3-S2 / T015).
func TestCreateGrant_RedirectURIWithoutSessionToken(t *testing.T) {
	t.Parallel()

	testAgentID := id.NewAgentID()
	handler := NewGrantsHandler(&mockConsentService{}, nil, newTestJWETokenService())

	req := newRequestWithPrincipal(
		"POST",
		"/api/consent/agent/"+testAgentID.String()+"/grants?redirect_uri=%2Fcallback",
		"user@example.com",
		GrantRequest{DelegatedOAuth2Tokens: []DelegatedTokenRequest{}},
	)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.CreateGrant(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code, rr.Body.String())
	var resp ErrorResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.Equal(t, "bad request", resp.Error)
}

// TestCreateGrant_SessionToken_PrincipalMismatch verifies that a session token issued
// for a different user produces 403.
func TestCreateGrant_SessionToken_PrincipalMismatch(t *testing.T) {
	t.Parallel()

	testAgentID := id.NewAgentID()

	ts := newTestJWETokenService()
	tokenForUserA := newTestSessionToken(ts, testAgentID, "userA@example.com", "/callback")

	handler := NewGrantsHandler(&mockConsentService{}, nil, ts)

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
