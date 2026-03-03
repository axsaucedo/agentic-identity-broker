// Package fixtures provides test tokens, configurations, and mock response data
// for ExtProc Token Exchange E2E tests.
//
// All tokens are static test values. For JWT-signed tokens (if needed for specific tests),
// use helpers.SignTestJWT with a generated key pair.
package fixtures

import "time"

// TestBearerTokens contains Bearer tokens for use in test scenarios.
// These are opaque tokens — the ExtProc service does not parse them,
// it passes them as subject_token to the identity broker.
const (
	// ValidBearerToken is a valid opaque Bearer token for happy path tests.
	ValidBearerToken = "test-bearer-token-subject-abc123"

	// ValidBearerToken2 is a second valid Bearer token for cache key differentiation tests.
	ValidBearerToken2 = "test-bearer-token-subject-def456"

	// AlternativeBearerToken is used to test cache miss scenarios with a different token.
	AlternativeBearerToken = "test-bearer-token-alternative-xyz789"

	// ExpiredBearerToken is a Bearer token associated with an expired session in mock responses.
	ExpiredBearerToken = "test-bearer-token-expired-999"
)

// TestResourceURIs contains resource URI values for use in test scenarios.
// These are passed as the resource parameter in RFC 8693 token exchange requests.
// Per FR-004, the :path pseudo-header must be an absolute URI with http/https scheme.
const (
	// ValidResourceURI is a valid absolute URI for happy path tests.
	ValidResourceURI = "http://mcp-server:9003/mcp"

	// ValidResourceURIHTTPS is a valid HTTPS absolute URI.
	ValidResourceURIHTTPS = "https://mcp-server.example.com/mcp"

	// AlternativeResourceURI is a second valid resource URI for cache differentiation tests.
	AlternativeResourceURI = "http://another-mcp-server:9004/tools"

	// EmptyResourceURI represents an empty path (should trigger 503 per FR-013).
	EmptyResourceURI = ""

	// RelativePathResourceURI is a relative path that should trigger 503 per FR-013.
	RelativePathResourceURI = "/mcp"

	// InvalidSchemeResourceURI has an unsupported scheme (should trigger 503 per FR-013).
	InvalidSchemeResourceURI = "ftp://example.com/resource"
)

// TestExchangedTokens contains the exchanged tokens returned by the mock identity broker.
const (
	// DefaultExchangedToken is returned by default mock token exchange responses.
	DefaultExchangedToken = "exchanged-access-token-default"

	// FreshExchangedToken is returned when a cache miss triggers a new exchange.
	FreshExchangedToken = "exchanged-access-token-fresh-001"

	// CachedExchangedToken is the token expected from a cache hit.
	CachedExchangedToken = "exchanged-access-token-cached-002"

	// RefreshedExchangedToken is returned after a cache expiry refresh.
	RefreshedExchangedToken = "exchanged-access-token-refreshed-003"
)

// TestClientAssertionTokens are ID tokens returned by the mock OAuth2 server
// and used as client assertions in token exchange requests.
const (
	// DefaultClientAssertionIDToken is the default mock ID token for client assertions.
	DefaultClientAssertionIDToken = "mock-id-token-for-client-assertion"
)

// TTLValues contains time duration values for cache testing.
var (
	// ShortTTL is a very short TTL used to test cache expiry scenarios.
	// Tests that use this TTL should wait slightly longer than this value.
	ShortTTL = 100 * time.Millisecond

	// DefaultTTL matches the default cache TTL in test config.
	DefaultTTL = 5 * time.Minute
)

// StandardExpiresIn is the standard expires_in value (in seconds) for mock responses.
const StandardExpiresIn = 3600
