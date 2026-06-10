package jwks

import (
	"context"
	"errors"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockJWKSPublisher struct {
	set jwk.Set
	err error
}

func (m *mockJWKSPublisher) PublishJWKS(_ context.Context) (jwk.Set, error) {
	return m.set, m.err
}

func TestNewPublishedAdapter(t *testing.T) {
	t.Parallel()

	t.Run("nil publisher rejected", func(t *testing.T) {
		t.Parallel()

		adapter, err := NewPublishedAdapter(nil)

		require.Error(t, err)
		assert.Nil(t, adapter)
		assert.ErrorContains(t, err, "cannot be nil")
	})

	t.Run("valid publisher accepted", func(t *testing.T) {
		t.Parallel()

		adapter, err := NewPublishedAdapter(&mockJWKSPublisher{set: jwk.NewSet()})

		require.NoError(t, err)
		require.NotNil(t, adapter)
	})
}

func TestPublishedAdapter_GetKeySet(t *testing.T) {
	t.Parallel()

	expected := jwk.NewSet()
	adapter, err := NewPublishedAdapter(&mockJWKSPublisher{set: expected})
	require.NoError(t, err)

	got, err := adapter.GetKeySet(context.Background())

	require.NoError(t, err)
	assert.Same(t, expected, got)
}

func TestPublishedAdapter_GetKeySet_PropagatesPublisherError(t *testing.T) {
	t.Parallel()

	adapter, err := NewPublishedAdapter(&mockJWKSPublisher{err: errors.New("publisher unavailable")})
	require.NoError(t, err)

	_, err = adapter.GetKeySet(context.Background())

	require.Error(t, err)
	assert.ErrorContains(t, err, "publisher unavailable")
}

func TestPublishedAdapter_GetKey(t *testing.T) {
	t.Parallel()

	keySet := jwk.NewSet()
	symmetricKey, err := jwk.Import([]byte("0123456789abcdef0123456789abcdef"))
	require.NoError(t, err)
	require.NoError(t, symmetricKey.Set(jwk.KeyIDKey, "kid-1"))
	require.NoError(t, keySet.AddKey(symmetricKey))

	adapter, err := NewPublishedAdapter(&mockJWKSPublisher{set: keySet})
	require.NoError(t, err)

	got, err := adapter.GetKey(context.Background(), "kid-1")

	require.NoError(t, err)
	require.NotNil(t, got)
	kid, ok := got.KeyID()
	require.True(t, ok)
	assert.Equal(t, "kid-1", kid)
}

func TestPublishedAdapter_GetKey_NotFound(t *testing.T) {
	t.Parallel()

	adapter, err := NewPublishedAdapter(&mockJWKSPublisher{set: jwk.NewSet()})
	require.NoError(t, err)

	_, err = adapter.GetKey(context.Background(), "missing")

	require.Error(t, err)
	assert.ErrorContains(t, err, "key not found in published JWKS")
}
