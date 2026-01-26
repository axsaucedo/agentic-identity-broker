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
│   ├── tsconfig.json     # TypeScript configuration
│   └── tailwind.config.ts # Tailwind CSS v4.0 configuration
├── docs/                 # Project documentation (e.g., API docs, setup guides)
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
├── tsconfig.json         # TypeScript configuration
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

#### 3.1.3. End-to-End Testing Architecture

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
- No manual encryption steps required in calling code - encryption is transparent

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

### Security & Encryption
- [ADR 008: Encryption Context Optimization](adrs/008-encryption-context-optimization.md) - Service-ID-only context binding performance optimization
- [ADR 009: Envelope Encryption Design](adrs/009-envelope-encryption-design.md) - DEK-per-session with AWS KMS and context binding

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

**EncryptionContext**: Additional authenticated data (AAD) bound to ciphertext during encryption but not encrypted itself. Used to provide cryptographic isolation between different services or tenants. Implemented as a map[string]string containing service_id and other binding metadata.

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

**Agent**: An AI agent registered in the identity broker system. Each agent has a unique client_id, display name, description, and optional URLs for governance documentation and user interface. Agents request delegated OAuth2 permissions from users through the consent flow. Optionally, agents may specify service requirements (mandatory and optional third-party services with required scopes).

**ServiceRequirement**: A value object representing a single third-party OAuth2 service that an agent requires or can optionally use. Each requirement specifies: (1) service_id - which third-party service (UUID reference), (2) requirement_type - whether "mandatory" or "optional", and (3) required_scopes - which OAuth2 scopes must be granted (string array). Stored as JSONB in the agent's service_requirements column. Validates structure at domain layer and referential integrity at application layer.

**RequirementType**: Enum with two values: "mandatory" (agent cannot function without this service, authorization blocked until requirement satisfied) and "optional" (agent can use if available, authorization proceeds regardless). Case-sensitive, lowercase only. Controls authorization flow behavior - only mandatory requirements block authorization.

**Mandatory Service**: A third-party OAuth2 service marked with requirement_type="mandatory" in an agent's service requirements. Authorization flow validates that user has an active, non-expired OAuth2 session with all mandatory services and that session scopes are a superset of required_scopes (case-sensitive). If validation fails, user is redirected to consent screen to establish missing sessions before authorization continues.

**Optional Service**: A third-party OAuth2 service marked with requirement_type="optional" in an agent's service requirements. Displayed in consent UI with visual distinction (neutral badge vs trust-deep for mandatory). Does not block authorization flow - if user lacks session or scopes, authorization proceeds anyway. Allows agents to degrade gracefully when optional integrations unavailable.

**ThirdpartyOAuth2Service**: External OAuth2 provider (e.g., GitHub, Google, Microsoft) registered in the system. Each service defines a set of OAuth scopes that can be delegated to agents. Services have a client_id, client_secret (stored securely, redacted in responses), and display name.

**OAuth Scope**: A specific permission defined by an OAuth2 provider (e.g., "repo", "user:email"). Each scope has a scope_value (the OAuth scope string) and a human-readable description. Scopes are defined per service and validated during grant creation.

**User Grant**: A record of a user (principal) delegating specific OAuth2 scopes to an agent for one or more third-party services. Grants have an optional expiration time (valid_until) and can be revoked at any time. Each user can have at most one active grant per agent (upsert semantics).

**Delegated Token**: Component of a grant specifying which OAuth2 service and which scopes from that service are delegated to the agent. A single grant can contain multiple delegated tokens for different services. Format: {thirdparty_oauth2_service_id, scopes[]}.

**Grant Expiration**: The point at which a grant becomes inactive (valid_until < NOW()). Expired grants are filtered out when listing grants. Grants with valid_until=null never expire (indefinite grants). Users must specify future timestamps when creating grants.

**Consent Flow**: The process where a user reviews agent metadata and available OAuth2 services, then decides which permissions to grant. Implemented via GET /api/consent/agent/:agent-id (view info) and POST /api/consent/agent/:agent-id/grants (grant permissions).

**Scope Validation**: Business rule (FR-018) enforcing that all requested scopes in a grant must exist in the corresponding service's scope configuration. Invalid scopes are rejected with a 400 error listing which scopes are not defined.

**Cascade Delete**: When an agent is deleted, all user grants referencing that agent are automatically deleted (FR-020). This maintains referential integrity and prevents orphaned grants. Implemented at the repository layer.

**Service Protection**: Business rule preventing deletion of an OAuth2 service if any active grants reference it (returns 409 Conflict). Ensures grants don't reference non-existent services. Requires revocation of all referencing grants before service deletion.

### AWS Encryption Vault Domain Model

**UserSession**: Domain aggregate representing the complete lifecycle of a user's session with a third-party OAuth2 provider. Contains encrypted access/refresh tokens, expiration metadata, and manages token encryption/decryption through the EncryptionPort. Enforces one session per (principal, service_id) with automatic token refresh and secure deletion.

**EncryptionContext**: Domain value object containing metadata that cryptographically binds encrypted tokens to their usage context. Implemented as an immutable map[string]string with service_id as the primary binding field. Prevents cross-service token usage and provides audit trail for encryption operations.

**EncryptionPort**: Port interface defining the boundary between domain logic and encryption adapters. Provides Encrypt/Decrypt methods with context parameter, enabling the domain to remain independent of specific encryption implementations (AWS KMS, local encryption, etc.). Implementations perform envelope encryption with DEK-per-session pattern and context binding validation.

### Third-Party OAuth2 Session Management

**UserSession**: An authenticated OAuth2 session between a user (principal) and a third-party service. Contains encrypted access/refresh tokens, scope, and expiration metadata. One session per (principal, service_id) pair enforced by database unique constraint. Aggregate root that owns the encrypted tokens and manages session lifecycle.

**OAuth2StateToken**: A JWE-encrypted ephemeral token that binds an OAuth2 callback to the initiating request. Contains principal, PKCE verifier, service_id, and redirect_uri claims. Short-lived (10 min TTL, max 15 min per spec) to limit CSRF exposure. Uses authenticated encryption (A256GCMKW + A256GCM) for tamper detection.

**PKCE**: Proof Key for Code Exchange (RFC 7636). Security extension for OAuth2 that prevents authorization code interception attacks. Uses code_verifier (random 32-128 byte secret, base64url-encoded) and code_challenge (SHA256 hash of verifier). Mandatory for all OAuth2 flows with no bypass allowed.

**Token Vault**: Secure storage for encrypted OAuth2 tokens. Tokens are encrypted using AES-GCM with encryption context binding them to principal, service_id, and session_id. Uses EncryptionPort for all cryptographic operations. All tokens stored as ciphertext (BYTEA in PostgreSQL).

**Session Termination**: User-initiated action to delete their OAuth2 session with a third-party service. Removes encrypted tokens from storage and displays warning about affected agents before deletion. Idempotent operation (safe to terminate non-existent sessions).

**OAuth2SessionService**: Domain service that orchestrates OAuth2 authorization flows and session lifecycle. Handles PKCE generation, JWE state token management, authorization URL construction, callback processing, token exchange with retry logic, session encryption/storage, and termination with dependent agent warnings.

**Encryption Context**: Additional authenticated data (AAD) included in token encryption. Binds ciphertext to principal, service_id, session_id, and purpose ("oauth2_token"). Stored as JSONB in PostgreSQL. Used for auditing and prevents cross-context token usage (tokens encrypted for one session cannot be decrypted for another).

### General Acronyms

**ADR**: Architecture Decision Record - Documents important architectural decisions and their rationale

**CLI**: Command-Line Interface

**YAML**: Yet Another Markup Language (configuration file format)

**TLS**: Transport Layer Security

**RBAC**: Role-Based Access Control