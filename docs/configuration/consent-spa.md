# Consent SPA Configuration

This document describes the configuration options for the consent Single Page Application (SPA) and its integration with the Go backend.

## Table of Contents

1. [Backend Configuration](#backend-configuration)
2. [Frontend Configuration](#frontend-configuration)
3. [Environment-Specific Configuration](#environment-specific-configuration)
4. [CORS Configuration](#cors-configuration)
5. [Session and Authentication](#session-and-authentication)
6. [Examples](#examples)

## Backend Configuration

The backend always serves the root-mounted SPA from `web/dist`, relative to the process working directory. It has no SPA-specific environment variables, YAML keys, or CLI flags.

Deployments must build the frontend and make `web/dist` available relative to the broker process. The browser routes are defined in [ADR 035](../../adrs/035-root-mounted-spa.md).

### Server Configuration

The SPA is served on the enduser server port (default: 8080).

**Environment Variables:**
- `ENDUSER_SERVER_PORT`: Port for enduser server (default: 8080)
- `ENDUSER_SERVER_HOST`: Host for enduser server (default: 0.0.0.0)

**YAML:**
```yaml
servers:
  enduser:
    port: 8080
    host: 0.0.0.0
    read_timeout: 30s
    write_timeout: 30s
    idle_timeout: 120s
```

## Frontend Configuration

The React SPA is configured through Vite and environment variables. Configuration is embedded at build time.

### API Base URL

**Environment Variable:** `VITE_API_BASE_URL`
**Default:** `/api`
**Type:** string (URL path)

**Description:** Base URL for API requests. Should be relative to the same origin.

**Usage in Code:**
```typescript
// src/services/api/client.ts
const client = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  // ...
});
```

**Example:**
```bash
# .env file (development)
VITE_API_BASE_URL=/api

# Build with custom API base URL
VITE_API_BASE_URL=/api/v2 npm run build
```

### Development Server Port

**Environment Variable:** `VITE_DEV_SERVER_PORT`
**Default:** `3000`
**Type:** number

**Description:** Port for the Vite development server. Only used during local development.

**Example:**
```bash
# .env file
VITE_DEV_SERVER_PORT=3001

# Or inline
VITE_DEV_SERVER_PORT=3001 npm run dev
```

### Build Configuration

Build settings are defined in `vite.config.ts`:

```typescript
export default defineConfig({
  base: '/',                  // Root SPA path
  build: {
    outDir: 'dist',           // SPA build output
    sourcemap: false,         // Disable in production
    minify: 'terser',         // Minification strategy
  },
  server: {
    port: 3000,               // Dev server port
    proxy: {                  // Proxy API requests to backend
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

## Environment-Specific Configuration

### Development

**Backend:**
```yaml
# config.dev.yaml

servers:
  enduser:
    port: 8080
    host: localhost
```

**Frontend:**
```bash
# web/.env.development
VITE_API_BASE_URL=/api
VITE_DEV_SERVER_PORT=3000
```

**Development Workflow:**
```bash
# Terminal 1: Backend
just dev

# Terminal 2: Frontend (with proxy)
cd web && npm run dev
```

Access frontend at: http://localhost:3000 (proxies API to :8080)

### Staging

**Backend:**
```yaml
# config.staging.yaml
servers:
  enduser:
    port: 8080
    host: 0.0.0.0
    read_timeout: 30s
    write_timeout: 30s
```

**Frontend:**
```bash
# Build with production settings
npm run build
```

**Deployment:**
```bash
# Build frontend
cd web && npm run build

# Build backend
just build-release

# Deploy binary with the built SPA files
./bin/agentic-identity-broker --config=config.staging.yaml
```

Access SPA at: https://staging.example.com/

### Production

**Backend:**
```yaml
# config.production.yaml
servers:
  enduser:
    port: 8080
    host: 0.0.0.0
    read_timeout: 60s
    write_timeout: 60s
    idle_timeout: 300s

log:
  level: info
  format: json
```

**Deployment:**
```bash
# Build frontend
cd web && npm run build

# Build optimized backend
just build-release

# Deploy
./bin/agentic-identity-broker --config=config.production.yaml
```

Access SPA at: https://agentic-identity-broker.example.com/

## CORS Configuration

CORS is configured on the backend for API requests.

**Environment Variables:**
- `CORS_ALLOWED_ORIGINS`: Comma-separated list of allowed origins
- `CORS_ALLOW_CREDENTIALS`: Allow credentials (cookies, auth headers)
- `CORS_MAX_AGE`: Preflight cache duration (seconds)

**YAML:**
```yaml
cors:
  allowed_origins:
    - https://agentic-identity-broker.example.com
    - https://staging.example.com
  allow_credentials: true
  max_age: 3600
```

**Development:**
```yaml
cors:
  allowed_origins:
    - http://localhost:3000  # Vite dev server
    - http://localhost:8080  # Go backend
  allow_credentials: true
  max_age: 3600
```

**Production (Same-Origin):**
When the SPA is served from the same origin as the API, CORS is not required:
```yaml
cors:
  allowed_origins: []  # Empty = same-origin only
```

## Session and Authentication

### Principal Header Configuration

**Environment Variable:** `ENDUSER_SERVER_AUTH_PRINCIPAL_HEADER_NAME`
**YAML Key:** `servers.enduser.authentication.preauth.principal_header_name`
**Default:** `X-Remote-User`

**Description:** HTTP header containing the authenticated user principal (set by reverse proxy).

**Example:**
```yaml
servers:
  enduser:
    authentication:
      preauth:
        enabled: true
        principal_header_name: X-Remote-User
```

**Common Header Names:**
- `X-Remote-User` (default)
- `X-Authenticated-User`
- `X-Forwarded-User`
- `X-Auth-Request-User` (oauth2-proxy)

### CSRF Token Configuration

CSRF tokens are managed automatically by the backend. Configuration:

**Token Properties:**
- **Header Name:** `X-CSRF-Token` (hardcoded)
- **Cookie Name:** `csrf_token` (hardcoded)
- **Token Length:** 32 bytes (base64-encoded)
- **TTL:** 24 hours
- **Storage:** In-memory (keyed by principal)

**Cookie Settings:**
- **HttpOnly:** false (JavaScript needs to read it)
- **Secure:** true (HTTPS only in production)
- **SameSite:** Strict
- **Path:** /

**No Configuration Required:** CSRF is enabled automatically for all mutating requests.

### Session Management

The system uses **stateless session management** with the principal from the reverse proxy:

1. User authenticates with reverse proxy
2. Proxy sets `X-Principal` header
3. Go backend extracts principal from header
4. Principal is used as session identifier

**No Backend Session Configuration:** Sessions are managed by the reverse proxy.

## Examples

### Example 1: Default Configuration (Development)

**Backend (config.yaml):**
```yaml
servers:
  enduser:
    port: 8080
    authentication:
      preauth:
        enabled: true
        principal_header_name: X-Remote-User
```

**Frontend (.env):**
```bash
VITE_API_BASE_URL=/api
```

**Start:**
```bash
# Build frontend
cd web && npm run build

# Start backend
just run
```

**Access:** http://localhost:8080/

### Example 2: Separate Dev Servers with Proxy

**Backend (running on :8080):**
```bash
just dev
```

**Frontend (vite.config.ts):**
```typescript
export default defineConfig({
  server: {
    port: 3000,
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
```

**Start:**
```bash
# Terminal 1: Backend
just dev

# Terminal 2: Frontend
cd web && npm run dev
```

**Access:** http://localhost:3000 (frontend) → proxies API to :8080 (backend)

### Example 3: Docker Deployment

**Dockerfile:**
```dockerfile
FROM node:24-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
COPY --from=frontend-builder /app/web/dist ./web/dist
RUN go build -o agentic-identity-broker ./cmd/agentic-identity-broker

FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /app/agentic-identity-broker .
COPY --from=backend-builder /app/web/dist ./web/dist
COPY config.production.yaml ./config.yaml

EXPOSE 8080
CMD ["./agentic-identity-broker", "--config=config.yaml"]
```

**config.production.yaml:**
```yaml
servers:
  enduser:
    port: 8080
    host: 0.0.0.0
```

### Example 4: Kubernetes Deployment

**ConfigMap:**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentic-identity-broker-config
data:
  config.yaml: |
    servers:
      enduser:
        port: 8080
        authentication:
          preauth:
            enabled: true
            principal_header_name: X-Auth-Request-User

```
**Deployment:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentic-identity-broker
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: agentic-identity-broker
        image: agentic-identity-broker:latest
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
      volumes:
      - name: config
        configMap:
          name: agentic-identity-broker-config
```

**Ingress (with oauth2-proxy):**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: agentic-identity-broker
  annotations:
    nginx.ingress.kubernetes.io/auth-url: "https://oauth2-proxy.example.com/oauth2/auth"
    nginx.ingress.kubernetes.io/auth-signin: "https://oauth2-proxy.example.com/oauth2/start"
    nginx.ingress.kubernetes.io/auth-response-headers: "X-Auth-Request-User,X-Auth-Request-Email"
spec:
  rules:
  - host: agentic-identity-broker.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: agentic-identity-broker
            port:
              number: 8080
```

### Example 5: CDN Deployment (Separate SPA and API)

**Backend Configuration:**
```yaml
# Configure CORS for CDN origin
cors:
  allowed_origins:
    - https://cdn.example.com
  allow_credentials: true
  max_age: 3600

servers:
  enduser:
    port: 8080
```

**Frontend Build:**
```bash
# Build with production API URL
VITE_API_BASE_URL=https://api.example.com/api npm run build

# Upload dist/ to CDN
aws s3 sync web/dist/ s3://my-cdn-bucket/
```

**Access:**
- SPA: https://cdn.example.com/
- API: https://api.example.com/api/

**Note:** Requires CORS configuration and careful handling of authentication cookies.

## Troubleshooting

### SPA Not Loading

**Issue:** 404 error when accessing `/`

**Solutions:**
1. Ensure frontend is built: `cd web && npm run build`
2. Verify the deployment contains `web/dist/index.html`
3. Ensure `web/dist` is relative to the broker process working directory
4. Check backend logs for file serving errors

### API Requests Failing

**Issue:** CORS errors or 404 on API calls

**Solutions:**
1. Verify API base URL: `VITE_API_BASE_URL=/api`
2. Check backend is running on expected port
3. Verify CORS configuration includes frontend origin
4. In development, ensure Vite proxy is configured

### CSRF Token Missing

**Issue:** 403 Forbidden on POST/PUT/DELETE requests

**Solutions:**
1. Ensure GET request is made first (to get token)
2. Check `X-CSRF-Token` header is included
3. Verify cookie `csrf_token` is set and sent
4. Check SameSite cookie settings
5. Verify principal is present in context

### Principal Not Found

**Issue:** 401 Unauthorized on protected endpoints

**Solutions:**
1. Verify reverse proxy is setting `X-Principal` header
2. Check header name matches configuration
3. Test with curl: `curl -H "X-Principal: test@example.com"`
4. Verify authentication middleware is enabled
5. Check backend logs for principal extraction errors

## Related Documentation

- Architecture documentation: see `ARCHITECTURE.md` in the repository root
- Consent frontend development guide: see `README.md` in the repository root
- [API Documentation](../api/consent-endpoints.md) - API reference
- [Configuration Guide](../configuration.md) - General configuration
