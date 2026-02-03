// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

// JWKSPort defines the interface for JSON Web Key Set (JWKS) operations.
// This port abstracts HTTP JWKS fetching and caching, allowing domain logic
// to remain independent of infrastructure concerns.
//
// Implementation: internal/adapters/jwks/adapter.go (handles HTTP fetching with caching)
//
// The adapter is responsible for:
// - Fetching JWKS from upstream OAuth2 server's jwks_uri
// - Caching the key set with automatic refresh
// - Handling HTTP errors and retries
// - Validating key set structure
type JWKSPort interface {
	// GetKeySet returns the cached JWKS key set.
	// The adapter handles background refresh based on configured intervals.
	// May block briefly if first fetch is needed.
	// Returns error if:
	// - HTTP fetch fails (connection error, timeout)
	// - Upstream server returns invalid JWKS format
	// - Context is cancelled
	GetKeySet(ctx context.Context) (jwk.Set, error)

	// GetKey retrieves a specific key from the cached JWKS by key ID (kid).
	// Preferred over iterating GetKeySet() when only one key is needed.
	// Returns error if:
	// - Key with given kid not found (no wrapped error, plain message)
	// - HTTP fetch fails or JWKS is invalid
	// - Context is cancelled
	// Note: Implementing adapters MUST cache results to avoid repeated HTTP fetches.
	GetKey(ctx context.Context, kid string) (jwk.Key, error)
}
