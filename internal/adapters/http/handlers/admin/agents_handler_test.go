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

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
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

// newAgentsHandlerForTest creates an AgentsHandler backed by a real domain service
// wrapping a mock repository. This ensures the architecture invariant holds in tests:
// the handler always goes through the domain service, never raw storage.
//
// MockProviderRepository and newTestEncryption are defined in services_handler_test.go
// and are available here because both files share the same package admin.
func newAgentsHandlerForTest(mockRepo *MockAgentRepository, mockServiceRepo *MockProviderRepository, logger *slog.Logger) *AgentsHandler {
	svc := thirdparty.NewThirdpartyOAuth2ProviderService(mockServiceRepo, newTestEncryption(), nil, false, logger)
	// Use multiAgentEnabled=true so existing CRUD tests don't need GetByClientID expectations.
	// T033 tests use newAgentsHandlerForTestWithMultiAgent with explicit flags.
	return NewAgentsHandler(mockRepo, svc, logger, true)
}

// newAgentsHandlerForTestWithMultiAgent creates an AgentsHandler with the given multiAgentEnabled flag.
func newAgentsHandlerForTestWithMultiAgent(mockRepo *MockAgentRepository, mockServiceRepo *MockProviderRepository, logger *slog.Logger, multiAgentEnabled bool) *AgentsHandler {
	svc := thirdparty.NewThirdpartyOAuth2ProviderService(mockServiceRepo, newTestEncryption(), nil, false, logger)
	return NewAgentsHandler(mockRepo, svc, logger, multiAgentEnabled)
}

func TestAgentsHandler_CreateAgent(t *testing.T) {
	logger := slog.Default()

	t.Run("successful creation", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

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
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

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

	t.Run("validation error", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

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
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

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
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		agent := &storage.Agent{
			ID:          agentID,
			ClientID:    "test-client",
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
		assert.Equal(t, "test-client", resp.ClientID)

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

	t.Run("agent not found", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		notFoundID := id.NewAgentID()
		reqBody := AgentRequest{
			ClientID:    "test-client",
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

	t.Run("preserves CIMD snapshot fields on update", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		now := time.Now().UTC()
		authMethod := "private_key_jwt"
		jwksURI := "https://agent.example.com/.well-known/jwks.json"
		cimdName := "Test Agent"
		cimdLogo := "https://agent.example.com/logo.png"
		existingAgent := &storage.Agent{
			ID:               agentID,
			ClientID:         "https://agent.example.com/client",
			DisplayName:      "Test Agent",
			Description:      "Test description",
			AuthMethod:       &authMethod,
			JwksURI:          &jwksURI,
			CIMDClientName:   &cimdName,
			CIMDLogoURI:      &cimdLogo,
			CIMDRedirectURIs: []string{"https://agent.example.com/callback"},
			CreatedAt:        now,
			UpdatedAt:        now,
		}

		reqBody := AgentRequest{
			ClientID:    "https://agent.example.com/client",
			DisplayName: "Updated Name",
			Description: "Updated description",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		mockRepo.On("Get", mock.Anything, agentID).Return(existingAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(a *storage.Agent) bool {
			return a.ID == agentID &&
				a.AuthMethod != nil && *a.AuthMethod == authMethod &&
				a.JwksURI != nil && *a.JwksURI == jwksURI &&
				a.CIMDClientName != nil && *a.CIMDClientName == cimdName &&
				a.CIMDLogoURI != nil && *a.CIMDLogoURI == cimdLogo &&
				len(a.CIMDRedirectURIs) == 1 && a.CIMDRedirectURIs[0] == "https://agent.example.com/callback"
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
				ClientID:    "client-1",
				DisplayName: "Agent 1",
				Description: "First agent",
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          agentID2,
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

	existingAgentID := id.NewAgentID()

	t.Run("Create: !multiAgentEnabled + duplicate client_id → 409 Conflict", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		existingAgent := &storage.Agent{
			ID:          existingAgentID,
			ClientID:    id.ClientID("shared-client"),
			DisplayName: "Existing Agent",
			Description: "Already registered",
		}

		// Handler should call GetByClientID before Create; return a conflict
		mockRepo.On("GetByClientID", mock.Anything, id.ClientID("shared-client")).Return(existingAgent, nil)

		reqBody := AgentRequest{
			ClientID:    "shared-client",
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
			ClientID:    "shared-client",
			DisplayName: "New Agent",
			Description: "Allowed duplicate",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/agents", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateAgent(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertNotCalled(t, "GetByClientID")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update: !multiAgentEnabled + client_id taken by another agent → 409 Conflict", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTestWithMultiAgent(mockRepo, mockServiceRepo, logger, false)

		targetAgentID := id.NewAgentID()
		otherAgentID := id.NewAgentID() // different agent that already owns the client_id

		targetAgent := &storage.Agent{
			ID:          targetAgentID,
			ClientID:    id.ClientID("old-client"),
			DisplayName: "Target Agent",
			Description: "Being updated",
		}
		otherAgent := &storage.Agent{
			ID:          otherAgentID,
			ClientID:    id.ClientID("shared-client"),
			DisplayName: "Other Agent",
			Description: "Already has shared-client",
		}

		mockRepo.On("Get", mock.Anything, targetAgentID).Return(targetAgent, nil)
		// GetByClientID for the new client_id returns a DIFFERENT agent — conflict
		mockRepo.On("GetByClientID", mock.Anything, id.ClientID("shared-client")).Return(otherAgent, nil)

		reqBody := AgentRequest{
			ClientID:    "shared-client",
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
			ClientID:    id.ClientID("my-client"),
			DisplayName: "Target Agent",
			Description: "Being updated",
		}

		mockRepo.On("Get", mock.Anything, targetAgentID).Return(targetAgent, nil)
		// GetByClientID returns the SAME agent — no conflict (self-update)
		mockRepo.On("GetByClientID", mock.Anything, id.ClientID("my-client")).Return(targetAgent, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		reqBody := AgentRequest{
			ClientID:    "my-client",
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

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateAgent", storage.ErrorKindValidation, nil, "client_uris[0] is not a valid HTTPS URL"),
		)

		reqBody := AgentRequest{
			ClientID:    "cimd-client",
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
		mockRepo.AssertExpectations(t)
	})

	t.Run("Create: duplicate client_uri returns 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("CreateAgent", storage.ErrorKindConflict, nil, "client_uri already registered by another agent"),
		)

		reqBody := AgentRequest{
			ClientID:    "cimd-conflict-client",
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
		existing := &storage.Agent{
			ID:          agentID,
			ClientID:    "cimd-update-client",
			DisplayName: "CIMD Agent",
			Description: "Test description",
		}
		mockRepo.On("Get", mock.Anything, agentID).Return(existing, nil)

		reqBody := AgentRequest{
			ClientID:    "cimd-update-client",
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
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update: duplicate client_uri returns 409", func(t *testing.T) {
		mockRepo := new(MockAgentRepository)
		mockServiceRepo := new(MockProviderRepository)
		handler := newAgentsHandlerForTest(mockRepo, mockServiceRepo, logger)

		agentID := id.NewAgentID()
		existing := &storage.Agent{
			ID:          agentID,
			ClientID:    "cimd-update-conflict-client",
			DisplayName: "CIMD Agent",
			Description: "Test description",
		}
		mockRepo.On("Get", mock.Anything, agentID).Return(existing, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(
			storage.NewStorageError("UpdateAgent", storage.ErrorKindConflict, nil, "client_uri already registered by another agent"),
		)

		reqBody := AgentRequest{
			ClientID:    "cimd-update-conflict-client",
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
