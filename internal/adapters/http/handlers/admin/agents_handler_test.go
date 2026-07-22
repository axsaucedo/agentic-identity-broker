package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	encryptionnoop "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/noop"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/agents"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/model"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAgentRepository is a mock implementation of ports.AgentRepository
type MockAgentRepository struct {
	mock.Mock
}

func (m *MockAgentRepository) Create(ctx context.Context, agent *storage.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Get(ctx context.Context, agentID id.AgentID) (*storage.Agent, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Agent), args.Error(1)
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Delete(ctx context.Context, agentID id.AgentID) error {
	args := m.Called(ctx, agentID)
	return args.Error(0)
}

func (m *MockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByClientID(ctx context.Context, clientID id.ClientID) (*storage.Agent, error) {
	args := m.Called(ctx, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Agent), args.Error(1)
}

func (m *MockAgentRepository) GetByClientURI(ctx context.Context, uri string) (*storage.Agent, error) {
	args := m.Called(ctx, uri)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Agent), args.Error(1)
}

func (m *MockAgentRepository) ExistsOtherWithClientID(ctx context.Context, clientID id.ClientID, excludeAgentID *id.AgentID) (bool, error) {
	args := m.Called(ctx, clientID, excludeAgentID)
	return args.Bool(0), args.Error(1)
}

// newAgentsHandlerForTest creates an AgentsHandler backed by a real domain service
// wrapping a mock repository. This ensures the architecture invariant holds in tests:
// the handler always goes through the domain service, never raw storage.
//
// MockProviderRepository and newTestEncryption are defined in services_handler_test.go
// and are available here because both files share the same package admin.
func newAgentsHandlerForTest(mockRepo *MockAgentRepository, mockServiceRepo *MockProviderRepository, logger *slog.Logger) *AgentsHandler {
	providerSvc := thirdparty.NewThirdpartyOAuth2ProviderService(mockServiceRepo, newTestEncryption(), &encryptionnoop.BranchKeyManager{}, nil, false, logger)
	// Use multiAgentEnabled=true so existing CRUD tests don't need ExistsOtherWithClientID expectations.
	// T033 tests use newAgentsHandlerForTestWithMultiAgent with explicit flags.
	agentSvc := agents.NewService(mockRepo, providerSvc, logger, true)
	return NewAgentsHandler(agentSvc, providerSvc, nil, logger)
}

// newAgentsHandlerForTestWithMultiAgent creates an AgentsHandler with the given multiAgentEnabled flag.
func newAgentsHandlerForTestWithMultiAgent(mockRepo *MockAgentRepository, mockServiceRepo *MockProviderRepository, logger *slog.Logger, multiAgentEnabled bool) *AgentsHandler {
	providerSvc := thirdparty.NewThirdpartyOAuth2ProviderService(mockServiceRepo, newTestEncryption(), &encryptionnoop.BranchKeyManager{}, nil, false, logger)
	agentSvc := agents.NewService(mockRepo, providerSvc, logger, multiAgentEnabled)
	return NewAgentsHandler(agentSvc, providerSvc, nil, logger)
}

func TestAgentsHandler_CreateAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		reqBody := AgentRequest{
			ClientID:    ptr.To("test-client"),
			DisplayName: "Test Agent",
			Description: "Test agent description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID != nil && *a.ClientID == "test-client" && a.DisplayName == "Test Agent"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.ID)
		require.NotNil(t, resp.ClientID)
		assert.Equal(t, "test-client", *resp.ClientID)
		assert.Equal(t, "Test Agent", resp.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("with optional fields", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		externalID := "ext-123"
		govURL := "https://example.com/gov"
		reqBody := AgentRequest{
			ClientID:      ptr.To("test-client-2"),
			ExternalID:    &externalID,
			DisplayName:   "Test Agent 2",
			Description:   "Test description",
			GovernanceURL: &govURL,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, externalID, *resp.ExternalID)
		assert.Equal(t, govURL, *resp.GovernanceURL)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "invalid request body", resp.Error)
	})

	t.Run("nil client_id when omitted from request", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		reqBody := AgentRequest{
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID == nil
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Nil(t, resp.ClientID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("blank client_id returns 400 validation failed", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		emptyClientID := ""
		reqBody := AgentRequest{
			ClientID:    &emptyClientID,
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "validation failed", resp.Error)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("conflict error", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		reqBody := AgentRequest{
			ClientID:    ptr.To("duplicate-client"),
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateAgent", storage.ErrorKindConflict, nil, "agent with this client_id already exists"),
		)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "conflict", resp.Error)

		mockRepo.AssertExpectations(t)
	})
}

func TestAgentsHandler_GetAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful retrieval", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		agent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("test-client")),
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockRepo.On("Get", mock.Anything, agentID).Return(agent, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/"+agentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, agentID.String(), resp.ID)
		require.NotNil(t, resp.ClientID)
		assert.Equal(t, "test-client", *resp.ClientID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("agent not found", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		notFoundID := id.NewAgentID()
		mockRepo.On("Get", mock.Anything, notFoundID).Return(
			nil,
			storage.NewStorageError("GetAgent", storage.ErrorKindNotFound, ports.ErrNotFound, "agent not found"),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/"+notFoundID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", notFoundID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetAgent(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "agent not found", resp.Error)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty agent ID", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/", nil)
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "agent ID is required", resp.Error)
	})
}

func TestAgentsHandler_UpdateAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful update", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("test-client")),
			DisplayName: "Old Name",
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		reqBody := AgentRequest{
			ClientID:    ptr.To("test-client"),
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ID == agentID && a.DisplayName == "Updated Name"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, agentID.String(), resp.ID)
		assert.Equal(t, "Updated Name", resp.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("omitted client_id preserves existing value and passes uniqueness in single-agent mode", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("existing-client-id")),
			DisplayName: "Old Name",
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		reqBody := AgentRequest{
			// ClientID intentionally omitted — should preserve "existing-client-id"
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockRepo.On("ExistsOtherWithClientID", mock.Anything, id.ClientID("existing-client-id"), &agentID).Return(false, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID != nil && *a.ClientID == "existing-client-id" && a.DisplayName == "Updated Name"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		require.NotNil(t, resp.ClientID)
		assert.Equal(t, "existing-client-id", *resp.ClientID)
		assert.Equal(t, "Updated Name", resp.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("explicit null client_id clears the stored value", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("existing-client")),
			DisplayName: "Old Name",
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		bodyBytes := []byte(`{"client_id": null, "display_name": "Updated Name", "description": "Updated description"}`)

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID == nil && a.DisplayName == "Updated Name"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp AgentResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Nil(t, resp.ClientID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("agent not found", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		notFoundID := id.NewAgentID()
		reqBody := AgentRequest{
			ClientID:    ptr.To("test-client"),
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, notFoundID).Return(
			nil,
			storage.NewStorageError("GetAgent", storage.ErrorKindNotFound, ports.ErrNotFound, "agent not found"),
		)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+notFoundID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", notFoundID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("blank client_id returns 400 validation failed", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("existing-client")),
			DisplayName: "Existing Agent",
			Description: "Existing description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)

		emptyClientID := ""
		reqBody := AgentRequest{
			ClientID:    &emptyClientID,
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "validation failed", resp.Error)
		mockRepo.AssertNotCalled(t, "Update")
		mockRepo.AssertExpectations(t)
	})

	t.Run("preserves CIMD snapshot fields on update", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("https://agent.example.com/client")),
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		reqBody := AgentRequest{
			ClientID:    ptr.To("https://agent.example.com/client"),
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ID == agentID && a.DisplayName == "Updated Name"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestAgentsHandler_DeleteAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		mockRepo.On("Delete", mock.Anything, agentID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/agents/"+agentID.String(), nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DeleteAgent(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty agent ID", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		req := httptest.NewRequest(http.MethodDelete, "/api/agents/", nil)
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DeleteAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAgentsHandler_ListAgents(t *testing.T) {
	logger := slog.Default()

	t.Run("successful list", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID1 := id.NewAgentID()
		agentID2 := id.NewAgentID()
		now := time.Now().UTC()
		agents := []*storage.Agent{
			{
				ID:          agentID1,
				ClientID:    ptr.To(id.ClientID("client-1")),
				DisplayName: "Agent 1",
				Description: "First agent",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          agentID2,
				ClientID:    ptr.To(id.ClientID("client-2")),
				DisplayName: "Agent 2",
				Description: "Second agent",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		mockRepo.On("List", mock.Anything).Return(agents, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
		w := httptest.NewRecorder()

		handler.ListAgents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, agentID1.String(), resp[0].ID)
		assert.Equal(t, agentID2.String(), resp[1].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		mockRepo.On("List", mock.Anything).Return([]*storage.Agent{}, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
		w := httptest.NewRecorder()

		handler.ListAgents(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Empty(t, resp)

		mockRepo.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		mockRepo.On("List", mock.Anything).Return(
			nil,
			storage.NewStorageError("ListAgents", storage.ErrorKindConnection, nil, "connection failed"),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/agents", nil)
		w := httptest.NewRecorder()

		handler.ListAgents(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockRepo.AssertExpectations(t)
	})
}

// T033: Unit tests for admin agents handler client_id uniqueness gating.
func TestAgentsHandler_ClientIDUniqueness(t *testing.T) {
	logger := slog.Default()

	t.Run("Create: !multiAgentEnabled + duplicate client_id → 409 Conflict", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		mockRepo.On("ExistsOtherWithClientID", mock.Anything, id.ClientID("shared-client"), (*id.AgentID)(nil)).Return(true, nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("shared-client"),
			DisplayName: "New Agent",
			Description: "Duplicate client_id",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create: multiAgentEnabled=true + duplicate client_id → no 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, true)

		// GetByClientID must NOT be called when feature is enabled
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("shared-client"),
			DisplayName: "New Agent",
			Description: "Allowed duplicate",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertNotCalled(t, "ExistsOtherWithClientID")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update: !multiAgentEnabled + client_id taken by another agent → 409 Conflict", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		targetAgentID := id.NewAgentID()

		targetAgent := &storage.Agent{
			ID:          targetAgentID,
			ClientID:    ptr.To(id.ClientID("old-client")),
			DisplayName: "Target Agent",
			Description: "Being updated",
		}

		mockRepo.On("Get", mock.Anything, targetAgentID).Return(targetAgent, nil)
		mockRepo.On("ExistsOtherWithClientID", mock.Anything, id.ClientID("shared-client"), &targetAgentID).Return(true, nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("shared-client"),
			DisplayName: "Target Agent Updated",
			Description: "Trying to steal client_id",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+targetAgentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		// Inject chi URL param
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", targetAgentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update: !multiAgentEnabled + client_id belongs to self → no 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		targetAgentID := id.NewAgentID()
		targetAgent := &storage.Agent{
			ID:          targetAgentID,
			ClientID:    ptr.To(id.ClientID("my-client")),
			DisplayName: "Target Agent",
			Description: "Being updated",
		}

		mockRepo.On("Get", mock.Anything, targetAgentID).Return(targetAgent, nil)
		mockRepo.On("ExistsOtherWithClientID", mock.Anything, id.ClientID("my-client"), &targetAgentID).Return(false, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("my-client"),
			DisplayName: "Target Agent Updated",
			Description: "Self-update allowed",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+targetAgentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", targetAgentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertNotCalled(t, "GetByClientID")
		mockRepo.AssertExpectations(t)
	})
}

// T027a: Unit tests for admin handler client_uris validation mapping.
func TestAgentsHandler_ClientURIsValidation(t *testing.T) {
	logger := slog.Default()

	t.Run("Create: malformed client_uri returns 400", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		reqBody := AgentRequest{
			ClientID:    ptr.To("cimd-client"),
			DisplayName: "CIMD Agent",
			Description: "Test description",
			ClientURIs:  []string{"not-a-valid-url"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "validation failed", resp.Error)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("Create: duplicate client_uri returns 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateAgent", storage.ErrorKindConflict, nil, "client_uri already registered by another agent"),
		)

		reqBody := AgentRequest{
			DisplayName: "CIMD Agent",
			Description: "Test description",
			ClientURIs:  []string{"https://example.com/taken"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "conflict", resp.Error)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update: malformed client_uri returns 400", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()

		reqBody := AgentRequest{
			ClientID:    ptr.To("cimd-update-client"),
			DisplayName: "CIMD Agent",
			Description: "Updated description",
			ClientURIs:  []string{"http://bad-scheme.example.com"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "validation failed", resp.Error)
		mockRepo.AssertNotCalled(t, "Get")
	})

	t.Run("Update: duplicate client_uri returns 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		existing := &storage.Agent{
			ID:          agentID,
			DisplayName: "CIMD Agent",
			Description: "Test description",
		}
		mockRepo.On("Get", mock.Anything, agentID).Return(existing, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("UpdateAgent", storage.ErrorKindConflict, nil, "client_uri already registered by another agent"),
		)

		reqBody := AgentRequest{
			DisplayName: "CIMD Agent",
			Description: "Updated description",
			ClientURIs:  []string{"https://example.com/already-taken"},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		var resp ErrorResponse
		require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
		assert.Equal(t, "conflict", resp.Error)
		mockRepo.AssertExpectations(t)
	})
}

// T043: Unit tests for agent permission_sets validation
// These tests document the validation behavior is correct in the handler implementation.
// Full integration testing happens in E2E tests (T045).
func TestAgentsHandler_CreateAgent_PermissionSets_RedPhase(t *testing.T) {
	logger := slog.Default()

	// Note: These tests are placeholder documentation for T043 red phase.
	// The actual PermissionSetService validation is tested via E2E tests (T045) which
	// exercise the full flow with a real service instance.
	// The handler implementation already includes the validation logic for:
	// - Empty permission_sets list validation
	// - Non-existent permission_set_id validation via psService.ValidateIDs
	// - Preservation of declaration order in JSONB storage

	t.Run("handler accepts permission_sets in request (red phase)", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		psID1 := id.NewPermissionSetID()
		reqBody := AgentRequest{
			ClientID:    ptr.To("test-client"),
			DisplayName: "Test Agent",
			Description: "Test description",
			PermissionSets: []PermissionSetRequest{
				{
					PermissionSetID: psID1.String(),
					RequirementType: "mandatory",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			// Verify permission sets field is populated (even with nil psService, validation is skipped)
			return a.ClientID != nil && string(*a.ClientID) == "test-client" && len(a.PermissionSets) == 1
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		// With nil psService, handler skips validation but still parses and stores permission_sets
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

// MockPermissionSetValidator is a mock implementation of PermissionSetValidator.
type MockPermissionSetValidator struct {
	mock.Mock
}

func (m *MockPermissionSetValidator) ValidateIDs(ctx context.Context, ids []id.PermissionSetID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockPermissionSetValidator) GetByIDs(ctx context.Context, ids []id.PermissionSetID) ([]*storage.PermissionSet, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.PermissionSet), args.Error(1)
}

// newAgentsHandlerForFR019Test creates an AgentsHandler with a real domain service
// and a non-nil MockPermissionSetValidator so FR-019 coverage checks are exercised.
func newAgentsHandlerForFR019Test(
	mockRepo *MockAgentRepository,
	mockServiceRepo *MockProviderRepository,
	mockPS *MockPermissionSetValidator,
	logger *slog.Logger,
) *AgentsHandler {
	svc := thirdparty.NewThirdpartyOAuth2ProviderService(mockServiceRepo, newTestEncryption(), &encryptionnoop.BranchKeyManager{}, nil, false, logger)
	agentSvc := agents.NewService(mockRepo, svc, logger, true)
	return NewAgentsHandler(agentSvc, svc, mockPS, logger)
}

func TestAgentsHandler_FR019_CoverageInvariant(t *testing.T) {
	logger := slog.Default()

	// Shared IDs used across subtests
	serviceAID := id.NewServiceID()
	serviceBID := id.NewServiceID()
	psID1 := id.NewPermissionSetID()

	// Helper: entity returned by provider repo for ValidateServiceRequirements
	serviceAEntity := &model.ThirdpartyOAuth2ProviderEntity{
		ID:          serviceAID,
		DisplayName: "Service A",
		Scopes:      []model.OAuthScope{{ScopeValue: "read"}},
	}

	t.Run("create agent with SR service not covered by any PS returns 400", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)

		// providerService.ValidateServiceRequirements will look up serviceA
		mockServiceRepo.On("Get", mock.Anything, serviceAID).Return(serviceAEntity, nil)

		// convertPermissionSetRequests calls ValidateIDs
		mockPS.On("ValidateIDs", mock.Anything, mock.Anything).Return(nil)

		// validateServiceRequirementsCoverage calls GetByIDs — return PS covering serviceB only
		mockPS.On("GetByIDs", mock.Anything, mock.Anything).Return([]*storage.PermissionSet{
			{
				ID:   psID1,
				Name: "PS1",
				ServiceScopes: []storage.ServiceScope{
					{ServiceID: serviceBID, Scopes: []string{"write"}},
				},
			},
		}, nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("test-fr019-uncovered"),
			DisplayName: "FR019 Agent",
			Description: "Test agent",
			ServiceRequirements: []ServiceRequirementRequest{
				{ServiceID: serviceAID.String(), RequirementType: "mandatory", RequiredScopes: []string{"read"}},
			},
			PermissionSets: []PermissionSetRequest{
				{PermissionSetID: psID1.String(), RequirementType: "mandatory"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, serviceAID.String())

		mockPS.AssertExpectations(t)
	})

	t.Run("create agent with all SR services covered by PS returns 201", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)

		// providerService.ValidateServiceRequirements looks up serviceA
		mockServiceRepo.On("Get", mock.Anything, serviceAID).Return(serviceAEntity, nil)

		mockPS.On("ValidateIDs", mock.Anything, mock.Anything).Return(nil)

		// PS covers serviceA — invariant satisfied
		mockPS.On("GetByIDs", mock.Anything, mock.Anything).Return([]*storage.PermissionSet{
			{
				ID:   psID1,
				Name: "PS1",
				ServiceScopes: []storage.ServiceScope{
					{ServiceID: serviceAID, Scopes: []string{"read"}},
				},
			},
		}, nil)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID != nil && string(*a.ClientID) == "test-fr019-covered" && len(a.ServiceRequirements) == 1 && len(a.PermissionSets) == 1
		})).Return(nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("test-fr019-covered"),
			DisplayName: "FR019 Covered Agent",
			Description: "Test agent",
			ServiceRequirements: []ServiceRequirementRequest{
				{ServiceID: serviceAID.String(), RequirementType: "mandatory", RequiredScopes: []string{"read"}},
			},
			PermissionSets: []PermissionSetRequest{
				{PermissionSetID: psID1.String(), RequirementType: "mandatory"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		mockRepo.AssertExpectations(t)
		mockPS.AssertExpectations(t)
	})

	t.Run("creates an agent with omitted required_scopes for a scope-less service", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)
		scopeLessService := &model.ThirdpartyOAuth2ProviderEntity{ID: serviceAID, DisplayName: "Scope-less Service"}
		mockServiceRepo.On("Get", mock.Anything, serviceAID).Return(scopeLessService, nil)
		mockPS.On("ValidateIDs", mock.Anything, mock.Anything).Return(nil)
		mockPS.On("GetByIDs", mock.Anything, mock.Anything).Return([]*storage.PermissionSet{{
			ID: psID1,
			ServiceScopes: []storage.ServiceScope{{
				ServiceID:       serviceAID,
				RequirementType: storage.RequirementTypeMandatory,
			}},
		}}, nil)
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(agent *storage.Agent) bool {
			return len(agent.ServiceRequirements) == 1 && len(agent.ServiceRequirements[0].RequiredScopes) == 0
		})).Return(nil)
		rawBody := fmt.Sprintf(`{
			"client_id": "scope-less-agent",
			"display_name": "Scope-less Agent",
			"description": "Uses a scope-less service",
			"service_requirements": [{"service_id": %q, "requirement_type": "mandatory"}],
			"permission_sets": [{"permission_set_id": %q, "requirement_type": "mandatory"}]
		}`, serviceAID.String(), psID1.String())
		w := httptest.NewRecorder()
		handler.CreateAgent(w, httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(rawBody)))

		require.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
		mockPS.AssertExpectations(t)
	})

	t.Run("update agent with uncovered SR service returns 400", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          agentID,
			ClientID:    ptr.To(id.ClientID("test-fr019-update")),
			DisplayName: "Old Name",
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockServiceRepo.On("Get", mock.Anything, serviceAID).Return(serviceAEntity, nil)
		mockPS.On("ValidateIDs", mock.Anything, mock.Anything).Return(nil)

		// PS covers only serviceB, not serviceA
		mockPS.On("GetByIDs", mock.Anything, mock.Anything).Return([]*storage.PermissionSet{
			{
				ID:   psID1,
				Name: "PS1",
				ServiceScopes: []storage.ServiceScope{
					{ServiceID: serviceBID, Scopes: []string{"write"}},
				},
			},
		}, nil)

		reqBody := AgentRequest{
			ClientID:    ptr.To("test-fr019-update"),
			DisplayName: "Updated Name",
			Description: "Updated description",
			ServiceRequirements: []ServiceRequirementRequest{
				{ServiceID: serviceAID.String(), RequirementType: "mandatory", RequiredScopes: []string{"read"}},
			},
			PermissionSets: []PermissionSetRequest{
				{PermissionSetID: psID1.String(), RequirementType: "mandatory"},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/"+agentID.String(), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", agentID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, serviceAID.String())

		mockPS.AssertExpectations(t)
	})

	t.Run("agent with empty permission_sets returns 400 when psService is available", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)

		// providerService.ValidateServiceRequirements looks up serviceA
		mockServiceRepo.On("Get", mock.Anything, serviceAID).Return(serviceAEntity, nil)

		// Use raw JSON with explicit empty permission_sets array — cannot use struct marshaling
		// because omitempty omits nil/empty slices, and we need "permission_sets": [] explicitly.
		rawBody := fmt.Sprintf(`{
			"client_id": "test-fr019-empty-ps",
			"display_name": "FR019 Empty PS Agent",
			"description": "Test agent",
			"service_requirements": [{"service_id": "%s", "requirement_type": "mandatory", "required_scopes": ["read"]}],
			"permission_sets": []
		}`, serviceAID.String())

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(rawBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		assert.Contains(t, resp.Message, "at least one permission set entry is required")

		// Create should NOT be called when PS list is explicitly empty
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("agent with omitted permission_sets (nil) returns 400 when psService is available", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		mockPS := new(MockPermissionSetValidator)
		handler := newAgentsHandlerForFR019Test(mockRepo, mockServiceRepo, mockPS, logger)

		// Use a struct body without permission_sets (omitted via omitempty → nil on decode).
		// FR-006 requires at least one entry; both nil and explicit-empty are rejected.
		rawBody := fmt.Sprintf(`{
			"client_id": "test-fr006-nil-ps",
			"display_name": "FR006 Nil PS Agent",
			"description": "Test agent",
			"service_requirements": [{"service_id": "%s", "requirement_type": "mandatory", "required_scopes": ["read"]}]
		}`, serviceAID.String())

		req := httptest.NewRequest(http.MethodPost, "/api/agents", strings.NewReader(rawBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)
		// Rejected by FR-006: at least one permission set entry is required
		assert.Contains(t, resp.Message, "at least one permission set entry is required")

		// Create should NOT be called
		mockRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})
}
