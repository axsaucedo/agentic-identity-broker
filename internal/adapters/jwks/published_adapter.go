package jwks

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwk"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// PublishedAdapter exposes a broker-published JWKS as a JWKSProvider-compatible source.
// It is used when a consumer should trust the same aggregated key surface the broker
// publishes at /oauth2/jwks.json, without making an in-process HTTP round-trip.
type PublishedAdapter struct {
	publisher ports.JWKSPublisherPort
}

// NewPublishedAdapter creates an adapter backed by a JWKS publisher.
func NewPublishedAdapter(publisher ports.JWKSPublisherPort) (*PublishedAdapter, error) {
	if publisher == nil {
		return nil, fmt.Errorf("jwks publisher cannot be nil")
	}

	return &PublishedAdapter{publisher: publisher}, nil
}

// GetKeySet returns the published JWKS snapshot.
func (a *PublishedAdapter) GetKeySet(ctx context.Context) (jwk.Set, error) {
	return a.publisher.PublishJWKS(ctx)
}

// GetKey returns a key by kid from the published JWKS snapshot.
func (a *PublishedAdapter) GetKey(ctx context.Context, kid string) (jwk.Key, error) {
	set, err := a.GetKeySet(ctx)
	if err != nil {
		return nil, err
	}

	key, ok := set.LookupKeyID(kid)
	if !ok {
		return nil, fmt.Errorf("key not found in published JWKS (kid: %s)", kid)
	}

	return key, nil
}
