package integration_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestListSessions_Success(t *testing.T) {
	// Setup
	sessionRepo := memory.NewInMemoryUserSessionRepository()
	serviceRepo := memory.NewThirdpartyServiceRepository()
	grantRepo := memory.NewUserGrantRepository()
	ctx := context.Background()

	principal := "user@example.com"
	serviceID := uuid.New().String()

	// Create service first
	service := &storage.ThirdpartyOAuth2Service{
		ID:          serviceID,
		DisplayName: "GitHub",
		ClientID:    "test-client-id",
		ClientSecret: "test-secret",
		IssuerURI:   "https://github.com",
		Endpoints: storage.OAuth2Endpoints{
			AuthorizeEndpoint: "https://github.com/login/oauth/authorize",
			TokenEndpoint:     "https://github.com/login/oauth/access_token",
		},
		Scopes: []storage.OAuthScope{
			{ScopeValue: "repo", Description: "Repository access"},
		},
	}
	err := serviceRepo.Create(ctx, service)
	assert.NoError(t, err)

	// Create a session
	session := &storage.UserSession{
		ID:                   uuid.New().String(),
		Principal:            principal,
		ServiceID:            serviceID,
		EncryptedAccessToken: []byte("token"),
		TokenType:            "Bearer",
		Scope:                []string{"repo", "user"},
		InitiatedAt:          time.Now(),
		CreatedAt:            time.Now(),
	}
	err = sessionRepo.Create(ctx, session)
	assert.NoError(t, err)

	// Create service with dependencies
	service2 := oauth2session.NewOAuth2SessionService(
		serviceRepo,
		sessionRepo,
		grantRepo,
		nil, // encryption not needed for this test
		nil, // jweKey not needed for this test
		oauth2session.DefaultConfig(),
		slog.Default(),
	)

	// Create handler
	handler := oauth2_sessions.NewHandler(
		sessionRepo,
		grantRepo,
		service2,
	)

	// Make request
	req := httptest.NewRequest("GET", "/api/third-party/sessions", nil)
	req.Header.Set("X-Remote-User", principal)
	w := httptest.NewRecorder()

	handler.ListSessions(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			Sessions []map[string]interface{} `json:"sessions"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Len(t, resp.Data.Sessions, 1)

	// Verify session data
	sessionData := resp.Data.Sessions[0]
	assert.Equal(t, session.ID, sessionData["id"])
	assert.Equal(t, serviceID, sessionData["service_id"])
	assert.Equal(t, "GitHub", sessionData["service_display_name"])
	assert.Equal(t, "Bearer", sessionData["token_type"])
}

func TestListSessions_MissingPrincipal(t *testing.T) {
	// Setup
	sessionRepo := memory.NewInMemoryUserSessionRepository()
	grantRepo := memory.NewUserGrantRepository()
	serviceRepo := memory.NewThirdpartyServiceRepository()

	service := oauth2session.NewOAuth2SessionService(
		serviceRepo,
		sessionRepo,
		grantRepo,
		nil,
		nil,
		oauth2session.DefaultConfig(),
		slog.Default(),
	)

	handler := oauth2_sessions.NewHandler(
		sessionRepo,
		grantRepo,
		service,
	)

	// Make request without X-Remote-User header
	req := httptest.NewRequest("GET", "/api/third-party/sessions", nil)
	w := httptest.NewRecorder()

	handler.ListSessions(w, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "unauthorized", resp["error"])
}

func TestListSessions_EmptySessions(t *testing.T) {
	// Setup
	sessionRepo := memory.NewInMemoryUserSessionRepository()
	grantRepo := memory.NewUserGrantRepository()
	serviceRepo := memory.NewThirdpartyServiceRepository()

	service := oauth2session.NewOAuth2SessionService(
		serviceRepo,
		sessionRepo,
		grantRepo,
		nil,
		nil,
		oauth2session.DefaultConfig(),
		slog.Default(),
	)

	handler := oauth2_sessions.NewHandler(
		sessionRepo,
		grantRepo,
		service,
	)

	// Make request for user with no sessions
	req := httptest.NewRequest("GET", "/api/third-party/sessions", nil)
	req.Header.Set("X-Remote-User", "user@example.com")
	w := httptest.NewRecorder()

	handler.ListSessions(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Data struct {
			Sessions []interface{} `json:"sessions"`
		} `json:"data"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Empty(t, resp.Data.Sessions)
}
