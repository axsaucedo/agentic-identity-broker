# API Documentation

## Health Check Endpoints

The Agentic Identity Broker provides independent health check endpoints on both the end-user and admin servers.

### Overview

Each server exposes a `/health` endpoint that returns the current operational status of that specific server. The servers operate independently, so each endpoint only reports the health of its own server instance.

### OpenAPI Specification

Complete API specification is available in: [specs/003-dual-port-server/contracts/health-api.yaml](../specs/003-dual-port-server/contracts/health-api.yaml)

### Endpoints

#### GET /health (End-User Server)

**URL**: `http://localhost:8000/health` (default port)

Returns the health status of the end-user server.

**Response Codes**:
- `200 OK`: Server is healthy and accepting requests
- `503 Service Unavailable`: Server is starting, shutting down, or unhealthy

**Response Body**:
```json
{
  "status": "healthy",
  "server": "enduser",
  "timestamp": "2025-12-15T10:30:00Z",
  "uptime_seconds": 3600
}
```

#### GET /health (Admin Server)

**URL**: `http://localhost:14000/health` (default port)

Returns the health status of the admin server.

**Response Codes**:
- `200 OK`: Server is healthy and accepting requests
- `503 Service Unavailable`: Server is starting, shutting down, or unhealthy

**Response Body**:
```json
{
  "status": "healthy",
  "server": "admin",
  "timestamp": "2025-12-15T10:30:00Z",
  "uptime_seconds": 3600
}
```

### Health States

| Status | HTTP Code | Description |
|--------|-----------|-------------|
| `healthy` | 200 | Server is operational and processing requests normally |
| `starting` | 503 | Server is initializing (binding to port, setting up routes) |
| `shutting_down` | 503 | Graceful shutdown in progress, completing in-flight requests |
| `unhealthy` | 503 | Server encountered an error and requires restart |

### Response Fields

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Current health state (see table above) |
| `server` | string | Server identifier: `"enduser"` or `"admin"` |
| `timestamp` | string | ISO 8601 timestamp of the health check (UTC) |
| `uptime_seconds` | integer | Number of seconds since the server started |

### Examples

#### Healthy Server

Request:
```bash
curl http://localhost:8000/health
```

Response (200 OK):
```json
{
  "status": "healthy",
  "server": "enduser",
  "timestamp": "2025-12-15T10:30:00Z",
  "uptime_seconds": 3600
}
```

#### Server Shutting Down

Request:
```bash
curl http://localhost:14000/health
```

Response (503 Service Unavailable):
```json
{
  "status": "shutting_down",
  "server": "admin",
  "timestamp": "2025-12-15T11:30:00Z",
  "uptime_seconds": 7200
}
```

### Integration with Load Balancers

The health endpoints are designed for integration with load balancers and monitoring systems:

- **Kubernetes**: Use as liveness and readiness probes
- **HAProxy**: Use as HTTP health check with 200 status code requirement
- **AWS ALB/NLB**: Use as target health check endpoint
- **Prometheus**: Can be scraped for uptime metrics

### Notes

- Health checks are fast (<10ms) and do not perform complex dependency checks
- Each server independently manages its own health state
- Health state transitions are thread-safe using atomic operations
- During graceful shutdown, health checks return 503 to signal load balancers to stop sending traffic
