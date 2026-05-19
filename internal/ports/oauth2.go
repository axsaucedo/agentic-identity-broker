// Package ports defines the ports (interfaces) for OAuth2 functionality.
package ports

import (
	"context"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/sessiontoken"
)

// MultiAgentVerifier verifies agent ID claims in proxied upstream token responses.
// When non-nil (feature enabled), VerifyAgentIDClaim is called after buffering the
// upstream response body. If verification fails, the response is withheld and an
// OAuth2 server_error is returned to the client (fail closed per SR-001).
// Nil means feature disabled — upstream response is passed through unchanged.
type MultiAgentVerifier interface {
	VerifyAgentIDClaim(ctx context.Context, responseBody []byte, expectedAgentID id.AgentID) error
}

// OAuth2Service defines the domain service interface for OAuth2 authorization,
// token proxying, and metadata discovery.
type OAuth2Service interface {
	// HandleAuthorization processes an OAuth2 authorization request, checking
	// client validity and user consent status, returning a decision (redirect URL or error).
	HandleAuthorization(ctx context.Context, req *AuthorizationRequest, principal id.Principal) (*AuthorizationDecision, error)

	// GenerateMetadata returns RFC 8414 OAuth2 metadata for auto-discovery.
	GenerateMetadata(ctx context.Context) (*MetadataResponse, error)
}

// AuthorizationRequest represents an OAuth2 authorization request (RFC 6749 Section 4.1.1).
// Fields are parsed from HTTP query parameters.
type AuthorizationRequest struct {
	// REQUIRED: OAuth2 client_id identifying the agent making the request
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
	// Action determines the response: "proceed", "redirect_to_consent", or "error".
	// "proceed" means the user has an active grant and the request can continue.
	// In proxy mode the handler redirects to the upstream OAuth2 server;
	// in issue_token mode the handler issues a local authorization code.
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

	// OPTIONAL: JWKS URI for public key discovery (present in issue_token mode)
	JWKSURI string `json:"jwks_uri,omitempty"`

	// OPTIONAL: Supported PKCE code challenge methods (present in issue_token mode)
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported,omitempty"`

	// OPTIONAL: Whether CIMD-based client_id resolution is supported (RFC draft)
	ClientIDMetadataDocumentSupported *bool `json:"client_id_metadata_document_supported,omitempty"`
}

// TokenMintingStrategy abstracts local token grant processing in issue_token mode.
// Grants are processed locally by the oauth2server.Provider.
// In proxy mode, grants are handled at the HTTP layer by proxyTokenGrantStrategy.
type TokenMintingStrategy interface {
	// HandleClientCredentials processes a client_credentials grant type request.
	// Returns the token response or an error.
	HandleClientCredentials(ctx context.Context, clientID id.ClientID, clientSecret, scope string) (*TokenResponse, error)

	// HandleAuthorizationCodeExchange processes an authorization_code grant type request.
	// Returns the token response or an error.
	HandleAuthorizationCodeExchange(ctx context.Context, clientID id.ClientID, clientSecret, code, redirectURI, codeVerifier string) (*TokenResponse, error)
}

// TokenResponse represents a successful OAuth2 token response from a minting strategy.
type TokenResponse struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	Scope       string
}

// AuthorizationCodeIssuer abstracts how the authorize endpoint issues authorization codes.
// In proxy mode, this is nil and the handler redirects to an upstream OAuth2 server.
// In issue_token mode, the endpoint issues authorization codes locally.
type AuthorizationCodeIssuer interface {
	// IssueAuthorizationCode processes a validated authorization request and returns
	// an authorization code. The handler is responsible for redirect_uri validation
	// and PKCE enforcement before calling this method.
	IssueAuthorizationCode(ctx context.Context, req *AuthorizationRequest, principal id.Principal) (code string, err error)
}

// SessionTokenValidator validates JWE authorization session tokens.
type SessionTokenValidator interface {
	ValidateAuthorizationSessionToken(token string, agentID id.AgentID, principal id.Principal) (*sessiontoken.AuthorizationSessionClaims, error)
}
