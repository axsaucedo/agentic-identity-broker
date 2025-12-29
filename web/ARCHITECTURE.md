# Frontend Architecture: OAuth2 Sessions Management

Visual architecture and data flow documentation for the OAuth2 sessions frontend.

---

## Component Hierarchy

```
App.tsx
  └── Router (/consent)
      └── Route (/oauth2/sessions)
          └── ThirdPartySessionsPage
              ├── useSessions() hook
              │   └── sessionsApi.listSessions()
              │
              ├── Loading State
              │   └── Grid
              │       └── Skeleton × 3
              │
              ├── Error State
              │   └── Alert + Retry Button
              │
              ├── Empty State
              │   └── EmptyState + Refresh Button
              │
              └── Content State
                  ├── Grid (responsive)
                  │   └── SessionCard[] (map)
                  │       ├── Card
                  │       │   ├── Header
                  │       │   │   ├── Service Name
                  │       │   │   └── Status Badge
                  │       │   ├── Body
                  │       │   │   ├── StatusIndicator (Encrypted)
                  │       │   │   ├── StatusIndicator (Agents)
                  │       │   │   ├── StatusIndicator (Timestamp)
                  │       │   │   └── Scopes (Badge[])
                  │       │   └── Footer
                  │       │       ├── View Details Button
                  │       │       └── Terminate Button
                  │       └── ...more cards
                  │
                  └── Termination Dialog (conditional)
                      ├── Dialog Header
                      ├── Warning (if dependent agents)
                      ├── Error Alert (if failed)
                      └── Actions
                          ├── Cancel Button
                          └── Confirm Button
```

---

## Data Flow

### 1. Initial Page Load

```
User navigates to /oauth2/sessions
    ↓
ThirdPartySessionsPage mounts
    ↓
useSessions() hook executes
    ↓
useState initializes with loading=true
    ↓
useEffect triggers fetchData()
    ↓
sessionsApi.listSessions() called
    ↓
Check apiCache (2min TTL)
    ↓
├─ Cache HIT: Return cached data
│       ↓
│   setState({ sessions, loading: false, error: null })
│       ↓
│   Component re-renders with data
│
└─ Cache MISS: Fetch from backend
        ↓
    apiClient.get('/api/third-party/sessions')
        ↓
    ├─ SUCCESS:
    │   ├─ Cache response (2min TTL)
    │   ├─ setState({ sessions, loading: false, error: null })
    │   └─ Component re-renders with data
    │
    └─ ERROR:
        ├─ Parse error message
        ├─ setState({ sessions: [], loading: false, error: message })
        └─ Component re-renders with error state
```

### 2. Session Termination Flow

```
User clicks "Terminate" on SessionCard
    ↓
onTerminate(serviceId) called
    ↓
ThirdPartySessionsPage.handleTerminate()
    ↓
Set selectedSessionId state
    ↓
Termination Dialog appears
    ↓
User clicks "Terminate Session"
    ↓
handleTerminationConfirm() called
    ↓
setTerminatingLoading(true)
    ↓
sessionsApi.terminateSession(serviceId)
    ↓
apiClient.delete('/api/third-party/:service-id/session')
    ↓
├─ SUCCESS:
│   ├─ apiCache.invalidatePattern('/third-party/*')
│   ├─ Close dialog
│   ├─ refetch() - fetch fresh data
│   └─ User sees updated list
│
└─ ERROR:
    ├─ Extract error message
    ├─ setTerminationError(message)
    ├─ Show error in dialog
    └─ User can retry or cancel
```

### 3. Refetch Flow

```
User clicks "Refresh" or "Retry"
    ↓
refetch() called
    ↓
setState({ ...prev, loading: true, error: null })
    ↓
sessionsApi.listSessions() (bypasses cache)
    ↓
apiClient.get('/api/third-party/sessions')
    ↓
├─ SUCCESS:
│   ├─ Update cache
│   ├─ setState({ sessions, loading: false, error: null })
│   └─ User sees fresh data
│
└─ ERROR:
    ├─ setState({ sessions: [], loading: false, error: message })
    └─ User sees error state
```

---

## State Management

### Component State (ThirdPartySessionsPage)

```typescript
// useSessions hook state
{
  sessions: SessionSummary[];    // Array of session objects
  loading: boolean;              // Global loading state
  error: string | null;          // Global error message
  refetch: () => Promise<void>;  // Manual refetch function
}

// Local component state
{
  selectedSessionId: string | null;     // Dialog open state
  terminatingLoading: boolean;          // Termination in progress
  terminationError: string | null;      // Termination error message
}
```

### API Cache State

```typescript
// In-memory cache (apiCache)
{
  '/third-party/sessions': {
    data: SessionSummary[];
    timestamp: number;
    ttl: 120000;  // 2 minutes
  }
}

// Cache operations:
- get(key): Check if cached and not expired
- set(key, data, ttl): Store with expiration
- invalidate(key): Remove specific entry
- invalidatePattern(pattern): Remove all matching entries
```

---

## Type System

### Core Types

```typescript
// API Response
interface ListSessionsResponse {
  data: {
    sessions: SessionSummary[];
  }
}

// Session Data
interface SessionSummary {
  id: string;                          // Unique identifier
  service_id: string;                  // Service key (google, github, etc.)
  service_display_name: string;        // Human-readable name
  token_type: string;                  // OAuth2 token type (Bearer)
  scope: string[];                     // Granted scopes
  initiated_at: string;                // ISO 8601 timestamp
  is_expired: boolean;                 // Full expiration flag
  access_token_expired: boolean;       // Access token only
  refresh_token_expires_at?: string;   // ISO 8601 timestamp (optional)
  dependent_agent_count: number;       // Number of dependent agents
  is_encrypted: boolean;               // Encryption flag
}

// Component Props
interface SessionCardProps {
  session: SessionSummary;
  onTerminate: (serviceId: string) => void;
  onViewDetails?: (serviceId: string) => void;
  loading?: boolean;
}

// Hook Return
interface UseSessionsReturn {
  sessions: SessionSummary[];
  loading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
}
```

---

## API Integration

### Request Pipeline

```
Component
    ↓
Hook (useSessions)
    ↓
API Client (sessionsApi)
    ↓
Cache Layer (apiCache)
    ↓ (cache miss)
HTTP Client (apiClient/axios)
    ↓
Request Interceptor
    ↓
Backend API (Go server)
```

### Response Pipeline

```
Backend API
    ↓
Response Interceptor
    ↓ (error handling)
HTTP Client (axios)
    ↓
Cache Layer (store)
    ↓
API Client (sessionsApi)
    ↓
Hook (useSessions)
    ↓
Component (setState)
    ↓
React Re-render
```

### Error Handling

```
API Error
    ↓
Response Interceptor
    ↓
├─ 401 Unauthorized → "Authentication failed"
├─ 403 Forbidden → "Insufficient permissions"
├─ 404 Not Found → "Resource not found"
├─ 500+ Server Error → "Service unavailable"
└─ Network Error → "Connection error"
    ↓
Promise.reject(ApiError)
    ↓
Hook catch block
    ↓
setState({ error: message })
    ↓
Component renders error state
```

---

## Design System Integration

### Component Tree

```
ThirdPartySessionsPage
├── Container (max-w-7xl mx-auto)
│   └── Stack (vertical, gap-lg)
│       ├── Header (text)
│       └── Grid (responsive, gap-md)
│           └── SessionCard[]
│               └── Card (padding-default, border-subtle, hover-lift)
│                   ├── Header
│                   │   └── Stack (horizontal, gap-md, justify-between)
│                   │       ├── Service Info (text)
│                   │       └── Badge (status, showDot)
│                   ├── Body
│                   │   └── Stack (vertical, gap-sm)
│                   │       ├── StatusIndicator (icon + label)
│                   │       ├── StatusIndicator (icon + label)
│                   │       ├── StatusIndicator (icon + label)
│                   │       └── Scopes
│                   │           └── Stack (horizontal, wrap)
│                   │               └── Badge[] (neutral, size-sm)
│                   └── Footer
│                       └── Stack (horizontal, gap-md, justify-end)
│                           ├── Button (outline, size-sm)
│                           └── Button (danger, size-sm)
```

### Styling Approach

**No Custom CSS** - All styling via:
1. **Design System Components**: Pre-built with variants
2. **Tailwind Utility Classes**: For layout and spacing
3. **Semantic Color Tokens**: From design system

Example:
```tsx
// ✅ Good: Design system + utilities
<div className="max-w-7xl mx-auto py-8 px-4">
  <Stack gap="lg">
    <Card padding="default" border="subtle" hover="lift">
      <Badge variant="success" showDot>Active</Badge>
    </Card>
  </Stack>
</div>

// ❌ Bad: Custom CSS
<div className="custom-container">
  <div className="custom-stack">
    <div className="custom-card">
      <span className="custom-badge">Active</span>
    </div>
  </div>
</div>
```

---

## Performance Optimizations

### 1. Code Splitting

```
App.tsx
    ↓
React.lazy(() => import('./pages/ThirdPartySessionsPage'))
    ↓
Separate bundle: ThirdPartySessionsPage.chunk.js
    ↓
Loaded only when route accessed
```

### 2. API Caching

```
First Request
    ↓
API Call → Cache MISS
    ↓
Fetch from backend (300ms)
    ↓
Store in cache with 2min TTL

Subsequent Requests (within 2min)
    ↓
API Call → Cache HIT
    ↓
Return cached data (instant)
```

### 3. Memoization

```tsx
// SessionCard: Status calculation
const { status, variant } = useMemo(() => {
  // Complex logic here
  return { status, variant };
}, [session]);  // Only recalculate if session changes

// SessionCard: Date formatting
const initiatedDate = useMemo(() => {
  return new Intl.DateTimeFormat(...).format(new Date(session.initiated_at));
}, [session.initiated_at]);  // Only reformat if date changes
```

### 4. Skeleton Loading

```
Initial Load
    ↓
Show 3 skeleton cards immediately (no spinner)
    ↓
Data arrives
    ↓
Fade transition to real cards
```

---

## Accessibility Architecture

### Semantic HTML Structure

```html
<div>                          <!-- Page container -->
  <div>                        <!-- Content wrapper -->
    <h1>Third-Party Sessions</h1>
    <p>Description...</p>
    <div>                      <!-- Grid container -->
      <div>                    <!-- Card 1 -->
        <h3>Service Name</h3>  <!-- Session heading -->
        <div role="status">    <!-- Status badge -->
        <button>View Details</button>
        <button>Terminate</button>
      </div>
      <!-- More cards... -->
    </div>
  </div>

  <!-- Dialog (when open) -->
  <div role="dialog" aria-modal="true" aria-labelledby="dialog-title">
    <h2 id="dialog-title">Terminate Session?</h2>
    <button>Cancel</button>
    <button>Confirm</button>
  </div>
</div>
```

### Keyboard Navigation Flow

```
Page Load → Focus on document
    ↓
Tab → First SessionCard "View Details" button
    ↓
Tab → "Terminate" button (same card)
    ↓
Tab → Next SessionCard "View Details"
    ↓
... continue through all cards ...
    ↓
Shift+Tab → Navigate backwards

Terminate Button Activated:
    ↓
Dialog Opens → Focus trapped in dialog
    ↓
Tab → "Cancel" button
    ↓
Tab → "Terminate Session" button
    ↓
Tab → Wraps to "Cancel" (focus trap)
    ↓
Escape or Cancel → Dialog closes, focus returns to Terminate button
```

### Screen Reader Announcements

```
Page Load:
  "Third-Party Sessions, heading level 1"
  "Manage your OAuth2 sessions..."

Loading State:
  "Loading... status, polite"

Error State:
  "Failed to Load Sessions, alert"
  "Network error. Retry button."

Session Card:
  "Google Drive, heading level 3"
  "Active status"
  "Encrypted status"
  "2 agents"
  "View Details button"
  "Terminate button"

Dialog Open:
  "Terminate Session? dialog"
  "Cancel button"
  "Terminate Session button"
```

---

## Testing Architecture

### Test Pyramid

```
                      /\
                     /  \
                    /E2E \      (Manual - Future)
                   /______\
                  /        \
                 / Integr-  \   (Hook tests with mocked API)
                /   ation    \
               /              \
              /______________  \
             /                  \
            /   Component Tests  \  (18 tests - SessionCard)
           /     (Unit Tests)     \
          /                        \
         /_________________________ \

Total: 33 automated tests
```

### Test Coverage Map

```
SessionCard.test.tsx (18 tests)
├── Rendering (8 tests)
│   ├── Basic information display
│   ├── Status badge variants (4)
│   ├── Agent count (singular/plural)
│   └── Scopes display
├── Actions (6 tests)
│   ├── Button rendering
│   ├── onTerminate callback
│   ├── onViewDetails callback
│   ├── Optional button handling
│   └── Expired state
├── Loading State (2 tests)
│   ├── Disabled buttons
│   └── Loading spinner
└── Accessibility (4 tests)
    ├── Semantic HTML
    ├── Button labels
    ├── Keyboard navigation
    └── Date formatting

useSessions.test.ts (15 tests)
├── Data Fetching (2 tests)
│   ├── Initial fetch
│   └── Empty list
├── Error Handling (4 tests)
│   ├── Generic errors
│   ├── 404 errors
│   ├── API error responses
│   └── Unknown errors
├── Refetch (3 tests)
│   ├── Basic refetch
│   ├── Error recovery
│   └── Loading state
├── Loading State (3 tests)
│   ├── Initial loading
│   ├── Success loading
│   └── Error loading
└── useSession helper (3 tests)
    ├── Find by service ID
    ├── Not found
    └── Loading state
```

---

## Security Architecture

### Authentication Flow

```
User Request
    ↓
Vite Dev Server (dev only)
│   └─ Inject: X-Remote-User: dev@example.com
    ↓
Backend API (Go)
│   └─ Verify: X-Remote-User header
    ↓
Session Lookup
    ↓
Response (filtered by user)
```

**Production:**
```
User Request
    ↓
oauth2-proxy / nginx
│   └─ Verify: OAuth2 token
│   └─ Set: X-Remote-User header
    ↓
Backend API (Go)
│   └─ Trust: X-Remote-User header (internal network only)
    ↓
Session Lookup
    ↓
Response (filtered by user)
```

### Data Security

**Frontend:**
- ✅ No tokens in localStorage
- ✅ No tokens in sessionStorage
- ✅ No tokens in Redux/state
- ✅ All data from authenticated API

**API Client:**
- ✅ HTTPS enforced in production
- ✅ Authentication headers only
- ✅ No credential storage
- ✅ Secure cookie handling (SameSite, Secure)

**Backend:**
- ✅ Session tokens encrypted at rest
- ✅ Refresh tokens encrypted
- ✅ Access control per user
- ✅ Audit logging

---

## Future Enhancements

### Phase 4 Integration Points

```
ThirdPartySessionsPage
    ↓
1. Add "Connect New Service" button
    ↓
    ServiceSelectionDialog
        ↓
        OAuth2AuthorizationFlow
            ↓
            Redirect to provider
            ↓
            Callback handler
            ↓
            Refresh sessions list

2. Add "Refresh Token" action
    ↓
    SessionCard
        ↓
        onClick={() => sessionsApi.refreshSession(serviceId)}
        ↓
        Show success toast
        ↓
        Update card status

3. Add "View Details" link
    ↓
    SessionDetailPage
        ├── Full token metadata
        ├── Dependent agents list
        ├── Audit log
        └── Back to list button
```

---

## Deployment Architecture

### Build Process

```
Development:
    Source (TSX) → Vite Dev Server → Browser
        ↓
    Hot Module Replacement (instant updates)

Production:
    Source (TSX)
        ↓
    TypeScript Compiler (tsc)
        ↓
    Vite Build
        ├─ Tree shaking
        ├─ Code splitting
        ├─ Minification
        └─ Asset optimization
        ↓
    dist/consent/
        ├─ index.html
        ├─ assets/
        │   ├─ main.js (50KB gzipped)
        │   ├─ ThirdPartySessionsPage.chunk.js (15KB)
        │   └─ styles.css
        └─ favicon.ico
        ↓
    Embedded in Go binary (embed.FS)
        ↓
    Served by chi router (/consent/*)
```

---

**Architecture Complete** ✅
