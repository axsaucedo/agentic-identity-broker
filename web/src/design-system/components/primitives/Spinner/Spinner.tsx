/**
 * Spinner Component
 *
 * Loading indicator with multiple sizes and semantic variants.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 4 sizes: xs, sm, md, lg
 * - 6 semantic color variants
 * - Optional label for accessibility
 * - Smooth CSS animation
 * - WCAG 2.1 AA compliant with proper ARIA attributes
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const spinnerVariants = cva('animate-spin', {
  variants: {
    size: {
      xs: 'w-3 h-3',
      sm: 'w-4 h-4',
      md: 'w-6 h-6',
      lg: 'w-8 h-8',
    },
    variant: {
      // Primary: Trust colors
      primary: 'text-trust-deep',

      // Success: Green
      success: 'text-success-primary',

      // Error: Red
      error: 'text-error-primary',

      // Warning: Amber
      warning: 'text-warning-primary',

      // Info: Blue
      info: 'text-info-primary',

      // Neutral: Gray
      neutral: 'text-neutral-500',

      // White: For dark backgrounds
      white: 'text-white',
    },
  },
  defaultVariants: {
    size: 'md',
    variant: 'primary',
  },
});

export interface SpinnerProps
  extends
    Omit<React.SVGAttributes<SVGSVGElement>, 'children'>,
    VariantProps<typeof spinnerVariants> {
  /** Accessible label for screen readers */
  label?: string;
}

/**
 * Spinner component for loading states.
 * Uses SVG animation for smooth rendering across all devices.
 *
 * @example
 * ```tsx
 * <Spinner />
 *
 * <Spinner size="sm" variant="success" label="Loading data..." />
 *
 * <Spinner size="lg" variant="primary" />
 * ```
 */
export const Spinner = React.forwardRef<SVGSVGElement, SpinnerProps>(
  ({ size, variant, className, label = 'Loading...', ...props }, ref) => {
    return (
      <svg
        ref={ref}
        className={cn(spinnerVariants({ size, variant }), className)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        role="status"
        aria-label={label}
        {...props}
      >
        {/* Background circle - lower opacity */}
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />

        {/* Animated arc - higher opacity */}
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />

        {/* Hidden text for screen readers */}
        <title>{label}</title>
      </svg>
    );
  },
);

Spinner.displayName = 'Spinner';

export default Spinner;
