package fixtures

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/google/uuid"
)

// GitHubSessionForPrincipal returns a fixture for an active GitHub OAuth2 session.
// The session contains encrypted access and refresh tokens for the given principal
// to use the GitHub third-party service.
// This is used in token exchange tests to provide stored tokens.
//
// Parameters:
//   - principal: The user principal (e.g., "user@example.com")
//
// Returns:
//   - A UserSession with:
//   - Access token valid for 1 hour
//   - Refresh token valid for 30 days
//   - Token type: Bearer
//   - Scopes: repo, user
func GitHubSessionForPrincipal(principal string) *storage.UserSession {
	now := time.Now()
	accessTokenExpires := now.Add(1 * time.Hour)
	refreshTokenExpires := now.Add(30 * 24 * time.Hour)

	return &storage.UserSession{
		ID:                    uuid.New().String(),
		Principal:             principal,
		ServiceID:             "github-service",
		EncryptedAccessToken:  []byte("encrypted-github-access-token"),  // Mock encrypted token
		EncryptedRefreshToken: []byte("encrypted-github-refresh-token"), // Mock encrypted token
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  &accessTokenExpires,
		RefreshTokenExpiresAt: &refreshTokenExpires,
		Scope:                 []string{"repo", "user"},
		EncryptionContext: storage.EncryptionContext{
			Principal: principal,
			ServiceID: "github-service",
			SessionID: uuid.New().String(),
			Purpose:   "oauth2_token",
		},
		InitiatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// ExpiredGitHubSessionForPrincipal returns a GitHub session where the access token is expired.
// This is used for testing token refresh logic in token exchange.
func ExpiredGitHubSessionForPrincipal(principal string) *storage.UserSession {
	now := time.Now()
	accessTokenExpired := now.Add(-1 * time.Hour) // Already expired
	refreshTokenValid := now.Add(7 * 24 * time.Hour)

	return &storage.UserSession{
		ID:                    uuid.New().String(),
		Principal:             principal,
		ServiceID:             "github-service",
		EncryptedAccessToken:  []byte("encrypted-expired-access-token"),
		EncryptedRefreshToken: []byte("encrypted-refresh-token"),
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  &accessTokenExpired,
		RefreshTokenExpiresAt: &refreshTokenValid,
		Scope:                 []string{"repo", "user"},
		EncryptionContext: storage.EncryptionContext{
			Principal: principal,
			ServiceID: "github-service",
			SessionID: uuid.New().String(),
			Purpose:   "oauth2_token",
		},
		InitiatedAt: now.Add(-2 * time.Hour),
		CreatedAt:   now.Add(-2 * time.Hour),
		UpdatedAt:   now.Add(-1 * time.Hour),
	}
}

// FullyExpiredSessionForPrincipal returns a session where both access and refresh tokens are expired.
// This is used for testing the error case where tokens cannot be refreshed.
func FullyExpiredSessionForPrincipal(principal, serviceID string) *storage.UserSession {
	now := time.Now()
	accessTokenExpired := now.Add(-2 * time.Hour)
	refreshTokenExpired := now.Add(-1 * time.Hour)

	return &storage.UserSession{
		ID:                    uuid.New().String(),
		Principal:             principal,
		ServiceID:             serviceID,
		EncryptedAccessToken:  []byte("encrypted-expired-access-token"),
		EncryptedRefreshToken: []byte("encrypted-expired-refresh-token"),
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  &accessTokenExpired,
		RefreshTokenExpiresAt: &refreshTokenExpired,
		Scope:                 []string{"read", "write"},
		EncryptionContext: storage.EncryptionContext{
			Principal: principal,
			ServiceID: serviceID,
			SessionID: uuid.New().String(),
			Purpose:   "oauth2_token",
		},
		InitiatedAt: now.Add(-7 * 24 * time.Hour),
		CreatedAt:   now.Add(-7 * 24 * time.Hour),
		UpdatedAt:   now.Add(-1 * time.Hour),
	}
}

// SessionForService returns a generic active session for the given principal and service.
// Useful for testing with different services.
func SessionForService(principal, serviceID string) *storage.UserSession {
	now := time.Now()
	accessTokenExpires := now.Add(1 * time.Hour)
	refreshTokenExpires := now.Add(30 * 24 * time.Hour)

	return &storage.UserSession{
		ID:                    uuid.New().String(),
		Principal:             principal,
		ServiceID:             serviceID,
		EncryptedAccessToken:  []byte("encrypted-access-token-" + serviceID),
		EncryptedRefreshToken: []byte("encrypted-refresh-token-" + serviceID),
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  &accessTokenExpires,
		RefreshTokenExpiresAt: &refreshTokenExpires,
		Scope:                 []string{"read", "write"},
		EncryptionContext: storage.EncryptionContext{
			Principal: principal,
			ServiceID: serviceID,
			SessionID: uuid.New().String(),
			Purpose:   "oauth2_token",
		},
		InitiatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// SessionWithoutRefreshToken returns an active session that does NOT have a refresh token.
// Access token is valid. This is used for testing scenarios where refresh is not possible.
func SessionWithoutRefreshToken(principal, serviceID string) *storage.UserSession {
	now := time.Now()
	accessTokenExpires := now.Add(1 * time.Hour)

	return &storage.UserSession{
		ID:                    uuid.New().String(),
		Principal:             principal,
		ServiceID:             serviceID,
		EncryptedAccessToken:  []byte("encrypted-access-token"),
		EncryptedRefreshToken: nil, // No refresh token
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  &accessTokenExpires,
		RefreshTokenExpiresAt: nil, // No refresh token expiration
		Scope:                 []string{"read"},
		EncryptionContext: storage.EncryptionContext{
			Principal: principal,
			ServiceID: serviceID,
			SessionID: uuid.New().String(),
			Purpose:   "oauth2_token",
		},
		InitiatedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
