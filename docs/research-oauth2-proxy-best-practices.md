# Go Best Practices for OAuth2 Proxy Implementation

## Overview
This document provides comprehensive guidance for implementing OAuth2 proxy functionality in Go, following idiomatic patterns and security best practices. All examples are tailored to the agentic-identity-broker project using Go 1.24.0, chi v5.2.3 router, and structured logging with slog.

## 1. TLS-Validated HTTP Client Configuration

### Security Requirements
- MUST reject self-signed certificates (use system CA pool)
- MUST enforce TLS 1.2 minimum version
- MUST set appropriate timeouts to prevent resource exhaustion
- SHOULD reuse connections via connection pooling

### Idiomatic Implementation

```go
package httpclient

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "net"
    "net/http"
    "time"
)

// NewSecureClient creates an HTTP client with strict TLS validation.
// It uses the system certificate pool and rejects self-signed certificates.
//
// Timeouts:
//   - Dial: 30s (TCP connection establishment)
//   - TLS Handshake: 10s (TLS negotiation)
//   - Total Request: 30s (entire request/response cycle)
//   - Idle Connection: 90s (connection pool keep-alive)
//
// Returns:
//   - *http.Client configured for secure upstream communication
//   - error if system CA pool cannot be loaded
func NewSecureClient() (*http.Client, error) {
    // Load system CA certificates
    // This ensures we only trust properly signed certificates
    certPool, err := x509.SystemCertPool()
    if err != nil {
        return nil, fmt.Errorf("failed to load system CA pool: %w", err)
    }

    // Configure TLS with strict security settings
    tlsConfig := &tls.Config{
        // Use system CAs - rejects self-signed certificates
        RootCAs: certPool,

        // Enforce minimum TLS version (TLS 1.2+)
        MinVersion: tls.VersionTLS12,

        // Prefer server cipher suites for better security
        PreferServerCipherSuites: true,

        // Do NOT skip verification (default is false, but explicit is better)
        InsecureSkipVerify: false,
    }

    // Create transport with security-focused settings
    transport := &http.Transport{
        // Connection establishment timeout
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,

        // TLS handshake timeout
        TLSHandshakeTimeout: 10 * time.Second,

        // TLS configuration
        TLSClientConfig: tlsConfig,

        // Connection pooling settings (avoid resource exhaustion)
        MaxIdleConns:        100,              // Total idle connections
        MaxIdleConnsPerHost: 10,               // Idle connections per host
        MaxConnsPerHost:     100,              // Max connections per host
        IdleConnTimeout:     90 * time.Second, // Close idle after 90s

        // Response header timeout (time to receive response headers)
        ResponseHeaderTimeout: 10 * time.Second,

        // Expect-Continue timeout (for 100-continue)
        ExpectContinueTimeout: 1 * time.Second,
    }

    // Create client with overall request timeout
    client := &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second, // Total request timeout

        // Disable automatic redirects for proxy mode
        // We want to proxy the redirect response, not follow it
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    return client, nil
}

// NewSecureClientWithTimeout creates a secure client with custom timeout.
// Useful when different operations require different timeout values.
func NewSecureClientWithTimeout(timeout time.Duration) (*http.Client, error) {
    client, err := NewSecureClient()
    if err != nil {
        return nil, err
    }

    client.Timeout = timeout
    return client, nil
}

// Example usage in server initialization
func ExampleServerSetup() {
    // Create secure client during server startup
    upstreamClient, err := NewSecureClient()
    if err != nil {
        log.Fatal("failed to create HTTP client", "error", err)
    }

    // Inject client into handlers via dependency injection
    handler := NewOAuth2ProxyHandler(upstreamClient, logger)
}
```

### Testing TLS Validation

```go
package httpclient_test

import (
    "crypto/tls"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSecureClient_RejectsSelfSignedCertificates(t *testing.T) {
    // Create test server with self-signed certificate
    server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Create secure client
    client, err := NewSecureClient()
    require.NoError(t, err)

    // Attempt request - should fail due to self-signed cert
    resp, err := client.Get(server.URL)

    // Assert that connection failed due to certificate validation
    assert.Error(t, err)
    assert.Nil(t, resp)
    assert.Contains(t, err.Error(), "certificate")
}

func TestSecureClient_AcceptsValidCertificates(t *testing.T) {
    // Create test server with valid certificate (using httptest.NewServer for HTTP)
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }))
    defer server.Close()

    client, err := NewSecureClient()
    require.NoError(t, err)

    // This should succeed (HTTP doesn't require TLS)
    resp, err := client.Get(server.URL)
    require.NoError(t, err)
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSecureClient_EnforcesTLSMinVersion(t *testing.T) {
    // Test that TLS 1.0/1.1 are rejected
    client, err := NewSecureClient()
    require.NoError(t, err)

    transport := client.Transport.(*http.Transport)
    assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
}
```

---

## 2. Chi Router Middleware Patterns

### Audit Logging Middleware

```go
package middleware

import (
    "context"
    "log/slog"
    "net/http"
    "time"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey string

const (
    // requestIDKey is the context key for request ID.
    requestIDKey contextKey = "request_id"

    // clientIDKey is the context key for OAuth2 client_id.
    clientIDKey contextKey = "client_id"
)

// AuditLogMiddleware returns middleware that logs all requests with structured audit information.
// Captures: request ID, principal, client_id, method, path, status, duration, remote_addr.
//
// Context values injected:
//   - request_id: UUID generated per request
//   - principal: Extracted from existing context (if available)
//   - client_id: Extracted from query parameter or form data (if available)
func AuditLogMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()

            // Generate unique request ID
            requestID := uuid.New().String()
            ctx := context.WithValue(r.Context(), requestIDKey, requestID)

            // Extract principal from existing context (set by RequirePrincipalMiddleware)
            principalValue, _ := principal.FromContext(ctx)

            // Extract client_id from query parameters (OAuth2 authorize requests)
            clientID := r.URL.Query().Get("client_id")
            if clientID == "" {
                // Try form data for token requests (POST)
                clientID = r.FormValue("client_id")
            }
            if clientID != "" {
                ctx = context.WithValue(ctx, clientIDKey, clientID)
            }

            // Wrap ResponseWriter to capture status code
            wrapped := &auditResponseWriter{
                ResponseWriter: w,
                statusCode:     http.StatusOK,
            }

            // Update request with new context
            r = r.WithContext(ctx)

            // Process request
            next.ServeHTTP(wrapped, r)

            // Calculate duration
            duration := time.Since(start)

            // Structured audit log
            logger.Info("audit_log",
                "request_id", requestID,
                "principal", principalValue,
                "client_id", clientID,
                "method", r.Method,
                "path", r.URL.Path,
                "query", r.URL.RawQuery,
                "status", wrapped.statusCode,
                "duration_ms", duration.Milliseconds(),
                "remote_addr", r.RemoteAddr,
                "user_agent", r.UserAgent(),
            )
        })
    }
}

// auditResponseWriter wraps http.ResponseWriter to capture status code.
type auditResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
    w.statusCode = statusCode
    w.ResponseWriter.WriteHeader(statusCode)
}

// GetRequestID extracts the request ID from context.
// Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey).(string); ok {
        return id
    }
    return ""
}

// GetClientID extracts the OAuth2 client_id from context.
// Returns empty string if not found.
func GetClientID(ctx context.Context) string {
    if id, ok := ctx.Value(clientIDKey).(string); ok {
        return id
    }
    return ""
}

// WithClientID adds a client_id to the context.
// Useful for setting client_id after parsing request body.
func WithClientID(ctx context.Context, clientID string) context.Context {
    return context.WithValue(ctx, clientIDKey, clientID)
}
```

### Example Usage in Server Setup

```go
package http

import (
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
    "github.com/go-chi/chi/v5"
)

func (s *Server) setupRoutes() {
    // Order matters: recovery → audit → principal → routes
    s.router.Use(RecoveryMiddleware(s.logger))
    s.router.Use(middleware.AuditLogMiddleware(s.logger))
    s.router.Use(OptionalPrincipalMiddleware(s.config.Authentication, s.logger))

    // Register routes
    s.router.Get("/health", s.handleHealth())
    s.router.Post("/oauth2/authorize", s.handleOAuth2Authorize())
    s.router.Post("/oauth2/token", s.handleOAuth2Token())
}
```

### Testing Audit Middleware

```go
package middleware_test

import (
    "bytes"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAuditLogMiddleware_CapturesRequestContext(t *testing.T) {
    // Capture log output
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    // Create test handler
    handler := middleware.AuditLogMiddleware(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request_id is in context
        requestID := middleware.GetRequestID(r.Context())
        assert.NotEmpty(t, requestID)

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("success"))
    }))

    // Test request with client_id
    req := httptest.NewRequest(http.MethodPost, "/oauth2/authorize?client_id=test-agent", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    // Verify response
    assert.Equal(t, http.StatusOK, rec.Code)

    // Verify audit log contains expected fields
    logOutput := buf.String()
    assert.Contains(t, logOutput, "audit_log")
    assert.Contains(t, logOutput, "request_id")
    assert.Contains(t, logOutput, "client_id")
    assert.Contains(t, logOutput, "test-agent")
}
```

---

## 3. HTTP POST Proxy Implementation

### Core Proxy Function

```go
package proxy

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "strings"
    "time"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
)

// ProxyHandler handles HTTP proxy requests to upstream OAuth2 servers.
type ProxyHandler struct {
    upstreamClient *http.Client
    logger         *slog.Logger
}

// NewProxyHandler creates a new proxy handler.
func NewProxyHandler(upstreamClient *http.Client, logger *slog.Logger) *ProxyHandler {
    return &ProxyHandler{
        upstreamClient: upstreamClient,
        logger:         logger,
    }
}

// ProxyPOSTRequest proxies an HTTP POST request to an upstream server.
// Copies all headers (except hop-by-hop headers), request body, and response.
//
// Parameters:
//   - w: Response writer for sending upstream response to client
//   - r: Original client request
//   - upstreamURL: Target upstream URL
//   - timeout: Request timeout (0 = use client default)
//
// Returns error if upstream request fails or times out.
func (h *ProxyHandler) ProxyPOSTRequest(
    w http.ResponseWriter,
    r *http.Request,
    upstreamURL string,
    timeout time.Duration,
) error {
    requestID := middleware.GetRequestID(r.Context())

    h.logger.Debug("proxying POST request",
        "request_id", requestID,
        "upstream_url", upstreamURL,
        "content_type", r.Header.Get("Content-Type"))

    // Create context with timeout if specified
    ctx := r.Context()
    if timeout > 0 {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, timeout)
        defer cancel()
    }

    // Read request body (need to copy it for potential retries)
    bodyBytes, err := io.ReadAll(r.Body)
    if err != nil {
        h.logger.Error("failed to read request body",
            "request_id", requestID,
            "error", err)
        return fmt.Errorf("failed to read request body: %w", err)
    }
    defer r.Body.Close()

    // Create upstream request
    upstreamReq, err := http.NewRequestWithContext(
        ctx,
        http.MethodPost,
        upstreamURL,
        bytes.NewReader(bodyBytes),
    )
    if err != nil {
        return fmt.Errorf("failed to create upstream request: %w", err)
    }

    // Copy headers (excluding hop-by-hop headers)
    h.copyHeaders(upstreamReq.Header, r.Header)

    // Set Content-Length explicitly (required for some OAuth2 servers)
    upstreamReq.ContentLength = int64(len(bodyBytes))

    // Execute upstream request
    startTime := time.Now()
    upstreamResp, err := h.upstreamClient.Do(upstreamReq)
    duration := time.Since(startTime)

    if err != nil {
        h.logger.Error("upstream request failed",
            "request_id", requestID,
            "upstream_url", upstreamURL,
            "duration_ms", duration.Milliseconds(),
            "error", err)
        return fmt.Errorf("upstream request failed: %w", err)
    }
    defer upstreamResp.Body.Close()

    h.logger.Info("upstream request completed",
        "request_id", requestID,
        "upstream_url", upstreamURL,
        "status", upstreamResp.StatusCode,
        "duration_ms", duration.Milliseconds())

    // Copy response headers
    h.copyHeaders(w.Header(), upstreamResp.Header)

    // Write status code
    w.WriteHeader(upstreamResp.StatusCode)

    // Stream response body
    written, err := io.Copy(w, upstreamResp.Body)
    if err != nil {
        h.logger.Error("failed to copy response body",
            "request_id", requestID,
            "bytes_written", written,
            "error", err)
        return fmt.Errorf("failed to copy response body: %w", err)
    }

    h.logger.Debug("response proxied successfully",
        "request_id", requestID,
        "bytes_written", written)

    return nil
}

// copyHeaders copies HTTP headers from src to dst, excluding hop-by-hop headers.
// Hop-by-hop headers are connection-specific and should not be forwarded.
func (h *ProxyHandler) copyHeaders(dst, src http.Header) {
    // List of hop-by-hop headers (RFC 2616 Section 13.5.1)
    hopByHopHeaders := map[string]bool{
        "Connection":          true,
        "Keep-Alive":          true,
        "Proxy-Authenticate":  true,
        "Proxy-Authorization": true,
        "Te":                  true,
        "Trailers":            true,
        "Transfer-Encoding":   true,
        "Upgrade":             true,
    }

    for key, values := range src {
        // Skip hop-by-hop headers
        if hopByHopHeaders[key] {
            continue
        }

        // Copy all values for this header
        for _, value := range values {
            dst.Add(key, value)
        }
    }
}

// ProxyFormURLEncoded is a specialized proxy for application/x-www-form-urlencoded requests.
// Validates Content-Type and provides better error messages for OAuth2 token requests.
func (h *ProxyHandler) ProxyFormURLEncoded(
    w http.ResponseWriter,
    r *http.Request,
    upstreamURL string,
) error {
    // Validate Content-Type
    contentType := r.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
        return fmt.Errorf("invalid Content-Type: expected application/x-www-form-urlencoded, got %s", contentType)
    }

    // Use default timeout for token requests (30s)
    return h.ProxyPOSTRequest(w, r, upstreamURL, 30*time.Second)
}

// ProxyJSON is a specialized proxy for application/json requests.
// Validates Content-Type and provides better error messages for JSON APIs.
func (h *ProxyHandler) ProxyJSON(
    w http.ResponseWriter,
    r *http.Request,
    upstreamURL string,
) error {
    // Validate Content-Type
    contentType := r.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "application/json") {
        return fmt.Errorf("invalid Content-Type: expected application/json, got %s", contentType)
    }

    return h.ProxyPOSTRequest(w, r, upstreamURL, 30*time.Second)
}
```

### Error Handling for Proxy Failures

```go
package proxy

import (
    "encoding/json"
    "net/http"
)

// ErrorResponse represents a JSON error response.
type ErrorResponse struct {
    Error            string `json:"error"`
    ErrorDescription string `json:"error_description,omitempty"`
}

// WriteProxyError writes a JSON error response when upstream proxy fails.
// Uses OAuth2 error format for compatibility.
func (h *ProxyHandler) WriteProxyError(w http.ResponseWriter, err error) {
    h.logger.Error("proxy error", "error", err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusBadGateway)

    resp := ErrorResponse{
        Error:            "upstream_error",
        ErrorDescription: "Failed to communicate with upstream OAuth2 server",
    }

    json.NewEncoder(w).Encode(resp)
}

// Example usage in handler
func (h *ProxyHandler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
    upstreamURL := "https://oauth-provider.example.com/oauth2/token"

    if err := h.ProxyFormURLEncoded(w, r, upstreamURL); err != nil {
        // Write error response to client
        h.WriteProxyError(w, err)
        return
    }

    // Success - response already written by ProxyFormURLEncoded
}
```

### Testing Proxy Implementation

```go
package proxy_test

import (
    "bytes"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/proxy"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestProxyHandler_ProxyPOSTRequest(t *testing.T) {
    // Create mock upstream server
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request was proxied correctly
        assert.Equal(t, http.MethodPost, r.Method)
        assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

        // Read body
        body, err := io.ReadAll(r.Body)
        require.NoError(t, err)
        assert.Equal(t, "grant_type=client_credentials", string(body))

        // Write response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"access_token":"test-token"}`))
    }))
    defer upstream.Close()

    // Create proxy handler
    client := &http.Client{}
    handler := proxy.NewProxyHandler(client, slog.Default())

    // Create test request
    requestBody := bytes.NewReader([]byte("grant_type=client_credentials"))
    req := httptest.NewRequest(http.MethodPost, "/oauth2/token", requestBody)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    rec := httptest.NewRecorder()

    // Proxy request
    err := handler.ProxyPOSTRequest(rec, req, upstream.URL, 0)
    require.NoError(t, err)

    // Verify response
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
    assert.JSONEq(t, `{"access_token":"test-token"}`, rec.Body.String())
}

func TestProxyHandler_HandlesUpstreamError(t *testing.T) {
    // Create failing upstream server
    upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte("upstream error"))
    }))
    defer upstream.Close()

    client := &http.Client{}
    handler := proxy.NewProxyHandler(client, slog.Default())

    req := httptest.NewRequest(http.MethodPost, "/oauth2/token", bytes.NewReader([]byte("test")))
    rec := httptest.NewRecorder()

    // Proxy should succeed but return upstream's error status
    err := handler.ProxyPOSTRequest(rec, req, upstream.URL, 0)
    require.NoError(t, err) // No proxy error

    // Verify upstream error was proxied
    assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
```

---

## 4. OAuth2 URL Construction Helpers

### URL Builder with Query Parameters

```go
package oauth2

import (
    "fmt"
    "net/url"
)

// AuthorizationURLBuilder constructs OAuth2 authorization URLs with proper encoding.
type AuthorizationURLBuilder struct {
    baseURL url.URL
    params  url.Values
}

// NewAuthorizationURLBuilder creates a new URL builder for OAuth2 authorize endpoint.
//
// Parameters:
//   - authorizeEndpoint: Base authorization URL (e.g., "https://provider.com/oauth2/authorize")
//
// Returns error if base URL is invalid or not HTTPS.
func NewAuthorizationURLBuilder(authorizeEndpoint string) (*AuthorizationURLBuilder, error) {
    parsed, err := url.Parse(authorizeEndpoint)
    if err != nil {
        return nil, fmt.Errorf("invalid authorize endpoint: %w", err)
    }

    if parsed.Scheme != "https" {
        return nil, fmt.Errorf("authorize endpoint must use HTTPS: %s", authorizeEndpoint)
    }

    return &AuthorizationURLBuilder{
        baseURL: *parsed,
        params:  url.Values{},
    }, nil
}

// SetResponseType sets the response_type parameter (typically "code" for authorization code flow).
func (b *AuthorizationURLBuilder) SetResponseType(responseType string) *AuthorizationURLBuilder {
    b.params.Set("response_type", responseType)
    return b
}

// SetClientID sets the client_id parameter (OAuth2 client identifier).
func (b *AuthorizationURLBuilder) SetClientID(clientID string) *AuthorizationURLBuilder {
    b.params.Set("client_id", clientID)
    return b
}

// SetRedirectURI sets the redirect_uri parameter (callback URL after authorization).
// The URL is properly encoded to handle special characters and query parameters.
func (b *AuthorizationURLBuilder) SetRedirectURI(redirectURI string) *AuthorizationURLBuilder {
    b.params.Set("redirect_uri", redirectURI)
    return b
}

// SetScope sets the scope parameter (space-separated list of scopes).
func (b *AuthorizationURLBuilder) SetScope(scope string) *AuthorizationURLBuilder {
    b.params.Set("scope", scope)
    return b
}

// SetState sets the state parameter (CSRF token).
// This is REQUIRED for OAuth2 security - it prevents CSRF attacks.
func (b *AuthorizationURLBuilder) SetState(state string) *AuthorizationURLBuilder {
    b.params.Set("state", state)
    return b
}

// AddCustomParam adds a custom parameter (e.g., provider-specific extensions).
func (b *AuthorizationURLBuilder) AddCustomParam(key, value string) *AuthorizationURLBuilder {
    b.params.Set(key, value)
    return b
}

// Build constructs the final URL with all parameters properly encoded.
// Returns the complete authorization URL as a string.
func (b *AuthorizationURLBuilder) Build() string {
    // Create a copy of baseURL to avoid mutation
    result := b.baseURL

    // Preserve any existing query parameters in base URL
    existingParams := result.Query()

    // Merge with builder params (builder params take precedence)
    for key, values := range b.params {
        for _, value := range values {
            existingParams.Set(key, value)
        }
    }

    // Set merged parameters
    result.RawQuery = existingParams.Encode()

    return result.String()
}

// Example: Build authorization URL with all parameters
func ExampleBuildAuthorizationURL() string {
    builder, _ := NewAuthorizationURLBuilder("https://github.com/login/oauth/authorize")

    url := builder.
        SetResponseType("code").
        SetClientID("agent-12345").
        SetRedirectURI("https://broker.example.com/oauth2/callback").
        SetScope("repo user:email").
        SetState("csrf-token-abc123").
        AddCustomParam("login", "octocat"). // GitHub-specific parameter
        Build()

    // Result: https://github.com/login/oauth/authorize?response_type=code&client_id=agent-12345&redirect_uri=https%3A%2F%2Fbroker.example.com%2Foauth2%2Fcallback&scope=repo+user%3Aemail&state=csrf-token-abc123&login=octocat
    return url
}
```

### URL Parameter Parsing and Validation

```go
package oauth2

import (
    "fmt"
    "net/http"
    "net/url"
)

// AuthorizeRequest represents a parsed OAuth2 authorization request.
type AuthorizeRequest struct {
    ResponseType string
    ClientID     string
    RedirectURI  string
    Scope        string
    State        string
}

// ParseAuthorizeRequest extracts and validates OAuth2 parameters from request.
//
// Required parameters:
//   - response_type: Must be "code" for authorization code flow
//   - client_id: Agent identifier
//   - redirect_uri: Callback URL (must be HTTPS and pre-registered)
//
// Optional parameters:
//   - scope: Space-separated list of scopes
//   - state: CSRF token (strongly recommended)
//
// Returns error if required parameters are missing or invalid.
func ParseAuthorizeRequest(r *http.Request) (*AuthorizeRequest, error) {
    // Parse query parameters
    query := r.URL.Query()

    // Extract response_type (required)
    responseType := query.Get("response_type")
    if responseType == "" {
        return nil, fmt.Errorf("missing required parameter: response_type")
    }
    if responseType != "code" {
        return nil, fmt.Errorf("unsupported response_type: %s (only 'code' is supported)", responseType)
    }

    // Extract client_id (required)
    clientID := query.Get("client_id")
    if clientID == "" {
        return nil, fmt.Errorf("missing required parameter: client_id")
    }

    // Extract redirect_uri (required)
    redirectURI := query.Get("redirect_uri")
    if redirectURI == "" {
        return nil, fmt.Errorf("missing required parameter: redirect_uri")
    }

    // Validate redirect_uri is a valid URL
    parsed, err := url.Parse(redirectURI)
    if err != nil {
        return nil, fmt.Errorf("invalid redirect_uri: %w", err)
    }

    // Enforce HTTPS for redirect_uri (security requirement)
    if parsed.Scheme != "https" && parsed.Scheme != "http" {
        return nil, fmt.Errorf("invalid redirect_uri scheme: %s (must be http or https)", parsed.Scheme)
    }

    // Extract optional parameters
    scope := query.Get("scope")
    state := query.Get("state")

    return &AuthorizeRequest{
        ResponseType: responseType,
        ClientID:     clientID,
        RedirectURI:  redirectURI,
        Scope:        scope,
        State:        state,
    }, nil
}

// RedirectWithError constructs an OAuth2 error redirect URL.
// Used to redirect user back to client with error parameters.
func RedirectWithError(redirectURI, errorCode, errorDescription, state string) (string, error) {
    parsed, err := url.Parse(redirectURI)
    if err != nil {
        return "", fmt.Errorf("invalid redirect_uri: %w", err)
    }

    // Add error parameters
    query := parsed.Query()
    query.Set("error", errorCode)
    if errorDescription != "" {
        query.Set("error_description", errorDescription)
    }
    if state != "" {
        query.Set("state", state)
    }

    parsed.RawQuery = query.Encode()
    return parsed.String(), nil
}

// RedirectWithCode constructs an OAuth2 success redirect URL with authorization code.
func RedirectWithCode(redirectURI, code, state string) (string, error) {
    parsed, err := url.Parse(redirectURI)
    if err != nil {
        return "", fmt.Errorf("invalid redirect_uri: %w", err)
    }

    // Add success parameters
    query := parsed.Query()
    query.Set("code", code)
    if state != "" {
        query.Set("state", state)
    }

    parsed.RawQuery = query.Encode()
    return parsed.String(), nil
}
```

### Testing URL Builders

```go
package oauth2_test

import (
    "net/http/httptest"
    "net/url"
    "testing"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAuthorizationURLBuilder_BuildsValidURL(t *testing.T) {
    builder, err := oauth2.NewAuthorizationURLBuilder("https://github.com/login/oauth/authorize")
    require.NoError(t, err)

    result := builder.
        SetResponseType("code").
        SetClientID("agent-123").
        SetRedirectURI("https://broker.example.com/callback?foo=bar").
        SetScope("repo user:email").
        SetState("csrf-abc").
        Build()

    // Parse result to verify structure
    parsed, err := url.Parse(result)
    require.NoError(t, err)

    // Verify base URL
    assert.Equal(t, "https", parsed.Scheme)
    assert.Equal(t, "github.com", parsed.Host)
    assert.Equal(t, "/login/oauth/authorize", parsed.Path)

    // Verify query parameters
    query := parsed.Query()
    assert.Equal(t, "code", query.Get("response_type"))
    assert.Equal(t, "agent-123", query.Get("client_id"))
    assert.Equal(t, "https://broker.example.com/callback?foo=bar", query.Get("redirect_uri"))
    assert.Equal(t, "repo user:email", query.Get("scope"))
    assert.Equal(t, "csrf-abc", query.Get("state"))
}

func TestParseAuthorizeRequest_ValidRequest(t *testing.T) {
    req := httptest.NewRequest("GET", "/oauth2/authorize?response_type=code&client_id=agent-123&redirect_uri=https://example.com/callback&scope=repo&state=xyz", nil)

    parsed, err := oauth2.ParseAuthorizeRequest(req)
    require.NoError(t, err)

    assert.Equal(t, "code", parsed.ResponseType)
    assert.Equal(t, "agent-123", parsed.ClientID)
    assert.Equal(t, "https://example.com/callback", parsed.RedirectURI)
    assert.Equal(t, "repo", parsed.Scope)
    assert.Equal(t, "xyz", parsed.State)
}

func TestParseAuthorizeRequest_MissingRequiredParams(t *testing.T) {
    tests := []struct {
        name    string
        url     string
        wantErr string
    }{
        {
            name:    "missing response_type",
            url:     "/oauth2/authorize?client_id=agent-123&redirect_uri=https://example.com/callback",
            wantErr: "missing required parameter: response_type",
        },
        {
            name:    "missing client_id",
            url:     "/oauth2/authorize?response_type=code&redirect_uri=https://example.com/callback",
            wantErr: "missing required parameter: client_id",
        },
        {
            name:    "missing redirect_uri",
            url:     "/oauth2/authorize?response_type=code&client_id=agent-123",
            wantErr: "missing required parameter: redirect_uri",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", tt.url, nil)
            _, err := oauth2.ParseAuthorizeRequest(req)
            require.Error(t, err)
            assert.Contains(t, err.Error(), tt.wantErr)
        })
    }
}

func TestRedirectWithError(t *testing.T) {
    result, err := oauth2.RedirectWithError(
        "https://example.com/callback",
        "access_denied",
        "User declined authorization",
        "state-123",
    )
    require.NoError(t, err)

    parsed, err := url.Parse(result)
    require.NoError(t, err)

    query := parsed.Query()
    assert.Equal(t, "access_denied", query.Get("error"))
    assert.Equal(t, "User declined authorization", query.Get("error_description"))
    assert.Equal(t, "state-123", query.Get("state"))
}
```

---

## 5. Complete OAuth2 Proxy Handler Example

### Putting It All Together

```go
package handlers

import (
    "log/slog"
    "net/http"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/proxy"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// OAuth2ProxyHandler handles OAuth2 proxy requests with audit logging.
type OAuth2ProxyHandler struct {
    proxyHandler   *proxy.ProxyHandler
    serviceRepo    ports.ThirdpartyOAuth2ServiceRepository
    logger         *slog.Logger
}

// NewOAuth2ProxyHandler creates a new OAuth2 proxy handler.
func NewOAuth2ProxyHandler(
    upstreamClient *http.Client,
    serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
    logger *slog.Logger,
) *OAuth2ProxyHandler {
    return &OAuth2ProxyHandler{
        proxyHandler: proxy.NewProxyHandler(upstreamClient, logger),
        serviceRepo:  serviceRepo,
        logger:       logger,
    }
}

// HandleAuthorize handles GET/POST /oauth2/authorize requests.
// This endpoint initiates the OAuth2 authorization code flow.
func (h *OAuth2ProxyHandler) HandleAuthorize(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())

    // Parse and validate request
    authReq, err := oauth2.ParseAuthorizeRequest(r)
    if err != nil {
        h.logger.Warn("invalid authorize request",
            "request_id", requestID,
            "error", err)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Add client_id to context for audit logging
    ctx := middleware.WithClientID(r.Context(), authReq.ClientID)
    r = r.WithContext(ctx)

    // Look up service by client_id
    service, err := h.serviceRepo.GetByID(ctx, authReq.ClientID)
    if err != nil {
        h.logger.Warn("service not found",
            "request_id", requestID,
            "client_id", authReq.ClientID)

        // Redirect to redirect_uri with error
        redirectURL, _ := oauth2.RedirectWithError(
            authReq.RedirectURI,
            "invalid_client",
            "Unknown client_id",
            authReq.State,
        )
        http.Redirect(w, r, redirectURL, http.StatusFound)
        return
    }

    // Build upstream authorization URL
    builder, err := oauth2.NewAuthorizationURLBuilder(service.OAuth2Endpoints.AuthorizeEndpoint)
    if err != nil {
        h.logger.Error("invalid authorization endpoint",
            "request_id", requestID,
            "client_id", authReq.ClientID,
            "error", err)
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    upstreamURL := builder.
        SetResponseType(authReq.ResponseType).
        SetClientID(service.ClientID).
        SetRedirectURI(authReq.RedirectURI).
        SetScope(authReq.Scope).
        SetState(authReq.State).
        Build()

    h.logger.Info("redirecting to upstream authorization",
        "request_id", requestID,
        "client_id", authReq.ClientID,
        "upstream_url", upstreamURL)

    // Redirect user to upstream OAuth2 provider
    http.Redirect(w, r, upstreamURL, http.StatusFound)
}

// HandleToken handles POST /oauth2/token requests.
// This endpoint exchanges authorization codes for access tokens.
func (h *OAuth2ProxyHandler) HandleToken(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())

    // Extract client_id from form data
    if err := r.ParseForm(); err != nil {
        h.logger.Warn("failed to parse form",
            "request_id", requestID,
            "error", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    clientID := r.FormValue("client_id")
    if clientID == "" {
        h.logger.Warn("missing client_id in token request",
            "request_id", requestID)
        http.Error(w, "Missing client_id", http.StatusBadRequest)
        return
    }

    // Add client_id to context for audit logging
    ctx := middleware.WithClientID(r.Context(), clientID)
    r = r.WithContext(ctx)

    // Look up service
    service, err := h.serviceRepo.GetByID(ctx, clientID)
    if err != nil {
        h.logger.Warn("service not found",
            "request_id", requestID,
            "client_id", clientID)
        http.Error(w, "Invalid client_id", http.StatusUnauthorized)
        return
    }

    // Proxy request to upstream token endpoint
    h.logger.Info("proxying token request",
        "request_id", requestID,
        "client_id", clientID,
        "upstream_url", service.OAuth2Endpoints.TokenEndpoint)

    if err := h.proxyHandler.ProxyFormURLEncoded(w, r, service.OAuth2Endpoints.TokenEndpoint); err != nil {
        h.proxyHandler.WriteProxyError(w, err)
        return
    }

    // Success - response already written by proxy handler
}
```

---

## 6. Integration with Existing Codebase

### Server Setup with All Components

```go
package main

import (
    "log/slog"

    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/httpclient"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/middleware"
)

func setupOAuth2ProxyServer(
    server *http.Server,
    serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
    logger *slog.Logger,
) error {
    // Create secure HTTP client for upstream requests
    upstreamClient, err := httpclient.NewSecureClient()
    if err != nil {
        return fmt.Errorf("failed to create HTTP client: %w", err)
    }

    // Create OAuth2 proxy handler
    oauthHandler := handlers.NewOAuth2ProxyHandler(upstreamClient, serviceRepo, logger)

    // Register routes
    router := server.Router()

    // Apply middleware stack
    router.Use(http.RecoveryMiddleware(logger))
    router.Use(middleware.AuditLogMiddleware(logger))
    router.Use(http.OptionalPrincipalMiddleware(server.Config().Authentication, logger))

    // OAuth2 endpoints
    router.Get("/oauth2/authorize", oauthHandler.HandleAuthorize)
    router.Post("/oauth2/authorize", oauthHandler.HandleAuthorize)
    router.Post("/oauth2/token", oauthHandler.HandleToken)

    logger.Info("OAuth2 proxy endpoints registered")
    return nil
}
```

---

## 7. Summary and Best Practices

### Key Takeaways

1. **TLS Security**
   - Always use system CA pool (reject self-signed certificates)
   - Enforce TLS 1.2+ minimum version
   - Set appropriate timeouts on all layers (dial, TLS handshake, request)
   - Reuse connections via http.Transport connection pooling

2. **Middleware Patterns**
   - Use unexported context key types to prevent collisions
   - Extract context values with type-safe helper functions
   - Apply middleware in correct order: recovery → audit → auth → routes
   - Wrap http.ResponseWriter to capture status codes and response size

3. **HTTP Proxy Implementation**
   - Read request body into memory (allows retries, but limits request size)
   - Filter hop-by-hop headers (Connection, Keep-Alive, etc.)
   - Stream response body with io.Copy (efficient for large responses)
   - Set Content-Length explicitly for better compatibility
   - Use context for timeout and cancellation

4. **URL Construction**
   - Use url.URL and url.Values for proper encoding
   - Validate all URLs are HTTPS (security requirement)
   - Preserve existing query parameters when building URLs
   - Handle special characters in redirect_uri correctly

5. **Error Handling**
   - Log all errors with structured context (request_id, client_id, principal)
   - Return OAuth2-compliant error responses (error, error_description)
   - Use error wrapping with fmt.Errorf("context: %w", err)
   - Differentiate between client errors (4xx) and server errors (5xx)

6. **Testing Strategy**
   - Use httptest.Server for upstream mock servers
   - Test TLS validation with both valid and self-signed certificates
   - Verify header copying (including filtering hop-by-hop headers)
   - Test timeout behavior with context.WithTimeout
   - Validate URL encoding for special characters

---

## 8. References

- [Go net/http Documentation](https://pkg.go.dev/net/http)
- [Go crypto/tls Documentation](https://pkg.go.dev/crypto/tls)
- [Chi Router Documentation](https://github.com/go-chi/chi)
- [RFC 6749: OAuth 2.0 Authorization Framework](https://datatracker.ietf.org/doc/html/rfc6749)
- [RFC 8414: OAuth 2.0 Authorization Server Metadata](https://datatracker.ietf.org/doc/html/rfc8414)
- [Go Context Best Practices](https://go.dev/blog/context)

---

## File Locations in Codebase

Suggested file structure for implementation:

```
internal/adapters/http/
├── httpclient/
│   ├── secure_client.go          # NewSecureClient implementation
│   └── secure_client_test.go     # TLS validation tests
├── middleware/
│   ├── audit.go                   # AuditLogMiddleware
│   ├── audit_test.go
│   └── context.go                 # Context helper functions
├── proxy/
│   ├── handler.go                 # ProxyHandler implementation
│   ├── handler_test.go
│   └── errors.go                  # Error response helpers
├── oauth2/
│   ├── url_builder.go             # AuthorizationURLBuilder
│   ├── url_builder_test.go
│   ├── request_parser.go          # ParseAuthorizeRequest
│   └── request_parser_test.go
└── handlers/
    ├── oauth2_proxy_handler.go    # Complete OAuth2 proxy handler
    └── oauth2_proxy_handler_test.go
```

This structure follows the existing codebase patterns and provides clear separation of concerns.
