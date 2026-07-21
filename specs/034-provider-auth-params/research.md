# Research: Provider Authorization Parameters

## Decision: Extend the existing provider entity with a string map

**Rationale**: `ThirdpartyOAuth2ProviderEntity` already owns all per-provider OAuth2 configuration and is used by the admin handler, both storage adapters, and `OAuth2SessionService`. Adding `AuthorizationParams map[string]string` keeps this static provider setting in the existing aggregate and requires no new port, entity ID, service, or dependency.

**Alternatives considered**:

- A new provider-parameter entity and repository: rejected; the map has no independent lifecycle or identity.
- Passing parameters through the end-user authorization request: rejected; it would trust unvalidated user input and violate the feature's security requirement.
- A global configuration section: rejected; values belong to an individual provider service.

## Decision: Validate in the domain entity before any persistence or encryption side effect

**Rationale**: `ValidateForCreate` and `ValidateForUpdate` are the existing authoritative pure validation gates. A small shared validator will reject whitespace-only names or values and reserved names before branch-key creation, encryption, or repository writes. Reserved names are compared case-insensitively as the conservative interpretation of broker-owned OAuth2 fields.

**Alternatives considered**:

- HTTP-handler-only validation: rejected; other callers could persist unsafe configuration.
- Filtering reserved names silently: rejected; administrators need an actionable error and silent omission obscures configuration mistakes.
- Detecting duplicate JSON object keys: rejected; the standard JSON decoder follows JSON-object semantics, where a map represents one final value per key. No separate conflicting pair can exist in the decoded request.

## Decision: Preserve the existing map only when the update request omits the field

**Rationale**: A Go map decoded from an omitted JSON field is `nil`, while an explicitly supplied empty object is an empty non-nil map. The handler can preserve the stored map only for `nil`; `{}` intentionally clears it. This matches the feature's optional-field update requirement without a new patch API or wrapper type.

**Alternatives considered**:

- Treat omitted and empty identically: rejected; it prevents an administrator from clearing configuration.
- Change the existing full-update API to a patch endpoint: rejected; it is unnecessary scope expansion.

## Decision: Store the map in the existing service row as JSONB

**Rationale**: The map is small, structured configuration with string keys and values. A `JSONB NOT NULL DEFAULT '{}'::jsonb` column preserves a service-local map, gives old rows empty behavior automatically, and keeps create/read/list/update within the existing repository record.

**Alternatives considered**:

- A normalized child table: rejected; it adds a repository and lifecycle for static key/value configuration that is always read and written with its parent.
- Serializing JSON into text: rejected; JSONB is already used for structured service data and provides native validation.

## Decision: Add parameters only to the generated upstream URL

**Rationale**: `OAuth2SessionService.InitiateOAuth2Flow` is the only path that builds the upstream authorization URL. It fetches the persisted provider through `ThirdpartyOAuth2ProviderService`, creates state and PKCE values, and redirects the handler to the result. Add the validated stored entries there, after broker-owned fields are created, without reading end-user query values beyond the existing broker `redirect_uri` input.

**Alternatives considered**:

- Add them in the HTTP handler: rejected; authorization URL construction belongs in the domain session service and direct handler forwarding risks query propagation.
- Modify the upstream URL from arbitrary request query parameters: rejected; this would make the browser a configuration authority.
- Add a provider-specific Zalando branch: rejected; a generic validated map already covers this and avoids provider-specific code.

## Decision: Reuse existing test layers and documentation surfaces

**Rationale**: Unit tests already cover provider validation, admin handler mapping, and session authorization URLs. The repository has memory and PostgreSQL test patterns, and E2E tests use production app bootstrap. The admin OpenAPI contract and `docs/guides/manage-agents-and-services.md` are the existing public service-management documentation.

**Alternatives considered**:

- New testing framework or helper abstraction: rejected; existing testify and Ginkgo/Gomega coverage is sufficient.
- Frontend work: rejected; this feature changes backend administration and upstream redirects only.
