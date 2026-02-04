/**
 * Design System Border Radius Tokens
 */

export const radius = {
  none: '0',
  sm: '0.125rem', // 2px
  base: '0.25rem', // 4px
  md: '0.375rem', // 6px
  lg: '0.5rem', // 8px
  xl: '0.75rem', // 12px
  '2xl': '1rem', // 16px
  '3xl': '1.5rem', // 24px
  full: '9999px',
} as const;

/**
 * Component-specific radius mapping
 */
export const componentRadius = {
  button: radius.md,
  card: radius.xl,
  input: radius.md,
  badge: radius.base,
  modal: radius['2xl'],
  avatar: radius.full,
} as const;

export type Radius = typeof radius;
export type ComponentRadius = typeof componentRadius;
