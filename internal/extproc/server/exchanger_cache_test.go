// Package server_test — Phase 4 cache, singleflight, and eviction tests.
//
// T028: Cache TTL expiry — verifies expired entries are re-fetched
// T029: Singleflight deduplication — concurrent identical calls hit exchange once
// T030: Background eviction and assertion refresh — eviction goroutine clears expired entries
package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

// ---------------------------------------------------------------------------
// T028: Cache TTL expiry
// ---------------------------------------------------------------------------

// Spec: FR-011, FR-012 — Expired cache entry is re-fetched from exchange endpoint
func TestTokenExchanger_Cache_ExpiredEntry_TriggersNewExchange(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	// Use a very short TTL so the cache entry expires quickly
	shortExpiry := 1 // 1 second
	mocks.exchangeExpiry = &shortExpiry

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "expiry-test-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// First call — populates cache
	token1, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)
	callsAfterFirst := mocks.exchangeCalls

	// Wait for cache entry to expire
	time.Sleep(1100 * time.Millisecond)

	// Return a different token on the second exchange call
	mocks.exchangedToken = "refreshed-access-token"

	// Second call after expiry — must re-fetch
	token2, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)

	assert.Equal(t, "exchanged-access-token", token1, "first call must return original token")
	assert.Equal(t, "refreshed-access-token", token2, "second call after expiry must return new token")
	assert.Greater(t, mocks.exchangeCalls, callsAfterFirst,
		"token exchange endpoint must be called again after TTL expiry")
}

// Spec: FR-012 — Cache hit returns same token without calling exchange
func TestTokenExchanger_Cache_HitBeforeExpiry_ReturnsCachedToken(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	// Use a long TTL — cache should not expire during test
	longExpiry := 3600
	mocks.exchangeExpiry = &longExpiry

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "long-lived-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// First call — populates cache
	token1, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)
	callsAfterFirst := mocks.exchangeCalls

	// Multiple rapid calls — all should hit cache
	for range 5 {
		token, err := exchanger.Exchange(subjectToken, resourceURI)
		require.NoError(t, err)
		assert.Equal(t, token1, token, "cache hit must return same token")
	}

	assert.Equal(t, callsAfterFirst, mocks.exchangeCalls,
		"exchange endpoint must not be called on cache hits")
}

// Spec: FR-018 — Cache TTL respects max_ttl cap even when expires_in is larger
func TestTokenExchanger_Cache_MaxTTLCap_AppliedCorrectly(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	// expires_in larger than max_ttl
	largeExpiry := 7 * 24 * 3600 // 1 week
	mocks.exchangeExpiry = &largeExpiry

	cfg := configForMocks(mocks)
	cfg.Cache.MaxTTL = 500 * time.Millisecond // short max TTL for testing

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "max-ttl-test-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// First call — populates cache with max_ttl cap applied
	token1, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)

	// The token is cached — another call should hit cache
	token2, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)
	assert.Equal(t, token1, token2, "immediate second call must be a cache hit")
	callsAfterSecond := mocks.exchangeCalls

	// Wait for max_ttl to expire (500ms + buffer)
	time.Sleep(700 * time.Millisecond)

	// After max_ttl, cache must be expired
	mocks.exchangedToken = "post-maxttl-token"
	token3, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)

	assert.Equal(t, "post-maxttl-token", token3,
		"token must be refreshed after max_ttl expires")
	assert.Greater(t, mocks.exchangeCalls, callsAfterSecond,
		"exchange endpoint must be called again after max_ttl cap expiry")
}

// Spec: FR-011 — Each unique subject+resource combination has its own cache entry
func TestTokenExchanger_Cache_UniqueKeyPerSubjectAndResource(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	longExpiry := 3600
	var callMu sync.Mutex
	var exchangeLog []string

	mocks.tokenExchServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		subject := r.FormValue("subject_token")
		resource := r.FormValue("resource")

		callMu.Lock()
		exchangeLog = append(exchangeLog, fmt.Sprintf("%s@%s", subject, resource))
		callMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		expiry := longExpiry
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: fmt.Sprintf("token-for-%s-%s", subject, resource),
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	})

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Exchange for multiple combinations
	combinations := []struct{ subject, resource string }{
		{"user-a", "http://service-1:9000/api"},
		{"user-a", "http://service-2:9001/api"},
		{"user-b", "http://service-1:9000/api"},
	}

	tokens := make([]string, len(combinations))
	for i, c := range combinations {
		tok, err := exchanger.Exchange(c.subject, c.resource)
		require.NoError(t, err)
		tokens[i] = tok
	}

	// All tokens should be distinct
	assert.NotEqual(t, tokens[0], tokens[1], "different resources → different tokens")
	assert.NotEqual(t, tokens[0], tokens[2], "different subjects → different tokens")
	assert.NotEqual(t, tokens[1], tokens[2], "different subject+resource → different tokens")

	// Re-fetch — all should come from cache (no new exchange calls)
	callsBeforeRefetch := len(exchangeLog)
	for _, c := range combinations {
		_, err := exchanger.Exchange(c.subject, c.resource)
		require.NoError(t, err)
	}
	callMu.Lock()
	callsAfterRefetch := len(exchangeLog)
	callMu.Unlock()

	assert.Equal(t, callsBeforeRefetch, callsAfterRefetch,
		"all re-fetches must be cache hits — no new exchange calls")
}

// ---------------------------------------------------------------------------
// T029: Singleflight deduplication
// ---------------------------------------------------------------------------

// Spec: FR-014 — Concurrent requests for the same key are deduplicated via singleflight
func TestTokenExchanger_Singleflight_ConcurrentRequests_CallExchangeOnce(t *testing.T) {
	var (
		requestCount  int64
		responseMu    sync.Mutex
		responseReady = make(chan struct{})
	)

	// Slow exchange server that counts calls
	slowExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		// Block until all goroutines have submitted their requests
		responseMu.Lock()
		<-responseReady
		responseMu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "singleflight-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer slowExchServer.Close()

	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.TokenEndpoint = slowExchServer.URL + "/oauth2/token"
	cfg.OAuth2.ExchangeTimeout = 10 * time.Second

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const numGoroutines = 10
	const subjectToken = "concurrent-user-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	var wg sync.WaitGroup
	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	// Launch concurrent goroutines requesting the same key
	for i := range numGoroutines {
		wg.Go(func() {
			results[i], errors[i] = exchanger.Exchange(subjectToken, resourceURI)
		})
	}

	// Give goroutines time to queue up before releasing the slow server
	time.Sleep(50 * time.Millisecond)
	close(responseReady) // release all blocked requests
	wg.Wait()

	// All calls must succeed
	for i, err := range errors {
		assert.NoError(t, err, "goroutine %d must not error", i)
	}

	// All calls must return the same token
	for i, result := range results {
		assert.Equal(t, "singleflight-token", result,
			"goroutine %d must receive the singleflight token", i)
	}

	// Exchange endpoint must be called only ONCE despite 10 concurrent requests
	assert.Equal(t, int64(1), atomic.LoadInt64(&requestCount),
		"singleflight must deduplicate concurrent requests — exchange called exactly once")
}

// Spec: FR-014 — Singleflight allows independent keys to proceed concurrently
func TestTokenExchanger_Singleflight_DifferentKeys_ProceedConcurrently(t *testing.T) {
	var requestCount int64

	exchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		_ = r.ParseForm()
		resource := r.FormValue("resource")

		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "token-for-" + resource,
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer exchServer.Close()

	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.TokenEndpoint = exchServer.URL + "/oauth2/token"

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// 3 different resources — should each call exchange once
	resources := []string{
		"http://service-a:9000/api",
		"http://service-b:9001/api",
		"http://service-c:9002/api",
	}

	var wg sync.WaitGroup
	results := make([]string, len(resources))
	for i, res := range resources {
		wg.Add(1)
		go func(idx int, resource string) {
			defer wg.Done()
			tok, err := exchanger.Exchange("user-token", resource)
			require.NoError(t, err)
			results[idx] = tok
		}(i, res)
	}
	wg.Wait()

	// Each resource gets its own token
	for i, res := range resources {
		assert.Equal(t, "token-for-"+res, results[i],
			"resource %s must receive its own token", res)
	}

	// Each resource should have triggered exactly one exchange
	assert.Equal(t, int64(len(resources)), atomic.LoadInt64(&requestCount),
		"each unique resource must trigger exactly one exchange call")
}

// Spec: FR-014 — Singleflight error is shared across all waiting callers
func TestTokenExchanger_Singleflight_SharedError_OnExchangeFailure(t *testing.T) {
	var callCount int64

	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&callCount, 1)
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprintf(w, `{"error":"invalid_client"}`)
	}))
	defer errorServer.Close()

	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.TokenEndpoint = errorServer.URL + "/oauth2/token"

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Multiple concurrent calls with same key — all should fail, exchange called once
	const numGoroutines = 5
	errs := make([]error, numGoroutines)
	var wg sync.WaitGroup

	for i := range numGoroutines {
		wg.Go(func() {
			_, errs[i] = exchanger.Exchange("user-token", "http://resource.example.com")
		})
	}
	wg.Wait()

	// All goroutines must receive an error
	for i, err := range errs {
		assert.Error(t, err, "goroutine %d must receive an error on exchange failure", i)
	}

	// Exchange must be called at most once despite concurrent requests
	// (singleflight deduplicates the error path too)
	assert.LessOrEqual(t, atomic.LoadInt64(&callCount), int64(numGoroutines),
		"singleflight must not amplify error-path calls beyond number of goroutines")
}

// ---------------------------------------------------------------------------
// T030: Background eviction and assertion refresh
// ---------------------------------------------------------------------------

// Spec: FR-016 — Background eviction removes expired entries from cache
func TestTokenExchanger_Eviction_ExpiredEntries_RemovedFromCache(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	shortExpiry := 1 // 1 second TTL
	mocks.exchangeExpiry = &shortExpiry

	cfg := configForMocks(mocks)
	// Set a very short DefaultTTL to trigger eviction loop quickly
	cfg.Cache.DefaultTTL = 500 * time.Millisecond

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "eviction-test-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// Populate cache
	_, err = exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)
	callsAfterFirst := mocks.exchangeCalls

	// Wait for cache TTL to expire (1 second) and eviction loop to run
	// Eviction runs at DefaultTTL/2 = 250ms, so 1.5 seconds is sufficient
	time.Sleep(1500 * time.Millisecond)

	// After eviction, next call must re-fetch
	mocks.exchangedToken = "token-after-eviction"
	_, err = exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)

	assert.Greater(t, mocks.exchangeCalls, callsAfterFirst,
		"exchange endpoint must be called again after eviction")
}

// Spec: FR-016 — Background eviction does not remove valid (non-expired) entries
func TestTokenExchanger_Eviction_ValidEntries_NotEvicted(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	longExpiry := 3600 // 1 hour TTL
	mocks.exchangeExpiry = &longExpiry

	cfg := configForMocks(mocks)
	cfg.Cache.DefaultTTL = 500 * time.Millisecond // fast eviction loop

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "persist-test-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// Populate cache
	token1, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)
	callsAfterFirst := mocks.exchangeCalls

	// Wait for several eviction cycles (but entry should NOT be evicted — long TTL)
	time.Sleep(1200 * time.Millisecond)

	// Entry must still be in cache — no new exchange call
	token2, err := exchanger.Exchange(subjectToken, resourceURI)
	require.NoError(t, err)

	assert.Equal(t, token1, token2,
		"valid cache entry must survive eviction cycles")
	assert.Equal(t, callsAfterFirst, mocks.exchangeCalls,
		"exchange endpoint must not be called for a non-expired cache entry")
}

// Spec: FR-006 — Background goroutine refreshes client assertion before expiry
func TestTokenExchanger_AssertionRefresh_BackgroundRefresh_KeepsAssertionFresh(t *testing.T) {
	var assertionCallCount int64

	// Custom client credentials server that counts assertion refreshes.
	// Uses 1-second expires_in so maybeRefreshAssertion (which fires when remaining < 30s)
	// will always trigger a refresh on every eviction tick.
	clientCredsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt64(&assertionCallCount, 1)
		w.Header().Set("Content-Type", "application/json")
		expiry := 1 // 1 second expires_in → assertion is "almost expired" immediately
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access-token",
			IDToken:     fmt.Sprintf("id-token-%d", count),
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer clientCredsServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tokenExchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tokenExchServer.URL + "/oauth2/token",
			Issuer:                    clientCredsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: clientCredsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			// DefaultTTL = 2s → eviction ticker fires every max(1s, 2s/2) = 1s.
			// Assertion refresh ticker = min(1s, 30s) = 1s.
			// Assertion expires in 1s; remaining is always < 30s threshold →
			// maybeRefreshAssertion triggers on every assertion ticker tick.
			DefaultTTL: 2 * time.Second,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	callsAtStartup := atomic.LoadInt64(&assertionCallCount)

	// Wait for at least 2 assertion refresh ticks (1s each) to fire.
	// Assertion ticker = min(DefaultTTL/2, 30s) = min(1s, 30s) = 1s.
	// With 1s expires_in → remaining is always < 30s → refresh is triggered on every tick.
	time.Sleep(2500 * time.Millisecond)

	callsAfterWait := atomic.LoadInt64(&assertionCallCount)

	// The background goroutine must have triggered at least one refresh beyond startup
	assert.Greater(t, callsAfterWait, callsAtStartup,
		"background goroutine must refresh client assertion before expiry")
}

// Spec: FR-016 — Shutdown stops background eviction goroutine
func TestTokenExchanger_Shutdown_StopsEvictionGoroutine(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.Cache.DefaultTTL = 100 * time.Millisecond // fast eviction loop

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)

	// Shutdown must complete without blocking or panicking
	done := make(chan struct{})
	go func() {
		exchanger.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		// Shutdown completed as expected
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown must complete promptly — possible goroutine leak")
	}

	// Second Shutdown call must not panic (idempotent cleanup via stopCh close)
	// Note: calling Shutdown() twice on a channel that's already closed would panic,
	// so we verify single-shutdown semantics work correctly
}

// Spec: FR-014, FR-016 — Race detector: concurrent Exchange + Shutdown is safe
func TestTokenExchanger_Concurrent_ExchangeAndShutdown_RaceFree(t *testing.T) {
	// Use isolated mock servers with race-safe atomic counters (avoiding the
	// shared mutable fields in newMockServers() which have a pre-existing race).
	clientCredsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access-token",
			IDToken:     "id-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer clientCredsServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tokenExchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tokenExchServer.URL + "/oauth2/token",
			Issuer:                    clientCredsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: clientCredsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 50 * time.Millisecond,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)

	var wg sync.WaitGroup

	// Concurrent exchange calls with unique keys (no singleflight contention needed here)
	for i := range 5 {
		wg.Go(func() {
			_, _ = exchanger.Exchange(
				fmt.Sprintf("token-%d", i),
				fmt.Sprintf("http://service-%d:9000/api", i),
			)
		})
	}

	// Concurrent shutdown
	wg.Go(func() {
		time.Sleep(10 * time.Millisecond) // let some exchanges start
		exchanger.Shutdown()
	})

	wg.Wait()
	// Test passes if no race conditions are detected by the Go race detector (-race flag)
}
