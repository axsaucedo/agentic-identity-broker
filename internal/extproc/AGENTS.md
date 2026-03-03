# ExtProc Token Exchange Service — `internal/extproc/`

**Prefer retrieval-led reasoning. Read source files before making assumptions about types, interfaces, or behaviour.**

This subtree is a **standalone application** (`cmd/extproc-token-exchange/`). It does **not** share domain objects, ports, or adapters with the identity broker's `internal/` hexagonal core. All packages under `internal/extproc/` are only for this service.

---

## Package Map

```
internal/extproc/
  config/
    config.go       Config struct tree (GRPCConfig, OAuth2Config, TLSConfig, CacheConfig, LogConfig)
    loader.go       Viper loader — Load(), LoadWithCommand(), LoadFromViper(), RegisterFlags()
    validate.go     Validate() — 13 fail-fast rules, all errors collected and returned together

  server/
    server.go       ExtProc gRPC server (Server struct), Process streaming RPC, request-header processing
    exchanger.go    TokenExchanger implementation: RFC 8693 exchange, in-memory cache, singleflight, background refresh
    server_test.go  Unit tests for the gRPC server (streaming, pass-through, error responses)
    exchanger_test.go   Unit tests for TokenExchanger — cache, singleflight, client assertion, TTL
    exchanger_cache_test.go  Focused cache & eviction tests
    tls_test.go     Unit tests for buildHTTPClient — TLS config, CA bundle, InsecureSkipVerify
```

Entry point and wiring live **outside** this tree:

```
cmd/extproc-token-exchange/
  main.go       Cobra Execute call
  root.go       Cobra root command — config load → logger → NewTokenExchanger → NewServer → gRPC lifecycle
```

E2E acceptance tests live in a **separate Ginkgo suite** (per spec FR-017):

```
tests/e2e/extproc/
  extproc_suite_test.go
  token_exchange_test.go   (12 scenarios, 1:1 mapping to spec.md)
  bootstrap/
  helpers/
  fixtures/
```

---

## Key Types and Interfaces

### `internal/extproc/config`

| Type | Purpose |
|---|---|
| `Config` | Root config: `GRPC`, `OAuth2`, `Cache`, `Log` |
| `GRPCConfig` | `Bind`, `Port`, `MaxConcurrentStreams` |
| `OAuth2Config` | `TokenEndpoint`, `Issuer`, `ClientID`, `ClientSecret`, `ClientCredentialsEndpoint`, `ClientAssertionType`, `ExchangeTimeout`, `TLS` |
| `TLSConfig` | `InsecureSkipVerify`, `CaBundlePath`, `AllowHTTP` |
| `CacheConfig` | `DefaultTTL`, `MaxTTL` |
| `LogConfig` | `Level`, `Format` |

**Loading pipeline** (in `loader.go`):
1. Viper with `EXTPROC_` prefix + dot→underscore key replacer
2. Optional YAML file via `EXTPROC_CONFIG_PATH`
3. CLI flags via `RegisterFlags()` + `bindFlags()` (highest precedence)
4. `${VAR}` expansion via `expandEnvVars()` (all string fields)
5. `Validate()` (fail-fast, all errors collected)

### `internal/extproc/server`

| Type/Symbol | Purpose |
|---|---|
| `Exchanger` interface | Port: `Exchange(subjectToken, resourceURI string) (string, error)` + `Shutdown()` |
| `Server` struct | Implements `ExternalProcessorServer`. Fields: `cfg`, `exchanger`, `logger` |
| `NewServer(cfg, exchanger, logger)` | Constructor |
| `Server.Process(stream)` | Streaming gRPC RPC — dispatches on `req.Request` type |
| `TokenExchanger` struct | Concrete `Exchanger` implementation |
| `NewTokenExchanger(cfg, logger)` | Constructor — acquires client assertion at startup (fail-fast), starts background goroutine |
| `ErrAssertionExpired` | Sentinel error → caller returns 503 |
| `tokenCacheKey` struct | Map key for token cache: `{subjectToken, resourceURI}` — struct avoids separator-injection |
| `cachedToken` | `accessToken string`, `expiresAt time.Time` |
| `assertionState` | `value`, `issuedAt`, `expiresAt` — atomically swapped via `atomic.Pointer[assertionState]` |

---

## Architecture Invariants

### 1. Standalone — no identity broker imports
`internal/extproc/` must **never** import from `internal/domain/`, `internal/ports/`, or `internal/adapters/`. It is a separate binary, not a plugin.

### 2. Exchanger is the only port interface
`Exchanger` (defined in `server/server.go`) is the hexagonal boundary inside this service. `Server` depends on the interface; `TokenExchanger` is the production implementation. Tests use mock implementations of `Exchanger`.

### 3. Config is loaded once at startup
`LoadWithCommand()` is called in `cmd/extproc-token-exchange/root.go`. No ad-hoc config reading elsewhere. The loaded `*Config` is passed by pointer.

### 4. Fail-fast at startup
- Config validation: 13 rules in `Validate()` — all errors are collected, not short-circuited.
- Client assertion: `NewTokenExchanger()` calls `refreshClientAssertion()` synchronously and returns an error if it fails.
- CA bundle: `buildHTTPClient()` reads and parses the bundle at construction time.

### 5. Token cache design
- Map key: `tokenCacheKey{subjectToken, resourceURI}` — Go struct map key (no hash, no separator injection risk).
- Lock: `sync.RWMutex cacheMu` with double-checked locking pattern inside `singleflight.Group.Do`.
- TTL: use `expires_in` from exchange response → fall back to `cache.default_ttl` → cap at `cache.max_ttl`.
- Eviction: background goroutine in `runEviction()`, fires every `DefaultTTL/2` (floor: 1s).

### 6. Client assertion refresh
- `assertionState` is stored via `atomic.Pointer[assertionState]` (lock-free reads).
- Background ticker in `runEviction()` fires every `min(DefaultTTL/2, 30s)`.
- Refresh triggers when remaining lifetime < `max(20% of total lifetime, 30s)`.
- `ErrAssertionExpired` is returned by `Exchange()` if the assertion is nil or expired — signals 503 to Envoy.

### 7. Body/trailer pass-through
Envoy ExtProc requires a **phase-specific** response type for each message received. Use `StreamedBodyResponse` (not `BodyMutation_Body`) for request and response bodies — agentgateway discards body content otherwise.

### 8. Resource URI construction
agentgateway sends HTTP/2 pseudo-headers (`:path`, `:scheme`, `:authority`) separately. `buildResourceURI()` assembles `{scheme}://{authority}{path}`. If `:path` is already absolute (starts with `http://` or `https://`), it is used as-is. `validateResourceURI()` enforces non-empty, absolute URI with `http`/`https` scheme.

---

## Validation Rules Summary (13 rules)

| # | Rule |
|---|---|
| 1 | `grpc.port` 1–65535 |
| 2 | `grpc.bind` non-empty |
| 3 | `oauth2.token_endpoint` valid URL with http/https scheme |
| 4 | `oauth2.issuer` valid URL with http/https scheme |
| 5 | `oauth2.client_id` non-empty |
| 6 | `oauth2.client_secret` non-empty (after env expansion) |
| 7 | `cache.default_ttl` positive duration |
| 8 | `token_endpoint` and `issuer` must use `https://` unless `oauth2.tls.allow_http: true` |
| 9 | `cache.max_ttl` positive duration |
| 10 | `oauth2.exchange_timeout` positive duration |
| 11 | `log.level` one of: `debug`, `info`, `warn`, `error` (empty → `info`) |
| 12 | `log.format` one of: `text`, `json` (empty → `text`) |
| 13 | `oauth2.client_assertion_type` one of: `id_token`, `access_token` (default: `id_token`) |

---

## Testing Conventions (this subtree)

- **Framework**: stdlib `testing` + `testify` (`require` for preconditions, `assert` for checks).
- **Package**: tests use `package server_test` or `package config_test` (external/black-box style).
- **Mocking**: hand-rolled mock `Exchanger` in server tests; `httptest.NewServer` for OAuth2 + token exchange endpoints in exchanger tests.
- **Config helpers**: tests call `extprocconfig.LoadFromViper(v)` with a pre-configured `*viper.Viper` to avoid env pollution.
- **No Ginkgo here**: unit tests use standard `t.Run` / table-driven subtests. Ginkgo/Gomega is reserved for `tests/e2e/extproc/`.
- **Race detector**: all tests must pass with `go test -race`. The `atomic.Pointer` assertion state and `sync.RWMutex` cache are designed for this.

---

## Configuration Reference (defaults)

| Key | Default | Notes |
|---|---|---|
| `grpc.bind` | `0.0.0.0` | |
| `grpc.port` | `50051` | |
| `grpc.max_concurrent_streams` | `100` | |
| `oauth2.token_endpoint` | _(required)_ | RFC 8693 exchange endpoint |
| `oauth2.issuer` | _(required)_ | Used to derive client_credentials URL if explicit endpoint not set |
| `oauth2.client_id` | _(required)_ | |
| `oauth2.client_secret` | _(required)_ | Never logged |
| `oauth2.client_credentials_endpoint` | `{issuer}/oauth/token` | Optional override |
| `oauth2.client_assertion_type` | `id_token` | `id_token` or `access_token` |
| `oauth2.exchange_timeout` | `5s` | |
| `oauth2.tls.insecure_skip_verify` | `false` | DEV ONLY |
| `oauth2.tls.ca_bundle_path` | `` | |
| `oauth2.tls.allow_http` | `false` | DEV ONLY — disables HTTPS enforcement |
| `cache.default_ttl` | `5m` | |
| `cache.max_ttl` | `1h` | |
| `log.level` | `info` | |
| `log.format` | `text` | |

Full example: `examples/config/extproc-token-exchange.yaml`

---

## Spec and Design Documents

| Document | Path |
|---|---|
| Feature spec | `specs/015-extproc-token-exchange/spec.md` |
| Implementation plan | `specs/015-extproc-token-exchange/plan.md` |
| Data model | `specs/015-extproc-token-exchange/data-model.md` |
| gRPC contract | `specs/015-extproc-token-exchange/contracts/extproc-grpc.md` |
| Configuration schema | `specs/015-extproc-token-exchange/contracts/configuration.md` |
| Docker Compose integration | `specs/015-extproc-token-exchange/contracts/docker-compose.md` |
| Quickstart | `specs/015-extproc-token-exchange/quickstart.md` |
