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

## Documentation

- [Architecture Overview](ARCHITECTURE.md) - System architecture and design decisions
- [User Documentation](docs/) - End-user guides and tutorials
- [Documentation Site](assets/docusaurus/) - Full documentation website

Run the documentation server locally:
```bash
just docs-serve
```

## License

[Specify your license here]

## Contact

[Specify contact information or team here]
