# Third-Party OAuth2 Session Management - Domain Invariants

**Feature**: Third-Party OAuth2 Session Management
**Feature Branch**: `008-thirdparty-oauth2-sessions`
**Date**: 2025-12-23
**Purpose**: Document critical domain invariants that MUST be maintained at all times

## What are Domain Invariants?

Domain invariants are business rules that must always hold true, regardless of the system's state. They are enforced through:
- Database constraints (hard guarantees)
- Domain entity validation (defensive programming)
- Service layer checks (business logic)
- API validation (fail-fast on invalid input)

Violating an invariant indicates a bug that must be fixed immediately.

---

## Invariant 1: One Session Per User-Service Pair

**Statement**: Each user (principal) can have at most one active OAuth2 session with any given third-party service.

**Enforcement**:
- **Database**: `UNIQUE CONSTRAINT user_sessions_principal_service_unique` on `(principal, service_id)`
- **Service Layer**: `UserSessionRepository.Create()` uses upsert semantics (ON CONFLICT DO UPDATE)
- **Rationale**: Prevents race conditions during OAuth2 callback; first successful callback wins

**Consequences**:
- If user initiates multiple OAuth2 flows with same service, only the first completed callback creates a session
- Subsequent callbacks replace existing session tokens (token rotation)
- No need for manual session cleanup before re-authentication

**Test Cases**:
- Concurrent callback requests for same (principal, service_id) → only one succeeds
- Re-authenticating with same service → old session replaced atomically
- Database constraint violation → StorageError with Kind=Conflict

---

## Invariant 2: All Tokens Always Encrypted

**Statement**: OAuth2 access and refresh tokens MUST never be stored in plaintext. All tokens are encrypted at rest using AES-GCM.

**Enforcement**:
- **Service Layer**: `OAuth2SessionService.HandleCallback()` encrypts tokens via `EncryptionPort` before storage
- **Database**: Tokens stored as `BYTEA` (ciphertext), never `TEXT`
- **API Layer**: Tokens never appear in API responses
- **Code Review**: No code path that stores plaintext tokens

**Consequences**:
- Encryption failure → session creation fails (fail-closed)
- Decryption failure → session retrieval fails (fail-closed)
- Database compromise does not expose plaintext tokens
- Performance overhead: ~1ms per encryption/decryption operation

**Test Cases**:
- Verify `encrypted_access_token` and `encrypted_refresh_token` columns contain ciphertext (entropy check)
- Verify decryption uses correct encryption context (AAD)
- Verify encryption failure prevents session creation

---

## Invariant 3: PKCE Mandatory

**Statement**: All OAuth2 authorization code flows MUST use PKCE (Proof Key for Code Exchange). No bypass is allowed.

**Enforcement**:
- **Service Layer**: `InitiateOAuth2Flow()` always calls `GeneratePKCE()` before building authorization URL
- **Code Review**: No code path that skips PKCE generation
- **State Token**: PKCE verifier embedded in JWE state token
- **Callback Validation**: Token exchange always includes `code_verifier` parameter

**Consequences**:
- Authorization code interception attacks are mitigated
- Slightly increased complexity (verifier generation + challenge computation)
- No compatibility issues (PKCE is widely supported)

**Test Cases**:
- Verify authorization URL includes `code_challenge` and `code_challenge_method=S256`
- Verify callback includes `code_verifier` in token exchange
- Verify PKCE verifier length is 32-128 bytes per RFC 7636

---

## Invariant 4: State Token TTL ≤ 15 Minutes

**Statement**: OAuth2 state tokens MUST have a time-to-live (TTL) of at most 15 minutes. Default is 10 minutes.

**Enforcement**:
- **Configuration**: `third_party_oauth2.state_token_ttl` validated at startup (max 15 minutes)
- **Service Layer**: `CreateStateToken()` sets `exp` claim to `now + TTL`
- **Callback Validation**: `ValidateStateToken()` rejects expired tokens with `StateTokenExpired` error

**Consequences**:
- Limits exposure window for state token leakage
- User must complete OAuth2 flow within 15 minutes
- Expired state token → callback fails with 400 Bad Request

**Test Cases**:
- Verify configuration validation rejects TTL > 15 minutes
- Verify state token expiration is honored (clock skew: 30 seconds)
- Verify expired state token rejected in callback

---

## Invariant 5: Encryption Context Binds Tokens to Session

**Statement**: Token encryption MUST include encryption context (AAD) binding ciphertext to `principal`, `service_id`, `session_id`, and `purpose="oauth2_token"`.

**Enforcement**:
- **Service Layer**: `createSession()` helper builds `EncryptionContext` struct before encryption
- **EncryptionPort**: `Encrypt()` method includes AAD in AES-GCM operation
- **Database**: Encryption context stored as JSONB for auditing
- **Decryption**: `Decrypt()` requires matching encryption context

**Consequences**:
- Tokens encrypted for one session cannot be decrypted for another (context mismatch)
- Prevents "ciphertext substitution" attacks
- Enables auditing (who encrypted which tokens, when)
- Slight storage overhead (~100 bytes JSON per session)

**Test Cases**:
- Verify decryption with mismatched context fails
- Verify encryption context contains all required fields
- Verify encryption context is immutable after creation

---

## Invariant Violation Handling

If an invariant is violated:

1. **Log Security Event**: Emit structured log with severity `ERROR` or `CRITICAL`
2. **Fail-Closed**: Reject the operation (do not proceed with inconsistent state)
3. **Return Error**: Return appropriate HTTP status code (400/403/409/500)
4. **Alert**: Trigger monitoring alert for critical invariants (#1, #2, #5)
5. **Audit**: Record violation in audit log with full context

**Example**:
```go
// Invariant violation detected
logger.Error("invariant violation: token encryption context mismatch",
    slog.String("session_id", sessionID),
    slog.String("principal", principal),
    slog.String("service_id", serviceID),
    slog.String("violation", "encryption_context_mismatch"))
return nil, ErrEncryptionContextMismatch
```

---

## Testing Invariants

Each invariant MUST have:
- Unit tests verifying enforcement
- Integration tests verifying database constraints
- Negative tests verifying violations are detected and rejected

See:
- `tests/unit/oauth2session/user_session_test.go`
- `tests/integration/user_sessions_test.go`
- `tests/unit/oauth2session/state_token_security_test.go`

---

## References

- RFC 7636: Proof Key for Code Exchange (PKCE)
- NIST SP 800-38D: AES-GCM
- Data Model: `/specs/008-thirdparty-oauth2-sessions/data-model.md`
- Architecture: `/ARCHITECTURE.md` (Glossary section)
