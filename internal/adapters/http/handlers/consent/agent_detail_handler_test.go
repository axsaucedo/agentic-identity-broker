package consent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// mockAgentDetailService is a mock implementation of consent.Service for testing.
type mockAgentDetailService struct {
	getAgentDetailFunc func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
}

func (m *mockAgentDetailService) GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
	if m.getAgentDetailFunc != nil {
		return m.getAgentDetailFunc(ctx, agentID)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentConsentInfo(ctx context.Context, agentID string) (*consent.AgentConsentInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) RevokeConsent(ctx context.Context, principal, agentID string) error {
	return errors.New("not implemented")
}

func (m *mockAgentDetailService) GetActiveGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetAgentDelegations(ctx context.Context, principal string) ([]consent.AgentDelegation, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentDetailService) GetUserGrants(ctx context.Context, principal, agentID string) ([]*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func TestGetAgentDetail_Success(t *testing.T) {
	// Setup
	agentID := "agent-123"
	governanceURL := "https://example.com/governance"
	userDocsURL := "https://example.com/docs"
	agentInterfaceURL := "https://example.com/interface"

	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			if agID != agentID {
				t.Errorf("expected agentID %s, got %s", agentID, agID)
			}

			agentDetail := &consent.AgentDetail{
				AgentID:              agentID,
				DisplayName:          "Test Agent",
				Description:          "A test agent for testing purposes",
				LogoURL:              nil,
				GovernanceURL:        &governanceURL,
				UserDocumentationURL: &userDocsURL,
				AgentInterfaceURL:    &agentInterfaceURL,
			}

			services := []consent.ThirdpartyService{
				{
					ServiceID:   "github",
					DisplayName: "GitHub",
					LogoURL:     nil,
					Scopes: []consent.ServiceScope{
						{Value: "read:user", Description: "Read user profile"},
						{Value: "repo", Description: "Full control of repositories"},
					},
				},
				{
					ServiceID:   "google",
					DisplayName: "Google",
					LogoURL:     nil,
					Scopes: []consent.ServiceScope{
						{Value: "email", Description: "View email address"},
					},
				},
			}

			return agentDetail, services, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agentId", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify agent details
	if response.Data.Agent.AgentID != agentID {
		t.Errorf("expected agent ID %s, got %s", agentID, response.Data.Agent.AgentID)
	}
	if response.Data.Agent.DisplayName != "Test Agent" {
		t.Errorf("expected display name 'Test Agent', got %s", response.Data.Agent.DisplayName)
	}
	if response.Data.Agent.GovernanceURL == nil || *response.Data.Agent.GovernanceURL != governanceURL {
		t.Errorf("governance URL mismatch")
	}

	// Verify services
	if len(response.Data.Services) != 2 {
		t.Errorf("expected 2 services, got %d", len(response.Data.Services))
	}
	if response.Data.Services[0].ServiceID != "github" {
		t.Errorf("expected first service to be 'github', got %s", response.Data.Services[0].ServiceID)
	}
	if len(response.Data.Services[0].Scopes) != 2 {
		t.Errorf("expected 2 scopes for GitHub, got %d", len(response.Data.Services[0].Scopes))
	}
}

func TestGetAgentDetail_AgentNotFound(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			return nil, nil, consent.ErrAgentNotFound
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agentId", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "not found" {
		t.Errorf("expected error 'not found', got %s", response.Error)
	}
}

func TestGetAgentDetail_MissingAgentID(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{}
	handler := NewAgentDetailHandler(mockService, nil)

	// Create request without agent ID
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/", nil)
	rctx := chi.NewRouteContext()
	// Don't add agentId parameter
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "bad request" {
		t.Errorf("expected error 'bad request', got %s", response.Error)
	}
}

func TestGetAgentDetail_ServiceError(t *testing.T) {
	// Setup
	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			return nil, nil, errors.New("database connection failed")
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agentId", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "internal server error" {
		t.Errorf("expected error 'internal server error', got %s", response.Error)
	}
}

func TestGetAgentDetail_EmptyServicesList(t *testing.T) {
	// Setup
	agentID := "agent-456"

	mockService := &mockAgentDetailService{
		getAgentDetailFunc: func(ctx context.Context, agID string) (*consent.AgentDetail, []consent.ThirdpartyService, error) {
			agentDetail := &consent.AgentDetail{
				AgentID:     agentID,
				DisplayName: "Test Agent",
				Description: "A test agent",
			}

			// Return empty services list
			services := []consent.ThirdpartyService{}

			return agentDetail, services, nil
		},
	}

	handler := NewAgentDetailHandler(mockService, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agentId", agentID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentDetail(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response GetAgentDetailResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify empty services list
	if len(response.Data.Services) != 0 {
		t.Errorf("expected 0 services, got %d", len(response.Data.Services))
	}
}

// Ensure mockAgentDetailService implements the required interface methods
var _ interface {
	GetAgentDetail(ctx context.Context, agentID string) (*consent.AgentDetail, []consent.ThirdpartyService, error)
} = (*mockAgentDetailService)(nil)
