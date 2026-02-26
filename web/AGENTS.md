# Consent Frontend — React SPA (`web/`)

**Prefer retrieval-led reasoning. Read source files before making assumptions about components, hooks, API types, or design tokens.**

## Overview

React-based consent management SPA where users review, grant, and revoke AI agent permissions to third-party OAuth2 services. Served from the Go backend at `/consent` (ADR 005). The UI implements the **Refined Trust Architecture** design aesthetic — sophisticated, trust-communicating, WCAG 2.1 AA compliant.

## Tech Stack

| Technology | Version | Role |
|---|---|---|
| React | 19 | UI framework (StrictMode) |
| TypeScript | 5.3+ | Type safety |
| Vite | 7 | Build tool + dev server |
| Tailwind CSS | 4 | Utility-first styling via semantic tokens |
| Headless UI | 2 | Accessible unstyled component primitives |
| React Router | 7 | Client-side routing (basename `/consent`) |
| Axios | 1 | HTTP client with interceptors |
| Framer Motion | 12 | Page transitions and animations |
| Vitest | 4 | Unit/integration testing |
| React Testing Library | 16 | Component testing |
| Storybook | 10 | Design system documentation and visual testing |

## Source Structure

```
src/
  App.tsx              Router setup: /, /agent/:agentId, /oauth2/sessions, *→404
  main.tsx             Entry point — React.StrictMode mount

  components/
    consent/           Consent-specific components (DelegationCard, ServiceCard, ScopeList,
                       GrantValidityControl, ServiceRequirementCard, GrantStatusBadge, etc.)
    layout/            AppLayout, Header — page chrome
    sessions/          SessionCard, TerminationDialog — OAuth2 session management
    ui/                Reusable UI primitives (Button, Toast, ErrorBoundary, Skeleton,
                       EmptyState, Switch, InlineError, PageTransition, DatePicker)

  design-system/       ★ AUTHORITATIVE design reference — read before styling anything
    components/        Design system component library (7 categories):
      primitives/        Button, Badge, Avatar, Spinner, Divider
      inputs/            TextInput, TextArea, Select, Checkbox, Radio, Switch, DatePicker
      data-display/      Card, Table, ScopeList, StatusIndicator
      layout/            AppLayout, Container, Stack, Grid, PageTransition
      navigation/        Tabs, Pagination, Breadcrumb
      overlays/          Modal, Tooltip, Dropdown, Popover
      feedback/          EmptyState, InlineError, Skeleton, Alert
      advanced/          Progress
    tokens/            Design tokens (colors, typography, spacing, shadows, radius, animation, z-index, breakpoints)
    utils/             cn() (clsx + tailwind-merge), a11y helpers, focus utilities
    docs/              ★ Read before any UI work:
      INDEX.md                Complete documentation index
      DESIGN_PRINCIPLES.md    Visual philosophy — "Refined Trust Architecture"
      COLOR_GUIDE.md          Full color palette with hex values + WCAG ratios
      TOKEN_GUIDE.md          All design tokens with usage examples
      COMPONENT_ARCHETYPES.md Foundational component specifications
      COMMON_MISTAKES.md      Anti-patterns with correct solutions
      MOTION_GUIDE.md         Animation timing and easing specifications
      COMPONENT_PAIRING_GUIDE.md  Component composition patterns
      COMPOSITION_PATTERNS.md     Complex layout recipes
      DECISION_TREES.md           Which component to use when
      ACCESSIBILITY_GUIDE.md      WCAG 2.1 AA compliance requirements

  hooks/               Custom React hooks
    useConsent          Consent overview data fetching + state
    useAgentGrants      Agent detail + grant management
    useSessions         OAuth2 session listing + operations
    useToggleGrant      Grant enable/disable toggle logic
    useUpdateValidity   Grant expiration date management
    useRetry            Generic retry with exponential backoff

  pages/               Route-level components (lazy-loaded via React.lazy)
    ConsentOverviewPage     / — agent delegation list
    AgentGrantDetailPage    /agent/:agentId — per-agent grant management
    ThirdPartySessionsPage  /oauth2/sessions — session listing and termination
    ErrorPage               * — 404 fallback
    NotFoundPage            Standalone 404

  services/
    api/
      client.ts        Axios instance — baseURL: /api, 30s timeout, error interceptors
      consent.ts       ConsentApiService class — getUserInfo, getAgentDelegations,
                       getAgentDetail, getAgentGrants, createOrUpdateGrant, revokeGrant
      sessions.ts      SessionsApiService — listSessions, getSessionDetail,
                       terminateSession, refreshSession
      cache.ts         In-memory TTL cache (5 min default) for GET responses
      index.ts         Barrel export
    storage/
      session.ts       Session token storage (sessionStorage + localStorage fallback)

  styles/
    index.css          Global styles + Tailwind directives
    fonts.css          Font imports (Crimson Pro, Manrope, JetBrains Mono)

  types/
    consent.ts         All API response/request TypeScript types (UserInfo, AgentDelegation,
                       AgentDetail, ThirdpartyService, UserGrant, ServiceScope, etc.)

  utils/
    validation.ts      Input validation (URL safety, etc.)
    scrollToError.ts   Scroll-to-first-error UX helper
```

## Path Aliases

Configured in `vite.config.ts`, `vitest.config.ts`, and `.storybook/main.ts`:

| Alias | Path |
|---|---|
| `@design-system` | `./src/design-system` |
| `@components` | `./src/components` |
| `@hooks` | `./src/hooks` |
| `@services` | `./src/services` |
| `@types` | `./src/types` |
| `@utils` | `./src/utils` |
| `@assets` | `./src/assets` |
| `@styles` | `./src/styles` |

Always use these aliases in imports — never relative paths that cross alias boundaries.

## Design System Rules

The design system at `src/design-system/` is the **single source of truth** for all visual decisions. Read `docs/COMMON_MISTAKES.md` before writing any styled component.

### Critical Rules

1. **Semantic color tokens only** — Use `text-trust-deep`, `bg-neutral-50`, `text-error-primary`, etc. Never use raw Tailwind grays (`gray-*` does not exist in this system). Use warm neutrals (`neutral-*`).
2. **Typography hierarchy** — Headings: `font-display` (Crimson Pro). Body: `font-sans` (Manrope). Code/technical: `font-mono` (JetBrains Mono). Headings must use `text-trust-deep` or `text-trust`, never `text-neutral-*`.
3. **Elevation via shadows, not borders** — Cards use multi-layer shadows. Borders are used sparingly for containment only (muted taupe/slate).
4. **Animation timing** — 150ms (hover), 200ms (state changes), 300ms (elevation/modal), 500ms (page transitions). All must respect `prefers-reduced-motion`.
5. **WCAG 2.1 AA** — All interactive elements must have visible focus indicators (2px navy ring at 2px offset). Color contrast ratios must meet AA. Use `eslint-plugin-jsx-a11y` rules.
6. **Component composition** — Prefer design system components from `design-system/components/` over ad-hoc implementations. Check `docs/DECISION_TREES.md` to select the right component.

### Semantic Color Palette

| Token Family | Hex (primary) | Use |
|---|---|---|
| `trust-*` | #0A2540 (deep), #1E4D6B, #E8F1F5 (light) | Primary brand, headings, actions |
| `cta-*` | #D97706 | Call-to-action buttons and links |
| `success-*` | #059669 | Granted permissions, success states |
| `error-*` | #DC2626 | Error states, destructive actions |
| `warning-*` | #D97706 | Warnings, attention signals |
| `neutral-*` | #faf9f7 → #1a1a1a (50–900 scale) | Backgrounds, body text, borders |

## API Client Pattern

- `services/api/client.ts` exports a configured Axios instance (`apiClient`)
- Base URL is `/api` (relative) — Vite proxy forwards to Go backend in dev, reverse proxy in prod
- **Authentication is external** — Vite dev proxy injects `X-Remote-User: dev@example.com`; in production, handled by upstream reverse proxy. No auth logic in frontend code.
- Service classes (`ConsentApiService`, `SessionsApiService`) provide typed methods wrapping `apiClient`
- GET responses are cached in-memory via `apiCache` (5 min TTL) — cache is invalidated on mutations
- Error responses are normalized to `ApiError` shape via response interceptor

### Key API Endpoints

| Method | Endpoint | Service Method |
|---|---|---|
| GET | `/api/me` | `consentApi.getUserInfo()` |
| GET | `/api/consent/agents` | `consentApi.getAgentDelegations()` |
| GET | `/api/consent/agent/:id` | `consentApi.getAgentDetail(id)` |
| GET | `/api/consent/agent/:id/grants` | `consentApi.getAgentGrants(id)` |
| PUT | `/api/consent/agent/:id/grant` | `consentApi.createOrUpdateGrant(id, grant)` |
| DELETE | `/api/consent/agent/:id/grant` | `consentApi.revokeGrant(id)` |
| GET | `/api/third-party/sessions` | `sessionsApi.listSessions()` |
| GET | `/api/third-party/:sid/session` | `sessionsApi.getSessionDetail(sid)` |
| DELETE | `/api/third-party/:sid/session` | `sessionsApi.terminateSession(sid)` |

## Development Setup

```bash
just web-install          # Install npm deps (once after clone)
just web-dev              # Vite dev server :3000 (proxies API → Go :8000)
just web-build            # Production build → web/dist/consent/
```

Vite dev server runs on port 3000. It proxies all non-frontend requests to `http://localhost:8000` and injects `X-Remote-User: dev@example.com` header. Docker dev uses `VITE_USE_POLLING=true` for file watching.

### Storybook

```bash
cd web && npm run storybook     # Port 6006
```

Storybook renders only `src/design-system/` stories and MDX docs. Used for visual testing and design system documentation — not for application-level components.

## Testing Conventions

- **Framework**: Vitest + React Testing Library + jsdom
- **Setup**: `vitest.setup.ts` — extends `expect` with jest-dom matchers, runs `cleanup()` after each test
- **File naming**: Co-located `*.test.tsx` (unit), `*.interactive.test.tsx` (interaction-heavy), `*.integration.test.tsx` (multi-component)
- **Testing approach**: Test user-visible behavior, not implementation details. Use `screen.getByRole`, `getByText`, `getByLabelText` — avoid `getByTestId` unless no semantic alternative exists.
- **Mocking**: Mock API services at the module level (`vi.mock('@services/api')`). Never mock React hooks directly — mock their data sources.
- **Accessibility**: `eslint-plugin-jsx-a11y` enforced. Every interactive component must be keyboard-navigable and have proper ARIA attributes.

## Architecture Boundary

The frontend communicates with the backend **exclusively through HTTP API endpoints** defined in `/api/enduser/openapi.yaml`. No direct Go imports, no shared types, no backend coupling beyond the API contract. TypeScript types in `types/consent.ts` mirror the OpenAPI response schemas — keep them in sync.

## Code Splitting

All page-level components are lazy-loaded via `React.lazy()` and wrapped in `<Suspense>`. This ensures the initial bundle only includes the router shell, loading fallback, error boundary, and toast provider. Route bundles are loaded on navigation.
