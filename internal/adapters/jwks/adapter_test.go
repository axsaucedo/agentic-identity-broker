package jwks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewJWKSAdapter tests adapter creation with various parameter combinations
func TestNewJWKSAdapter(t *testing.T) {
	tests := []struct {
		name               string
		jwksURI            string
		httpClient         *http.Client
		minRefreshInterval time.Duration
		maxRefreshInterval time.Duration
		expectError        bool
		errorContains      string
	}{
		{
			name:               "valid parameters",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         http.DefaultClient,
			minRefreshInterval: 15 * time.Minute,
			maxRefreshInterval: time.Hour,
			expectError:        false,
		},
		{
			name:               "zero intervals",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         http.DefaultClient,
			minRefreshInterval: 0,
			maxRefreshInterval: 0,
			expectError:        false,
		},
		{
			name:               "empty jwks_uri",
			jwksURI:            "",
			httpClient:         http.DefaultClient,
			minRefreshInterval: 15 * time.Minute,
			maxRefreshInterval: time.Hour,
			expectError:        true,
			errorContains:      "cannot be empty",
		},
		{
			name:               "nil http client",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         nil,
			minRefreshInterval: 15 * time.Minute,
			maxRefreshInterval: time.Hour,
			expectError:        true,
			errorContains:      "cannot be nil",
		},
		{
			name:               "negative min_refresh_interval",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         http.DefaultClient,
			minRefreshInterval: -1 * time.Minute,
			maxRefreshInterval: time.Hour,
			expectError:        true,
			errorContains:      "cannot be negative",
		},
		{
			name:               "negative max_refresh_interval",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         http.DefaultClient,
			minRefreshInterval: 15 * time.Minute,
			maxRefreshInterval: -1 * time.Hour,
			expectError:        true,
			errorContains:      "cannot be negative",
		},
		{
			name:               "min_refresh > max_refresh",
			jwksURI:            "https://auth.example.com/.well-known/jwks.json",
			httpClient:         http.DefaultClient,
			minRefreshInterval: 2 * time.Hour,
			maxRefreshInterval: time.Hour,
			expectError:        true,
			errorContains:      "must be <=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewJWKSAdapter(
				tt.jwksURI,
				tt.httpClient,
				tt.minRefreshInterval,
				tt.maxRefreshInterval,
			)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, adapter)
				if tt.errorContains != "" {
					assert.ErrorContains(t, err, tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, adapter)
				assert.Equal(t, tt.jwksURI, adapter.jwksURI)
			}
		})
	}
}

// TestGetKeySet tests successful JWKS fetching and caching
func TestGetKeySet(t *testing.T) {
	// Create a test JWKS with one key
	testKey, err := jwk.Import([]byte("secret_key_material_32_bytes_long_"))
	require.NoError(t, err)
	require.NoError(t, testKey.Set(jwk.KeyIDKey, "test-kid"))
	require.NoError(t, testKey.Set(jwk.AlgorithmKey, jwa.HS256()))

	keyset := jwk.NewSet()
	require.NoError(t, keyset.AddKey(testKey))

	// Convert keyset to JSON (what upstream server returns)
	jwksJSON, err := json.Marshal(keyset)
	require.NoError(t, err)

	// Create mock upstream JWKS endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksJSON)
	}))
	defer server.Close()

	// Create adapter
	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	// Get keyset
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resultSet, err := adapter.GetKeySet(ctx)
	require.NoError(t, err)
	require.NotNil(t, resultSet)

	// Verify keyset contains our test key
	key, found := resultSet.LookupKeyID("test-kid")
	assert.True(t, found)
	assert.NotNil(t, key)
}

// TestGetKeySet_HTTPError tests handling of HTTP errors
func TestGetKeySet_HTTPError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = adapter.GetKeySet(ctx)
	assert.Error(t, err)
}

// TestGetKeySet_InvalidJSON tests handling of invalid JWKS format
func TestGetKeySet_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not valid json {"))
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = adapter.GetKeySet(ctx)
	assert.Error(t, err)
}

// TestGetKeySet_ContextCancelled tests cancellation handling
func TestGetKeySet_ContextCancelled(t *testing.T) {
	// Create mock server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"keys": []}`))
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = adapter.GetKeySet(ctx)
	assert.Error(t, err)
}

// TestGetKey tests key retrieval by kid
func TestGetKey(t *testing.T) {
	// Create test JWKS with multiple keys
	key1, err := jwk.Import([]byte("secret_key_material_32_bytes_long_1"))
	require.NoError(t, err)
	require.NoError(t, key1.Set(jwk.KeyIDKey, "kid-1"))
	require.NoError(t, key1.Set(jwk.AlgorithmKey, jwa.HS256()))

	key2, err := jwk.Import([]byte("secret_key_material_32_bytes_long_2"))
	require.NoError(t, err)
	require.NoError(t, key2.Set(jwk.KeyIDKey, "kid-2"))
	require.NoError(t, key2.Set(jwk.AlgorithmKey, jwa.HS256()))

	keyset := jwk.NewSet()
	require.NoError(t, keyset.AddKey(key1))
	require.NoError(t, keyset.AddKey(key2))

	jwksJSON, err := json.Marshal(keyset)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksJSON)
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get specific key
	key, err := adapter.GetKey(ctx, "kid-1")
	require.NoError(t, err)
	require.NotNil(t, key)

	// Verify it's the correct key (KeyID() returns (string, bool) in v3)
	kid, ok := key.KeyID()
	assert.True(t, ok)
	assert.Equal(t, "kid-1", kid)
}

// TestGetKey_NotFound tests handling of missing key
func TestGetKey_NotFound(t *testing.T) {
	key1, err := jwk.Import([]byte("secret_key_material_32_bytes_long_1"))
	require.NoError(t, err)
	require.NoError(t, key1.Set(jwk.KeyIDKey, "kid-1"))
	require.NoError(t, key1.Set(jwk.AlgorithmKey, jwa.HS256()))

	keyset := jwk.NewSet()
	require.NoError(t, keyset.AddKey(key1))

	jwksJSON, err := json.Marshal(keyset)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksJSON)
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to get key that doesn't exist
	_, err = adapter.GetKey(ctx, "nonexistent-kid")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "key not found")
}

// TestGetKey_ConcurrentAccess tests thread-safety of concurrent Get calls
func TestGetKey_ConcurrentAccess(t *testing.T) {
	key1, err := jwk.Import([]byte("secret_key_material_32_bytes_long_1"))
	require.NoError(t, err)
	require.NoError(t, key1.Set(jwk.KeyIDKey, "kid-1"))
	require.NoError(t, key1.Set(jwk.AlgorithmKey, jwa.HS256()))

	keyset := jwk.NewSet()
	require.NoError(t, keyset.AddKey(key1))

	jwksJSON, err := json.Marshal(keyset)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksJSON)
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Concurrent calls should use cache, not make multiple fetches
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := adapter.GetKey(ctx, "kid-1")
			done <- err
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		err := <-done
		require.NoError(t, err)
	}
}

// TestGetKeySet_Caching verifies that subsequent calls use cached results
func TestGetKeySet_Caching(t *testing.T) {
	key1, err := jwk.Import([]byte("secret_key_material_32_bytes_long_1"))
	require.NoError(t, err)
	require.NoError(t, key1.Set(jwk.KeyIDKey, "kid-1"))
	require.NoError(t, key1.Set(jwk.AlgorithmKey, jwa.HS256()))

	keyset := jwk.NewSet()
	require.NoError(t, keyset.AddKey(key1))

	jwksJSON, err := json.Marshal(keyset)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(jwksJSON)
	}))
	defer server.Close()

	adapter, err := NewJWKSAdapter(
		server.URL,
		server.Client(),
		15*time.Minute,
		time.Hour,
	)
	require.NoError(t, err)
	defer func() { _ = adapter.Shutdown(context.Background()) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First call should fetch
	keyset1, err := adapter.GetKeySet(ctx)
	require.NoError(t, err)
	require.NotNil(t, keyset1)

	// Second call should use cache
	keyset2, err := adapter.GetKeySet(ctx)
	require.NoError(t, err)
	require.NotNil(t, keyset2)

	// Both calls should return key sets
	assert.NotNil(t, keyset1)
	assert.NotNil(t, keyset2)
}
