package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserGrant_Validate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name    string
		grant   *UserGrant
		wantErr string
	}{
		{
			name: "valid grant with expiration",
			grant: &UserGrant{
				ID:         "750e8400-e29b-41d4-a716-446655440002",
				Principal:  "user123@example.com",
				AgentID:    "550e8400-e29b-41d4-a716-446655440000",
				ValidUntil: &future,
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo", "user:email"},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "valid grant without expiration (indefinite)",
			grant: &UserGrant{
				ID:         "750e8400-e29b-41d4-a716-446655440002",
				Principal:  "user123@example.com",
				AgentID:    "550e8400-e29b-41d4-a716-446655440000",
				ValidUntil: nil,
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "multiple delegations",
			grant: &UserGrant{
				ID:        "750e8400-e29b-41d4-a716-446655440002",
				Principal: "user123@example.com",
				AgentID:   "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440003",
						Scopes:                    []string{"read:user", "read:org"},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "missing ID",
			grant: &UserGrant{
				Principal: "user123@example.com",
				AgentID:   "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "grant ID cannot be empty",
		},
		{
			name: "missing principal",
			grant: &UserGrant{
				ID:      "750e8400-e29b-41d4-a716-446655440002",
				AgentID: "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "principal is required",
		},
		{
			name: "missing agent_id",
			grant: &UserGrant{
				ID:        "750e8400-e29b-41d4-a716-446655440002",
				Principal: "user123@example.com",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "agent_id is required",
		},
		{
			name: "valid_until in the past",
			grant: &UserGrant{
				ID:         "750e8400-e29b-41d4-a716-446655440002",
				Principal:  "user123@example.com",
				AgentID:    "550e8400-e29b-41d4-a716-446655440000",
				ValidUntil: &past,
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "valid_until must be in the future",
		},
		{
			name: "no delegations",
			grant: &UserGrant{
				ID:                    "750e8400-e29b-41d4-a716-446655440002",
				Principal:             "user123@example.com",
				AgentID:               "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{},
			},
			wantErr: "at least one delegated service is required",
		},
		{
			name: "delegation missing service_id",
			grant: &UserGrant{
				ID:        "750e8400-e29b-41d4-a716-446655440002",
				Principal: "user123@example.com",
				AgentID:   "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						Scopes: []string{"repo"},
					},
				},
			},
			wantErr: "delegation 0: thirdparty_oauth2_service_id is required",
		},
		{
			name: "delegation with no scopes",
			grant: &UserGrant{
				ID:        "750e8400-e29b-41d4-a716-446655440002",
				Principal: "user123@example.com",
				AgentID:   "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{},
					},
				},
			},
			wantErr: "delegation 0: at least one scope is required",
		},
		{
			name: "delegation with duplicate scopes",
			grant: &UserGrant{
				ID:        "750e8400-e29b-41d4-a716-446655440002",
				Principal: "user123@example.com",
				AgentID:   "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo", "user:email", "repo"},
					},
				},
			},
			wantErr: "delegation 0: duplicate scope 'repo'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.grant.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUserGrant_IsActive(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name   string
		grant  *UserGrant
		active bool
	}{
		{
			name: "indefinite grant is active",
			grant: &UserGrant{
				ValidUntil: nil,
			},
			active: true,
		},
		{
			name: "future expiration is active",
			grant: &UserGrant{
				ValidUntil: &future,
			},
			active: true,
		},
		{
			name: "past expiration is not active",
			grant: &UserGrant{
				ValidUntil: &past,
			},
			active: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.active, tt.grant.IsActive())
		})
	}
}

func TestUserGrant_Copy(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	original := &UserGrant{
		ID:         "750e8400-e29b-41d4-a716-446655440002",
		Principal:  "user123@example.com",
		AgentID:    "550e8400-e29b-41d4-a716-446655440000",
		ValidUntil: &future,
		DelegatedOAuth2Tokens: []DelegatedToken{
			{
				ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
				Scopes:                    []string{"repo", "user:email"},
			},
			{
				ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440003",
				Scopes:                    []string{"read:user"},
			},
		},
	}

	copy := original.Copy()

	// Verify equality
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.Principal, copy.Principal)
	assert.Equal(t, original.AgentID, copy.AgentID)
	assert.Equal(t, len(original.DelegatedOAuth2Tokens), len(copy.DelegatedOAuth2Tokens))

	// Verify deep copy of scopes
	copy.DelegatedOAuth2Tokens[0].Scopes[0] = "modified"
	assert.Equal(t, "repo", original.DelegatedOAuth2Tokens[0].Scopes[0])
	assert.Equal(t, "modified", copy.DelegatedOAuth2Tokens[0].Scopes[0])

	// Verify deep copy of ValidUntil
	*copy.ValidUntil = time.Now().Add(48 * time.Hour)
	assert.NotEqual(t, original.ValidUntil.Unix(), copy.ValidUntil.Unix())
}

func TestUserGrant_Copy_Nil(t *testing.T) {
	var grant *UserGrant
	copy := grant.Copy()
	assert.Nil(t, copy)
}

func TestUserGrant_ValidateForCreate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		grant   *UserGrant
		wantErr string
	}{
		{
			name: "valid grant for creation",
			grant: &UserGrant{
				Principal:  "user123@example.com",
				AgentID:    "550e8400-e29b-41d4-a716-446655440000",
				ValidUntil: &future,
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "missing principal",
			grant: &UserGrant{
				AgentID: "550e8400-e29b-41d4-a716-446655440000",
				DelegatedOAuth2Tokens: []DelegatedToken{
					{
						ThirdpartyOAuth2ServiceID: "650e8400-e29b-41d4-a716-446655440001",
						Scopes:                    []string{"repo"},
					},
				},
			},
			wantErr: "principal is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.grant.ValidateForCreate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
