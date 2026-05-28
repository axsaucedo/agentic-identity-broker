# Configuration Guide

This guide explains how to configure the Agentic Identity Broker for different deployment environments.

## Table of Contents

- [Overview](#overview)
- [Configuration Sources](#configuration-sources)
- [Precedence Rules](#precedence-rules)
- [Environment-Specific Configuration](#environment-specific-configuration)
- [YAML Configuration](#yaml-configuration)
- [Command-Line Flags](#command-line-flags)
- [Configuration Reference](#configuration-reference)
- [Available Settings](#available-settings)
- [Security Best Practices](#security-best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

The Identity Broker supports multiple configuration sources with clear precedence rules. You can combine .env files, YAML configuration, and command-line flags to achieve flexible, environment-specific configuration without code changes.

### Key Features

- **Multiple Sources**: .env files, YAML, CLI flags
- **Environment Variable Substitution**: Use `${VAR_NAME}` in YAML files
- **Environment-Specific**: Automatic loading of .env.{environment} files
- **Secure by Default**: Sensitive values automatically redacted in logs
- **Clear Validation**: Helpful error messages with fix instructions
- **Audit Logging**: JSON audit log of all configuration sources

## Configuration Sources

The Identity Broker loads configuration from four sources (in order of precedence):

### 1. Built-in Defaults

Default values applied if no other source provides a value:
- `log.level`: `info`
- `log.format`: `text`

### 2. .env Files

Environment-specific files loaded automatically based on the `GO_ENV` environment variable:

1. `.env` - Base configuration (committed to git as .env.example)
2. `.env.local` - Local overrides (gitignored)
3. `.env.{GO_ENV}` - Environment-specific (e.g., .env.production)
4. `.env.{GO_ENV}.local` - Environment-specific local overrides (gitignored)

**Example** `.env`:
```bash
# Base configuration
IDENTITY_BROKER_LOG_LEVEL=info
IDENTITY_BROKER_LOG_FORMAT=text
```

**Example** `.env.production`:
```bash
# Production overrides
IDENTITY_BROKER_LOG_LEVEL=warn
IDENTITY_BROKER_LOG_FORMAT=json
```

### 3. YAML Configuration File

Optional YAML file for structured configuration. Supports environment variable substitution using `${VAR_NAME}` syntax.

**File Location** (in order of precedence):
1. Path from `--config` CLI flag
2. Path from `IDENTITY_BROKER_CONFIG_PATH` environment variable
3. `config.yaml` in current directory

**Example** `config.yaml`:
```yaml
log:
  level: ${IDENTITY_BROKER_LOG_LEVEL}
  format: ${IDENTITY_BROKER_LOG_FORMAT}
```

### 4. Command-Line Flags

Highest precedence - overrides all other sources.

```bash
agentic-identity-broker --log-level debug --log-format json
```

## Precedence Rules

When the same configuration key is provided by multiple sources, the value from the highest-precedence source wins:

```
CLI Flags > YAML > .env Files > Defaults
   (3)       (2)      (1)        (0)
```

**Example**:
- Defaults set `log.level=info`
- `.env` sets `IDENTITY_BROKER_LOG_LEVEL=warn`
- `config.yaml` sets `log.level=${IDENTITY_BROKER_LOG_LEVEL}` (expands to `warn`)
- CLI flag `--log-level=debug` is provided

**Result**: `log.level=debug` (from CLI flag)

## Environment-Specific Configuration

Use the `GO_ENV` environment variable to control which .env files are loaded:

### Development (default)
```bash
# GO_ENV defaults to "development" if not set
agentic-identity-broker

# Loads: .env → .env.local → .env.development → .env.development.local
```

### Production
```bash
GO_ENV=production agentic-identity-broker

# Loads: .env → .env.local → .env.production → .env.production.local
```

### Staging
```bash
GO_ENV=staging agentic-identity-broker

# Loads: .env → .env.local → .env.staging → .env.staging.local
```

## YAML Configuration

### Basic Structure

```yaml
log:
  level: info    # debug, info, warn, error
  format: text   # text, json
```

### Environment Variable Substitution

Reference environment variables using `${VAR_NAME}` syntax:

```yaml
log:
  level: ${IDENTITY_BROKER_LOG_LEVEL}
  format: ${IDENTITY_BROKER_LOG_FORMAT}
```

### Nested Variable References

Environment variables can reference other environment variables:

```bash
# In .env
BASE_LEVEL=info
IDENTITY_BROKER_LOG_LEVEL=${BASE_LEVEL}
```

```yaml
# In config.yaml
log:
  level: ${IDENTITY_BROKER_LOG_LEVEL}  # Expands to "info"
```

**Note**: Circular references are detected and will cause an error. Maximum expansion depth is 10 levels.

### Security Validation

The following patterns are **rejected** for security:
- Command substitution: `$(command)` or backticks
- Shell metacharacters: `;`, `|`, `&`, `>`, `<`
- These protections prevent command injection attacks

## Command-Line Flags

### Available Flags

```bash
agentic-identity-broker [flags]

Flags:
  -c, --config string       config file path (overrides IDENTITY_BROKER_CONFIG_PATH)
      --log-level string    log level: debug, info, warn, error
      --log-format string   log format: text, json
  -h, --help               help for agentic-identity-broker
```

### Examples

```bash
# Override log level
agentic-identity-broker --log-level debug

# Use custom config file
agentic-identity-broker --config /etc/agentic-identity-broker/config.yaml

# Multiple flags
agentic-identity-broker --log-level debug --log-format json

# Short form for config
agentic-identity-broker -c config.production.yaml --log-level warn
```

## Configuration Reference

This section provides a comprehensive quick-reference table for all configuration options. For detailed explanations and examples, see the [Available Settings](#available-settings) section below.

### Current Configuration Options

#### Logging Configuration

| Option | Type | Default Value | Valid Values | Required? | Environment Variable | CLI Flag | Description |
|--------|------|---------------|--------------|-----------|----------------------|----------|-------------|
| `log.level` | enum | `info` | `debug`, `info`, `warn`, `error` | No | `IDENTITY_BROKER_LOG_LEVEL` | `--log-level` | Sets logging verbosity level. Use `debug` for troubleshooting, `info` for normal operation, `warn` for production. |
| `log.format` | enum | `text` | `text`, `json` | No | `IDENTITY_BROKER_LOG_FORMAT` | `--log-format` | Sets log output format. Use `json` for production and log aggregation systems. |

#### Encryption Configuration

| Option | Type | Default Value | Valid Values | Required? | Environment Variable | CLI Flag | Description |
|--------|------|---------------|--------------|-----------|----------------------|----------|-------------|
| `encryption.key` | string | - | AWS KMS ARN or `${ENV_VAR}` | Yes | `IDENTITY_BROKER_ENCRYPTION_KEY` | N/A | Key Encryption Key (KEK) for OAuth token envelope encryption. Use AWS KMS ARN for production or `${ENCRYPTION_KEK}` for development. Sensitive - redacted in logs. |

**Encryption Configuration Notes:**
- AWS KMS ARN format: `arn:aws:kms:eu-central-1:123456789012:key/key-id` or `arn:aws:kms:eu-central-1:123456789012:alias/alias-name`
- Environment variable reference: `${ENCRYPTION_KEK}` must resolve to a base64-encoded 256-bit AES key
- Application validates KMS key accessibility on startup (AWS KMS mode) or environment variable presence (env var mode)
- Startup fails with clear error if KEK is unavailable or inaccessible
- Tokens encrypted with old KEK remain decryptable after KEK rotation

#### Server Configuration

The Identity Broker runs two independent HTTP servers on separate ports:
- **End-User Server**: Public-facing API for authentication and identity operations (default port 8000)
- **Admin Server**: Internal management API for monitoring and administration (default port 14000)

| Option | Type | Default Value | Valid Values | Required? | Environment Variable | CLI Flag | Description |
|--------|------|---------------|--------------|-----------|----------------------|----------|-------------|
| `server.enduser.port` | integer | `8000` | 1-65535 | No | `IDENTITY_BROKER_SERVER_ENDUSER_PORT` | `--server.enduser.port` | Port for end-user server. Must differ from admin port. |
| `server.enduser.bind` | string | `::` | IPv4/IPv6 address or hostname | No | `IDENTITY_BROKER_SERVER_ENDUSER_BIND` | `--server.enduser.bind` | Bind address for end-user server. Use `::` for dual-stack (IPv6+IPv4), `0.0.0.0` for IPv4 only, or `127.0.0.1` for localhost only. |
| `server.admin.port` | integer | `14000` | 1-65535 | No | `IDENTITY_BROKER_SERVER_ADMIN_PORT` | `--server.admin.port` | Port for admin server. Must differ from end-user port. |
| `server.admin.bind` | string | `::` | IPv4/IPv6 address or hostname | No | `IDENTITY_BROKER_SERVER_ADMIN_BIND` | `--server.admin.bind` | Bind address for admin server. In production, restrict to private network (e.g., `10.0.1.0`) or use firewall rules. |
| `server.shutdown.timeout` | duration | `30s` | 1s-5m | No | `IDENTITY_BROKER_SERVER_SHUTDOWN_TIMEOUT` | `--server.shutdown.timeout` | Maximum time to wait for in-flight requests to complete during graceful shutdown. Use longer timeouts (60s) in production. |

**Server Configuration Notes:**
- Both servers start atomically - if one fails to bind, both are stopped
- Servers run independently after startup - failure of one doesn't affect the other
- Health endpoints are available on both servers at `/health`
- Graceful shutdown waits for in-flight requests to complete (up to timeout)

#### IPv4/IPv6 Dual-Stack Support

The Identity Broker supports flexible network binding:

- **Dual-Stack (default)**: Bind to `::` accepts both IPv6 and IPv4 connections on systems with dual-stack support
- **IPv6 Only**: Bind to `::1` (localhost) or specific IPv6 addresses
- **IPv4 Only**: Bind to `0.0.0.0` (all interfaces) or `127.0.0.1` (localhost) for IPv4-only systems
- **Automatic Fallback**: If IPv6 binding fails, automatically falls back to IPv4 with a warning log

**Example YAML configurations** are provided in `examples/config/`:
- `config.ipv6-only.yaml` - Dual-stack with IPv6 preference
- `config.ipv4-only.yaml` - IPv4-only configuration

#### Graceful Shutdown

The broker implements graceful shutdown to ensure requests complete cleanly:

1. On receiving SIGTERM or SIGINT signal, health status changes to `shutting_down`
2. New requests are rejected with HTTP 503 Service Unavailable
3. In-flight requests are allowed to complete (up to `server.shutdown.timeout`)
4. After timeout, remaining connections are forcefully closed
5. Process exits cleanly

**Example:**
```bash
# Start broker
agentic-identity-broker &

# Graceful shutdown (waits for requests)
kill -TERM $(pgrep agentic-identity-broker)

# Force shutdown (immediate)
kill -KILL $(pgrep agentic-identity-broker)
```

#### Authentication Configuration

Pre-authentication mode allows the Identity Broker to trust authenticated principals from a reverse proxy. The reverse proxy handles user authentication and passes the authenticated user identifier via an HTTP header.

| Option | Type | Default Value | Valid Values | Required? | Environment Variable | Description |
|--------|------|---------------|--------------|-----------|----------------------|-------------|
| `server.enduser.authentication.preauth.principal_header_name` | string | `X-Remote-User` | Any HTTP header name | No | `IDENTITY_BROKER_SERVER_ENDUSER_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME` | HTTP header containing the authenticated user principal for end-user server |
| `server.admin.authentication.preauth.principal_header_name` | string | `X-Remote-User` | Any HTTP header name | No | `IDENTITY_BROKER_SERVER_ADMIN_AUTHENTICATION_PREAUTH_PRINCIPAL_HEADER_NAME` | HTTP header containing the authenticated user principal for admin server |

**Pre-Authentication Configuration Notes:**
- The principal header must be set by a **trusted reverse proxy only** (nginx, Traefik, HAProxy, etc.)
- Never expose the Identity Broker to untrusted networks - always use a reverse proxy for authentication
- The principal value is trimmed of leading/trailing whitespace
- Maximum principal length: 200 characters (longer principals are rejected with 400 Bad Request)
- Empty/missing headers result in 401 Unauthorized on protected routes
- Optional authentication routes allow missing principals

**Example YAML configurations:**

Development (custom header):
```yaml
server:
  enduser:
    port: 8000
    bind: "::"
    authentication:
      preauth:
        principal_header_name: X-Authenticated-User
  admin:
    port: 14000
    bind: "::"
    authentication:
      preauth:
        principal_header_name: X-Remote-User
```

Production (environment variable):
```yaml
server:
  enduser:
    port: 8000
    bind: "::"
    authentication:
      preauth:
        principal_header_name: ${IDENTITY_BROKER_PRINCIPAL_HEADER}
  admin:
    port: 14000
    bind: "10.0.1.0"  # Restrict to private network
    authentication:
      preauth:
        principal_header_name: ${IDENTITY_BROKER_PRINCIPAL_HEADER}
```

Then set the environment variable:
```bash
export IDENTITY_BROKER_PRINCIPAL_HEADER="X-Authenticated-User"
```

**Nginx Reverse Proxy Example:**
```nginx
upstream identity_broker {
    server localhost:8000;
}

server {
    listen 80;
    server_name api.example.com;

    location / {
        auth_request /auth;
        proxy_pass http://identity_broker;
        proxy_set_header X-Remote-User $remote_user;
    }

    location = /auth {
        internal;
        auth_basic "Restricted";
        auth_basic_user_file /etc/nginx/.htpasswd;
        return 200;
    }
}
```

**Notes:**
- All configuration options have built-in defaults and are optional unless marked "Required"
- Environment variables follow the pattern: `IDENTITY_BROKER_{SECTION}_{KEY}` (uppercased)
- CLI flags follow the pattern: `--{section}-{key}` (lowercase with hyphens)
- See [Precedence Rules](#precedence-rules) for how values from different sources are resolved
- For YAML configuration syntax, see [YAML Configuration](#yaml-configuration)

### Future Configuration Options

The following configuration sections are planned for future releases. This table documents the intended structure for extensibility:

| Option | Type | Default Value | Valid Values | Required? | Environment Variable | CLI Flag | Description |
|--------|------|---------------|--------------|-----------|----------------------|----------|-------------|
| `server.host` | string | `0.0.0.0` | Any valid hostname/IP | No | `IDENTITY_BROKER_SERVER_HOST` | `--server-host` | Server bind address. Use `127.0.0.1` for local-only access. |
| `server.port` | integer | `8080` | 1-65535 | No | `IDENTITY_BROKER_SERVER_PORT` | `--server-port` | Server listen port for HTTP requests. |
| `server.tls.enabled` | boolean | `false` | `true`, `false` | No | `IDENTITY_BROKER_SERVER_TLS_ENABLED` | `--server-tls-enabled` | Enable TLS/HTTPS for secure connections. |
| `server.tls.cert_file` | string | - | Valid file path | Yes (if TLS enabled) | `IDENTITY_BROKER_SERVER_TLS_CERT_FILE` | `--server-tls-cert-file` | Path to TLS certificate file (PEM format). |
| `server.tls.key_file` | string | - | Valid file path | Yes (if TLS enabled) | `IDENTITY_BROKER_SERVER_TLS_KEY_FILE` | `--server-tls-key-file` | Path to TLS private key file (PEM format). |
| `database.type` | enum | `postgres` | `postgres`, `mysql`, `sqlite` | No | `IDENTITY_BROKER_DATABASE_TYPE` | `--database-type` | Database backend type for persistent storage. |
| `database.host` | string | `localhost` | Valid hostname/IP | Yes | `IDENTITY_BROKER_DATABASE_HOST` | `--database-host` | Database server hostname or IP address. |
| `database.port` | integer | `5432` | 1-65535 | No | `IDENTITY_BROKER_DATABASE_PORT` | `--database-port` | Database server port (defaults: PostgreSQL 5432, MySQL 3306). |
| `database.name` | string | `identity_broker` | Valid database name | Yes | `IDENTITY_BROKER_DATABASE_NAME` | `--database-name` | Database name to use for broker data. |
| `database.username` | string | - | Valid username | Yes | `IDENTITY_BROKER_DATABASE_USERNAME` | `--database-username` | Database authentication username. |
| `database.password` | string | - | Any string | Yes | `IDENTITY_BROKER_DATABASE_PASSWORD` | N/A | Database authentication password. Sensitive - redacted in logs. |
| `database.ssl_mode` | enum | `prefer` | `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` | No | `IDENTITY_BROKER_DATABASE_SSL_MODE` | `--database-ssl-mode` | SSL/TLS mode for database connections. |
| `database.max_connections` | integer | `25` | 1-1000 | No | `IDENTITY_BROKER_DATABASE_MAX_CONNECTIONS` | `--database-max-connections` | Maximum number of open database connections in pool. |
| `auth.jwt.secret` | string | - | Base64 string (min 32 bytes) | Yes | `IDENTITY_BROKER_AUTH_JWT_SECRET` | N/A | JWT signing secret key. Sensitive - redacted in logs. |
| `auth.jwt.expiry` | duration | `1h` | Valid duration (e.g., `30m`, `2h`) | No | `IDENTITY_BROKER_AUTH_JWT_EXPIRY` | `--auth-jwt-expiry` | JWT token expiration duration. |
| `auth.session.timeout` | duration | `24h` | Valid duration | No | `IDENTITY_BROKER_AUTH_SESSION_TIMEOUT` | `--auth-session-timeout` | User session inactivity timeout. |
| `auth.providers` | list | `[]` | Array of provider configs | Yes | N/A | N/A | List of configured identity providers (OAuth, SAML, etc.). |

**Future Options Notes:**
- Options in this table are for planning and design purposes
- Not yet implemented in the current release
- Structure may change based on requirements and feedback
- Sensitive fields (passwords, secrets, keys) will never have CLI flags for security
- Duration values accept formats like: `30s`, `5m`, `1h`, `24h`

### Naming Conventions

The configuration system follows consistent naming patterns across all sources:

1. **YAML Paths**: Use dot notation with lowercase keys (e.g., `log.level`, `server.tls.enabled`)
2. **Environment Variables**: Prefix + uppercase + underscores (e.g., `IDENTITY_BROKER_LOG_LEVEL`, `IDENTITY_BROKER_SERVER_TLS_ENABLED`)
3. **CLI Flags**: Lowercase with hyphens (e.g., `--log-level`, `--server-tls-enabled`)
4. **Nested Config**: Each level adds a separator (`.` in YAML, `_` in env vars, `-` in flags)

### Configuration by Use Case

**Development Environment:**
```yaml
log:
  level: debug
  format: text
```

**Production Environment:**
```yaml
log:
  level: warn
  format: json
```

**Troubleshooting:**
```bash
# Temporarily override to debug level
agentic-identity-broker --log-level debug
```

For complete configuration examples and detailed explanations, continue to the [Available Settings](#available-settings) section.

## Available Settings

### Logging Configuration

#### log.level

**Description**: Sets the logging verbosity level.

**Valid Values**: `debug`, `info`, `warn`, `error`

**Default**: `info`

**Environment Variable**: `IDENTITY_BROKER_LOG_LEVEL`

**CLI Flag**: `--log-level`

**Examples**:
```bash
# .env file
IDENTITY_BROKER_LOG_LEVEL=debug

# YAML file
log:
  level: warn

# CLI flag
--log-level error
```

#### log.format

**Description**: Sets the log output format.

**Valid Values**: `text`, `json`

**Default**: `text`

**Environment Variable**: `IDENTITY_BROKER_LOG_FORMAT`

**CLI Flag**: `--log-format`

**Examples**:
```bash
# .env file
IDENTITY_BROKER_LOG_FORMAT=json

# YAML file
log:
  format: json

# CLI flag
--log-format json
```

**Recommendation**: Use `json` format in production for structured logging and log aggregation.

### Encryption Configuration

#### encryption.key

**Description**: Specifies the Key Encryption Key (KEK) for envelope encryption of OAuth2 tokens at rest. Supports two modes: AWS KMS for production deployments, or environment variable injection for development.

**Valid Formats**:
- AWS KMS ARN: `arn:aws:kms:region:account:key/key-id` or `arn:aws:kms:region:account:alias/alias-name`
- Base64-encoded AES key: Raw base64-encoded 256-bit (32-byte) key (can be populated from environment variable)

**Default**: None (required for token encryption; startup fails if not provided and encryption is enabled)

**Environment Variable**: `IDENTITY_BROKER_ENCRYPTION_KEY`

**CLI Flag**: None (only configurable via YAML or environment)

**Examples**:

**Production (AWS KMS)**:
```yaml
# config.yaml - Production deployment
encryption:
  key: "arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012"
```

```bash
# .env.production
IDENTITY_BROKER_ENCRYPTION_KEY=arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012
```

**Development (Base64-Encoded Key from Environment Variable)**:
```bash
# .env.local - Development with base64-encoded key in environment variable
ENCRYPTION_KEK=$(openssl rand -base64 32)
IDENTITY_BROKER_ENCRYPTION_KEY="${ENCRYPTION_KEK}"
```

```yaml
# config.yaml - Development deployment (interpolates ENCRYPTION_KEK from environment)
encryption:
  key_encryption_key: "${ENCRYPTION_KEK}"
```

**AWS KMS Setup Instructions**:

1. Create a customer-managed key in AWS KMS:
```bash
aws kms create-key \
  --description "Identity Broker OAuth Token Encryption Key" \
  --region eu-central-1
```

2. Create an alias for easier reference:
```bash
aws kms create-alias \
  --alias-name "alias/identity-broker-encryption" \
  --target-key-id "arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012"
```

3. Grant IAM permissions to the Identity Broker service role:
```bash
aws kms create-grant \
  --key-id "arn:aws:kms:eu-central-1:123456789012:key/12345678-1234-1234-1234-123456789012" \
  --grantee-principal "arn:aws:iam::123456789012:role/IdentityBrokerRole" \
  --operations "Encrypt" "Decrypt" "GenerateDataKey" "DescribeKey"
```

4. Reference the key in configuration:
```yaml
encryption:
  key_encryption_key: "arn:aws:kms:eu-central-1:123456789012:alias/identity-broker-encryption"
```

**Startup Validation**:

- **AWS KMS mode**: On startup, the application validates that the KMS key is accessible and the service has required permissions. Startup fails with clear error message if validation fails.
- **Environment variable mode**: On startup, the application validates that the environment variable is set and contains valid base64-encoded key material. Startup fails if not.

**Security Considerations**:

- **AWS KMS**: KEK never leaves AWS KMS boundaries. The application only sees encrypted Data Encryption Keys (DEKs). All cryptographic operations happen server-side in KMS.
- **Environment Variable**: KEK is loaded into application memory at startup.
- **Rotation**: KMS keys can be rotated without application restart. Old tokens remain decryptable with rotated keys.
- **Permissions**: Ensure service role has `kms:Decrypt` and `kms:GenerateDataKey` permissions. Overly broad permissions should be avoided.

**Recommendations**:

| Environment | Approach | Configuration |
|---|---|---|
| Production | AWS KMS | Use customer-managed key ARN or alias |
| Staging | AWS KMS | Separate key or AWS KMS key per environment |
| Development | Environment Variable | `${ENCRYPTION_KEK}` from .env.local |
| CI/CD | Environment Variable | `${ENCRYPTION_KEK}` from CI/CD secrets |
| Testing | Environment Variable | Random generated key per test |

## Security Best Practices

### Sensitive Values

Any configuration key starting with `IDENTITY_BROKER_` or containing these keywords is considered sensitive and will be redacted in logs:
- `PASSWORD`
- `SECRET`
- `TOKEN`
- `KEY`
- `CREDENTIAL`
- `AUTH`

**Example**:
```bash
IDENTITY_BROKER_API_KEY=secret123
DB_PASSWORD=mypassword
```

Both values will be shown as `***REDACTED***` in startup summary and audit logs.

### Never Commit Secrets

1. Use `.env.local` and `.env.{environment}.local` for local secrets (gitignored)
2. Commit `.env.example` and `.env.production.example` as templates
3. Use environment variables or secret management systems in production

### File Permissions

Configuration files should not be world-readable:

```bash
# Recommended permissions
chmod 600 .env
chmod 600 config.yaml
```

### Environment Variable Injection

The system validates against command injection attempts. The following will be rejected:

```yaml
# REJECTED: Command substitution
log:
  level: $(malicious_command)

# REJECTED: Shell metacharacters
database:
  host: localhost; rm -rf /
```

## Troubleshooting

### Configuration Not Loading

**Symptom**: Application uses default values instead of your configuration.

**Solutions**:
1. Check file locations - .env files must be in the current directory or use absolute paths
2. Verify environment variable names start with `IDENTITY_BROKER_`
3. Check YAML syntax is valid (use `yamllint` or online validator)
4. Use `--help` to verify flag names

### Undefined Environment Variable Error

**Symptom**: Error message: "environment variable 'VAR_NAME' is not set"

**Cause**: YAML file references `${VAR_NAME}` but variable doesn't exist in environment.

**Solutions**:
1. Set the environment variable: `export VAR_NAME=value`
2. Add to .env file: `VAR_NAME=value`
3. Remove the ${} reference from YAML and use a literal value

### Invalid Configuration Value

**Symptom**: Error message: "invalid value 'X' for field 'Y'"

**Cause**: Configuration value doesn't match expected format or enum values.

**Solutions**:
1. Check error message for expected values (e.g., "expected: debug, info, warn, or error")
2. Verify spelling and case (values are case-sensitive)
3. Check for extra whitespace or quotes in values

### Circular Reference Detected

**Symptom**: Error message: "circular reference detected: VAR1 → VAR2 → VAR1"

**Cause**: Environment variables reference each other in a loop.

**Solution**: Break the circular dependency:

```bash
# WRONG
VAR1=${VAR2}
VAR2=${VAR1}

# CORRECT
VAR1=value1
VAR2=${VAR1}
```

### Viewing Effective Configuration

To see which configuration values are being used and from which sources:

```bash
agentic-identity-broker

# Output shows:
# === Configuration Summary ===
#   log.level: debug [source: cli]
#   log.format: json [source: yaml (/path/to/config.yaml)]
#
# === Configuration Sources ===
#   [0] default: defaults (loaded at ...)
#   [1] env_file: /path/to/.env (loaded at ...)
#   [2] yaml: /path/to/config.yaml (loaded at ...)
#   [3] cli: cli (loaded at ...)
```

### Audit Log

For compliance and troubleshooting, check the JSON audit log (first output on startup):

```json
{
  "timestamp": "2025-12-15T09:00:00Z",
  "event": "configuration_loaded",
  "sources": [...],
  "keys": ["log.level", "log.format"],
  "redacted_keys": []
}
```

### OAuth2 Authorization Server Configuration

#### oauth2_authorization_server.multi_agent_client

**Description**: Controls whether multiple agents may share the same upstream OAuth2 `client_id` (stored as `agent.client_id`). When disabled (default), the broker enforces per-agent *upstream* `client_id` uniqueness; incoming OAuth2 requests are always resolved by the agent's internal UUID (`agent.id`), regardless of this setting. When enabled, multiple agents can share one upstream OAuth2 application, and agent identity is additionally verified via a custom claim injected into the upstream authorize redirect and checked in the upstream token response.

**Configuration block** (nested under `oauth2_authorization_server`):

| Option | Type | Default | Valid Values | Required | Description |
|--------|------|---------|--------------|----------|-------------|
| `multi_agent_client.enabled` | boolean | `false` | `true`, `false` | No | Allow multiple agents to share one upstream OAuth2 `client_id`. When `false`, duplicate `client_id` on agent create/update returns `409 Conflict`. |
| `multi_agent_client.agent_id_param_name` | string | — | Any URL-safe string | Yes if `enabled=true` | Query parameter appended to the upstream authorization redirect URL carrying the agent's internal UUID. Must match the claim name expected by your upstream OAuth2 server. |
| `multi_agent_client.agent_id_claim_name` | string | — | Any string | Yes if `enabled=true` | JWT claim in the upstream token response carrying the agent's internal UUID. The broker verifies this claim on every token proxy response; tokens lacking it are rejected (fail closed). |

**Startup validation**: If `enabled = true` and either `agent_id_param_name` or `agent_id_claim_name` is empty, the broker fails to start with a clear error message.

**Feature disabled (default)**:
```yaml
oauth2_authorization_server:
  upstream_issuer_uri: "https://auth.example.com"
  upstream_token_endpoint: "https://auth.example.com/token"
  public_base_url: "https://broker.example.com"

  multi_agent_client:
    enabled: false   # default — each agent must have a unique client_id
```

**Feature enabled**:
```yaml
oauth2_authorization_server:
  upstream_issuer_uri: "https://auth.example.com"
  upstream_token_endpoint: "https://auth.example.com/token"
  public_base_url: "https://broker.example.com"

  multi_agent_client:
    enabled: true
    agent_id_param_name: "x_agent_id"   # injected into authorize redirect as ?x_agent_id=<agent.id>
    agent_id_claim_name: "x_agent_id"   # expected in upstream token response JWT
```

**Token exchange CEL expression update**: When using `multi_agent_client`, you must also update `token_exchange.claim_extraction.agent_id_expression` in your configuration:

| Feature mode | Required CEL expression | Notes |
|---|---|---|
| `enabled = false` (default) | `resolveAgentIdByClientId(subject_token.azp)` | Uses the `resolveAgentIdByClientId` helper to map upstream `client_id` → `agent.id` |
| `enabled = true` | `subject_token.x_agent_id` (use your `agent_id_claim_name`) | Reads the agent UUID directly from the token claim |

> **Breaking change**: The previous expression `subject_token.azp` is no longer valid. Token exchange resolves agents by `agent.id` (UUID), not `agent.client_id`. See [docs/changelog.md](../docs/changelog.md) for migration instructions.

See `examples/config/oauth2-authorization-server.yaml` for a complete configuration example with both modes commented.

### OAuth2 Server Mode Configuration

#### oauth2.auth_server

**Description**: Controls the broker's OAuth2 operating mode. Three symmetric modes are supported:

- **`proxy`**: OAuth2 requests are forwarded to an upstream authorization server. The broker acts as a transparent proxy — it handles consent and delegation, then routes the final authorization to the upstream. No local token issuance; no JWKS endpoint.
- **`local`**: The broker acts as a standalone OAuth2 authorization server, minting its own JWT access tokens signed with managed asymmetric keys. Supports `client_credentials` and `authorization_code` (with PKCE) grant types, and exposes RFC 8414 discovery and JWKS endpoints.
- **`hybrid`**: Both proxy and local paths coexist. Agents are classified by their properties: agents with an upstream `ClientID` are routed to the proxy path; local agents (no `ClientID`, no `client_uris`) and CIMD agents (`client_uris` set) are issued local tokens. Requires both `proxy` and `local` configuration sections.

> **Note**: The mode name `issue_token` (used in earlier versions) is no longer valid. Use `local` instead.

**Configuration block** (nested under `oauth2.auth_server`):

| Option | Type | Default | Valid Values | Required? | Environment Variable | CLI Flag | Description |
|--------|------|---------|--------------|-----------|----------------------|----------|-------------|
| `oauth2.auth_server.mode` | enum | — | `proxy`, `local`, `hybrid` | Yes (if any auth server field is set) | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_MODE` | — | Operating mode. `proxy` forwards to upstream; `local` mints tokens locally; `hybrid` supports both. |
| `oauth2.auth_server.proxy.upstream_issuer_uri` | string | — | Valid HTTPS URI | Yes (if `proxy` or `hybrid`) | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_ISSUER_URI` | — | Upstream OAuth2 issuer URI. Used for proxy path routing. |
| `oauth2.auth_server.proxy.upstream_authorize_endpoint` | string | — | Valid HTTPS URI | Yes (if `proxy` or `hybrid`) | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_AUTHORIZE_ENDPOINT` | — | Upstream authorization endpoint. |
| `oauth2.auth_server.proxy.upstream_token_endpoint` | string | — | Valid HTTPS URI | Yes (if `proxy` or `hybrid`) | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_PROXY_UPSTREAM_TOKEN_ENDPOINT` | — | Upstream token endpoint. |
| `oauth2.auth_server.local.token_ttl` | duration | `1h` | Go duration (e.g. `30m`, `2h`) | No | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_LOCAL_TOKEN_TTL` | — | Validity period for locally issued JWT access tokens. |
| `oauth2.auth_server.local.token_claims_expression` | string | `""` | CEL expression | No | `IDENTITY_BROKER_OAUTH2_AUTH_SERVER_LOCAL_TOKEN_CLAIMS_EXPRESSION` | — | CEL expression to inject custom claims into issued JWTs. |

**Required**: `oauth2_authorization_server.mode` is mandatory — the broker rejects startup when the block is absent or `mode` is empty or invalid. Set `mode` to `proxy`, `local`, or `hybrid` before deploying.

**Startup validation**: The broker validates configuration at startup and rejects incompatible combinations — proxy-only fields in local mode, local-only fields in proxy mode, or missing sections in hybrid mode.

**Proxy mode**:
```yaml
oauth2:
  auth_server:
    mode: "proxy"
    proxy:
      upstream_issuer_uri: "https://auth.example.com"
      upstream_authorize_endpoint: "https://auth.example.com/oauth/authorize"
      upstream_token_endpoint: "https://auth.example.com/oauth/token"
```

**Local mode**:
```yaml
oauth2:
  auth_server:
    mode: "local"
    local:
      token_ttl: "1h"
      token_claims_expression: '{"team": agent.display_name}'
```

**Hybrid mode** (proxy and local agents coexist):
```yaml
oauth2:
  auth_server:
    mode: "hybrid"
    proxy:
      upstream_issuer_uri: "https://auth.example.com"
      upstream_authorize_endpoint: "https://auth.example.com/oauth/authorize"
      upstream_token_endpoint: "https://auth.example.com/oauth/token"
    local:
      token_ttl: "1h"
```

**Security notes**:
- PKCE is always enforced (S256 only) for authorization code grants. No plaintext challenge method.
- Client secrets are hashed with Argon2id and never stored in plaintext.
- Signing key private material is encrypted at rest via `EncryptionPort`.
- Authorization codes are single-use with 60-second TTL, stored as SHA-256 hashes.
- Signing key decryption failure prevents token issuance (fail-closed).
- In `proxy` mode, locally registered agents (no `ClientID`) are rejected. In `local` mode, proxy agents (with `ClientID`) are rejected. Mode boundaries are strict.

See [docs/features/oauth2-server-mode.md](docs/features/oauth2-server-mode.md) for comprehensive end-user documentation.

## Observability / OpenTelemetry

The Identity Broker supports configurable OpenTelemetry (OTel) tracing, metrics, and logging export via OTLP (gRPC or HTTP). When disabled (default), the OTel SDK is never initialized and there is zero overhead.

### Quickstart

See `examples/config/telemetry.yaml` for a complete production-ready example. The minimal configuration to enable tracing:

```yaml
telemetry:
  enabled: true
  service_name: agentic-identity-broker
  exporter:
    protocol: grpc
    endpoint: otel-collector:4317
    insecure: true   # only for non-production environments
```

### OTel Configuration Reference

| Parameter | Type | Default | Environment Variable | Description |
|---|---|---|---|---|
| `telemetry.enabled` | bool | `false` | `IDENTITY_BROKER_TELEMETRY_ENABLED` | Master switch. When false, no OTel SDK is initialized and overhead is zero. |
| `telemetry.service_name` | string | `"agentic-identity-broker"` | `IDENTITY_BROKER_TELEMETRY_SERVICE_NAME` | Value for the `service.name` OTel resource attribute. Appears on all spans, metrics, and logs. |
| `telemetry.resource_attributes` | map[string]string | `{}` | N/A (config file only) | Additional OTel resource attributes added to every telemetry signal (e.g., `deployment.environment: production`). |
| `telemetry.traces.enabled` | bool | `true` (when telemetry.enabled) | `IDENTITY_BROKER_TELEMETRY_TRACES_ENABLED` | Enable trace export. Disabling traces suppresses otelchi HTTP spans and storage child spans. |
| `telemetry.traces.sampling_rate` | float64 | `1.0` | `IDENTITY_BROKER_TELEMETRY_TRACES_SAMPLING_RATE` | Fractional sampling rate for traces (0.0–1.0). `1.0` = 100% sampled. Uses `ParentBased(TraceIDRatioBased(rate))`. |
| `telemetry.traces.propagators` | []string | `["ottrace","b3multi","baggage"]` | N/A (config file only) | Propagators to register globally. Supported values: `ottrace` (OpenTracing interop), `b3multi` (Zipkin B3 multiple headers), `b3` (B3 single header), `tracecontext` (W3C), `baggage` (W3C). Propagators are registered unconditionally when telemetry is enabled, even if `traces.enabled` is false. |
| `telemetry.metrics.enabled` | bool | `true` (when telemetry.enabled) | `IDENTITY_BROKER_TELEMETRY_METRICS_ENABLED` | Enable metrics export. When enabled, process runtime metrics (goroutines, memory, GC) and HTTP request metrics (via otelchi) are exported automatically. |
| `telemetry.metrics.export_interval` | duration | `30s` | `IDENTITY_BROKER_TELEMETRY_METRICS_EXPORT_INTERVAL` | How often metrics are pushed to the collector. Must be positive. Example: `30s`, `1m`. |
| `telemetry.logs.enabled` | bool | `true` (when telemetry.enabled) | `IDENTITY_BROKER_TELEMETRY_LOGS_ENABLED` | Enable OTLP log export via the `slog` bridge. Set to `false` if the collector does not support `opentelemetry.proto.collector.logs.v1.LogsService`. |
| `telemetry.exporter.protocol` | string | `"grpc"` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_PROTOCOL` | OTLP transport protocol. Accepted values: `grpc`, `http`, `https`. The `https` protocol uses the HTTP OTLP exporter with TLS; bare `host:port` endpoints are auto-prefixed with `https://`. |
| `telemetry.exporter.endpoint` | string | `""` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_ENDPOINT` | OTLP collector endpoint. **Required when `telemetry.enabled=true`**. Format: `host:port` for gRPC, `http(s)://host:port` for HTTP/HTTPS. |
| `telemetry.exporter.headers` | map[string]string | `{}` | N/A (config file only) | Additional HTTP/gRPC headers sent with every export request (e.g., authentication tokens). Use `${ENV_VAR}` substitution to avoid committing secrets. |
| `telemetry.exporter.timeout` | duration | `10s` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_TIMEOUT` | Per-export request timeout. Must be positive. Example: `5s`, `30s`. |
| `telemetry.exporter.compression` | string | `"none"` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_COMPRESSION` | Payload compression for all OTLP exporters. Accepted values: `none`, `gzip`. |
| `telemetry.exporter.insecure` | bool | `false` | `IDENTITY_BROKER_TELEMETRY_EXPORTER_INSECURE` | Disable TLS for the OTLP exporter. **Do not use in production** — a startup warning is emitted when this is true. Only meaningful for gRPC; for HTTP the URL scheme controls TLS. |

### Environment Variable Mapping

All OTel settings can be overridden via environment variables using the `IDENTITY_BROKER_TELEMETRY_` prefix, following the same Viper binding rules as other settings. For example:

```bash
IDENTITY_BROKER_TELEMETRY_ENABLED=true
IDENTITY_BROKER_TELEMETRY_EXPORTER_ENDPOINT=otel-collector:4317
IDENTITY_BROKER_TELEMETRY_EXPORTER_INSECURE=true
IDENTITY_BROKER_TELEMETRY_SERVICE_NAME=my-broker-instance
```

### What is Instrumented

When tracing is enabled, the following operations emit child spans:

| Span Name | Operation | Key Attributes |
|---|---|---|
| HTTP span (per route) | All inbound HTTP requests (via otelchi) | `http.method`, `http.route`, `http.status_code` |
| `storage.get.agent` | PostgreSQL `AgentRepository.Get` | `db.system=postgresql`, `db.operation=GetAgent` |
| `storage.list.userGrants` | PostgreSQL `UserGrantRepository.ListByPrincipal` | `db.system=postgresql` |
| `storage.upsert.userGrant` | PostgreSQL `UserGrantRepository.Create` (upsert) | `db.system=postgresql` |
| `storage.get.thirdPartyService` | PostgreSQL `ThirdpartyServiceRepository.Get` | `db.system=postgresql` |
| `encryption.encrypt` | AWS KMS envelope encryption | `encryption.key_type=aws_kms` |
| `encryption.decrypt` | AWS KMS envelope decryption | `encryption.key_type=aws_kms` |
| `jwks.fetch` | JWKS key set fetch/cache lookup | `url.full` |
| `oauth2.token_exchange` | Upstream OAuth2 token proxy | `http.method=POST`, `http.status_code` |

### Security Notes

- `telemetry.exporter.insecure` defaults to `false` (TLS required by default). A startup WARN is emitted if set to true.
- Span attributes never contain tokens, encryption keys, PII, credentials, or request body content.
- `telemetry.exporter.headers` values (e.g., API keys) should reference environment variables via `${VAR_NAME}` syntax to avoid committing secrets to configuration files.

## Getting Help

- Review error messages carefully - they include fix instructions
- Check the startup summary to see which sources were loaded
- Verify file paths are absolute or relative to current directory
- Ensure GO_ENV matches your environment name
- Review [ADR 002](../adrs/002-configuration-libraries.md) for implementation details
- Check [ARCHITECTURE.md](../ARCHITECTURE.md) for configuration subsystem architecture
