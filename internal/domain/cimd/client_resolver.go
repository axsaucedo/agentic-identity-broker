package cimd

import (
	"context"
	"errors"
	"strings"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// CIMDClientResolver resolves client_id values by:
//   - Detecting URL-format client_id (https://) → CIMD fetch path
//   - Falling back to opaque UUID lookup for non-URL client_id
type CIMDClientResolver struct {
	agentRepo   ports.AgentRepository
	cimdService *Service
}

// NewCIMDClientResolver creates a CIMD-enabled client resolver.
func NewCIMDClientResolver(agentRepo ports.AgentRepository, cimdService *Service) *CIMDClientResolver {
	return &CIMDClientResolver{
		agentRepo:   agentRepo,
		cimdService: cimdService,
	}
}

// ResolveClient resolves a client_id to an Agent and optional CIMD metadata.
// For URL-format client_id: validates URL, looks up agent by client URI, fetches/validates CIMD.
// For opaque client_id: parses UUID and looks up agent by ID.
func (r *CIMDClientResolver) ResolveClient(ctx context.Context, clientID id.ClientID) (*ports.ClientResolution, error) {
	if strings.Contains(string(clientID), "://") {
		return r.resolveCIMD(ctx, string(clientID))
	}
	return r.resolveOpaque(ctx, clientID)
}

func (r *CIMDClientResolver) resolveCIMD(ctx context.Context, rawURL string) (*ports.ClientResolution, error) {
	// Validate URL format first (fast rejection before any I/O)
	if _, err := ParseClientIDMetadataDocumentURL(rawURL); err != nil {
		return nil, &ports.ClientIDError{Code: "invalid_request", Desc: "invalid client_id URL: " + err.Error()}
	}

	// Look up agent by pre-registered client URI (FR-026)
	agent, err := r.agentRepo.GetByClientURI(ctx, rawURL)
	if err != nil {
		if isNotFoundErr(err) {
			return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
		}
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "Failed to validate client"}
	}

	// Fetch and validate CIMD document
	doc, err := r.cimdService.Resolve(ctx, rawURL, agent)
	if err != nil {
		var snapErr *SnapshotPersistenceError
		if errors.As(err, &snapErr) {
			return nil, &ports.ClientIDError{Code: "server_error", Desc: "Internal error validating client"}
		}
		return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "CIMD document validation failed: " + err.Error()}
	}

	return &ports.ClientResolution{Agent: agent, CIMDMetadata: toDTO(doc)}, nil
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
		RedirectURIs:  doc.RedirectURIs,
		AuthMethod:    doc.AuthMethod,
		GrantTypes:    doc.GrantTypes,
		ResponseTypes: doc.ResponseTypes,
		JwksURI:       doc.JwksURI,
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
		if isNotFoundErr(err) {
			return nil, &ports.ClientIDError{Code: "invalid_client", Desc: "Client not registered"}
		}
		return nil, &ports.ClientIDError{Code: "server_error", Desc: "Failed to validate client"}
	}

	return &ports.ClientResolution{Agent: agent}, nil
}

func isNotFoundErr(err error) bool {
	if errors.Is(err, ports.ErrNotFound) {
		return true
	}
	var storageErr *storage.StorageError
	return errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound
}
