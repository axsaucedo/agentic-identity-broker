package oauth2

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/urivalidation"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OpaqueClientResolver resolves a client_id as a UUID-based agent identifier.
// It rejects any client_id with a URI scheme (URL-format) with invalid_client,
// enforcing the CIMD gate when cimd.enabled is false.
type OpaqueClientResolver struct {
	agentRepo ports.AgentRepository
}

// NewOpaqueClientResolver creates a resolver for opaque (UUID-based) client IDs.
func NewOpaqueClientResolver(agentRepo ports.AgentRepository) *OpaqueClientResolver {
	return &OpaqueClientResolver{agentRepo: agentRepo}
}

// ResolveClient parses the client_id as a UUID and looks up the agent.
// Returns invalid_client if the client_id has a URI scheme (CIMD disabled gate).
func (r *OpaqueClientResolver) ResolveClient(ctx context.Context, clientID id.ClientID) (*ports.ClientResolution, error) {
	if urivalidation.ValidateCIMDClientURL(string(clientID)) == nil {
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "URL-format client_id requires CIMD support (disabled)"}
	}

	agentUUID, parseErr := id.ParseAgentID(string(clientID))
	if parseErr != nil {
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
	}

	agent, err := r.agentRepo.Get(ctx, agentUUID)
	if err != nil {
		if ports.IsNotFoundErr(err) {
			return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
		}
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "Failed to validate client"}
	}

	return &ports.ClientResolution{Agent: agent}, nil
}
