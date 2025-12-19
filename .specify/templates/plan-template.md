# Implementation Plan: [FEATURE]

**Branch**: `[###-feature-name]` | **Date**: [DATE] | **Spec**: [link]
**Input**: Feature specification from `/specs/[###-feature-name]/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

[Extract from feature spec: primary requirement + technical approach from research]

## Technical Context

<!--
  ACTION REQUIRED: Replace the content in this section with the technical details
  for the project. The structure here is presented in advisory capacity to guide
  the iteration process.
-->

**Language/Version**: [e.g., Python 3.11, Swift 5.9, Rust 1.75 or NEEDS CLARIFICATION]  
**Primary Dependencies**: [e.g., FastAPI, UIKit, LLVM or NEEDS CLARIFICATION]  
**Storage**: [if applicable, e.g., PostgreSQL, CoreData, files or N/A]  
**Testing**: [e.g., pytest, XCTest, cargo test or NEEDS CLARIFICATION]  
**Target Platform**: [e.g., Linux server, iOS 15+, WASM or NEEDS CLARIFICATION]
**Project Type**: [single/web/mobile - determines source structure]  
**Performance Goals**: [domain-specific, e.g., 1000 req/s, 10k lines/sec, 60 fps or NEEDS CLARIFICATION]  
**Constraints**: [domain-specific, e.g., <200ms p95, <100MB memory, offline-capable or NEEDS CLARIFICATION]  
**Scale/Scope**: [domain-specific, e.g., 10k users, 1M LOC, 50 screens or NEEDS CLARIFICATION]

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Before proceeding, verify compliance with [.specify/memory/constitution.md](.specify/memory/constitution.md):

**Design Preconditions (BLOCKING)**:

- [ ] **Domain Model**: Have entities, aggregates, value objects been identified and documented?
- [ ] **Domain Concepts**: Will new domain terms be added to ARCHITECTURE.md Glossary?
- [ ] **Configuration Design**: Have all config requirements been identified with YAML examples?
- [ ] **Config Examples**: Will example YAML snippets be added to examples/config/?
- [ ] **API Design First**: Will APIs be designed (OpenAPI spec) and confirmed BEFORE implementation?
- [ ] **API Documentation**: Will OpenAPI specs be created in `/api/enduser/` or `/api/admin/` as applicable?
- [ ] **API Changes**: Are all API changes confirmed by user/stakeholder (document in PR)?
- [ ] **Database Design**: Will all schema changes use go-migrate naming in `/migrations/`?

**Implementation Considerations**:

- [ ] **Security-First**: Are security features enabled by default? No bypasses or optional security?
- [ ] **Architecture Docs**: Will ARCHITECTURE.md be updated if this touches architecture?
- [ ] **ADRs**: Does this require an ADR in adrs/ for major decisions?
- [ ] **Library-First Security**: Are we using vetted libraries for crypto/security (no custom implementations)?
- [ ] **Zalando Guidelines**: Will APIs follow Zalando RESTful API and Event Guidelines?
- [ ] **End-User Docs**: Will API documentation be rendered in `docs/api/` with examples?
- [ ] **Migration Testing**: Will migrations be tested (apply/rollback) in PostgreSQL integration tests?
- [ ] **Hexagonal Architecture**: Does domain logic use ports (interfaces) with clear adapter separation?
- [ ] **Persistence Patterns**: If adding persistence, will it follow specs/004-persistence-layer/quickstart.md?

*If any BLOCKING check fails, stop and clarify requirements. Implementation cannot begin until all preconditions complete.*

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: [Document the selected structure and reference the real
directories captured above]

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
