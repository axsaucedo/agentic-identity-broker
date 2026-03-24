package consent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/go-chi/chi/v5"
)

// TestRevokeGrantHandler_Success tests 204 No Content on successful revocation.
func TestRevokeGrantHandler_Success(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()
	revokeCalled := false

	mockService := &mockConsentService{
		revokeConsentForPrincipalFunc: func(ctx context.Context, p id.Principal, agentID id.AgentID) error {
			revokeCalled = true
			if p != id.Principal("user@example.com") {
				t.Errorf("expected principal 'user@example.com', got '%s'", p)
			}
			if agentID != testAgentID {
				t.Errorf("expected agentID '%s', got '%s'", testAgentID, agentID)
			}
			return nil
		},
	}
	handler := NewRevokeGrantHandler(mockService, nil)

	req := newRequestWithPrincipal("DELETE", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	if !revokeCalled {
		t.Error("expected RevokeConsentForPrincipal to be called")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

// TestRevokeGrantHandler_NoPrincipal tests 401 when principal is missing.
// Principal check must happen BEFORE UUID parse (AGENTS.md rule: 401 before 400).
func TestRevokeGrantHandler_NoPrincipal(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()
	handler := NewRevokeGrantHandler(nil, nil)

	req := httptest.NewRequest("DELETE", "/api/consent/agent/"+testAgentID.String()+"/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

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

// TestRevokeGrantHandler_NoPrincipal_InvalidUUID tests that 401 is returned even when
// the agent-id is also invalid — principal check is first (SR-001).
func TestRevokeGrantHandler_NoPrincipal_InvalidUUID(t *testing.T) {
	t.Parallel()
	handler := NewRevokeGrantHandler(nil, nil)

	req := httptest.NewRequest("DELETE", "/api/consent/agent/not-a-uuid/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	// Must be 401, NOT 400 — principal check is before UUID parse
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d (principal checked first), got %d", http.StatusUnauthorized, rr.Code)
	}
}

// TestRevokeGrantHandler_InvalidAgentID tests 400 when agent-id is not a valid UUID.
func TestRevokeGrantHandler_InvalidAgentID(t *testing.T) {
	t.Parallel()
	handler := NewRevokeGrantHandler(nil, nil)

	req := newRequestWithPrincipal("DELETE", "/api/consent/agent/not-a-uuid/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "not-a-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

// TestRevokeGrantHandler_NotFound tests 404 when ErrGrantNotFound is returned (SR-001).
func TestRevokeGrantHandler_NotFound(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()

	mockService := &mockConsentService{
		revokeConsentForPrincipalFunc: func(ctx context.Context, p id.Principal, agentID id.AgentID) error {
			return fmt.Errorf("wrapped: %w", consent.ErrGrantNotFound)
		},
	}
	handler := NewRevokeGrantHandler(mockService, nil)

	req := newRequestWithPrincipal("DELETE", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	var errResp ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp.Error != "not found" {
		t.Errorf("expected error 'not found', got '%s'", errResp.Error)
	}
}

// TestRevokeGrantHandler_ServiceError tests 500 on unexpected service errors (fail-closed SR-003).
func TestRevokeGrantHandler_ServiceError(t *testing.T) {
	t.Parallel()
	testAgentID := id.NewAgentID()

	mockService := &mockConsentService{
		revokeConsentForPrincipalFunc: func(ctx context.Context, p id.Principal, agentID id.AgentID) error {
			return errors.New("database connection lost")
		},
	}
	handler := NewRevokeGrantHandler(mockService, nil)

	req := newRequestWithPrincipal("DELETE", "/api/consent/agent/"+testAgentID.String()+"/grants", "user@example.com", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", testAgentID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// TestRevokeGrantHandler_NoPrincipalWithInvalidAgentID_PrincipalCheckedFirst tests
// that principal validation error takes precedence over agent ID validation error.
func TestRevokeGrantHandler_PrincipalCheckedFirst(t *testing.T) {
	t.Parallel()
	handler := NewRevokeGrantHandler(nil, nil)

	// No principal in context AND invalid UUID
	req := httptest.NewRequest("DELETE", "/api/consent/agent/bad-uuid/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "bad-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.RevokeGrant(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized (principal checked before UUID), got %d", rr.Code)
	}
}
