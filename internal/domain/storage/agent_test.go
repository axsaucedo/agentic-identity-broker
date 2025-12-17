package storage

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent_Validate(t *testing.T) {
	validGovernanceURL := "https://governance.example.com"
	validUserDocURL := "https://docs.example.com"
	validAgentURL := "https://chat.example.com"

	tests := []struct {
		name    string
		agent   *Agent
		wantErr string
	}{
		{
			name: "valid agent with all fields",
			agent: &Agent{
				ID:                   "550e8400-e29b-41d4-a716-446655440000",
				ClientID:             "test-client",
				DisplayName:          "Test Agent",
				Description:          "A test agent for validation",
				GovernanceURL:        &validGovernanceURL,
				UserDocumentationURL: &validUserDocURL,
				AgentInterfaceURL:    &validAgentURL,
			},
			wantErr: "",
		},
		{
			name: "valid agent with minimal fields",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				ClientID:    "test-client",
				DisplayName: "Test Agent",
				Description: "A test agent",
			},
			wantErr: "",
		},
		{
			name: "missing ID",
			agent: &Agent{
				ClientID:    "test-client",
				DisplayName: "Test Agent",
				Description: "A test agent",
			},
			wantErr: "agent ID cannot be empty",
		},
		{
			name: "missing client_id",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				DisplayName: "Test Agent",
				Description: "A test agent",
			},
			wantErr: "client_id is required",
		},
		{
			name: "missing display_name",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				ClientID:    "test-client",
				Description: "A test agent",
			},
			wantErr: "display_name is required",
		},
		{
			name: "display_name too long",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				ClientID:    "test-client",
				DisplayName: strings.Repeat("a", 256),
				Description: "A test agent",
			},
			wantErr: "display_name exceeds 255 characters",
		},
		{
			name: "missing description",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				ClientID:    "test-client",
				DisplayName: "Test Agent",
			},
			wantErr: "description is required",
		},
		{
			name: "description too long",
			agent: &Agent{
				ID:          "550e8400-e29b-41d4-a716-446655440000",
				ClientID:    "test-client",
				DisplayName: "Test Agent",
				Description: strings.Repeat("a", 1001),
			},
			wantErr: "description exceeds 1000 characters",
		},
		{
			name: "invalid governance_url",
			agent: &Agent{
				ID:            "550e8400-e29b-41d4-a716-446655440000",
				ClientID:      "test-client",
				DisplayName:   "Test Agent",
				Description:   "A test agent",
				GovernanceURL: stringPtr("not-a-url"),
			},
			wantErr: "governance_url is not a valid HTTP/HTTPS URL",
		},
		{
			name: "invalid user_documentation_url",
			agent: &Agent{
				ID:                   "550e8400-e29b-41d4-a716-446655440000",
				ClientID:             "test-client",
				DisplayName:          "Test Agent",
				Description:          "A test agent",
				UserDocumentationURL: stringPtr("ftp://invalid.com"),
			},
			wantErr: "user_documentation_url is not a valid HTTP/HTTPS URL",
		},
		{
			name: "invalid agent_interface_url",
			agent: &Agent{
				ID:                "550e8400-e29b-41d4-a716-446655440000",
				ClientID:          "test-client",
				DisplayName:       "Test Agent",
				Description:       "A test agent",
				AgentInterfaceURL: stringPtr("javascript:alert(1)"),
			},
			wantErr: "agent_interface_url is not a valid HTTP/HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgent_ValidateForCreate(t *testing.T) {
	validGovernanceURL := "https://governance.example.com"

	tests := []struct {
		name    string
		agent   *Agent
		wantErr string
	}{
		{
			name: "valid agent for creation",
			agent: &Agent{
				ClientID:      "test-client",
				DisplayName:   "Test Agent",
				Description:   "A test agent",
				GovernanceURL: &validGovernanceURL,
			},
			wantErr: "",
		},
		{
			name: "missing client_id",
			agent: &Agent{
				DisplayName: "Test Agent",
				Description: "A test agent",
			},
			wantErr: "client_id is required",
		},
		{
			name: "display_name exceeds limit with character count",
			agent: &Agent{
				ClientID:    "test-client",
				DisplayName: strings.Repeat("a", 256),
				Description: "A test agent",
			},
			wantErr: "display_name exceeds 255 characters (got 256)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.agent.ValidateForCreate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgent_Copy(t *testing.T) {
	externalID := "ext-123"
	governanceURL := "https://governance.example.com"
	userDocURL := "https://docs.example.com"
	agentURL := "https://chat.example.com"

	original := &Agent{
		ID:                   "550e8400-e29b-41d4-a716-446655440000",
		ClientID:             "test-client",
		ExternalID:           &externalID,
		DisplayName:          "Test Agent",
		Description:          "A test agent",
		GovernanceURL:        &governanceURL,
		UserDocumentationURL: &userDocURL,
		AgentInterfaceURL:    &agentURL,
	}

	// Create copy
	copy := original.Copy()

	// Verify copy is equal
	assert.Equal(t, original.ID, copy.ID)
	assert.Equal(t, original.ClientID, copy.ClientID)
	assert.Equal(t, original.DisplayName, copy.DisplayName)
	assert.Equal(t, original.Description, copy.Description)
	require.NotNil(t, copy.ExternalID)
	assert.Equal(t, *original.ExternalID, *copy.ExternalID)
	require.NotNil(t, copy.GovernanceURL)
	assert.Equal(t, *original.GovernanceURL, *copy.GovernanceURL)

	// Verify deep copy (modifying copy doesn't affect original)
	*copy.ExternalID = "modified"
	assert.Equal(t, "ext-123", *original.ExternalID)
	assert.Equal(t, "modified", *copy.ExternalID)
}

func TestAgent_Copy_Nil(t *testing.T) {
	var agent *Agent
	copy := agent.Copy()
	assert.Nil(t, copy)
}

func stringPtr(s string) *string {
	return &s
}
