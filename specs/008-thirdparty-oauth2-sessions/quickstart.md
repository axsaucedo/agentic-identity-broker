# Quickstart Guide: Third-Party OAuth2 Session Management

**Feature**: 008-thirdparty-oauth2-sessions  
**Audience**: Developers implementing OAuth2 session functionality  
**Purpose**: Step-by-step guide for implementing the feature

## Table of Contents

1. [Overview](#overview)
2. [Dependencies Setup](#dependencies-setup)
3. [Implementing the Domain Service](#implementing-the-domain-service)
4. [Implementing the State Token Handler](#implementing-the-state-token-handler)
5. [Implementing the Repository](#implementing-the-repository)
6. [Implementing HTTP Handlers](#implementing-http-handlers)
7. [Configuration Integration](#configuration-integration)
8. [Testing Patterns](#testing-patterns)
9. [Frontend Integration](#frontend-integration)

---

## Overview

This feature adds OAuth2 session management for third-party services:

```
┌─────────────┐      ┌─────────────────────┐      ┌──────────────────┐
│   Frontend  │      │   Identity Broker   │      │  Third-Party     │
│    (SPA)    │      │   (Go Backend)      │      │  OAuth2 Service  │
└──────┬──────┘      └──────────┬──────────┘      └────────┬─────────┘
       │                        │                          │
       │ 1. Click "Login"       │                          │
       │───────────────────────>│                          │
       │                        │                          │
       │ 2. Redirect to authorize endpoint                 │
       │<───────────────────────│                          │
       │                        │                          │
       │ 3. User authorizes at third-party                 │
       │──────────────────────────────────────────────────>│
       │                        │                          │
       │ 4. Redirect to callback with code                 │
       │<──────────────────────────────────────────────────│
       │                        │                          │
       │ 5. Forward callback to broker                     │
       │───────────────────────>│                          │
       │                        │                          │
       │                        │ 6. Exchange code for tokens
       │                        │─────────────────────────>│
       │                        │                          │
       │                        │ 7. Receive tokens        │
       │                        │<─────────────────────────│
       │                        │                          │
       │                        │ 8. Encrypt & store tokens│
       │                        │                          │
       │ 9. Redirect to sessions page with success         │
       │<───────────────────────│                          │
```

### Architecture

```
internal/
├── domain/oauth2session/
│   ├── service.go              # OAuth2SessionService (orchestrates flows)
│   ├── state_token.go          # JWE state token creation/validation
│   ├── errors.go               # Domain-specific errors
│   └── service_test.go
├── ports/
│   ├── storage.go              # + UserSessionRepository interface
│   └── oauth2.go               # StateTokenPort interface (optional)
└── adapters/
    ├── http/
    │   └── oauth2_session_handlers.go
    └── storage/
        ├── memory/user_session.go
        └── postgres/user_session.go
```

---

## Dependencies Setup

### Add Go Dependencies

```bash
# OAuth2 library for authorization code flow with PKCE
go get golang.org/x/oauth2

# JWX library for JWE state tokens
go get github.com/lestrrat-go/jwx/v3
```

### Verify go.mod

```go
require (
    golang.org/x/oauth2 v0.34.0
    github.com/lestrrat-go/jwx/v3 v3.0.12
    // ... existing dependencies
)
```

---

## Implementing the Domain Service

### Step 1: Create Service Structure

Location: `internal/domain/oauth2session/service.go`

```go
package oauth2session

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "golang.org/x/oauth2"
    "github.com/google/uuid"
    
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Service orchestrates OAuth2 authorization flows for third-party services.
type Service struct {
    sessionRepo  ports.UserSessionRepository
    serviceRepo  ports.ThirdpartyOAuth2ServiceRepository
    grantRepo    ports.UserGrantRepository
    encryption   ports.EncryptionPort
    stateToken   *StateTokenService
    httpClient   *http.Client
    config       Config
}

// Config holds configuration for the OAuth2 session service.
type Config struct {
    CallbackURLTemplate string        // e.g., "https://broker.example.com/api/third-party/%s/oauth2/callback"
    StateTokenTTL       time.Duration // Default: 10 minutes
    RetryMaxAttempts    int           // Default: 3
    RetryBaseDelay      time.Duration // Default: 1 second
}

// NewService creates a new OAuth2 session service.
func NewService(
    sessionRepo ports.UserSessionRepository,
    serviceRepo ports.ThirdpartyOAuth2ServiceRepository,
    grantRepo ports.UserGrantRepository,
    encryption ports.EncryptionPort,
    jweKey []byte,
    config Config,
) *Service {
    if config.StateTokenTTL == 0 {
        config.StateTokenTTL = 10 * time.Minute
    }
    if config.RetryMaxAttempts == 0 {
        config.RetryMaxAttempts = 3
    }
    if config.RetryBaseDelay == 0 {
        config.RetryBaseDelay = time.Second
    }

    return &Service{
        sessionRepo:  sessionRepo,
        serviceRepo:  serviceRepo,
        grantRepo:    grantRepo,
        encryption:   encryption,
        stateToken:   NewStateTokenService(jweKey, config.StateTokenTTL),
        httpClient:   &http.Client{Timeout: 30 * time.Second},
        config:       config,
    }
}
```

### Step 2: Implement InitiateOAuth2Flow

```go
// InitiateFlowParams contains parameters for initiating an OAuth2 flow.
type InitiateFlowParams struct {
    Principal   string
    ServiceID   string
    RedirectURI string
    RequestHost string // For same-origin validation
}

// InitiateOAuth2Flow starts an OAuth2 authorization code flow with PKCE.
func (s *Service) InitiateOAuth2Flow(ctx context.Context, params InitiateFlowParams) (string, error) {
    // Validate principal
    if params.Principal == "" {
        return "", ErrUnauthenticated
    }

    // Validate redirect_uri matches request host (same-origin)
    if err := validateRedirectURI(params.RedirectURI, params.RequestHost); err != nil {
        return "", fmt.Errorf("%w: %v", ErrInvalidRedirectURI, err)
    }

    // Load service configuration
    service, err := s.serviceRepo.Get(ctx, params.ServiceID)
    if err != nil {
        return "", fmt.Errorf("failed to load service: %w", err)
    }

    // Generate PKCE verifier using oauth2 library
    verifier := oauth2.GenerateVerifier()

    // Create JWE state token
    stateToken, err := s.stateToken.Create(StateTokenClaims{
        Principal:    params.Principal,
        PKCEVerifier: verifier,
        ServiceID:    params.ServiceID,
        RedirectURI:  params.RedirectURI,
    })
    if err != nil {
        return "", fmt.Errorf("failed to create state token: %w", err)
    }

    // Build OAuth2 config
    callbackURL := fmt.Sprintf(s.config.CallbackURLTemplate, params.ServiceID)
    config := &oauth2.Config{
        ClientID:     service.ClientID,
        ClientSecret: service.ClientSecret,
        Endpoint: oauth2.Endpoint{
            AuthURL:  service.Endpoints.AuthorizeEndpoint,
            TokenURL: service.Endpoints.TokenEndpoint,
        },
        RedirectURL: callbackURL,
        Scopes:      extractScopeValues(service.Scopes),
    }

    // Generate authorization URL with PKCE challenge
    authURL := config.AuthCodeURL(
        stateToken,
        oauth2.S256ChallengeOption(verifier),
        oauth2.AccessTypeOffline, // Request refresh token
    )

    return authURL, nil
}

func extractScopeValues(scopes []storage.OAuthScope) []string {
    values := make([]string, len(scopes))
    for i, s := range scopes {
        values[i] = s.ScopeValue
    }
    return values
}
```

### Step 3: Implement CompleteOAuth2Flow

```go
// CompleteFlowParams contains parameters from the OAuth2 callback.
type CompleteFlowParams struct {
    Principal        string
    ServiceID        string
    Code             string
    State            string
    Error            string // OAuth2 error from third-party
    ErrorDescription string
}

// CompleteOAuth2Flow processes the OAuth2 callback.
func (s *Service) CompleteOAuth2Flow(ctx context.Context, params CompleteFlowParams) (*storage.UserSession, error) {
    // Check for OAuth2 error from third-party
    if params.Error != "" {
        return nil, &OAuth2Error{
            Code:        params.Error,
            Description: params.ErrorDescription,
        }
    }

    // Validate and decrypt state token
    claims, err := s.stateToken.Validate(params.State)
    if err != nil {
        // Log security event
        s.logSecurityEvent("state_token_validation_failed", map[string]string{
            "service_id": params.ServiceID,
            "reason":     err.Error(),
        })
        return nil, ErrInvalidStateToken
    }

    // Validate principal matches (CSRF protection)
    if claims.Principal != params.Principal {
        s.logSecurityEvent("principal_mismatch", map[string]string{
            "service_id":      params.ServiceID,
            "expected":        claims.Principal,
            "actual":          params.Principal,
        })
        return nil, ErrPrincipalMismatch
    }

    // Validate service ID matches
    if claims.ServiceID != params.ServiceID {
        return nil, ErrServiceIDMismatch
    }

    // Load service configuration
    service, err := s.serviceRepo.Get(ctx, params.ServiceID)
    if err != nil {
        return nil, fmt.Errorf("failed to load service: %w", err)
    }

    // Build OAuth2 config
    callbackURL := fmt.Sprintf(s.config.CallbackURLTemplate, params.ServiceID)
    config := &oauth2.Config{
        ClientID:     service.ClientID,
        ClientSecret: service.ClientSecret,
        Endpoint: oauth2.Endpoint{
            AuthURL:  service.Endpoints.AuthorizeEndpoint,
            TokenURL: service.Endpoints.TokenEndpoint,
        },
        RedirectURL: callbackURL,
    }

    // Exchange code for tokens with retry
    token, err := s.exchangeWithRetry(ctx, config, params.Code, claims.PKCEVerifier)
    if err != nil {
        return nil, err
    }

    // Create session
    session, err := s.createSession(ctx, params.Principal, params.ServiceID, token)
    if err != nil {
        return nil, err
    }

    // Log success event
    s.logAuditEvent("session_established", map[string]string{
        "session_id": session.ID,
        "service_id": session.ServiceID,
        "principal":  session.Principal,
    })

    return session, nil
}

func (s *Service) createSession(
    ctx context.Context,
    principal, serviceID string,
    token *oauth2.Token,
) (*storage.UserSession, error) {
    sessionID := uuid.New().String()
    
    encryptionContext := map[string]string{
        "principal":  principal,
        "service_id": serviceID,
    }

    // Encrypt access token
    encryptedAccessToken, err := s.encryption.Encrypt(
        ctx,
        []byte(token.AccessToken),
        encryptionContext,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to encrypt access token: %w", err)
    }

    // Encrypt refresh token if present
    var encryptedRefreshToken []byte
    if token.RefreshToken != "" {
        encryptedRefreshToken, err = s.encryption.Encrypt(
            ctx,
            []byte(token.RefreshToken),
            encryptionContext,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
        }
    }

    session := &storage.UserSession{
        ID:                    sessionID,
        Principal:             principal,
        ServiceID:             serviceID,
        EncryptedAccessToken:  encryptedAccessToken,
        EncryptedRefreshToken: encryptedRefreshToken,
        TokenType:             token.TokenType,
        AccessTokenExpiresAt:  timePtr(token.Expiry),
        Scope:                 extractScopes(token),
        InitiatedAt:           time.Now(),
        EncryptionContext:     encryptionContext,
    }

    // Create session (handles race condition via upsert)
    created, err := s.sessionRepo.Create(ctx, session)
    if err != nil {
        return nil, fmt.Errorf("failed to store session: %w", err)
    }

    return created, nil
}
```

### Step 4: Implement Token Exchange with Retry

```go
func (s *Service) exchangeWithRetry(
    ctx context.Context,
    config *oauth2.Config,
    code, verifier string,
) (*oauth2.Token, error) {
    var lastErr error
    maxDelay := 4 * time.Second

    for attempt := 1; attempt <= s.config.RetryMaxAttempts; attempt++ {
        // Set custom HTTP client on context
        ctx := context.WithValue(ctx, oauth2.HTTPClient, s.httpClient)
        
        token, err := config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
        if err == nil {
            return token, nil
        }

        lastErr = err

        // Don't retry OAuth2 errors (these are definitive)
        var retrieveErr *oauth2.RetrieveError
        if errors.As(err, &retrieveErr) {
            return nil, &OAuth2Error{
                Code:        retrieveErr.ErrorCode,
                Description: retrieveErr.ErrorDescription,
            }
        }

        // Calculate backoff delay
        delay := s.config.RetryBaseDelay * time.Duration(1<<(attempt-1))
        if delay > maxDelay {
            delay = maxDelay
        }

        select {
        case <-ctx.Done():
            return nil, ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }

    return nil, fmt.Errorf("token exchange failed after %d attempts: %w",
        s.config.RetryMaxAttempts, lastErr)
}
```

---

## Implementing the State Token Handler

Location: `internal/domain/oauth2session/state_token.go`

```go
package oauth2session

import (
    "encoding/json"
    "errors"
    "time"

    "github.com/lestrrat-go/jwx/v3/jwa"
    "github.com/lestrrat-go/jwx/v3/jwe"
    "github.com/lestrrat-go/jwx/v3/jwk"
)

// StateTokenService handles JWE state token creation and validation.
type StateTokenService struct {
    key jwk.Key
    ttl time.Duration
}

// StateTokenClaims contains the claims in a state token.
type StateTokenClaims struct {
    Principal    string `json:"sub"`
    PKCEVerifier string `json:"pkce_verifier"`
    ServiceID    string `json:"service_id"`
    RedirectURI  string `json:"redirect_uri"`
    IssuedAt     int64  `json:"iat"`
    ExpiresAt    int64  `json:"exp"`
}

// NewStateTokenService creates a new state token service.
func NewStateTokenService(key []byte, ttl time.Duration) *StateTokenService {
    symmetricKey, _ := jwk.FromRaw(key)
    return &StateTokenService{
        key: symmetricKey,
        ttl: ttl,
    }
}

// Create creates a new JWE state token.
func (s *StateTokenService) Create(claims StateTokenClaims) (string, error) {
    now := time.Now()
    claims.IssuedAt = now.Unix()
    claims.ExpiresAt = now.Add(s.ttl).Unix()

    payload, err := json.Marshal(claims)
    if err != nil {
        return "", err
    }

    encrypted, err := jwe.Encrypt(
        payload,
        jwe.WithKey(jwa.A256KW(), s.key),
        jwe.WithContentEncryption(jwa.A256GCM()),
    )
    if err != nil {
        return "", err
    }

    return string(encrypted), nil
}

// Validate validates and decrypts a JWE state token.
func (s *StateTokenService) Validate(token string) (*StateTokenClaims, error) {
    decrypted, err := jwe.Decrypt(
        []byte(token),
        jwe.WithKey(jwa.A256KW(), s.key),
    )
    if err != nil {
        return nil, errors.New("failed to decrypt state token")
    }

    var claims StateTokenClaims
    if err := json.Unmarshal(decrypted, &claims); err != nil {
        return nil, errors.New("failed to parse state token claims")
    }

    // Check expiration
    if time.Now().Unix() > claims.ExpiresAt {
        return nil, errors.New("state token expired")
    }

    // Validate required claims
    if claims.Principal == "" || claims.PKCEVerifier == "" || claims.ServiceID == "" {
        return nil, errors.New("missing required claims")
    }

    return &claims, nil
}
```

---

## Implementing the Repository

### Step 1: Add Interface to Ports

Location: `internal/ports/storage.go`

```go
// UserSessionRepository defines storage operations for user OAuth2 sessions.
type UserSessionRepository interface {
    Create(ctx context.Context, session *storage.UserSession) (*storage.UserSession, error)
    Get(ctx context.Context, id string) (*storage.UserSession, error)
    FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error)
    Delete(ctx context.Context, id string) error
    DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error
    ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error)
    CountByService(ctx context.Context, serviceID string) (int, error)
}
```

### Step 2: Implement PostgreSQL Adapter

Location: `internal/adapters/storage/postgres/user_session.go`

```go
package postgres

import (
    "context"
    "database/sql"
    "errors"

    "github.com/jmoiron/sqlx"
    "github.com/lib/pq"
    
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
)

// Create creates a new user session, handling race conditions via ON CONFLICT.
func (a *Adapter) CreateUserSession(ctx context.Context, session *storage.UserSession) (*storage.UserSession, error) {
    query := `
        INSERT INTO user_sessions (
            id, principal, service_id, 
            encrypted_access_token, encrypted_refresh_token,
            token_type, access_token_expires_at, refresh_token_expires_at,
            scope, initiated_at, encryption_context,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()
        )
        ON CONFLICT (principal, service_id) DO NOTHING
        RETURNING *`

    var result storage.UserSession
    err := a.db.QueryRowxContext(ctx, query,
        session.ID, session.Principal, session.ServiceID,
        session.EncryptedAccessToken, session.EncryptedRefreshToken,
        session.TokenType, session.AccessTokenExpiresAt, session.RefreshTokenExpiresAt,
        pq.Array(session.Scope), session.InitiatedAt, session.EncryptionContext,
    ).StructScan(&result)

    if errors.Is(err, sql.ErrNoRows) {
        // Conflict: session already exists, fetch existing
        return a.FindUserSessionByPrincipalAndService(ctx, session.Principal, session.ServiceID)
    }
    if err != nil {
        return nil, storage.NewStorageError("CreateUserSession", storage.ErrorKindConnection, err, err.Error())
    }

    return &result, nil
}

// DeleteByPrincipalAndService deletes session for a principal and service.
func (a *Adapter) DeleteUserSessionByPrincipalAndService(ctx context.Context, principal, serviceID string) error {
    query := `DELETE FROM user_sessions WHERE principal = $1 AND service_id = $2`
    _, err := a.db.ExecContext(ctx, query, principal, serviceID)
    if err != nil {
        return storage.NewStorageError("DeleteUserSession", storage.ErrorKindConnection, err, err.Error())
    }
    return nil
}
```

---

## Implementing HTTP Handlers

Location: `internal/adapters/http/oauth2_session_handlers.go`

```go
package http

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
    "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/principal"
)

// OAuth2SessionHandler handles OAuth2 session endpoints.
type OAuth2SessionHandler struct {
    service *oauth2session.Service
}

// RegisterRoutes registers OAuth2 session routes.
func (h *OAuth2SessionHandler) RegisterRoutes(r chi.Router) {
    r.Get("/api/third-party/sessions", h.ListSessions)
    r.Get("/api/third-party/{serviceId}/oauth2/authorize", h.InitiateFlow)
    r.Get("/api/third-party/{serviceId}/oauth2/callback", h.HandleCallback)
    r.Delete("/api/third-party/{serviceId}/session", h.TerminateSession)
    r.Get("/api/third-party/{serviceId}/session/affected-agents", h.GetAffectedAgents)
}

// InitiateFlow handles GET /api/third-party/{serviceId}/oauth2/authorize
func (h *OAuth2SessionHandler) InitiateFlow(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Get principal from context (set by middleware)
    p, ok := principal.FromContext(ctx)
    if !ok {
        respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
        return
    }

    serviceID := chi.URLParam(r, "serviceId")
    redirectURI := r.URL.Query().Get("redirect_uri")

    if redirectURI == "" {
        respondError(w, http.StatusBadRequest, "missing_redirect_uri", "redirect_uri is required")
        return
    }

    authURL, err := h.service.InitiateOAuth2Flow(ctx, oauth2session.InitiateFlowParams{
        Principal:   p,
        ServiceID:   serviceID,
        RedirectURI: redirectURI,
        RequestHost: r.Host,
    })
    if err != nil {
        handleServiceError(w, err)
        return
    }

    http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback handles GET /api/third-party/{serviceId}/oauth2/callback
func (h *OAuth2SessionHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    p, ok := principal.FromContext(ctx)
    if !ok {
        respondError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
        return
    }

    serviceID := chi.URLParam(r, "serviceId")
    code := r.URL.Query().Get("code")
    state := r.URL.Query().Get("state")
    oauthError := r.URL.Query().Get("error")
    errorDesc := r.URL.Query().Get("error_description")

    session, err := h.service.CompleteOAuth2Flow(ctx, oauth2session.CompleteFlowParams{
        Principal:        p,
        ServiceID:        serviceID,
        Code:             code,
        State:            state,
        Error:            oauthError,
        ErrorDescription: errorDesc,
    })

    // Extract redirect_uri from state token for redirect
    redirectURI := h.extractRedirectURI(state)
    if redirectURI == "" {
        redirectURI = "/consent/sessions" // Fallback
    }

    if err != nil {
        // Redirect with error, in real code use the url class to append the new query parameters
        http.Redirect(w, r, redirectURI+"?error="+url.QueryEscape(err.Error()), http.StatusFound)
        return
    }

    // Redirect with success, in real code use the url class to append the new query parameters
    http.Redirect(w, r, redirectURI+"?success=true&session_id="+session.ID, http.StatusFound)
}
```

---

## Configuration Integration

### Add to Config Schema

Location: `internal/config/schema.go`

```go
type Config struct {
    // ... existing fields ...
    
    ThirdPartyOAuth2 ThirdPartyOAuth2Config `mapstructure:"third_party_oauth2"`
}

type ThirdPartyOAuth2Config struct {
    JWESigningKey      string `mapstructure:"jwe_signing_key"`
    StateTokenTTL      int    `mapstructure:"state_token_ttl"`       // seconds, default: 600
    PKCEVerifierLength int    `mapstructure:"pkce_verifier_length"`  // bytes, default: 32
}
```

### Example Configuration

Location: `examples/config/third-party-oauth2.yaml`

```yaml
# Third-party OAuth2 Session Configuration
# Documentation: docs/configuration.md

third_party_oauth2:
  # JWE signing key for state tokens (REQUIRED)
  # Must be 32 bytes for A256KW algorithm
  # Load from environment variable in production
  jwe_signing_key: ${IDENTITY_BROKER_JWE_SIGNING_KEY}
  
  # State token time-to-live in seconds (default: 600 = 10 minutes)
  # Shorter TTL reduces exposure window but may impact slow authorizations
  state_token_ttl: 600
  
  # PKCE verifier length in bytes (default: 32)
  # Range: 32-128 bytes per RFC 7636
  pkce_verifier_length: 32
```

---

## Testing Patterns

### Unit Test: State Token Service

```go
func TestStateTokenService_CreateAndValidate(t *testing.T) {
    key := make([]byte, 32)
    _, _ = rand.Read(key)
    
    service := NewStateTokenService(key, 10*time.Minute)
    
    claims := StateTokenClaims{
        Principal:    "alice@example.com",
        PKCEVerifier: "test-verifier",
        ServiceID:    "service-123",
        RedirectURI:  "https://example.com/callback",
    }
    
    token, err := service.Create(claims)
    require.NoError(t, err)
    require.NotEmpty(t, token)
    
    validated, err := service.Validate(token)
    require.NoError(t, err)
    assert.Equal(t, claims.Principal, validated.Principal)
    assert.Equal(t, claims.PKCEVerifier, validated.PKCEVerifier)
}

func TestStateTokenService_ValidateExpired(t *testing.T) {
    key := make([]byte, 32)
    _, _ = rand.Read(key)
    
    // TTL of -1 second = already expired
    service := NewStateTokenService(key, -1*time.Second)
    
    token, _ := service.Create(StateTokenClaims{
        Principal:    "alice@example.com",
        PKCEVerifier: "test-verifier",
        ServiceID:    "service-123",
        RedirectURI:  "https://example.com/callback",
    })
    
    _, err := service.Validate(token)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expired")
}
```

### Integration Test: Full OAuth2 Flow

```go
func TestOAuth2Flow_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test containers and adapters
    ctx := context.Background()
    db := setupTestPostgres(t)
    defer db.Close()
    
    // Create mock OAuth2 server
    mockOAuth2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Return mock tokens
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "access_token":  "mock-access-token",
            "refresh_token": "mock-refresh-token",
            "token_type":    "Bearer",
            "expires_in":    3600,
        })
    }))
    defer mockOAuth2.Close()
    
    // Create service with mock OAuth2 endpoints
    service := setupTestService(db, mockOAuth2.URL)
    
    // 1. Initiate flow
    authURL, err := service.InitiateOAuth2Flow(ctx, oauth2session.InitiateFlowParams{
        Principal:   "alice@example.com",
        ServiceID:   testServiceID,
        RedirectURI: "https://example.com/sessions",
        RequestHost: "example.com",
    })
    require.NoError(t, err)
    
    // Extract state from URL
    parsed, _ := url.Parse(authURL)
    state := parsed.Query().Get("state")
    
    // 2. Complete flow (simulating callback)
    session, err := service.CompleteOAuth2Flow(ctx, oauth2session.CompleteFlowParams{
        Principal: "alice@example.com",
        ServiceID: testServiceID,
        Code:      "mock-auth-code",
        State:     state,
    })
    require.NoError(t, err)
    assert.NotEmpty(t, session.ID)
    assert.Equal(t, "alice@example.com", session.Principal)
    
    // 3. Verify session is stored
    stored, err := service.sessionRepo.FindByPrincipalAndService(ctx, "alice@example.com", testServiceID)
    require.NoError(t, err)
    assert.Equal(t, session.ID, stored.ID)
}
```

---

## Frontend Integration

### Session List Hook

Location: `web/src/hooks/useThirdPartySessions.ts`

```typescript
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';

export interface ThirdPartyServiceWithSession {
  service: {
    id: string;
    display_name: string;
    description?: string;
    scopes: Array<{ scope_value: string; description: string }>;
  };
  session: {
    session_id: string;
    initiated_at: string;
    access_token_expires_at?: string;
    refresh_token_expires_at?: string;
    is_expired: boolean;
    has_refresh_token: boolean;
    scope: string[];
    dependent_agent_count: number;
    tokens_encrypted: boolean;
  } | null;
}

export function useThirdPartySessions() {
  return useQuery({
    queryKey: ['third-party-sessions'],
    queryFn: async (): Promise<ThirdPartyServiceWithSession[]> => {
      const response = await api.get('/api/third-party/sessions');
      return response.data.data;
    },
  });
}
```

### Login Button Handler

```typescript
function handleLogin(serviceId: string) {
  const redirectUri = encodeURIComponent(window.location.origin + '/consent/sessions');
  window.location.href = `/api/third-party/${serviceId}/oauth2/authorize?redirect_uri=${redirectUri}`;
}
```

---

## Checklist

Before marking implementation complete:

- [ ] Domain service implemented with all methods
- [ ] State token service with JWE encryption
- [ ] PostgreSQL migration created and tested
- [ ] Repository adapter for PostgreSQL
- [ ] Repository adapter for in-memory (testing)
- [ ] HTTP handlers registered on router
- [ ] Configuration added to schema
- [ ] Configuration example in examples/config/
- [ ] Unit tests for state token service
- [ ] Unit tests for domain service
- [ ] Integration tests for full OAuth2 flow
- [ ] Frontend hooks and components
- [ ] OpenAPI spec merged into /api/enduser/openapi.yaml
- [ ] ARCHITECTURE.md glossary updated
- [ ] Structured audit logging for security events
