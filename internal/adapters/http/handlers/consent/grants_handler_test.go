package consent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// Helper to create request with principal context
func newRequestWithPrincipal(method, path, principalValue string, body interface{}) *http.Request {
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
	handler := NewGrantsHandler(nil, nil)

	// Create request without principal
	req := httptest.NewRequest("GET", "/api/consent/agent/agent-123/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
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
	handler := NewGrantsHandler(nil, nil)

	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"repo"},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/consent/agent/agent-123/grants", bytes.NewBuffer(jsonBody))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify response
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestCreateGrant_InvalidJSON(t *testing.T) {
	handler := NewGrantsHandler(nil, nil)

	req := newRequestWithPrincipal("POST", "/api/consent/agent/agent-123/grants", "user@example.com", nil)
	req.Body = httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("invalid json"))).Body

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
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

func TestCreateGrant_EmptyTokensRevokes(t *testing.T) {
	revokeCalled := false
	mockService := &mockConsentService{
		revokeConsentFunc: func(ctx context.Context, principal, agentID string) error {
			revokeCalled = true
			if principal != "user@example.com" {
				t.Errorf("expected principal 'user@example.com', got '%s'", principal)
			}
			if agentID != "agent-123" {
				t.Errorf("expected agentID 'agent-123', got '%s'", agentID)
			}
			return nil
		},
	}
	handler := NewGrantsHandler(mockService, nil)

	// Create request with empty tokens (revoke)
	reqBody := GrantRequest{
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/agent-123/grants", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify RevokeConsent was called
	if !revokeCalled {
		t.Error("expected RevokeConsent to be called")
	}

	// Verify response
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestCreateGrant_ValidUntilInPast(t *testing.T) {
	handler := NewGrantsHandler(nil, nil)

	pastTime := time.Now().Add(-1 * time.Hour)
	reqBody := GrantRequest{
		ValidUntil: &pastTime,
		DelegatedOAuth2Tokens: []DelegatedTokenRequest{
			{
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"repo"},
			},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/consent/agent/agent-123/grants", bytes.NewBuffer(jsonBody))
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
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
	handler := NewGrantsHandler(nil, nil)

	validUntil := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	grant := &storage.UserGrant{
		ID:         "grant-123",
		Principal:  "user@example.com",
		AgentID:    "agent-456",
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"repo", "user"},
			},
			{
				ThirdpartyOAuth2ServiceID: "google",
				Scopes:                    []string{"email"},
			},
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	response := handler.toGrantResponse(grant)

	// Verify basic fields
	if response.ID != "grant-123" {
		t.Errorf("expected ID 'grant-123', got '%s'", response.ID)
	}
	if response.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got '%s'", response.Principal)
	}
	if response.AgentID != "agent-456" {
		t.Errorf("expected agentID 'agent-456', got '%s'", response.AgentID)
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
	if token1.ThirdpartyOAuth2ServiceID != "github" {
		t.Errorf("expected service 'github', got '%s'", token1.ThirdpartyOAuth2ServiceID)
	}
	if len(token1.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(token1.Scopes))
	}

	token2 := response.DelegatedOAuth2Tokens[1]
	if token2.ThirdpartyOAuth2ServiceID != "google" {
		t.Errorf("expected service 'google', got '%s'", token2.ThirdpartyOAuth2ServiceID)
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
			name:           "generic error",
			serviceError:   errors.New("database error"),
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
						ThirdpartyOAuth2ServiceID: "github",
						Scopes:                    []string{"repo"},
					},
				},
			}

			req := newRequestWithPrincipal("POST", "/api/consent/agent/agent-123/grants", "user@example.com", reqBody)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("agent-id", "agent-123")
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
	// Create mock service
	now := time.Now()
	futureTime := now.Add(24 * time.Hour)

	var capturedRequest *consent.GrantRequest
	mockService := &mockConsentService{
		grantConsentFunc: func(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
			capturedRequest = req
			return &storage.UserGrant{
				ID:         "grant-new-123",
				Principal:  "user@example.com",
				AgentID:    "agent-123",
				ValidUntil: &futureTime,
				DelegatedOAuth2Tokens: []storage.DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "github",
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
				ThirdpartyOAuth2ServiceID: "github",
				Scopes:                    []string{"repo", "user"},
			},
		},
	}

	req := newRequestWithPrincipal("POST", "/api/consent/agent/agent-123/grants", "user@example.com", reqBody)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()

	handler.CreateGrant(rr, req)

	// Verify service was called and request was captured
	if capturedRequest == nil {
		t.Fatal("expected grant request to be passed to service")
	}
	if capturedRequest.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got '%s'", capturedRequest.Principal)
	}
	if capturedRequest.AgentID != "agent-123" {
		t.Errorf("expected agent_id 'agent-123', got '%s'", capturedRequest.AgentID)
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

	if response.ID != "grant-new-123" {
		t.Errorf("expected ID 'grant-new-123', got '%s'", response.ID)
	}
	if response.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got '%s'", response.Principal)
	}
	if response.AgentID != "agent-123" {
		t.Errorf("expected agent_id 'agent-123', got '%s'", response.AgentID)
	}
}

// TestGetGrants_Success verifies successful grant retrieval
func TestGetGrants_Success(t *testing.T) {
	now := time.Now()

	mockService := &mockConsentService{
		getActiveGrantsFunc: func(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
			return []*storage.UserGrant{
				{
					ID:        "grant-1",
					Principal: "user@example.com",
					AgentID:   "agent-123",
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: "github",
							Scopes:                    []string{"repo"},
						},
					},
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "grant-2",
					Principal: "user@example.com",
					AgentID:   "agent-123",
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: "google",
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

	req := newRequestWithPrincipal("GET", "/api/consent/agent/agent-123/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
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
	mockService := &mockConsentService{
		getActiveGrantsFunc: func(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
			return nil, consent.ErrAgentNotFound
		},
	}
	handler := NewGrantsHandler(mockService, nil)

	req := newRequestWithPrincipal("GET", "/api/consent/agent/unknown-agent/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "unknown-agent")
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
