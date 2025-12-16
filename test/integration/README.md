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

Podman is supported as an alternative to Docker. Note: testcontainers-go v0.40.0 has limitations with Podman on some platforms.

**On Linux with rootless Podman:**
```bash
# Podman socket should be available automatically
export DOCKER_HOST=unix://$XDG_RUNTIME_DIR/podman/podman.sock
go test -tags=integration -v ./test/integration/storage/...
```

**On macOS with Podman machine:**
```bash
# Start the Podman machine
podman machine start

# Attempt to run integration tests
# Note: May skip if testcontainers-go can't access the socket (this is expected behavior)
go test -tags=integration -v ./test/integration/storage/...
```

**Known Issues:**
- testcontainers-go v0.40.0 requires rootless container support which Podman on macOS (via VM) may not expose correctly
- Podman on macOS uses SSH connections to the VM, which testcontainers-go may not auto-detect
- When integration with Podman fails, tests gracefully skip with appropriate message
- For reliable integration testing on macOS, Docker Desktop is recommended
- See: https://golang.testcontainers.org/ for testcontainers-go configuration options

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

### "Podman socket not found" error
```bash
# Check if podman socket is running
ls -l /run/podman/podman.sock

# Start podman socket if not running
podman system service --time=0 unix:///run/podman/podman.sock &
```

### "Connection refused" error
```bash
# Verify DOCKER_HOST is set correctly
echo $DOCKER_HOST

# Try connecting directly
curl --unix-socket /run/podman/podman.sock http://localhost/v1.0.0/libpod/info
```

### Container fails to start
```bash
# Check if image is available
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

# Disable Ryuk cleanup for debugging
TESTCONTAINERS_RYUK_DISABLED=true go test -tags=integration -v ./test/integration/storage/...
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
