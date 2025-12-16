# Security Checklist for Storage Layer

This document outlines security practices and checks for the agentic-identity-broker storage layer.

## Connection String Security

### T105: Connection Strings Never Logged in Plain Text

**Objective**: Ensure database connection strings containing credentials are never logged.

**Implementation**:
- ✅ Connection URL is stored in `ports.StorageConfig` but never passed to logging functions
- ✅ `ConnectionParameters.Redacted()` method masks passwords and sensitive query parameters
- ✅ Adapter code never logs `config.Postgres.ConnectionURL` directly
- ✅ Error messages from adapter never include connection strings

**Verification Tests**:
- `TestRedactedConnection_PasswordMasking` - Verifies password masking in redacted URLs
- `TestRedactedConnection_NoCredentials` - Verifies URLs without credentials work correctly
- `TestConnectionURLWithSpecialCharacters` - Verifies special characters are handled safely

**Runtime Behavior**:
```go
// CORRECT: Using redacted URL for logging
redactedURL := cp.Redacted()  // Returns: postgresql://user@localhost:5432/db

// INCORRECT: Would expose password (not used in code)
// log.Error("Connection failed", cp.ConnectionURL)  // postgresql://user:password@host/db
```

### T106: TLS/SSL Support in PostgreSQL Configuration

**Objective**: Ensure TLS/SSL can be configured for PostgreSQL connections.

**Implementation**:
- ✅ PostgreSQL adapter accepts `sslmode` query parameter in connection URL
- ✅ Configuration validates SSL-related query parameters:
  - `sslmode`: require, verify-full, disable (standard PostgreSQL modes)
  - `sslcert`: Path to SSL certificate
  - `sslkey`: Path to SSL key
  - `sslrootcert`: Path to root CA certificate
- ✅ Connection URL validation in `ValidatePostgresURL()` accepts SSL parameters

**Production Configuration**:
```yaml
# Production: Require SSL/TLS
storage:
  backend: postgres
  postgres:
    # Using environment variable with SSL requirement
    connection_url: ${IDENTITY_BROKER_STORAGE_POSTGRES_URL:postgresql://user@localhost/db?sslmode=require}
```

**Verification Tests**:
- `TestConnectionValidation_SSLModeParameter` - Validates sslmode parameter handling
- All SSL-related parameters are accepted in connection URLs

### T107: SSL Verification Failure Causes Startup Failure

**Objective**: Ensure SSL certificate verification failures prevent application startup.

**Implementation**:
- ✅ PostgreSQL adapter calls `db.PingContext()` during `Initialize()` to verify connection
- ✅ Connection errors (including SSL certificate verification failures) return `ErrorKindConnection`
- ✅ Startup sequence fails if `Initialize()` returns an error
- ✅ Application exits with non-zero status code if storage fails to initialize

**Verification**:
- Production configuration uses `sslmode=require` or `sslmode=verify-full`
- Connection pool verification happens before marking storage as ready
- Health check validates SSL connection is maintained

**Test Coverage**:
- `TestPostgresAdapter_Initialize_ConnectionFailed` - Verifies connection failures during init
- `TestPostgresAdapter_HealthCheck_NotInitialized` - Verifies health checks fail for uninitialized DB

## Configuration Security

### T108: Environment Variables Properly Handled

**Objective**: Ensure sensitive configuration values are protected.

**Implementation**:
- ✅ Connection URL sourced from environment variable: `IDENTITY_BROKER_STORAGE_POSTGRES_URL`
- ✅ Configuration redactor identifies sensitive keys:
  - `*_PASSWORD`
  - `*_SECRET`
  - `*_TOKEN`
  - `*_KEY`
  - `IDENTITY_BROKER_*` prefixed variables
- ✅ Sensitive values printed as `[REDACTED]` in logs and debug output
- ✅ Configuration validation never exposes raw values in error messages

**Configuration Usage**:
```bash
# Development: In-memory storage (no credentials)
identity-broker --config config.dev.yaml

# Production: PostgreSQL with SSL
export IDENTITY_BROKER_STORAGE_POSTGRES_URL="postgresql://user:password@prod-db.example.com:5432/identity_broker?sslmode=require"
identity-broker --config config.prod.yaml
```

**Verification Tests**:
- `TestRedact` - Verifies sensitive keys are redacted in configuration
- `TestIsSensitive` - Verifies key classification logic
- Configuration loader properly binds environment variables

### T109: Sensitive Data Never Exposed in Error Messages

**Objective**: Ensure error messages don't accidentally leak credentials or sensitive data.

**Implementation**:
- ✅ `StorageError` wraps adapter errors without exposing sensitive details
- ✅ Error messages use user-friendly descriptions: "failed to connect to PostgreSQL database"
- ✅ Specific error types (`ErrorKind`) allow error handling without exposing internals
- ✅ User IDs and emails redacted in validation errors
- ✅ Database connection errors return generic messages, not raw database errors

**Example Error Messages**:
```go
// CORRECT: Generic message
err := storage.NewStorageError("Initialize", storage.ErrorKindConnection, dbErr,
    "failed to connect to PostgreSQL database")

// WRONG: Would expose credentials (not used in code)
// fmt.Sprintf("failed to connect to %s", connectionURL)
```

## Security Review Checklist for Contributors

When extending the storage layer, ensure:

- [ ] No connection strings are logged or printed in debug output
- [ ] All database errors are wrapped in `StorageError` before being returned
- [ ] Sensitive configuration values use environment variables with `IDENTITY_BROKER_` prefix
- [ ] New adapter implementations include connection redaction for any credentials
- [ ] Error messages are user-friendly and don't expose internals
- [ ] TLS/SSL support is available for any database connection type
- [ ] Connection pool settings are appropriate for expected load
- [ ] Context timeouts are enforced for all database operations
- [ ] Race condition testing passes: `go test -race ./...`
- [ ] All new tests pass without warnings or errors

## Migration Guide: Adding New Storage Backends

When adding a new storage backend (e.g., MongoDB, Redis):

1. **Create adapter package**: `internal/adapters/storage/{backend_name}/`
2. **Implement interfaces**: `ports.StorageLifecycle` and `ports.UserRepository`
3. **Create domain errors**: Use `storage.ErrorKind` for error classification
4. **Handle credentials**: Implement redaction similar to `ConnectionParameters.Redacted()`
5. **Connection validation**: Add URL/config validation in adapter constructor
6. **Update factory**: Add case to `NewAdapter()` switch statement
7. **Add tests**: Minimum 20+ unit/integration tests
8. **Security review**: Verify no credentials are exposed
9. **Documentation**: Add to configuration examples and troubleshooting guide

## Testing Security

All security tests must pass before committing:

```bash
# Run security-focused tests
go test ./internal/domain/storage/... -run "Redacted|Validation|Security"

# Run all tests with race detection
go test -race ./...

# Generate coverage report
go test -cover ./internal/adapters/storage/...
```

## References

- PostgreSQL Connection Security: https://www.postgresql.org/docs/current/ssl-tcp.html
- OWASP: Sensitive Data Exposure - https://owasp.org/www-project-top-ten/
- Go Context Usage: https://pkg.go.dev/context
