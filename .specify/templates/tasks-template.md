---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: The examples below include test tasks. Tests are OPTIONAL - only include them if explicitly requested in the feature specification.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `src/`, `tests/` at repository root
- **Web app**: `backend/src/`, `frontend/src/`
- **Mobile**: `api/src/`, `ios/src/` or `android/src/`
- Paths shown below assume single project - adjust based on plan.md structure

<!-- 
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.
  
  The /speckit.tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Endpoints from contracts/
  
  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment
  
  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Create project structure per implementation plan
- [ ] T002 Initialize [language] project with [framework] dependencies
- [ ] T003 [P] Configure linting and formatting tools

---

## Phase 2: Design Preconditions (Blocking Prerequisites)

**Purpose**: Domain model, configuration, API, and database design MUST all be complete before implementation

**⚠️ CRITICAL**: No code implementation can begin until this entire phase is complete

### Phase 2a: Domain Model & Glossary

Examples of domain design tasks:

- [ ] T004 Identify all entities, aggregates, and value objects for this feature
- [ ] T004a Document domain model (entities, aggregates, value objects, domain events)
- [ ] T004b [P] Add all new domain terms to ARCHITECTURE.md Glossary section
- [ ] T004c Document invariants and consistency rules for key aggregates

**Checkpoint**: Domain model complete and documented

### Phase 2b: Configuration Design

Examples of configuration design tasks:

- [ ] T005 Identify all configuration parameters needed for this feature
- [ ] T005a Create example YAML showing all new config options with defaults
- [ ] T005b [P] Commit configuration examples to `examples/config/[feature_name].yaml`
- [ ] T005c Update `examples/config/README.md` to reference new configuration section

**Checkpoint**: Configuration requirements designed with YAML examples

### Phase 2c: API Design

Examples of API design tasks:

- [ ] T006 Design end-user API and document in `/api/enduser/openapi.yaml` (per Constitution Principle IV)
- [ ] T006a [P] Get user/stakeholder confirmation for end-user API design
- [ ] T007 [P] Design admin API in `/api/admin/openapi.yaml` if applicable (per Constitution Principle IV)
- [ ] T007a [P] Get user/stakeholder confirmation for admin API design (if applicable)

**Checkpoint**: APIs designed and confirmed by user/stakeholder

### Phase 2d: Database Design

Examples of database design tasks:

- [ ] T008 Design database schema and document migrations to create (or confirm no DB changes needed)
- [ ] T008a Create migration files: `[NNN]_[description].up.sql` and `.down.sql` (per go-migrate format)
- [ ] T008b [P] Document all schema changes (tables, indexes, constraints)

**Checkpoint**: Database schema designed, migrations documented

---

## Phase 2.5: Foundational Infrastructure

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on your project):

- [ ] T009 [P] Implement authentication/authorization framework
- [ ] T010 [P] Setup API routing and middleware structure
- [ ] T011 Create base models/entities that all stories depend on
- [ ] T012 Configure error handling and logging infrastructure
- [ ] T013 Setup environment configuration management

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1 (OPTIONAL - only if tests requested) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T014 [P] [US1] Contract test for [endpoint] in tests/contract/test_[name].go
- [ ] T015 [P] [US1] Integration test for [user journey] in tests/integration/test_[name].go

### Implementation for User Story 1

- [ ] T016 [P] [US1] Create [Entity1] model in internal/domain/[entity1].go
- [ ] T017 [P] [US1] Create [Entity2] model in internal/domain/[entity2].go
- [ ] T018 [US1] Implement [Service] in internal/domain/services/[service].go (depends on T016, T017)
- [ ] T019 [US1] Implement [endpoint/feature] in internal/adapters/[location]/[file].go
- [ ] T020 [US1] Add validation and error handling per Constitution Principle I (security-first)
- [ ] T021 [US1] Add structured logging for user story 1 operations

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2 (OPTIONAL - only if tests requested) ⚠️

- [ ] T022 [P] [US2] Contract test for [endpoint] in tests/contract/test_[name].go
- [ ] T023 [P] [US2] Integration test for [user journey] in tests/integration/test_[name].go

### Implementation for User Story 2

- [ ] T024 [P] [US2] Create [Entity] model in internal/domain/[entity].go
- [ ] T025 [US2] Implement [Service] in internal/domain/services/[service].go
- [ ] T026 [US2] Implement [endpoint/feature] in internal/adapters/[location]/[file].go
- [ ] T027 [US2] Integrate with User Story 1 components (if needed)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3 (OPTIONAL - only if tests requested) ⚠️

- [ ] T028 [P] [US3] Contract test for [endpoint] in tests/contract/test_[name].go
- [ ] T029 [P] [US3] Integration test for [user journey] in tests/integration/test_[name].go

### Implementation for User Story 3

- [ ] T030 [P] [US3] Create [Entity] model in internal/domain/[entity].go
- [ ] T031 [US3] Implement [Service] in internal/domain/services/[service].go
- [ ] T032 [US3] Implement [endpoint/feature] in internal/adapters/[location]/[file].go

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

### Constitution Compliance (MANDATORY)

**Design Phase**:
- [ ] TXXX Verify domain model design is documented in ARCHITECTURE.md Glossary
- [ ] TXXX Verify configuration design YAML examples exist in `examples/config/`
- [ ] TXXX Verify configuration examples referenced in `examples/config/README.md`
- [ ] TXXX Verify API designs documented in `/api/enduser/openapi.yaml` or `/api/admin/openapi.yaml`
- [ ] TXXX Verify user/stakeholder confirmed API designs (document reference in PR)
- [ ] TXXX Verify database schema design documented (migrations or confirmation of no DB changes)

**Implementation Phase**:
- [ ] TXXX [P] Update `/api/enduser/openapi.yaml` or `/api/admin/openapi.yaml` with implemented APIs
- [ ] TXXX [P] Verify API implementation matches confirmed OpenAPI specification exactly
- [ ] TXXX Update `docs/api/` with end-user API documentation and integration examples
- [ ] TXXX [P] Verify configuration uses unified system configuration port (not custom loading)
- [ ] TXXX [P] Create/verify database migrations in `/migrations/` follow sequential numbering (NNN format)
- [ ] TXXX [P] Write/update integration tests verifying migrations (apply, rollback, data integrity)
- [ ] TXXX Update ARCHITECTURE.md with architectural changes and new domain concepts
- [ ] TXXX Update ARCHITECTURE.md Glossary with new domain terms (if not done in design phase)
- [ ] TXXX [P] Create/update ADR in adrs/ for major architectural decisions
- [ ] TXXX Verify security features are enabled by default (no optional bypasses)
- [ ] TXXX Verify domain logic uses ports (interfaces) and adapters
- [ ] TXXX [P] Add structured logging for security-critical operations
- [ ] TXXX [P] Verify all PostgreSQL-backed repositories tested in integration tests

### Additional Polish

- [ ] TXXX Code cleanup and refactoring
- [ ] TXXX Performance optimization across all stories
- [ ] TXXX [P] Additional unit tests (if requested) in tests/unit/
- [ ] TXXX Run quickstart.md validation (if exists)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Design Preconditions (Phase 2)**: Depends on Setup completion - BLOCKS all implementation
  - **CRITICAL**: Domain model must be designed and documented
  - **CRITICAL**: Configuration requirements must be designed with YAML examples
  - **CRITICAL**: APIs must be designed and confirmed by user/stakeholder BEFORE implementation
  - **CRITICAL**: Database schema must be designed and migrations planned (or confirmed no DB changes)
  - Phase 2a, 2b, 2c, 2d can proceed in parallel, but all must complete before Phase 2.5 begins
- **Foundational Infrastructure (Phase 2.5)**: Depends on ALL of Phase 2 completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Phase 2 + Phase 2.5 completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Phase 2 + Phase 2.5 completion - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Phase 2 + Phase 2.5 completion - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Phase 2 + Phase 2.5 completion - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Tests (if included) MUST be written and FAIL before implementation
- Models before services
- Services before endpoints
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Models within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all tests for User Story 1 together (if tests requested):
Task: "Contract test for [endpoint] in tests/contract/test_[name].py"
Task: "Integration test for [user journey] in tests/integration/test_[name].py"

# Launch all models for User Story 1 together:
Task: "Create [Entity1] model in src/models/[entity1].py"
Task: "Create [Entity2] model in src/models/[entity2].py"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Design Preconditions (CRITICAL - blocks all stories)
   - 2a: Domain model & glossary
   - 2b: Configuration design with YAML examples
   - 2c: API design + user/stakeholder confirmation
   - 2d: Database schema design
3. Complete Phase 2.5: Foundational Infrastructure (CRITICAL - blocks all stories)
4. Complete Phase 3: User Story 1
5. **STOP and VALIDATE**: Test User Story 1 independently
6. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup → Design Preconditions (all 4 areas) → Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test independently → Deploy/Demo
4. Add User Story 3 → Test independently → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers or agents:

1. Parallel design work (Phase 2) - use specialized sub-agents:
   - **Phase 2a (Domain Model)**: Use architecture-reviewer or domain-architect specialist
   - **Phase 2b (Configuration Design)**: Use backend-developer or platform-engineer specialist
   - **Phase 2c (API Design)**: Use api-designer specialist (coordinates with user for confirmation)
   - **Phase 2d (Database Schema)**: Use database-administrator specialist
2. Remaining team/agents complete Setup (Phase 1) while design in progress
3. Once ALL of Phase 2 complete and confirmed (especially API sign-off):
   - All team/agents complete Foundational Infrastructure (Phase 2.5) together
4. Once Phase 2.5 done, assign user stories to specialized agents:
   - **User Story 1**: golang-pro or feature-lead or frontend-developer
   - **User Story 2**: another agent like in User Story 1
   - **User Story 3**: another agent like in User Story 1
5. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
