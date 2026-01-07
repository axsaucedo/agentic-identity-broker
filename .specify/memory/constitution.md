<!--
Sync Impact Report
==================
Version Change: 1.5.1 → 1.6.0
Rationale: MINOR version bump - new principle added for explicit governance of dependency injection patterns.
  This is additive guidance (Principle XII) that establishes binding rules for service instantiation and
  HTTP route registration, preventing architectural drift and enabling clear separation of concerns between
  wiring logic (Builder pattern) and route registration (routing functions).

Modified Principles:
- None (existing principles unchanged - Principle XII is new and complementary)

Added Principles:
- Principle XII: Dependency Injection & Component Wiring (centralizes DI pattern governance)

Added Sections:
- Compliance Checklist: Added Principle XII verification items
- When Constraints Cannot Be Met: Added Principle XII exception handling

Removed Sections:
- None

Templates Status:
- ⚠ tasks-template.md: PENDING - should add Principle XII verification task to Phase N
  - Location: Phase N: Constitution Compliance Verification
  - Suggested task: Verify dependency injection follows Builder pattern; routing functions don't instantiate
- 🔲 spec-template.md: No updates needed
- 🔲 plan-template.md: No updates needed (Constitution Check already covers architecture)

Follow-up TODOs: None - Principle XII is complete and self-contained

Previous Version History:
- 1.5.0 → 1.5.1: Clarified testing requirements in tasks-template.md (PATCH)
- 1.4.0 → 1.5.0: Added Governance > Task List Requirements section (MINOR)
- 1.3.1 → 1.4.0: Added Principle XI (Design System Compliance & Consistency) (MINOR)

Rationale for Principle XII (Dependency Injection & Component Wiring):
  Separating construction (Builder pattern) from routing (dedicated routing functions) maintains clean
  architectural boundaries, enables testability through dependency injection, prevents circular dependencies,
  and makes the dependency graph explicit and auditable. The Builder pattern centralizes wiring logic, making
  it easy to understand initialization order and dependencies at a glance. Routing functions remain thin and
  focused solely on route registration, which prevents accidental coupling to internal implementation details.
  This pattern follows hexagonal architecture principles (Principle VI): domain logic (Builder) is isolated
  from infrastructure concerns (routing), and dependencies flow from configuration/infrastructure toward
  domain logic, not vice versa. Establishing explicit rules prevents architectural drift.
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

### IV. API Documentation & OpenAPI Transparency

All APIs (end-user and administrative) MUST be documented in OpenAPI format before implementation,
and API contracts are binding commitments to consumers.

**Rules**:
- End-user APIs (Backend for Frontend, OAuth2 flows) MUST be documented in `/api/enduser/openapi.yaml`
- Administrative/system APIs MUST be documented in `/api/admin/openapi.yaml`
- All API documentation MUST use OpenAPI 3.0+ specification format
- API documentation MUST include: endpoint paths, HTTP methods, parameters, request/response bodies,
  error codes, examples, and authentication requirements
- End-user API documentation MUST also be rendered in [docs/api/](docs/api/) directory with examples for
  typical use cases and integration patterns
- **API changes MUST be confirmed by the user/stakeholder before implementation begins**
- API changes MUST NOT be made to fix bugs or address edge cases without explicit user/stakeholder confirmation
- When an API change is confirmed, the OpenAPI specification MUST be updated in the same PR as implementation
- All APIs MUST follow Zalando RESTful API and Event Guidelines
  (https://opensource.zalando.com/restful-api-guidelines/)
- API design decisions (e.g., naming conventions, error response format, pagination strategy) MUST be
  documented in [ARCHITECTURE.md](ARCHITECTURE.md) or via ADR
- Breaking API changes MUST be documented in [docs/changelog.md](docs/changelog.md)

**Rationale**: APIs are contracts with consumers. Explicit OpenAPI documentation in designated locations
ensures clients can discover, understand, and implement integrations correctly. Designing APIs before
implementation ensures contracts are stable and well-thought-out. Requiring user confirmation for API changes
prevents silent breaking changes and establishes APIs as commitments. Following Zalando guidelines ensures
consistency with industry best practices for API design.

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

### IX. Persistence Pattern Consistency & Database Migration Management

All entities requiring persistence MUST follow established patterns, and ALL database schema changes
MUST be managed via migrations with comprehensive testing.

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
- **Database Schema Changes MUST be managed via migrations (not ad-hoc schema edits)**:
  - ALL schema changes MUST be in `/migrations/` directory using go-migrate naming conventions
  - Migration file format: `[NNN]_[description].up.sql` and `[NNN]_[description].down.sql` where NNN is sequential (001, 002, 003...)
  - Example: `004_create_sessions_table.up.sql`, `004_create_sessions_table.down.sql`
  - Migration files MUST include both UP (apply schema change) and DOWN (rollback) operations
  - Each migration MUST be atomic: either fully apply or fully rollback on failure
- **All migrations MUST be tested**:
  - PostgreSQL integration tests MUST verify each migration applies cleanly
  - Integration tests MUST verify rollback functionality (DOWN migrations)
  - Rollback tests MUST verify data integrity: migrations can be applied and rolled back repeatedly
    without data loss or corruption
  - Integration tests MUST run against real PostgreSQL instance (e.g., via testcontainers)
- **All PostgreSQL-backed repositories MUST be tested in integration tests**:
  - Every repository adapter (in `/internal/adapters/storage/postgres/`) MUST have integration tests
  - Integration tests MUST verify persistence behavior against real PostgreSQL
  - Integration tests MUST use the same migrations that production uses
  - Integration tests MUST validate error handling and edge cases (constraints, timeouts, etc.)
- See [ADR 004: Storage Layer Architecture](adrs/004-storage-layer-architecture.md) for binding architectural decisions

**Rationale**: The 004-persistence-layer feature established comprehensive, battle-tested persistence
patterns following Go best practices (small interfaces, Interface Segregation Principle, hexagonal
architecture). Requiring all entities to follow these patterns ensures consistency, maintainability,
and prevents ad-hoc persistence implementations that violate architectural principles. Database
migrations are a contract with production systems and must be auditable, reversible, and tested.
Using go-migrate naming conventions ensures consistency and prevents merge conflicts. Testing all
migrations and repositories in integration tests validates that schema and persistence layers work
correctly with real PostgreSQL.

**Example**: When adding a `Session` entity that needs persistence, the implementer must:
1. Define `SessionRepository` interface in `internal/ports/storage.go`
2. Create migration files: `004_create_sessions_table.up.sql` and `004_create_sessions_table.down.sql` (increment sequence number)
3. Write integration test to verify migration applies and rolls back cleanly
4. Implement methods in both memory and postgres adapters
5. Add `Sessions() SessionRepository` accessor to Adapter struct
6. Use sqlx for PostgreSQL queries (not GORM or raw database/sql)
7. Wrap all errors in `domain.StorageError`
8. Write unit and integration tests following quickstart.md patterns
9. Verify all repository methods work in integration tests against real PostgreSQL

### X. API-First Development

APIs MUST be designed and documented before implementation begins, and all API changes require
explicit user/stakeholder confirmation. API contracts are treated as commitments that cannot be
changed without confirmation, even to fix bugs.

**Rules**:
- API design MUST occur before implementation: OpenAPI spec created and reviewed with user/stakeholder FIRST
- Implementation MUST follow the confirmed OpenAPI specification exactly
- All API changes (additions, modifications, removals) MUST be confirmed by the user/stakeholder in writing
  before implementation begins
- **API changes for bug fixes are NOT exempt**: even corrections to API behavior require user confirmation
- Rationale: A bug in one system may be relied upon in another system. API changes must be conscious choices,
  not collateral damage
- User confirmation should be documented via PR review comments, issue discussions, or design documents
- Once an API is deployed/released, NEVER change its behavior without a major version bump or creating a new endpoint
- Deprecation warnings MUST be added to OpenAPI docs before removing endpoints
- When user requests API change during implementation, create ADR documenting the change with confirmation reference
- API versioning strategy MUST be documented in [ARCHITECTURE.md](ARCHITECTURE.md)

**Rationale**: Designing APIs before implementation ensures that system contracts are stable,
well-thought-out, and user-centric. Requiring user/stakeholder confirmation makes API changes explicit
and prevents accidental breaking changes. Treating API bugs as design issues (requiring confirmation to
change) rather than simple fixes establishes accountability for the API contract and prevents cascading
failures in dependent systems. Systems built on APIs assume behavior is intentional; changing API behavior
without confirmation breaks that trust.

### XI. Design System Compliance & Consistency

All frontend components MUST use the design system; universal patterns MUST be contributed back to
ensure consistency, accessibility, and brand identity.

**Rules**:
- All frontend components MUST be built using the design system located at `web/src/design-system/`
- Documentation entrypoint is [web/src/design-system/docs/INDEX.md](../web/src/design-system/docs/INDEX.md)
- Before building new components, implementers MUST:
  - Review [DECISION_TREES.md](../web/src/design-system/docs/DECISION_TREES.md) for variant selection guidance
  - Reference [COMPONENT_PAIRING_GUIDE.md](../web/src/design-system/docs/COMPONENT_PAIRING_GUIDE.md) for composition patterns
  - Study [COMMON_MISTAKES.md](../web/src/design-system/docs/COMMON_MISTAKES.md) to avoid anti-patterns
  - Use semantic color tokens (`trust-deep`, `success-primary`, `neutral-*`) NOT extended palettes (`navy-*`, `emerald-*`, `gray-*`)
  - Follow [DESIGN_PRINCIPLES.md](../web/src/design-system/docs/DESIGN_PRINCIPLES.md) for Refined Trust Architecture aesthetic
- New application-specific components MUST use design system primitives (Button, Card, Badge, etc.) not custom implementations
- If a component is universally applicable (not app-specific), it MUST be added to the design system in `web/src/design-system/components/`
- All design system components MUST follow:
  - Refined Trust Architecture aesthetic (navy brand colors #0A2540, warm neutrals, Crimson Pro serif headings)
  - WCAG 2.1 AA accessibility standards (4.5:1 text contrast minimum, 3:1 UI component contrast minimum)
  - Semantic HTML with proper ARIA attributes for screen reader support
  - Tailwind CSS v4 with @theme directive for design tokens
  - Class Variance Authority (CVA) for component variants
- Component-specific styling MUST NOT bypass design tokens or introduce custom CSS that breaks visual consistency
- Storybook stories MUST be included for all new design system components with visual regression testing
- Color usage MUST follow semantic tokens defined in [TOKEN_GUIDE.md](../web/src/design-system/docs/TOKEN_GUIDE.md)
- Typography MUST use font families: Crimson Pro (headings), Manrope (body), JetBrains Mono (code/technical values)

**Rationale**: Design system compliance ensures visual consistency across the application, maintains brand
identity (Refined Trust Architecture), reduces development time through component reuse, and guarantees
accessibility standards (WCAG 2.1 AA). The design system documentation provides comprehensive guidance
enabling 92-95% autonomous decision-making for AI agents building components (verified via independent
ui-designer assessment). Requiring universal components to be contributed back prevents fragmentation
and ensures patterns are shared across the application. Semantic tokens prevent color inconsistencies and
make theming possible. This principle establishes frontend quality standards and prevents ad-hoc styling
that undermines user experience and accessibility.

### XII. Dependency Injection & Component Wiring

All service instantiation and dependency wiring MUST occur via the Builder pattern in
[internal/app/builder.go](internal/app/builder.go). HTTP route registration MUST be exclusively
performed by routing functions in [internal/adapters/http/routing/admin.go](internal/adapters/http/routing/admin.go)
and [internal/adapters/http/routing/enduser.go](internal/adapters/http/routing/enduser.go); routing files
MUST NOT instantiate or wire services.

**Rules**:
- All service instantiation MUST use the Builder pattern (`internal/app/builder.go`)
- The Builder receives all dependencies (config, storage, logger, encryption, etc.) via `With*()` methods
- Services MUST be constructed in the `Build()` method respecting dependency ordering and initialization phases
- HTTP route registration MUST use dedicated routing functions: `SetupAdminRoutes()` and `SetupEnduserRoutes()`
- Routing functions MUST receive pre-wired handlers/services via parameter, NEVER instantiate them
- Routing functions MUST be in `internal/adapters/http/routing/` directory with naming convention `Setup*Routes()`
- Routing functions MUST only register routes and middleware, NOT create instances or perform wiring
- Services MUST be passed to routing functions as part of handler structs (e.g., `app.AdminHandlers`, `app.EnduserHandlers`)
- If a handler needs a service, it MUST be injected via the App builder, not created in routing or handlers
- All adapters (handlers, middleware, services) MUST depend on ports (interfaces), not concrete implementations
- Configuration passed to routing functions MUST be read-only; routing functions MUST NOT modify application state
- Routing functions MUST follow the pattern: receive router and pre-wired dependencies, register routes, return nothing

**Rationale**: Separating construction (Builder) from routing (dedicated routing functions) maintains
clean architectural boundaries, enables testability through dependency injection, prevents circular
dependencies, and makes the dependency graph explicit and auditable. The Builder pattern centralizes
wiring logic, making it easy to understand initialization order and dependencies at a glance. Routing
functions remain thin, focused solely on route registration, which prevents accidental coupling to
internal implementation details. This pattern follows hexagonal architecture principles (Principle VI):
domain logic (Builder) is isolated from infrastructure concerns (routing), and dependencies flow from
configuration/infrastructure toward domain logic, not vice versa. Establishing explicit rules prevents
architectural drift and ensures consistency across all HTTP route registration.

**Example**: When adding a new OAuth2 handler, the implementer must:
1. Create the handler in `internal/adapters/http/handlers/**/**`
2. Add the handler field to `app.EnduserHandlers` or the `app.AdminHandlers` struct
3. Instantiate the handler in `app.Builder.Build()` method with all necessary dependencies
4. Pass the pre-wired handler to `SetupEnduserRoutes()` or `SetupAdminRoutes()` as part of the handlers struct
5. Register routes in `SetupEnduserRoutes()` or `SetupAdminRoutes()` using the passed handler, WITHOUT creating new instances

## Development Requirements

### Compliance Checklist

**PRECONDITIONS (must complete BEFORE implementation begins)**:

- [ ] Domain model designed: all entities, aggregates, and value objects identified and documented
- [ ] Domain concepts added to [ARCHITECTURE.md](ARCHITECTURE.md) Glossary section
- [ ] Configuration requirements designed: example YAML snippets showing all new config options
- [ ] Configuration examples committed to [examples/config/](examples/config/) for reference
- [ ] APIs designed and documented in OpenAPI format (confirm with user/stakeholder per Principle X)
- [ ] Database schema designed (migration files and SQL documented, or confirm no DB changes)
- [ ] Frontend components designed: review [web/src/design-system/docs/INDEX.md](../web/src/design-system/docs/INDEX.md) and ensure design system can be used
- [ ] Universal components identified: plan to add them to design system in `web/src/design-system/components/` (if applicable)

**Implementation Phase**:

- [ ] Security controls are enabled by default and fail closed
- [ ] [ARCHITECTURE.md](ARCHITECTURE.md) reflects architectural changes (if any)
- [ ] Major decisions recorded in [adrs/](adrs/) with correct numbering
- [ ] Code follows patterns established in accepted ADRs (especially ADR 004 for persistence)
- [ ] APIs implemented exactly as documented in OpenAPI specification
- [ ] End-user API documentation in [docs/api/](docs/api/) with examples (if applicable)
- [ ] Configuration implementation uses unified system port, not custom loading
- [ ] Domain logic uses ports (interfaces) and adapters are separated
- [ ] No custom cryptography; security features use vetted libraries
- [ ] Structured logging present for security-critical operations
- [ ] Automated tests included (unit, integration, or both) with meaningful coverage
- [ ] No Bash scripts used for code correctness validation (only infrastructure tasks)
- [ ] New persistence entities follow quickstart.md patterns (if applicable)
- [ ] All database schema changes in `/migrations/` with go-migrate naming conventions
- [ ] Migrations tested: verified apply cleanly, rollback works, can repeat without data loss
- [ ] All PostgreSQL-backed repositories tested in integration tests
- [ ] Frontend components use design system primitives and semantic tokens (if applicable)
- [ ] Universal components added to design system with Storybook stories (if applicable)
- [ ] No custom CSS bypassing design tokens or introducing inconsistent styling
- [ ] WCAG 2.1 AA accessibility verified: 4.5:1 text contrast, 3:1 UI component contrast (if applicable)
- [ ] All service instantiation uses Builder pattern in `internal/app/builder.go` (Principle XII)
- [ ] HTTP routes registered via `SetupAdminRoutes()` and `SetupEnduserRoutes()` functions (Principle XII)
- [ ] Routing functions receive pre-wired dependencies, do NOT instantiate services (Principle XII)
- [ ] All handlers and services depend on ports (interfaces), not concrete implementations (Principle XII)

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

If Principle IV (API Documentation & OpenAPI Transparency) cannot be satisfied:

1. STOP implementation immediately
2. Document why API cannot be documented in OpenAPI or why user confirmation cannot be obtained
3. Escalate to project maintainers for guidance
4. Do NOT implement undocumented or unconfirmed APIs

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

If Principle IX (Persistence Pattern Consistency & Database Migration Management) cannot be satisfied:

1. STOP implementation immediately
2. Document why quickstart.md patterns or migration management cannot be followed
3. Propose alternative approach with technical justification
4. Create an ADR documenting the exception and rationale
5. Do NOT implement non-standard persistence patterns or schema changes without an accepted ADR

If Principle X (API-First Development) requires deviation:

1. STOP implementation immediately
2. Document why API cannot be designed before implementation or why confirmation cannot be obtained
3. Escalate to project maintainers for guidance
4. Do NOT implement APIs without design confirmation

If Principle XI (Design System Compliance & Consistency) cannot be satisfied:

1. STOP implementation immediately
2. Document why design system components cannot be used or why universal patterns cannot be contributed
3. Propose alternative approach with technical justification
4. Escalate to project maintainers for exception approval
5. Do NOT implement frontend components that bypass design system without explicit approval

If Principle XII (Dependency Injection & Component Wiring) cannot be satisfied:

1. STOP implementation immediately
2. Document why Builder pattern cannot be used or why routing functions cannot be thin adapters
3. Propose alternative DI approach with technical justification (e.g., service locator, factory pattern)
4. Create an ADR documenting the exception and alternative pattern
5. Do NOT implement service instantiation in routing functions or bypass Builder pattern without an accepted ADR

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
- Reviewers MUST verify database migrations use go-migrate naming and are tested (Principle IX)
- Reviewers MUST verify API changes have user confirmation (Principle X)
- Reviewers MUST verify frontend components use design system and universal patterns are contributed (Principle XI)
- Reviewers MUST verify WCAG 2.1 AA accessibility compliance for frontend components (Principle XI)
- Reviewers MUST verify dependency injection uses Builder pattern and routing functions are thin (Principle XII)
- Template files in [.specify/templates/](.specify/templates/) provide execution workflows that enforce these principles

### Task List Requirements

Every feature's `tasks.md` file MUST include these mandatory sections from [tasks-template.md](.specify/templates/tasks-template.md):

**MANDATORY SECTIONS** (cannot be omitted):
1. **Phase 2: Design Preconditions** - Constitution PRECONDITIONS implementation
   - Phase 2a: Domain Model & Glossary (Principles II, V)
   - Phase 2b: Configuration Design (Principle VII)
   - Phase 2c: API Design (Principles IV, X)
   - Phase 2d: Database Design (Principle IX)
   - Phase 2e: Frontend/Design System Review (Principle XI, if applicable)

2. **Phase N: Constitution Compliance Verification** - Constitution Implementation Phase checklist
   - Design Phase Verification tasks
   - Implementation Phase Verification tasks grouped by principle
   - All verification tasks with explicit principle references

**CUSTOMIZABLE SECTIONS** (adapt to feature):
- Phase 1: Setup (project-specific initialization)
- Phase 2.5: Foundational Infrastructure (feature-specific foundation)
- Phase 3+: User Stories (based on spec.md user stories)
- Additional Polish (optional improvements)

**Enforcement**:
- AI agents generating tasks.md MUST retain Phase 2 and Phase N sections verbatim (adapting only task descriptions to the specific feature)
- The tasks-template.md uses 🔒 emoji and [MANDATORY] markers to clearly distinguish mandatory from customizable sections
- Omitting mandatory sections violates this constitution and blocks feature completion

**Version**: 1.6.0 | **Ratified**: 2025-12-14 | **Last Amended**: 2026-01-05
