# Migration Guide: Moving to the Design System

This guide helps developers migrate existing components and code to use the new Refined Trust Architecture design system.

## Overview

The design system provides:
- **30+ production-ready components** (feedback, input, layout, navigation, data display, overlays, advanced)
- **Consistent styling** via Tailwind CSS v4 and design tokens
- **Accessibility built-in** (WCAG 2.1 AA compliant)
- **TypeScript support** with full type safety
- **Storybook documentation** with interactive examples

## Migration Phases

### Phase 1: Evaluation (Week 1)
- Audit existing components
- Identify candidates for replacement
- Plan migration order
- Set up environment

### Phase 2: Replace UI Primitives (Week 2-3)
- Replace buttons, links, inputs
- Replace form components
- Update basic styling

### Phase 3: Replace Complex Components (Week 3-4)
- Replace modals, dropdowns, menus
- Replace navigation patterns
- Replace data display components

### Phase 4: Refactor Application Components (Week 4-5)
- Update domain-specific components
- Use composition patterns
- Remove custom styling

### Phase 5: Polish & Optimize (Week 5-6)
- Test accessibility
- Optimize performance
- Update documentation

## Step-by-Step Migration

### Step 1: Install/Update Dependencies

```bash
# Ensure Tailwind CSS v4 is installed
npm install -D tailwindcss@4.0

# Ensure design system is available
npm install @design-system  # or local path alias
```

### Step 2: Update Tailwind Config

```typescript
// tailwind.config.ts
import type { Config } from 'tailwindcss'

export default {
  content: [
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      // Add custom tokens if needed
    },
  },
  plugins: [],
} satisfies Config
```

### Step 3: Import from Design System

```typescript
// Old way
import { CustomButton } from './components/ui/CustomButton'
import { CustomInput } from './components/ui/CustomInput'

// New way
import { Button } from '@design-system/components/primitives/Button'
import { TextInput } from '@design-system/components/inputs/TextInput'
```

## Component Migration Examples

### Buttons

**Before (Custom Button):**
```tsx
<button
  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
  onClick={handleClick}
  disabled={isLoading}
>
  {isLoading ? 'Loading...' : 'Submit'}
</button>
```

**After (Design System Button):**
```tsx
import { Button } from '@design-system/components/primitives/Button'

<Button
  onClick={handleClick}
  disabled={isLoading}
  loading={isLoading}
>
  Submit
</Button>
```

### Form Inputs

**Before (Custom Input):**
```tsx
<div className="flex flex-col gap-2">
  <label htmlFor="email">Email</label>
  <input
    id="email"
    type="email"
    className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
    placeholder="you@example.com"
    value={email}
    onChange={(e) => setEmail(e.target.value)}
  />
  {error && <span className="text-red-600 text-sm">{error}</span>}
</div>
```

**After (Design System TextInput):**
```tsx
import { TextInput } from '@design-system/components/inputs/TextInput'

<TextInput
  id="email"
  label="Email"
  type="email"
  placeholder="you@example.com"
  value={email}
  onChange={(e) => setEmail(e.target.value)}
  errorMessage={error}
/>
```

### Cards

**Before (Custom Card):**
```tsx
<div className="bg-white rounded-lg shadow-md p-6 border border-gray-200">
  <h3 className="text-lg font-semibold">{title}</h3>
  <p className="text-gray-600 mt-2">{description}</p>
</div>
```

**After (Design System Card):**
```tsx
import { Card } from '@design-system/components/data-display/Card'

<Card padding="lg">
  <h3 className="text-lg font-semibold">{title}</h3>
  <p className="text-gray-600 mt-2">{description}</p>
</Card>
```

### Modals

**Before (Custom Modal):**
```tsx
{isOpen && (
  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div className="bg-white rounded-lg shadow-lg max-w-md p-6">
      <h2 className="text-xl font-bold">{title}</h2>
      <div className="mt-4">{children}</div>
      <div className="mt-6 flex gap-2 justify-end">
        <button onClick={onClose}>Cancel</button>
        <button onClick={onConfirm}>Confirm</button>
      </div>
    </div>
  </div>
)}
```

**After (Design System Modal):**
```tsx
import { Modal } from '@design-system/components/overlays/Modal'

<Modal
  isOpen={isOpen}
  onClose={onClose}
  title={title}
>
  {children}
  <div className="flex gap-2 justify-end mt-6">
    <Button variant="secondary" onClick={onClose}>Cancel</Button>
    <Button variant="primary" onClick={onConfirm}>Confirm</Button>
  </div>
</Modal>
```

### Navigation

**Before (Custom Navigation):**
```tsx
<div className="flex items-center gap-4 px-6 py-4 bg-gray-900 text-white">
  <div className="text-xl font-bold">Logo</div>
  <nav className="flex gap-6 ml-auto">
    <a href="/home" className="hover:text-gray-300">Home</a>
    <a href="/settings" className="hover:text-gray-300">Settings</a>
    <a href="/logout" className="hover:text-gray-300">Logout</a>
  </nav>
</div>
```

**After (Design System AppLayout + Navigation):**
```tsx
import { AppLayout } from '@design-system/components/layout/AppLayout'
import { Stack } from '@design-system/components/layout/Stack'

<AppLayout
  header={
    <Stack direction="row" gap="lg" align="center" className="px-6 py-4 bg-navy-900">
      <div className="text-xl font-bold text-white">Logo</div>
      <nav className="ml-auto flex gap-6">
        <Link href="/home">Home</Link>
        <Link href="/settings">Settings</Link>
        <Link href="/logout">Logout</Link>
      </nav>
    </Stack>
  }
>
  {children}
</AppLayout>
```

## Migration Checklist

### Before Starting Migration

- [ ] Design system is imported correctly
- [ ] Tailwind CSS v4 configured
- [ ] Team trained on design system
- [ ] Storybook set up and running
- [ ] Migration plan documented
- [ ] Testing strategy defined

### During Migration

For each component being migrated:

- [ ] Read Storybook documentation
- [ ] Identify exact replacement component
- [ ] Map old props to new props
- [ ] Update import statements
- [ ] Remove custom CSS/styling
- [ ] Update tests
- [ ] Test accessibility (keyboard, screen reader)
- [ ] Check responsive behavior
- [ ] Verify visual appearance

### After Migration

- [ ] Run full test suite
- [ ] Audit Lighthouse score
- [ ] Test with screen reader
- [ ] Get design review
- [ ] Update component documentation
- [ ] Record migration in changelog

## Common Migration Patterns

### Pattern 1: Custom Layout → Stack Component

```tsx
// Old
<div className="flex flex-col gap-4">
  <div className="flex justify-between items-center">
    <h2>{title}</h2>
    <button>Close</button>
  </div>
  {children}
</div>

// New
<Stack gap="md">
  <Stack direction="row" justify="space-between" align="center">
    <h2>{title}</h2>
    <Button>Close</Button>
  </Stack>
  {children}
</Stack>
```

### Pattern 2: Custom Styling → Design System Props

```tsx
// Old
<div className="px-6 py-4 bg-gray-50 rounded-lg shadow border border-gray-200">
  Content
</div>

// New
<Card
  padding="lg"
  border="subtle"
  hover="lift"
  backgroundColor="gray-50"
>
  Content
</Card>
```

### Pattern 3: Custom Form → Design System Components

```tsx
// Old
<form onSubmit={handleSubmit}>
  <div className="mb-4">
    <label htmlFor="name" className="block mb-1 font-medium">Name</label>
    <input
      id="name"
      type="text"
      className="w-full px-3 py-2 border border-gray-300 rounded-md"
      value={name}
      onChange={(e) => setName(e.target.value)}
    />
  </div>
  <button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded-md">
    Submit
  </button>
</form>

// New
<Stack as="form" gap="md" onSubmit={handleSubmit}>
  <TextInput
    id="name"
    label="Name"
    type="text"
    value={name}
    onChange={(e) => setName(e.target.value)}
  />
  <Button type="submit">Submit</Button>
</Stack>
```

## Testing During Migration

### Unit Tests

```typescript
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from '@design-system/components/primitives/Button'

test('migrated button works correctly', async () => {
  const handleClick = vi.fn()
  render(<Button onClick={handleClick}>Click me</Button>)

  const button = screen.getByRole('button', { name: /click me/i })
  await userEvent.click(button)

  expect(handleClick).toHaveBeenCalledOnce()
})
```

### Accessibility Tests

```typescript
import { axe } from 'jest-axe'

test('migrated component is accessible', async () => {
  const { container } = render(<MigratedComponent />)
  const results = await axe(container)
  expect(results).toHaveNoViolations()
})
```

### Visual Tests

```bash
# Use Storybook visual regression testing
npm run test:visual

# Compare screenshots before/after migration
```

## Performance Optimization

After migration, optimize performance:

```typescript
// 1. Use React.memo for expensive components
export const MigratedComponent = React.memo(({ data }) => (
  <Card>{data.name}</Card>
))

// 2. Memoize callbacks
const handleChange = useCallback((value) => {
  setData(value)
}, [])

// 3. Use useMemo for expensive computations
const processedData = useMemo(() => {
  return data.map(transform)
}, [data])

// 4. Lazy load components
const HeavyComponent = lazy(() => import('./HeavyComponent'))
```

## Rollback Strategy

If issues arise during migration:

1. **Keep old code in feature branch** for easy rollback
2. **Tag release points** before major migrations
3. **Feature flag migrations** to gradual rollout
4. **Monitor error logs** for compatibility issues
5. **Have backup plan** to revert if needed

```tsx
// Use feature flag for gradual rollout
const useNewButton = useFeatureFlag('use-new-design-system-button')

if (useNewButton) {
  return <Button />
} else {
  return <LegacyButton />
}
```

## Documentation Updates

After migrating components:

1. **Update component documentation**
   - Link to Storybook examples
   - Document props and usage
   - Include accessibility notes

2. **Update team wiki/handbook**
   - Recommended patterns
   - Common pitfalls
   - Troubleshooting guide

3. **Record in changelog**
   - What changed
   - Migration path
   - Timeline

Example:
```markdown
## v2.0.0 - Design System Migration

### Changed
- Button component moved to design system
- Button props slightly different (loading vs spinner)

### Migration Path
```migrate-button.md

### Timeline
- v2.0.0: Both old and new buttons available
- v2.1.0: Old button deprecated
- v3.0.0: Old button removed
```

## Common Issues & Solutions

### Issue: Props Don't Match

**Solution:** Check Storybook or component documentation for correct prop names.

```tsx
// Old
<Button loading={true} />

// New (if prop name changed)
<Button disabled={true} aria-busy="true" />
```

### Issue: Styling Doesn't Match

**Solution:** Use design system props and classes consistently.

```tsx
// ✗ Mixing styles
<Card className="custom-color-class">

// ✓ Use design system only
<Card backgroundColor="blue-50">
```

### Issue: TypeScript Errors

**Solution:** Install types and ensure imports are correct.

```bash
npm install @types/design-system
# or for local alias
npm install --save-dev typescript
```

## Success Metrics

Track migration success:

- **Test coverage**: Should maintain >80%
- **Accessibility score**: Should reach WCAG AA
- **Lighthouse score**: Should improve or stay same
- **Bundle size**: May reduce due to shared components
- **Development speed**: Should improve with component reuse
- **Time to migrate**: Track to estimate remaining work

## Timeline Example

```
Week 1: Evaluation & Planning
  - Day 1-2: Audit existing components
  - Day 3-4: Plan migration order
  - Day 5: Team training on design system

Week 2-3: Core Components
  - Migrate: Buttons, Links, Inputs
  - Migrate: Checkboxes, Radios, Toggles

Week 3-4: Complex Components
  - Migrate: Modals, Dropdowns, Navigation
  - Migrate: Tables, Pagination

Week 4-5: Application Components
  - Refactor: Domain-specific components
  - Update: Tests and documentation

Week 5-6: Polish
  - Performance optimization
  - Accessibility audit
  - Final testing & deployment
```

## Resources

- [Design System Documentation](./README.md)
- [Storybook](http://localhost:6006)
- [Composition Patterns](./COMPOSITION_PATTERNS.md)
- [Accessibility Guide](./ACCESSIBILITY_GUIDE.md)
- [Design Principles](./DESIGN_PRINCIPLES.md)

## Support & Questions

- Check Storybook documentation for component usage
- Review this migration guide for common patterns
- Ask team lead for guidance on complex cases
- Open GitHub issue for design system bugs

## Conclusion

Migration to the design system:
- ✅ Improves consistency
- ✅ Reduces custom code
- ✅ Enhances accessibility
- ✅ Speeds up development
- ✅ Maintains component quality

Follow this guide, take it step-by-step, and don't hesitate to ask for help. Happy migrating!
