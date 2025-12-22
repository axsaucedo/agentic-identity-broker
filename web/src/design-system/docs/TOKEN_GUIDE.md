# Design Token Guide

Design tokens are the visual design decisions encoded as data. This guide covers the tokens available in the Refined Trust Architecture design system and how to use them.

## Overview

Design tokens in this system are managed through:
- **Tailwind CSS v4**: CSS variables via `@theme` directive
- **CSS Custom Properties**: For runtime customization
- **TypeScript Type System**: For compile-time safety
- **Class Variance Authority (CVA)**: For component variants

## Color System

The color palette is based on semantic meanings, brand identity, and accessibility requirements.

### Semantic Colors

```typescript
// Primary Brand Color
navy-50 through navy-950    // Blue spectrum (brand primary)
navy-700: #0084d1           // Primary action color

// Success (Positive/Approved)
emerald-50 through emerald-950
emerald-700: #059669        // Success/approved indicator

// Warning (Caution/Pending)
amber-50 through amber-950
amber-700: #b45309          // Warning/pending indicator

// Error (Destructive/Denied)
red-50 through red-950
red-700: #dc2626            // Error/denied indicator

// Neutral (Backgrounds/Text)
gray-50 through gray-950
gray-600: #4b5563           // Body text
gray-700: #1f2937           // Heading text
```

### Color Usage

| Use Case | Token | Example |
|----------|-------|---------|
| Primary buttons | navy-700 | `<Button variant="primary">` |
| Success status | emerald-700 | `<GrantStatusBadge status="approved">` |
| Warning/Pending | amber-700 | `<Badge variant="warning">` |
| Error/Denied | red-700 | `<Alert variant="error">` |
| Link hover | navy-600 | Navigation links |
| Disabled state | gray-400 | Inactive form inputs |
| Borders | gray-200 | Card outlines |
| Backgrounds | gray-50 | Page backgrounds |

### Color Accessibility

All semantic colors meet WCAG AA contrast requirements:
- Text contrast: Minimum 4.5:1 (for body text)
- UI component contrast: Minimum 3:1 (for graphics)
- Use ColorSnack or WebAIM for verification

**When choosing colors:**
1. Prefer semantic colors (navy, emerald, amber, red)
2. Check contrast with intended background
3. Consider colorblind accessibility (don't rely on color alone)
4. Test with accessibility tools

## Spacing System

Tailwind's spacing scale follows a consistent 4px base unit (0.25rem).

```typescript
// Core Spacing Scale
0       // 0px       - No space (useful for removing margins)
px      // 1px       - Divider lines
0.5     // 2px       - Very tight spacing
1       // 4px       - xs: Extra small spacing
2       // 8px       - Extra small
3       // 12px      - Small
4       // 16px      - md: Medium (default)
6       // 24px      - lg: Large
8       // 32px      - Extra large
10      // 40px      - XXL
12      // 48px      - XXXL
16      // 64px      - 2XL
```

### Named Spacing (in components)

Components use semantic names that map to this scale:

```typescript
size: 'xs' | 'sm' | 'md' | 'lg' | 'xl'

// Typical mapping
'xs' → 0.5  (2px)   or  2 (8px)    - Compact
'sm' → 3    (12px)                 - Small
'md' → 4    (16px)                 - Default
'lg' → 6    (24px)                 - Large
'xl' → 8    (32px)                 - Extra large
```

### Spacing Usage

```tsx
// Padding (internal spacing)
<Card padding="lg" />           // 24px internal padding
<Button size="sm" />            // Compact button

// Gaps (space between children)
<Stack gap="md" />              // 16px between items
<Grid gap="lg" />               // 24px between grid cells

// Margins (external spacing)
<div className="mb-4" />        // 16px margin bottom
<section className="my-6" />    // 24px margin top/bottom

// Margins within components
'mt-1' → 4px  (top)
'mt-2' → 8px  (top)
'mb-3' → 12px (bottom)
'mb-4' → 16px (bottom)
```

## Typography System

Typography is managed through semantic HTML and Tailwind's text utilities.

### Font Families

```typescript
// Display/Headings (Brand Primary)
'font-display' → Crimson Pro, serif
  - Weight 700 (bold) for main headings
  - Weight 600 (semibold) for subheadings
  - Letter-spacing: -0.02em for authority
  - Usage: h1, h2, h3, page titles, modal headers

// Body/UI (Humanist Sans-Serif)
'font-sans' → Manrope, sans-serif
  - Weight 400 (regular) for body text
  - Weight 500 (medium) for emphasized text
  - Letter-spacing: -0.01em for readability
  - Usage: Body text, buttons, form labels, descriptions
  - Base size: 1rem (16px)
  - Line height: 1.5 for comfortable reading

// Monospace (Technical Values)
'font-mono' → JetBrains Mono, monospace
  - Weight 500 (medium) for technical content
  - Usage: OAuth scopes, agent IDs, API tokens, code blocks
  - Typically displayed with subtle background highlight
```

**Rationale:**
- **Crimson Pro** conveys trust, authority, and seriousness—essential for consent UI
- **Manrope** provides excellent readability and humanist approachability
- **JetBrains Mono** offers clarity for technical values while maintaining visual consistency

### Font Sizes & Line Heights

```typescript
// Tailwind text-* scale
'text-xs'     → 12px  / 1rem      (overline text)
'text-sm'     → 14px  / 1.25rem   (small/secondary)
'text-base'   → 16px  / 1.5rem    (body text - default)
'text-lg'     → 18px  / 1.75rem   (heading 4)
'text-xl'     → 20px  / 1.75rem   (heading 3)
'text-2xl'    → 24px  / 2rem      (heading 2)
'text-3xl'    → 30px  / 2.25rem   (heading 1)
'text-4xl'    → 36px  / 2.25rem   (display)
```

### Font Weights

```typescript
'font-normal'   → 400   (regular text)
'font-medium'   → 500   (emphasized text)
'font-semibold' → 600   (strong emphasis)
'font-bold'     → 700   (headings)
```

### Typography Usage

```tsx
// Semantic HTML + Tailwind
<h1 className="text-4xl font-bold">Main Title</h1>
<h2 className="text-2xl font-bold">Section</h2>
<h3 className="text-xl font-semibold">Subsection</h3>
<p className="text-base font-normal">Body text</p>
<p className="text-sm text-gray-600">Secondary text</p>
<span className="text-xs text-gray-500">Overline</span>

// Component sizing
<Button size="sm" />    // Smaller text inside
<Badge size="md" />     // Default text sizing
<Heading level={2} />   // Semantic heading level
```

## Shadow System

Shadows provide depth and hierarchy in the interface.

```typescript
// Tailwind shadow scale
'shadow-sm'        // Subtle (cards, small elements)
'shadow'           // Default (most interactive elements)
'shadow-md'        // Medium (modals, dropdowns)
'shadow-lg'        // Large (overlays, prominent elements)
'shadow-lg-premium' // Custom premium shadow (design system specific)

// Shadow usage
<Card className="shadow" />           // Default card shadow
<Modal className="shadow-lg" />       // Prominent modal
<button className="hover:shadow-md" /> // Hover elevation
```

## Border Radius

Consistent rounded corners throughout the system.

```typescript
// Tailwind radius scale
'rounded-none'  → 0px       (sharp corners)
'rounded-sm'    → 0.125rem  (1px - very subtle)
'rounded'       → 0.25rem   (4px - default)
'rounded-md'    → 0.375rem  (6px)
'rounded-lg'    → 0.5rem    (8px - prominent)
'rounded-xl'    → 0.75rem   (12px)
'rounded-full'  → 9999px    (circular)

// Typical usage
<Card className="rounded" />           // 4px (default)
<Badge className="rounded-md" />       // 6px
<Avatar className="rounded-lg" />      // 8px
<Button className="rounded-lg" />      // 8px
<Checkbox className="rounded-sm" />    // 1px (checkbox)
```

## Transitions & Animations

Smooth, purposeful animations that enhance usability.

```typescript
// Transition durations
150ms   → Fast interactions (hover effects, small changes)
300ms   → Default (most component animations)
500ms   → Slow (page transitions, large changes)

// Easing functions
'cubic-bezier(0.4, 0, 0.2, 1)'  → Default (smooth)
'cubic-bezier(0.4, 0, 1, 1)'     → Ease-out (enter)
'cubic-bezier(0, 0, 0.2, 1)'     → Ease-in (exit)

// Usage
<Transition duration={300} easing="ease-out">
  <Modal />
</Transition>
```

## Z-Index Scale

Layering strategy for overlays and stacked elements.

```typescript
// Tailwind z-index scale
'z-0'   → 0       (default)
'z-10'  → 10      (tooltips, popovers)
'z-20'  → 20      (dropdowns)
'z-30'  → 30      (modals, important overlays)
'z-40'  → 40      (notification toasts)
'z-50'  → 50      (full-page overlays, critical modals)

// Component defaults
Tooltip  → z-10
Dropdown → z-50
Modal    → z-50
Toast    → z-40
Popover  → z-20
```

## Using Design Tokens in Code

### Via Tailwind Classes (Recommended)

```tsx
// Most common approach
<div className="bg-gray-50 text-gray-700 p-4 rounded-lg shadow">
  <h2 className="text-2xl font-bold text-gray-900">Title</h2>
  <p className="mt-2 text-sm text-gray-600">Description</p>
</div>
```

### Via CSS Variables

```tsx
// If needing dynamic theming
<div style={{
  backgroundColor: 'var(--color-gray-50)',
  color: 'var(--color-gray-700)',
  padding: 'var(--spacing-4)',
  borderRadius: 'var(--radius-lg)',
}}>
  {children}
</div>
```

### Via Component Props

```tsx
// Most semantic approach
<Card padding="lg" hover="lift" backgroundColor="gray-50">
  <heading>Title</heading>
  <p>Description</p>
</Card>

<Stack gap="md" direction="column" align="start">
  <Button variant="primary" size="lg" />
  <Button variant="secondary" size="lg" />
</Stack>
```

## Token Customization

### For Application-Specific Theming

1. **Color overrides** in `tailwind.config.ts`:
```typescript
export default {
  theme: {
    extend: {
      colors: {
        'brand-primary': '#0084d1',
        'brand-secondary': '#6366f1',
      }
    }
  }
}
```

2. **CSS variable overrides**:
```css
:root {
  --color-primary: #0084d1;
  --color-secondary: #6366f1;
}

@media (prefers-color-scheme: dark) {
  :root {
    --color-primary: #1e90ff;
  }
}
```

### Respecting User Preferences

```typescript
// Dark mode support (automatic via Tailwind)
<div className="dark:bg-gray-900 dark:text-white">
  {children}
</div>

// Reduced motion support
<div className="motion-safe:animate-in motion-reduce:animate-none">
  {children}
</div>

// High contrast support (automatic via semantic colors)
```

## Accessibility with Design Tokens

1. **Always verify color contrast** when using custom colors
2. **Use semantic colors** (navy, emerald, amber, red) which are pre-verified
3. **Avoid color-only encoding** - use icons, text, or patterns too
4. **Test with ColorSnack** or WebAIM Contrast Checker
5. **Respect prefers-reduced-motion** in animations

## Token Maintenance

Design tokens are maintained in:
- `tailwind.config.ts` - Tailwind configuration
- Component CVA files - Component-specific variants
- Storybook docs - Visual reference
- This guide - Documentation

When proposing new tokens:
1. Identify the design need
2. Check if existing token works
3. Verify accessibility compliance
4. Document in this guide
5. Update Tailwind config
6. Test across components

## Summary

Design tokens provide:
- **Consistency** across all applications
- **Accessibility** through verified color and sizing choices
- **Flexibility** to customize for brand or context
- **Maintainability** through centralized management
- **Performance** via efficient CSS generation

By using design tokens consistently, we ensure a cohesive, accessible, and maintainable user experience.
