// Package server — exchanger.go implements RFC 8693 token exchange with
// in-memory caching, singleflight deduplication, and background client
// assertion refresh. It is safe for concurrent use.
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"golang.org/x/sync/singleflight"

	"github.com/sony/gobreaker/v2"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// tokenCacheKey uniquely identifies a cached token by subject + resource.
// Using a struct as a map key avoids separator-injection attacks.
type tokenCacheKey struct {
	subjectToken string
	resourceURI  string
}

// ErrAssertionExpired is returned by Exchange when the stored client assertion
// has expired and the background refresh has not yet succeeded.
// Callers should surface this as a 503 Service Unavailable to signal a transient
// infrastructure failure rather than a client error.
var ErrAssertionExpired = errors.New("client assertion expired")

// cachedToken holds an exchanged access token and its expiry time.
type cachedToken struct {
	accessToken string
	expiresAt   time.Time
}

// isExpired reports whether the cached token has expired.
func (c *cachedToken) isExpired() bool {
	return time.Now().After(c.expiresAt)
}

// tokenExchangeResponse is the JSON shape from the RFC 8693 token exchange endpoint.
type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   *int   `json:"expires_in"` // pointer: nil means absent
}

// assertionState is the atomically-swapped snapshot of the client assertion.
// Storing all fields together ensures readers always observe a consistent snapshot.
type assertionState struct {
	value     string
	issuedAt  time.Time
	expiresAt time.Time
}

// TokenExchanger implements the Exchanger interface.
// It manages a token cache, singleflight group, circuit breaker, and HTTP
// client for outbound calls to the identity broker and the upstream OAuth2 server.
type TokenExchanger struct {
	cfg    *extprocconfig.Config
	client *http.Client
	logger *slog.Logger

	// assertion holds the current client assertion (id_token or access_token).
	// Written only by the background refresh goroutine; read atomically by Exchange callers.
	// Reading a slightly stale but still-valid assertion is acceptable and avoids locking.
	assertion atomic.Pointer[assertionState]

	// cache holds exchanged tokens keyed by subject+resource.
	cacheMu sync.RWMutex
	cache   map[tokenCacheKey]*cachedToken

	// sfGroup deduplicates concurrent cache refresh calls for the same key.
	sfGroup singleflight.Group

	// cb protects outbound token exchange calls with a circuit breaker
	// to prevent thundering herd when the identity broker recovers.
	cb *gobreaker.CircuitBreaker[string]

	// stopCh signals the background goroutine to stop.
	stopCh chan struct{}
}

// NewTokenExchanger creates a new TokenExchanger and performs a startup
// client_credentials grant to obtain the initial client assertion.
// Returns an error if the startup assertion grant fails (fail-fast per FR-006).
func NewTokenExchanger(cfg *extprocconfig.Config, logger *slog.Logger) (*TokenExchanger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}

	httpClient, err := buildHTTPClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("building HTTP client: %w", err)
	}

	te := &TokenExchanger{
		cfg:    cfg,
		client: httpClient,
		logger: logger,
		cache:  make(map[tokenCacheKey]*cachedToken),
		cb:     newGobreakerCB(cfg, logger),
		stopCh: make(chan struct{}),
	}

	// Fail-fast: acquire initial client assertion at startup.
	if err := te.refreshClientAssertion(); err != nil {
		return nil, fmt.Errorf("startup client assertion failed: %w", err)
	}

	// Start background eviction goroutine.
	go te.runEviction()

	return te, nil
}

// Exchange exchanges subjectToken for a downstream token scoped to resourceURI.
// Results are cached; concurrent requests for the same key are deduplicated via singleflight.
func (te *TokenExchanger) Exchange(subjectToken, resourceURI string) (string, error) {
	key := tokenCacheKey{subjectToken: subjectToken, resourceURI: resourceURI}

	// Fast path: cache hit.
	te.cacheMu.RLock()
	if entry, ok := te.cache[key]; ok && !entry.isExpired() {
		te.cacheMu.RUnlock()
		return entry.accessToken, nil
	}
	te.cacheMu.RUnlock()

	// Slow path: singleflight-deduplicated exchange.
	sfKey := subjectToken + "\x00" + resourceURI
	result, err, _ := te.sfGroup.Do(sfKey, func() (interface{}, error) {
		// Re-check cache inside singleflight to handle races.
		te.cacheMu.RLock()
		if entry, ok := te.cache[key]; ok && !entry.isExpired() {
			te.cacheMu.RUnlock()
			return entry.accessToken, nil
		}
		te.cacheMu.RUnlock()

		// Circuit breaker: wrap doExchange in gobreaker.Execute so that
		// gobreaker tracks successes/failures and opens/closes automatically.
		// gobreaker returns ErrOpenState when the circuit is open, or
		// ErrTooManyRequests when the half-open probe slot is taken.
		// Both are mapped to ErrCircuitOpen for callers.
		token, err := te.cb.Execute(func() (string, error) {
			tok, ttl, err := te.doExchange(subjectToken, resourceURI)
			if err != nil {
				return "", err
			}

			te.cacheMu.Lock()
			te.cache[key] = &cachedToken{
				accessToken: tok,
				expiresAt:   time.Now().Add(ttl),
			}
			te.cacheMu.Unlock()

			return tok, nil
		})
		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
				return "", ErrCircuitOpen
			}
			return "", err
		}
		return token, nil
	})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// Shutdown signals the background goroutine to stop and releases resources.
func (te *TokenExchanger) Shutdown() {
	close(te.stopCh)
}

// doExchange performs the RFC 8693 token exchange HTTP call.
// Returns the exchanged access token and the TTL to cache it for.
// Fails fast with ErrAssertionExpired if the stored assertion has expired.
func (te *TokenExchanger) doExchange(subjectToken, resourceURI string) (string, time.Duration, error) {
	s := te.assertion.Load()
	if s == nil || s.value == "" {
		te.logger.Error("client assertion unavailable: no assertion stored")
		return "", 0, ErrAssertionExpired
	}
	if !s.expiresAt.IsZero() && time.Now().After(s.expiresAt) {
		te.logger.Error("client assertion expired: background refresh did not complete in time",
			"expired_at", s.expiresAt.Format(time.RFC3339))
		return "", 0, ErrAssertionExpired
	}
	assertion := s.value

	// Log the full exchange request parameters for observability.
	te.logger.Debug("extproc: token exchange request",
		"token_endpoint", te.cfg.OAuth2.TokenEndpoint,
		"grant_type", "urn:ietf:params:oauth:grant-type:token-exchange",
		"subject_token_type", "urn:ietf:params:oauth:token-type:access_token",
		"resource", resourceURI,
		"client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
	)

	ctx, cancel := context.WithTimeout(context.Background(), te.cfg.OAuth2.ExchangeTimeout)
	defer cancel()

	form := url.Values{
		"grant_type":            {"urn:ietf:params:oauth:grant-type:token-exchange"},
		"subject_token":         {subjectToken},
		"subject_token_type":    {"urn:ietf:params:oauth:token-type:access_token"},
		"resource":              {resourceURI},
		"client_assertion":      {assertion},
		"client_assertion_type": {"urn:ietf:params:oauth:client-assertion-type:jwt-bearer"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		te.cfg.OAuth2.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", 0, fmt.Errorf("building token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := te.client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("reading token exchange response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		te.logger.Warn("token exchange returned non-200",
			"status", resp.StatusCode,
			"resource", resourceURI)
		return "", 0, fmt.Errorf("token exchange returned status %d", resp.StatusCode)
	}

	var exResp tokenExchangeResponse
	if err := json.Unmarshal(body, &exResp); err != nil {
		return "", 0, fmt.Errorf("parsing token exchange response: %w", err)
	}
	if exResp.AccessToken == "" {
		return "", 0, fmt.Errorf("token exchange response missing access_token")
	}

	ttl := te.computeTTL(exResp.ExpiresIn)
	return exResp.AccessToken, ttl, nil
}

// computeTTL returns the cache TTL for an exchanged token.
// Uses expires_in if present, falls back to default_ttl, capped at max_ttl.
func (te *TokenExchanger) computeTTL(expiresIn *int) time.Duration {
	var ttl time.Duration
	if expiresIn != nil && *expiresIn > 0 {
		ttl = time.Duration(*expiresIn) * time.Second
	} else {
		ttl = te.cfg.Cache.DefaultTTL
	}
	if ttl > te.cfg.Cache.MaxTTL {
		ttl = te.cfg.Cache.MaxTTL
	}
	return ttl
}

// refreshClientAssertion obtains a new client assertion (id_token or access_token based on config)
// via the client_credentials grant using golang.org/x/oauth2/clientcredentials.
// Thread-safe: may be called from the background refresh goroutine.
func (te *TokenExchanger) refreshClientAssertion() error {
	endpoint := te.cfg.OAuth2.ClientCredentialsEndpoint
	if endpoint == "" {
		endpoint = te.cfg.OAuth2.Issuer + "/oauth/token"
	}

	ctx, cancel := context.WithTimeout(context.Background(), te.cfg.OAuth2.ExchangeTimeout)
	defer cancel()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, te.client)

	conf := &clientcredentials.Config{
		ClientID:     te.cfg.OAuth2.ClientID,
		ClientSecret: te.cfg.OAuth2.ClientSecret,
		TokenURL:     endpoint,
		Scopes:       clientCredentialsScopes(te.cfg.OAuth2.ClientCredentialsScopes),
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := conf.Token(ctx)
	if err != nil {
		te.logger.Error("client credentials grant failed", "error", err)
		return fmt.Errorf("client credentials grant failed: %w", err)
	}

	var assertion string
	switch te.cfg.OAuth2.ClientAssertionType {
	case "id_token":
		idToken, ok := token.Extra("id_token").(string)
		if !ok || idToken == "" {
			return fmt.Errorf("client credentials response missing id_token (required for client assertion with client_assertion_type=id_token)")
		}
		assertion = idToken
	case "access_token":
		if token.AccessToken == "" {
			return fmt.Errorf("client credentials response missing access_token (required for client assertion with client_assertion_type=access_token)")
		}
		assertion = token.AccessToken
	default:
		return fmt.Errorf("invalid client_assertion_type: %q", te.cfg.OAuth2.ClientAssertionType)
	}

	te.assertion.Store(&assertionState{value: assertion, issuedAt: time.Now(), expiresAt: token.Expiry})

	te.logger.Debug("client assertion refreshed",
		"assertion_type", te.cfg.OAuth2.ClientAssertionType,
		"expires_at", token.Expiry.Format(time.RFC3339))
	return nil
}

// runEviction runs two independent background tasks until Shutdown() is called:
//   - Cache eviction: fires every DefaultTTL/2 (per spec, floor 1s) to sweep expired entries.
//   - Assertion refresh: fires every min(DefaultTTL/2, 30s) (floor 1s) to proactively renew
//     the client assertion before expiry, independent of the eviction cadence.
func (te *TokenExchanger) runEviction() {
	evictInterval := te.cfg.Cache.DefaultTTL / 2
	if evictInterval < time.Second {
		evictInterval = time.Second
	}

	assertionInterval := te.cfg.Cache.DefaultTTL / 2
	if assertionInterval < time.Second {
		assertionInterval = time.Second
	}
	if assertionInterval > 30*time.Second {
		assertionInterval = 30 * time.Second
	}

	evictTicker := time.NewTicker(evictInterval)
	assertionTicker := time.NewTicker(assertionInterval)
	defer evictTicker.Stop()
	defer assertionTicker.Stop()

	for {
		select {
		case <-evictTicker.C:
			te.evictExpired()
		case <-assertionTicker.C:
			te.maybeRefreshAssertion()
		case <-te.stopCh:
			return
		}
	}
}

// evictExpired removes expired entries from the cache.
func (te *TokenExchanger) evictExpired() {
	te.cacheMu.Lock()
	defer te.cacheMu.Unlock()
	for key, entry := range te.cache {
		if entry.isExpired() {
			delete(te.cache, key)
		}
	}
}

// maybeRefreshAssertion refreshes the client assertion proactively.
// It triggers when the remaining lifetime is less than the larger of:
//   - 20% of the total token lifetime (exp − iat), or
//   - 30 seconds (hard floor to ensure refresh before imminent expiry).
func (te *TokenExchanger) maybeRefreshAssertion() {
	s := te.assertion.Load()
	if s == nil || s.expiresAt.IsZero() {
		return
	}

	lifetime := s.expiresAt.Sub(s.issuedAt)
	if s.issuedAt.IsZero() || lifetime <= 0 {
		// Fall back to hard floor when lifetime is unknown.
		lifetime = 0
	}
	threshold := lifetime / 5 // 20% of lifetime
	if threshold < 30*time.Second {
		threshold = 30 * time.Second
	}

	if remaining := time.Until(s.expiresAt); remaining < threshold {
		if err := te.refreshClientAssertion(); err != nil {
			te.logger.Error("background client assertion refresh failed",
				"error", err)
		}
	}
}

// clientCredentialsScopes returns the OAuth2 scopes for the client_credentials grant.
// When no scopes are configured, or when all configured entries are blank after trimming,
// it falls back to ["openid"] to obtain an id_token.
func clientCredentialsScopes(configured []string) []string {
	filtered := make([]string, 0, len(configured))
	for _, s := range configured {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	if len(filtered) == 0 {
		return []string{"openid"}
	}
	return filtered
}

// buildHTTPClient constructs an http.Client respecting the TLS configuration.
// Supports InsecureSkipVerify and CaBundlePath from the TLS config block.
// Returns an error if CaBundlePath is set but the file cannot be read or parsed.
func buildHTTPClient(cfg *extprocconfig.Config) (*http.Client, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.OAuth2.TLS.InsecureSkipVerify, //nolint:gosec // controlled by explicit operator config
	}

	if cfg.OAuth2.TLS.CaBundlePath != "" {
		pemData, err := os.ReadFile(cfg.OAuth2.TLS.CaBundlePath)
		if err != nil {
			return nil, fmt.Errorf("reading ca_bundle_path %q: %w", cfg.OAuth2.TLS.CaBundlePath, err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("ca_bundle_path %q contains no valid PEM certificates", cfg.OAuth2.TLS.CaBundlePath)
		}
		tlsCfg.RootCAs = pool
	}

	transport := &http.Transport{
		TLSClientConfig: tlsCfg,
	}

	return &http.Client{
		Timeout:   cfg.OAuth2.ExchangeTimeout,
		Transport: transport,
	}, nil
}
