// Package jwks provides the JWKS adapter for token exchange.
// This adapter abstracts HTTP fetching and caching of JSON Web Key Sets from upstream OAuth2 servers.
package jwks

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/lestrrat-go/httprc/v3"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Adapter implements the JWKSPort interface for fetching and caching JWKS from an upstream server.
// Each adapter instance maintains its own cached key set and refresh schedule.
//
// The adapter uses lestrrat-go/jwx/v3's jwk.Cache for automatic background refresh:
// - MinInterval: Minimum time between refresh attempts (prevents excessive refresh)
// - MaxInterval: Maximum time between refresh attempts (ensures periodic refresh)
// - HTTP client: For fetching JWKS from upstream server (injected into httprc.Client)
//
// Thread-safe: All methods are safe for concurrent use via internal synchronization by jwk.Cache.
type Adapter struct {
	cache        *jwk.Cache
	jwksURI      string
	httprcClient *httprc.Client
	mu           sync.RWMutex
}

// NewJWKSAdapter creates a new JWKS adapter that fetches and caches JWKs from the given JWKS URI.
//
// Parameters:
//   - jwksURI: The upstream OAuth2 server's JWKS endpoint URI (e.g., "https://auth.example.com/.well-known/jwks.json")
//   - httpClient: HTTP client for fetching JWKS (allows injection of timeouts, proxies, custom TLS, etc.)
//   - minRefreshInterval: Minimum time between refresh attempts (e.g., 15 minutes)
//     This prevents excessive refresh attempts if the server is unreachable.
//   - maxRefreshInterval: Maximum time between refresh attempts (e.g., 1 hour)
//     The cache will attempt to refresh JWKS at this interval to get updated keys.
//
// Returns an adapter ready to fetch and cache JWKS. Actual fetching is lazy (happens on first call).
//
// Error handling:
// - Validates jwksURI is not empty
// - Validates httpClient is not nil
// - Validates intervals are non-negative and minRefresh <= maxRefresh
// - Returns error if parameters are invalid
//
// Thread-safe for concurrent Get operations after creation.
func NewJWKSAdapter(
	jwksURI string,
	httpClient *http.Client,
	minRefreshInterval time.Duration,
	maxRefreshInterval time.Duration,
) (*Adapter, error) {
	// Validate parameters
	if jwksURI == "" {
		return nil, fmt.Errorf("jwks_uri cannot be empty")
	}
	if httpClient == nil {
		return nil, fmt.Errorf("http_client cannot be nil")
	}
	if minRefreshInterval < 0 {
		return nil, fmt.Errorf("min_refresh_interval cannot be negative")
	}
	if maxRefreshInterval < 0 {
		return nil, fmt.Errorf("max_refresh_interval cannot be negative")
	}
	if minRefreshInterval > maxRefreshInterval {
		return nil, fmt.Errorf("min_refresh_interval (%v) must be <= max_refresh_interval (%v)", minRefreshInterval, maxRefreshInterval)
	}

	// Create httprc client with custom HTTP client for upstream JWKS fetching
	httprcClient := httprc.NewClient()

	// Create new cache with httprc client
	cache, err := jwk.NewCache(context.Background(), httprcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwks cache: %w", err)
	}

	// Register the JWKS endpoint with refresh intervals and custom HTTP client
	// WithWaitReady(false) prevents blocking on first fetch; fetching happens lazily
	err = cache.Register(
		context.Background(),
		jwksURI,
		jwk.WithMinInterval(minRefreshInterval),
		jwk.WithMaxInterval(maxRefreshInterval),
		jwk.WithWaitReady(false),
		jwk.WithHttprcResourceOption(httprc.WithHTTPClient(httpClient)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register jwks_uri with cache: %w", err)
	}

	adapter := &Adapter{
		cache:        cache,
		jwksURI:      jwksURI,
		httprcClient: httprcClient,
	}

	return adapter, nil
}

// GetKeySet returns the cached JWKS key set from the upstream OAuth2 server.
//
// On first call, this will fetch the JWKS from the upstream server (may block briefly).
// Subsequent calls return cached results. The adapter handles automatic background refresh
// based on configured intervals.
//
// Returns the entire jwk.Set which can be searched for specific keys (by kid or other criteria).
//
// Error handling:
// - If first fetch fails: HTTP error (connection, timeout), JWKS format invalid, etc.
// - If context is cancelled: returns context cancellation error
// - If upstream returns non-200 status: returns HTTP error
//
// Thread-safe: Safe for concurrent calls (synchronized by jwk.Cache internally).
func (a *Adapter) GetKeySet(ctx context.Context) (jwk.Set, error) {
	ctx, span := otel.Tracer("jwks").Start(ctx, "jwks.fetch")
	defer span.End()
	span.SetAttributes(attribute.String("url.full", a.jwksURI))

	// Check if resource is ready. If not, force a refresh to populate the cache.
	// This handles the case where Register was called with WithWaitReady(false).
	if !a.cache.Ready(ctx, a.jwksURI) {
		_, err := a.cache.Refresh(ctx, a.jwksURI)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch jwks from %s: %w", a.jwksURI, err)
		}
	}

	// Fetch the JWKS from the upstream server (uses cache internally)
	// Lookup returns cached results immediately if available
	keyset, err := a.cache.Lookup(ctx, a.jwksURI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jwks from %s: %w", a.jwksURI, err)
	}

	return keyset, nil
}

// GetKey retrieves a specific key from the cached JWKS by key ID (kid).
//
// Preferred over iterating through GetKeySet() when only one key is needed.
// This method searches the key set for a key with the given kid and returns it.
//
// Parameters:
//   - ctx: Request context for cancellation support
//   - kid: The key ID to search for (typically from JWT header)
//
// Returns the jwk.Key if found, or an error if:
// - JWKS fetch fails
// - Key with given kid not found in key set
// - Context is cancelled
//
// Thread-safe: Safe for concurrent calls.
func (a *Adapter) GetKey(ctx context.Context, kid string) (jwk.Key, error) {
	keyset, err := a.GetKeySet(ctx)
	if err != nil {
		return nil, err
	}

	// Find key by kid
	key, found := keyset.LookupKeyID(kid)
	if !found {
		return nil, fmt.Errorf("key not found in jwks (kid: %s)", kid)
	}

	return key, nil
}

// Shutdown gracefully shuts down the adapter, stopping background refresh goroutines
// and closing any open connections. Should be called during application shutdown.
//
// Parameters:
//   - ctx: Context for cancellation support (typically a shutdown context with timeout)
//
// Returns error if shutdown fails.
//
// It is safe to call Shutdown multiple times.
func (a *Adapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cache == nil {
		return nil
	}

	return a.cache.Shutdown(ctx)
}
