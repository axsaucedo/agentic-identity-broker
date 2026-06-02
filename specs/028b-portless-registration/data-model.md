# Data Model: Portless Redirect URI Registration

**Branch**: `feat/ephemeral-runtime-port` | **Date**: 2026-06-02

## Overview

This feature introduces no new entities, no schema changes, and no configuration parameters. The entire change is a **new pure function** and updates to two call sites.

---

## New: `MatchesRedirectURI` (Value Function)

**Package**: `internal/domain/storage`
**File**: `agent.go` (added alongside `IsValidRedirectURI`)

```
MatchesRedirectURI(registered string, incoming string) bool
```

**Semantics**:

| Component | Loopback host (`localhost`, `127.0.0.1`) | Non-loopback host |
|---|---|---|
| Scheme | Must match exactly | Must match exactly |
| Host | Must match exactly | Must match exactly |
| Port | **Ignored** | Must match exactly |
| Path | Must match exactly | Must match exactly |
| Query | Must match exactly | Must match exactly |

**Loopback detection**: `u.Hostname() == "localhost" || u.Hostname() == "127.0.0.1"` — applied to the **registered** URI. If the registered URI is loopback, port is stripped from both before comparison.

**Failure modes**:
- Either URI fails `url.Parse`: returns `false`
- Registered URI has empty host: returns `false`
- All component mismatches: returns `false`

**No side effects**: pure function, no logger, no external dependencies.

---

## Changed: Redirect URI List-Membership Check

**Two sites updated** to use `MatchesRedirectURI` instead of `==`:

### Site 1 — `internal/domain/oauth2/service.go`

```
Before:
    if req.RedirectURI == allowed { uriAllowed = true; break }

After:
    if redirect.MatchesRedirectURI(allowed, req.RedirectURI) {
        uriAllowed = true
        // audit log: registered_redirect_uri=allowed, incoming_redirect_uri=req.RedirectURI
        break
    }
```

The slog audit entry is emitted unconditionally on every loopback match (SR-003). Non-loopback matches require no audit entry beyond existing error logging.

### Site 2 — `internal/domain/oauth2server/provider.go`

```
Before (contains helper):
    if v == item { return true }

After:
    if redirect.MatchesRedirectURI(v, item) { return true }
```

Audit logging here is best-effort: the `contains` helper currently has no logger reference. The audit requirement (SR-003) is fully satisfied by site 1 (which handles both CIMD and opaque flows). If the fosite path is taken, a logger can be passed or a separate check added post-`contains` call in the caller.

---

## Unchanged: `IsValidRedirectURI`

`internal/domain/storage/agent.go:IsValidRedirectURI` — no changes.

Already accepts portless loopback URIs (`http://localhost/callback`) as structurally valid. Write-time validation is orthogonal to runtime matching semantics.

---

## Unchanged: `validateRedirectOrigin` (CIMD)

`internal/domain/oauth2/cimd/document.go:validateRedirectOrigin` — no changes.

This validates CIMD document fetch-time same-origin constraints. Already skips same-origin for loopback. Port is not checked for loopback in this path.

---

## No New Packages

Both callers (`service.go`, `provider.go`) already import `internal/domain/storage`. No import changes needed at either call site.

## No Schema Changes

- No new database tables or columns
- No new migration files
- No changes to `Agent`, `UserSession`, `UserGrant`, or any other persisted entity

---

## No Configuration Changes

- No new fields in `internal/ports/config.go`
- No Helm chart updates required
- No new YAML configuration examples
