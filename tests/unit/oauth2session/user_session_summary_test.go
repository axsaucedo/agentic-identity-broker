package oauth2session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

func TestNewUserSessionSummary(t *testing.T) {
	now := time.Now()
	refreshExp := now.Add(24 * time.Hour)

	session := &storage.UserSession{
		ID:                    "session-123",
		ServiceID:             "service-uuid",
		TokenType:             "Bearer",
		Scope:                 []string{"repo", "user"},
		InitiatedAt:           now.Add(-1 * time.Hour),
		RefreshTokenExpiresAt: &refreshExp,
	}

	summary := storage.NewUserSessionSummary(session, "GitHub", 3)

	assert.Equal(t, "session-123", summary.ID)
	assert.Equal(t, "service-uuid", summary.ServiceID)
	assert.Equal(t, "GitHub", summary.ServiceDisplayName)
	assert.Equal(t, "Bearer", summary.TokenType)
	assert.Equal(t, []string{"repo", "user"}, summary.Scope)
	assert.Equal(t, 3, summary.DependentAgentCount)
	assert.True(t, summary.IsEncrypted)
	assert.False(t, summary.IsExpired)
}

func TestUserSessionSummary_WithExpiredSession(t *testing.T) {
	now := time.Now()
	refreshExp := now.Add(-1 * time.Hour) // Expired

	session := &storage.UserSession{
		ID:                    "session-123",
		ServiceID:             "service-uuid",
		RefreshTokenExpiresAt: &refreshExp,
	}

	summary := storage.NewUserSessionSummary(session, "GitHub", 0)

	assert.True(t, summary.IsExpired)
}

func TestUserSessionSummary_WithExpiredAccessToken(t *testing.T) {
	now := time.Now()
	accessExp := now.Add(-1 * time.Hour) // Expired access token
	refreshExp := now.Add(24 * time.Hour) // Valid refresh token

	session := &storage.UserSession{
		ID:                    "session-123",
		ServiceID:             "service-uuid",
		AccessTokenExpiresAt:  &accessExp,
		RefreshTokenExpiresAt: &refreshExp,
	}

	summary := storage.NewUserSessionSummary(session, "GitHub", 2)

	assert.True(t, summary.AccessTokenExpired)
	assert.False(t, summary.IsExpired) // Session not expired if refresh token valid
}
