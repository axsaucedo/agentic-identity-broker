# Research: Hybrid OAuth Server Modes

**Feature**: 030-hybrid-oauth-modes | **Date**: 2026-05-12

## R1: Mode Strategy Pattern — Interface vs Composite

**Decision**: Use a `ModeStrategy` interface with three implementations (proxy, local, hybrid) where hybrid delegates to proxy or local based on agent classification.

**Rationale**: The spec requires mode to drive wiring at startup (FR-014). A strategy interface aligns with the existing `AuthorizationProceedStrategy` and `TokenGrantStrategy` patterns (ADR 014). The hybrid strategy composes proxy and local strategies rather than being a third independent implementation — this avoids duplicating logic.

**Alternatives considered**:
- Single strategy with mode field + runtime branching — rejected per FR-014 (no runtime `if mode ==` checks in handlers)
- Separate handler sets per mode wired by builder — rejected because hybrid needs both sets coexisting, and agent classification determines which path to take per-request

**Implementation approach**: The `ModeStrategy` acts at a higher level than `ProceedStrategy`/`TokenGrantStrategy`. It sits in the resolution/classification layer and selects which proceed/grant strategy to invoke. In hybrid mode, the builder wires both strategy sets, and the mode strategy dispatches based on agent class.

## R2: Agent Classification Location

**Decision**: Classification is a method on the Agent entity: `agent.Class() AgentClass`. The `AgentClass` type is defined in `internal/domain/storage/` alongside the Agent entity.

**Rationale**: Classification derives entirely from Agent's own fields (`ClientID`, `ClientURIs`). It's natural to ask an agent "what class are you?" rather than passing the agent to an external classifier. The Agent entity already lives in domain code (`internal/domain/storage/`), so this keeps the logic co-located with the data it inspects.

**Alternatives considered**:
- Free function `ClassifyAgent(agent)` in `internal/domain/oauth2/` — rejected because it separates logic from the data it operates on; classification is intrinsic to the entity
- Classification in HTTP handler — rejected because it's domain logic, not HTTP concern

## R3: Config Restructuring Approach

**Decision**: Restructure `OAuth2AuthServerConfig` to have nested `Proxy` and `Local` sections. Mode validation ensures only relevant sections are populated. The flat fields (`UpstreamIssuerURI`, `TokenTTL`, etc.) move into their respective subsections.

**Rationale**: The spec requires each mode to have a dedicated config section (FR-015). Hybrid mode requires both sections (FR-016). Nested sections make cross-mode field ambiguity impossible at the structural level.

**Alternatives considered**:
- Keep flat structure with mode-based validation only — rejected because FR-015 explicitly requires dedicated sections per mode
- Separate config types per mode (union type) — adds complexity without benefit since Viper maps to a single struct

**Migration path**: Existing configs using flat fields (`upstream_issuer_uri` at top level) need to be updated. The `issue_token` → `local` rename happens in Phase 0 as a pure refactor.

## R4: Hybrid Mode Builder Wiring

**Decision**: In hybrid mode, the builder creates both `issueTokenProceedStrategy` (renamed `localProceedStrategy`) and `proxyProceedStrategy`, both `localTokenGrantStrategy` and `proxyTokenGrantStrategy`. A dispatching strategy wraps both and selects based on agent classification at request time.

**Rationale**: The builder already creates one set or the other based on mode (`internal/app/builder.go:609-653`). For hybrid, both sets are created. The dispatch happens in a thin wrapper that classifies the agent and delegates — this is not a runtime mode check (it's agent-class dispatch), satisfying FR-014.

**Alternatives considered**:
- Two separate routers/mux for proxy and local paths — overengineered, requires URL path splitting that doesn't exist in the spec
- Middleware-based classification — moves domain logic into HTTP middleware, violating hexagonal architecture

## R5: CIMD Agent UUID Rejection (FR-005)

**Decision**: Enforce in the client resolver: after resolving an agent by UUID, check if `len(agent.ClientURIs) > 0`. If so, reject with error "CIMD agents must be addressed via their URL-format client_id".

**Rationale**: FR-005 is explicit: UUID-format `client_id` resolving to a CIMD agent must be rejected. This prevents bypassing CIMD validation. The check belongs in the resolver since that's where resolution format and agent properties are both available.

**Alternatives considered**:
- Check in mode strategy — too late, classification should not run on improperly addressed agents
- Check in handler — domain logic belongs in domain layer

## R6: Deprecation Handling for `issue_token`

**Decision**: Config validation returns a clear error: `"mode 'issue_token' has been renamed to 'local' — please update your configuration"`. No silent fallback (SR-003).

**Rationale**: Explicit error with remediation guidance. The spec is clear: no silent migration.

## R7: ADR Decision

**Decision**: No new ADR. The spec, plan, and research documents already capture all architectural decisions (mode strategy pattern, agent classification, config restructuring, mode renaming). ADR 014 remains the binding ADR for local token issuance via fosite; this feature extends its patterns without changing the architectural foundation.

**Rationale**: An ADR would duplicate content already present in the spec artifacts. The spec + plan serve the same purpose — documenting decisions, rationale, and alternatives considered.
