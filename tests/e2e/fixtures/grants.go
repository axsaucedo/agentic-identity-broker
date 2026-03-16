package fixtures

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// ActiveGrant returns a user grant that is currently active (not expired).
// Principal, agentID, and serviceID must be provided by caller.
// ValidUntil is set to 1 hour in the future.
func ActiveGrant(principal, agentID, serviceID string, scopes []string) *storage.UserGrant {
	now := time.Now()
	validUntil := now.Add(1 * time.Hour)

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID(serviceID),
				Scopes:                    scopes,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ExpiredGrant returns a user grant that has already expired.
// Principal, agentID, and serviceID must be provided by caller.
// ValidUntil is set to 1 hour in the past.
func ExpiredGrant(principal, agentID, serviceID string, scopes []string) *storage.UserGrant {
	now := time.Now()
	validUntil := now.Add(-1 * time.Hour)

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID(serviceID),
				Scopes:                    scopes,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GrantExpiringIn returns a user grant that expires in the specified duration.
// Useful for testing edge cases like grants expiring soon.
func GrantExpiringIn(principal, agentID, serviceID string, scopes []string, duration time.Duration) *storage.UserGrant {
	now := time.Now()
	validUntil := now.Add(duration)

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID(serviceID),
				Scopes:                    scopes,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IndefiniteGrant returns a user grant that never expires (ValidUntil is nil).
// Indefinite grants remain active until explicitly revoked.
func IndefiniteGrant(principal, agentID, serviceID string, scopes []string) *storage.UserGrant {
	now := time.Now()

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: nil, // No expiration
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID(serviceID),
				Scopes:                    scopes,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GrantWithMultipleServices returns a user grant with multiple delegated OAuth2 services.
// Useful for testing scopes across different providers.
func GrantWithMultipleServices(principal, agentID string) *storage.UserGrant {
	now := time.Now()
	validUntil := now.Add(24 * time.Hour)

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID("a0000000-0000-0000-0000-000000000001"),
				Scopes:                    []string{"repo", "user"},
			},
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID("a0000000-0000-0000-0000-000000000002"),
				Scopes:                    []string{"calendar", "drive"},
			},
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID("a0000000-0000-0000-0000-000000000003"),
				Scopes:                    []string{"mail.read", "calendar.read"},
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GrantWithService returns a user grant for a specific OAuth2 service.
// Useful for testing service-specific grant scenarios.
func GrantWithService(principal, agentID, serviceID string, scopes []string) *storage.UserGrant {
	now := time.Now()
	validUntil := now.Add(30 * 24 * time.Hour) // 30 days

	return &storage.UserGrant{
		ID:         id.NewGrantID(),
		Principal:  id.Principal(principal),
		AgentID:    id.MustParseAgentID(agentID),
		ValidUntil: &validUntil,
		DelegatedOAuth2Tokens: []storage.DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: id.MustParseServiceID(serviceID),
				Scopes:                    scopes,
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
