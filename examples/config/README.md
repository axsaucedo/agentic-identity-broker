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

### `jwt-preauth.yaml`

JWT Pre-Authentication configuration. Demonstrates:
- Signed JWT validation with JWKS endpoint (production recommended)
- Unsigned JWT support for service mesh environments (`verification: none`)
- CEL expressions for principal and profile attribute extraction (display name, email, picture URL)
- Backward-compatible configuration with plain-header fallback
- Mutual exclusivity enforcement (`verification: none` + `jwks_uri` → startup error)

**Usage:**
```bash
# Include jwt section in your main configuration file under server.enduser.authentication
# See jwt-preauth.yaml for complete examples of signed, unsigned, and fallback configurations
server:
  enduser:
    authentication:
      jwt:
        jwks_uri: https://auth.example.com/.well-known/jwks.json
        claim_extraction:
          principal_expression: "claims.sub"
          email_expression: "claims.email"
```

**Key Features:**
- Cryptographic JWT signature verification via JWKS (default mode)
- Optional unsigned JWT support for trusted environments (explicit opt-in)
- CEL-based claim extraction for principal, display name, email, and picture URL
- Enriched `/api/me` response with profile attributes
- Fail-closed security: invalid JWTs always rejected, no silent fallback
- Bearer token auto-detection for Authorization header

### `token-exchange.yaml`

RFC 8693 OAuth 2.0 Token Exchange configuration. Demonstrates:
- CEL expressions for claim extraction from JWT tokens (principal and agent ID)
- Gateway authorization rules using CEL expressions
- Automatic token refresh configuration
- Security-first design with mandatory JWT validation
- Support for both minimal and complex authorization scenarios

**Usage:**
```bash
# Include token_exchange section in your main configuration file
./agentic-identity-broker --config ./examples/config/config.yaml

# Where config.yaml includes token_exchange section:
token_exchange:
  claim_extraction:
    principal_expression: "subject_token.sub"
    agent_client_id_expression: "subject_token.azp"
  authorization:
    type: "cel"
    cel:
      expression: 'true'  # or more complex authorization rules
  refresh:
    enabled: true
```

**Key Features:**
- RFC 8693 compliant token exchange endpoint
- JWT validation against upstream OAuth2 JWKS (mandatory, no bypass)
- CEL-based gateway authorization policies
- Resource-based service discovery via protected_resources
- Automatic token refresh with configurable behavior
- Support for custom claim extraction expressions
- Complete audit logging of token exchange events

### `telemetry.yaml`

OpenTelemetry observability configuration. Demonstrates:
- Enabling distributed tracing, metrics, and log correlation
- Service name and resource attribute configuration
- Trace sampling rate and propagator selection
- Metrics export interval configuration
- OTLP exporter protocol (gRPC or HTTP), endpoint, headers with `${ENV_VAR}` substitution
- TLS configuration (enabled by default; `insecure: false` is the secure default)

Telemetry is **disabled by default** — no OTel SDK code runs when `enabled: false`, incurring zero overhead. Enable only when an OTLP-compatible collector is available.

**Usage:**
```bash
# Set the OTLP exporter auth token (if required by your collector)
export OTEL_EXPORTER_AUTH_TOKEN="your-collector-token"

# Enable telemetry by including this section in your configuration
./agentic-identity-broker --config ./examples/config/telemetry.yaml
```

**Key Features:**
- Zero overhead when disabled (no-op provider, no SDK initialization)
- Configurable sampling rate (default: 1.0 for development, 0.1 recommended for production)
- gRPC (default) or HTTP OTLP exporter protocol
- TLS enabled by default (`insecure: false`) — explicit opt-in required to disable
- Environment variable substitution for sensitive headers (e.g., auth tokens)
- Compatible with any OTLP-capable collector (Grafana Agent, OpenTelemetry Collector, Datadog, etc.)

**Security Note:** Never set `insecure: true` in production. This disables TLS for telemetry export, exposing trace data and metric data in transit. The broker will emit a WARN log if insecure mode is enabled at startup.

See [docs/configuration.md](../../docs/configuration.md) for the full list of configuration parameters and environment variable mappings.

### `oauth2-authorization-server.yaml`

OAuth2 Authorization Server proxy configuration. Demonstrates:
- Upstream OAuth2 server URLs (authorization, token endpoints)
- Supported OAuth2 grant types and response types
- Upstream request timeout configuration
- TLS certificate validation (enforced by default)
- **Note**: The broker's public URL used for RFC 8414 metadata/issuer is configured via `server.enduser.public_url` (not inside the `oauth2_authorization_server` block)

**Usage:**
```bash
# Include in your main configuration file
./identity-broker --config ./examples/config/config.yaml

# Where config.yaml references oauth2-authorization-server section:
# oauth2_authorization_server:
#   upstream_issuer_uri: "https://auth.example.com"
#   upstream_authorize_endpoint: "https://auth.example.com/authorize"
#   upstream_token_endpoint: "https://auth.example.com/token"
#
# The broker's public URL (used as the OAuth2 issuer) is set separately:
# server:
#   enduser:
#     public_url: "https://identity-broker.example.com"
```

**Key Features:**
- RFC 6749 compliant OAuth2 Authorization Code flow
- RFC 8414 OAuth2 metadata discovery (/.well-known/oauth-authorization-server)
- Integration with broker's consent UI (Feature 007)
- Agent registry validation (Feature 006)
- TLS certificate validation enforced (no self-signed certificates)
- Audit logging for authorization requests

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

## Observability

The broker supports configurable OpenTelemetry (OTel) distributed tracing, metrics, and log correlation. Telemetry is **disabled by default** with zero overhead.

### Quick Start (Development)

```bash
# Start a local OpenTelemetry Collector (e.g., via Docker)
docker run --rm -p 4317:4317 otel/opentelemetry-collector-contrib

# Enable telemetry with gRPC exporter and insecure connection for local dev
./agentic-identity-broker --config ./examples/config/telemetry.yaml
```

### Configuration Reference

See [`telemetry.yaml`](telemetry.yaml) for a full production-ready example covering:
- `telemetry.enabled` — master switch (default: `false`)
- `telemetry.service_name` — service name on all spans and metrics
- `telemetry.resource_attributes` — custom key-value pairs added to all telemetry
- `telemetry.traces.sampling_rate` — fraction of traces to export (0.0–1.0)
- `telemetry.traces.propagators` — trace context propagation formats (`tracecontext`, `baggage`)
- `telemetry.metrics.export_interval` — how often to push metrics to the collector
- `telemetry.exporter.protocol` — `grpc` (default) or `http`
- `telemetry.exporter.endpoint` — OTLP collector address (`host:port` for gRPC, `https://host:port` for HTTP)
- `telemetry.exporter.headers` — additional headers (e.g., auth tokens via `${ENV_VAR}`)
- `telemetry.exporter.timeout` — per-export timeout
- `telemetry.exporter.insecure` — disable TLS (default: `false`, **never use in production**)

### Environment Variable Mapping

| YAML Key | Environment Variable |
|---|---|
| `telemetry.enabled` | `IDENTITY_BROKER_TELEMETRY_ENABLED` |
| `telemetry.service_name` | `IDENTITY_BROKER_TELEMETRY_SERVICE_NAME` |
| `telemetry.exporter.endpoint` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_ENDPOINT` |
| `telemetry.exporter.protocol` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_PROTOCOL` |
| `telemetry.exporter.insecure` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_INSECURE` |

See [docs/configuration.md](../../docs/configuration.md) for the complete list.

### ADR Reference

[ADR 011](../../adrs/011-opentelemetry-provider-pattern.md) documents the app-layer OTel provider pattern, `otelchi` library choice, and context-based span propagation approach.

## See Also

- [Configuration Guide](../../docs/configuration.md) - Comprehensive documentation
- [ADR 002](../../adrs/002-configuration-libraries.md) - Library selection rationale
- [ADR 011](../../adrs/011-opentelemetry-provider-pattern.md) - OpenTelemetry provider pattern
- [Architecture](../../ARCHITECTURE.md) - Configuration subsystem architecture
