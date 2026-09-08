# Docker Compose Development Setup

This guide explains how to use Docker Compose to run the complete Agentic Identity Broker stack locally with hot-reload support for both backend and frontend.

## Quick Start (5 minutes)

### Prerequisites

- Docker installed and running, with Docker Compose v2 (`docker compose`)
- `justfile` (included in repo)

### Step 1: Start All Services

```bash
just dev-docker
```

This single command:
1. Creates `.env.compose` from `.env` template (if needed)
2. Builds Docker images for all services
3. Starts all services with hot reload enabled
4. Auto-runs seed data scripts

**Expected output:**
- Services starting log messages
- Logs from all containers streaming to stdout
- "Seed data scripts" completing after broker health check

### Step 2: Access Services

Open in your browser:

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8000
- **Admin API**: http://localhost:14000
- **Upstream OAuth2**: http://localhost:9001
- **Third-Party OAuth2**: http://localhost:9000
- **Sample Agent**: http://localhost:9002

### Step 3: Verify Everything Works

```bash
# Check all services are healthy
just compose-health

# View logs from any service
just compose-logs-backend
just compose-logs-frontend

# Make a code change (modify a Go file)
# → Air rebuilds in ~1-2 seconds
# → Check logs: just compose-logs-backend
```

### Step 4: Stop All Services

Press `Ctrl+C` in the terminal, or:

```bash
just compose-down
```

---

## Services Overview

| Service | Port | Purpose | Notes |
|---------|------|---------|-------|
| `identity-broker` | 8000, 14000 | Main backend API | Hot reload with Air |
| `frontend` | 3000 | React UI (Vite dev server) | HMR enabled |
| `upstream-oauth2` | 9001 | Mock upstream OAuth2 provider | Fixed build |
| `third-party-oauth2` | 9000 | Mock third-party service | Fixed build |
| `sample-agent` | 9002 | Sample OAuth2 client | Fixed build |
| `seed-broker-data` | - | Runs seed script once | Auto-seed agents/services |
| `seed-third-party` | - | Registers mock service once | Auto-seed mock OAuth2 service |

---

## Configuration Files

The project uses separate configuration files for Docker Compose and native development:

### `config.yaml` (Native Development)

Default configuration file for running the broker directly on your host machine using `just run` or `just dev`. Uses `localhost` and `127.0.0.1` addresses for all services:

```yaml
oauth2_authorization_server:
  upstream_issuer_uri: http://127.0.0.1:9001
  upstream_authorize_endpoint: http://127.0.0.1:9001/oauth/authorize
  upstream_token_endpoint: http://127.0.0.1:9001/oauth/token
```

**When to use:** Running the broker natively on your host machine (recommended for debugging, profiling, or when you don't need the full stack).

### `config.docker.yaml` (Docker Compose)

Docker-specific configuration that uses internal container DNS names for service-to-service communication within the Docker Compose network:

```yaml
oauth2_authorization_server:
  upstream_issuer_uri: http://upstream-oauth2:9001
  upstream_authorize_endpoint: http://upstream-oauth2:9001/oauth/authorize
  upstream_token_endpoint: http://upstream-oauth2:9001/oauth/token
```

**When to use:** Running the full stack with `just dev-docker` or `just compose-up`. The docker-compose.yml automatically sets `IDENTITY_BROKER_CONFIG_PATH=config.docker.yaml` to use this file.

### Why Separate Configs?

- **Native development**: `just run` needs `localhost`/`127.0.0.1` addresses since services aren't in a Docker network
- **Docker Compose**: Container-to-container communication uses DNS names like `upstream-oauth2` which only resolve inside the Docker network
- **No manual switching**: The environment automatically picks the right config based on how you run it

---

## Hot Reload Development

### Backend (Go with Air)

When you modify any `.go` file in `./cmd` or `./internal`:

1. Air detects the change
2. Recompiles binary to `tmp/agentic-identity-broker`
3. Restarts the server
4. New behavior available after the server restarts

**View logs:**
```bash
just compose-logs-backend
```

**Manual restart if needed:**
```bash
just compose-restart-backend
```

### Frontend (React with Vite HMR)

When you modify files in `./web/src`:

1. Vite detects the change
2. Browser receives HMR update
3. Component hot-replaces without full page reload
4. Changes visible in ~500ms

**View logs:**
```bash
just compose-logs-frontend
```

**Manual restart if needed:**
```bash
just compose-restart-frontend
```

### Configuration Changes

If you modify `.air.toml` or `vite.config.ts`, restart the affected service:

```bash
just compose-restart-backend    # For Air config changes
just compose-restart-frontend   # For Vite config changes
```

---

## Seed Data

The docker-compose setup automatically seeds sample data on startup:

### Seed Broker Data (seed-broker-data service)

Creates sample agents and services via `tests/fixtures/seed-sample-data.sh`:
- **Weather Assistant** agent
- **Task Manager Pro** agent
- **OAuth2 Test Client** agent
- **Weather API** service
- **Calendar API** service
- **Email API** service

Access via Admin API:
```bash
curl http://localhost:14000/api/agents
```

### Seed Third-Party Service (seed-third-party service)

Registers the mock OAuth2 service via `scripts/register-mock-thirdparty-service.sh`:
- Registers mock OAuth2 service with identity broker
- Creates OAuth2 endpoints for testing
- Registers scopes (profile, email, read, write)

### Manual Seed Execution

If auto-seed fails or you want to reseed:

```bash
# Seed broker data
just compose-seed-data

# Register third-party service
just compose-register-third-party
```

---

## Common Workflows

### View All Logs in Real-Time

```bash
just compose-logs
```

### View Logs from One Service

```bash
just compose-logs-backend
just compose-logs-frontend
just compose-logs-service upstream-oauth2
```

### Restart a Service (after code changes)

```bash
just compose-restart-backend
just compose-restart-frontend
```

### Check Service Health

```bash
just compose-health
```

Shows:
- Running containers and their status
- Health check status for each service
- Connectivity test results

### Run Services in Background

```bash
just compose-up-detached

# Later, stop them:
just compose-down
```

### Clean Everything (containers, volumes, temp files)

```bash
just compose-clean
```

---

## Troubleshooting

### "Address already in use" Error

One or more ports (8000, 3000, 9000, 9001, 9002) are already in use.

**Solution 1: Stop conflicting processes**
```bash
# Find what's using port 8000
lsof -i :8000

# Kill the process (example PID 12345)
kill -9 12345
```

**Solution 2: Use different ports**
Edit `docker-compose.yml` and change port mappings:
```yaml
identity-broker:
  ports:
    - "8001:8000"    # Maps container 8000 to host 8001
    - "14001:14000"
```

### Frontend Can't Reach Backend

Backend API proxy failing with connection errors.

**Check:**
```bash
# 1. Backend is running
curl http://localhost:8000/health

# 2. Inside frontend container, can reach backend
just compose-logs-frontend  # Look for proxy errors

# 3. Check Vite proxy config
# Edit: web/vite.config.ts
# The proxy target should be http://localhost:8000 (host) or http://identity-broker:8000 (container)
```

**Verify proxy is working:**
```bash
curl -s http://localhost:3000/api/health  # Should proxy to backend
```

### Seed Data Didn't Run

Seed services may fail if broker isn't healthy.

**Check seed logs:**
```bash
# In one terminal:
just compose-logs-service seed-broker-data

# In another terminal:
just compose-logs-service seed-third-party
```

**Manual reseed:**
```bash
just compose-seed-data
just compose-register-third-party
```

### File Changes Not Detected

Air uses polling to detect changes on bind-mounted files.

**Expected behavior:**
- Go changes detected in ~1-2 seconds (polling delay)
- React changes detected in ~500ms (Vite HMR is faster)

**Check polling is enabled:**
```bash
grep "poll = " .air.toml
# Should show: poll = true
```

**Manually restart if needed:**
```bash
just compose-restart-backend
```

### Container Startup Too Slow

First startup builds Docker images (~1-2 min). Subsequent starts are faster.

**To speed up:**
- Use SSD for Docker storage
- Allocate more resources in Docker Desktop
- Pre-build images: `docker compose build` before `docker compose up`


### "health check: too many retries"

Services failing health checks - likely startup is slow.

**Check logs:**
```bash
just compose-logs

# Increase startup delay in docker-compose.yml:
# identity-broker:
#   healthcheck:
#     start_period: 20s  # Increase from 10s
```

---

## VS Code Devcontainer

For unified development experience inside a container:

### Setup

1. Install **Remote - Containers** extension in VS Code
2. Open project: `code .`
3. Click **"Reopen in Container"** (notification bottom right)
4. Wait for devcontainer to build (~2 min first time)

### Usage

Inside devcontainer:

```bash
# All commands work the same
just dev-docker     # Start all services
just compose-logs   # View logs
just compose-health # Check health
just verify         # Run the full verification gate
just fmt            # Format code
```

### Port Forwarding

All ports (8000, 3000, 9000-9002) auto-forward to host:

- From host browser: `http://localhost:3000` (works)
- From devcontainer terminal: `http://identity-broker:3000` (inside container network)

### Benefits

- No local setup needed (Go, Node.js, tools all in container)
- Consistent environment across team
- Services defined in `docker-compose.yml` auto-start
- Full terminal access inside container

---

## Environment Variables

### Backend (.env.compose)

```bash
IDENTITY_BROKER_LOG_LEVEL=debug              # Log verbosity
IDENTITY_BROKER_LOG_FORMAT=text               # text or json
IDENTITY_BROKER_JWE_SIGNING_KEY=...          # Encryption key
IDENTITY_BROKER_STATE_TOKEN_TTL=10m          # Token lifetime
IDENTITY_BROKER_STORAGE_BACKEND=memory       # memory or postgres
```

### Frontend (.env.compose)

```bash
VITE_API_URL=http://localhost:8000           # Backend URL
VITE_LOG_LEVEL=debug                          # Vite logging
```

### Seed Scripts (.env.compose)

```bash
ADMIN_API=http://identity-broker:14000/api   # Admin endpoint (container DNS)
BROKER_HEALTH_URL=http://identity-broker:14000/health  # Health check endpoint
MOCK_SERVER_URL=http://third-party-oauth2:9000        # Mock server DNS name
```

---

## Production Docker Image

For production deployment (not for development):

### Build Production Image

```bash
just build-all           # Build Go backend + React frontend
just docker-build-prod   # Create production Docker image
```

### Run Production Image

```bash
just docker-run-prod
# Runs at http://localhost:8000
# Frontend at http://localhost:8000/consent
```

### Key Differences from Dev

- Pre-built binaries (no hot reload)
- Production-optimized (minified, no source maps)
- Alpine Linux base (10x smaller)
- Single container (no compose orchestration)

---

## Tips & Tricks

### Quick Local Testing

```bash
# Full test: build, start, verify health, cleanup
just quick-test
```

### Monitor All Logs

```bash
# Terminal 1: Stream all logs
just compose-logs

# Terminal 2: Work on code
# Services auto-rebuild and logs appear in Terminal 1
```

### Debug a Specific Service

```bash
# Get into running container
docker exec -it aib-broker /bin/sh

# Then run commands directly
curl http://localhost:8000/health
ps aux | grep air
```

### Validate Configuration

```bash
# Check docker-compose.yml syntax
just compose-validate

# Check all environment variables are set
cat .env.compose | grep -v "^#"
```

### Network Isolation Testing

All services are on custom bridge network `aib-network`:

```bash
# List network
docker network inspect aib-network

# Test service DNS resolution
docker exec aib-broker ping upstream-oauth2  # Should work
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Host Machine                             │
│  http://localhost:3000  http://localhost:8000                   │
│            ↓                        ↓                             │
├─────────────────────────────────────────────────────────────────┤
│                      Docker Compose Network                      │
│                        (aib-network)                              │
│                                                                   │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐  │
│  │  Frontend (3000) │  │  Backend (8000)  │  │  Upstream    │  │
│  │  - React + Vite  │  │  - Go + Air       │  │  OAuth2      │  │
│  │  - HMR enabled   │  │  - Hot reload    │  │  (9001)      │  │
│  │                  │  │                  │  │              │  │
│  └──────────────────┘  └──────────────────┘  └──────────────┘  │
│         ↑                        ↑                                │
│         │ Vite proxy            │ API calls                      │
│         └────────────────────────┘                               │
│                                                                   │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐  │
│  │ Third-Party      │  │  Sample Agent    │  │  Seed        │  │
│  │ OAuth2 (9000)    │  │  (9002)          │  │  Services    │  │
│  │                  │  │                  │  │  (one-shot)  │  │
│  └──────────────────┘  └──────────────────┘  └──────────────┘  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘

Key Features:
  ✓ Hot reload: Go changes rebuild in 1-2s (Air)
  ✓ HMR: React changes apply in ~500ms (Vite)
  ✓ Auto-seed: Sample data created on startup
  ✓ Service DNS: Services communicate via names
  ✓ Port mapping: All ports forwarded to host
```

---

## Next Steps

- **Modify code**: Edit `.go` or `.tsx` files → auto-rebuild
- **Check logs**: `just compose-logs` to debug
- **Run tests**: `just verify` (inside devcontainer or on host)
- **Build for prod**: `just docker-build-prod`

Happy developing!
