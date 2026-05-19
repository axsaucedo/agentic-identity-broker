package sessiontoken

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	key, err := jwk.Import([]byte("test-32-byte-key-must-be-exact-x"))
	require.NoError(t, err)
	return NewService(jwe.New(key))
}

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
	assert.Nil(t, claims.CIMDMetadata)

	assert.True(t, !claims.IssuedAt.Before(before) && !claims.IssuedAt.After(after))
	assert.InDelta(t, (10 * time.Minute).Seconds(), claims.ExpiresAt.Sub(claims.IssuedAt).Seconds(), 1)
	assert.False(t, claims.IsExpired())
}

func TestNewAuthorizationSessionClaims_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		agentID     id.AgentID
		principal   id.Principal
		originalURL string
		wantErr     string
	}{
		{
			name:        "zero agent ID",
			agentID:     id.AgentID{},
			principal:   id.NewPrincipal("user@example.com"),
			originalURL: "https://example.com",
			wantErr:     "agentID must not be zero",
		},
		{
			name:        "zero principal",
			agentID:     id.NewAgentID(),
			principal:   id.Principal(""),
			originalURL: "https://example.com",
			wantErr:     "principal must not be zero",
		},
		{
			name:        "empty originalURL",
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

func TestService_ValidateAuthorizationSessionToken_Success(t *testing.T) {
	svc := newTestService(t)
	agentID := id.NewAgentID()
	p := id.NewPrincipal("user@example.com")

	claims, _ := NewAuthorizationSessionClaims(agentID, p, "https://example.com/authorize?client_id=test", nil)
	token, err := svc.Create(claims)
	require.NoError(t, err)

	validated, err := svc.ValidateAuthorizationSessionToken(token, agentID, p)
	require.NoError(t, err)
	assert.Equal(t, agentID, validated.AgentID)
	assert.Equal(t, p, validated.Principal)
}

func TestService_ValidateAuthorizationSessionToken_Expired(t *testing.T) {
	svc := newTestService(t)
	agentID := id.NewAgentID()
	p := id.NewPrincipal("user@example.com")

	past := time.Now().Add(-time.Hour)
	claims := &AuthorizationSessionClaims{
		AgentID:     agentID,
		Principal:   p,
		OriginalURL: "https://example.com/authorize",
		IssuedAt:    past,
		ExpiresAt:   past,
	}
	token, err := svc.jweTokenService.Encrypt(claims)
	require.NoError(t, err)

	_, err = svc.ValidateAuthorizationSessionToken(token, agentID, p)
	assert.ErrorIs(t, err, ErrSessionExpired)
}

func TestService_ValidateAuthorizationSessionToken_AgentMismatch(t *testing.T) {
	svc := newTestService(t)
	agentID := id.NewAgentID()
	otherAgent := id.NewAgentID()
	p := id.NewPrincipal("user@example.com")

	claims, _ := NewAuthorizationSessionClaims(agentID, p, "https://example.com/authorize?client_id=test", nil)
	token, err := svc.Create(claims)
	require.NoError(t, err)

	_, err = svc.ValidateAuthorizationSessionToken(token, otherAgent, p)
	assert.ErrorIs(t, err, ErrSessionAgentMismatch)
}

func TestService_ValidateAuthorizationSessionToken_PrincipalMismatch(t *testing.T) {
	svc := newTestService(t)
	agentID := id.NewAgentID()
	p := id.NewPrincipal("user@example.com")
	otherP := id.NewPrincipal("other@example.com")

	claims, _ := NewAuthorizationSessionClaims(agentID, p, "https://example.com/authorize?client_id=test", nil)
	token, err := svc.Create(claims)
	require.NoError(t, err)

	_, err = svc.ValidateAuthorizationSessionToken(token, agentID, otherP)
	assert.ErrorIs(t, err, ErrSessionPrincipalMismatch)
}

func TestService_ValidateAuthorizationSessionToken_InvalidToken(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ValidateAuthorizationSessionToken("garbage-token", id.NewAgentID(), id.NewPrincipal("x"))
	assert.ErrorIs(t, err, ErrSessionInvalidToken)
}
