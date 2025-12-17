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
├── web/                  # Contains all client-side code for user interfaces
│   ├── src/              # Main source code for frontend applications
│   │   ├── components/   # Reusable UI components
│   │   ├── pages/        # Application pages/views
│   │   ├── assets/       # Images, fonts, and other static assets
│   │   ├── services/     # Frontend services for API interaction
│   │   └── store/        # State management (e.g., Redux, Vuex, Context API)
│   ├── public/           # Publicly accessible assets (e.g., index.html)
│   ├── tests/            # Frontend unit and E2E tests
│   └── package.json      # Frontend dependencies and scripts
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

## 10. Project Identification

Project Name: [Insert Project Name]

Repository URL: [Insert Repository URL]

Primary Contact/Team: [Insert Lead Developer/Team Name]

Date of Last Update: [YYYY-MM-DD]

## 11. Glossary / Acronyms

Define any project-specific terms or acronyms.)

### Configuration Domain

**Configuration Schema**: The complete structure defining all valid configuration options including their types, default values, validation rules, and sensitivity level. Represented by the Config struct in code.

**Configuration Source**: A source of configuration data (defaults, .env files, YAML file, CLI flags) with associated precedence level and loading mechanism. Each source contributes values that may override lower-precedence sources.

**Environment Variable Reference**: A placeholder in configuration (using ${VAR_NAME} syntax) that references an environment variable for runtime substitution. Supports nested expansion with circular reference detection.

**Source Precedence**: The priority order determining which configuration value wins when multiple sources provide the same key. Order (lowest to highest): Defaults < .env Files < YAML < CLI Flags.

**Sensitive Value**: Configuration value that should be redacted in logs and output. Identified by IDENTITY_BROKER_ prefix or keywords (password, secret, token, key, credential, auth).

**ConfigPort**: Hexagonal architecture port (interface) for accessing configuration. Domain logic depends on this interface, not concrete implementations.

**Configuration Adapter**: Implementation of ConfigPort using Viper/Cobra/godotenv. Located in internal/config/ directory.

### Session Management Domain

**Principal**: The authenticated user identifier extracted from an HTTP header set by a reverse proxy after user authentication. Typically an email address, username, or unique ID. Examples: "alice@example.com", "user-123". Maximum length: 200 characters. Retrieved from context using principal.FromContext(ctx).

**Principal Extraction**: The process of reading a principal value from a configured HTTP header, validating it, and making it available throughout request processing. Implemented by RequirePrincipalMiddleware (rejects invalid) and OptionalPrincipalMiddleware (non-rejecting).

**Request Context**: Go context.Context object passed through an HTTP request and its downstream handlers, carrying request-scoped values including the authenticated principal. Context is created per-request and cancelled when the request completes. Principal is stored using an unexported context key type for type safety.

**RequirePrincipalMiddleware**: HTTP middleware that validates principal presence and validity, rejecting requests with missing/empty principals (401 Unauthorized) or oversized principals (400 Bad Request). Applied to protected routes requiring authentication. Located in internal/adapters/http/principal_middleware.go.

**OptionalPrincipalMiddleware**: HTTP middleware that extracts principals if present and valid, but never rejects requests. Applied globally to all routes, allowing downstream handlers to check for optional authentication. Located in internal/adapters/http/principal_middleware.go.

**Principal Header Name**: Configuration key specifying which HTTP header contains the principal value (e.g., "X-Remote-User", "X-Authenticated-User"). Configurable per server via servers.*.authentication.preauth.principal_header_name. Default: "X-Remote-User".

**PreAuthenticationConfig**: Configuration structure enabling pre-authentication mode where a trusted reverse proxy handles authentication and provides the principal via HTTP header. Supports future extension with JWT and other authentication methods. Located in internal/ports/config.go.

### General Acronyms

**ADR**: Architecture Decision Record - Documents important architectural decisions and their rationale

**CLI**: Command-Line Interface

**YAML**: Yet Another Markup Language (configuration file format)

**TLS**: Transport Layer Security

**RBAC**: Role-Based Access Control