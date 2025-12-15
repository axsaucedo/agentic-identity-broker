<!--
Sync Impact Report
==================
Version Change: 1.1.0 → 1.2.0
Rationale: Materially expanded Principle II (Architecture Documentation & Decision Records) to make ADRs
  binding and added Principle IX (Persistence Pattern Consistency) to enforce established storage patterns
  from quickstart.md for all new entities requiring persistence.

Modified Principles:
- II. Architecture Documentation & Decision Records: Added binding requirement that ADRs are not merely
  documentation but binding architectural decisions that MUST be followed

Added Principles:
- IX. Persistence Pattern Consistency (NEW)

Added Sections: None (integrated into Core Principles)

Removed Sections: None

Templates Status:
- ✅ plan-template.md: Constitution Check section includes ADR and persistence pattern requirements
- ✅ spec-template.md: Requirements section aligned with persistence patterns
- ✅ tasks-template.md: Task phases support persistence implementation per quickstart.md
- ✅ ARCHITECTURE.md: Will be updated with persistence subsystem (004-persistence-layer)
- ✅ specs/004-persistence-layer/quickstart.md: Established as canonical persistence pattern reference
- ✅ adrs/: ADR 004 (Storage Layer Architecture) established as binding pattern

Follow-up TODOs: None - all principles fully defined

Rationale for Principle II Enhancement:
  ADRs document major architectural decisions but must be understood as binding constraints, not optional
  suggestions. When an ADR establishes a pattern (e.g., ADR 004: interface segregation, sqlx for PostgreSQL),
  all subsequent code must follow that pattern unless a new ADR explicitly supersedes it. This prevents
  architectural drift and ensures consistency.

Rationale for Principle IX (Persistence Pattern Consistency):
  The 004-persistence-layer feature established comprehensive persistence patterns documented in
  specs/004-persistence-layer/quickstart.md including: small focused interfaces (StorageLifecycle,
  per-entity repositories), hexagonal architecture separation, sqlx/pgx usage, error wrapping patterns,
  and testing strategies. This principle codifies that all future entities requiring persistence MUST
  follow these established patterns to maintain architectural consistency.
-->

# Agentic Identity Broker Constitution

## Core Principles

### I. Security-First Development

Security is NON-NEGOTIABLE and MUST NOT be bypassed or made optional in this codebase.

**Rules**:
- Security features MUST be enabled by default; disabling them MUST require explicit configuration
- Signature validation is NEVER optional—no code path shall skip cryptographic verification
- Encryption MUST NOT have accidental fallbacks to plaintext or weaker modes
- Security controls MUST fail closed: if verification fails, the operation MUST abort
- All security-critical operations MUST be auditable via structured logging

**Rationale**: This project guarantees security for agentic platforms. A single bypass or default-off security control undermines the entire trust model and exposes downstream systems to compromise.

### II. Architecture Documentation & Decision Records

Architecture and major decisions MUST be documented, and Architecture Decision Records are BINDING.

**Rules**:
- [ARCHITECTURE.md](ARCHITECTURE.md) is the single source of truth for system architecture
- When features change or touch architecture, [ARCHITECTURE.md](ARCHITECTURE.md) MUST be updated in the same PR
- Non-functional requirements (performance, security, scalability) MUST be documented in [ARCHITECTURE.md](ARCHITECTURE.md)
- Major architectural decisions MUST be recorded in [adrs/](adrs/) directory as Architecture Decision Records
- ADR files MUST follow the format `NNN-decision-title.md` (e.g., `001-hexagonal-architecture.md`)
- ADRs MUST include: Context, Decision, Consequences, Status (Proposed/Accepted/Deprecated/Superseded)
- **ADRs with status "Accepted" are BINDING**: all code MUST follow patterns and decisions documented in accepted ADRs
- Deviation from accepted ADRs is NOT permitted without creating a new superseding ADR
- When an ADR establishes a pattern (e.g., interface design, library choice, architectural pattern), that pattern MUST be followed consistently across the codebase
- ADRs supersede ad-hoc implementation choices: if an ADR exists for a domain (e.g., ADR 004 for storage), implementers MUST follow it

**Rationale**: Agents and developers require rapid, accurate understanding of architectural constraints to contribute effectively and maintain consistency across the system. Making ADRs binding (not merely advisory) prevents architectural drift, ensures pattern consistency, and makes architectural governance explicit and enforceable.

**Example**: ADR 004 (Storage Layer Architecture) establishes interface segregation, sqlx for PostgreSQL, and specific error handling patterns. Any new storage-backed entity must follow these decisions, not introduce alternative approaches.

### III. Library-First Security Implementation

Cryptographic and security features MUST use battle-tested libraries, NEVER custom implementations.

**Rules**:
- Cryptography MUST be implemented via established libraries (e.g., Go crypto/*, golang.org/x/crypto)
- Custom cryptographic implementations are FORBIDDEN in this codebase
- Security primitives (signing, verification, encryption, key derivation) MUST delegate to vetted libraries
- If a security requirement cannot be satisfied with existing libraries, implementation MUST STOP and the constraint MUST be escalated for clarification

**Rationale**: Cryptographic implementations are notoriously error-prone. This project's security guarantees depend on correct cryptography; rolling our own introduces unacceptable risk.

### IV. Documentation & API Transparency

End-user documentation MUST be comprehensive, accurate, and kept in sync with implementation.

**Rules**:
- The [docs/](docs/) directory contains all end-user documentation
- When a new feature is implemented, end-user documentation MUST be updated in the same PR
- All public APIs (HTTP, gRPC, CLI) MUST be documented in [docs/](docs/)
- API documentation MUST include: purpose, parameters, return values, error conditions, examples
- Breaking changes to APIs MUST be clearly marked in documentation and changelogs

**Rationale**: Users of this broker require clear, correct documentation to integrate securely. Outdated or missing API docs lead to misconfiguration and security vulnerabilities.

### V. Domain-Driven Design & Glossary Management

Domain concepts MUST be explicitly modeled, documented, and maintained in the glossary.

**Rules**:
- This project follows Domain-Driven Design (DDD) principles
- When the domain model is updated or new domain concepts are introduced, they MUST be added to [ARCHITECTURE.md](ARCHITECTURE.md)
- All domain terms MUST be added to the Glossary section in [ARCHITECTURE.md](ARCHITECTURE.md)
- Glossary entries MUST include: term name, definition, relationships to other domain concepts
- Ubiquitous language from the domain model MUST be used consistently in code, docs, and discussions

**Rationale**: Consistent domain language prevents ambiguity, reduces cognitive load, and ensures that code structure reflects the problem domain.

### VI. Hexagonal Architecture & Clean Boundaries

Backend architecture MUST use hexagonal architecture with clear port/adapter separation.

**Rules**:
- The backend (Go) follows hexagonal architecture (also known as ports and adapters)
- Domain logic MUST depend on ports (interfaces), NOT concrete implementations
- Adapters MUST implement ports: driving adapters (inbound, e.g., HTTP handlers) and driven adapters (outbound, e.g., database clients)
- This is NOT dogmatic: pragmatic deviations are allowed, but domain logic MUST remain insulated via interfaces
- The directory structure MUST be explained in [ARCHITECTURE.md](ARCHITECTURE.md)
- If directory structure is unclear or misleading, it MUST be clarified with the user and updated in [ARCHITECTURE.md](ARCHITECTURE.md)

**Rationale**: Hexagonal architecture enables testability, flexibility, and maintainability by decoupling domain logic from infrastructure concerns.

### VII. Configuration-Driven Design

All runtime configuration MUST use the unified configuration system; ad-hoc configuration is forbidden.

**Rules**:
- Features MUST NOT implement custom configuration loading; they MUST use the system-wide configuration port defined in [internal/ports/config.go](internal/ports/config.go)
- All configuration settings MUST support multiple sources (files, environment variables, CLI flags) with clear precedence
- Configuration structure MUST be defined in [internal/config/schema.go](internal/config/schema.go) with validation rules enforced at startup
- End-user documentation for feature-specific configuration MUST be added to [docs/configuration.md](docs/configuration.md)
- Feature-specific configuration examples MUST be added to [examples/config/](examples/config/) directory
- Configuration guide [examples/config/README.md](examples/config/README.md) MUST be referenced and updated as new features add configuration options
- See [Flexible Configuration Feature Documentation](docs/configuration.md) for complete usage guidance
- Implementation reference: [Feature 002 - Flexible Configuration](specs/002-flexible-configuration/)

**Rationale**: Unified configuration prevents duplication, ensures consistent precedence rules across
the system, reduces operational confusion, and simplifies deployment across development/staging/production
environments. The 002-flexible-configuration feature established this system; all features MUST integrate
with it rather than bypassing it.

### VIII. Test-Driven Development & Automated Testing

Code quality and correctness MUST be ensured through automated tests, not manual Bash validation.

**Rules**:
- New features MUST include automated tests (unit, integration, or both) written before or alongside implementation (TDD)
- Automated tests MUST cover happy paths, error cases, and edge cases
- Test files MUST use the Go standard library testing package or established testing frameworks (testify for assertions only)
- Table-driven tests MUST be used for validation logic and parameterized scenarios
- Bash scripts MUST NOT be used for code correctness validation; they are reserved for infrastructure tasks (CI/CD, deployment, system-level checks)
- Test coverage MUST be verifiable via `go test ./... -cover` for Go packages
- Tests MUST be maintainable, readable, and include clear assertions with meaningful failure messages
- Integration tests MUST test real behavior end-to-end (e.g., actual file I/O, database operations)
- Unit tests MUST isolate functionality via mocking/interfaces where appropriate
- See [task templates](./templates/tasks-template.md) for test organization patterns

**Rationale**: Automated tests catch regressions early, document expected behavior, enable refactoring
with confidence, and scale better than manual validation. Bash-based validation is fragile and
unmaintainable; it MUST be reserved for infrastructure concerns (e.g., smoke tests in CI/CD) rather
than code correctness.

### IX. Persistence Pattern Consistency

All entities requiring persistence MUST follow established patterns documented in quickstart.md.

**Rules**:
- New entities requiring persistence MUST follow patterns in [specs/004-persistence-layer/quickstart.md](specs/004-persistence-layer/quickstart.md)
- Storage interfaces MUST use small, focused interfaces following Interface Segregation Principle:
  - Separate repository interface per entity (e.g., `UserRepository`, `ProductRepository`)
  - Lifecycle operations in separate interface (`StorageLifecycle`)
  - Maximum 5-7 methods per repository interface
- Repository interfaces MUST be defined in [internal/ports/storage.go](internal/ports/storage.go)
- Adapters MUST implement all repository interfaces:
  - In-memory adapter in `internal/adapters/storage/memory/` (for development/testing)
  - PostgreSQL adapter in `internal/adapters/storage/postgres/` (for production)
- PostgreSQL adapters MUST use sqlx library (not raw database/sql or ORM)
- All storage errors MUST be wrapped in domain `StorageError` type (never expose adapter-specific errors)
- Storage operations MUST respect configured timeouts (5s read, 10s write defaults)
- Adapter factory pattern MUST be used: return `*Adapter` struct with accessor methods
- Domain services MUST depend on specific repository interfaces, NOT concrete adapters
- All new repositories MUST include:
  - Unit tests for both in-memory and PostgreSQL adapters
  - Integration tests for PostgreSQL adapter using testcontainers
  - Table-driven tests for validation logic
- See [ADR 004: Storage Layer Architecture](adrs/004-storage-layer-architecture.md) for binding architectural decisions

**Rationale**: The 004-persistence-layer feature established comprehensive, battle-tested persistence
patterns following Go best practices (small interfaces, Interface Segregation Principle, hexagonal
architecture). Requiring all entities to follow these patterns ensures consistency, maintainability,
and prevents ad-hoc persistence implementations that violate architectural principles. The patterns
are documented in quickstart.md and enforced via ADR 004.

**Example**: When adding a `Session` entity that needs persistence, the implementer must:
1. Define `SessionRepository` interface in `internal/ports/storage.go`
2. Implement methods in both memory and postgres adapters
3. Add `Sessions() SessionRepository` accessor to Adapter struct
4. Use sqlx for PostgreSQL queries (not GORM or raw database/sql)
5. Wrap all errors in `domain.StorageError`
6. Write unit and integration tests following quickstart.md patterns

## Development Requirements

### Compliance Checklist

Before any feature PR is merged, verify:

- [ ] Security controls are enabled by default and fail closed
- [ ] [ARCHITECTURE.md](ARCHITECTURE.md) reflects architectural changes (if any)
- [ ] Major decisions recorded in [adrs/](adrs/) with correct numbering
- [ ] Code follows patterns established in accepted ADRs (especially ADR 004 for persistence)
- [ ] End-user documentation in [docs/](docs/) updated for new/changed APIs
- [ ] Domain concepts added to [ARCHITECTURE.md](ARCHITECTURE.md) Glossary section
- [ ] Domain logic uses ports (interfaces) and adapters are separated
- [ ] No custom cryptography; security features use vetted libraries
- [ ] Structured logging present for security-critical operations
- [ ] New features use the system configuration port, not custom config loading
- [ ] Automated tests included (unit, integration, or both) with meaningful coverage
- [ ] No Bash scripts used for code correctness validation (only infrastructure tasks)
- [ ] New persistence entities follow quickstart.md patterns (if applicable)

### When Constraints Cannot Be Met

If Principle II (Binding ADRs) requires deviation:

1. STOP implementation immediately
2. Document why the accepted ADR pattern cannot be followed
3. Draft a new ADR proposing an alternative approach with rationale
4. Mark the new ADR as superseding the previous ADR
5. Do NOT implement the deviation without an accepted superseding ADR

If Principle III (Library-First Security) cannot be satisfied:

1. STOP implementation immediately
2. Document the specific security requirement and why existing libraries are insufficient
3. Escalate to project maintainers for guidance
4. Do NOT proceed with custom cryptographic code without explicit approval

If Principle VII (Configuration-Driven Design) cannot be satisfied:

1. STOP implementation immediately
2. Document the specific configuration requirement and why the unified system is insufficient
3. Escalate to project maintainers for guidance
4. Do NOT implement custom configuration without explicit approval

If Principle VIII (TDD & Automated Testing) cannot be satisfied:

1. Document the specific reason why automated testing is not feasible
2. Propose alternative validation approach (e.g., infrastructure-level testing)
3. Escalate to project maintainers for exception approval
4. Do NOT use Bash scripts for code correctness validation without explicit justification

If Principle IX (Persistence Pattern Consistency) cannot be satisfied:

1. STOP implementation immediately
2. Document why quickstart.md patterns cannot be followed for this specific entity
3. Propose alternative approach with technical justification
4. Create an ADR documenting the exception and rationale
5. Do NOT implement non-standard persistence patterns without an accepted ADR

## Governance

### Amendment Procedure

1. Propose amendment via PR to [.specify/memory/constitution.md](.specify/memory/constitution.md)
2. Include rationale, impacted templates, and migration plan
3. Increment CONSTITUTION_VERSION per semantic versioning rules
4. Update Sync Impact Report with changes
5. Propagate changes to dependent templates and documentation

### Versioning Policy

- **MAJOR**: Backward incompatible governance/principle removals or redefinitions
- **MINOR**: New principle/section added or materially expanded guidance
- **PATCH**: Clarifications, wording, typo fixes, non-semantic refinements

### Compliance Review

- All PRs MUST verify compliance with this constitution
- Reviewers MUST challenge complexity and request justification when principles are violated
- Reviewers MUST verify adherence to accepted ADRs (Principle II)
- Reviewers MUST verify persistence implementations follow quickstart.md patterns (Principle IX)
- Template files in [.specify/templates/](.specify/templates/) provide execution workflows that enforce these principles

**Version**: 1.2.0 | **Ratified**: 2025-12-14 | **Last Amended**: 2025-12-15
