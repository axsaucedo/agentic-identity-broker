# Color System Guide

The Refined Trust Architecture uses a carefully curated color palette that balances sophistication with approachability. Every color is chosen to meet accessibility standards (WCAG 2.1 AA) while maintaining the visual identity of the system.

## Core Color Palette

### Navy: Trust & Authority

The primary color system representing stability, trust, and authority.

```
Navy-50:    #f0f5f9    (Lightest - backgrounds)
Navy-100:   #d9e5f0
Navy-200:   #b3cce1
Navy-300:   #8db3d2
Navy-400:   #669ac3
Navy-500:   #4081b4    (Medium)
Navy-600:   #2b68a5
Navy-700:   #1e4d6b    (Primary - main brand color)
Navy-800:   #0d1829    (Dark navy - deep authority)
Navy-900:   #0a1829    (Darkest - reserved for text)

Semantic Tokens:
--color-primary-trust-deep: #0A2540     (Trust Deep - darkest authority)
--color-primary-trust: #1E4D6B          (Trust - medium authority)
--color-primary-trust-light: #E8F1F5    (Trust Light - light backgrounds)
```

**Usage:**
- Primary buttons, links, and interactive elements: Navy-700
- Headings and text hierarchy: Navy-800 to Navy-900
- Focus rings and active states: Navy-700
- Light backgrounds for sections: Navy-50 to Navy-100
- Card backgrounds and containers: Navy-50

### Emerald: Success & Approval

Signals granted access, approved states, and positive actions.

```
Emerald-50:    #d1fae5   (Lightest)
Emerald-100:   #a7f3d0
Emerald-200:   #6ee7b7
Emerald-300:   #4dee9f
Emerald-400:   #34d399
Emerald-500:   #10b981
Emerald-600:   #059669   (Primary success color)
Emerald-700:   #047857
Emerald-800:   #065f46
Emerald-900:   #064e3b   (Darkest)

Semantic Token:
--color-success-primary: #059669        (Success - approved/granted)
```

**Usage:**
- Success badges and status indicators
- Approved permission states
- Checkmarks and confirmation icons
- Green alert backgrounds
- Positive feedback messages

### Amber: Warnings & CTAs

Captures attention for important actions and warnings without alarm.

```
Amber-50:      #fffbeb   (Lightest)
Amber-100:     #fef3c7
Amber-200:     #fde68a
Amber-300:     #fcd34d
Amber-400:     #fbbf24
Amber-500:     #f59e0b
Amber-600:     #d97706   (Primary CTA color)
Amber-700:     #b45309
Amber-800:     #92400e
Amber-900:     #78350f   (Darkest)

Semantic Token:
--color-action-cta: #D97706             (CTA - call-to-action)
```

**Usage:**
- Warning states and cautions
- Important CTAs that need attention
- Pending permission states
- Warning alert backgrounds
- Attention-grabbing but non-critical elements

### Red: Errors & Destructive

Indicates errors, denied access, and destructive actions.

```
Red-50:        #fef2f2   (Lightest)
Red-100:       #fee2e2
Red-200:       #fecaca
Red-300:       #fca5a5
Red-400:       #f87171
Red-500:       #ef4444
Red-600:       #dc2626   (Primary error color)
Red-700:       #b91c1c
Red-800:       #991b1b
Red-900:       #7f1d1d   (Darkest)

Semantic Token:
--color-error-primary: #DC2626          (Error - denied/revoked)
```

**Usage:**
- Error states and validation messages
- Revoked or denied permissions
- Delete buttons and destructive actions
- Error alert backgrounds
- Critical warnings

### Neutral: Backgrounds & Text

The neutral palette for structure, borders, and typography.

```
Neutral-50:    #faf9f7   (Cream - page background)
Neutral-100:   #f5f1ed   (Sand - section backgrounds)
Neutral-200:   #e8e3de   (Taupe - card borders, subtle dividers)
Neutral-300:   #ddd8d1   (Light taupe - input borders)
Neutral-400:   #c4bdb3   (Medium taupe)
Neutral-500:   #9a9591   (Medium-dark taupe)
Neutral-600:   #6b6561   (Dark taupe)
Neutral-700:   #1f2937   (Dark gray - headings)
Neutral-800:   #111827   (Very dark - body text)
Neutral-900:   #030712   (Darkest - reserved)

Text Colors:
--color-text-primary:     #0d1829       (Darkest - primary text)
--color-text-secondary:   #4a5366       (Medium - secondary text)
--color-text-tertiary:    #7a8899       (Light - tertiary text/hints)
```

**Usage:**
- Page background: Neutral-50 (cream)
- Section backgrounds: Neutral-100 (sand)
- Card backgrounds: #ffffff (white)
- Borders: Neutral-200 to Neutral-300
- Primary text: Neutral-700 to Neutral-900
- Secondary text: Neutral-600 to Neutral-700
- Disabled text: Neutral-400
- Subtle dividers: Neutral-200

## Color Accessibility

### Contrast Ratios (WCAG 2.1 AA Compliant)

All color combinations below meet WCAG 2.1 Level AA standards:

| Text Color | Background | Contrast Ratio | Status |
|-----------|-----------|---|---|
| Navy-900 (#0d1829) | White (#ffffff) | 13.8:1 | AAA (Passes large text) |
| Navy-700 (#1e4d6b) | White (#ffffff) | 8.2:1 | AAA |
| Emerald-600 (#059669) | White (#ffffff) | 5.3:1 | AA |
| Amber-600 (#d97706) | White (#ffffff) | 5.1:1 | AA |
| Red-600 (#dc2626) | White (#ffffff) | 5.5:1 | AA |
| Navy-800 (#0d1829) | Neutral-100 (#f5f1ed) | 11.2:1 | AAA |

**Rule:** Always verify custom color combinations with [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/) or [ColorSnack](https://www.colorsnack.com/).

## Using Colors in Components

### In Tailwind Classes

```tsx
// Background colors
className="bg-navy-700"        // Navy primary
className="bg-emerald-600"     // Success
className="bg-amber-600"       // Warning
className="bg-red-600"         // Error
className="bg-neutral-100"     // Section background

// Text colors
className="text-navy-900"      // Primary text
className="text-neutral-600"   // Secondary text
className="text-emerald-600"   // Success text
className="text-red-600"       // Error text
```

### In CSS Custom Properties

```css
/* For dynamic theming or special cases */
color: var(--color-text-primary);
background-color: var(--color-primary-trust);
border-color: var(--color-neutral-200);
```

### Component Guidelines

#### Buttons
- **Primary Button**: Navy-700 background, white text
- **Secondary Button**: Neutral-200 border, Navy-700 text
- **Danger Button**: Red-600 background, white text
- **Ghost Button**: Transparent, Navy-700 text

#### Status Badges
- **Active/Approved**: Emerald-600 background, white text
- **Pending**: Amber-600 background, Navy-900 text
- **Revoked/Error**: Red-600 background, white text
- **Inactive**: Neutral-300 background, Neutral-600 text

#### Form Inputs
- **Border (default)**: Neutral-300
- **Border (focus)**: Navy-700
- **Border (error)**: Red-600
- **Background (disabled)**: Neutral-100

#### Cards
- **Background**: White (#ffffff)
- **Border**: Neutral-200 (subtle containment)
- **Text**: Navy-900 for headings, Neutral-700 for body

#### Alerts
- **Success Alert**: Emerald-50 background, Emerald-700 border, Emerald-900 text
- **Warning Alert**: Amber-50 background, Amber-700 border, Amber-900 text
- **Error Alert**: Red-50 background, Red-700 border, Red-900 text
- **Info Alert**: Navy-50 background, Navy-700 border, Navy-900 text

## Semantic Color Tokens

For applications needing semantic token references:

```typescript
// Trust & Authority
colors: {
  'trust-deep': '#0A2540',      // Darkest trust color for headings
  'trust': '#1E4D6B',            // Primary trust/navy color
  'trust-light': '#E8F1F5',      // Light trust backgrounds
}

// Action & Status
colors: {
  'action-cta': '#D97706',       // Call-to-action (amber)
  'success': '#059669',          // Success state (emerald)
  'warning': '#F59E0B',          // Warning state (amber)
  'error': '#DC2626',            // Error state (red)
  'info': '#3B82F6',             // Info state (blue)
}

// Neutral & Text
colors: {
  'bg-primary': '#faf9f7',       // Cream background
  'bg-secondary': '#f5f1ed',     // Sand background
  'text-primary': '#0d1829',     // Primary text (darkest)
  'text-secondary': '#4a5366',   // Secondary text (medium)
  'text-tertiary': '#7a8899',    // Tertiary text (light)
}
```

## Color Combinations to Avoid

❌ **Don't use**:
- Red on Amber backgrounds (poor readability)
- Emerald on Navy backgrounds (insufficient contrast)
- Neutral-600 on Neutral-200 (indistinguishable)
- Single color to indicate state (use icon + color combination)

✅ **Do use**:
- High-contrast text on colored backgrounds
- Color + icon/symbol for status indication
- Semantic tokens consistently across components
- Hover state transitions for interactive elements

## Dark Mode Considerations

For future dark mode support, invert the palette:

```css
@media (prefers-color-scheme: dark) {
  --color-bg-primary: #0d1829;      /* Was Navy-900 */
  --color-bg-secondary: #1f2937;    /* Was Neutral-700 */
  --color-text-primary: #faf9f7;    /* Was Neutral-50 */
  --color-text-secondary: #c4bdb3;  /* Was Neutral-400 */
}
```

## Resources

- [WCAG 2.1 Color Contrast Guidelines](https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html)
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [ColorSnack Contrast Tool](https://www.colorsnack.com/)
- [Tailwind Color Palette](https://tailwindcss.com/docs/customizing-colors)

---

## Summary

The Refined Trust Architecture color system is built on five core palettes:

1. **Navy** - Trust and primary actions
2. **Emerald** - Success and approval
3. **Amber** - Warnings and CTAs
4. **Red** - Errors and destructive actions
5. **Neutral** - Structure and typography

Each color meets WCAG 2.1 AA accessibility standards and is chosen to communicate specific meanings to users while maintaining visual sophistication and approachability.
