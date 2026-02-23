# ADR 012: Encryption Layer Separation (Domain vs. Storage Adapter)

## Status
Accepted

## Context
The project uses OAuth2 tokens that require encryption at rest. Two approaches are possible:

1. **Adapter Encryption**: Repository adapter handles encryption/decryption
   - Example: ThirdpartyServiceRepository (historical)
   - Adapter receives plaintext, returns plaintext
   - Repository has EncryptionPort dependency

2. **Domain Encryption**: Domain service handles encryption/decryption
   - Example: UserSessionRepository (current recommended)
   - Service encrypts before calling repository
   - Repository operates on encrypted bytes
   - Repository has no crypto dependencies

## Decision
Adopt "Domain Encryption" as the standard for all new features.

## Rationale

### Hexagonal Architecture Purity (++++)
- Domain service owns business logic (including data protection)
- Repository is pure persistence layer (no crypto concerns)
- Cleaner dependency graph: domain → port, adapter → domain+port

### Separation of Concerns (+++)
- Service layer: Business logic + encryption orchestration
- Repository layer: CRUD operations only
- Each layer has single responsibility

### Adapter Simplicity (++)
- Repositories: 50-100 lines (CRUD only)
- No encryption logic in adapters
- Easier to add new storage backends

### Testability (++)
- Repository tests: Test persistence without mocking encryption
- Service tests: Test encryption logic without storage
- Clear test boundaries

### Consistency (+)
- Same pattern across all sensitive data entities
- Reduced cognitive load for developers
- Clearer code review standards

## Consequences

### Benefits
✅ Domain logic clearly visible in service layer
✅ Repositories are infrastructure-agnostic
✅ Easy to swap storage backends
✅ Simpler error handling (no crypto logic in adapters)

### Trade-offs
⚠️  Requires Pattern A (ThirdpartyService) to migrate
⚠️  Documentation needed (architecture not obvious)
⚠️  Service layer becomes more complex

## Related
- ADR 004: Storage Layer Architecture (hexagonal pattern)
- ADR 008: Encryption Context Optimization (service_id only)
- Feature 012: Token Vault (envelope encryption implementation)
