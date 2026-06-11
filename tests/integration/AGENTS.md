# Integration Tests (`tests/integration/`)

**Prefer retrieval-led reasoning. Read the existing suite helpers before adding new infrastructure-backed tests.**

This directory mixes two styles of integration coverage:

- **Self-contained integration** — `go test ./tests/integration/...`
- **Infra-backed integration** — `go test -tags=integration ...`

## Helper Inventory

### Shared PostgreSQL bootstrap

Use `tests/integration/bootstrap/postgres.go` before writing any new PostgreSQL-backed test.

Key helpers:

- `bootstrap.RequireSharedPostgres(t)` — starts or reuses one PostgreSQL container for the package
- `(*SharedPostgres).SetupDatabaseFromTemplate(...)` — clones a fresh database from a prepared template
- `(*SharedPostgres).ApplyMigrationsUpTo(...)` — applies a bounded migration set into a template DB
- `(*SharedPostgres).ApplyMigration(...)`, `QuerySQL(...)`, `ExecuteSQL(...)` — SQL helpers for migration-style assertions
- `bootstrap.TerminateSharedPostgres(ctx)` — call from `TestMain` in infra-backed packages that use the shared container

**Preferred pattern:**

1. Start one shared container per package
2. Build a template DB once
3. Clone a fresh DB per test
4. Close adapters and drop the cloned DB in `defer`

When a file contains many small CRUD-style assertions against the same schema, prefer **file-level shared DB consolidation**:

1. Clone one DB for the test group (often one top-level `TestXxxRepo`)
2. Reuse one adapter/repository across sequential `t.Run(...)` subtests
3. Generate fresh fixtures inside each subtest
4. Keep special global-state assertions isolated if needed

**Avoid:** starting a brand-new PostgreSQL container in every test unless the test is specifically validating container startup behavior.

### AWS emulator bootstrap

`tests/integration/bootstrap/aws_emulator.go` provides:

- `StartAWSEmulator()`
- `StartAWSEmulatorForSuite()`

Use the suite-level variant whenever multiple tests in a package need KMS + DynamoDB.

### HTTP readiness helper

Self-contained HTTP integration tests should prefer `waitForEndpoint(t, url)` from `tests/integration/http_wait_test.go` instead of fixed `time.Sleep(...)` delays.

## Performance Rules

### 1. Keep infra packages parallel-safe

CI runs infra-backed packages with package parallelism (`go test -tags=integration -p 2 ...`).

That means:

- no cross-package shared mutable files
- no assumptions about package execution order
- restore any package-global environment mutations
- isolate databases per test with cloned DBs, not shared mutable schemas

### 2. Prefer cloned databases over repeated migrations

If several tests need the same migrated schema:

- create a template DB once
- clone it for each test
- for very small repo assertions, consider sharing one cloned DB across sequential subtests in the same file

This is much faster than:

- starting a container per test
- rerunning the full migration chain per test
- re-initializing a fresh adapter for every tiny CRUD assertion

### 3. Bounded failure-path tests only

For timeout, retry, or cancellation scenarios:

- shorten retry delays / TTLs in test config
- use handlers that unblock on context cancellation or a short fallback timeout
- avoid multi-second sleeps unless the duration itself is the behavior under test

### 4. Use SQL helpers for migration assertions

Migration tests should prefer `QuerySQL` / `ExecuteSQL` helpers over ad-hoc shell parsing. Keep assertions focused on:

- schema existence
- column/index presence
- representative data behavior

### 5. How to keep shared-DB subtests stable

If multiple subtests share one cloned DB, all of the following are required:

- **Unique data per subtest** — generate fresh IDs, principals, client IDs, code hashes, KIDs, service IDs, etc.
- **No `t.Parallel()` inside the shared DB group** — these subtests are intentionally sequential.
- **Avoid global count assertions** unless you capture a baseline first (`before`, `after`) or filter only the records created by that subtest.
- **Prefer presence/absence checks** over exact table cardinality when prior subtests may leave unrelated rows behind.
- **Split out special cases** into a separate cloned DB when the test depends on a pristine table, global ordering, or global process state.
- **Keep global-state tests isolated** — examples: tests that replace the global OTel tracer provider or mutate package-wide settings.
- **Do not rely on subtest order for correctness** — order can make timing reproducible, but each subtest should still stand on its own fixtures.

## Package Guidance

### `infra/`

- Good for migration-specific or AWS-emulator-backed assertions
- Put shared suite lifecycle in `main_test.go`
- Reuse the shared PostgreSQL helper when testing schema changes

### `migrations/`

- Use `framework.go` for migration lifecycle tests
- Keep migration coverage representative, not exhaustive per column unless the migration risk justifies it

### `storage/infra/`

- Use the shared PostgreSQL bootstrap
- Keep repository behavior isolated per cloned DB
- Reuse adapter setup helpers instead of rebuilding them in every test

### top-level self-contained tests

- Prefer polling over sleeps for server readiness
- Use in-memory adapters unless the behavior specifically requires real infra

## Anti-patterns

Flag these in review:

- one Docker/Postgres container per test when a shared template would work
- repeated full migration application inside every test body
- fixed sleeps for server startup or readiness
- retry/cancellation tests that depend on production-sized delays
- package-global state that prevents safe `-p 2` execution
