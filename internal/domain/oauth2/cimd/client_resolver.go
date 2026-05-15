package cimd

import (
	"context"
	"log/slog"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/urivalidation"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// CIMDClientResolver resolves client_id values by:
//   - Detecting URL-format client_id (https://) → CIMD fetch path
//   - Falling back to opaque UUID lookup for non-URL client_id
type CIMDClientResolver struct {
	agentRepo   ports.AgentRepository
	cimdService *Service
	logger      *slog.Logger
}

// NewCIMDClientResolver creates a CIMD-enabled client resolver.
func NewCIMDClientResolver(agentRepo ports.AgentRepository, cimdService *Service, logger *slog.Logger) *CIMDClientResolver {
	return &CIMDClientResolver{
		agentRepo:   agentRepo,
		cimdService: cimdService,
		logger:      logger,
	}
}

// ResolveClient resolves a client_id to an Agent and optional CIMD metadata.
// For URL-format client_id: validates URL, looks up agent by client URI, fetches/validates CIMD.
// For opaque client_id: parses UUID and looks up agent by ID.
func (r *CIMDClientResolver) ResolveClient(ctx context.Context, clientID id.ClientID) (*ports.ClientResolution, error) {
	if urivalidation.ValidateCIMDClientURL(string(clientID)) == nil {
		return r.resolveCIMD(ctx, string(clientID))
	}
	return r.resolveOpaque(ctx, clientID)
}

func (r *CIMDClientResolver) resolveCIMD(ctx context.Context, rawURL string) (*ports.ClientResolution, error) {
	// Look up agent by pre-registered client URI (FR-026)
	agent, err := r.agentRepo.GetByClientURI(ctx, rawURL)
	if err != nil {
		if ports.IsNotFoundErr(err) {
			return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
		}
		r.logger.ErrorContext(ctx, "failed to look up agent by client URI", "error", err, "client_uri", rawURL)
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "Failed to validate client"}
	}

	if agent.ClientMode() == storage.AmbiguousClient {
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
	}

	// Fetch and validate CIMD document
	doc, err := r.cimdService.Resolve(ctx, rawURL, agent)
	if err != nil {
		r.logger.WarnContext(ctx, "CIMD document validation failed", "error", err, "client_uri", rawURL)
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "CIMD document validation failed"}
	}

	return ports.NewClientResolution(agent, toDTO(doc)), nil
}

// toDTO converts a domain ClientIDMetadataDocument to the port DTO.
func toDTO(doc *ClientIDMetadataDocument) *ports.CIMDMetadataDTO {
	if doc == nil {
		return nil
	}
	return &ports.CIMDMetadataDTO{
		ClientID:      doc.ClientID,
		ClientName:    doc.ClientName,
		LogoURI:       doc.LogoURI,
		RedirectURIs:  append([]string(nil), doc.RedirectURIs...),
		AuthMethod:    doc.AuthMethod,
		GrantTypes:    append([]string(nil), doc.GrantTypes...),
		ResponseTypes: append([]string(nil), doc.ResponseTypes...),
		PolicyURI:     doc.PolicyURI,
		TosURI:        doc.TosURI,
	}
}

func (r *CIMDClientResolver) resolveOpaque(ctx context.Context, clientID id.ClientID) (*ports.ClientResolution, error) {
	agentUUID, parseErr := id.ParseAgentID(string(clientID))
	if parseErr != nil {
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
	}

	agent, err := r.agentRepo.Get(ctx, agentUUID)
	if err != nil {
		if ports.IsNotFoundErr(err) {
			return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
		}
		r.logger.ErrorContext(ctx, "failed to look up agent by ID", "error", err, "agent_id", agentUUID)
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "Failed to validate client"}
	}

	// CIMD agents must be addressed by URL, not UUID. Ambiguous agents (both fields set) are invalid.
	mode := agent.ClientMode()
	if mode == storage.CIMDClient || mode == storage.AmbiguousClient {
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
	}

	return ports.NewClientResolution(agent, nil), nil
}
