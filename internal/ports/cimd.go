package ports

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// ClientResolver resolves a client_id from an authorization request to an Agent
// and optional CIMD metadata. The builder selects the implementation based on
// cimd.enabled configuration.
//
// When CIMD is disabled: OpaqueClientResolver is wired — rejects URL-format
// client_id values with invalid_client immediately.
//
// When CIMD is enabled: CIMDClientResolver is wired — handles URL-format
// client_id via CIMD fetch/validate/cache, delegates non-URL client_id to
// opaque UUID resolution.
type ClientResolver interface {
	ResolveClient(ctx context.Context, clientID id.ClientID) (*ClientResolution, error)
}

// ClientResolution is the result of a client_id resolution.
// CIMDMetadata is nil for opaque (UUID-based) client_id values.
type ClientResolution struct {
	Agent        *storage.Agent
	CIMDMetadata *CIMDMetadataDTO
}

// CIMDMetadataDTO carries CIMD document fields relevant to the authorization flow.
// It is a port-layer DTO — not the domain type — to avoid import cycles between
// ports/ and domain/oauth2server/cimd.
type CIMDMetadataDTO struct {
	ClientID      string
	ClientName    string
	LogoURI       string
	RedirectURIs  []string
	AuthMethod    string
	GrantTypes    []string
	ResponseTypes []string
	PolicyURI     string
	TosURI        string
}

// CIMDFetcher fetches Client ID Metadata Documents from remote HTTPS endpoints
// with SSRF protection, timeout, and size limits.
type CIMDFetcher interface {
	Fetch(ctx context.Context, url string) (*CIMDFetchResult, error)
}

// CIMDFetchResult holds the result of a successful CIMD document fetch.
// The body is raw bytes to avoid an import cycle — callers parse it with
// cimd.ParseDocument from internal/domain/oauth2server/cimd.
type CIMDFetchResult struct {
	Body         []byte
	CacheControl string
	Expires      string
}

// ClientIDError signals a client resolution failure with an OAuth2 error code.
// Both OpaqueClientResolver and CIMDClientResolver return this type so that
// HandleAuthorization can extract Code/Desc via errors.As without an import cycle.
type ClientIDError struct {
	Code string
	Desc string
}

func (e *ClientIDError) Error() string { return e.Code + ": " + e.Desc }

// SSRFBlockedError is returned when a CIMD fetch is rejected because the resolved
// IP falls in a blocked range (RFC 6890 or operator-configured). Defined at the
// ports layer so both the adapter (which creates it) and the domain service (which
// checks for it via errors.As) can reference it without import cycles.
type SSRFBlockedError struct {
	IP string
}

func (e *SSRFBlockedError) Error() string {
	return "SSRF protection: resolved IP " + e.IP + " is in a blocked range"
}
