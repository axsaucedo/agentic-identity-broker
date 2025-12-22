/**
 * Accessibility Helpers
 * Utilities for building accessible components (WCAG 2.1 AA)
 */

/**
 * Generate ARIA props for common patterns
 */
export function getAriaProps(options: {
  label?: string;
  labelledBy?: string;
  describedBy?: string;
  role?: string;
  expanded?: boolean;
  selected?: boolean;
  disabled?: boolean;
  required?: boolean;
  invalid?: boolean;
}) {
  const props: Record<string, string | boolean | undefined> = {};

  if (options.label) props['aria-label'] = options.label;
  if (options.labelledBy) props['aria-labelledby'] = options.labelledBy;
  if (options.describedBy) props['aria-describedby'] = options.describedBy;
  if (options.role) props['role'] = options.role;
  if (options.expanded !== undefined) props['aria-expanded'] = options.expanded;
  if (options.selected !== undefined) props['aria-selected'] = options.selected;
  if (options.disabled !== undefined) props['aria-disabled'] = options.disabled;
  if (options.required !== undefined) props['aria-required'] = options.required;
  if (options.invalid !== undefined) props['aria-invalid'] = options.invalid;

  return props;
}

/**
 * Announce message to screen readers
 */
export function announceToScreenReader(message: string, priority: 'polite' | 'assertive' = 'polite') {
  const announcement = document.createElement('div');
  announcement.setAttribute('role', 'status');
  announcement.setAttribute('aria-live', priority);
  announcement.setAttribute('aria-atomic', 'true');
  announcement.className = 'sr-only';
  announcement.textContent = message;

  document.body.appendChild(announcement);

  // Remove after announcement
  setTimeout(() => {
    document.body.removeChild(announcement);
  }, 1000);
}

/**
 * Check if user prefers reduced motion
 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

/**
 * Generate unique ID for form elements (for aria-describedby, aria-labelledby)
 */
let idCounter = 0;
export function generateId(prefix = 'ds'): string {
  idCounter += 1;
  return `${prefix}-${idCounter}`;
}
