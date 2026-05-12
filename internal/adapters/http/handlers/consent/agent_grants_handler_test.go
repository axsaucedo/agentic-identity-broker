package consent

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/go-chi/chi/v5"
)

// mockAgentGrantsService is a mock implementation of consent.Service for testing.
type mockAgentGrantsService struct {
	getUserGrantsFunc func(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error)
}

func (m *mockAgentGrantsService) GetUserGrants(ctx context.Context, principal id.Principal, agentID id.AgentID) ([]*storage.UserGrant, error) {
	if m.getUserGrantsFunc != nil {
		return m.getUserGrantsFunc(ctx, principal, agentID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAgentGrantsService) GetAgentConsentInfo(ctx context.Context, agentID id.AgentID) (*consent.AgentConsentInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentGrantsService) GrantConsent(ctx context.Context, req *consent.GrantRequest) (*storage.UserGrant, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentGrantsService) RevokeConsent(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	return errors.New("not implemented")
}

func (m *mockAgentGrantsService) GetAgentDelegations(ctx context.Context, principal id.Principal) ([]consent.AgentDelegation, error) {
	return nil, errors.New("not implemented")
}

func (m *mockAgentGrantsService) RevokeConsentForPrincipal(ctx context.Context, principal id.Principal, agentID id.AgentID) error {
	return nil
}

func (m *mockAgentGrantsService) GetAgentWithServiceRequirements(ctx context.Context, userPrincipal id.Principal, agentID id.AgentID) (*storage.Agent, []consent.ServiceRequirementStatus, error) {
	return nil, nil, errors.New("not implemented")
}

func TestGetAgentGrants_Success(t *testing.T) {
	t.Parallel()
	// Setup
	principalValue := "user@example.com"
	agentID := id.NewAgentID()
	grantID := id.NewGrantID()
	serviceID := id.NewServiceID()
	validUntil := time.Now().Add(24 * time.Hour)

	mockService := &mockAgentGrantsService{
		getUserGrantsFunc: func(ctx context.Context, p id.Principal, agID id.AgentID) ([]*storage.UserGrant, error) {
			if p != id.Principal(principalValue) {
				t.Errorf("expected principal %s, got %s", principalValue, p)
			}
			if agID != agentID {
				t.Errorf("expected agentID %s, got %s", agentID, agID)
			}

			return []*storage.UserGrant{
				{
					ID:         grantID,
					Principal:  id.Principal(principalValue),
					AgentID:    agentID,
					ValidUntil: &validUntil,
					DelegatedOAuth2Tokens: []storage.DelegatedToken{
						{
							ThirdpartyOAuth2ServiceID: serviceID,
							Scopes:                    []string{"read:user", "repo"},
						},
					},
					CreatedAt: time.Now().Add(-48 * time.Hour),
					UpdatedAt: time.Now().Add(-1 * time.Hour),
				},
			}, nil
		},
	}

	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String()+"/grants", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response GetAgentGrantsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify grant exists (due to 1:1 relationship, only one grant is returned)
	if response.Data == nil {
		t.Error("expected grant to be returned, got nil")
	}

	// Verify grant details
	if response.Data.ID != grantID.String() {
		t.Errorf("expected grant ID '%s', got %s", grantID.String(), response.Data.ID)
	}
	if response.Data.Principal != principalValue {
		t.Errorf("expected principal %s, got %s", principalValue, response.Data.Principal)
	}
	if response.Data.ValidUntil == nil {
		t.Error("expected ValidUntil to be set")
	}
	if len(response.Data.DelegatedOAuth2Tokens) != 1 {
		t.Errorf("expected 1 delegated token, got %d", len(response.Data.DelegatedOAuth2Tokens))
	}
}

func TestGetAgentGrants_EmptyGrants(t *testing.T) {
	t.Parallel()
	// Setup - user hasn't granted this agent access yet
	principalValue := "user@example.com"
	agentID := id.NewAgentID()

	mockService := &mockAgentGrantsService{
		getUserGrantsFunc: func(ctx context.Context, p id.Principal, agID id.AgentID) ([]*storage.UserGrant, error) {
			return []*storage.UserGrant{}, nil // Empty grants
		},
	}

	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String()+"/grants", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

	// Assert
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var response GetAgentGrantsResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify no grant exists (Data should be nil when user hasn't granted access)
	if response.Data != nil {
		t.Errorf("expected Data to be nil for empty grants, got %v", response.Data)
	}
}

func TestGetAgentGrants_MissingPrincipal(t *testing.T) {
	t.Parallel()
	// Setup
	mockService := &mockAgentGrantsService{}
	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request without principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/agent-123/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", "agent-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

	// Assert
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if response.Error != "unauthorized" {
		t.Errorf("expected error 'unauthorized', got %s", response.Error)
	}
}

func TestGetAgentGrants_MissingAgentID(t *testing.T) {
	t.Parallel()
	// Setup
	mockService := &mockAgentGrantsService{}
	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request without agent ID
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent//grants", nil)
	ctx := principal.WithPrincipal(req.Context(), "user@example.com")
	rctx := chi.NewRouteContext()
	// Don't add agentId parameter
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

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

func TestGetAgentGrants_AgentNotFound(t *testing.T) {
	t.Parallel()
	// Setup
	principalValue := "user@example.com"
	agentID := id.NewAgentID()

	mockService := &mockAgentGrantsService{
		getUserGrantsFunc: func(ctx context.Context, p id.Principal, agID id.AgentID) ([]*storage.UserGrant, error) {
			return nil, consent.ErrAgentNotFound
		},
	}

	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String()+"/grants", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

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

func TestGetAgentGrants_ServiceError(t *testing.T) {
	t.Parallel()
	// Setup
	principalValue := "user@example.com"
	agentID := id.NewAgentID()

	mockService := &mockAgentGrantsService{
		getUserGrantsFunc: func(ctx context.Context, p id.Principal, agID id.AgentID) ([]*storage.UserGrant, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewAgentGrantsHandler(mockService, nil)

	// Create request with principal in context
	req := httptest.NewRequest(http.MethodGet, "/api/consent/agent/"+agentID.String()+"/grants", nil)
	ctx := principal.WithPrincipal(req.Context(), principalValue)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("agent-id", agentID.String())
	ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute
	handler.GetAgentGrants(rr, req)

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

func TestToUserGrantDTO_Conversion(t *testing.T) {
	t.Parallel()
	// Setup
	grantID := id.NewGrantID()
	agentID := id.NewAgentID()
	serviceID := id.NewServiceID()
	validUntil := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	grant := &storage.UserGrant{
		ID:         grantID,
		Principal:  id.Principal("user@example.com"),
		AgentID:    agentID,
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: serviceID,
				Scopes:                    []string{"read:user", "repo"},
			},
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	mockService := &mockAgentGrantsService{}
	handler := NewAgentGrantsHandler(mockService, nil)

	// Execute
	dto := handler.toUserGrantDTO(grant)

	// Assert
	if dto.ID != grantID.String() {
		t.Errorf("expected ID '%s', got %s", grantID.String(), dto.ID)
	}
	if dto.Principal != "user@example.com" {
		t.Errorf("expected principal 'user@example.com', got %s", dto.Principal)
	}
	if dto.AgentID != agentID.String() {
		t.Errorf("expected agentID '%s', got %s", agentID.String(), dto.AgentID)
	}
	if dto.ValidUntil == nil {
		t.Fatal("expected ValidUntil to be set")
	}
	if !dto.ValidUntil.Equal(validUntil) {
		t.Errorf("ValidUntil mismatch: expected %v, got %v", validUntil, dto.ValidUntil)
	}
	if len(dto.DelegatedOAuth2Tokens) != 1 {
		t.Errorf("expected 1 delegated token, got %d", len(dto.DelegatedOAuth2Tokens))
	}
	if dto.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID != serviceID.String() {
		t.Errorf("expected service ID '%s', got %s", serviceID.String(), dto.DelegatedOAuth2Tokens[0].ThirdpartyOAuth2ServiceID)
	}
	if len(dto.DelegatedOAuth2Tokens[0].Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(dto.DelegatedOAuth2Tokens[0].Scopes))
	}

	// Verify RFC3339 timestamp format
	expectedCreatedAt := "2025-01-01T00:00:00Z"
	if dto.CreatedAt != expectedCreatedAt {
		t.Errorf("expected createdAt %s, got %s", expectedCreatedAt, dto.CreatedAt)
	}
	expectedUpdatedAt := "2025-06-01T00:00:00Z"
	if dto.UpdatedAt != expectedUpdatedAt {
		t.Errorf("expected updatedAt %s, got %s", expectedUpdatedAt, dto.UpdatedAt)
	}
}

// Ensure mockAgentGrantsService implements the ConsentService interface
var _ ConsentService = (*mockAgentGrantsService)(nil)
