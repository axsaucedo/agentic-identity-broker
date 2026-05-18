# Architecture Overview
This document serves as a critical, living template designed to equip agents with a rapid and comprehensive understanding of the codebase's architecture, enabling efficient navigation and effective contribution from day one. Update this document as the codebase evolves.

## 1. Project Structure
This section provides a high-level overview of the project's directory and file structure, categorised by architectural layer or major functional area. It is essential for quickly navigating the codebase, locating relevant files, and understanding the overall organization and separation of concerns.


[Project Root]/

├── cmd/                  # Main source code for backend services
├── internal/             # Code that is internal
├── pkg/                  # Code that is ok to be used when this is 
├── config/               # Backend configuration files
├── test/                 # Backend unit and integration tests
├── build/Dockerfile      # Dockerfile for backend deployment
├── web/                  # Single Page Application (Consent Frontend)
│   ├── src/              # Main source code for React application
│   │   ├── components/   # Reusable UI components
│   │   │   ├── consent/  # Consent-specific components (DelegationCard, ServiceCard, etc.)
│   │   │   ├── layout/   # Layout components (AppLayout)
│   │   │   └── ui/       # Generic UI components (Button, Switch, DatePicker, etc.)
│   │   ├── pages/        # Application pages/views (ConsentOverviewPage, AgentGrantDetailPage)
│   │   ├── hooks/        # Custom React hooks (useConsent, useAgentGrants, useToggleGrant)
│   │   ├── services/     # Frontend services
│   │   │   ├── api/      # API client and service layer (axios-based)
│   │   │   └── storage/  # Client-side storage utilities
│   │   ├── types/        # TypeScript type definitions
│   │   ├── utils/        # Utility functions (validation, formatting)
│   │   ├── assets/       # Images, fonts, and other static assets
│   │   └── styles/       # Global styles (Tailwind CSS)
│   ├── public/           # Publicly accessible assets (favicon, etc.)
│   ├── dist/consent/     # Build output directory (served by Go backend)
│   ├── tests/            # Frontend unit and integration tests (Vitest)
│   ├── package.json      # Frontend dependencies and scripts
│   ├── vite.config.ts    # Vite build configuration
│   ├── tsconfig.json     # TypeScript solution references
│   ├── tsconfig.app.json # App TypeScript configuration
│   ├── tsconfig.test.json # Test TypeScript configuration
│   ├── tsconfig.build.json # Build TypeScript configuration
│   └── tailwind.config.ts # Tailwind CSS v4.0 configuration
├── docs/                 # Project documentation (e.g., API docs, setup guides)
├── infra/                # Infrastructure as Code
│   └── cdk/              # AWS CDK (Go) – encryption infrastructure (KMS, DynamoDB, IAM)
├── scripts/              # Automation scripts (e.g., deployment, data seeding)
├── .github/              # GitHub Actions or other CI/CD configurations
├── .gitignore            # Specifies intentionally untracked files to ignore
├── README.md             # Project overview and quick start guide
└── ARCHITECTURE.md       # This document



## 2. High-Level System Diagram
Provide a simple block diagram (e.g., a C4 Model Level 1: System Context diagram, or a basic component diagram) or a clear text-based description of the major components and their interactions. Focus on how data flows, services communicate, and key architectural boundaries.
 
[User] <--> [Frontend Application] <--> [Backend Service 1] <--> [Database 1]
                                    |
                                    +--> [Backend Service 2] <--> [External API]                           

## 3. Core Components
(List and briefly describe the main components of the system. For each, include its primary responsibility and key technologies used.)

### 3.1. Identity Broker Service

Name: Agentic Identity Broker

Description: Core service providing secure identity management, authentication, and authorization for AI agents and autonomous systems. Implements hexagonal architecture with clear separation of domain logic, ports, and adapters.

Technologies: Go 1.21+, Viper (configuration), Cobra (CLI)

Deployment: Containerized service (Docker), deployable to Kubernetes, AWS ECS, or standalone

#### 3.1.1. Configuration Subsystem

**Purpose**: Flexible multi-source configuration management with environment-specific support, security-first design, and clear precedence rules.

**Architecture**: Hexagonal (ports and adapters pattern)

**Components**:
- **Port** (internal/ports/config.go): ConfigPort interface defining domain boundary
- **Adapter** (internal/config/loader.go): Viper-based implementation loading from multiple sources
- **Domain Types** (internal/domain/config/): LogLevel, LogFormat enums with validation
- **Domain Errors** (internal/domain/config/errors.go): ConfigError with error wrapping support

**Configuration Sources** (in precedence order, lowest to highest):
1. **Defaults**: Built-in default values (log.level=info, log.format=text)
2. **.env Files**: Environment-specific files (.env → .env.local → .env.{GO_ENV} → .env.{GO_ENV}.local)
3. **YAML File**: config.yaml with ${VAR} environment variable substitution
4. **CLI Flags**: Command-line flags (--log-level, --log-format, --config)

**Security Features**:
- Sensitive value redaction (IDENTITY_BROKER_* prefix and keywords: password, secret, token, key)
- Command injection prevention (rejects $(cmd), backticks, shell metacharacters)
- Circular reference detection (max depth: 10)
- Fail-closed on errors (graceful termination with clear messages)
- File permission validation
- Structured audit logging (JSON to stdout)

**Flow**:
```
Application Startup
  → Load Defaults
  → Load .env Files (godotenv)
  → Load YAML (Viper)
  → Expand ${VAR} References (with security validation)
  → Bind CLI Flags (Cobra)
  → Unmarshal to Config struct
  → Validate (custom validators)
  → Emit Audit Log
  → Display Startup Summary
  → Return Config to Application
```

**Performance**: Configuration loading completes in <250ms (within 5s startup budget)

**Technologies**:
- Viper v1.19.0+ (unified configuration management)
- Cobra v1.8.1+ (CLI framework)
- godotenv v1.5.1+ (.env file support)
- Custom validators (fast, zero-allocation validation)

**Future Extensions**: Hot-reloading (Reload method defined but not implemented), additional config categories (server, database, auth)

#### 3.1.2. Single Page Application (Consent Frontend)

**Purpose**: User-facing web interface for managing OAuth2 consent delegations to AI agents.

**Architecture**: React 18 Single Page Application with TypeScript, served from Go backend

**Technology Stack**:
- **Frontend Framework**: React 18.2+ with TypeScript 5.3+
- **Build Tool**: Vite 5.0+ (fast ESM-based bundler)
- **Styling**: Tailwind CSS v4.0 (utility-first CSS framework)
- **UI Components**: Headless UI 1.7+ (accessible, unstyled components)
- **HTTP Client**: Axios 1.6+ (promise-based HTTP client)
- **Router**: React Router DOM 6.20+ (client-side routing)
- **Testing**: Vitest 1.0+ (fast unit test framework)
- **State Management**: React hooks + Context API (no external state library)

**Directory Structure**:
```
web/
├── src/
│   ├── components/       # React components
│   │   ├── consent/      # Consent-specific components
│   │   │   ├── DelegationCard.tsx        # Agent delegation card
│   │   │   ├── DelegationList.tsx        # List of delegations
│   │   │   ├── ServiceCard.tsx           # OAuth2 service card
│   │   │   ├── ServiceGrantList.tsx      # List of service grants
│   │   │   ├── ScopeList.tsx             # Scope selection UI
│   │   │   ├── GrantStatusBadge.tsx      # Grant status indicator
│   │   │   └── GrantValidityControl.tsx  # Expiration date control
│   │   ├── layout/       # Layout components
│   │   │   └── AppLayout.tsx             # Main app layout
│   │   └── ui/           # Reusable UI components
│   │       ├── Button.tsx                # Button component
│   │       ├── Switch.tsx                # Toggle switch
│   │       ├── DatePicker.tsx            # Date picker
│   │       ├── ErrorBoundary.tsx         # Error boundary
│   │       ├── InlineError.tsx           # Error display
│   │       ├── EmptyState.tsx            # Empty state UI
│   │       └── Skeleton.tsx              # Loading skeleton
│   ├── pages/            # Application pages
│   │   ├── ConsentOverviewPage.tsx       # List of all agent delegations
│   │   ├── AgentGrantDetailPage.tsx      # Agent-specific grant management
│   │   └── ErrorPage.tsx                 # Error page
│   ├── hooks/            # Custom React hooks
│   │   ├── useConsent.ts                 # Fetch agent delegations
│   │   ├── useAgentGrants.ts             # Fetch agent grants
│   │   ├── useToggleGrant.ts             # Toggle grant scopes
│   │   ├── useUpdateValidity.ts          # Update grant expiration
│   │   └── useRetry.ts                   # Retry with exponential backoff
│   ├── services/         # Service layer
│   │   ├── api/          # API clients
│   │   │   ├── client.ts                 # Axios client configuration
│   │   │   ├── consent.ts                # Consent API methods
│   │   │   └── index.ts                  # API exports
│   │   └── storage/      # Client-side storage
│   │       └── session.ts                # Session storage utilities
│   ├── types/            # TypeScript types
│   │   ├── consent.ts                    # Consent domain types
│   │   └── index.ts                      # Type exports
│   ├── utils/            # Utility functions
│   │   └── validation.ts                 # Input validation
│   ├── App.tsx           # Root component
│   └── main.tsx          # Application entry point
├── dist/consent/         # Build output (served by Go)
├── vite.config.ts        # Vite configuration
├── tsconfig.json         # TypeScript solution references
├── tsconfig.app.json     # App TypeScript configuration
├── tsconfig.test.json    # Test TypeScript configuration
├── tsconfig.build.json   # Build TypeScript configuration
├── tailwind.config.ts    # Tailwind CSS configuration
└── package.json          # Dependencies and scripts
```

**Build Pipeline**:
1. **Development**: `npm run dev` runs Vite dev server (http://localhost:3000)
2. **Build**: `npm run build` compiles TypeScript and bundles with Vite
3. **Output**: Static files written to `dist/consent/` directory
4. **Deployment**: Go backend serves files from `dist/consent/` at `/consent` path

**SPA Serving Pattern**:
```
User Request: /consent/agents
  ↓
Go HTTP Server (Port 8080)
  ↓
Static File Handler (/consent/*)
  ↓ (404 fallback for client-side routes)
Serve index.html
  ↓
Browser loads React app
  ↓
React Router handles /agents route
  ↓
Component fetches data from /api/consent/agents
  ↓
Go API Handler returns JSON
```

**Key Features**:
- **Client-Side Routing**: React Router handles all `/consent/*` routes without page reloads
- **History API Fallback**: Go backend serves `index.html` for all `/consent/*` paths (SPA fallback)
- **API Integration**: Frontend makes requests to `/api/consent/*` endpoints on same domain
- **CSRF Protection**: All mutating requests include CSRF token from cookie
- **Session Management**: Principal extracted from `X-Principal` header (set by reverse proxy)
- **Type Safety**: Full TypeScript coverage with strict mode enabled
- **Responsive Design**: Tailwind CSS utilities for mobile-first responsive UI
- **Accessibility**: Headless UI components ensure WCAG 2.1 compliance
- **Error Handling**: Error boundaries and retry logic for resilient UX

**Component Hierarchy**:
```
App
├── ErrorBoundary
│   └── AppLayout
│       ├── ConsentOverviewPage
│       │   └── DelegationList
│       │       └── DelegationCard (per agent)
│       │           └── GrantStatusBadge
│       └── AgentGrantDetailPage
│           ├── ServiceGrantList
│           │   └── ServiceCard (per OAuth2 service)
│           │       ├── Switch (toggle grant)
│           │       └── ScopeList (scope checkboxes)
│           └── GrantValidityControl
│               └── DatePicker (expiration date)
```

**State Management**:
- **Local State**: React `useState` for component-level state
- **Server State**: Custom hooks with axios for API data fetching
- **Context**: React Context API for global UI state (theme, error messages)
- **No Redux/MobX**: Hooks + Context sufficient for current requirements

**API Communication**:
- **Base URL**: `/api` (relative, same origin)
- **Authentication**: Session-based (X-Principal header from proxy)
- **CSRF**: X-CSRF-Token header required for POST/PUT/DELETE
- **Error Handling**: Axios interceptors for global error handling
- **Retry Logic**: Exponential backoff for transient failures

**Testing Strategy**:
- **Unit Tests**: Vitest for component and hook testing
- **Integration Tests**: Test component + API interactions with mocked backend
- **E2E Tests**: (Future) Playwright for full user flows
- **Coverage Target**: >80% for critical paths

**Performance Optimizations**:
- **Code Splitting**: Lazy loading of routes with React.lazy()
- **Tree Shaking**: Vite removes unused code automatically
- **Minification**: Terser minification in production builds
- **Caching**: Immutable asset URLs with content hashing
- **Bundle Size**: Target <200KB gzipped for initial load

**Security Considerations**:
- **XSS Prevention**: React escapes all user input by default
- **CSRF Protection**: Token-based CSRF protection for mutating requests
- **Content Security Policy**: (Future) CSP headers from Go backend
- **Dependency Scanning**: Regular npm audit for vulnerabilities
- **TypeScript**: Compile-time type checking prevents runtime errors

#### 3.1.3. Builder Pattern and Dependency Injection

**Purpose**: Central DI wiring for all application components. All service instantiation happens in `internal/app/builder.go` via the `Builder` struct and `NewBuilder().With*().Build()` pattern.

**OTel Provider Wiring** (Feature 017):

The OTel provider is initialized in `Build()` and follows the app-layer provider pattern documented in [ADR 011](adrs/011-opentelemetry-provider-pattern.md):

```
Build()
  ├── NewProvider(ctx, cfg.Telemetry, logger)    // Initialize TracerProvider, MeterProvider, LoggerProvider
  ├── otel.SetTracerProvider(tp)                  // Register globally for child spans in adapters
  ├── otel.SetMeterProvider(mp)                   // Register globally (otelchi uses this for HTTP metrics)
  └── store shutdown func on App.ShutdownTelemetry
```

**Test Override Pattern** (mirrors `WithEncryption`):

```go
// WithTracerProvider sets a custom TracerProvider for testing.
// When set, this provider is registered as the global provider instead of
// the one created by NewProvider().
func (b *Builder) WithTracerProvider(tp *sdktrace.TracerProvider) *Builder
```

Tests use `bootstrap.NewInMemoryTracerProvider()` to get a `*tracetest.SpanRecorder`-backed provider and inject it via `builder.WithTracerProvider(tp)`.

**Graceful Shutdown Sequence**:

```
HTTP servers drain → tp.Shutdown(ctx) → mp.Shutdown(ctx) → lp.Shutdown(ctx) → process exit
```

The composite shutdown function is stored as `App.ShutdownTelemetry func(context.Context) error` and called after HTTP servers have drained all in-flight requests.

#### 3.1.4. End-to-End Testing Architecture

**Purpose**: Comprehensive E2E acceptance tests that validate the complete OAuth2 Authorization Server functionality through real HTTP requests and production code paths.

**Architecture**: Ginkgo BDD-style tests with separation of stable test scenarios (HTTP contract) and volatile setup code (production bootstrap wrappers).

**Key Design Principles** (see [ADR 007](adrs/007-e2e-testing-with-ginkgo.md)):

1. **Use Production Bootstrap**: Tests use production `app.Builder`, `httpAdapter.Server`, and `storage.Adapter` - no custom test implementations
2. **Stability Through Separation**: Test scenarios focus on HTTP contracts (stable), setup code wraps production bootstrap (volatile)
3. **Real HTTP Testing**: Tests use `httptest.Server` with full routing, middleware, and handler stack
4. **Test Isolation**: Each test gets fresh storage and server via `BeforeEach` - no shared state
5. **BDD Organization**: `Describe`/`Context`/`It` blocks map to acceptance scenarios (Given/When/Then)

**Test Structure**:
```
tests/e2e/
├── e2e_suite_test.go              # Ginkgo test suite entry point
├── bootstrap/                      # VOLATILE - Wraps production bootstrap
│   ├── server_factory.go          # Uses app.Builder from production
│   ├── test_server.go             # Wraps production HTTP server
│   └── storage.go                 # Uses production storage.Adapter
├── fixtures/                       # STABLE - Test data (domain model)
│   ├── agents.go, grants.go, principals.go, config.go
├── helpers/                        # STABLE - Test utilities
│   ├── http_helpers.go, mock_upstream.go
├── matchers/                       # STABLE - Custom Gomega matchers
│   └── oauth2_matchers.go
└── Test scenarios (62 scenarios total)
    ├── oauth2_authorize_test.go   # Authorization endpoint (23 scenarios)
    ├── oauth2_token_test.go       # Token endpoint (12 scenarios)
    ├── oauth2_metadata_test.go    # Metadata endpoint (8 scenarios)
    ├── oauth2_security_test.go    # Security tests (12 scenarios)
    └── oauth2_edge_cases_test.go  # Edge cases (7 scenarios)
```

**Running E2E Tests**:
```bash
# Run all E2E tests
just test-e2e

# Run with coverage report
just test-e2e-coverage

# Watch mode (auto-rerun on changes)
just test-e2e-watch

# Run specific scenarios
ginkgo -v --focus="Authorization Endpoint" ./tests/e2e/
```

**Key Benefits**:
- **Refactoring-Resistant**: Tests survive routing/DI changes because they use production bootstrap
- **Fast Execution**: In-memory storage, no database containers (< 60 seconds for full suite)
- **Clear Mapping**: Each `It()` block maps to one acceptance scenario in spec.md
- **Readable Output**: Ginkgo output is hierarchical and understandable by non-developers
- **Production Parity**: Tests exercise the same code path as production (DI, routing, middleware)

**Documentation**: Comprehensive E2E testing guide with examples, patterns, and anti-patterns in [tests/e2e/README.md](tests/e2e/README.md)

**Technologies**:
- Ginkgo v2 (BDD test framework)
- Gomega (assertion library)
- `net/http/httptest` (HTTP test server)
- Production `app.Builder` and `httpAdapter.Server` (no custom test implementations)

#### 3.1.4. Public API Documentation

**Purpose**: Comprehensive OpenAPI 3.0.3 documentation of all HTTP APIs exposed by the Identity Broker service.

**Architecture**: Dual-port HTTP server with clear separation between end-user and administrative APIs.

**API Specifications**:

- **[/api/enduser/openapi.yaml](/api/enduser/openapi.yaml)** - End-user server (Port 8000)
  - Canonical OpenAPI documentation for all end-user facing APIs
  - Endpoints: Health check, user info, consent management (agent delegations, service grants)
  - Authentication: Pre-authentication via reverse proxy (X-Remote-User header) + session-based
  - Response envelope: Consistent `{"data": ...}` structure for resource endpoints
  - Error handling: Standardized error response format with code and message

- **[/api/admin/openapi.yaml](/api/admin/openapi.yaml)** - Admin server (Port 14000)
  - Canonical OpenAPI documentation for all administrative APIs
  - Endpoints: Health check, agent management (CRUD), service management (CRUD)
  - Authentication: Pre-authentication via reverse proxy (admin-level access controlled upstream)
  - Security emphasis: Client secrets always redacted in responses (SR-003)
  - Referential integrity: 409 Conflict responses when deleting services with active grants
  - OAuth2 support: Service metadata for OIDC discovery

**Key Features**:
- **OpenAPI 3.0.3 Compliant**: Specifications follow OpenAPI 3.0.3 standard for interoperability
- **Zalando Guidelines Compliant**: APIs follow Zalando RESTful API and Event Guidelines (https://opensource.zalando.com/restful-api-guidelines/)
- **Pre-Implementation Design**: All APIs designed and documented before implementation (Constitution Principle X)
- **Examples Included**: Realistic examples for all endpoints covering success and error cases
- **User-Confirmed**: API specifications confirmed with users/stakeholders before implementation (Constitution Principle IV & X)

**Dual-Port Architecture**:
```
End-User Server (Port 8000):
  ├── GET /health
  ├── GET /api/me
  └── /api/consent/*
      ├── GET /agents
      ├── GET /agent/{agent-id}
      ├── POST /agent/{agent-id}/grants

Admin Server (Port 14000):
  ├── GET /health
  ├── /api/agents/*
  │   ├── POST / (create)
  │   ├── GET / (list)
  │   ├── GET /{id}
  │   ├── PUT /{id}
  │   └── DELETE /{id}
  └── /api/services/*
      ├── POST / (create)
      ├── GET / (list)
      ├── GET /{id}
      ├── PUT /{id}
      └── DELETE /{id}
```

**Usage**:
- Import specifications into Swagger UI, Redoc, or other OpenAPI tooling
- Generate API client libraries for multiple languages via OpenAPI generators
- Validate API implementation compliance against documented spec
- Reference for integration testing and contract validation

#### 3.1.4.1. OAuth2 Server Mode (`local`)

**Mode Selection**: The broker operates in one of two mutually exclusive modes, configured via `oauth2.auth_server.mode`:

| Mode | Value | Behavior |
|------|-------|----------|
| Proxy (default) | `proxy` | Forwards OAuth2 requests to an upstream authorization server. The broker acts as a mediating proxy and does not mint tokens. |
| Local | `local` | The broker acts as a standalone OAuth2 authorization server, minting its own JWT access tokens signed with managed asymmetric keys. |

**Strategy Pattern**: Handler behavior switches at startup based on mode:

- `OAuth2AuthorizeHandler` uses an `AuthorizationCodeIssuer` strategy — **nil** in proxy mode, non-nil in `local` mode. When nil, authorization requests are forwarded upstream; when non-nil, the broker generates authorization codes locally.
- `OAuth2TokenHandler` uses a `TokenMintingStrategy` — **nil** in proxy mode, non-nil in `local` mode. When nil, token requests are proxied upstream; when non-nil, the broker mints JWT access tokens.

**Type Containment**: All [fosite](https://github.com/ory/fosite) OAuth2 server types are contained in `internal/domain/oauth2server/`. This package encapsulates the OAuth2 authorization server domain logic (authorization code storage, client authentication, token signing) and **never leaks fosite types** into ports, adapters/http, or app packages.

**Import Rules**:
- `internal/domain/oauth2server/` may import `internal/ports/` and `internal/domain/` packages.
- `internal/domain/oauth2server/` must **never** import adapter packages or `internal/app/`.
- No other package in the codebase may import fosite types directly — all interaction flows through `oauth2server` domain interfaces.

**Endpoints added in `local` mode**:
```
End-User Server (Port 8000):
  ├── GET  /.well-known/oauth-authorization-server   (RFC 8414 discovery)
  ├── GET  /oauth2/jwks.json                         (signing key set)
  ├── GET  /oauth2/authorize                         (authorization code grant)
  └── POST /oauth2/token                             (token issuance)

Admin Server (Port 14000):
  ├── /api/agents/{id}/client-credentials/*
  │   ├── POST   /                (generate broker-issued credentials)
  │   ├── GET    /                (get credential metadata)
  │   └── DELETE /                (revoke credentials)
  └── /api/oauth2-server/signing-keys/*
      ├── POST   /                (add signing key)
      ├── GET    /                (list active keys)
      ├── PUT    /{kid}/current   (promote key to current)
      └── DELETE /{kid}           (remove key)
```

#### 3.1.5. Encryption Vault for OAuth Tokens (Feature 012)

**Purpose**: Secure at-rest encryption of OAuth2 tokens using envelope encryption with AWS Encryption SDK, protecting tokens from unauthorized access while maintaining developer transparency.

**Architecture**: Port-adapter pattern implementing EncryptionPort interface with AWS KMS hierarchical keyring (production) and environment variable KEK injection (development).

**Core Components**:

- **EncryptionPort** (internal/ports/encryption.go): Domain interface defining Encrypt/Decrypt methods with encryption context parameter
- **AWSAdapter** (internal/adapters/encryption/aws/): AWS Encryption SDK implementation with:
  - Envelope encryption: DEK-per-token with KEK wrapping
  - Context binding: Service isolation via encryption context AAD
  - Support for AWS KMS ARN (production) and ${ENV_VAR} (development)
  - Memory protection via AWS SDK baseline
  - Hierarchical keyring with DynamoDB branch key caching (production)

**Encryption Model**:

```
Token Encryption Flow:
  1. Generate fresh DEK (Data Encryption Key) for this token
  2. Encrypt token plaintext with DEK using AESGCMSIV
  3. Bind encryption context (service_id) to DEK encryption as AAD
  4. Wrap DEK with KEK (from AWS KMS or environment)
  5. Bind same encryption context to KEK wrapping as AAD
  6. Return serialized envelope: [wrapped_DEK || ciphertext || auth_tag]

Token Decryption Flow:
  1. Extract wrapped DEK, ciphertext, auth_tag from envelope
  2. Provide encryption context (service_id) to decoder
  3. Unwrap DEK with KEK, verifying context matches (context mismatch = fail-closed)
  4. Decrypt ciphertext with DEK, verifying auth tag
  5. Verify context at both DEK and KEK layers
  6. Return plaintext token or error
```

**KEK Storage Mechanisms**:

1. **Production (AWS KMS ARN)**:
   - KEK reference via AWS KMS customer-managed key ARN
   - Hierarchical keyring uses DynamoDB for branch key caching
   - Reduces KMS API calls while maintaining security
   - Configuration: `encryption.key: "arn:aws:kms:region:account:key/key-id"`
   - No plaintext KEK in application memory (AWS SDK handles)

2. **Development (Base64-Encoded Key)**:
   - KEK provided as base64-encoded AES-256 key (typically from environment variable)
   - Raw AES keyring (no AWS KMS dependency)
   - Configuration: `encryption.key: "${ENCRYPTION_KEK}"` (resolves to base64 key)
   - Environment variable interpolation allows flexible key injection

**Service Integration**:

- **OAuth2SessionService**: Transparently encrypts tokens on CreateSession, decrypts on retrieval
- **UserSessionRepository**: Stores EncryptedAccessToken and EncryptedRefreshToken as BYTEA columns
- **ThirdpartyOAuth2ProviderService** (`internal/domain/thirdparty/`): Exclusively owns encryption/decryption of provider `client_secret` via the `Secret` value object (see below). No other layer touches `EncryptionPort` for provider secrets.
- No manual encryption steps required in calling code - encryption is transparent

**Secret Value Object** (`internal/domain/model/secret.go`):

The `Secret` value object enforces encryption safety at the type level for provider credentials:

```
Create flow:
  Handler → ThirdpartyOAuth2ProviderEntity{Secret: NewPlaintextSecret(req.ClientSecret)}
  ThirdpartyOAuth2ProviderService.Create():
    → Secret.GetPlaintext() → encrypt via EncryptionPort → NewEncryptedSecret(ciphertext)
    → repo.Create(entity)  ← entity.Secret is now encrypted; plaintext is gone
  Repository adapter:
    → Secret.GetCiphertext() → store as BYTEA  ← fails if not encrypted first

Retrieval flow:
  Repository adapter → NewEncryptedSecret(row.SecretCiphertext)
  ThirdpartyOAuth2ProviderService.Get():
    → Secret.GetCiphertext() → decrypt via EncryptionPort → NewPlaintextSecret(plaintext)
    → return entity  ← plaintext available to caller only through service boundary
```

**Mandatory Encryption**: Encryption is required in all environments. Builder returns a startup error if no encryption configuration is provided — there is no NoOp fallback. This is enforced in `internal/app/builder.go`.

**Security Properties**:

- **Context Binding**: Service-level isolation - tokens encrypted for service A cannot be used for service B (context verification at DEK and KEK layers)
- **Fail-Closed**: No plaintext fallback on encryption/decryption failure (errors propagate)
- **Authenticated Encryption**: AESGCMSIV provides both confidentiality and authenticity
- **Fresh DEK Per Token**: Unique DEK for each token prevents cross-token analysis
- **Audit Logging**: All operations logged with error_kind, service_id, and operation type

**Performance**:

- Local encryption/decryption: <5ms per operation
- AWS KMS operations: 50-200ms (depends on KMS latency + DynamoDB branch key caching)
- Session operations: <100ms typical (includes token encryption overhead)
- Branch key cache improves production performance by reducing KMS calls

**Testing**:

- E2E tests (24 scenarios) covering all acceptance criteria from spec
- Unit tests for adapter error handling, context verification, DEK uniqueness
- Integration tests with LocalStack KMS and real PostgreSQL storage
- Backward compatibility tests for KEK rotation scenarios

**Encryption Architecture Pattern**:

The project follows **Domain Service Encryption** for all sensitive data encryption. This pattern maintains hexagonal architecture purity by placing encryption logic in the domain service layer rather than the storage adapter layer.

**Standard Pattern**:
- Encryption logic resides in domain service layer (business concern)
- Repository stores opaque encrypted bytes (infrastructure concern)
- Clean hexagonal architecture boundaries
- No encryption dependencies in storage adapters

**Data Flow**:
```
Service Layer (OAuth2SessionService):
  ↓ Encrypts tokens with EncryptionPort
  ↓ Creates domain entity with encrypted bytes
Repository Layer (UserSessionRepository):
  ↓ Stores encrypted bytes as BYTEA (opaque)
  ↓ Returns encrypted bytes on retrieval
Service Layer (OAuth2SessionService):
  ↓ Decrypts tokens with EncryptionPort
  ↓ Returns plaintext to caller
```

**Example Implementations**: UserSessionRepository + OAuth2SessionService (tokens); ThirdpartyOAuth2ProviderRepository + ThirdpartyOAuth2ProviderService (provider client_secret via `Secret` VO)

**See Also**:
- ADR 009: Envelope Encryption Design (cryptographic approach)
- ADR 012: Encryption Layer Separation (architectural pattern)

#### 3.1.x. Client ID Metadata Document (CIMD) Subsystem

**Purpose**: Allow AI agents to identify themselves via a publicly resolvable HTTPS URL as `client_id`. The broker fetches a JSON document from that URL, validates it, and presents its metadata on the consent screen. Enabled via `oauth2_authorization_server.cimd.enabled`.

**New Port**: `internal/ports/cimd.go` defines `CIMDFetcher` (outbound, infrastructure-side) and `ClientResolver` (strategy interface injected into `OAuth2AuthorizationService`).

**New Domain Package**: `internal/domain/oauth2/cimd/` contains `ClientIDMetadataDocumentURL`, `SSRFBlocklist`, `ClientIDMetadataDocument`, `CIMDCache`, `CIMDService`.

**Authorization Flow with URL-based `client_id`**:

```
OAuth2 /authorize request
  ↓ ClientResolver.ResolveClient(client_id)
  ↓  ├─ URL detected → AgentClientResolver (cimdService != nil)
  ↓  │    ↓ Validate URL (scheme, path, no credentials, no dot-segments)
  ↓  │    ↓ AgentRepository.GetByClientURI → resolve Agent
  ↓  │    ↓ CIMDService.FetchAndValidate(url, agent)
  ↓  │         ↓ Cache hit? → return cached document
  ↓  │         ↓ CIMDFetcher.Fetch (SSRF blocklist enforced at dial time)
  ↓  │         ↓ Validate: client_id match, redirect_uris present, auth_method safe
  ↓  │         ↓ Cache store with HTTP-header-derived TTL (clamped to operator bounds)
  ↓  │    ↓ Return ClientResolution{Agent, CIMDDocument}
  ↓  └─ UUID detected → AgentClientResolver (opaque path, cimdService may be nil)
  ↓ HandleAuthorization: CIMD metadata present → create AuthorizationSession
  ↓ Redirect to consent with ?session_id= (no CIMD params in URL)
  ↓ Consent handler loads AuthorizationSession (trusted server-side state)
  ↓ User grants → grants endpoint consumes session → authorization code redirect
```

**Security Properties**: SSRF blocked at TCP-connect time (TOCTOU-safe); CIMD params never relay through browser URL (AuthorizationSession binds context server-side, SR-013/SR-014).

**See Also**: ADR 015 — CIMD Fetcher Architecture (SSRF hardening, caching, strategy pattern)

### 3.2. Envoy External Processor (ExtProc) Token Exchange Service

**Name**: extproc-token-exchange

**Purpose**: Standalone gRPC microservice that implements the Envoy External Processor protocol for transparent OAuth2 token exchange. When deployed alongside agentgateway, the service intercepts incoming HTTP requests via Envoy's ExtProc filter, extracts Bearer tokens from request headers, performs RFC 8693 token exchange against the identity broker, and replaces the Authorization header with the exchanged token. Exchanged tokens are cached in-memory with singleflight deduplication to optimize performance.

**Architecture**: Hexagonal (ports and adapters)

**Components**:

#### 3.2.1. Domain Model

**Exchanger** (Port): Interface defining token exchange logic as a port. Implementations perform RFC 8693 token exchange and manage token caching independently of the gRPC protocol.

**TokenExchanger** (Adapter): Concrete implementation of Exchanger. Manages:
- In-memory token cache keyed by `(subjectToken, resourceURI)` struct
- Background client assertion refresh goroutine (startup + 80% TTL/30s threshold)
- Singleflight deduplication for concurrent exchange requests
- HTTP client for RFC 8693 token exchange requests

**CachedToken**: Value object representing a cached token with:
- `accessToken`: The exchanged token value
- `expiresAt`: Absolute expiration timestamp
- Computed from `expires_in` response field (capped at `cache.max_ttl`, defaulting to `cache.default_ttl`)

**TokenCacheKey**: Struct used as Go map key: `{subjectToken, resourceURI}`. Using struct keys prevents separator-injection attacks compared to string concatenation.

#### 3.2.2. gRPC Server

**Server**: Implements Envoy's `ExternalProcessorServer` interface with:
- **Process RPC**: Streaming bidirectional RPC handling all Envoy ExtProc phases (RequestHeaders, RequestBody, ResponseHeaders, ResponseBody, RequestTrailers, ResponseTrailers)
- **RequestHeaders Phase**: Primary processing phase where Bearer token extraction and exchange occurs
- **Other Phases**: Pass-through responses with phase-specific response types
- **ImmediateResponse**: Error response mechanism (500 on exchange failure, 503 on invalid URI)

#### 3.2.3. Request Processing

**RequestHeaders Processing**:
1. Extract Bearer token from Authorization header (pass through if absent or non-Bearer)
2. Extract resource URI from `:path` pseudo-header
3. Validate URI (absolute, http/https scheme, non-empty host) — SSRF mitigation
4. Call `Exchanger.Exchange()` with (token, uri)
5. On success: replace Authorization header with `"Bearer " + exchangedToken`
6. On failure: return 500 ImmediateResponse, reject request, log failure

**Helper Functions**:
- `extractBearerToken()`: Parse "Bearer <token>" format
- `extractHeader()`: Case-insensitive header lookup
- `validateResourceURI()`: URI parsing and scheme validation
- `replaceAuthorizationHeader()`: Build HeadersResponse with header mutation
- `immediateResponse()`: Build ImmediateResponse with status code and JSON body

#### 3.2.4. Configuration

**Configuration Subsystem**: Separate from broker config, uses `EXTPROC_` environment prefix.

**Config Structure**:
- **GRPCConfig**: `bind`, `port`, `max_concurrent_streams`
- **OAuth2Config**: `issuer`, `token_endpoint`, `client_id`, `client_secret`, `client_credentials_endpoint`, `client_credentials_scopes`, `client_assertion_type`, `exchange_timeout`, `tls`
- **TLSConfig**: `allow_http` (fail-closed unless true), `insecure_skip_verify`, `ca_bundle_path`
- **CacheConfig**: `default_ttl`, `max_ttl`
- **LogConfig**: `level`, `format`
- **CircuitBreakerConfig**: `max_failures`, `reset_timeout`
- **MetricsConfig**: `enabled`, `export_interval`
- **LogsConfig**: `enabled`
- **TelemetryConfig**: `enabled`, `service_name`, `resource_attributes`, `traces` (enabled, sampling_rate, propagators), `metrics` (enabled, export_interval), `logs` (enabled), `exporter` (protocol, endpoint, insecure, headers, timeout, compression)

**Validation Rules** (19 rules, fail-fast at startup):
1. grpc.port must be 1–65535
2. grpc.bind must not be empty
3. oauth2.token_endpoint must be a valid URL with http/https scheme
4. oauth2.issuer must be a valid URL with http/https scheme
5. oauth2.client_id must not be empty
6. oauth2.client_secret must not be empty after env var expansion
7. cache.default_ttl must be a positive duration
8. Token endpoint and issuer must use https:// unless `oauth2.tls.allow_http: true`
9. cache.max_ttl must be a positive duration
10. oauth2.exchange_timeout must be a positive duration
11. log.level must be one of debug, info, warn, error
12. log.format must be one of text, json
13. oauth2.client_assertion_type must be one of id_token, access_token
14. circuit_breaker.max_failures must be >= 1
15. circuit_breaker.reset_timeout must be a positive duration
16. telemetry.exporter.endpoint must not be empty when telemetry enabled
17. telemetry.exporter.protocol must be one of grpc, http, https
18. telemetry.traces.sampling_rate must be in range [0.0, 1.0]
19. telemetry.exporter.timeout must be a positive duration when telemetry enabled

**Configuration Loading**:
- Viper-based loader with EXTPROC_ prefix
- YAML file with `${VAR}` expansion for secret injection
- Example config: `examples/config/extproc-token-exchange.yaml`

#### 3.2.5. Security Features

**Fail-Closed**: Exchange failures return 500 ImmediateResponse; original bearer token is never forwarded.

**SSRF Mitigation**: `validateResourceURI()` enforces absolute URIs with http/https schemes only.

**Token Redaction**: Bearer tokens and exchanged tokens absent from all logs and error responses.

**Client Secret Protection**: client_secret stored in memory only (config), never logged (logged as `[REDACTED]`), redacted from error responses.

**Client Assertion Refresh**: Background goroutine acquires ID token from client_credentials grant at startup (fail-fast), refreshes proactively within 30s of expiry, retries on failure.

**Cache Expiry**: Tokens automatically expired and evicted; background goroutine sweeps every `cache.default_ttl / 2`.

**TLS Enforcement**: Token endpoint and issuer must use https:// unless `oauth2.tls.allow_http: true` (development only).

#### 3.2.6. Performance

- **RequestHeaders processing**: <10ms typical (cache hit)
- **Token exchange**: <200ms typical (cache miss, includes HTTP RPC)
- **Singleflight deduplication**: Concurrent requests for same key trigger exactly one exchange
- **Cache eviction**: Background goroutine runs non-blocking

**Throughput**: Handles 1000s of requests/sec at typical latencies.

#### 3.2.7. Deployment

**Entry Point**: `cmd/extproc-token-exchange/main.go` with Cobra CLI

**Binary**: `extproc-token-exchange` (single Go binary, ~24MB)

**Containerization**: `Dockerfile` for Docker Compose integration

**Lifecycle**:
- Load configuration (fail-fast on invalid)
- Initialize logger
- Initialize telemetry provider (if enabled; bounded by exporter timeout, interruptible by signal)
- Wire slog-to-OTel bridge (if telemetry + logs enabled)
- Create TokenExchanger (acquires client assertion; otelhttp wraps outbound HTTP)
- Create gRPC server
- Register ExternalProcessorServer
- Listen on configured bind/port
- Handle graceful shutdown on signal (GracefulStop → telemetry flush with 5s deadline)

#### 3.2.8. Testing

**E2E Test Suite** (separate from main broker tests): `tests/e2e/extproc/`
- 12 acceptance tests (1:1 mapping to spec scenarios)
- Ginkgo/Gomega BDD framework
- In-process bootstrap (gRPC server + mock OAuth2 servers)
- Comprehensive coverage: token exchange, caching, singleflight, URI validation, startup validation

**Unit Tests**: `internal/extproc/server/*_test.go`, `internal/extproc/config/*_test.go`
- 50+ tests covering all paths
- RFC 8693 request/response format validation
- Client assertion acquisition and refresh
- Cache TTL capping and eviction
- Error handling and security features
- Race detector clean

**Coverage**: 79.9% (server + config packages)

**Technologies**:
- Ginkgo v2 (BDD test framework)
- Gomega (assertion library)
- `google.golang.org/grpc` (gRPC framework)
- `github.com/envoyproxy/go-control-plane` (ExtProc protocol)
- `golang.org/x/sync/singleflight` (concurrent request deduplication)

#### 3.2.9. Dependencies

**Direct**: google.golang.org/grpc, github.com/envoyproxy/go-control-plane, golang.org/x/sync/singleflight, spf13/cobra, spf13/viper

**Testing**: github.com/onsi/ginkgo/v2, github.com/onsi/gomega, httptest (standard library)

**Storage**: In-memory only (no database required)

#### 3.2.10. Directory Structure

```
cmd/extproc-token-exchange/
├── main.go                    # Entry point
└── root.go                    # Cobra root command, gRPC server lifecycle

internal/extproc/
├── config/
│   ├── config.go              # Configuration types
│   ├── loader.go              # Viper loader + validation
│   ├── validate.go            # Validation rules
│   └── loader_test.go         # Config tests
└── server/
    ├── server.go              # ExtProc gRPC Process RPC
    ├── exchanger.go           # TokenExchanger with cache & singleflight
    ├── server_test.go         # Server unit tests
    └── exchanger_test.go      # Exchange unit tests

tests/e2e/extproc/
├── extproc_suite_test.go      # Ginkgo suite runner
├── token_exchange_test.go     # 12 acceptance tests
├── bootstrap/
│   └── bootstrap.go           # Server setup + mock servers
├── helpers/
│   ├── grpc_helpers.go        # gRPC utilities
│   └── matchers.go            # Gomega matchers
└── fixtures/
    ├── tokens.go              # Test token fixtures
    └── configs.go             # Configuration fixtures

examples/config/
└── extproc-token-exchange.yaml # Documented example config
```

#### 3.2.11. Telemetry Support

ExtProc leverages the broker's shared OpenTelemetry infrastructure (ADR 011, ADR 027) for distributed tracing, metrics, and log correlation. The `cmd/extproc-token-exchange/` composition layer imports `internal/adapters/telemetry` to initialize and manage the OTel provider lifecycle. The ExtProc domain and gRPC server use the global OTel tracer and meter patterns; no changes to adapter constructors are required. See ADR 027 for the cross-boundary dependency rationale and implementation approach.

## 4. Data Stores

(List and describe the databases and other persistent storage solutions used.)

### 4.1. [Data Store Type 1]

Name: [e.g., Primary User Database, Analytics Data Warehouse]

Type: [e.g., PostgreSQL, MongoDB, Redis, S3, Firestore]

Purpose: [Briefly describe what data it stores and why.]

Key Schemas/Collections: [List important tables/collections, e.g., users, products, orders (no need for full schema, just names)]

### 4.2. [Data Store Type 2]

Name: [e.g., Cache, Message Queue]

Type: [e.g., Redis, Kafka, RabbitMQ]

Purpose: [Briefly describe its purpose, e.g., "Used for caching frequently accessed data" or "Inter-service communication."]

## 5. External Integrations / APIs

(List any third-party services or external APIs the system interacts with.)

Service Name 1: [e.g., Stripe, SendGrid, Google Maps API]

Purpose: [Briefly describe its function, e.g., "Payment processing."]

Integration Method: [e.g., REST API, SDK]

## 6. Deployment & Infrastructure

Cloud Provider: [e.g., AWS, GCP, Azure, On-premise]

Key Services Used: [e.g., EC2, Lambda, S3, RDS, Kubernetes, Cloud Functions, App Engine]

CI/CD Pipeline: [e.g., GitHub Actions, GitLab CI, Jenkins, CircleCI]

Monitoring & Logging: [e.g., Prometheus, Grafana, CloudWatch, Stackdriver, ELK Stack]

## 7. Security Considerations

(Highlight any critical security aspects, authentication mechanisms, or data encryption practices.)

Authentication: [e.g., OAuth2, JWT, API Keys]

Authorization: [e.g., RBAC, ACLs]

Data Encryption: [e.g., TLS in transit, AES-256 at rest]

Key Security Tools/Practices: [e.g., WAF, regular security audits]

## 8. Development & Testing Environment

Local Setup Instructions: [Link to CONTRIBUTING.md or brief steps]

Testing Frameworks: [e.g., Jest, Pytest, JUnit]

Code Quality Tools: [e.g., ESLint, Black, SonarQube]

## 9. Future Considerations / Roadmap

(Briefly note any known architectural debts, planned major changes, or significant future features that might impact the architecture.)

[e.g., "Migrate from monolith to microservices."]

[e.g., "Implement event-driven architecture for real-time updates."]

## 10. Architecture Decision Records (ADRs)

This section lists all architectural decisions made for this project. ADRs document important technical choices, their rationale, alternatives considered, and consequences.

### Core Infrastructure
- [ADR 002: Configuration Libraries](adrs/002-configuration-libraries.md) - Multi-source configuration with Viper, Cobra, and godotenv
- [ADR 003: Chi Framework Selection](adrs/003-chi-framework.md) - HTTP routing framework choice
- [ADR 004: Dual-Server Isolation](adrs/004-dual-server-isolation.md) - Separate end-user and admin servers
- [ADR 004: Storage Layer Architecture](adrs/004-storage-layer-architecture.md) - Hexagonal architecture for persistence

### Frontend & User Interface
- [ADR 005: SPA Serving Pattern](adrs/005-spa-serving-pattern.md) - Serving React SPA from Go backend
- [ADR 006: Frontend Stack](adrs/006-frontend-stack.md) - React 18 + Vite + Tailwind CSS v4

### Testing & Quality
- [ADR 007: E2E Testing with Ginkgo](adrs/007-e2e-testing-with-ginkgo.md) - BDD-style E2E tests using production bootstrap

### RFC 8693 Token Exchange
- [ADR 008: Token Exchange JWKS Adapter Pattern](adrs/008-token-exchange-jwks-adapter-pattern.md) - HTTP abstraction for JWKS fetching and caching

### Security & Encryption
- [ADR 008: Encryption Context Optimization](adrs/008-encryption-context-optimization.md) - Service-ID-only context binding performance optimization
- [ADR 009: Envelope Encryption Design](adrs/009-envelope-encryption-design.md) - DEK-per-session with AWS KMS and context binding
- [ADR 010: CDK Encryption Infrastructure](adrs/010-cdk-encryption-infrastructure.md) - AWS CDK (Go) for KMS, DynamoDB, and IAM provisioning
- [ADR 012: Encryption Layer Separation](adrs/012-encryption-layer-separation.md) - Domain service encryption pattern for hexagonal architecture

### Observability
- [ADR 011: OpenTelemetry Provider Pattern](adrs/011-opentelemetry-provider-pattern.md) - App-layer OTel provider, otelchi middleware choice, context-based span propagation

### Client ID Metadata Document (CIMD)
- [ADR 015: CIMD Fetcher Architecture](adrs/015-cimd-fetcher-architecture.md) - SSRF-hardened HTTP client, in-process caching, hexagonal port, strategy pattern for opaque vs URL-based client IDs

## 11. Project Identification

Project Name: Agentic Identity Broker

Repository URL: [Insert Repository URL]

Primary Contact/Team: [Insert Lead Developer/Team Name]

Date of Last Update: 2026-01-06

## 12. Glossary / Acronyms

Define any project-specific terms or acronyms.)

### Configuration Domain

**Configuration Schema**: The complete structure defining all valid configuration options including their types, default values, validation rules, and sensitivity level. Represented by the Config struct in code.

**Configuration Source**: A source of configuration data (defaults, .env files, YAML file, CLI flags) with associated precedence level and loading mechanism. Each source contributes values that may override lower-precedence sources.

**Environment Variable Reference**: A placeholder in configuration (using ${VAR_NAME} syntax) that references an environment variable for runtime substitution. Supports nested expansion with circular reference detection.

**Source Precedence**: The priority order determining which configuration value wins when multiple sources provide the same key. Order (lowest to highest): Defaults < .env Files < YAML < CLI Flags.

**Sensitive Value**: Configuration value that should be redacted in logs and output. Identified by IDENTITY_BROKER_ prefix or keywords (password, secret, token, key, credential, auth).

**ConfigPort**: Hexagonal architecture port (interface) for accessing configuration. Domain logic depends on this interface, not concrete implementations.

**Configuration Adapter**: Implementation of ConfigPort using Viper/Cobra/godotenv. Located in internal/config/ directory.

### Encryption Domain

**Envelope Encryption**: A cryptographic pattern where data is encrypted with a Data Encryption Key (DEK), then the DEK is encrypted with a Key Encryption Key (KEK). This enables secure storage with only a single KMS call per session while protecting token material with symmetric encryption.

**DEK**: Data Encryption Key. A symmetric encryption key (AES-256) used to encrypt sensitive data like OAuth2 tokens. Generated randomly per session, never stored in plaintext, and always wrapped by the KEK before storage.

**KEK**: Key Encryption Key. A key used to encrypt/wrap the DEK. In AWS implementation, this is an AWS KMS customer-managed key (CMK) referenced by ARN. The KEK never leaves the secure boundary and is managed by AWS KMS.

**EncryptionContext**: Additional authenticated data (AAD) bound to ciphertext during encryption but not encrypted itself. Used to provide cryptographic isolation between different services. Implemented as a map[string]string containing only the service_id field for performance optimization (ADR 008).

**EncryptionPort**: Hexagonal architecture interface for encryption operations. Abstracts the domain from specific encryption implementations (AWS KMS, envelope encryption, etc.), allowing testability and implementation flexibility while ensuring consistent encryption behavior.

**AAD**: Additional Authenticated Data. Data that is authenticated but not encrypted as part of AEAD (Authenticated Encryption with Associated Data) schemes. Used in encryption context to prevent cross-context token usage.

**AESGCMSIV**: AES in Galois/Counter Mode with Synthetic Initialization Vector. A misuse-resistant authenticated encryption mode that provides both confidentiality and authenticity. Used for DEK-based token encryption with deterministic nonce generation.

### Session Management Domain

**Principal**: The authenticated user identifier extracted from an HTTP header set by a reverse proxy after user authentication. Typically an email address, username, or unique ID. Examples: "alice@example.com", "user-123". Maximum length: 200 characters. Retrieved from context using principal.FromContext(ctx).

**Principal Extraction**: The process of reading a principal value from a configured HTTP header, validating it, and making it available throughout request processing. Implemented by RequirePrincipalMiddleware (rejects invalid) and OptionalPrincipalMiddleware (non-rejecting).

**Request Context**: Go context.Context object passed through an HTTP request and its downstream handlers, carrying request-scoped values including the authenticated principal. Context is created per-request and cancelled when the request completes. Principal is stored using an unexported context key type for type safety.

**RequirePrincipalMiddleware**: HTTP middleware that validates principal presence and validity, rejecting requests with missing/empty principals (401 Unauthorized) or oversized principals (400 Bad Request). Applied to protected routes requiring authentication. Located in internal/adapters/http/principal_middleware.go.

**OptionalPrincipalMiddleware**: HTTP middleware that extracts principals if present and valid, but never rejects requests. Applied globally to all routes, allowing downstream handlers to check for optional authentication. Located in internal/adapters/http/principal_middleware.go.

**Principal Header Name**: Configuration key specifying which HTTP header contains the principal value (e.g., "X-Remote-User", "X-Authenticated-User"). Configurable per server via servers.*.authentication.preauth.principal_header_name. Default: "X-Remote-User".

**PreAuthenticationConfig**: Configuration structure enabling pre-authentication mode where a trusted reverse proxy handles authentication and provides the principal via HTTP header. Supports future extension with JWT and other authentication methods. Located in internal/ports/config.go.

### Domain Model and Consent Management

**Agent**: An AI agent registered in the identity broker system. Each agent has a unique client_id (the `ClientID` field — a short opaque identifier used in existing OAuth2/consent flows via `GetByClientID`; optional in the Admin API, and when omitted there is no auto-generation fallback per ADR 017, so `client_id` remains NULL), display name, description, and optional URLs for governance documentation and user interface. Agents may additionally register one or more **client_uris** (Client ID Metadata Document URLs per IETF draft-parecki-oauth-client-id-metadata-document) which provide an alternative resolution path via `GetByClientURI` for CIMD-aware clients. The `client_id` remains the canonical primary identifier when present; client_uris are supplementary discovery handles that resolve to the same Agent entity. Agents request delegated OAuth2 permissions from users through the consent flow. Optionally, agents may specify service requirements (mandatory and optional third-party services with required scopes).

**ServiceRequirement**: A value object representing a single third-party OAuth2 service that an agent requires or can optionally use. Each requirement specifies: (1) service_id - which third-party service (UUID reference), (2) requirement_type - whether "mandatory" or "optional", and (3) required_scopes - which OAuth2 scopes must be granted (string array). Stored as JSONB in the agent's service_requirements column. Validates structure at domain layer and referential integrity at application layer.

**RequirementType**: Enum with two values: "mandatory" (agent cannot function without this service, authorization blocked until requirement satisfied) and "optional" (agent can use if available, authorization proceeds regardless). Case-sensitive, lowercase only. Controls authorization flow behavior - only mandatory requirements block authorization.

**Mandatory Service**: A third-party OAuth2 service marked with requirement_type="mandatory" in an agent's service requirements. Authorization flow validates that user has an active, non-expired OAuth2 session with all mandatory services and that session scopes are a superset of required_scopes (case-sensitive). If validation fails, user is redirected to consent screen to establish missing sessions before authorization continues.

**Optional Service**: A third-party OAuth2 service marked with requirement_type="optional" in an agent's service requirements. Displayed in consent UI with visual distinction (neutral badge vs trust-deep for mandatory). Does not block authorization flow - if user lacks session or scopes, authorization proceeds anyway. Allows agents to degrade gracefully when optional integrations unavailable.

**ThirdpartyOAuth2Provider**: External OAuth2 provider (e.g., GitHub, Google, Microsoft) registered in the system. Each provider defines a set of OAuth scopes that can be delegated to agents. Providers have a client_id, client_secret (stored as a `Secret` value object — encrypted at rest, redacted in API responses), and display name. Represented as `model.ThirdpartyOAuth2ProviderEntity` in `internal/domain/model/`. All encryption and decryption of the client secret is owned exclusively by `ThirdpartyOAuth2ProviderService` in `internal/domain/thirdparty/`.

**Secret**: Immutable value object in `internal/domain/model/` with two mutually exclusive states: plaintext (`NewPlaintextSecret(value)`) and encrypted (`NewEncryptedSecret(ciphertext)`). `GetPlaintext()` fails on encrypted state; `GetCiphertext()` fails on plaintext state. `Redacted()` always returns `"REDACTED"` regardless of state. Prevents accidental plaintext leakage at the type level — storage adapters can never accidentally persist unencrypted secrets because `GetCiphertext()` will error if encryption was not performed first.

**OAuth Scope**: A specific permission defined by an OAuth2 provider (e.g., "repo", "user:email"). Each scope has a scope_value (the OAuth scope string) and a human-readable description. Scopes are defined per service and validated during grant creation.

**User Grant**: A record of a user (principal) delegating specific OAuth2 scopes to an agent for one or more third-party services. Grants have an optional expiration time (valid_until) and can be revoked at any time. Each user can have at most one active grant per agent (upsert semantics).

**GrantRevoked** — Domain event representing a user's explicit deletion of their grant for an agent. Emitted as a structured audit log entry carrying principal, agent_id, grant_id, and revoked_at.

**Delegated Token**: Component of a grant specifying which OAuth2 service and which scopes from that service are delegated to the agent. A single grant can contain multiple delegated tokens for different services. Format: {thirdparty_oauth2_service_id, scopes[]}.

**Grant Expiration**: The point at which a grant becomes inactive (valid_until < NOW()). Expired grants are filtered out when listing grants. Grants with valid_until=null never expire (indefinite grants). Users must specify future timestamps when creating grants.

**Consent Flow**: The process where a user reviews agent metadata and available OAuth2 services, then decides which permissions to grant. Implemented via GET /api/consent/agent/:agent-id (view info) and POST /api/consent/agent/:agent-id/grants (grant permissions).

**Scope Validation**: Business rule (FR-018) enforcing that all requested scopes in a grant must exist in the corresponding service's scope configuration. Invalid scopes are rejected with a 400 error listing which scopes are not defined.

**Cascade Delete**: When an agent is deleted, all user grants referencing that agent are automatically deleted (FR-020). This maintains referential integrity and prevents orphaned grants. Implemented at the repository layer.

**Service Protection**: Business rule preventing deletion of an OAuth2 service if any active grants reference it (returns 409 Conflict). Ensures grants don't reference non-existent services. Requires revocation of all referencing grants before service deletion.

**OAuth2Flavor**: Named enumeration on `ThirdpartyOAuth2Service` identifying the credential format and future token acquisition mechanism. Current values: `standard` (plain client secret string), `google` (Google service account JSON key). Designed for extension. Stored in the `oauth2_flavor` column of `thirdparty_oauth2_services`. Defaults to `standard` for backward compatibility.

**ClientCredential**: The authentication material stored in the `client_secret` field of a `ThirdpartyOAuth2Service`. Structure varies by `OAuth2Flavor`: a plain secret string for `standard`, a serialized Google service account JSON string for `google`. Always encrypted at rest. Field name preserved for API backward compatibility.

**GoogleServiceAccountKey**: Structured value object representing the parsed contents of a Google service account JSON key file. Required fields: `type` (must be `"service_account"`), `private_key`, `client_email`, `token_uri`, `client_id`. Validated structurally; cryptographic format of the private key is not verified at configuration time. Parsed exclusively during request validation; not stored as a separate entity.

### AWS Encryption Vault Domain Model

**UserSession**: Domain aggregate representing the complete lifecycle of a user's session with a third-party OAuth2 provider. Contains encrypted access/refresh tokens, expiration metadata, and manages token encryption/decryption through the EncryptionPort. Enforces one session per (principal, service_id) with automatic token refresh and secure deletion.

**EncryptionContext**: Domain value object containing metadata that cryptographically binds encrypted tokens to their usage context. Implemented as an immutable map[string]string with service_id as the primary binding field. Prevents cross-service token usage and provides audit trail for encryption operations.

**EncryptionPort**: Port interface defining the boundary between domain logic and encryption adapters. Provides Encrypt/Decrypt methods with context parameter, enabling the domain to remain independent of specific encryption implementations (AWS KMS, local encryption, etc.). Implementations perform envelope encryption with DEK-per-session pattern and context binding validation.

### Multi-Agent OAuth2 Client Delegation

**MultiAgentClientConfig**: Configuration value object enabling multiple agents to share one upstream OAuth2 client ID. Contains feature gate (`enabled`), `agent_id_param_name`, and `agent_id_claim_name`.

**resolveAgentIdByClientId**: CEL helper function (registered only when feature is disabled) that maps an upstream `client_id` claim value to the broker's internal `agent.id`. Safe because client_id uniqueness is enforced in disabled mode.

### Third-Party OAuth2 Session Management

**UserSession**: An authenticated OAuth2 session between a user (principal) and a third-party service. Contains encrypted access/refresh tokens, scope, and expiration metadata. One session per (principal, service_id) pair enforced by database unique constraint. Aggregate root that owns the encrypted tokens and manages session lifecycle.

**OAuth2StateToken**: A JWE-encrypted ephemeral token that binds an OAuth2 callback to the initiating request. Contains principal, PKCE verifier, service_id, and redirect_uri claims. Short-lived (10 min TTL, max 15 min per spec) to limit CSRF exposure. Uses authenticated encryption (A256GCMKW + A256GCM) for tamper detection.

**PKCE**: Proof Key for Code Exchange (RFC 7636). Security extension for OAuth2 that prevents authorization code interception attacks. Uses code_verifier (random 32-128 byte secret, base64url-encoded) and code_challenge (SHA256 hash of verifier). Mandatory for all OAuth2 flows with no bypass allowed.

**Token Vault**: Secure storage for encrypted OAuth2 tokens. Tokens are encrypted using AES-GCM with encryption context binding them to principal, service_id, and session_id. Uses EncryptionPort for all cryptographic operations. All tokens stored as ciphertext (BYTEA in PostgreSQL).

**Session Termination**: User-initiated action to delete their OAuth2 session with a third-party service. Removes encrypted tokens from storage and displays warning about affected agents before deletion. Idempotent operation (safe to terminate non-existent sessions).

**OAuth2SessionService**: Domain service that orchestrates OAuth2 authorization flows and session lifecycle. Handles PKCE generation, JWE state token management, authorization URL construction, callback processing, token exchange with retry logic, session encryption/storage, and termination with dependent agent warnings.

**Encryption Context**: Additional authenticated data (AAD) included in token encryption. Binds ciphertext to principal, service_id, session_id, and purpose ("oauth2_token"). Stored as JSONB in PostgreSQL. Used for auditing and prevents cross-context token usage (tokens encrypted for one session cannot be decrypted for another).

### RFC 8693 Token Exchange

**TokenExchangeRequest**: RFC 8693 token exchange request containing grant_type, subject_token, client_assertion, and resource parameters. Parsed from form-urlencoded POST body to /oauth2/token endpoint. Immutable value object after parsing.

**TokenExchangeResponse**: RFC 8693 compliant response containing access_token, token_type, issued_token_type, and optional expires_in. Returned as JSON from successful token exchange. Format enables clients to use the exchanged token with third-party services.

**ClientAssertion**: JWT authenticating the privileged client (API gateway or reverse proxy) making the token exchange request. Contains privileged client identifier in 'sub' claim. Validated against upstream OAuth2 server's JWKS. Represents the privileged client's identity and authorization to perform token exchange.

**SubjectToken**: JWT containing both user principal and agent identifier from the upstream OAuth2 server. Principal extracted via configurable CEL expression (default: sub claim). Agent identifier extracted via configurable CEL expression (default: azp claim). Identifies the end-user and agent on whose behalf token exchange is requested.

**ResourceURI**: URI identifying the target resource or third-party service for token exchange. Normalized (trailing slashes removed) before storage and lookup. Matched against service protected_resources to determine which third-party service to exchange tokens for. Example: "https://api.github.com" or "https://github.com/api/v3".

**Privileged Client**: API gateway or reverse proxy that initiates token exchange on behalf of agents. Authenticates using client_assertion JWT. Acts as intermediary between agent and identity broker, passing through user's subject_token for exchange.

**CEL Authorization**: Common Expression Language policy evaluation for privileged client authorization. Expression evaluated against client_assertion claims and request context. Expression must return boolean; defaults to "true" (allow all valid privileged clients). Enables flexible authorization policies beyond basic JWT validation.

**Protected Resources**: Array of normalized resource URIs on ThirdpartyOAuth2Provider that identify which resources map to that provider for RFC 8693 token exchange. Used to discover correct service when processing token exchange requests. URIs are normalized (trailing slashes removed) for consistent matching. Stored as TEXT[] column in PostgreSQL with GIN index for efficient lookups.

### ExtProc (Envoy External Processor) Domain

**ExtProc**: Envoy's External Processor (ExtProc) gRPC protocol allowing a standalone microservice to intercept and modify HTTP requests/responses in real-time. The extproc-token-exchange service implements this protocol to transparently exchange OAuth2 tokens.

**agentgateway**: Envoy-based reverse proxy deployed alongside the identity broker and ExtProc service. Configures Envoy's ExtProc filter to delegate token exchange decisions to the extproc-token-exchange microservice. Routes requests from agents through the ExtProc filter before forwarding to upstream services.

**Exchanged Token**: OAuth2 access token obtained via RFC 8693 token exchange, scoped to a specific downstream service (resource URI). Replaces the original Bearer token in request headers. Used by agents to access third-party services without exposing their original credentials.

**Singleflight Refresh**: Deduplication pattern preventing concurrent duplicate token exchange requests for identical (subjectToken, resourceURI) pairs. Uses `golang.org/x/sync/singleflight` to ensure exactly one exchange request completes while others wait for the result. Improves performance and reduces load on identity broker.

**Client Assertion**: JWT containing privileged client credentials (API gateway or reverse proxy) used to authenticate the token exchange request to the identity broker. Generated via client_credentials grant at ExtProc startup. Automatically refreshed in background goroutine within 30 seconds of expiry.

**MCP Streamable HTTP**: Model Context Protocol transport mode allowing JSON-RPC communication over HTTP with streaming capabilities. Used by ExtProc to forward tool calls to MCP servers while maintaining transparent token exchange for authentication.

### OAuth2 Server Mode

**BrokerClientCredential**: OAuth2 client credentials generated by the broker and bound to exactly one Agent. Contains hashed client secret (Argon2id, PHC format). `client_id` equals the agent's UUID string — no separate field needed. One credential per agent enforced by UNIQUE on `client_credentials.client_id`. Lifecycle: generated on demand via Admin API, replaced atomically on rotation, cascade-deleted with agent. Located in `internal/domain/storage/broker_client_credential.go`.

**SigningKey**: Asymmetric key pair (ES256 or RS256) used to sign locally-issued JWT access tokens. Private material stored PEM-encoded and encrypted via `EncryptionPort`. Exactly one key marked `is_current` at any time. Keys remain in JWKS until explicitly soft-deleted via `removed_at`. Located in `internal/domain/storage/signing_key.go`.

**AuthorizationCode**: Ephemeral, single-use code issued by the authorization endpoint and exchanged for an access token. Stored as SHA-256 hash. Expires after 60 seconds. Invalidated atomically on first use via `UPDATE ... SET used_at WHERE used_at IS NULL`. PKCE (S256) always required. Located in `internal/domain/storage/authorization_code.go`.

**ClientID** (credential): String-typed identifier for broker-issued OAuth2 client credentials. Always equals the agent's UUID string. Globally unique. Bound to authorization codes for rotation safety. Located in `internal/domain/id/string_ids.go`.

**KeyID**: String-typed identifier for JWT signing keys (`kid` claim). UUID format, immutable after creation. Used in JWT headers to identify the signing key for token validation. Located in `internal/domain/id/string_ids.go`.

**OAuth2ServerProvider**: Domain service (`Provider` struct) in `internal/domain/oauth2server/provider.go` that wires fosite protocol handlers (AuthorizeExplicitGrantHandler, ClientCredentialsGrantHandler, pkce.Handler) with custom strategies (JWXAccessTokenStrategy, RandomCodeStrategy). Exposes domain-native methods (HandleClientCredentials, HandleAuthorize, HandleAuthorizationCodeExchange). Fosite types are contained within this package and never leak into ports or HTTP handlers.

**TokenClaimsExpression**: CEL expression evaluated at token issuance time to produce custom JWT claims. Has access to `agent`, `principal`, and `request` variables. Return type must be `map[string]dyn`. Base claim keys (iss, sub, iat, exp, jti, kid, agent_id, scope) are silently stripped from the result to prevent override. Compiled at startup — invalid expressions cause startup failure (fail-closed). Located in `internal/domain/oauth2server/token_claims_cel.go`.

**Domain Model Invariants**: (1) One credential per agent — enforced by UNIQUE constraint on `client_credentials.client_id`. (2) Exactly one `is_current` signing key among active keys — enforced by application logic in `SigningKeyService` and transactional `SetCurrent` in PostgreSQL adapter. (3) Authorization codes are single-use with 60-second TTL — enforced by atomic `MarkUsed` (UPDATE WHERE used_at IS NULL) and expiry check before token exchange.

**ModeStrategy**: Domain interface that determines whether a classified agent is permitted in the active OAuth server mode. Single method: `AcceptsClientMode(ClientMode) bool`. Three implementations wired by the builder at startup: `proxyModeStrategy` (accepts ProxyClient only), `localModeStrategy` (accepts CIMDClient and LocalClient), `hybridModeStrategy` (accepts all client modes). Strategy is injected once at startup — no runtime mode checks in handlers. Located in `internal/domain/oauth2/mode_strategy.go`.

**OAuthServerMode**: Enumeration (`internal/domain/oauth2/servermode`) defining the three legal broker operating modes: `proxy` (all agents forwarded to an upstream OAuth2 server), `local` (all tokens minted locally by the broker), `hybrid` (both proxy and local agents coexist; dispatch per request based on `ClientMode`). Stored as a string in config; typed as `servermode.Mode` to prevent unchecked string comparisons in handler and service code.

**ClientMode**: Enum (`internal/domain/storage`) classifying an Agent at request time based on its registered identifiers. `ProxyClient` — has a `ClientID`; requests forwarded to upstream. `LocalClient` — no `ClientID` and no `ClientURIs`; tokens minted locally. `CIMDClient` — has `ClientURIs` but no `ClientID`; tokens minted locally after CIMD document fetch. `AmbiguousClient` — has both `ClientID` and `ClientURIs`; rejected as `invalid_client` during client resolution. Computed by `Agent.ClientMode()` — never stored.

**ProxyModeConfig**: Configuration value object (`internal/ports/config.go`) carrying the upstream OAuth2 server coordinates required when `mode` is `proxy` or `hybrid`: `upstream_issuer_uri`, `upstream_authorize_endpoint`, `upstream_token_endpoint`, `upstream_timeout_seconds`. All fields are ignored (and must be empty) in `local` mode.

**LocalModeConfig**: Configuration value object (`internal/ports/config.go`) carrying local token issuance parameters required when `mode` is `local` or `hybrid`: `token_ttl`, `token_claims_expression`, and optional `issuer_uri`. `issuer_uri` overrides the JWT `iss` claim independently of `server.enduser.public_url`, enabling deployments behind CDNs or reverse proxies. Defaults to `server.enduser.public_url` when absent. All fields are ignored (and must be empty) in `proxy` mode.

**TokenGrantResolution**: Port-layer DTO (`internal/ports/oauth2.go`) returned by `OAuth2Service.ResolveForTokenGrant`. Carries `AgentID`, `ClientID` (nil for local/CIMD agents), and `ClientMode` — the minimal scalar projection of `storage.Agent` that the token grant adapter layer needs. The domain entity itself (`storage.Agent`) is consumed by the service and never crosses the adapter boundary.

**TokenGrantStrategy**: Adapter-layer interface (`internal/adapters/http/enduser`) for processing OAuth2 token grant requests. Receives `*ports.TokenGrantResolution` rather than a domain entity. Three implementations: `proxyTokenGrantStrategy` (forwards to upstream, replaces broker UUID with upstream `client_id`), `localGrantStrategy` (delegates to fosite via `TokenMintingStrategy`), `hybridTokenGrantStrategy` (dispatches to proxy or local sub-strategy based on `resolution.ClientMode`).

**AuthorizationProceedStrategy**: Adapter-layer interface for the "proceed" branch of an authorization decision — invoked when a grant exists and the broker should advance the flow. `proxyProceedStrategy` issues a 302 to `decision.RedirectURL` (the upstream authorize URL). `localProceedStrategy` calls `AuthorizationCodeIssuer.IssueAuthorizationCode` and redirects with `code=` to the client's `redirect_uri`. `hybridProceedStrategy` dispatches to proxy or local based on `decision.ClientMode`.

**Mode Strategy Pattern**: The mechanism by which `OAuthServerMode` drives the entire request-handling topology at startup time rather than via runtime branching. The builder selects and wires the appropriate `ModeStrategy` (domain, controls `AcceptsClientMode`) and `AuthorizationProceedStrategy`/`TokenGrantStrategy` (adapters) based on the configured mode. For hybrid mode, both proxy and local strategies are created and wrapped in dispatching composites. Handlers and services never inspect the configured mode string — they receive pre-wired strategies.

### Client ID Metadata Document (CIMD) Domain

**ClientIDMetadataDocument**: Immutable value object representing a parsed and validated CIMD JSON document fetched from a client's registered HTTPS URL. Validated at construction time: `client_id` field must exactly match the fetch URL, `redirect_uris` must not be empty, `token_endpoint_auth_method` must not be a client-secret variant, and `client_name` must not match the keyword blocklist. Located in `internal/domain/oauth2/cimd/`.

**CIMDCacheEntry**: In-process (non-persisted) cache record keyed by the Client ID Metadata Document URL. Fields: URL (cache key), parsed Document, FetchedAt timestamp, ExpiresAt (computed from HTTP cache headers clamped to operator TTL bounds). Stored in a `sync.RWMutex`-protected map; expired entries are lazily evicted on next access. Located in `internal/domain/oauth2/cimd/`.

**ClientIDMetadataDocumentURL**: Value object representing a validated HTTPS URL used as a `client_id`. Validated at parse time — invalid URLs cannot be constructed. Enforces: HTTPS scheme only, non-empty path component, no `.`/`..` path segments, no fragment (`#`), no userinfo (credentials), and port must be 443 or absent. Located in `internal/domain/oauth2/cimd/`.

**SSRFBlocklist**: Immutable value object holding the set of CIDR ranges blocked for CIMD HTTP fetches. Initialized at startup from RFC 6890 Special-Purpose Address Registry defaults plus operator `extra_blocked_cidrs`. Consulted by the SSRF-hardened fetcher adapter's custom `net.Dialer.Control` callback to reject resolved IP addresses before TCP connect. Located in `internal/domain/oauth2/cimd/`.

**BrandPinMismatchDetected**: Domain audit event emitted as a structured log entry when a CIMD document's `client_name` differs from the registered Agent's `DisplayName`. Non-blocking — authorization proceeds, but the mismatch is recorded. Fields: AgentID, Agent.DisplayName, CIMD client_name.

**ClientResolver**: Strategy interface injected into `OAuth2AuthorizationService` that resolves a `client_id` from an authorization request to an Agent and optional CIMD metadata. One implementation — `AgentClientResolver` — serves both modes, selected by `cimdService` presence: when nil (CIMD disabled), URL-format client IDs are rejected with `invalid_client`; when set (CIMD enabled), URL-format client IDs are routed through CIMD fetch/validate/cache, non-URL IDs fall through to UUID lookup. Located in `internal/ports/cimd.go` (interface) and `internal/domain/oauth2/client_resolver.go`.

**CIMDFetcher**: Hexagonal port interface (outbound, infrastructure-side) for fetching Client ID Metadata Documents from remote HTTPS endpoints with SSRF protection, configurable timeout, and response size limits. Analogous to `JWKSPort`. Implemented by the SSRF-hardened HTTP fetcher adapter in `internal/adapters/cimd/fetcher.go` which uses a custom `net.Dialer.Control` callback for TOCTOU-safe IP address validation before TCP connect.

**ClientResolution**: DTO returned by `ClientResolver.ResolveClient()`. Contains the resolved `*storage.Agent` and an optional `*cimd.ClientIDMetadataDocument` (nil for opaque UUID client IDs). Used by `OAuth2AuthorizationService` to carry CIMD metadata into the consent session.

### General Acronyms

**ADR**: Architecture Decision Record - Documents important architectural decisions and their rationale

**CLI**: Command-Line Interface

**YAML**: Yet Another Markup Language (configuration file format)

**TLS**: Transport Layer Security

**RBAC**: Role-Based Access Control

### JWT Pre-Authentication

**PrincipalProfile**: Enriched user identity value object containing principal identifier, display name, email, and picture URL. Extracted from pre-authentication source (JWT or plain header). Request-scoped, immutable. Stored in request context via `principal.WithProfile()` alongside the existing string principal for backward compatibility. Located in `internal/domain/principal/profile.go`.

**JWTAuthConfig**: Configuration value object defining JWT-based pre-authentication behavior: HTTP header name, verification mode (`jwks` or `none`), JWKS endpoint, audience/issuer constraints, and CEL claim extraction expressions. Validated at startup with mutual exclusivity rules (`verification: none` + `jwks_uri` → startup error). Located in `internal/ports/config.go` as `JWTConfig`.

**JWTAuthenticator**: Port interface for JWT authentication in the pre-auth layer. Abstracts JWT parsing, signature verification (JWKS or none), temporal validation, and CEL-based claim extraction. Returns `AuthResult` containing extracted principal and optional profile attributes. Implemented by jwx adapter in `internal/adapters/jwtauth/`. Located in `internal/domain/jwtauth/authenticator.go`.

**JWTVerificationMode**: String enum (`"jwks"` or `"none"`) controlling JWT signature verification behavior. `"jwks"` (default) requires JWKS URI and validates cryptographic signatures against published key sets. `"none"` accepts unsigned JWTs (alg: "none") for trusted upstream environments such as service meshes. Unsigned mode requires explicit opt-in and is mutually exclusive with `jwks_uri`.

**JWTValidationFailed**: Domain event emitted when JWT pre-authentication fails. Contains failure reason (e.g., `invalid_signature`, `token_expired`, `audience_mismatch`), header name, and remote address. Logged as structured audit data for security monitoring per FR-020/SR-005. Not persisted — emitted as structured log entries.
