# Data Model: Third-Party OAuth2 Session Management

**Feature**: 008-thirdparty-oauth2-sessions  
**Date**: 2025-12-22  
**Purpose**: Document domain entities, relationships, and validation rules

## Table of Contents

1. [Entity Relationship Diagram](#entity-relationship-diagram)
2. [Entities](#entities)
3. [Value Objects](#value-objects)
4. [Domain Events](#domain-events)
5. [Database Schema](#database-schema)
6. [Glossary Updates](#glossary-updates)

---

## Entity Relationship Diagram

```
┌─────────────────────┐     1:N     ┌─────────────────────┐
│    Principal        │◄────────────│    UserSession      │
│    (User ID)        │             │    (OAuth2 tokens)  │
└─────────────────────┘             └──────────┬──────────┘
                                               │
                                               │ N:1
                                               ▼
                                    ┌─────────────────────┐
                                    │ ThirdpartyOAuth2    │
                                    │ Service             │
                                    │ (already exists)    │
                                    └─────────────────────┘

┌─────────────────────┐             ┌─────────────────────┐
│  OAuth2StateToken   │─────────────│    UserSession      │
│  (JWE, ephemeral)   │  creates    │    (after callback) │
└─────────────────────┘             └─────────────────────┘

Relationships:
- Principal 1:N UserSession (one user can have sessions with multiple services)
- ThirdpartyOAuth2Service 1:N UserSession (one service can have sessions with multiple users)
- UserSession has UNIQUE(principal, service_id) constraint
- UserGrant references UserSession via delegated_oauth2_tokens[].thirdparty_oauth2_service_id

**Agent Count Derivation**:
The dependent_agent_count displayed in the UI is calculated by querying the user_grants table.
For a given UserSession with (principal, service_id), the count is derived from:
```sql
SELECT COUNT(*) FROM user_grants 
WHERE principal = ? 
AND delegated_oauth2_tokens @> '[{"thirdparty_oauth2_service_id": ?}]'::jsonb
```
This counts all agents that have delegated_oauth2_tokens referencing the specific thirdparty_oauth2_service_id.
```

---

## Entities

### UserSession

Represents an authenticated OAuth2 session between a user and a third-party service. Contains encrypted access and refresh tokens.

**Location**: `internal/domain/storage/user_session.go`

```go
package storage

import (
    "errors"
    "fmt"
    "time"
)

// UserSession represents an authenticated OAuth2 session between a user (principal)
// and a third-party OAuth2 service. Tokens are stored encrypted.
type UserSession struct {
    // ID is the unique identifier for this session (UUID)
    ID string `json:"id" db:"id"`
    
    // Principal is the authenticated user identifier (from session/header)
    Principal string `json:"principal" db:"principal"`
    
    // ServiceID references the ThirdpartyOAuth2Service
    ServiceID string `json:"service_id" db:"service_id"`
    
    // EncryptedAccessToken is the AES-256-GCM encrypted OAuth2 access token
    EncryptedAccessToken []byte `json:"-" db:"encrypted_access_token"`
    
    // EncryptedRefreshToken is the AES-256-GCM encrypted OAuth2 refresh token (nullable)
    EncryptedRefreshToken []byte `json:"-" db:"encrypted_refresh_token"`
    
    // TokenType is the OAuth2 token type (typically "Bearer")
    TokenType string `json:"token_type" db:"token_type"`
    
    // AccessTokenExpiresAt is when the access token expires (nullable if unknown)
    AccessTokenExpiresAt *time.Time `json:"access_token_expires_at,omitempty" db:"access_token_expires_at"`
    
    // RefreshTokenExpiresAt is when the refresh token expires (nullable if unknown)
    // Session is marked "expired" in UI only when this is set and in the past
    RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at,omitempty" db:"refresh_token_expires_at"`
    
    // Scope is the list of OAuth2 scopes granted for this session
    Scope []string `json:"scope" db:"scope"`
    
    // InitiatedAt is when the OAuth2 flow was completed and tokens were stored
    InitiatedAt time.Time `json:"initiated_at" db:"initiated_at"`
    
    // EncryptionContext stores metadata used for token encryption AAD
    EncryptionContext map[string]string `json:"-" db:"encryption_context"`
    
    // CreatedAt is the database row creation timestamp
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    
    // UpdatedAt is the database row update timestamp
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate performs validation on the UserSession entity.
func (s *UserSession) Validate() error {
    if s.ID == "" {
        return errors.New("session ID cannot be empty")
    }
    if s.Principal == "" {
        return errors.New("principal is required")
    }
    if len(s.Principal) > 200 {
        return fmt.Errorf("principal exceeds 200 characters (got %d)", len(s.Principal))
    }
    if s.ServiceID == "" {
        return errors.New("service_id is required")
    }
    if len(s.EncryptedAccessToken) == 0 {
        return errors.New("encrypted_access_token is required")
    }
    if s.TokenType == "" {
        return errors.New("token_type is required")
    }
    if s.InitiatedAt.IsZero() {
        return errors.New("initiated_at is required")
    }
    return nil
}

// ValidateForCreate validates a session before creation.
func (s *UserSession) ValidateForCreate() error {
    if s.Principal == "" {
        return errors.New("principal is required")
    }
    if s.ServiceID == "" {
        return errors.New("service_id is required")
    }
    if len(s.EncryptedAccessToken) == 0 {
        return errors.New("encrypted_access_token is required")
    }
    if s.TokenType == "" {
        return errors.New("token_type is required")
    }
    return nil
}

// IsExpired returns true if the refresh token has expired.
// A session is only considered expired when the refresh token expires,
// not when the access token expires (access tokens can be refreshed).
func (s *UserSession) IsExpired() bool {
    if s.RefreshTokenExpiresAt == nil {
        return false // No expiration = never expires
    }
    return s.RefreshTokenExpiresAt.Before(time.Now())
}

// HasRefreshToken returns true if a refresh token is stored.
func (s *UserSession) HasRefreshToken() bool {
    return len(s.EncryptedRefreshToken) > 0
}
```

### UserSessionSummary

View model for session list endpoint (includes service display name and agent count).

```go
// UserSessionSummary is a read-only view of a session with related data.
// Used for listing sessions with service info and dependent agent count.
type UserSessionSummary struct {
    SessionID            string     `json:"session_id"`
    ServiceID            string     `json:"service_id"`
    ServiceDisplayName   string     `json:"service_display_name"`
    TokenType            string     `json:"token_type"`
    Scope                []string   `json:"scope"`
    InitiatedAt          time.Time  `json:"initiated_at"`
    AccessTokenExpiresAt *time.Time `json:"access_token_expires_at,omitempty"`
    RefreshTokenExpiresAt *time.Time `json:"refresh_token_expires_at,omitempty"`
    IsExpired            bool       `json:"is_expired"`
    HasRefreshToken      bool       `json:"has_refresh_token"`
    DependentAgentCount  int        `json:"dependent_agent_count"`
}
```

---

## Value Objects

### OAuth2StateTokenClaims

Claims embedded in JWE state token during OAuth2 flow. Not persisted.

**Location**: `internal/domain/oauth2session/state_token.go`

```go
// OAuth2StateTokenClaims contains the claims embedded in a JWE state token.
// These claims bind the OAuth2 callback to the initiating request and user.
type OAuth2StateTokenClaims struct {
    // Principal is the user who initiated the OAuth2 flow
    Principal string `json:"sub"`
    
    // PKCEVerifier is the code verifier for PKCE validation
    PKCEVerifier string `json:"pkce_verifier"`
    
    // ServiceID is the third-party service being authenticated
    ServiceID string `json:"service_id"`
    
    // RedirectURI is the original redirect_uri from authorize request
    RedirectURI string `json:"redirect_uri"`
    
    // IssuedAt is when the token was created (Unix timestamp)
    IssuedAt int64 `json:"iat"`
    
    // ExpiresAt is when the token expires (Unix timestamp)
    ExpiresAt int64 `json:"exp"`
}

// Validate validates the claims after decryption.
func (c *OAuth2StateTokenClaims) Validate() error {
    if c.Principal == "" {
        return errors.New("principal claim is required")
    }
    if c.PKCEVerifier == "" {
        return errors.New("pkce_verifier claim is required")
    }
    if c.ServiceID == "" {
        return errors.New("service_id claim is required")
    }
    if c.RedirectURI == "" {
        return errors.New("redirect_uri claim is required")
    }
    if c.ExpiresAt == 0 {
        return errors.New("exp claim is required")
    }
    return nil
}

// IsExpired returns true if the token has expired.
func (c *OAuth2StateTokenClaims) IsExpired() bool {
    return time.Now().Unix() > c.ExpiresAt
}
```

### AffectedAgent

Information about an agent that will lose access when a session is terminated.

```go
// AffectedAgent represents an agent that depends on a user's session.
// Used in termination warning dialog.
type AffectedAgent struct {
    AgentID     string `json:"agent_id"`
    DisplayName string `json:"display_name"`
}
```

---

## Domain Events

Events emitted by the OAuth2SessionService for audit logging and future event-driven features.

```go
// SessionEstablishedEvent is emitted when a user successfully completes
// an OAuth2 flow and tokens are stored.
type SessionEstablishedEvent struct {
    SessionID   string    `json:"session_id"`
    Principal   string    `json:"principal"`
    ServiceID   string    `json:"service_id"`
    Scope       []string  `json:"scope"`
    InitiatedAt time.Time `json:"initiated_at"`
    Timestamp   time.Time `json:"timestamp"`
}

// SessionTerminatedEvent is emitted when a user explicitly terminates
// a session and tokens are deleted.
type SessionTerminatedEvent struct {
    SessionID           string    `json:"session_id"`
    Principal           string    `json:"principal"`
    ServiceID           string    `json:"service_id"`
    AffectedAgentCount  int       `json:"affected_agent_count"`
    Timestamp           time.Time `json:"timestamp"`
}

// StateTokenValidationFailedEvent is emitted when state token validation
// fails at callback. Security audit event.
type StateTokenValidationFailedEvent struct {
    Reason      string    `json:"reason"` // "expired", "principal_mismatch", "service_id_mismatch", "decryption_failed"
    ServiceID   string    `json:"service_id"`
    Principal   string    `json:"principal,omitempty"` // Current principal if available
    RequestedBy string    `json:"requested_by,omitempty"` // Principal in token if available
    Timestamp   time.Time `json:"timestamp"`
}
```

---

## Database Schema

### Migration: 004_create_user_sessions.up.sql

```sql
-- Migration: Create user_sessions table
-- Feature: 008-thirdparty-oauth2-sessions
-- Description: Stores encrypted OAuth2 tokens for user sessions with third-party services

CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY,
    principal VARCHAR(200) NOT NULL,
    service_id UUID NOT NULL REFERENCES thirdparty_oauth2_services(id) ON DELETE RESTRICT,
    encrypted_access_token BYTEA NOT NULL,
    encrypted_refresh_token BYTEA,  -- Nullable, some services don't provide refresh tokens
    token_type VARCHAR(50) NOT NULL DEFAULT 'Bearer',
    access_token_expires_at TIMESTAMP,  -- Nullable if expiration unknown
    refresh_token_expires_at TIMESTAMP,  -- Nullable if refresh token not provided or never expires
    scope TEXT[] NOT NULL DEFAULT '{}',  -- Array of OAuth2 scopes
    initiated_at TIMESTAMP NOT NULL,
    encryption_context JSONB NOT NULL DEFAULT '{}',  -- AAD metadata for encryption
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- One session per user per service (first callback wins on race condition)
    CONSTRAINT uq_principal_service UNIQUE(principal, service_id)
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_sessions_principal ON user_sessions(principal);
CREATE INDEX IF NOT EXISTS idx_sessions_service ON user_sessions(service_id);
CREATE INDEX IF NOT EXISTS idx_sessions_principal_service ON user_sessions(principal, service_id);
CREATE INDEX IF NOT EXISTS idx_sessions_initiated ON user_sessions(initiated_at);

-- Comments
COMMENT ON TABLE user_sessions IS 'Stores encrypted OAuth2 tokens for user sessions with third-party services';
COMMENT ON COLUMN user_sessions.id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN user_sessions.principal IS 'User identifier (from authentication)';
COMMENT ON COLUMN user_sessions.service_id IS 'Third-party service (foreign key with RESTRICT delete)';
COMMENT ON COLUMN user_sessions.encrypted_access_token IS 'AES-256-GCM encrypted access token';
COMMENT ON COLUMN user_sessions.encrypted_refresh_token IS 'AES-256-GCM encrypted refresh token (nullable)';
COMMENT ON COLUMN user_sessions.token_type IS 'OAuth2 token type (typically Bearer)';
COMMENT ON COLUMN user_sessions.access_token_expires_at IS 'When access token expires (nullable)';
COMMENT ON COLUMN user_sessions.refresh_token_expires_at IS 'When refresh token expires - session marked expired only when this passes';
COMMENT ON COLUMN user_sessions.scope IS 'Array of OAuth2 scopes granted';
COMMENT ON COLUMN user_sessions.initiated_at IS 'When OAuth2 flow completed and tokens stored';
COMMENT ON COLUMN user_sessions.encryption_context IS 'Encryption AAD metadata (principal, service_id, session_id)';
COMMENT ON CONSTRAINT uq_principal_service ON user_sessions IS 'Ensures one session per user per service';
```

### Migration: 004_create_user_sessions.down.sql

```sql
-- Migration: Drop user_sessions table
-- Feature: 008-thirdparty-oauth2-sessions

DROP TABLE IF EXISTS user_sessions;
```

---

## Glossary Updates

Add to ARCHITECTURE.md Glossary section:

### Third-Party OAuth2 Session Domain

**UserSession**: Represents an authenticated OAuth2 session between a user (principal) and a third-party service. Contains encrypted access and refresh tokens, scope information, and expiration metadata. A user can have at most one session per service (UNIQUE constraint on principal + service_id).

**OAuth2StateToken**: A short-lived JWE (JSON Web Encryption) token used to bind an OAuth2 authorization flow to the initiating user and request. Contains principal, PKCE verifier, service ID, and redirect URI. Valid for 10 minutes by default. Used for CSRF protection and flow binding.

**OAuth2SessionService**: Domain service that orchestrates OAuth2 authorization code flows with third-party services. Coordinates PKCE generation, state token creation, callback validation, token exchange, and encrypted token storage. Located in `internal/domain/oauth2session/`.

**PKCE (Proof Key for Code Exchange)**: RFC 7636 security mechanism that protects OAuth2 authorization code flows against interception attacks. Uses a code verifier (high-entropy random string) and code challenge (SHA256 hash of verifier). The verifier is stored in the JWE state token.

**JWE State Binding**: Security pattern using JWE to bind OAuth2 callback to the original request. The state token is encrypted, preventing tampering, and contains claims that must match the callback context (principal, service ID). Provides CSRF protection.

**Session Expiration**: A user session is considered "expired" only when the refresh token expires (not the access token). Expired access tokens can be transparently refreshed in future iterations. The UI shows "Expired" status when refresh_token_expires_at < NOW().

**Dependent Agent Count**: The number of agents that have active grants referencing a user's session with a third-party service. Displayed in the termination warning dialog. Calculated by querying user_grants where delegated_oauth2_tokens contains the service_id.

---

## Repository Interface

Add to `internal/ports/storage.go`:

```go
// UserSessionRepository defines storage operations for user OAuth2 sessions.
// One session per user per service (UNIQUE constraint on principal + service_id).
type UserSessionRepository interface {
    // Create creates a new user session.
    // On conflict (existing session for same principal + service), returns existing session.
    // Returns error if:
    // - Service ID doesn't exist (StorageError with Kind=NotFound)
    // - Storage connection fails (StorageError with Kind=Connection)
    Create(ctx context.Context, session *storage.UserSession) (*storage.UserSession, error)
    
    // Get retrieves a user session by ID.
    // Returns StorageError with Kind=NotFound if session not found.
    Get(ctx context.Context, id string) (*storage.UserSession, error)
    
    // FindByPrincipalAndService retrieves session for a principal and service.
    // Returns nil, nil if no session exists (not an error).
    FindByPrincipalAndService(ctx context.Context, principal, serviceID string) (*storage.UserSession, error)
    
    // Delete deletes a user session by ID.
    // Idempotent: safe to delete non-existent sessions.
    Delete(ctx context.Context, id string) error
    
    // DeleteByPrincipalAndService deletes session for a principal and service.
    // Idempotent: safe to call if no session exists.
    DeleteByPrincipalAndService(ctx context.Context, principal, serviceID string) error
    
    // ListByPrincipal retrieves all sessions for a principal.
    // Returns empty slice if no sessions exist (not an error).
    ListByPrincipal(ctx context.Context, principal string) ([]*storage.UserSession, error)
    
    // CountByService returns number of sessions for a service.
    // Used to block service deletion when sessions exist.
    CountByService(ctx context.Context, serviceID string) (int, error)
}
```
