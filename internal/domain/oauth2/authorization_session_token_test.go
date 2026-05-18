package oauth2

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewAuthorizationSessionClaims_NilCIMDMetadata verifies that nil CIMDMetadata
// produces valid claims containing all required fields with correct 10-minute TTL (T009).
func TestNewAuthorizationSessionClaims_NilCIMDMetadata(t *testing.T) {
	agentID := id.NewAgentID()
	principal := id.NewPrincipal("user@example.com")
	originalURL := "https://broker.example.com/oauth2/authorize?client_id=test"

	before := time.Now()
	claims, err := NewAuthorizationSessionClaims(agentID, principal, originalURL, nil)
	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, agentID, claims.AgentID)
	assert.Equal(t, principal, claims.Principal)
	assert.Equal(t, originalURL, claims.OriginalURL)
	assert.Nil(t, claims.CIMDMetadata, "nil CIMDMetadata must produce claims with nil CIMDMetadata")

	assert.True(t, !claims.IssuedAt.Before(before) && !claims.IssuedAt.After(after))
	ttl := claims.ExpiresAt.Sub(claims.IssuedAt)
	assert.InDelta(t, authorizationSessionTokenTTL.Seconds(), ttl.Seconds(), 1)

	assert.False(t, claims.IsExpired())
}

// TestNewAuthorizationSessionClaims_ValidationErrors verifies required field validation.
func TestNewAuthorizationSessionClaims_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		agentID     id.AgentID
		principal   id.Principal
		originalURL string
		wantErr     string
	}{
		{
			name:        "zero agent ID returns error",
			agentID:     id.AgentID{},
			principal:   id.NewPrincipal("user@example.com"),
			originalURL: "https://example.com",
			wantErr:     "agentID must not be zero",
		},
		{
			name:        "zero principal returns error",
			agentID:     id.NewAgentID(),
			principal:   id.Principal(""),
			originalURL: "https://example.com",
			wantErr:     "principal must not be zero",
		},
		{
			name:        "empty originalURL returns error",
			agentID:     id.NewAgentID(),
			principal:   id.NewPrincipal("user@example.com"),
			originalURL: "",
			wantErr:     "originalURL must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := NewAuthorizationSessionClaims(tt.agentID, tt.principal, tt.originalURL, nil)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Nil(t, claims)
		})
	}
}
