# Frontend E2E Tests (`tests/e2e/frontend/`)

Browser-based tests using Playwright via `playwright-go`.

## Rules (inherit from `../AGENTS.md`)

All E2E rules apply: 1:1 spec-to-test mapping, spec traceability comments, production bootstrap. See [../AGENTS.md](../AGENTS.md).

## Directory Layout

```
tests/e2e/frontend/
  frontend_suite_test.go         Ginkgo suite — PlaywrightHelper init in BeforeSuite
  consent_flow_test.go           Consent UI grant/deny scenarios
  revoke_grant_flow_test.go      Grant revocation UI scenarios
```

## Playwright Lifecycle

- `PlaywrightHelper` manages browser lifecycle: init in `BeforeSuite`, browser context per test in `BeforeEach`
- Page objects in `../pages/` abstract selectors — tests call functional interaction methods, never raw selectors

## Page Objects (`../pages/`)

| File | Type | Responsibilities |
|---|---|---|
| `page.go` | `Page` | Base: navigation, waiting |
| `consent_page.go` | `ConsentPage` | Approve/deny consent interactions |

## Writing Frontend Tests

- Use page objects — never interact with raw selectors in test files
- Add new page objects to `../pages/` when covering a new UI screen
- Browser context is isolated per test via `BeforeEach` — no shared browser state
