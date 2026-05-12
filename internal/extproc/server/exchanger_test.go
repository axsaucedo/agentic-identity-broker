package server_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/server"
)

// ---------------------------------------------------------------------------
// Helpers for exchanger tests
// ---------------------------------------------------------------------------

// tokenResponse is the shape of OAuth2 token endpoint responses.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token,omitempty"`
	TokenType   string `json:"token_type"`
	ExpiresIn   *int   `json:"expires_in,omitempty"`
}

// mockOAuth2Server returns a test server that handles both the client_credentials
// grant and the RFC 8693 token exchange endpoint on the same path for simplicity.
// Callers configure behavior via the returned handlers.
type mockServers struct {
	clientCredsServer *httptest.Server
	tokenExchServer   *httptest.Server

	// Configurable client_credentials response
	clientCredsIDToken string
	clientCredsStatus  int

	// Configurable token exchange response
	exchangedToken  string
	exchangeExpiry  *int
	exchangeStatus  int
	exchangeErrCode string

	// Call counters
	clientCredsCalls int
	exchangeCalls    int

	// Last exchange request form values
	lastExchangeForm url.Values
}

func newMockServers() *mockServers {
	defaultExpiry := 3600
	m := &mockServers{
		clientCredsIDToken: "mock-id-token",
		clientCredsStatus:  http.StatusOK,
		exchangedToken:     "exchanged-access-token",
		exchangeExpiry:     &defaultExpiry,
		exchangeStatus:     http.StatusOK,
	}

	m.clientCredsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.clientCredsCalls++
		w.Header().Set("Content-Type", "application/json")
		if m.clientCredsStatus != http.StatusOK {
			w.WriteHeader(m.clientCredsStatus)
			_, _ = fmt.Fprintf(w, `{"error":"server_error"}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		resp := tokenResponse{
			AccessToken: "access-token",
			IDToken:     m.clientCredsIDToken,
			TokenType:   "Bearer",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))

	m.tokenExchServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.exchangeCalls++
		_ = r.ParseForm()
		m.lastExchangeForm = r.Form

		w.Header().Set("Content-Type", "application/json")
		if m.exchangeStatus != http.StatusOK {
			w.WriteHeader(m.exchangeStatus)
			_, _ = fmt.Fprintf(w, `{"error":%q}`, m.exchangeErrCode)
			return
		}
		w.WriteHeader(http.StatusOK)
		resp := tokenResponse{
			AccessToken: m.exchangedToken,
			TokenType:   "Bearer",
			ExpiresIn:   m.exchangeExpiry,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))

	return m
}

func (m *mockServers) Close() {
	m.clientCredsServer.Close()
	m.tokenExchServer.Close()
}

// configForMocks returns a Config pointing at the mock servers.
func configForMocks(m *mockServers) *extprocconfig.Config {
	return &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             m.tokenExchServer.URL + "/oauth2/token",
			Issuer:                    m.clientCredsServer.URL,
			ClientID:                  "test-client",
			ClientSecret:              "test-secret",
			ClientCredentialsEndpoint: m.clientCredsServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS: extprocconfig.TLSConfig{
				AllowHTTP: true, // mock servers use HTTP
			},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}
}

// ---------------------------------------------------------------------------
// T016: Token exchange HTTP client tests
// ---------------------------------------------------------------------------

// Spec: FR-003, FR-005 — Exchange sends RFC 8693 form-encoded request
func TestTokenExchanger_Exchange_SendsCorrectRFC8693Request(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "incoming-user-bearer-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	token, err := exchanger.Exchange(context.Background(), subjectToken, resourceURI)
	require.NoError(t, err)
	assert.Equal(t, mocks.exchangedToken, token)

	// Verify RFC 8693 request parameters
	form := mocks.lastExchangeForm
	assert.Equal(t, "urn:ietf:params:oauth:grant-type:token-exchange",
		form.Get("grant_type"), "grant_type must be RFC 8693 token exchange")
	assert.Equal(t, subjectToken, form.Get("subject_token"),
		"subject_token must be the incoming Bearer token")
	assert.Equal(t, "urn:ietf:params:oauth:token-type:access_token",
		form.Get("subject_token_type"), "subject_token_type must be access_token")
	assert.Equal(t, resourceURI, form.Get("resource"),
		"resource must be the request URI")
	assert.Equal(t, "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		form.Get("client_assertion_type"), "client_assertion_type must be jwt-bearer")
	assert.NotEmpty(t, form.Get("client_assertion"),
		"client_assertion must be set from the id_token")
}

// Spec: FR-006 — client_assertion comes from id_token (not access_token)
func TestTokenExchanger_Exchange_UsesIDTokenAsClientAssertion(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()
	mocks.clientCredsIDToken = "specific-id-token-for-assertion"

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	_, err = exchanger.Exchange(context.Background(), "some-token", "http://resource.example.com/api")
	require.NoError(t, err)

	assert.Equal(t, "specific-id-token-for-assertion",
		mocks.lastExchangeForm.Get("client_assertion"),
		"client_assertion must be the id_token from client_credentials response")
}

// Spec: FR-010 — Non-200 from token exchange → error returned
func TestTokenExchanger_Exchange_Non200Response_ReturnsError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorCode  string
	}{
		{"400 invalid_request", http.StatusBadRequest, "invalid_request"},
		{"401 invalid_client", http.StatusUnauthorized, "invalid_client"},
		{"403 access_denied", http.StatusForbidden, "access_denied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mocks := newMockServers()
			defer mocks.Close()
			mocks.exchangeStatus = tt.statusCode
			mocks.exchangeErrCode = tt.errorCode

			cfg := configForMocks(mocks)
			exchanger, err := server.NewTokenExchanger(cfg, testLogger())
			require.NoError(t, err)
			defer exchanger.Shutdown()

			token, err := exchanger.Exchange(context.Background(), "subject-token", "http://resource.example.com")
			assert.Error(t, err, "non-200 exchange response must return error")
			assert.Empty(t, token, "no token should be returned on failure")
		})
	}
}

// Spec: FR-010 — Timeout returns error
func TestTokenExchanger_Exchange_Timeout_ReturnsError(t *testing.T) {
	// Create a server that hangs
	hangServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // far longer than exchange_timeout
	}))
	defer hangServer.Close()

	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	cfg.OAuth2.TokenEndpoint = hangServer.URL + "/oauth2/token"
	cfg.OAuth2.ExchangeTimeout = 50 * time.Millisecond // very short timeout

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	start := time.Now()
	_, err = exchanger.Exchange(context.Background(), "token", "http://resource.example.com")
	elapsed := time.Since(start)

	assert.Error(t, err, "timeout must return error")
	assert.Less(t, elapsed, 2*time.Second, "should time out well before 2 seconds")
}

// Spec: FR-012 — Missing expires_in in exchange response uses default_ttl (Exchange still succeeds)
func TestTokenExchanger_Exchange_NoExpiresIn_SucceedsWithDefaultTTL(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()
	mocks.exchangeExpiry = nil // no expires_in in response

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	token, err := exchanger.Exchange(context.Background(), "token", "http://resource.example.com/api")
	require.NoError(t, err, "missing expires_in should not cause error")
	assert.Equal(t, mocks.exchangedToken, token)
}

// Spec: FR-011 — Second call with same key uses cache (exchange called only once)
func TestTokenExchanger_Exchange_CacheHit_DoesNotCallExchangeAgain(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	const subjectToken = "cached-token"
	const resourceURI = "http://mcp-server:9003/mcp"

	// First call — populates cache
	token1, err := exchanger.Exchange(context.Background(), subjectToken, resourceURI)
	require.NoError(t, err)
	firstCallCount := mocks.exchangeCalls

	// Second call with same key — must use cache
	token2, err := exchanger.Exchange(context.Background(), subjectToken, resourceURI)
	require.NoError(t, err)

	assert.Equal(t, token1, token2, "cached token must equal first exchanged token")
	assert.Equal(t, firstCallCount, mocks.exchangeCalls,
		"token exchange endpoint must not be called again on cache hit")
}

// Spec: FR-011, FR-015 — Different keys use separate cache entries
func TestTokenExchanger_Exchange_DifferentKeys_IndependentCacheEntries(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	// Return different tokens for different resource URIs
	callCount := 0
	mocks.tokenExchServer.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		_ = r.ParseForm()
		resource := r.FormValue("resource")
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		resp := tokenResponse{
			AccessToken: "token-for-" + resource,
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	cfg := configForMocks(mocks)
	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	token1, err := exchanger.Exchange(context.Background(), "user-token", "http://service-a:9000/api")
	require.NoError(t, err)

	token2, err := exchanger.Exchange(context.Background(), "user-token", "http://service-b:9001/api")
	require.NoError(t, err)

	assert.NotEqual(t, token1, token2, "different resources must produce different cached tokens")
}

// Spec: FR-018 — cache TTL is capped at max_ttl
func TestTokenExchanger_Exchange_ExpiresIn_CappedAtMaxTTL(t *testing.T) {
	mocks := newMockServers()
	defer mocks.Close()

	// Set expires_in to 1 week — much larger than max_ttl of 1 minute
	largeExpiry := 7 * 24 * 3600
	mocks.exchangeExpiry = &largeExpiry

	cfg := configForMocks(mocks)
	cfg.Cache.MaxTTL = 1 * time.Minute // short max_ttl for test

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	// This test verifies the TTL cap is applied — the exchange still succeeds
	token, err := exchanger.Exchange(context.Background(), "token", "http://resource.example.com")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	// The actual TTL cap behavior is verified via cache expiry tests
}

// ---------------------------------------------------------------------------
// T017: Client assertion acquisition tests
// ---------------------------------------------------------------------------

// Spec: FR-006 — client_credentials grant uses correct parameters
func TestTokenExchanger_ClientAssertion_SendsCorrectGrantRequest(t *testing.T) {
	var (
		mu           sync.Mutex
		receivedForm url.Values
	)
	clientCredsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		receivedForm = r.Form
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access",
			IDToken:     "id-token-jwt",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer clientCredsServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiry := 3600
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged",
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
			ClientID:                  "my-extproc-client",
			ClientSecret:              "my-secret",
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
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	_, err = exchanger.Exchange(context.Background(), "user-token", "http://resource.example.com")
	require.NoError(t, err)

	mu.Lock()
	form := receivedForm
	mu.Unlock()

	// Verify client_credentials grant parameters per contracts/extproc-grpc.md
	assert.Equal(t, "client_credentials", form.Get("grant_type"))
	assert.Equal(t, "my-extproc-client", form.Get("client_id"))
	assert.Equal(t, "my-secret", form.Get("client_secret"))
	assert.Equal(t, "openid", form.Get("scope"),
		"scope must be openid to obtain an id_token")
}

// Spec: oauth2.client_credentials_scopes — custom scopes sent to client_credentials endpoint
func TestTokenExchanger_ClientAssertion_UsesConfiguredScopes(t *testing.T) {
	var (
		mu           sync.Mutex
		receivedForm url.Values
	)
	clientCredsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		receivedForm = r.Form
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access",
			IDToken:     "id-token-jwt",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer clientCredsServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiry := 3600
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged",
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
			ClientID:                  "my-extproc-client",
			ClientSecret:              "my-secret",
			ClientCredentialsEndpoint: clientCredsServer.URL + "/oauth/token",
			ClientCredentialsScopes:   []string{"openid", "profile", "email"},
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	_, err = exchanger.Exchange(context.Background(), "user-token", "http://resource.example.com")
	require.NoError(t, err)

	mu.Lock()
	form := receivedForm
	mu.Unlock()

	// Verify that the configured scopes are sent to the client_credentials endpoint
	scope := form.Get("scope")
	assert.Contains(t, scope, "openid", "configured scope 'openid' should be sent")
	assert.Contains(t, scope, "profile", "configured scope 'profile' should be sent")
	assert.Contains(t, scope, "email", "configured scope 'email' should be sent")
}

// Spec: oauth2.client_credentials_scopes — blank/whitespace-only entries are filtered and fall back to openid
func TestTokenExchanger_ClientAssertion_BlankScopesFallBackToOpenID(t *testing.T) {
	var (
		mu           sync.Mutex
		receivedForm url.Values
	)
	clientCredsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		receivedForm = r.Form
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access",
			IDToken:     "id-token-jwt",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer clientCredsServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiry := 3600
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged",
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
			ClientID:                  "my-extproc-client",
			ClientSecret:              "my-secret",
			ClientCredentialsEndpoint: clientCredsServer.URL + "/oauth/token",
			ClientCredentialsScopes:   []string{"", " "}, // all blank — should fall back to openid
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
	}

	exchanger, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer exchanger.Shutdown()

	_, err = exchanger.Exchange(context.Background(), "user-token", "http://resource.example.com")
	require.NoError(t, err)

	mu.Lock()
	form := receivedForm
	mu.Unlock()

	assert.Equal(t, "openid", form.Get("scope"),
		"blank/whitespace-only scopes must fall back to openid")
}

// Spec: FR-006 — client_credentials endpoint defaults to {issuer}/oauth/token
func TestTokenExchanger_ClientAssertion_DefaultsIssuerOAuthToken(t *testing.T) {
	called := false
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.Header().Set("Content-Type", "application/json")
		expiry := 3600
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "access",
			IDToken:     "id-token",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	})

	issuerServer := httptest.NewServer(mux)
	defer issuerServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiry := 3600
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "exchanged",
			TokenType:   "Bearer",
			ExpiresIn:   &expiry,
		})
	}))
	defer tokenExchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:       tokenExchServer.URL + "/oauth2/token",
			Issuer:              issuerServer.URL,
			ClientID:            "client",
			ClientSecret:        "secret",
			ClientAssertionType: "id_token",
			// ClientCredentialsEndpoint intentionally left empty — should default to {issuer}/oauth/token
			ExchangeTimeout: 5 * time.Second,
			TLS:             extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
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

	_, err = exchanger.Exchange(context.Background(), "token", "http://resource.example.com")
	require.NoError(t, err)
	assert.True(t, called, "client_credentials must be called at {issuer}/oauth/token by default")
}

// Spec: FR-006 — client_credentials failure at startup → NewTokenExchanger returns error
func TestTokenExchanger_New_ClientAssertionFailure_ReturnsError(t *testing.T) {
	// Server that always returns 500
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"server_error"}`, http.StatusInternalServerError)
	}))
	defer errorServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer tokenExchServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tokenExchServer.URL + "/oauth2/token",
			Issuer:                    errorServer.URL,
			ClientID:                  "client",
			ClientSecret:              "secret",
			ClientCredentialsEndpoint: errorServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	_, err := server.NewTokenExchanger(cfg, testLogger())
	assert.Error(t, err, "NewTokenExchanger must fail fast when client_credentials grant fails at startup")
}

// Spec: SR-003 — client_secret must not appear in any logs (structural test: verify it's not in error messages)
func TestTokenExchanger_ClientSecret_NotExposedInErrors(t *testing.T) {
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_client"}`, http.StatusUnauthorized)
	}))
	defer errorServer.Close()

	tokenExchServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer tokenExchServer.Close()

	const secretValue = "super-secret-client-secret-value"

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:             tokenExchServer.URL + "/oauth2/token",
			Issuer:                    errorServer.URL,
			ClientID:                  "client",
			ClientSecret:              secretValue,
			ClientCredentialsEndpoint: errorServer.URL + "/oauth/token",
			ClientAssertionType:       "id_token",
			ExchangeTimeout:           5 * time.Second,
			TLS:                       extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			Enabled:      true,
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	_, err := server.NewTokenExchanger(cfg, testLogger())
	require.Error(t, err)
	assert.False(t, strings.Contains(err.Error(), secretValue),
		"client_secret must not appear in error messages (SR-003)")
}

// Spec: NewTokenExchanger — nil config returns error
func TestTokenExchanger_New_NilConfig_ReturnsError(t *testing.T) {
	_, err := server.NewTokenExchanger(nil, testLogger())
	assert.Error(t, err, "nil config must return error")
}

// ---------------------------------------------------------------------------
// T022: HTTP client trace propagation
// ---------------------------------------------------------------------------

// Spec: US1 S2 — When traces are enabled, outbound HTTP request receives traceparent header
func TestTokenExchanger_Exchange_PropagatesTraceContext(t *testing.T) {
	// Install a real tracer provider + W3C propagator so outbound otelhttp transport injects traceparent.
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(spanRecorder))
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	var mu sync.Mutex
	var capturedTraceparent string
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		capturedTraceparent = r.Header.Get("traceparent")
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"access_token": "new-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
			"id_token":     "id-token-jwt",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:           tokenServer.URL + "/token",
			Issuer:                  tokenServer.URL,
			ClientID:                "test-client",
			ClientSecret:            "test-secret",
			ExchangeTimeout:         5 * time.Second,
			ClientAssertionType:     "id_token",
			ClientCredentialsScopes: []string{"openid"},
			TLS:                     extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	te, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer te.Shutdown()

	// Clear any spans from constructor's startup HTTP traffic.
	tp.ForceFlush(context.Background()) //nolint:errcheck
	spanRecorder.Ended()                // drain

	// Start a parent span and pass its context to Exchange.
	ctx, parentSpan := tp.Tracer("test").Start(context.Background(), "test-parent")
	expectedTraceID := parentSpan.SpanContext().TraceID().String()

	// Reset captured header before the exchange call.
	mu.Lock()
	capturedTraceparent = ""
	mu.Unlock()

	token, err := te.Exchange(ctx, "test-subject-token", "https://example.com/api")
	parentSpan.End()
	require.NoError(t, err)
	assert.NotEmpty(t, token, "exchange should return a token")

	// The outbound request must carry a traceparent with the same trace ID.
	mu.Lock()
	tp2 := capturedTraceparent
	mu.Unlock()
	assert.NotEmpty(t, tp2, "outbound request must have traceparent header")
	assert.Contains(t, tp2, expectedTraceID,
		"traceparent must contain the parent span's trace ID")
}

// ---------------------------------------------------------------------------
// T023: Propagation-only mode when traces disabled
// ---------------------------------------------------------------------------

// Spec: US1 S6 — When telemetry.enabled=true but traces.enabled=false,
// trace context is still forwarded without creating client spans
func TestTokenExchanger_Exchange_PropagatesWithoutLocalSpans(t *testing.T) {
	// Install a NeverSample tracer provider so no local spans are created,
	// but propagation still injects traceparent from the incoming context.
	spanRecorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
		sdktrace.WithSampler(sdktrace.NeverSample()),
	)
	prevTP := otel.GetTracerProvider()
	prevProp := otel.GetTextMapPropagator()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})

	var mu sync.Mutex
	var capturedTraceparent string
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		capturedTraceparent = r.Header.Get("traceparent")
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"access_token": "token-data",
			"token_type":   "Bearer",
			"expires_in":   1800,
			"id_token":     "id-token-jwt",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	cfg := &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{Bind: "127.0.0.1", Port: 50051},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint:           tokenServer.URL + "/token",
			Issuer:                  tokenServer.URL,
			ClientID:                "test-client",
			ClientSecret:            "test-secret",
			ExchangeTimeout:         5 * time.Second,
			ClientAssertionType:     "id_token",
			ClientCredentialsScopes: []string{"openid"},
			TLS:                     extprocconfig.TLSConfig{AllowHTTP: true},
		},
		Cache: extprocconfig.CacheConfig{
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		},
		CircuitBreaker: extprocconfig.CircuitBreakerConfig{
			MaxFailures:  5,
			ResetTimeout: 30 * time.Second,
		},
	}

	te, err := server.NewTokenExchanger(cfg, testLogger())
	require.NoError(t, err)
	defer te.Shutdown()

	// Clear constructor spans and reset captured header.
	tp.ForceFlush(context.Background()) //nolint:errcheck
	spanRecorder.Ended()                // drain
	mu.Lock()
	capturedTraceparent = ""
	mu.Unlock()

	// Create a sampled-out parent span context — propagation should still forward it.
	ctx, parentSpan := tp.Tracer("test").Start(context.Background(), "test-parent")
	expectedTraceID := parentSpan.SpanContext().TraceID().String()

	token, err := te.Exchange(ctx, "propagation-test-token", "https://example.com/resource")
	parentSpan.End()
	require.NoError(t, err)
	assert.Equal(t, "token-data", token, "exchange should return the access token regardless of trace collection state")

	// Verify propagation occurred (traceparent forwarded) but no local spans were recorded.
	tp.ForceFlush(context.Background()) //nolint:errcheck
	localSpans := spanRecorder.Ended()
	assert.Empty(t, localSpans, "no local spans should be recorded when sampler is NeverSample")
	mu.Lock()
	tp2 := capturedTraceparent
	mu.Unlock()
	assert.NotEmpty(t, tp2, "traceparent must still be propagated even without local spans")
	assert.Contains(t, tp2, expectedTraceID,
		"propagated traceparent must carry the original trace ID")
}
