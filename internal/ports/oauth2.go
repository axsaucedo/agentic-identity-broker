// Package ports defines the ports (interfaces) for OAuth2 functionality.
package ports

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
)

// OAuth2Service defines the domain service interface for OAuth2 authorization,
// token proxying, and metadata discovery.
type OAuth2Service interface {
	// HandleAuthorization processes an OAuth2 authorization request, checking
	// client validity and user consent status, returning a decision (redirect URL or error).
	HandleAuthorization(ctx context.Context, req *AuthorizationRequest, principal string) (*AuthorizationDecision, error)

	// GenerateMetadata returns RFC 8414 OAuth2 metadata for auto-discovery.
	GenerateMetadata(ctx context.Context) (*MetadataResponse, error)
}

// AuthorizationRequest represents an OAuth2 authorization request (RFC 6749 Section 4.1.1).
// Fields are parsed from HTTP query parameters.
type AuthorizationRequest struct {
	// REQUIRED: OAuth2 client identifier (maps to registered Agent.ClientID)
	ClientID id.ClientID

	// REQUIRED: Client's callback URL for authorization code
	RedirectURI string

	// OPTIONAL: Space-delimited requested scopes
	Scope string

	// RECOMMENDED: Opaque value for CSRF protection
	State string

	// REQUIRED: Must be "code" for authorization code flow
	ResponseType string

	// OPTIONAL: PKCE code challenge (RFC 7636)
	CodeChallenge string

	// OPTIONAL: PKCE method ("S256" or "plain")
	CodeChallengeMethod string

	// INTERNAL: Full original URL for redirect to consent UI
	OriginalURL string
}

// AuthorizationDecision represents the broker's decision for an authorization request.
// Either redirects to upstream or to consent UI, or returns an error.
type AuthorizationDecision struct {
	// Action determines the response: "redirect_to_upstream", "redirect_to_consent", or "error"
	Action string

	// RedirectURL is the target URL for HTTP 302 redirect
	RedirectURL string

	// ErrorCode is the OAuth2 error code (if Action == "error")
	ErrorCode string

	// ErrorDesc is the human-readable error description
	ErrorDesc string
}

// MetadataResponse represents OAuth2 Authorization Server Metadata (RFC 8414).
// Enables OAuth2 clients to auto-discover the broker's endpoints and capabilities.
type MetadataResponse struct {
	// REQUIRED: Issuer identifier (broker's public base URL)
	Issuer string `json:"issuer"`

	// REQUIRED: URL of the authorization endpoint
	AuthorizationEndpoint string `json:"authorization_endpoint"`

	// REQUIRED: URL of the token endpoint
	TokenEndpoint string `json:"token_endpoint"`

	// REQUIRED: Supported response types (at minimum ["code"])
	ResponseTypesSupported []string `json:"response_types_supported"`

	// REQUIRED: Supported grant types (at minimum ["authorization_code"])
	GrantTypesSupported []string `json:"grant_types_supported"`

	// OPTIONAL: Supported token endpoint authentication methods
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported,omitempty"`

	// OPTIONAL: Scopes supported by the broker
	ScopesSupported []string `json:"scopes_supported,omitempty"`

	// OPTIONAL: Claim types supported
	ClaimTypesSupported []string `json:"claim_types_supported,omitempty"`
}
