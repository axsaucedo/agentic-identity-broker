# Middleware Usage Guide

This guide explains how to use the principal extraction middleware in the Identity Broker for authenticating requests from a reverse proxy.

## Overview

The Identity Broker includes two middleware components for extracting and validating user principals from HTTP headers:

- **RequirePrincipalMiddleware**: Enforces authentication (rejects requests without valid principals)
- **OptionalPrincipalMiddleware**: Extracts principals optionally (never rejects requests)

Both middleware components are applied via chi router groups and provide request-scoped context access to the authenticated principal.

## Architecture

```
HTTP Request
    ↓
Recovery & Logging Middleware
    ↓
OptionalPrincipalMiddleware (global)
    ↓
Router Matches Route
    ↓
RequirePrincipalMiddleware (if applied to route group)
    ↓
Handler (accesses principal via context.Context)
    ↓
HTTP Response
```

## Configuration

Principals are extracted from an HTTP header configured per-server:

```yaml
server:
  enduser:
    port: 8000
    bind: "::"
    authentication:
      preauth:
        principal_header_name: X-Remote-User  # HTTP header to read
  admin:
    port: 14000
    bind: "::"
    authentication:
      preauth:
        principal_header_name: X-Remote-User
```

For detailed configuration options, see [Configuration Guide - Authentication Configuration](../configuration.md#authentication-configuration).

## Using Optional Principal Extraction (Global)

By default, all routes have OptionalPrincipalMiddleware applied globally. This extracts principals if present and valid, but never rejects requests.

### Example: Accessing Optional Principal

```go
import (
    "context"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// In an HTTP handler
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Check if principal is present
    if p, ok := principal.FromContext(r.Context()); ok {
        fmt.Printf("Authenticated as: %s\n", p)
    } else {
        fmt.Println("No principal (unauthenticated)")
    }
}
```

### Scenarios

**Request with principal header:**
```http
GET /api/public HTTP/1.1
Host: api.example.com
X-Remote-User: alice@example.com
```

Handler receives context with principal: `alice@example.com`

**Request without principal header:**
```http
GET /api/public HTTP/1.1
Host: api.example.com
```

Handler receives context without principal (optional)

**Request with oversized principal:**
```http
GET /api/public HTTP/1.1
Host: api.example.com
X-Remote-User: [201-character string]
```

Handler receives context without principal (too long, ignored in optional mode)

## Using Required Principal Extraction (Route Groups)

Use RequirePrincipalMiddleware on specific route groups to enforce authentication:

### Example: Protected Routes

```go
import (
    httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
    "github.com/go-chi/chi/v5"
)

func setupRoutes(router *chi.Mux, cfg ports.ServerInstanceConfig, logger *slog.Logger) {
    // Optional principal available globally

    // Protected routes require valid principal
    router.Group(func(r chi.Router) {
        // Apply required principal middleware to this group
        r.Use(httpAdapter.RequirePrincipalMiddleware(cfg.Authentication, logger))

        // All routes in this group require valid principal
        r.Get("/admin/users", handleListUsers)
        r.Post("/admin/settings", handleUpdateSettings)
        r.Delete("/admin/sessions/{id}", handleRevokeSession)
    })

    // Public routes (optional principal via global middleware)
    router.Get("/health", handleHealth)
    router.Get("/info", handleInfo)
}
```

### Request Handling

**Valid Request** → 200 OK, handler invoked:
```http
GET /admin/users HTTP/1.1
X-Remote-User: alice@example.com
```

**Missing Header** → 401 Unauthorized:
```http
GET /admin/users HTTP/1.1
```

Response:
```json
{
  "error": "missing or empty principal"
}
```

**Empty Header** → 401 Unauthorized:
```http
GET /admin/users HTTP/1.1
X-Remote-User:
```

Response:
```json
{
  "error": "missing or empty principal"
}
```

**Whitespace-Only Header** → 401 Unauthorized:
```http
GET /admin/users HTTP/1.1
X-Remote-User:
```

Response (after trimming):
```json
{
  "error": "missing or empty principal"
}
```

**Oversized Principal** → 400 Bad Request:
```http
GET /admin/users HTTP/1.1
X-Remote-User: [201+ character string]
```

Response:
```json
{
  "error": "principal exceeds maximum length of 200 characters (got 201)"
}
```

## Accessing Principal in Handlers

### Getting the Principal (With Type Safety)

```go
import "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"

func MyProtectedHandler(w http.ResponseWriter, r *http.Request) {
    p, ok := principal.FromContext(r.Context())
    if !ok {
        // This should never happen in a protected route with RequirePrincipalMiddleware
        http.Error(w, "Principal not found", http.StatusInternalServerError)
        return
    }

    // p is guaranteed to be non-empty and <= 200 characters
    fmt.Printf("User: %s\n", p)
}
```

### Getting the Principal (With Panic)

Use when you're absolutely certain the principal exists (e.g., inside protected routes):

```go
func MyProtectedHandler(w http.ResponseWriter, r *http.Request) {
    p := principal.MustFromContext(r.Context())
    fmt.Printf("User: %s\n", p)  // p is guaranteed to be present
}
```

### Handling Missing Principal (Optional Routes)

```go
func MyOptionalHandler(w http.ResponseWriter, r *http.Request) {
    resp := map[string]interface{}{}

    if p, ok := principal.FromContext(r.Context()); ok {
        resp["authenticated_as"] = p
        resp["authenticated"] = true
    } else {
        resp["authenticated"] = false
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

## Principal Validation Rules

The middleware enforces these validation rules:

| Scenario | Status | Error | Notes |
|----------|--------|-------|-------|
| Valid principal | Allowed | None | "alice", "user@example.com", "用户", "🚀" |
| Missing header | 401 | missing or empty principal | RequirePrincipalMiddleware only |
| Empty header value | 401 | missing or empty principal | After trimming: "" |
| Whitespace-only | 401 | missing or empty principal | After trimming: "   " → "" |
| Principal > 200 chars | 400 | principal exceeds maximum length | Exact length included in error |
| Leading/trailing spaces | OK | None | Automatically trimmed to "alice" |
| Unicode characters | OK | None | UTF-8 preserved: "用户", "🚀" |
| Internal spaces | OK | None | "alice smith" preserved |

## Error Handling

All errors from principal middleware return JSON responses with consistent structure:

```json
{
  "error": "description of what went wrong"
}
```

### Error Types

**401 Unauthorized** (RequirePrincipalMiddleware):
- Missing header: `"missing or empty principal"`
- Empty value: `"missing or empty principal"`
- Whitespace-only: `"missing or empty principal"`

**400 Bad Request** (RequirePrincipalMiddleware):
- Too long: `"principal exceeds maximum length of 200 characters (got 201)"`

**200 OK** (OptionalPrincipalMiddleware):
- Always succeeds (never returns error)
- Invalid principals silently ignored

## Custom Headers

Configure custom principal header names for your reverse proxy:

### Development with Custom Header

```yaml
server:
  enduser:
    authentication:
      preauth:
        principal_header_name: X-Authenticated-User
```

### Environment-Specific via Environment Variable

```yaml
server:
  enduser:
    authentication:
      preauth:
        principal_header_name: ${IDENTITY_BROKER_PRINCIPAL_HEADER}
```

Then:
```bash
export IDENTITY_BROKER_PRINCIPAL_HEADER="X-Custom-Header"
```

### Reverse Proxy Examples

**Nginx:**
```nginx
proxy_set_header X-Remote-User $remote_user;
```

**Traefik:**
```toml
[entryPoints.web]
    [entryPoints.web.forwardedHeaders]
        trustedIPs = ["127.0.0.1"]

[backends.backend1]
    [backends.backend1.servers.server1]
        url = "http://agentic-identity-broker:8000"
```

**HAProxy:**
```haproxy
http-request set-header X-Remote-User %[req.hdr(Authorization),extract(7)]
```

## Performance Characteristics

- **Context operations**: <100ns (per operation)
- **Middleware execution**: ~500ns-1µs
- **Total added latency**: <1ms at p95

For detailed benchmarks, see performance test results in `internal/adapters/http/principal_middleware_bench_test.go`.

## Security Considerations

### Trust Boundary

The middleware **trusts** the HTTP header value implicitly. The security model assumes:

1. **Reverse proxy is trusted**: Only a trusted reverse proxy should set the principal header
2. **Network is secured**: Use TLS/HTTPS for all communication
3. **Header cannot be spoofed**: The Identity Broker must not be exposed directly to untrusted networks

### Never Expose Directly

❌ **WRONG** - Direct exposure:
```
User → Identity Broker (port 8000)
```

✅ **CORRECT** - Via trusted reverse proxy:
```
User → Reverse Proxy (nginx/Traefik) → Identity Broker (port 8000)
```

### Validation Layers

1. **Header extraction**: Read from configured header
2. **Trimming**: Remove leading/trailing whitespace
3. **Length validation**: Reject > 200 characters
4. **Type safety**: Use unexported context key type
5. **Context scoping**: Per-request isolation

## Integration with Chi Router

The middleware integrates seamlessly with chi's router and middleware patterns:

### Basic Setup

```go
router := chi.NewRouter()

// Global middleware (applies to all routes)
router.Use(RecoveryMiddleware(logger))
router.Use(LoggingMiddleware(logger))
router.Use(OptionalPrincipalMiddleware(authConfig, logger))

// Protected group
router.Group(func(r chi.Router) {
    r.Use(RequirePrincipalMiddleware(authConfig, logger))
    r.Get("/admin/users", handleUsers)
})

// Public routes
router.Get("/health", handleHealth)
```

### Nested Groups

```go
router.Group(func(r chi.Router) {
    r.Use(RequirePrincipalMiddleware(authConfig, logger))

    r.Get("/admin/users", handleUsers)

    // Nested group inherits parent middleware
    r.Group(func(r chi.Router) {
        r.Use(AdminOnlyMiddleware())
        r.Delete("/admin/system/restart", handleSystemRestart)
    })
})
```

## Testing

### Unit Tests

For unit test examples, see `internal/adapters/http/principal_middleware_test.go`:
- 40+ test scenarios
- Valid principals (simple, email, unicode, emoji)
- Invalid principals (missing, empty, oversized)
- Custom headers
- Whitespace handling

### Integration Tests

For integration test examples, see `tests/integration/principal_middleware_test.go`:
- End-to-end chi router tests
- Real HTTP server testing
- Multiple middleware combinations
- Unicode support verification

### Manual Testing

```bash
# Valid principal
curl -H "X-Remote-User: alice" http://localhost:8000/admin/users

# Missing header (protected route returns 401)
curl http://localhost:8000/admin/users

# Oversized principal (returns 400)
curl -H "X-Remote-User: $(python3 -c 'print(\"a\"*201)')" http://localhost:8000/admin/users

# Optional route (always succeeds)
curl http://localhost:8000/health
```

## Troubleshooting

### Principal Not Found

**Symptom**: Handler receives no principal even with header present

**Causes**:
1. Header name mismatch (configured vs. sent)
2. Principal header not set by reverse proxy
3. Using wrong context key

**Solution**:
- Verify header name in config matches reverse proxy
- Check middleware is applied to route
- Use `principal.FromContext()` to retrieve

### Always Getting 401

**Symptom**: All protected routes return 401 even with valid header

**Causes**:
1. Reverse proxy not setting header
2. Header name misconfiguration
3. Principal validation rejecting valid values

**Solution**:
- Check logs for "Missing or empty principal" message
- Verify header with `curl -v`
- Check configuration value

### Middleware Not Applied

**Symptom**: Principal not available in handler despite middleware setup

**Causes**:
1. Middleware not registered in router
2. Wrong middleware type used
3. Route not covered by middleware group

**Solution**:
- Verify `router.Use()` or `r.Use()` called
- Check route is in correct group
- Use `router.Group()` for selective application

## Further Reading

- [Configuration Guide](../configuration.md#authentication-configuration)
- [Architecture Documentation](../ARCHITECTURE.md#session-management-domain)
- [Principal Domain Package](../../internal/domain/principal/README.md)
