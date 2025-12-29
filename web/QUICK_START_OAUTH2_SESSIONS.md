# Quick Start: OAuth2 Sessions Frontend

A quick reference guide for developers working with the OAuth2 sessions frontend.

## Getting Started

### Start Development Server

```bash
# Terminal 1: Backend
just run

# Terminal 2: Frontend with hot reload
just web-dev
```

Access: http://localhost:3000/consent/oauth2/sessions

### Run Tests

```bash
cd web

# All tests
npm test

# Watch mode
npm test -- --watch

# Specific test
npm test -- --run src/components/sessions/SessionCard.test.tsx
```

### Build for Production

```bash
just web-build
just run
```

Access: http://localhost:8000/consent/oauth2/sessions

---

## Component Usage

### SessionCard

```tsx
import { SessionCard } from '@components/sessions';

function MyComponent() {
  const handleTerminate = (serviceId: string) => {
    console.log('Terminate:', serviceId);
  };

  return (
    <SessionCard
      session={session}
      onTerminate={handleTerminate}
      onViewDetails={(id) => navigate(`/details/${id}`)}
      loading={isTerminating}
    />
  );
}
```

### useSessions Hook

```tsx
import { useSessions } from '@hooks';

function MyPage() {
  const { sessions, loading, error, refetch } = useSessions();

  if (loading) return <Spinner />;
  if (error) return <Alert>{error}</Alert>;

  return (
    <div>
      {sessions.map(session => (
        <SessionCard key={session.id} session={session} />
      ))}
      <Button onClick={refetch}>Refresh</Button>
    </div>
  );
}
```

### Sessions API

```tsx
import { sessionsApi } from '@services/api';

// List all sessions
const sessions = await sessionsApi.listSessions();

// Get session details
const details = await sessionsApi.getSessionDetails('google');

// Terminate session
await sessionsApi.terminateSession('google');

// Refresh session
const updated = await sessionsApi.refreshSession('google');
```

---

## API Response Format

```json
{
  "data": {
    "sessions": [
      {
        "id": "uuid-here",
        "service_id": "google",
        "service_display_name": "Google Drive",
        "token_type": "Bearer",
        "scope": ["read:email", "write:files"],
        "initiated_at": "2024-01-01T12:00:00Z",
        "is_expired": false,
        "access_token_expired": false,
        "refresh_token_expires_at": "2024-12-31T23:59:59Z",
        "dependent_agent_count": 2,
        "is_encrypted": true
      }
    ]
  }
}
```

---

## Status Badge Logic

```typescript
// Active (green)
!is_expired && !access_token_expired && refresh_token_expires_at > 7 days

// Expiring Soon (amber)
!is_expired && !access_token_expired && refresh_token_expires_at < 7 days

// Access Token Expired (amber)
!is_expired && access_token_expired

// Expired (red)
is_expired
```

---

## Design System Components

**Approved Components:**
- `Card` - Session container
- `Button` - Actions
- `Badge` - Status indicators
- `StatusIndicator` - Metadata (icon + label)
- `Stack` - Flexbox layouts
- `Grid` - Responsive grid
- `Skeleton` - Loading states
- `Alert` - Error messages
- `EmptyState` - No data

**Semantic Colors:**
- `success-primary` - Active sessions
- `warning-primary` - Expiring/expired tokens
- `error-primary` - Fully expired
- `neutral-50` to `neutral-900` - Text/backgrounds
- `trust-deep` - Primary brand

---

## Common Tasks

### Add New Session Status

1. Update status calculation in `SessionCard.tsx`:
```tsx
const { status, variant } = useMemo(() => {
  // Add new condition here
  if (session.someNewField) {
    return { status: 'New Status', variant: 'info' as const };
  }
  // ... existing logic
}, [session]);
```

2. Add test case in `SessionCard.test.tsx`:
```tsx
it('renders "New Status" for new condition', () => {
  const session = createMockSession({ someNewField: true });
  render(<SessionCard session={session} onTerminate={vi.fn()} />);
  expect(screen.getByText('New Status')).toBeInTheDocument();
});
```

### Add New Session Metadata

1. Update `SessionSummary` interface in `sessions.ts`:
```tsx
export interface SessionSummary {
  // ... existing fields
  new_field: string;
}
```

2. Display in `SessionCard.tsx`:
```tsx
<StatusIndicator
  icon={<NewIcon />}
  label={session.new_field}
/>
```

### Add New Action Button

1. Add handler to `SessionCard.tsx`:
```tsx
interface SessionCardProps {
  // ... existing props
  onNewAction?: (serviceId: string) => void;
}

// In render:
<Button onClick={() => onNewAction?.(session.service_id)}>
  New Action
</Button>
```

2. Add test:
```tsx
it('calls onNewAction when button clicked', async () => {
  const onNewAction = vi.fn();
  render(<SessionCard session={session} onNewAction={onNewAction} />);

  await user.click(screen.getByText('New Action'));
  expect(onNewAction).toHaveBeenCalledWith('google');
});
```

---

## Troubleshooting

### "Failed to fetch sessions"

**Check:**
1. Backend is running: `curl http://localhost:8000/health`
2. API endpoint exists: `curl http://localhost:8000/api/third-party/sessions`
3. Vite proxy configured: Check `vite.config.ts`
4. Network tab in DevTools for actual error

### "Module not found" errors

**Fix:**
```bash
cd web
npm install
```

### Tests failing with "act()" warnings

**Normal behavior** - React Testing Library warnings that don't affect functionality.

**To suppress:**
```tsx
import { act } from '@testing-library/react';

await act(async () => {
  await result.current.refetch();
});
```

### Styles not applying

**Check:**
1. Tailwind classes are valid
2. Design system component props are correct
3. PostCSS is running (dev server restart)

---

## File Locations

```
web/src/
├── components/sessions/
│   ├── SessionCard.tsx           # Main card component
│   └── SessionCard.test.tsx      # Component tests
├── pages/
│   └── ThirdPartySessionsPage.tsx # Main page
├── hooks/
│   ├── useSessions.ts            # Data fetching hook
│   └── useSessions.test.ts       # Hook tests
└── services/api/
    └── sessions.ts               # API client
```

---

## Resources

- **Design System**: `/web/src/design-system/`
- **Component Docs**: `/web/src/components/sessions/README.md`
- **Implementation Summary**: `/FRONTEND_IMPLEMENTATION_SUMMARY.md`
- **Storybook**: `npm run storybook` (port 6006)

---

## Need Help?

1. Check `/web/src/components/sessions/README.md` for detailed docs
2. Review test files for usage examples
3. Look at existing consent components for patterns
4. Check design system components in `/web/src/design-system/`

---

## Pre-commit Checklist

```bash
just check  # Runs fmt, vet, lint, test
```

Or individually:
```bash
just fmt      # Format code
just vet      # Static analysis
just lint     # Linting
just test     # Run tests
```

---

**Happy coding!** 🚀
