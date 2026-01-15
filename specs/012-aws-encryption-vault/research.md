# Phase 0 Research: AWS Encryption Vault for OAuth Tokens

**Date**: 2026-01-15
**Branch**: `012-aws-encryption-vault`
**Spec**: [spec.md](./spec.md)

## Executive Summary

Phase 0 research confirms that the codebase architecture is **well-suited for AWS Encryption Vault implementation**. Existing patterns (port/adapter architecture, builder dependency injection, unified configuration system, E2E test infrastructure) provide excellent foundation. Key finding: **no migration of schema required** - session table already has encrypted token fields and encryption context support.

**Critical Path Items**:
1. Add AWS SDK dependencies (AWS Encryption SDK, Material Providers Library, KMS client)
2. Implement memory protection utilities using memguard
3. Update storage adapters to use encryption port (transparent encryption/decryption)
4. Replace noop encryption adapter with AWS SDK wrapper
5. Extend unified configuration for AWS KMS and environment-variable KEK storage
6. Write E2E acceptance tests mapped 1:1 to spec scenarios

## Detailed Research Findings

### 1. AWS SDK Integration Status

**Current State**:
- ✅ Go project uses standard library cryptography (`golang.org/x/crypto`)
- ❌ No AWS SDK dependencies currently present in `go.mod`
- ❌ No existing AWS Encryption SDK integration

**Required Packages**:
```
github.com/aws/aws-sdk-go-v2/service/kms          // KMS client for key management
github.com/aws/aws-cryptographic-material-providers-library/releases/go/mpl  // Keyring management
github.com/aws/aws-encryption-sdk-go/v3            // AESGCMSIV authenticated encryption
```

**Implementation Impact**: Low - packages are isolated behind EncryptionPort interface, no leakage to domain logic.

### 2. EncryptionPort Interface Foundation

**Current Definition** (`internal/ports/encryption.go`):
```go
type EncryptionPort interface {
    Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)
}
```

**Design Quality**:
- ✅ Context parameter supports cancellation and timeouts
- ✅ Map-based encryption context aligns with AWS AAD (Additional Authenticated Data) patterns
- ✅ Proper error handling convention
- ✅ Stateless interface design (no connection management needed)

**Integration Points**:
- Builder already has `WithEncryption()` injection point
- Storage adapters can use port without modification
- No interface changes needed

### 3. Storage Layer Already Supports Encryption

**UserSession Domain Model** (`internal/domain/storage/user_session.go`):
```go
type UserSession struct {
    EncryptedAccessToken  []byte            `json:"-" db:"encrypted_access_token"`
    EncryptedRefreshToken []byte            `json:"-" db:"encrypted_refresh_token"`
    EncryptionContext     EncryptionContext `json:"encryption_context" db:"encryption_context"`
    // ... other fields
}

type EncryptionContext struct {
    Principal string `json:"principal"`
    ServiceID string `json:"service_id"`
    SessionID string `json:"session_id"`
    Purpose   string `json:"purpose"`
}
```

**Critical Discovery**:
- ✅ Schema ALREADY has encrypted token columns
- ✅ EncryptionContext ALREADY defined with required 4 fields
- ✅ JSONB storage for context already established
- ✅ No database migrations required

**Implication**: Feature is nearly ready from schema perspective; focus is on encryption logic and memory protection.

### 4. Storage Adapter Patterns

**Both memory and PostgreSQL adapters**:
- ✅ Follow interface segregation with small, focused repository interfaces
- ✅ Proper error wrapping to domain `StorageError` type
- ✅ Support transparent encryption/decryption through port interface injection
- ✅ Existing `noop.NewNoOpEncryption()` provides development baseline

**PostgreSQL Adapter Insights**:
- Uses `sqlx` (not ORM) per ADR 004
- Handles JSONB serialization for EncryptionContext correctly
- Proper upsert semantics with `ON CONFLICT`
- Thread-safe connection pooling

**Memory Adapter Insights**:
- Thread-safe map operations with `sync.RWMutex`
- Dual indexing for efficient lookups
- Upsert semantics matching PostgreSQL behavior

### 5. Builder Pattern and Dependency Injection

**Current Builder** (`internal/app/builder.go`):
```go
type Builder struct {
    config     *ports.Config
    storage    *storage.Adapter
    logger     *slog.Logger
    encryption ports.EncryptionPort  // Already here!
}

func (b *Builder) WithEncryption(encryptor ports.EncryptionPort) *Builder {
    b.encryption = encryptor
    return b
}
```

**Implementation Ready**:
- ✅ Encryption port injection already supported
- ✅ Chainable builder pattern established
- ✅ Proper validation of required dependencies
- ✅ Clear initialization phases

**Integration Path**: Simply instantiate AWS SDK adapter and wire via `WithEncryption()` in main.go

### 6. Unified Configuration System

**Architecture** (`internal/config/`):
- ✅ Viper-based with precedence: defaults → .env → YAML → CLI flags
- ✅ Environment variable prefixing established (`IDENTITY_BROKER_*`)
- ✅ Circular reference detection and injection prevention
- ✅ Source metadata tracking for debugging

**Configuration Extension Required**:
```yaml
encryption:
  backend: "aws-kms"  # or "env-var" for development
  aws:
    region: "us-east-1"
    kms_key_id: "arn:aws:kms:..."
  env_var:
    kek_var_name: "ENCRYPTION_KEK"
```

**Integration Path**: Extend existing schema with AWS-specific fields, no architectural changes needed.

### 7. E2E Testing Infrastructure

**Ginkgo/Gomega Setup** (`tests/e2e/`):
- ✅ Full production app bootstrap via `TestServer` wrapper
- ✅ Real HTTP server with production routing
- ✅ Authentication simulation via `X-Remote-User` header
- ✅ Fixture-based test data management
- ✅ Per-test isolation (fresh server/storage)

**Testcontainers Integration**:
- ✅ Real PostgreSQL for integration tests
- ✅ Database cleanup between tests
- ✅ Migration verification

**E2E Test Capability**: Ready to write 24+ acceptance tests mapped 1:1 to spec scenarios.

### 8. Memory Protection Status

**Current State**:
- ❌ No memory protection utilities exist
- ❌ Tokens stored as plain `[]byte` without explicit zeroization
- ❌ No memory locking (mlock) for key material
- ⚠️ Core dump exclusion not implemented

**Implementation Requirement**: Add memory_protection package with:
- Buffer zeroization utilities (defer-based cleanup)
- Memory locking (mprotect via `golang.org/x/sys/unix`)
- Core dump exclusion (RLIMIT_CORE setting)
- Secure comparison helpers

**Recommended Library**: `github.com/awnumar/memguard` for production-grade memory protection.

### 9. Domain Error Handling

**StorageError Pattern** (`internal/domain/storage/errors.go`):
- ✅ Well-structured error types with error kinds
- ✅ Proper error wrapping for cause chains
- ✅ Domain-friendly messages without implementation leaks

**Encryption Error Extensions**:
Need to define new error kinds:
- `ErrorKindEncryptionFailed` - DEK generation, context binding, KEK wrapping failed
- `ErrorKindDecryptionFailed` - DEK unwrapping, token decryption failed
- `ErrorKindContextMismatch` - Context verification failed

**Integration Path**: Extend existing error handling pattern with encryption-specific kinds.

### 10. Constitution Compliance Matrix

| Principle | Status | Integration Path |
|-----------|--------|------------------|
| I. Security-First | ✅ Ready | Fail-closed on any encryption error; no plaintext fallback |
| II. Architecture Docs | ⏳ Phase 1 | Update ARCHITECTURE.md with encryption domain |
| III. Library-First Security | ✅ Ready | Use AWS Encryption SDK (battle-tested); no custom crypto |
| IV. API Documentation | ⏳ Phase 1 | Document encryption configuration in OpenAPI |
| V. Domain-Driven Design | ✅ Ready | EncryptionContext already modeled; add to glossary |
| VI. Hexagonal Architecture | ✅ Ready | Encryption port interfaces with adapter pattern |
| VII. Configuration-Driven | ✅ Ready | Extend unified config system for AWS/env-var backends |
| VIII. Test-Driven Development | ⏳ Phase 2f | Write E2E tests FIRST; verify red phase; implement |
| IX. Persistence Patterns | ✅ Ready | No schema changes; use existing patterns |
| XII. Dependency Injection | ✅ Ready | Builder pattern; wire via `WithEncryption()` |
| XIII. E2E Acceptance Tests | ⏳ Phase 2f | 24+ tests mapped to spec scenarios |

## Critical Implementation Decisions

### 1. One DEK Per Session (Performance Optimization)

**Decision**: Use one DEK per session (covering both access and refresh tokens) instead of DEK per token.

**Rationale**:
- Both tokens share same context (principal, service_id, session_id, purpose)
- Eliminates redundant DEK generation and KEK wrapping operations
- Reduces AWS KMS calls from 2 to 1 per session operation
- Keeps encryption within 100ms performance budget

**Security Impact**: No degradation - cross-session DEKs still isolate sessions; context verification prevents cross-context reuse.

### 2. AWS Encryption SDK with AESGCMSIV

**Decision**: Use AWS Encryption SDK's AESGCMSIV (Encrypt-then-MAC with counter mode) for authenticated encryption.

**Rationale**:
- Battle-tested implementation (used by AWS services)
- Chosen-ciphertext attack protection via authentication tag
- Post-quantum cryptography readiness (Go 1.24+ support)
- AEAD (Authenticated Encryption with Associated Data) for context binding

**No Alternatives**: Custom AEAD implementations forbidden by Constitution Principle III.

### 3. Two KEK Storage Backends

**Decision**: Support AWS KMS (production) and environment-variable (development/containers) KEK storage.

**Trade-off**:
- Environment variables create key material exposure risk (process memory)
- Mitigated via memory protection (mlock, buffer zeroing, core dump exclusion)
- Worth complexity for developer experience and container deployments

### 4. No Schema Changes Needed

**Decision**: Leverage existing session table schema without migrations.

**Evidence**:
- `encrypted_access_token` column exists
- `encrypted_refresh_token` column exists
- `encryption_context` JSONB column exists with proper serialization

**Implication**: Focus implementation effort on encryption logic, not database changes.

## Implementation Roadmap

### Phase 1 - Design Artifacts (This Phase)
1. ✅ Complete research document (you are here)
2. Generate `data-model.md` with domain model specifics
3. Generate API contracts in `contracts/` directory
4. Generate `quickstart.md` for feature integration
5. Update `ARCHITECTURE.md` with encryption domain

### Phase 2 - Implementation
1. Phase 2a: Add AWS SDK dependencies to `go.mod`
2. Phase 2b: Implement memory protection utilities
3. Phase 2c: Implement AWS SDK encryption adapter
4. Phase 2d: Extend configuration schema for AWS KMS and env-var
5. Phase 2e: Update storage adapters to use encryption port
6. Phase 2f: Write E2E acceptance tests (red phase)
7. Phase 3+: Implementation per E2E green phase

### Phase 3 - Constitutional Verification
- Verify TDD (tests written first, red phase, green phase)
- Verify no custom cryptography
- Verify fail-closed behavior
- Verify memory protection effectiveness
- Verify E2E test coverage (100% of spec scenarios)

## Open Questions / Clarifications Resolved

**Q1**: Does schema need updating for encrypted tokens?
**A**: ✅ No - encrypted token columns and encryption context already exist.

**Q2**: Is the encryption port interface adequate for AWS integration?
**A**: ✅ Yes - supports context parameters and error handling needed.

**Q3**: What's the builder integration point?
**A**: ✅ `Builder.WithEncryption()` already exists and ready to use.

**Q4**: How does unified config system work?
**A**: ✅ Viper-based with precedence; can extend with AWS-specific fields.

**Q5**: Is DEK-per-token viable performance-wise?
**A**: ❌ No (would require 2 KMS calls per session). Changed to DEK-per-session in spec.

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|-----------|
| AWS SDK dependency bloat | Medium | Isolated behind EncryptionPort; no domain leakage |
| Memory protection complexity | Medium | Use memguard library; test memory behavior explicitly |
| KMS latency in performance budget | Medium | DEK-per-session reduces calls; 100ms budget excludes KMS |
| Environment-variable KEK exposure | Medium-High | Mitigated by mlock, buffer zeroing, core dump exclusion; documented tradeoff |
| Configuration validation | Low | Extend existing validated config system |

## Success Criteria for Phase 1 → Phase 2 Transition

- [x] Research document complete (no NEEDS CLARIFICATION items remain)
- [ ] data-model.md generated with complete domain model
- [ ] API contracts documented for encryption configuration
- [ ] quickstart.md provides integration guide
- [ ] ARCHITECTURE.md updated with encryption domain terminology
- [ ] Constitution Check re-evaluated and all gates pass
- [ ] Phase 2f E2E test design ready (tests written before implementation)

---

**Status**: ✅ Phase 0 Research Complete - Ready for Phase 1 Design Artifacts

**Next**: Generate Phase 1 design documents (data-model.md, contracts/, quickstart.md)
