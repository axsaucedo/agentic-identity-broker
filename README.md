# Agentic Identity Broker

A secure identity management, authentication, and authorization service for AI agents and autonomous systems. Built with Go using hexagonal architecture principles for maintainability and extensibility.

## Features

- Flexible multi-source configuration (defaults, .env files, YAML, CLI flags)
- Security-first design with sensitive value redaction and command injection prevention
- Hexagonal architecture with clear separation of concerns
- Environment-specific configuration support
- Comprehensive validation and audit logging

## Prerequisites

- Go 1.23.0 or higher
- [just](https://github.com/casey/just) command runner (optional but recommended)
- [Air](https://github.com/air-verse/air) for hot-reload development (optional)
- [golangci-lint](https://golangci-lint.run/) for code quality checks (optional)

## Quick Start

### Installation

1. Clone the repository:
```bash
git clone https://github.com/agentic-identity-broker/agentic-identity-broker.git
cd agentic-identity-broker
```

2. Install dependencies:
```bash
just deps
```

3. Install development tools (optional):
```bash
just install-tools
```

### Building

Build the application:
```bash
just build
```

The binary will be created at `./bin/identity-broker`.

For a production-optimized build with smaller binary size:
```bash
just build-release
```

### Running

Run the application:
```bash
just run
```

Or run directly:
```bash
./bin/identity-broker
```

### Development

Start the development server with hot-reload:
```bash
just dev
```

This uses [Air](https://github.com/air-verse/air) to automatically rebuild and restart the application when source files change.

## Configuration

The application supports multiple configuration sources with the following precedence (lowest to highest):

1. **Defaults**: Built-in default values
2. **.env Files**: Environment-specific files (`.env` → `.env.local` → `.env.{GO_ENV}` → `.env.{GO_ENV}.local`)
3. **YAML File**: `config.yaml` with `${VAR}` environment variable substitution
4. **CLI Flags**: Command-line flags override all other sources

### Configuration Files

Create a `.env` file for local development:
```env
IDENTITY_BROKER_LOG_LEVEL=debug
IDENTITY_BROKER_LOG_FORMAT=json
```

Or use a `config.yaml` file:
```yaml
log:
  level: debug
  format: json
```

See [examples/](examples/) directory for more configuration examples.

## Development Commands

All development tasks are managed using [just](https://github.com/casey/just). Run `just --list` to see all available commands:

### Building & Running
- `just build` - Build the binary to ./bin/identity-broker
- `just build-release` - Build optimized binary for production (30% smaller)
- `just run` - Build and run the application
- `just dev` - Start hot-reload development server
- `just clean` - Remove build artifacts

### Testing & Quality
- `just test` - Run all tests with race detection
- `just test-coverage` - Generate HTML coverage report
- `just test-coverage-summary` - Display coverage summary in terminal
- `just check` - Run all quality checks (fmt, vet, lint, test)

### Code Quality
- `just fmt` - Format code with gofmt
- `just vet` - Run go vet static analysis
- `just lint` - Run golangci-lint (with fallback to go vet)

### Dependencies & Tools
- `just deps` - Manage Go module dependencies
- `just install-tools` - Install Air and golangci-lint

### Documentation
- `just docs-serve` - Start documentation server locally
- `just docs-build` - Build documentation site
- `just docs-deploy` - Deploy documentation to GitHub Pages

## Testing

Run all tests:
```bash
just test
```

Generate coverage report:
```bash
just test-coverage
```

The coverage report will be generated at `coverage/coverage.html`.

## Code Quality

Format code:
```bash
just fmt
```

Run static analysis:
```bash
just vet
```

Run linter:
```bash
just lint
```

Run all quality checks:
```bash
just check
```

## Architecture

The project follows hexagonal architecture (ports and adapters pattern) with clear separation between:

- **Domain Logic**: Core business logic independent of external concerns
- **Ports**: Interfaces defining boundaries between layers
- **Adapters**: Implementations of ports using specific technologies

For detailed architecture documentation, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Project Structure

```
.
├── cmd/                    # Application entry points
│   └── identity-broker/    # Main application
├── internal/               # Private application code
│   ├── config/            # Configuration adapter
│   ├── domain/            # Domain models and logic
│   └── ports/             # Port interfaces
├── examples/              # Configuration examples
├── docs/                  # User documentation
├── assets/docusaurus/     # Documentation site
└── tests/                 # Integration tests
```

## Security

The application implements security-first design principles:

- **Sensitive Value Redaction**: Automatic redaction of passwords, tokens, and secrets in logs
- **Command Injection Prevention**: Validation to prevent shell command injection
- **Circular Reference Detection**: Protection against infinite loops in configuration
- **Fail-Closed**: Graceful termination on configuration errors
- **Audit Logging**: Structured JSON audit logs for compliance

For security concerns, please see our security policy.

## Contributing

We welcome contributions! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run quality checks (`just check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Consent Frontend Development

The consent frontend is a React-based Single Page Application (SPA) for managing OAuth2 delegations to AI agents.

### Prerequisites

- Node.js 18.0.0 or higher
- npm (comes with Node.js)

### Setup

1. Navigate to the web directory:
```bash
cd web
```

2. Install dependencies:
```bash
npm install
```

3. Configure environment (optional):
```bash
# Create .env file for local development
cp .env.example .env
```

### Development Workflow

There are several ways to develop the consent frontend depending on your needs.

#### Quick Start with Justfile (Recommended)

The easiest way to run both backend and frontend:

```bash
# Terminal 1: Start Go backend
just run

# Terminal 2: Start frontend dev server with hot reload
just web-dev
```

Access the frontend at http://localhost:3000 (Vite dev server with HMR)
API requests automatically proxy to http://localhost:8000 (Go backend)

**How the Vite Proxy Works:**

The Vite dev server is configured to proxy API requests from the frontend to the backend:

```
Browser Request: http://localhost:3000/api/consent/agents
       ↓
Vite Dev Server (port 3000)
       ↓ (proxy configuration in vite.config.ts)
       ↓ Automatically injects: X-Remote-User: dev@example.com
Go Backend (port 8000)
       ↓
API Handler returns JSON
       ↓
Vite proxies response back to browser
```

**Authentication During Development:**
The Vite proxy automatically adds the `X-Remote-User: dev@example.com` header to all API requests. This simulates the authentication header that would normally be set by an upstream proxy (oauth2-proxy, nginx, etc.) in production. All API requests appear as if they're coming from the authenticated user `dev@example.com`.

This setup provides:
- **Hot Module Replacement (HMR)**: Instant updates without page reload
- **Same-origin requests**: No CORS issues during development
- **API debugging**: See requests in Go backend logs
- **Frontend debugging**: Use React DevTools in browser
- **Simulated authentication**: X-Remote-User header automatically injected for local testing

**All Frontend Commands:**

```bash
just web-install          # Install npm dependencies
just web-dev              # Start Vite dev server (port 3000)
just web-build            # Build production bundle to web/dist/consent/
just build-all            # Build both Go backend and frontend
```

#### Alternative: Integrated Build (No HMR)

Build frontend and run from Go backend:
```bash
cd web && npm run build && cd .. && just run
```

Access at http://localhost:8000/consent (no hot reload, requires rebuild for changes)

#### Frontend-Only Development

**Start Development Server:**
```bash
cd web && npm run dev
# or from project root:
just web-dev
```
This starts the Vite dev server at http://localhost:3000 with hot module replacement.

**Run Tests:**
```bash
npm run test              # Run tests in watch mode
npm run test:coverage     # Generate coverage report
```

**Lint and Format:**
```bash
npm run lint              # Run ESLint
npm run format            # Format code with Prettier
```

**Build for Production:**
```bash
npm run build
# or from project root:
just web-build
```
This compiles TypeScript and bundles the app to `web/dist/consent/` directory.

**Preview Production Build:**
```bash
npm run preview
```

### Integration with Go Backend

The frontend is served by the Go backend at `/consent`:

1. **Build the frontend**: `just web-build` (creates `web/dist/consent/`)
2. **Start the backend**: `just run` (from project root)
3. **Access the app**: http://localhost:8000/consent

The backend serves static files from `web/dist/consent/` and handles API requests at `/api/consent/*`.

### Environment Variables

Frontend configuration (optional `.env` file in `web/` directory):

```bash
# API base URL (defaults to /api)
VITE_API_BASE_URL=/api

# Development server port (defaults to 3000)
VITE_DEV_SERVER_PORT=3000
```

### Project Structure

```
web/
├── src/
│   ├── components/       # React components
│   │   ├── consent/      # Consent-specific components
│   │   ├── layout/       # Layout components
│   │   └── ui/           # Reusable UI components
│   ├── pages/            # Application pages
│   ├── hooks/            # Custom React hooks
│   ├── services/         # API client and services
│   ├── types/            # TypeScript type definitions
│   └── utils/            # Utility functions
├── dist/consent/         # Build output (served by Go backend)
├── package.json          # Dependencies and scripts
├── vite.config.ts        # Vite configuration
├── tsconfig.json         # TypeScript configuration
└── tailwind.config.ts    # Tailwind CSS configuration
```

### Technology Stack

- **React 18.2+**: Frontend framework
- **TypeScript 5.3+**: Type-safe JavaScript
- **Vite 5.0+**: Fast build tool with HMR
- **Tailwind CSS v4.0**: Utility-first CSS framework
- **Headless UI**: Accessible UI components
- **Axios**: HTTP client
- **React Router DOM**: Client-side routing
- **Vitest**: Unit testing framework

### Testing

**Run Tests:**
```bash
npm run test
```

**Coverage Report:**
```bash
npm run test:coverage
open coverage/index.html
```

**Test Structure:**
- Unit tests: `*.test.tsx` or `*.test.ts`
- Integration tests: `*.integration.test.tsx`
- Test files located next to source files

### Deployment

**Build for Production:**
```bash
npm run build
```

Output: `web/dist/consent/` directory with optimized static files.

**Deploy with Go Backend:**
1. Build frontend: `cd web && npm run build`
2. Build Go binary: `just build-release`
3. Deploy `bin/identity-broker` with embedded `web/dist/consent/`

The Go backend automatically serves the SPA from the embedded directory.

### Troubleshooting

**Port Already in Use:**
```bash
# Change port in vite.config.ts or use environment variable
VITE_DEV_SERVER_PORT=3001 npm run dev
```

**API Connection Issues:**
- Verify Go backend is running on port 8000
- Check Vite proxy configuration in `vite.config.ts`
- Ensure CORS is configured correctly

**Build Errors:**
```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

**TypeScript Errors:**
```bash
# Regenerate TypeScript types
npm run build
```

## Documentation

- [Architecture Overview](ARCHITECTURE.md) - System architecture and design decisions
- [User Documentation](docs/) - End-user guides and tutorials
- [API Documentation](docs/api/) - REST API reference
- [Documentation Site](assets/docusaurus/) - Full documentation website

Run the documentation server locally:
```bash
just docs-serve
```

## License

[Specify your license here]

## Contact

[Specify contact information or team here]
