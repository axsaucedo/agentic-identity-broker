# Integration Tests

This directory contains integration tests for the persistence layer storage adapters.

## Running Integration Tests

### Prerequisites

Integration tests require either Docker or Podman to run.

#### With Docker
```bash
# Ensure Docker daemon is running
docker ps

# Run integration tests
go test -tags=integration -v ./test/integration/storage/...

# Or using justfile
just test-integration
```

#### With Podman

Podman is fully supported as an alternative to Docker, including on macOS.

**On Linux with rootless Podman:**
```bash
# Podman socket should be available automatically
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
go test -tags=integration -v ./test/integration/storage/...
```

**On macOS with Podman machine:**
```bash
# Start the Podman machine and get the socket path
podman machine start

# The output will show the socket path, typically:
# /var/folders/42/xfyh9ksn6sndqbtl0ybbtr700000gn/T/podman/podman-machine-default-api.sock

# Set DOCKER_HOST environment variable
export DOCKER_HOST='unix:///var/folders/42/xfyh9ksn6sndqbtl0ybbtr700000gn/T/podman/podman-machine-default-api.sock'

# Run tests
go test -tags=integration -v ./test/integration/storage/...

# Or with justfile
just test-integration
```

**Note:** The postgres_test.go init function automatically disables Ryuk cleanup by default. This is necessary because Ryuk tries to use a network named "bridge", which conflicts with Podman's network mode system (where "bridge" is a reserved network mode, not a network name). This is handled transparently - no additional configuration needed.

## Test Organization

### Structure
```
test/integration/storage/
├── README.md                 # This file
├── lifecycle_test.go        # Memory adapter lifecycle tests (no build tag)
└── postgres_test.go         # PostgreSQL adapter tests (integration build tag)
```

### Memory Adapter Tests (No Build Tag)
Run with standard `go test`:
```bash
go test -v ./test/integration/storage/...
```

Tests:
- `TestMemoryAdapter_FullLifecycle` - Complete workflow
- `TestMemoryAdapter_ConcurrentOperations` - 100+ goroutines
- `TestMemoryAdapter_TimeoutHandling` - Context timeout behavior
- `TestMemoryAdapter_PaginationSupport` - Offset/limit pagination
- `TestMemoryAdapter_ErrorRecovery` - Error handling scenarios

### PostgreSQL Adapter Tests (With Integration Build Tag)
Requires container runtime (Docker or Podman):
```bash
go test -tags=integration -v ./test/integration/storage/...
```

Tests:
- `TestPostgresAdapter_Initialize_ConnectionFailed` - Invalid connection handling
- `TestPostgresAdapter_HealthCheck_NotInitialized` - Health check validation
- `TestPostgresAdapter_CreateUser_ValidationErrors` - Input validation
- `TestPostgresAdapter_ContextCancellation` - Context cancellation handling
- `TestPostgresAdapter_FullLifecycle_Integration` - Full lifecycle with real PostgreSQL container

## Container Runtime Detection

The integration tests automatically detect available container runtimes in this order:
1. Docker (via `docker ps`)
2. Podman (via `podman ps`)

If neither is available, tests are skipped with appropriate message.

## Environment Variables

### Container Runtime Configuration
- `DOCKER_HOST` - Used by testcontainers to connect to Podman socket
  - Example: `unix:///run/podman/podman.sock`

### Test Configuration
- `TESTCONTAINERS_RYUK_DISABLED` - Disable resource cleanup (useful for debugging)
  - Set to `true` to keep containers running after test failure

## Troubleshooting

### "Connection refused" or "Cannot connect to container runtime"

**On macOS with Podman:**
```bash
# Verify DOCKER_HOST is set correctly with podman socket path
echo $DOCKER_HOST

# Start podman machine if not running
podman machine start

# Get the correct socket path from podman machine start output
# It will be something like:
# unix:///var/folders/42/xfyh9ksn6sndqbtl0ybbtr700000gn/T/podman/podman-machine-default-api.sock

# Export and retry tests
export DOCKER_HOST='unix:///<your-socket-path>'
go test -tags=integration -v ./test/integration/storage/...
```

**On Linux with rootless Podman:**
```bash
# Verify socket exists
ls -l $XDG_RUNTIME_DIR/podman/podman.sock

# Set DOCKER_HOST if needed
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
go test -tags=integration -v ./test/integration/storage/...
```

### Container image not available
```bash
# Check if postgres:15-alpine image is available
podman images | grep postgres

# Pull image manually if needed
podman pull postgres:15-alpine

# Run tests with verbose output
TESTCONTAINERS_LOGS=true go test -tags=integration -v ./test/integration/storage/...
```

### Tests hang or timeout
```bash
# Check for orphaned containers
podman ps -a | grep postgres

# Clean up if needed
podman rm -f $(podman ps -aq --filter ancestor=postgres:15-alpine)

# Run tests with debug output
TESTCONTAINERS_LOGS=true go test -tags=integration -v ./test/integration/storage/...
```

## CI/CD Integration

### GitHub Actions Example
```yaml
- name: Run Integration Tests
  env:
    DOCKER_HOST: unix:///run/podman/podman.sock
  run: |
    # Start podman socket
    podman system service --time=0 unix:///run/podman/podman.sock &
    sleep 2

    # Run tests
    go test -tags=integration -v ./test/integration/storage/...
```

### GitLab CI Example
```yaml
integration-tests:
  image: golang:1.24
  services:
    - podman
  script:
    - export DOCKER_HOST=unix:///run/podman/podman.sock
    - podman system service --time=0 unix:///run/podman/podman.sock &
    - sleep 2
    - go test -tags=integration -v ./test/integration/storage/...
```

## Performance Notes

Integration tests with real PostgreSQL containers are slower than unit tests:
- Setup time: 2-5 seconds (pulling image, starting container)
- Test execution: 1-10 seconds per test
- Cleanup time: 1-2 seconds

For CI/CD, consider:
- Running integration tests separately from unit tests
- Caching container images to speed up setup
- Running tests in parallel (with appropriate test isolation)

## References

- [testcontainers-go Documentation](https://golang.testcontainers.org/)
- [Podman Documentation](https://podman.io/)
- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
