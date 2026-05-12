package fixtures

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ptr"
)

// ValidAgent returns a valid test agent with all required fields.
// ClientID: test-client-valid
// DisplayName: Test Agent Valid
// RedirectURIs: ["https://client.example.com/cb"]
// ID and timestamps are generated fresh for each call.
func ValidAgent() *storage.Agent {
	now := time.Now()
	return &storage.Agent{
		ID:           id.NewAgentID(),
		ClientID:     ptr.To(id.ClientID("test-client-valid")),
		DisplayName:  "Test Agent Valid",
		Description:  "A valid test agent for E2E testing with all required fields",
		RedirectURIs: []string{"https://client.example.com/cb"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AnotherAgent returns an alternative valid test agent.
// ClientID: test-client-another
// DisplayName: Test Agent Another
// RedirectURIs: ["https://client.example.com/cb"]
// ID and timestamps are generated fresh for each call.
func AnotherAgent() *storage.Agent {
	now := time.Now()
	return &storage.Agent{
		ID:           id.NewAgentID(),
		ClientID:     ptr.To(id.ClientID("test-client-another")),
		DisplayName:  "Test Agent Another",
		Description:  "Another valid test agent for E2E testing with different client ID",
		RedirectURIs: []string{"https://client.example.com/cb"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AgentWithClientID returns a valid test agent with a specific client ID.
// Useful for testing scenarios that require a known client ID.
func AgentWithClientID(clientID string) *storage.Agent {
	now := time.Now()
	return &storage.Agent{
		ID:          id.NewAgentID(),
		ClientID:    ptr.To(id.ClientID(clientID)),
		DisplayName: "Test Agent " + clientID,
		Description: "Test agent with custom client ID: " + clientID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AgentWithURLs returns a valid test agent with all optional URL fields.
// Includes governance URL, user documentation URL, and agent interface URL.
func AgentWithURLs() *storage.Agent {
	now := time.Now()
	governanceURL := "https://governance.example.com/agent"
	userDocURL := "https://docs.example.com/user-guide"
	agentInterfaceURL := "https://agent.example.com"

	return &storage.Agent{
		ID:                   id.NewAgentID(),
		ClientID:             ptr.To(id.ClientID("test-client-with-urls")),
		DisplayName:          "Agent With URLs",
		Description:          "Test agent with all URL fields populated for documentation and governance",
		GovernanceURL:        &governanceURL,
		UserDocumentationURL: &userDocURL,
		AgentInterfaceURL:    &agentInterfaceURL,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}
