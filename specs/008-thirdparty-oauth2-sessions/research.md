# Research: Third-Party OAuth2 Session Management

**Feature**: 008-thirdparty-oauth2-sessions  
**Date**: 2025-12-22  
**Purpose**: Resolve technical unknowns and document library best practices

## Table of Contents

1. [OAuth2 Library (golang.org/x/oauth2)](#oauth2-library)
2. [JWE/JWS Library (lestrrat-go/jwx)](#jwejws-library)
3. [Domain Service Design](#domain-service-design)
4. [State Token Security Patterns](#state-token-security-patterns)
5. [Token Storage Patterns](#token-storage-patterns)
6. [Retry and Error Handling](#retry-and-error-handling)

---

## OAuth2 Library

### Decision: Use `golang.org/x/oauth2` for OAuth2 Authorization Code Flow with PKCE

**Rationale**: 
- Official Go OAuth2 library maintained by the Go team
- Native support for PKCE (RFC 7636) since v0.13.0
- Used by 46,000+ projects, battle-tested
- Clean API for authorization code flow: `AuthCodeURL()`, `Exchange()`
- Built-in token refresh support via `TokenSource`

**Alternatives Considered**:
- `github.com/ory/fosite` - Full OAuth2 server implementation, too heavy for client-side flows
- Custom implementation - Forbidden by Constitution Principle III (Library-First Security)

### Key APIs

```go
import "golang.org/x/oauth2"

// Generate PKCE verifier (32-byte high-entropy random string)
verifier := oauth2.GenerateVerifier()

// Create OAuth2 config
config := &oauth2.Config{
    ClientID:     service.ClientID,
    ClientSecret: service.ClientSecret,
    Endpoint: oauth2.Endpoint{
        AuthURL:  service.Endpoints.AuthorizeEndpoint,
        TokenURL: service.Endpoints.TokenEndpoint,
    },
    RedirectURL: callbackURL,
    Scopes:      scopes,
}

// Generate authorization URL with PKCE challenge
authURL := config.AuthCodeURL(
    stateToken,                           // JWE state token
    oauth2.S256ChallengeOption(verifier), // PKCE challenge
    oauth2.AccessTypeOffline,             // Request refresh token
)

// Exchange authorization code for tokens (with PKCE verifier)
token, err := config.Exchange(
    ctx,
    authorizationCode,
    oauth2.VerifierOption(verifier), // PKCE verifier validation
)
```

### PKCE Flow Integration

1. **Authorize Endpoint**:
   - Generate PKCE verifier using `oauth2.GenerateVerifier()`
   - Compute S256 challenge using `oauth2.S256ChallengeOption()`
   - Include verifier in JWE state token (encrypted)
   - Redirect to third-party with challenge

2. **Callback Endpoint**:
   - Decrypt JWE state token
   - Extract PKCE verifier from token
   - Call `config.Exchange()` with `oauth2.VerifierOption(verifier)`
   - Library handles PKCE validation at token endpoint

### Error Handling

```go
// RetrieveError contains OAuth2 error details from token endpoint
var retrieveErr *oauth2.RetrieveError
if errors.As(err, &retrieveErr) {
    // retrieveErr.ErrorCode = "invalid_grant", "access_denied", etc.
    // retrieveErr.ErrorDescription = human-readable message
    log.Error("OAuth2 error", 
        "code", retrieveErr.ErrorCode,
        "description", retrieveErr.ErrorDescription)
}
```

---

## JWE/JWS Library

### Decision: Use `github.com/lestrrat-go/jwx/v3` for JWE State Tokens

**Rationale**:
- Complete JWE/JWS/JWK/JWT implementation (RFC 7515, 7516, 7517, 7518)
- 2,300+ stars, mature project (10 years)
- Supports authenticated encryption (A256GCM)
- Clean API with symmetric key encryption
- Used by 5,000+ projects

**Alternatives Considered**:
- `gopkg.in/square/go-jose.v2` - Valid alternative, but less active maintenance
- Custom JWE implementation - Forbidden by Constitution Principle III

### Key APIs

```go
import (
    "github.com/lestrrat-go/jwx/v3/jwa"
    "github.com/lestrrat-go/jwx/v3/jwe"
    "github.com/lestrrat-go/jwx/v3/jwk"
)

// Create symmetric key from configuration
key, err := jwk.FromRaw([]byte(signingKey))

// State token claims
type StateTokenClaims struct {
    Principal   string    `json:"principal"`
    PKCEVerifier string   `json:"pkce_verifier"`
    ServiceID   string    `json:"service_id"`
    RedirectURI string    `json:"redirect_uri"`
    IssuedAt    time.Time `json:"iat"`
    ExpiresAt   time.Time `json:"exp"`
}

// Encrypt (create JWE)
payload, _ := json.Marshal(claims)
encrypted, err := jwe.Encrypt(
    payload,
    jwe.WithKey(jwa.A256KW(), key),           // Key wrapping algorithm
    jwe.WithContentEncryption(jwa.A256GCM()), // Authenticated encryption
)

// Decrypt (validate JWE)
decrypted, err := jwe.Decrypt(
    encrypted,
    jwe.WithKey(jwa.A256KW(), key),
)
```

### Algorithm Choice

| Algorithm | Purpose | Spec |
|-----------|---------|------|
| `A256KW` | Key wrapping (AES-256 Key Wrap) | RFC 7518 §4.4 |
| `A256GCM` | Content encryption (AES-256-GCM authenticated encryption) | RFC 7518 §5.3 |

**Security Properties**:
- Confidentiality: Claims are encrypted, cannot be read by user/attacker
- Integrity: GCM provides authentication, tampering detected
- Key binding: Symmetric key in config, not exposed to clients

---

## Domain Service Design

### Decision: New `OAuth2SessionService` Domain Service

**Rationale**:
- Orchestrating OAuth2 flows is domain logic, not adapter logic
- Needs to coordinate multiple ports: storage, encryption, HTTP client
- Separates OAuth2 session concerns from existing consent service
- User's directive: "Orchestrating the OAuth2 flows is part of the domain"

**Location**: `internal/domain/oauth2session/`

### Service Interface

```go
package oauth2session

// Service orchestrates OAuth2 authorization flows for third-party services.
// It uses the standard golang.org/x/oauth2 library for flow execution
// and lestrrat-go/jwx for JWE state token handling.
type Service struct {
    sessionRepo     ports.UserSessionRepository
    serviceRepo     ports.ThirdpartyOAuth2ServiceRepository
    encryption      ports.EncryptionPort
    stateToken      StateTokenService
    httpClient      *http.Client
    config          OAuth2SessionConfig
}

// OAuth2SessionConfig holds configuration for the OAuth2 session service.
type OAuth2SessionConfig struct {
    JWESigningKey      []byte        // Symmetric key for JWE
    StateTokenTTL      time.Duration // Default: 10 minutes
    PKCEVerifierLength int           // Default: 32 bytes
    RetryConfig        RetryConfig   // Exponential backoff config
}

type RetryConfig struct {
    MaxAttempts int           // Default: 3
    BaseDelay   time.Duration // Default: 1 second
    MaxDelay    time.Duration // Default: 4 seconds
}
```

### Service Methods

```go
// InitiateOAuth2Flow starts an OAuth2 authorization code flow with PKCE.
// Returns the authorization URL to redirect the user to.
// 
// Validates:
// - Principal is authenticated
// - Service exists and is configured
// - Redirect URI matches request origin (same-origin validation)
//
// Creates JWE state token containing principal, PKCE verifier, service ID, redirect URI.
func (s *Service) InitiateOAuth2Flow(ctx context.Context, params InitiateFlowParams) (string, error)

// CompleteOAuth2Flow processes the OAuth2 callback after user authorization.
// Validates state token, exchanges code for tokens, stores encrypted tokens.
//
// Validates:
// - State token decrypts successfully
// - Principal in token matches current principal (CSRF protection)
// - Service ID in token matches callback parameter
// - PKCE verifier is valid
//
// On success, creates UserSession with encrypted tokens.
// On race condition (existing session), returns existing session.
func (s *Service) CompleteOAuth2Flow(ctx context.Context, params CompleteFlowParams) (*UserSession, error)

// ListUserSessions returns all sessions for a principal with status info.
// Includes service display name, agent count, expiration status.
func (s *Service) ListUserSessions(ctx context.Context, principal string) ([]UserSessionSummary, error)

// TerminateSession deletes a user's session with a third-party service.
// Returns affected agent count for confirmation dialog.
func (s *Service) TerminateSession(ctx context.Context, principal, serviceID string) error

// GetAffectedAgents returns agents that depend on a session (for termination warning).
func (s *Service) GetAffectedAgents(ctx context.Context, principal, serviceID string) ([]AffectedAgent, error)
```

### Dependency on OAuth2 and JWX Libraries in Domain

Per user directive: *"It is ok to use the jwx and oauth2 libraries in the domain."*

The domain service directly uses:
- `golang.org/x/oauth2` - For `oauth2.Config`, `GenerateVerifier()`, `Exchange()`
- `github.com/lestrrat-go/jwx/v3` - For `jwe.Encrypt()`, `jwe.Decrypt()`

This is acceptable because:
1. These are stable, well-defined interfaces (RFC-based)
2. No adapter abstraction would add value (would just proxy the same calls)
3. OAuth2 flow orchestration IS domain logic
4. Both libraries are vetted security implementations (Constitution Principle III)

### HTTP Client Abstraction

The HTTP client used for token exchange IS abstracted via dependency injection:

```go
// In domain service constructor
func NewService(opts ...Option) *Service {
    s := &Service{
        httpClient: http.DefaultClient, // Default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Option for custom HTTP client (testing, timeouts, tracing)
func WithHTTPClient(client *http.Client) Option {
    return func(s *Service) {
        s.httpClient = client
    }
}

// Usage in token exchange
func (s *Service) exchangeCode(ctx context.Context, config *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
    ctx = context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)
    return config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}
```

---

## State Token Security Patterns

### JWE Structure

```
JWE Compact Serialization:
BASE64URL(Header).BASE64URL(Encrypted Key).BASE64URL(IV).BASE64URL(Ciphertext).BASE64URL(Auth Tag)
```

### Claims Structure

```go
type OAuth2StateTokenClaims struct {
    // Principal of the user who initiated the flow
    Principal string `json:"sub"`
    
    // PKCE code verifier (high-entropy random string)
    PKCEVerifier string `json:"pkce_verifier"`
    
    // Service ID for the third-party OAuth2 service
    ServiceID string `json:"service_id"`
    
    // Original redirect URI from authorize request
    RedirectURI string `json:"redirect_uri"`
    
    // Token issue time
    IssuedAt int64 `json:"iat"`
    
    // Token expiration time (iat + TTL)
    ExpiresAt int64 `json:"exp"`
}
```

### Validation Rules (Fail-Closed)

| Check | Failure Response | Security Rationale |
|-------|------------------|-------------------|
| JWE decryption fails | 400 Bad Request | Tampered or corrupted token |
| Token expired | 400 Bad Request | Replay attack prevention |
| Principal mismatch | 403 Forbidden | CSRF attack prevention |
| Service ID mismatch | 400 Bad Request | Flow binding violation |

### Key Management

- Key loaded from environment variable: `IDENTITY_BROKER_JWE_SIGNING_KEY`
- Minimum length: 32 bytes (256 bits for A256KW)
- Never logged or exposed in configuration dumps
- Key rotation: Requires new tokens, old tokens fail decryption (acceptable for short TTL)
- If no key configured, generate one in memory and log a warning that this only works during development

---

## Token Storage Patterns

### Encryption Context

Following existing `EncryptionPort` pattern:

```go
encryptionContext := map[string]string{
    "principal":  session.Principal,
    "service_id": session.ServiceID,
}

encryptedAccessToken, err := s.encryption.Encrypt(
    ctx,
    []byte(token.AccessToken),
    encryptionContext,
)
```

The encryption context binds the ciphertext to the session:
- Decryption fails if context doesn't match
- Prevents token reuse across sessions
- Audit trail for encrypted data

### UserSession Entity

```go
type UserSession struct {
    ID                     string    `db:"id"`
    Principal              string    `db:"principal"`
    ServiceID              string    `db:"service_id"`
    EncryptedAccessToken   []byte    `db:"encrypted_access_token"`
    EncryptedRefreshToken  []byte    `db:"encrypted_refresh_token"` // Nullable
    TokenType              string    `db:"token_type"`
    AccessTokenExpiresAt   *time.Time `db:"access_token_expires_at"` // Nullable
    RefreshTokenExpiresAt  *time.Time `db:"refresh_token_expires_at"` // Nullable
    Scope                  []string  `db:"scope"`
    InitiatedAt            time.Time `db:"initiated_at"`
    CreatedAt              time.Time `db:"created_at"`
    UpdatedAt              time.Time `db:"updated_at"`
}
```

### Database Constraint for Race Conditions

```sql
CONSTRAINT uq_principal_service UNIQUE(principal, service_id)
```

On conflict (race condition):
1. First callback wins, inserts session
2. Subsequent callbacks get constraint violation
3. Query for existing session, return success

---

## Retry and Error Handling

### Token Exchange Retry Strategy

Per FR-021: Retry up to 3 times with exponential backoff (1s, 2s, 4s).

```go
type RetryConfig struct {
    MaxAttempts int           // 3
    BaseDelay   time.Duration // 1 * time.Second
    MaxDelay    time.Duration // 4 * time.Second
}

func (s *Service) exchangeWithRetry(ctx context.Context, config *oauth2.Config, code, verifier string) (*oauth2.Token, error) {
    var lastErr error
    
    for attempt := 1; attempt <= s.config.RetryConfig.MaxAttempts; attempt++ {
        token, err := s.exchangeCode(ctx, config, code, verifier)
        if err == nil {
            return token, nil
        }
        
        lastErr = err
        
        // Don't retry OAuth2 errors (these are definitive failures)
        var retrieveErr *oauth2.RetrieveError
        if errors.As(err, &retrieveErr) {
            return nil, s.translateOAuth2Error(retrieveErr)
        }
        
        // Calculate delay with exponential backoff
        delay := s.config.RetryConfig.BaseDelay * time.Duration(1<<(attempt-1))
        if delay > s.config.RetryConfig.MaxDelay {
            delay = s.config.RetryConfig.MaxDelay
        }
        
        log.Warn("Token exchange failed, retrying",
            "attempt", attempt,
            "max_attempts", s.config.RetryConfig.MaxAttempts,
            "delay", delay,
            "error", err,
        )
        
        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }
    
    return nil, fmt.Errorf("token exchange failed after %d attempts: %w", 
        s.config.RetryConfig.MaxAttempts, lastErr)
}
```

### Error Translation

```go
func (s *Service) translateOAuth2Error(err *oauth2.RetrieveError) error {
    switch err.ErrorCode {
    case "access_denied":
        return ErrUserDeniedAccess
    case "invalid_grant":
        return ErrInvalidGrant
    case "invalid_scope":
        return ErrInvalidScope
    default:
        return fmt.Errorf("oauth2 error: %s - %s", err.ErrorCode, err.ErrorDescription)
    }
}
```

---

## Summary of Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| OAuth2 library | `golang.org/x/oauth2` | Official Go library, PKCE support, battle-tested |
| JWE library | `github.com/lestrrat-go/jwx/v3` | Complete JOSE implementation, authenticated encryption |
| JWE algorithm | A256KW + A256GCM | 256-bit key wrapping + authenticated encryption |
| Domain service location | `internal/domain/oauth2session/` | Separates OAuth2 session logic from consent |
| Library usage in domain | Allowed | Per user directive, both are vetted RFC implementations |
| HTTP client | Injected dependency | Testability, timeout control, tracing |
| State token TTL | 10 minutes (configurable) | Short exposure window, sufficient for flow completion |
| Retry strategy | 3 attempts, exponential backoff | Network resilience without overwhelming provider |
