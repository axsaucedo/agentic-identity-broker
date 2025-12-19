# Storage Layer Troubleshooting Guide

This guide helps diagnose and resolve common storage layer issues.

## Startup Issues

### "Failed to Initialize Storage"

**Symptoms**: Application exits at startup with storage initialization error

**Check List**:

1. **Verify backend is configured**
   ```bash
   # Check config file
   cat config.yaml | grep -A 5 "^storage:"
   # Should show: backend: memory or backend: postgres
   ```

2. **If using PostgreSQL**:
   ```bash
   # Verify connection URL is set
   echo $IDENTITY_BROKER_STORAGE_POSTGRES_URL
   # Should not be empty

   # Test PostgreSQL connection
   psql $IDENTITY_BROKER_STORAGE_POSTGRES_URL -c "SELECT 1"
   ```

3. **Enable debug logging**:
   ```bash
   identity-broker --log-level=debug --config config.yaml
   ```

### "Database Connection Failed"

**Symptoms**: `ErrorKindConnection` error at startup

**Possible Causes**:

| Cause | Fix |
|-------|-----|
| PostgreSQL server not running | `docker run -d -e POSTGRES_PASSWORD=password postgres:15` |
| Wrong hostname in connection URL | Check DNS resolution: `nslookup db.example.com` |
| Port unreachable | `telnet localhost 5432` |
| SSL/TLS misconfiguration | Try `sslmode=disable` temporarily to diagnose |
| Firewall blocking connection | Check security groups/firewall rules |

**Debug Commands**:
```bash
# Test connection with psql directly
psql -h localhost -U postgres -d identity_broker

# Check network connectivity
telnet localhost 5432

# Verify connection URL format
# Expected: postgresql://user:password@host:port/dbname[?params]
```

### "Schema Not Found" or "Table Not Found"

**Symptoms**: `ErrorKindValidation` with message about missing tables

**Solution**: Run database migrations
```bash
# Create database and schema
identity-broker migrate up

# Verify schema was created
psql $IDENTITY_BROKER_STORAGE_POSTGRES_URL -c "
  SELECT tablename FROM pg_tables WHERE schemaname='public';"
```

Expected tables:
- `schema_migrations` - Tracks applied migrations
- `users` - Stores user entities

## Runtime Issues

### Timeout Errors

**Symptoms**: `ErrorKindTimeout` with "operation exceeded timeout"

**Likely Causes**:

1. **Database is slow**
   - Check PostgreSQL performance: `EXPLAIN ANALYZE SELECT * FROM users;`
   - Monitor CPU/memory: `top` or cloud provider dashboard
   - Check slow query log

2. **Connection pool exhausted**
   ```bash
   # Check active connections
   psql $IDENTITY_BROKER_STORAGE_POSTGRES_URL -c "
     SELECT count(*) as active_connections
     FROM pg_stat_activity;"
   ```

3. **Network latency**
   - Measure latency: `ping -c 5 db.example.com`
   - Check traceroute: `traceroute db.example.com`
   - Move application closer to database

**Solution - Increase Timeouts**:
```yaml
storage:
  backend: postgres
  timeouts:
    read: 10s    # Increase from 5s
    write: 20s   # Increase from 10s
```

### High Memory Usage with In-Memory Backend

**Symptoms**: Memory usage grows over time with memory backend

**Root Cause**: In-memory adapter has no eviction policy

**Solutions**:

1. **Restart application periodically**
   ```bash
   # Use deployment's rolling restart
   kubectl rollout restart deployment/identity-broker
   ```

2. **Switch to PostgreSQL for production**
   ```yaml
   # config.prod.yaml
   storage:
     backend: postgres  # Better for persistent, scalable storage
   ```

3. **Clean up old users** (development only)
   ```bash
   curl -X DELETE http://localhost:14000/admin/users/old_user_id
   ```

### Duplicate Key Conflicts

**Symptoms**: `ErrorKindConflict` when creating users

**Root Cause**: User with that ID already exists

**Solution - Verify User Doesn't Exist**:
```bash
# Check if user exists
curl http://localhost:8000/users/user123

# If 404, then safe to create
# If 200, user exists - update instead or use different ID
```

## Connection Pool Issues

### "Too Many Connections"

**Symptoms**: PostgreSQL error "too many connections"

**Solution - Reduce Pool Size**:

The default pool configuration is:
- Max open: 25
- Max idle: 5
- Max lifetime: 1 hour
- Max idle time: 15 minutes

To reduce load, either:

1. **Decrease number of identity-broker instances**
2. **Use connection pooler** (PgBouncer, pgpool)
   ```yaml
   storage:
     postgres:
       connection_url: postgresql://user@pgbouncer:6432/db
   ```

### Connection Timeouts

**Symptoms**: Occasional `ErrorKindTimeout` errors

**Check**: Connection pool health
```bash
# Monitor connections in real-time
watch -n 1 "psql $IDENTITY_BROKER_STORAGE_POSTGRES_URL -c
  'SELECT count(*) FROM pg_stat_activity;'"
```

**Solutions**:
1. Reduce `Max open connections` if PostgreSQL has connection limits
2. Use connection pooler to centralize connection management
3. Increase pool idle timeout if connections are closed prematurely

## Security Issues

### Sensitive Data in Logs

**Symptom**: Connection string appears in log output

**Solution - Already Built In**:
- Connection strings are automatically redacted in logs
- Never log raw configuration values
- Use `--log-level=error` in production to reduce verbosity

**Verify redaction is working**:
```bash
identity-broker --log-level=debug --config config.prod.yaml 2>&1 | grep -i password
# Should output nothing - password should be redacted as "[REDACTED]"
```

### SSL Certificate Verification Failure

**Symptoms**: `ErrorKindConnection` with SSL/certificate error

**Debug**:
```bash
# Test connection with SSL verification
psql "postgresql://user@host/db?sslmode=verify-full" -c "SELECT 1"

# If fails, check certificate
openssl s_client -connect host:5432 -showcerts

# Verify certificate is valid
openssl x509 -in /path/to/cert.pem -text -noout
```

**Solutions**:

1. **Accept self-signed certificate**
   ```yaml
   storage:
     postgres:
       connection_url: "postgresql://...?sslmode=require&sslrootcert=/path/to/ca.pem"
   ```

2. **Disable SSL (development only)**
   ```yaml
   storage:
     postgres:
       connection_url: "postgresql://...?sslmode=disable"
   ```

3. **Fix certificate**
   - Ensure certificate CN matches hostname
   - Update certificate expiration
   - Install intermediate certificates

## Data Integrity Issues

### Users Missing After Restart

**If Using In-Memory Backend**:
- This is expected behavior - in-memory adapter loses all data on restart
- Switch to PostgreSQL for persistent storage
- Or populate seed data on startup

**If Using PostgreSQL**:

1. **Verify data is in database**
   ```bash
   psql $IDENTITY_BROKER_STORAGE_POSTGRES_URL -c "
     SELECT id, email FROM users LIMIT 5;"
   ```

2. **Check application isn't deleting data**
   - Review recent changes
   - Check logs for DELETE statements
   - Look for cleanup/purge processes

3. **Verify backup integrity**
   ```bash
   # List backups
   pg_basebackup -D /tmp/backup -v
   ```

### Concurrent Access Issues

**Symptoms**: Occasional `ErrorKindConflict` or stale data

**In-Memory Adapter**:
- Uses `sync.RWMutex` for thread safety
- No race conditions (verified with `go test -race`)

**PostgreSQL Adapter**:
- Database handles concurrent access via transactions
- If issues persist, check:
  ```bash
  psql -c "SELECT * FROM pg_stat_activity;" # Check for locks
  pg_locks    # Check for blocking queries
  ```

## Performance Tuning

### Slow Read Operations

**Check Query Performance**:
```bash
psql -c "
EXPLAIN ANALYZE
SELECT id, email, created_at, updated_at FROM users LIMIT 100;"
```

**Solutions**:
1. Create indexes on frequently searched columns
2. Increase `read` timeout if database is returning data
3. Move database closer to application (reduce network latency)

### Slow Write Operations

**Check Write Performance**:
```bash
time psql -c "INSERT INTO users (id, email, created_at, updated_at)
              VALUES ('test', 'test@example.com', NOW(), NOW());"
```

**Solutions**:
1. Check disk I/O: `iostat` or cloud provider metrics
2. Increase `write` timeout temporarily
3. Batch writes using transactions (if implemented)
4. Archive old data if table is very large

## Debugging with Environment Variables

Enable detailed logging:
```bash
export IDENTITY_BROKER_LOG_LEVEL=debug
export IDENTITY_BROKER_LOG_FORMAT=json

identity-broker --config config.yaml
```

JSON logs can be parsed and analyzed:
```bash
# Filter for storage errors
identity-broker | jq 'select(.component=="storage")'

# Count errors by type
identity-broker | jq -r '.error_kind' | sort | uniq -c
```

## Testing Connectivity

### Script to Verify Storage Setup

```bash
#!/bin/bash

echo "Testing storage connectivity..."

# Check in-memory backend
echo "Testing in-memory backend..."
cat > test.yaml <<EOF
storage:
  backend: memory
  timeouts:
    read: 5s
    write: 10s
EOF

identity-broker --config test.yaml &
sleep 2
curl -f http://localhost:8000/health || echo "Health check failed"
pkill -f identity-broker

# Check PostgreSQL backend
if [ ! -z "$IDENTITY_BROKER_STORAGE_POSTGRES_URL" ]; then
  echo "Testing PostgreSQL backend..."
  cat > test.yaml <<EOF
storage:
  backend: postgres
  postgres:
    connection_url: $IDENTITY_BROKER_STORAGE_POSTGRES_URL
  timeouts:
    read: 5s
    write: 10s
EOF

  identity-broker --config test.yaml &
  sleep 3
  curl -f http://localhost:8000/health || echo "PostgreSQL health check failed"
  pkill -f identity-broker
fi

echo "Storage connectivity tests complete"
```

## Getting Help

If you cannot resolve the issue:

1. **Check logs with timestamps**
   ```bash
   identity-broker --log-level=debug 2>&1 | tee app.log
   ```

2. **Capture error context**
   - Note the exact error message and error kind
   - Record timestamps and operations that failed
   - Save configuration (with passwords redacted)

3. **Report issue with**:
   - Error kind and message
   - Configuration (with redacted credentials)
   - Debug logs
   - Steps to reproduce
   - Environment information (OS, Go version, PostgreSQL version)

## References

- [Storage Architecture](../adrs/004-storage-layer-architecture.md)
- [Security Checklist](../SECURITY.md)
- [Extension Guide](../docs/STORAGE_EXTENSION_GUIDE.md)
- PostgreSQL Documentation: https://www.postgresql.org/docs/
