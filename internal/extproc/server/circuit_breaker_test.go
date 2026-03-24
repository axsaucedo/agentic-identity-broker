// Package server_test — circuit_breaker_test.go covers the circuit breaker
// protecting token exchange calls to the identity broker.
//
// Tests verify:
//   - Circuit opens after max_failures consecutive errors
//   - Open circuit rejects calls immediately with ErrCircuitOpen
//   - Half-open state allows a single probe after reset_timeout
//   - Successful probe closes the circuit
//   - Failed probe re-opens the circuit
//   - Cache hits bypass the circuit breaker (no backend call needed)
//   - Circuit breaker is thread-safe under concurrent access
package server_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	httpv3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

// eventuallyTimeout is the extra buffer added to resetTimeout when waiting for
// circuit state transitions under test. 500ms gives ample slack for CI scheduling
// jitter without making tests noticeably slow.
const eventuallyTimeout = 500 * time.Millisecond

// eventuallyPollInterval is the polling cadence for require.Eventually calls that
// probe circuit breaker state transitions. 20ms balances test speed with CPU waste.
const eventuallyPollInterval = 20 * time.Millisecond

// circuitBreakerConfig returns a config with a low threshold for fast circuit breaker testing.
func circuitBreakerConfig(m *mockServers, maxFailures int, resetTimeout time.Duration) *extprocconfig.Config {
	cfg := configForMocks(m)
	cfg.CircuitBreaker = extprocconfig.CircuitBreakerConfig{
		MaxFailures:  maxFailures,
		ResetTimeout: resetTimeout,
	}
	return cfg
}

// ---------------------------------------------------------------------------
// Circuit breaker: open after max_failures
// ---------------------------------------------------------------------------

// Circuit opens after max_failures consecutive exchange errors.
func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"

	const maxFailures = 3
	cfg := circuitBreakerConfig(mocks, maxFailures, 30*time.Second)

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Trigger max_failures consecutive errors (each with a unique key to avoid singleflight caching errors)
	for i := 0; i < maxFailures; i++ {
		_, err := exchanger.Exchange(fmt.Sprintf("token-%d", i), fmt.Sprintf("http://resource.example.com/%d", i))
		assert.Error(t, err, "exchange %d should fail", i)
		assert.NotErrorIs(t, err, server.ErrCircuitOpen,
			"exchange %d should fail with backend error, not circuit open", i)
	}

	// Next call should be rejected by the circuit breaker
	_, err = exchanger.Exchange("token-after-trip", "http://resource.example.com/after")
	assert.ErrorIs(t, err, server.ErrCircuitOpen,
		"circuit must be open after %d consecutive failures", maxFailures)
}

// ---------------------------------------------------------------------------
// Circuit breaker: fast-fail when open
// ---------------------------------------------------------------------------

// Open circuit rejects calls immediately without hitting the backend.
func TestCircuitBreaker_OpenCircuit_RejectsImmediately(t *testing.T) {
	var backendCalls int64

	exchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&backendCalls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, `{"error":"server_error"}`)
	}))
	defer exchServer.Close()

	mocks := newMockServers()
	defer mocks.Close()

	const maxFailures = 2
	cfg := circuitBreakerConfig(mocks, maxFailures, 1*time.Minute)
	cfg.OAuth2.TokenEndpoint = exchServer.URL + "/oauth2/token"

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Trip the circuit
	for i := 0; i < maxFailures; i++ {
		_, _ = exchanger.Exchange(fmt.Sprintf("token-%d", i), fmt.Sprintf("http://resource.example.com/%d", i))
	}
	callsAfterTrip := atomic.LoadInt64(&backendCalls)

	// Additional calls must NOT hit the backend
	for i := 0; i < 5; i++ {
		_, err := exchanger.Exchange(fmt.Sprintf("token-open-%d", i), fmt.Sprintf("http://resource.example.com/open-%d", i))
		assert.ErrorIs(t, err, server.ErrCircuitOpen)
	}

	assert.Equal(t, callsAfterTrip, atomic.LoadInt64(&backendCalls),
		"no backend calls should be made while circuit is open")
}

// ---------------------------------------------------------------------------
// Circuit breaker: half-open → probe succeeds → closed
// ---------------------------------------------------------------------------

// After reset_timeout, a successful probe closes the circuit.
func TestCircuitBreaker_HalfOpen_SuccessfulProbe_ClosesCircuit(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	const maxFailures = 2
	const resetTimeout = 200 * time.Millisecond
	cfg := circuitBreakerConfig(mocks, maxFailures, resetTimeout)

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Trip the circuit with failures
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"
	for i := 0; i < maxFailures; i++ {
		_, _ = exchanger.Exchange(fmt.Sprintf("token-%d", i), fmt.Sprintf("http://resource.example.com/%d", i))
	}

	// Verify circuit is open
	_, err = exchanger.Exchange("token-check", "http://resource.example.com/check")
	assert.ErrorIs(t, err, server.ErrCircuitOpen, "circuit must be open")

	// Restore backend and wait until the probe succeeds (circuit moves through half-open → closed).
	// Using Eventually avoids a fixed sleep that would be flaky under CI load.
	mocks.exchangeStatus = http.StatusOK
	mocks.exchangedToken = "recovered-token"

	var token string
	require.Eventually(
		t,
		func() bool {
			tok, err := exchanger.Exchange("token-probe", "http://resource.example.com/probe")
			if err != nil {
				// Circuit may still be open; keep retrying until reset_timeout expires.
				return false
			}
			token = tok
			return token == "recovered-token"
		},
		resetTimeout+eventuallyTimeout,
		eventuallyPollInterval,
		"half-open probe should succeed with recovered backend",
	)

	// Subsequent calls should work normally (circuit is closed)
	token2, err := exchanger.Exchange("token-normal", "http://resource.example.com/normal")
	require.NoError(t, err, "circuit should be closed after successful probe")
	assert.Equal(t, "recovered-token", token2)
}

// ---------------------------------------------------------------------------
// Circuit breaker: half-open → probe fails → re-opens
// ---------------------------------------------------------------------------

// After reset_timeout, a failed probe re-opens the circuit.
func TestCircuitBreaker_HalfOpen_FailedProbe_ReopensCircuit(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	const maxFailures = 2
	const resetTimeout = 200 * time.Millisecond
	cfg := circuitBreakerConfig(mocks, maxFailures, resetTimeout)

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Trip the circuit
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"
	for i := 0; i < maxFailures; i++ {
		_, _ = exchanger.Exchange(fmt.Sprintf("token-%d", i), fmt.Sprintf("http://resource.example.com/%d", i))
	}

	// Wait until a probe is allowed (circuit transitions to half-open after reset_timeout).
	// Using Eventually avoids a fixed sleep that can be flaky under CI scheduling pressure.
	// The backend is still failing, so the probe should fail and re-open the circuit.
	require.Eventually(
		t,
		func() bool {
			_, err := exchanger.Exchange("token-probe", "http://resource.example.com/probe")
			// The probe either reached the backend (non-circuit error) or the circuit is open.
			// We stop polling as soon as we get a non-circuit-open error — that means a probe was attempted.
			return err != nil && !errors.Is(err, server.ErrCircuitOpen)
		},
		resetTimeout+eventuallyTimeout,
		eventuallyPollInterval,
		"probe should be attempted after reset_timeout with backend still down",
	)

	// Circuit should be re-opened — immediate calls should be rejected
	_, err = exchanger.Exchange("token-after-probe", "http://resource.example.com/after-probe")
	assert.ErrorIs(t, err, server.ErrCircuitOpen,
		"circuit must re-open after failed probe")
}

// ---------------------------------------------------------------------------
// Circuit breaker: cache hits bypass circuit breaker
// ---------------------------------------------------------------------------

// Cached tokens are returned even when the circuit is open.
func TestCircuitBreaker_CacheHit_BypassesCircuitBreaker(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	const maxFailures = 2
	cfg := circuitBreakerConfig(mocks, maxFailures, 1*time.Minute)

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// Cache a successful token first
	const cachedKey = "cached-token"
	const cachedResource = "http://resource.example.com/cached"
	token1, err := exchanger.Exchange(cachedKey, cachedResource)
	require.NoError(t, err)

	// Trip the circuit with failures on different keys
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"
	for i := 0; i < maxFailures; i++ {
		_, _ = exchanger.Exchange(fmt.Sprintf("fail-token-%d", i), fmt.Sprintf("http://resource.example.com/fail-%d", i))
	}

	// Circuit is now open — but cached entry should still work
	token2, err := exchanger.Exchange(cachedKey, cachedResource)
	require.NoError(t, err, "cache hit must succeed even when circuit is open")
	assert.Equal(t, token1, token2, "cached token must be returned")
}

// ---------------------------------------------------------------------------
// Circuit breaker: success resets failure count
// ---------------------------------------------------------------------------

// A successful exchange resets the consecutive failure counter.
func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	const maxFailures = 3
	cfg := circuitBreakerConfig(mocks, maxFailures, 30*time.Second)

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// 2 failures (under threshold)
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"
	for i := 0; i < maxFailures-1; i++ {
		_, _ = exchanger.Exchange(fmt.Sprintf("token-%d", i), fmt.Sprintf("http://resource.example.com/%d", i))
	}

	// 1 success — resets failure count
	mocks.exchangeStatus = http.StatusOK
	mocks.exchangedToken = "success-token"
	_, err = exchanger.Exchange("token-success", "http://resource.example.com/success")
	require.NoError(t, err)

	// 2 more failures — should NOT trip because count was reset
	mocks.exchangeStatus = http.StatusInternalServerError
	mocks.exchangeErrCode = "server_error"
	for i := 0; i < maxFailures-1; i++ {
		_, err := exchanger.Exchange(fmt.Sprintf("token-after-%d", i), fmt.Sprintf("http://resource.example.com/after-%d", i))
		assert.Error(t, err)
		assert.NotErrorIs(t, err, server.ErrCircuitOpen,
			"circuit should not be open — failure count was reset by success")
	}
}

// ---------------------------------------------------------------------------
// Circuit breaker: concurrent access is safe
// ---------------------------------------------------------------------------

// Concurrent exchanges must not cause a data race on the circuit breaker.
func TestCircuitBreaker_ConcurrentAccess_RaceFree(t *testing.T) {
	// Use isolated race-safe servers (atomic counters, no shared mutable state)
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

	var failCount int64
	exchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt64(&failCount, 1)
		if count <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, `{"error":"server_error"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer exchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             exchServer.URL + "/oauth2/token",
			Issuer:                    clientCredsServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: clientCredsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			MaxFailures:  10, // high threshold to avoid opening during test
			ResetTimeout: 30 * time.Second,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _ = exchanger.Exchange(
				fmt.Sprintf("token-%d", idx),
				fmt.Sprintf("http://service-%d:9000/api", idx),
			)
		}(i)
	}
	wg.Wait()
	// Test passes if no race conditions are detected by the Go race detector (-race flag)
}

// ---------------------------------------------------------------------------
// Circuit breaker: server.go integration — ErrCircuitOpen returns 503
// ---------------------------------------------------------------------------

// ErrCircuitOpen from exchanger → 503 ImmediateResponse with appropriate body.
func TestServer_Process_CircuitOpen_Returns503(t *testing.T) {
	exchanger := &mockExchanger{
		exchangeFunc: func(_, _ string) (string, error) {
			return "", server.ErrCircuitOpen
		},
	}
	client, cleanup := startTestServer(t, exchanger)
	defer cleanup()

	resp, err := sendRequestHeaders(t, client, map[string]string{
		":path":         "http://mcp-server:9003/mcp",
		"authorization": "Bearer some-token",
	})
	require.NoError(t, err)

	immResp, ok := resp.Response.(*extprocv3.ProcessingResponse_ImmediateResponse)
	require.True(t, ok, "circuit open must produce an ImmediateResponse")
	assert.Equal(t, int32(httpv3.StatusCode_ServiceUnavailable),
		int32(immResp.ImmediateResponse.Status.Code),
		"circuit open must return HTTP 503")
	assert.Contains(t, string(immResp.ImmediateResponse.Body),
		"circuit breaker", "error body must mention circuit breaker")
}
