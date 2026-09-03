// Package ports defines interfaces for hexagonal architecture boundaries.
package ports

import (
	"context"

	"github.com/lestrrat-go/jwx/v4/jwk"
)

// JWKSHealthPort exposes operational health information for a network-backed
// JWKS adapter. It intentionally shares the HealthState() shape used by the
// publisher-facing health facet so a publisher can delegate upstream freshness
// without inventing translation glue at the domain boundary.
//
// Implementations should report degraded only when the cached upstream material
// is unavailable or has gone stale without a successful refresh.
type JWKSHealthPort interface {
	HealthState() ComponentHealth
}

// JWKSPort defines the interface for JSON Web Key Set (JWKS) operations.
// It combines key retrieval with operational health reporting for the broker's
// network-backed JWKS sources.
//
// Implementation: internal/adapters/jwks/adapter.go (handles HTTP fetching,
// caching, and refresh health)
//
// The adapter is responsible for:
// - Fetching JWKS from upstream OAuth2 server's jwks_uri
// - Caching the key set with automatic refresh
// - Reporting whether cached upstream key material is still servable
// - Handling HTTP errors and retries
// - Validating key set structure
type JWKSPort interface {
	JWKSHealthPort

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
