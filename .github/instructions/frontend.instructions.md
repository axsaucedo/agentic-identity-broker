---
applyTo: "web/**"
---

# Consent Frontend — React SPA

Full reference: `web/AGENTS.md`

## Tech Stack

React 19 + TypeScript 5.3+ + Vite 7 + Tailwind CSS 4 + Headless UI 2 + React Router 7 + Axios + Framer Motion 12 + Vitest 4 + React Testing Library 16 + Storybook 10

## Source Structure

```
src/
  App.tsx              Router: /, /delegations, /agents/:agentId, /sessions, *→404
  components/          consent/ (delegation cards), layout/ (chrome), sessions/, ui/ (primitives)
  design-system/       ★ AUTHORITATIVE design reference — read docs/ before styling
    components/        7 categories: primitives, inputs, data-display, layout, navigation, overlays, feedback
    tokens/            Colors, typography, spacing, shadows, radius, animation
    docs/              INDEX.md, DESIGN_PRINCIPLES.md, COLOR_GUIDE.md, COMMON_MISTAKES.md, etc.
  hooks/               useConsent, useAgentGrants, useSessions, useToggleGrant, useUpdateValidity, useRetry
  pages/               Lazy-loaded route components
  services/api/        client.ts (Axios), consent.ts, sessions.ts, cache.ts (5 min TTL)
  types/consent.ts     API response/request TypeScript types
```

## Path Aliases (use these, not relative paths)

`@design-system` → `./src/design-system`, `@components`, `@hooks`, `@services`, `@types`, `@utils`, `@assets`, `@styles`

## Design System Rules (Critical)

1. **Semantic color tokens only** — `text-trust-deep`, `bg-neutral-50`, `text-error-primary`. Never raw Tailwind grays. Use warm neutrals (`neutral-*`).
2. **Typography** — Headings: `font-display` (Crimson Pro). Body: `font-sans` (Manrope). Code: `font-mono` (JetBrains Mono). Headings use `text-trust-deep` or `text-trust`.
3. **Elevation via shadows, not borders** — Cards use multi-layer shadows.
4. **Animation timing** — 150ms hover, 200ms state, 300ms modal, 500ms page. Respect `prefers-reduced-motion`.
5. **WCAG 2.1 AA** — Visible focus indicators (2px navy ring at 2px offset). Proper contrast ratios.
6. **Component composition** — Prefer `design-system/components/` over ad-hoc. Check `docs/DECISION_TREES.md`.

Read `design-system/docs/COMMON_MISTAKES.md` before writing any styled component.

## Semantic Color Palette

| Family | Primary Hex | Use |
|---|---|---|
| `trust-*` | #0A2540 | Brand, headings, actions |
| `cta-*` | #D97706 | Call-to-action buttons/links |
| `success-*` | #059669 | Granted permissions |
| `error-*` | #DC2626 | Errors, destructive actions |
| `neutral-*` | #faf9f7→#1a1a1a | Backgrounds, text, borders |

## API Client Pattern

- Base URL `/api` — Vite proxy → Go backend in dev, reverse proxy in prod
- Auth is external (upstream proxy injects `X-Remote-User`)
- Service classes (`ConsentApiService`, `SessionsApiService`) wrap typed Axios calls
- GET responses cached (5 min TTL), invalidated on mutations

## Testing Conventions

- **Framework**: Vitest + React Testing Library + jsdom
- **Files**: Co-located `*.test.tsx`, `*.interactive.test.tsx`, `*.integration.test.tsx`
- **Approach**: Test user-visible behavior. Use `getByRole`, `getByText`, `getByLabelText` — avoid `getByTestId`.
- **Mocking**: Mock API services at module level (`vi.mock('@services/api')`). Never mock hooks directly.
- **A11y**: `eslint-plugin-jsx-a11y` enforced. Every interactive component keyboard-navigable.

## Architecture Boundary

Frontend communicates with backend **exclusively through HTTP API endpoints** in `/api/enduser/openapi.yaml`. No Go imports, no shared types. TypeScript types in `types/consent.ts` mirror OpenAPI schemas — keep in sync.
