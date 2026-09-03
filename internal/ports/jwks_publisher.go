package ports

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v4/jwk"
)

// JWKSPublisherPort sentinel errors — returned by JWKSPublisherPort implementations and
// handled by adapters that consume the port. Defined here so adapters do not need to import
// the concrete domain package to interpret publisher failures.
var (
	// ErrUpstreamUnavailable is returned when the broker cannot publish a complete
	// verification surface because upstream key material is unavailable.
	ErrUpstreamUnavailable = errors.New("upstream JWKS unavailable")

	// ErrKidConflict is returned when merging key sources would create an
	// ambiguous kid collision.
	ErrKidConflict = errors.New("duplicate kid across key sources")
)

// JWKSPublisherPort produces the aggregated JWKS for the broker's /oauth2/jwks.json endpoint.
// The implementation aggregates keys from mode-appropriate sources (local signing keys,
// upstream JWKS, or both) and enforces kid uniqueness invariants.
type JWKSPublisherPort interface {
	PublishJWKS(ctx context.Context) (jwk.Set, error)
}

// JWKSPublisherHealthPort exposes the publisher's upstream-JWKS health view for
// the end-user health endpoint. It intentionally matches JWKSHealthPort so the
// publisher can surface adapter freshness directly while still keeping the
// publisher and adapter roles distinct in the ports layer.
//
// Implementations return ComponentHealthHealthy while upstream key material is
// still servable from cache and ComponentHealthDegraded once it becomes
// unavailable or stale.
type JWKSPublisherHealthPort interface {
	HealthState() ComponentHealth
}
