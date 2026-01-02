# Configuration Examples

This directory contains example configuration files for the Agentic Identity Broker. Each file demonstrates different configuration approaches for various deployment scenarios.

## Quick Start

The simplest way to run the application:

```bash
# Use all defaults (log level: info, format: text)
./agentic-identity-broker

# Specify a configuration file
./agentic-identity-broker --config ./examples/config/config.development.yaml

# Override settings via CLI flags
./agentic-identity-broker --config config.yaml --log-level debug --log-format json

# Use environment-specific configuration
GO_ENV=production ./agentic-identity-broker --config ./examples/config/config.production.yaml
```

## Configuration Files

### `config.minimal.yaml`

The most minimal valid configuration. Demonstrates:
- Only required fields (log section with level and format)
- Uses explicit default values
- No environment variable substitution
- Suitable for testing or simple deployments

**Usage:**
```bash
./agentic-identity-broker --config ./examples/config/config.minimal.yaml
```

### `config.development.yaml`

Development environment configuration. Demonstrates:
- Debug logging level for detailed troubleshooting
- Text format for human-readable logs
- Commented-out sections showing future expansion points
- Quick iteration and local testing

**Usage:**
```bash
./agentic-identity-broker --config ./examples/config/config.development.yaml
```

Or combine with environment variables:
```bash
GO_ENV=development ./agentic-identity-broker --config ./examples/config/config.development.yaml
```

### `config.staging.yaml`

Staging/pre-production environment configuration. Demonstrates:
- Environment variable substitution with defaults
- Info logging level for normal operation visibility
- JSON format for structured logging and monitoring
- TLS configuration setup (commented)
- Higher connection pool limits
- Production-like settings but with logging visibility

**Usage:**
```bash
IDENTITY_BROKER_LOG_LEVEL=info IDENTITY_BROKER_LOG_FORMAT=json \
  ./agentic-identity-broker --config ./examples/config/config.staging.yaml
```

### `config.production.yaml`

Production environment configuration. Demonstrates:
- All sensitive values from environment variables (never hardcoded)
- Info logging with JSON format
- TLS mandatory with environment-driven paths
- High connection pool limits
- Monitoring and metrics configuration
- Security-first approach
- Detailed comments on production best practices

**Usage:**
```bash
# Set required environment variables
export IDENTITY_BROKER_LOG_LEVEL=info
export IDENTITY_BROKER_LOG_FORMAT=json
export IDENTITY_BROKER_DB_CONNECTION="postgresql://prod-user:password@prod-db:5432/identity_broker"
export IDENTITY_BROKER_JWT_SECRET="your-production-secret-key"
export IDENTITY_BROKER_TLS_CERT_PATH="/etc/agentic-identity-broker/tls/cert.pem"
export IDENTITY_BROKER_TLS_KEY_PATH="/etc/agentic-identity-broker/tls/key.pem"

./agentic-identity-broker --config ./examples/config/config.production.yaml
```

### `config.yaml.example`

The original template file. Demonstrates:
- Basic YAML structure
- All logging configuration keys
- Environment variable syntax (${VAR_NAME})

**Usage:**
```bash
# Copy as template and customize
cp examples/config/config.yaml.example config.yaml
./agentic-identity-broker --config config.yaml
```

### `third-party-oauth2.yaml`

Third-party OAuth2 session management configuration. Demonstrates:
- JWE signing key configuration for state tokens
- State token TTL configuration (max 15 minutes)
- PKCE code verifier length configuration (32-128 bytes)
- Security-focused comments explaining each setting
- Environment variable substitution for sensitive keys

This configuration is required when enabling OAuth2 session management with third-party
services (GitHub, Google, Microsoft, etc.). It controls how the broker orchestrates
OAuth2 authorization flows and stores encrypted tokens.

**Usage:**
```bash
# Generate JWE signing key
export IDENTITY_BROKER_JWE_SIGNING_KEY="$(openssl rand -base64 32)"

# Run with third-party OAuth2 configuration
./agentic-identity-broker --config ./examples/config/third-party-oauth2.yaml
```

**Security Note:** The JWE signing key MUST be kept secret. It protects OAuth2 state
tokens during the authorization flow. Compromise allows state token forgery and CSRF attacks.

## Configuration Sources and Precedence

The application loads configuration from multiple sources with this precedence (highest to lowest):

1. **CLI Flags** (highest precedence)
   ```bash
   ./agentic-identity-broker --log-level debug --log-format json
   ```

2. **YAML File** (from --config flag or IDENTITY_BROKER_CONFIG_PATH)
   ```bash
   ./agentic-identity-broker --config ./examples/config/config.production.yaml
   ```

3. **.env Files** (environment-specific loading)
   - `.env` (always loaded)
   - `.env.local` (local overrides, not in version control)
   - `.env.{GO_ENV}` (environment-specific, e.g., `.env.production`)
   - `.env.{GO_ENV}.local` (local overrides for environment-specific, not in version control)

   Example with GO_ENV=production:
   ```bash
   GO_ENV=production ./agentic-identity-broker
   ```

4. **Defaults** (lowest precedence)
   - log.level: `info`
   - log.format: `text`

## Environment Variables

### Setting Environment Variables

**Via command line:**
```bash
export IDENTITY_BROKER_LOG_LEVEL=debug
./agentic-identity-broker
```

**Via .env file:**
Create `.env` in the application directory:
```
IDENTITY_BROKER_LOG_LEVEL=debug
IDENTITY_BROKER_LOG_FORMAT=json
IDENTITY_BROKER_API_KEY=your-api-key
```

**Via YAML substitution:**
```yaml
log:
  level: ${IDENTITY_BROKER_LOG_LEVEL}
  format: ${IDENTITY_BROKER_LOG_FORMAT}
```

### Sensitive Values

Values with keys containing `IDENTITY_BROKER_`, `password`, `secret`, `token`, `key`, or `credential` are redacted in logs:

```
Configuration Summary:
  log.level: info [source: CLI]
  log.format: json [source: YAML]
  IDENTITY_BROKER_API_KEY: ***REDACTED*** [source: Environment]
```

## Examples by Deployment Type

### Local Development

```bash
# Option 1: Use defaults only
./agentic-identity-broker

# Option 2: Use development config with debug logging
./agentic-identity-broker --config ./examples/config/config.development.yaml

# Option 3: Mix config file and CLI flags
./agentic-identity-broker --config ./examples/config/config.development.yaml --log-level debug
```

### Docker Container

```dockerfile
FROM golang:1.21 as builder
WORKDIR /app
COPY . .
RUN go build -o agentic-identity-broker ./cmd/agentic-identity-broker

FROM alpine:latest
COPY --from=builder /app/agentic-identity-broker /usr/local/bin/
COPY --from=builder /app/examples/config/config.production.yaml /etc/agentic-identity-broker/config.yaml

ENV IDENTITY_BROKER_LOG_LEVEL=info
ENV IDENTITY_BROKER_LOG_FORMAT=json

CMD ["agentic-identity-broker", "--config", "/etc/agentic-identity-broker/config.yaml"]
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentic-identity-broker-config
data:
  config.yaml: |
    log:
      level: ${IDENTITY_BROKER_LOG_LEVEL:info}
      format: ${IDENTITY_BROKER_LOG_FORMAT:json}

---
apiVersion: v1
kind: Secret
metadata:
  name: agentic-identity-broker-secrets
type: Opaque
stringData:
  IDENTITY_BROKER_JWT_SECRET: "your-secret-key"
  IDENTITY_BROKER_DB_CONNECTION: "postgresql://..."

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentic-identity-broker
spec:
  template:
    spec:
      containers:
      - name: agentic-identity-broker
        image: agentic-identity-broker:latest
        args:
        - --config
        - /etc/agentic-identity-broker/config.yaml
        env:
        - name: IDENTITY_BROKER_LOG_LEVEL
          value: "info"
        - name: IDENTITY_BROKER_LOG_FORMAT
          value: "json"
        - name: IDENTITY_BROKER_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: agentic-identity-broker-secrets
              key: IDENTITY_BROKER_JWT_SECRET
        volumeMounts:
        - name: config
          mountPath: /etc/agentic-identity-broker
      volumes:
      - name: config
        configMap:
          name: agentic-identity-broker-config
```

## Environment-Specific .env Files

The application supports environment-specific .env files. Create them in the application root:

### `.env`
Loaded by all environments. Contains defaults:
```
IDENTITY_BROKER_LOG_LEVEL=info
IDENTITY_BROKER_LOG_FORMAT=text
```

### `.env.local`
Local overrides for development. Not in version control:
```
IDENTITY_BROKER_LOG_LEVEL=debug
IDENTITY_BROKER_LOG_FORMAT=json
```

### `.env.production`
Production-specific settings in version control (no secrets):
```
IDENTITY_BROKER_LOG_LEVEL=info
IDENTITY_BROKER_LOG_FORMAT=json
```

### `.env.production.local`
Local production overrides. Not in version control:
```
# Only for local testing of production config
IDENTITY_BROKER_DB_CONNECTION=postgresql://localhost:5432/test_db
```

## Configuration Validation

The application validates configuration at startup. Invalid values cause a clear error:

```
Error: Configuration error: invalid value "invalid" for field "log.level"
Expected: one of [debug, info, warn, error]
Source: CLI flag --log-level

Please fix the configuration and try again.
```

## Security Best Practices

1. **Never commit sensitive values** to version control
   ```bash
   # ❌ Don't do this
   echo "IDENTITY_BROKER_API_KEY=secret123" >> config.yaml
   git add config.yaml

   # ✅ Do this instead
   echo "IDENTITY_BROKER_API_KEY=${YOUR_API_KEY}" >> config.yaml
   ```

2. **Use .env.*.local files** for local overrides
   ```bash
   # These files are in .gitignore and won't be committed
   cat .env.production.local  # Safe to add secrets here
   ```

3. **Use environment variables in production**
   ```bash
   # Always prefer environment variables in containerized environments
   docker run -e IDENTITY_BROKER_JWT_SECRET=your-secret agentic-identity-broker
   ```

4. **Check file permissions** on configuration files
   ```bash
   # Restrict access to configuration files with secrets
   chmod 600 config.yaml
   chmod 600 .env.production.local
   ```

5. **Audit logs track all sources**
   ```json
   {
     "timestamp": "2025-12-15T10:30:00Z",
     "level": "info",
     "message": "configuration_loaded",
     "sources": [".env", "config.yaml", "cli_flags"],
     "config_keys": ["log.level", "log.format"],
     "redacted_keys": ["IDENTITY_BROKER_JWT_SECRET"]
   }
   ```

## Troubleshooting

### "Required field 'log.level' is missing"

The log level must be set via one of these methods:
```bash
# Via CLI flag
./agentic-identity-broker --log-level info

# Via YAML file
echo "log:\n  level: info" > config.yaml
./agentic-identity-broker --config config.yaml

# Via .env file
echo "LOG_LEVEL=info" > .env
./agentic-identity-broker
```

### "Environment variable 'IDENTITY_BROKER_API_KEY' not set"

Set the environment variable before running:
```bash
export IDENTITY_BROKER_API_KEY=your-api-key
./agentic-identity-broker --config config.yaml
```

### "Permission denied reading 'config.yaml'"

Fix file permissions:
```bash
chmod +r config.yaml
./agentic-identity-broker --config config.yaml
```

### Wrong configuration loaded

Check the startup summary to see which configuration source was used:
```
Configuration Summary:
  log.level: debug [source: CLI]
  log.format: json [source: YAML]
```

The `[source: ...]` annotation shows where each value came from.

## See Also

- [Configuration Guide](../../docs/configuration.md) - Comprehensive documentation
- [ADR 002](../../adrs/002-configuration-libraries.md) - Library selection rationale
- [Architecture](../../ARCHITECTURE.md) - Configuration subsystem architecture
