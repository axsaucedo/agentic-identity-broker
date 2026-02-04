# Upstream OAuth2 Mock Server

This is a mock upstream OAuth2 server used for testing the identity broker's OAuth2 proxy functionality.

## Overview

The Upstream OAuth2 Mock Server simulates a real OAuth2 Authorization Server. It implements:
- OAuth2 authorization endpoint (`/oauth/authorize`)
- OAuth2 token endpoint (`/oauth/token`)
- Health check endpoint (`/health`)
- Consent/authorization screen with distinctive teal styling

## Visual Differentiation

This mock is **visually distinct** from the third-party service mock:
- **Upstream OAuth2 Server**: Teal/cyan background (`#00d4aa` to `#00bfa5` gradient)
- **Third-party Service Mock**: Blue background (different gradient)

This makes it easy to see which mock server is being used during manual testing.

## Configuration

The server is configured via `config.yaml`:

```yaml
server:
  port: 9001                    # HTTP port (default: 9001)
  bind: "127.0.0.1"            # Bind address

oauth2:
  client_id: "upstream-oauth2-client"
  client_secret: "upstream-oauth2-secret-xyz"
  access_token_ttl: 3600s
  refresh_token_ttl: 604800s
  scopes:
    - name: "openid"
      description: "OpenID Connect scope"
    # ... additional scopes

mock_user:
  sub: "upstream-user@example.com"
  name: "Upstream Test User"
  email: "upstream-user@example.com"
```

## Running the Server

### From Project Root

```bash
# Build and run the upstream OAuth2 mock server
go run ./mocks/upstream-oauth2-server/cmd/mock-upstream-oauth2-server/main.go
```

### From the Mock Directory

```bash
cd mocks/upstream-oauth2-server
go run ./cmd/mock-upstream-oauth2-server/main.go
```

The server will start on `http://127.0.0.1:9001`.

## Testing the Mock

### Health Check

```bash
curl http://127.0.0.1:9001/health
# Response: {"status": "ok", "service": "upstream-oauth2-server"}
```

### Authorization Flow

1. Navigate to the authorization endpoint:
   ```
   http://127.0.0.1:9001/oauth/authorize?client_id=upstream-oauth2-client&response_type=code&redirect_uri=http://localhost:3000/callback&scope=openid+profile+email&state=abc123
   ```

2. You'll see the consent screen with **teal background** (distinguishing it from third-party mock)

3. Click "Approve" to grant authorization and receive an authorization code

### Manual Testing with Identity Broker

To test the identity broker's OAuth2 proxy functionality:

1. **Terminal 1**: Start the upstream OAuth2 server
   ```bash
   go run ./mocks/upstream-oauth2-server/cmd/mock-upstream-oauth2-server/main.go
   ```

2. **Terminal 2**: Start the identity broker
   ```bash
   just dev
   ```

3. **Terminal 3**: Start the consent UI (if needed)
   ```bash
   just web-dev
   ```

4. Test the authorization flow through the identity broker's OAuth2 endpoints

## Integration with Identity Broker

Configure the identity broker's `config.yaml` to use this mock as the upstream OAuth2 server:

```yaml
third_party_oauth2:
  upstream_issuer_uri: http://127.0.0.1:9001
  upstream_authorize_endpoint: http://127.0.0.1:9001/oauth/authorize
  upstream_token_endpoint: http://127.0.0.1:9001/oauth/token
  # ... other config
```

## Endpoints

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/health` | GET | Health check |
| `/oauth/authorize` | GET | Authorization endpoint (shows consent screen) |
| `/oauth/authorize` | POST | Consent form submission |
| `/oauth/token` | POST | Token endpoint (server-to-server) |

## Logs

The server outputs structured JSON logs to stdout, including:
- Authorization requests (with full PKCE parameters)
- Token requests (with basic auth details)
- Consent submissions (with approval/denial status)

Monitor logs to debug OAuth2 flow issues.

## Debugging

To see detailed logs including token responses:

```bash
go run ./mocks/upstream-oauth2-server/cmd/mock-upstream-oauth2-server/main.go 2>&1 | jq .
```

Or grep for specific events:

```bash
# Find token requests
go run ./mocks/upstream-oauth2-server/cmd/mock-upstream-oauth2-server/main.go 2>&1 | grep "token_request"
```

## Differences from Third-Party Service Mock

| Aspect | Upstream OAuth2 | Third-Party Service |
|--------|-----------------|-------------------|
| Port | 9001 | 9000 |
| Background Color | Teal/cyan gradient | Blue gradient |
| Client ID | `upstream-oauth2-client` | `mock-oauth2-client-dev` |
| User Subject | `upstream-user@example.com` | `testuser@example.com` |
| Use Case | Tests broker's upstream OAuth2 proxy | Tests third-party OAuth2 sessions |
| Badge | 🌐 UPSTREAM OAUTH2 | (None/different) |

This differentiation makes it immediately clear which OAuth2 server you're interacting with during testing.
