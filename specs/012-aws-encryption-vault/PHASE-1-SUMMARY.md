# Phase 1 Design Summary: Encryption Vault for OAuth Tokens

**Date**: 2026-01-16 | **Feature**: 012-aws-encryption-vault | **Status**: Phase 1 Complete

---

## Executive Summary

Phase 1 design for the AWS Encryption Vault feature is complete. All design artifacts have been created, validated against the Constitution, and are ready for Phase 2 implementation. The design establishes:

- **Service-layer encryption** (not storage-adapter encryption)
- **Hierarchical Keyring with Branch Key caching** (15-minute TTL recommended)
- **Memguard protection** for cached keys and token transit
- **Fail-closed security** with no plaintext fallback
- **Context binding** (service_id only) for cross-service prevention

---

## Phase 1 Deliverables

### Core Specification Documents

| Document | Purpose | Status |
|----------|---------|--------|
| [spec.md](./spec.md) | Feature specification (7 user stories, 30 requirements, 24 acceptance scenarios) | ✅ Complete (v1.3) |
| [plan.md](./plan.md) | Implementation plan with Constitution Check and testing strategy | ✅ Complete (v1.0) |
| [research.md](./research.md) | Phase 0 technical research findings | ✅ Complete |
| [data-model.md](./data-model.md) | Domain model with entities, aggregates, value objects | ✅ Complete (v1.0) |
| [quickstart.md](./quickstart.md) | Developer integration guide with code examples | ✅ Complete (v1.0) |

### API Contracts (contracts/ directory)

| Contract | Purpose | Status |
|----------|---------|--------|
| [encryption-port.md](./contracts/encryption-port.md) | EncryptionPort interface specification with Encrypt/Decrypt signatures | ✅ Complete (v1.0) |
| [configuration.md](./contracts/configuration.md) | Configuration contract for KEK storage (AWS KMS vs. environment variable) | ✅ Complete (v1.0) |
| [error-contract.md](./contracts/error-contract.md) | Error types, error handling patterns, HTTP mapping | ✅ Complete (v1.0) |
| [domain-events.md](./contracts/domain-events.md) | Domain events (SessionEncrypted, SessionDecrypted, failures) | ✅ Complete (v1.0) |
| [aws-kms-architecture.md](./contracts/aws-kms-architecture.md) | AWS KMS architectural decision (Hierarchical Keyring vs. Direct KMS) | ✅ Complete (v1.0) |
| [memguard-integration.md](./contracts/memguard-integration.md) | Implementation guide for memguard + AWS Encryption SDK integration | ✅ Complete (v1.0) |

---

## Architecture Decisions

### 1. Service-Layer Encryption (Not Storage-Adapter)

**Decision**: OAuth2SessionService owns encryption/decryption logic; storage adapters handle encrypted tokens transparently.

**Rationale**:
- Simpler storage adapter implementation (no encryption awareness)
- Clear separation of concerns
- Easier to test service layer independently
- Service layer can orchestrate both access and refresh token encryption atomically

**Implementation**:
```
User Request
    ↓
OAuth2SessionService.CreateSession()
    ↓
[Encrypt access token via port.Encrypt()]
[Encrypt refresh token via port.Encrypt()]
    ↓
Create UserSession with encrypted fields
    ↓
repository.Create(userSession)  ← Repository sees encrypted tokens only
    ↓
return session to caller
```

### 2. Hierarchical Keyring with Branch Key Caching

**Decision**: Use AWS Encryption SDK's Hierarchical Keyring with memguard-protected Branch Key cache (15-minute TTL recommended).

**Rationale**:
- Performance: ~1-5ms cached operations vs. 50-200ms per direct KMS call
- Cost: 10x cheaper (~$0.03 vs. $0.30 per 10K operations)
- Availability: Tolerates temporary KMS outages (within TTL window)
- Security: 15-minute TTL + memguard protection balances usability vs. security
- Scalability: Supports 100+ service contexts with <1MB memory overhead

**Three-Layer Architecture**:
```
KEK (AWS KMS) → Branch Key (cached, memguard-protected) → DEK (per-session)
```

### 3. Context Binding: service_id Only

**Decision**: Reduce EncryptionContext from 4 fields (principal, service_id, session_id, purpose) to single field: service_id.

**Rationale**:
- Optimizes KMS operations (one Branch Key per service_id)
- Provides service-level isolation (prevents cross-service token reuse)
- Sufficient for defense-in-depth (context binding at both DEK and KEK layers)
- Simpler configuration and management

**Result**: Tokens encrypted for service_id="oauth2" cannot be decrypted with service_id="github".

### 4. Memguard Protection Strategy

**Decision**: Protect cached Branch Keys, plaintext tokens in transit, and DEKs using memguard.

**Protection Guarantees**:
- Branch Keys in cache: MLOCK (prevent swap), DONTDUMP (exclude from core dumps)
- Plaintext tokens: Wrapped in memguard enclaves during encrypt/decrypt
- Memory zeroing: Explicit cryptographic-grade zeroing on cleanup

**AWS SDK Integration**:
- Custom `CryptographicMaterialsCache` implementation
- Put: Move Branch Keys into memguard enclaves with TTL
- Get: Return LockedBuffer for AWS SDK operations
- Automatic cleanup: Enclave destruction on TTL expiration

### 5. Fail-Closed Security

**Decision**: No plaintext fallback; all encryption/decryption failures result in application-level failures.

**Implementation**:
- Encryption failures prevent session storage
- Decryption failures prevent session retrieval
- Context verification failures immediately fail decryption
- Integrity violations immediately fail decryption (no retry)
- KEK unavailability causes application failure (no degradation)

---

## Configuration Design

### Single Unified Configuration Field

```yaml
encryption:
  key_encryption_key: <AWS_KMS_ARN_or_ENV_VAR_REFERENCE>
```

**Production** (AWS KMS):
```yaml
encryption:
  key_encryption_key: arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012
```

**Development** (Environment Variable):
```yaml
encryption:
  key_encryption_key: ${ENCRYPTION_KEK}  # Resolves via config system interpolation
```

**Automatic Detection**:
- ARN detected: Use AWS KMS with Hierarchical Keyring
- Environment variable: Use raw key material (dev/containers)

---

## Security Requirements Met

| Requirement | Implementation | Status |
|-------------|-----------------|--------|
| Envelope encryption (DEK + KEK) | AWS Encryption SDK AESGCMSIV | ✅ |
| Fresh DEK per session | AWS SDK generates per operation | ✅ |
| Context binding (AAD) | service_id bound at DEK and KEK layers | ✅ |
| Authenticated encryption | AESGCMSIV provides integrity verification | ✅ |
| KEK security | AWS KMS (prod) or memguard (dev) | ✅ |
| KEK rotation support | Backward compatibility via version byte | ✅ |
| Memory protection | Memguard for Branch Keys, tokens, DEKs | ✅ |
| Core dump exclusion | MADV_DONTDUMP via memguard | ✅ |
| No plaintext logging | Sanitized logging documented | ✅ |
| Fail-closed behavior | No fallback on any encryption/decryption failure | ✅ |

---

## Domain Model Summary

### Entities (with unique identity)

None added; encryption is transparent to existing entities.

### Aggregates (consistency boundaries)

- **UserSession** (existing): Now contains EncryptedAccessToken and EncryptedRefreshToken BYTEA fields

### Value Objects (no identity)

- **EncryptionContext**: `{"service_id": "oauth2"}` (single-field map)

### Domain Events (added)

- **SessionEncrypted**: Tokens encrypted successfully, session stored
- **SessionDecrypted**: Tokens decrypted successfully, returned to caller
- **SessionEncryptionFailed**: Encryption failed, storage prevented
- **SessionDecryptionFailed**: Decryption failed, retrieval failed

### Error Types

| ErrorKind | Cause | HTTP Status |
|-----------|-------|-------------|
| `encryption_failed` | DEK generation or AESGCMSIV failed | 500 |
| `decryption_failed` | DEK unwrap or token decryption failed | 500 |
| `context_mismatch` | Encryption context doesn't match (wrong service_id) | 403 |
| `integrity_violation` | Auth tag verification failed (possible tampering) | 403 |
| `kek_unavailable` | AWS KMS unreachable or env var not set | 503 |

---

## API Contracts Defined

### EncryptionPort Interface

```go
type EncryptionPort interface {
    Encrypt(ctx context.Context, plaintext []byte, encryptionContext map[string]string) ([]byte, error)
    Decrypt(ctx context.Context, ciphertext []byte, encryptionContext map[string]string) ([]byte, error)
}
```

### Configuration Interface

```yaml
encryption:
  key_encryption_key: <value>  # ARN or ${ENV_VAR}
  kms:
    branch_key_ttl: 15m        # Recommended for OAuth
    cache_limit_entries: 100   # Concurrent service contexts
    memory_protection: true     # Enable memguard
```

### Error Interface

```go
type EncryptionError struct {
    Kind    ErrorKind
    Message string // Sanitized
    Wrapped error
}
```

---

## Testing Strategy (Phase 2 Design)

### E2E Acceptance Tests (24 scenarios)

| User Story | Scenarios | Test Focus |
|------------|-----------|-----------|
| US1: Envelope Encryption | 4 | DEK context binding, KEK wrapping, fail-closed |
| US2: Secure KEK Storage | 4 | AWS KMS, logging, access control, rotation |
| US3: DEK Generation | 3 | Fresh per session, unique, memory protection |
| US4: Env Var KEK | 3 | Load from env var, use for wrap/unwrap, persist |
| US5: Transparent Encryption | 3 | Repository Create/Get automatic encryption |
| US6: Cross-Service Prevention | 3 | Same service succeeds, different fails, reuse fails |
| US7: Post-Quantum Ready | 4 | PQC available, with/without PQC |

### Test Execution (TDD Red-Phase-First)

1. **Phase 2f (Design)**: Write all 24 tests FIRST (will all fail)
2. **Verify Red**: `ginkgo -v ./tests/e2e/encryption_vault_test.go` → all FAIL
3. **Implementation**: Build adapters, services, configuration incrementally
4. **Verify Green**: Tests turn GREEN as features complete
5. **No Test Changes**: Only fixture/data adjustments during implementation

---

## Phase 1 Constitution Check: ✅ PASSED

All design preconditions verified:

- [x] Domain model (UserSession, EncryptionContext, EncryptionPort)
- [x] Domain concepts (envelope encryption, DEK, KEK, AAD, etc.)
- [x] Configuration design (single field, ${env_var} interpolation)
- [x] Configuration examples (AWS KMS and env var)
- [x] API design (EncryptionPort, service-layer ownership)
- [x] Database design (BYTEA and JSONB columns exist)
- [x] Security-first (fail-closed, memory protection, context verification)
- [x] Library-first (AWS Encryption SDK, memguard, no custom crypto)
- [x] Hexagonal architecture (EncryptionPort interface, AWS adapter, storage adapters)
- [x] Persistence patterns (existing UserSession aggregate, transparent encryption)

**Gate**: ✅ Ready for Phase 2 implementation

---

## Next Steps (Phase 2)

### Phase 2a: E2E Test Design
- Write all 24 acceptance tests in Ginkgo/Gomega
- Verify tests fail (red phase)
- Identify test infrastructure needs (bootstrap, fixtures)

### Phase 2b: Infrastructure
- Testcontainers setup for PostgreSQL
- LocalStack setup for AWS KMS (optional for tests)
- Environment variable KEK configuration

### Phase 2c: Implementation
- AWS Encryption SDK adapter with Hierarchical Keyring
- Memguard-protected custom cache implementation
- Service-layer encryption/decryption in OAuth2SessionService
- Configuration system integration
- Domain event publishing

### Phase 2d: Storage Adapter Updates
- Memory adapter: Pass-through for encrypted tokens (no changes to logic)
- PostgreSQL adapter: Pass-through for encrypted tokens (no changes to logic)

### Phase 2e: Testing & Validation
- All E2E tests green
- Unit test coverage for adapters and services
- Performance benchmarking vs. targets
- Security audit

---

## Summary of Design Decisions

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| **Encryption Ownership** | Service layer | Clear separation, easier testing |
| **Key Architecture** | Hierarchical Keyring + Branch Key Caching | Performance + cost vs. security |
| **Branch Key TTL** | 15 minutes (recommended) | Balance between security window and KMS availability |
| **Context Fields** | service_id only | Optimizes KMS operations, provides service isolation |
| **Memory Protection** | Memguard for Branch Keys and token transit | Prevents core dumps, swap, and memory pressure eviction |
| **Error Handling** | Fail-closed, no fallback | Security-first, no degradation |
| **Configuration** | Single field with auto-detection | Simplifies ops, leverages existing config system |

---

## Document Navigation

### For Implementation Teams
- Start with [quickstart.md](./quickstart.md) - Developer integration patterns
- Reference [encryption-port.md](./contracts/encryption-port.md) - Interface specification
- Study [memguard-integration.md](./contracts/memguard-integration.md) - Production implementation

### For Security Review
- Review [aws-kms-architecture.md](./contracts/aws-kms-architecture.md) - Architecture decision
- Review [error-contract.md](./contracts/error-contract.md) - Error handling and security events
- Review [domain-events.md](./contracts/domain-events.md) - Audit and monitoring

### For Operations
- Reference [configuration.md](./contracts/configuration.md) - Configuration and deployment
- Review [quickstart.md](./quickstart.md) section 7 - Monitoring and logging
- Review [data-model.md](./data-model.md) - Understanding the architecture

### For Testing
- Reference [plan.md](./plan.md) - 24 acceptance scenarios and testing strategy
- Reference [quickstart.md](./quickstart.md) section 6 - Testing patterns

---

## Artifacts Checklist

**Documentation**:
- [x] spec.md (feature specification)
- [x] plan.md (implementation plan)
- [x] research.md (Phase 0 research)
- [x] data-model.md (domain model)
- [x] quickstart.md (integration guide)
- [x] PHASE-1-SUMMARY.md (this document)

**Contracts**:
- [x] encryption-port.md (interface specification)
- [x] configuration.md (configuration contract)
- [x] error-contract.md (error types and handling)
- [x] domain-events.md (domain events)
- [x] aws-kms-architecture.md (KMS architecture decision)
- [x] memguard-integration.md (memguard implementation)

**Missing (Phase 2)**:
- [ ] tests/e2e/encryption_vault_test.go (24 acceptance tests)
- [ ] internal/adapters/encryption/aws/adapter.go (AWS SDK adapter)
- [ ] internal/adapters/encryption/aws/adapter_test.go (adapter tests)
- [ ] internal/domain/encryption/errors.go (error types)
- [ ] ARCHITECTURE.md updates (glossary and domain terminology)
- [ ] adrs/NNNN-envelope-encryption-design.md (architecture decision record)

---

**Phase 1 Status**: ✅ COMPLETE

**Ready for**: Phase 2 Implementation

**Last Updated**: 2026-01-16

**Version**: 1.0
