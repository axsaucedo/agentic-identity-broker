package fixtures

import (
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// SharedUpstreamClientID is the upstream OAuth2 client_id shared by multi-agent test agents.
// Multiple agents sharing this client_id is the core of feature 021.
const SharedUpstreamClientID = "shared-upstream"

// MultiAgentAlpha returns the first test agent sharing SharedUpstreamClientID.
// Its agent.id (UUID) is used as the OAuth2 client_id in broker requests.
// ClientID "shared-upstream" is the upstream OAuth2 client credential.
func MultiAgentAlpha() *storage.Agent {
	now := time.Now()
	return &storage.Agent{
		ID:          id.NewAgentID(),
		ClientID:    id.ClientID(SharedUpstreamClientID),
		DisplayName: "Multi-Agent Alpha",
		Description: "First agent sharing an upstream OAuth2 client_id (feature 021 testing)",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// MultiAgentBeta returns the second test agent sharing SharedUpstreamClientID.
// Distinct agent.id from MultiAgentAlpha, same upstream ClientID.
// Used to verify that claim verification correctly rejects cross-agent token use.
func MultiAgentBeta() *storage.Agent {
	now := time.Now()
	return &storage.Agent{
		ID:          id.NewAgentID(),
		ClientID:    id.ClientID(SharedUpstreamClientID),
		DisplayName: "Multi-Agent Beta",
		Description: "Second agent sharing an upstream OAuth2 client_id (feature 021 testing)",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// MultiAgentEnabledConfig returns a ports.Config with multi_agent_client feature enabled.
// The upstream URL is pointed at the provided mock server.
//
// Multi-agent settings:
//   - Enabled: true
//   - AgentIDParamName: "x_agent_id"  (injected on upstream authorize redirect)
//   - AgentIDClaimName: "x_agent_id"  (verified in proxied token responses)
//
// Token exchange is configured to read the agent ID directly from the "x_agent_id" claim:
//
//	agent_id_expression: "subject_token.x_agent_id"
//
// NOTE: This fixture references ports.MultiAgentClientConfig which is added in Phase 2.5 (T015/T016).
// Tests using this fixture will not compile until Phase 2.5 is complete.
func MultiAgentEnabledConfig(upstreamURL string) *ports.Config {
	config := OAuth2ConfigWithUpstream(upstreamURL)
	config.OAuth2AuthServer.MultiAgentClient = ports.MultiAgentClientConfig{
		Enabled:          true,
		AgentIDParamName: "x_agent_id",
		AgentIDClaimName: "x_agent_id",
	}
	// When feature is enabled, the CEL policy reads the agent ID claim directly
	config.TokenExchange.ClaimExtraction.AgentIDExpression = "subject_token.x_agent_id"
	return config
}

// MultiAgentDisabledConfig returns a ports.Config with multi_agent_client feature disabled.
// The upstream URL is pointed at the provided mock server.
//
// When disabled, upstream client_id uniqueness is enforced and the resolveAgentIdByClientId
// CEL helper function is registered for token exchange policies.
//
// Token exchange is configured to use the resolveAgentIdByClientId function:
//
//	agent_id_expression: "resolveAgentIdByClientId(subject_token.azp)"
//
// NOTE: This fixture references ports.MultiAgentClientConfig which is added in Phase 2.5 (T015/T016).
// Tests using this fixture will not compile until Phase 2.5 is complete.
func MultiAgentDisabledConfig(upstreamURL string) *ports.Config {
	config := OAuth2ConfigWithUpstream(upstreamURL)
	config.OAuth2AuthServer.MultiAgentClient = ports.MultiAgentClientConfig{
		Enabled: false,
	}
	// When feature is disabled, CEL uses resolveAgentIdByClientId to map upstream client_id → agent.id
	config.TokenExchange.ClaimExtraction.AgentIDExpression = "resolveAgentIdByClientId(subject_token.azp)"
	return config
}
