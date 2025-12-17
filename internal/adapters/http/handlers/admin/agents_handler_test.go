package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
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

func (m *MockAgentRepository) Get(ctx context.Context, id string) (*storage.Agent, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*storage.Agent), args.Error(1)
}

func (m *MockAgentRepository) Update(ctx context.Context, agent *storage.Agent) error {
	args := m.Called(ctx, agent)
	return args.Error(0)
}

func (m *MockAgentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockAgentRepository) List(ctx context.Context) ([]*storage.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*storage.Agent), args.Error(1)
}

func TestAgentsHandler_CreateAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		reqBody := AgentRequest{
			ClientID:    "test-client",
			DisplayName: "Test Agent",
			Description: "Test agent description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ClientID == "test-client" && a.DisplayName == "Test Agent"
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
		assert.Equal(t, "test-client", resp.ClientID)
		assert.Equal(t, "Test Agent", resp.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("with optional fields", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		externalID := "ext-123"
		govURL := "https://example.com/gov"
		reqBody := AgentRequest{
			ClientID:      "test-client-2",
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
		handler := NewAgentsHandler(mockRepo, logger)

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

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		reqBody := AgentRequest{
			ClientID:    "", // Empty client_id
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateAgent", storage.ErrorKindValidation, nil, "client_id is required"),
		)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var resp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "validation failed", resp.Error)

		mockRepo.AssertExpectations(t)
	})

	t.Run("conflict error", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		reqBody := AgentRequest{
			ClientID:    "duplicate-client",
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
		handler := NewAgentsHandler(mockRepo, logger)

		now := time.Now().UTC()
		agent := &storage.Agent{
			ID:          "agent-123",
			ClientID:    "test-client",
			DisplayName: "Test Agent",
			Description: "Test description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mockRepo.On("Get", mock.Anything, "agent-123").Return(agent, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/agent-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.GetAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "agent-123", resp.ID)
		assert.Equal(t, "test-client", resp.ClientID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("agent not found", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		mockRepo.On("Get", mock.Anything, "non-existent").Return(
			nil,
			storage.NewStorageError("GetAgent", storage.ErrorKindNotFound, ports.ErrNotFound, "agent not found"),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/agents/non-existent", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "non-existent")
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
		handler := NewAgentsHandler(mockRepo, logger)

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
		handler := NewAgentsHandler(mockRepo, logger)

		now := time.Now().UTC()
		existingAgent := &storage.Agent{
			ID:          "agent-123",
			ClientID:    "test-client",
			DisplayName: "Old Name",
			Description: "Old description",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		reqBody := AgentRequest{
			ClientID:    "test-client",
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, "agent-123").Return(existingAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ID == "agent-123" && a.DisplayName == "Updated Name"
		})).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/agent-123", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp AgentResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "agent-123", resp.ID)
		assert.Equal(t, "Updated Name", resp.DisplayName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("agent not found", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		reqBody := AgentRequest{
			ClientID:    "test-client",
			DisplayName: "Test Agent",
			Description: "Test description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, "non-existent").Return(
			nil,
			storage.NewStorageError("GetAgent", storage.ErrorKindNotFound, ports.ErrNotFound, "agent not found"),
		)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/non-existent", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "non-existent")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		req := httptest.NewRequest(http.MethodPut, "/api/agents/agent-123", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.UpdateAgent(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAgentsHandler_DeleteAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful deletion", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

		mockRepo.On("Delete", mock.Anything, "agent-123").Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api/agents/agent-123", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("agent-id", "agent-123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()

		handler.DeleteAgent(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty agent ID", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

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
		handler := NewAgentsHandler(mockRepo, logger)

		now := time.Now().UTC()
		agents := []*storage.Agent{
			{
				ID:          "agent-1",
				ClientID:    "client-1",
				DisplayName: "Agent 1",
				Description: "First agent",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          "agent-2",
				ClientID:    "client-2",
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
		assert.Equal(t, "agent-1", resp[0].ID)
		assert.Equal(t, "agent-2", resp[1].ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		handler := NewAgentsHandler(mockRepo, logger)

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
		handler := NewAgentsHandler(mockRepo, logger)

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
