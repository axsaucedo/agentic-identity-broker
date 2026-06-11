# Frontend E2E Tests (`tests/e2e/frontend/`)

> **Prefer retrieval-led reasoning. Read existing test files and page objects before writing new tests.**

Browser-based tests using Playwright via `playwright-go`.

## Rules (inherit from `../AGENTS.md`)

All E2E rules apply: 1:1 spec-to-test mapping, spec traceability comments, production bootstrap. See [../AGENTS.md](../AGENTS.md).

## Directory Layout

```
tests/e2e/frontend/
  frontend_suite_test.go          Ginkgo suite — PlaywrightHelper init in BeforeSuite
  *_flow_test.go                  Consent, revoke, and CIMD browser journeys
  *_consent_test.go               Consent-focused CIMD browser scenarios
  csrf_grant_save_test.go         CSRF-protected grant submission regression coverage
  permission_sets_frontend_test.go Permission-set rendering and interaction coverage
  selection_preservation_test.go  Selection round-trip persistence coverage
```

## Playwright Lifecycle

- `frontend_suite_test.go` launches one browser in `BeforeSuite` and creates a fresh browser context + page in `BeforeEach`
- Normal verification runs the suite with multiple Ginkgo workers (`GINKGO_FRONTEND_PROCS`), so every spec must remain isolated to its own test server, browser context, and storage
- Page objects in `../pages/` abstract selectors — tests call functional interaction methods, never raw selectors
- Use `bootstrap.TestLogger()` / suite helpers instead of writing directly to `GinkgoWriter` for routine logging

## Page Objects (`../pages/`)

| File | Type | Responsibilities |
|---|---|---|
| `page.go` | `Page` | Base: navigation, waiting |
| `consent_page.go` | `ConsentPage` | Approve/deny consent interactions |

## Writing Frontend Tests

- Use page objects — never interact with raw selectors in test files
- Add new page objects to `../pages/` when covering a new UI screen
- Browser context is isolated per test via `BeforeEach` — no shared browser state
- Prefer Playwright waits / Gomega `Eventually` over `time.Sleep(...)`

## Screenshots and Performance

- `Page.TakeScreenshot()` is a **no-op unless** `E2E_CAPTURE_SCREENSHOTS=true`
- Normal verification runs with screenshots disabled for speed
- The dedicated GitHub workflow `.github/workflows/frontend-e2e-screenshots.yml` enables screenshots and updates tracked PNG artifacts
- Do **not** add screenshots to every new spec by default. Only capture one when the screenshot itself is a maintained artifact or review aid
- If a test needs screenshots, it must also remain compatible with the screenshot workflow's serial execution (run without `--procs`) and use stable, unique filenames
