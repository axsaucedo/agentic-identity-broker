/**
 * Design System Color Tokens (TypeScript)
 * Semantic color tokens that map to CSS custom properties
 */

export const semanticColors = {
  primary: {
    trustDeep: 'var(--color-primary-trust-deep)',
    trust: 'var(--color-primary-trust)',
    trustLight: 'var(--color-primary-trust-light)',
    trustHover: 'var(--color-primary-trust-hover)',
  },
  action: {
    cta: 'var(--color-action-cta)',
    ctaHover: 'var(--color-action-cta-hover)',
    ctaActive: 'var(--color-action-cta-active)',
    ctaDisabled: 'var(--color-action-cta-disabled)',
  },
  success: {
    primary: 'var(--color-success-primary)',
    light: 'var(--color-success-light)',
    dark: 'var(--color-success-dark)',
  },
  error: {
    primary: 'var(--color-error-primary)',
    light: 'var(--color-error-light)',
    dark: 'var(--color-error-dark)',
  },
  warning: {
    primary: 'var(--color-warning-primary)',
    light: 'var(--color-warning-light)',
    dark: 'var(--color-warning-dark)',
  },
  info: {
    primary: 'var(--color-info-primary)',
    light: 'var(--color-info-light)',
    dark: 'var(--color-info-dark)',
  },
  text: {
    primary: 'var(--color-text-primary)',
    secondary: 'var(--color-text-secondary)',
    tertiary: 'var(--color-text-tertiary)',
    disabled: 'var(--color-text-disabled)',
    inverse: 'var(--color-text-inverse)',
  },
  bg: {
    primary: 'var(--color-bg-primary)',
    secondary: 'var(--color-bg-secondary)',
    tertiary: 'var(--color-bg-tertiary)',
    elevated: 'var(--color-bg-elevated)',
    overlay: 'var(--color-bg-overlay)',
  },
  border: {
    primary: 'var(--color-border-primary)',
    secondary: 'var(--color-border-secondary)',
    focus: 'var(--color-border-focus)',
  },
} as const;

export type SemanticColors = typeof semanticColors;
