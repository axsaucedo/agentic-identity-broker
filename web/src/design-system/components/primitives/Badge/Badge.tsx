/**
 * Badge Component
 *
 * Versatile badge component for status indicators, labels, and counts.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 5 semantic variants: success, error, warning, info, neutral
 * - 3 sizes: sm, md, lg
 * - Optional icon support (before/after)
 * - Optional dot indicator
 * - Pill or rounded shape
 * - WCAG 2.1 AA compliant
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const badgeVariants = cva(
  // Base styles - applied to all variants
  'inline-flex items-center justify-center font-medium border transition-colors',
  {
    variants: {
      variant: {
        // Success: Green - positive status (solid background per COLOR_GUIDE.md)
        success: 'bg-success-primary text-white border-success-primary',

        // Error: Red - negative/destructive status (solid background per COLOR_GUIDE.md)
        error: 'bg-error-primary text-white border-error-primary',

        // Warning: Amber - caution status (solid background per COLOR_GUIDE.md)
        warning: 'bg-warning-primary text-trust-deep border-warning-primary',

        // Info: Blue - informational status (solid background per COLOR_GUIDE.md)
        info: 'bg-info-primary text-white border-info-primary',

        // Neutral: Warm neutral - default/neutral status
        neutral: 'bg-neutral-300 text-neutral-600 border-neutral-300',

        // Primary: Trust colors - brand-related badges (solid background per COLOR_GUIDE.md)
        primary: 'bg-trust text-white border-trust',
      },
      size: {
        sm: 'px-2 py-0.5 text-xs gap-1',
        md: 'px-2.5 py-1 text-sm gap-1.5',
        lg: 'px-3 py-1.5 text-base gap-2',
      },
      shape: {
        rounded: 'rounded-md',
        pill: 'rounded-full',
      },
    },
    defaultVariants: {
      variant: 'neutral',
      size: 'md',
      shape: 'pill',
    },
  },
);

export interface BadgeProps
  extends
    React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {
  /** Icon to display before children */
  iconBefore?: React.ReactNode;
  /** Icon to display after children */
  iconAfter?: React.ReactNode;
  /** Show a dot indicator before the content */
  showDot?: boolean;
}

/**
 * Badge component for displaying status, labels, or counts.
 * Supports icons, dots, and multiple semantic variants.
 *
 * @example
 * ```tsx
 * <Badge variant="success">Active</Badge>
 *
 * <Badge variant="warning" showDot>
 *   Pending
 * </Badge>
 *
 * <Badge variant="info" iconBefore={<InfoIcon />}>
 *   New
 * </Badge>
 *
 * <Badge variant="neutral" size="sm" shape="rounded">
 *   Beta
 * </Badge>
 * ```
 */
export const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  (
    {
      variant,
      size,
      shape,
      className,
      children,
      iconBefore,
      iconAfter,
      showDot = false,
      ...props
    },
    ref,
  ) => {
    // Icon sizing based on badge size
    const iconSize =
      size === 'sm' ? 'w-3 h-3' : size === 'lg' ? 'w-4 h-4' : 'w-3.5 h-3.5';
    const dotSize =
      size === 'sm' ? 'w-1.5 h-1.5' : size === 'lg' ? 'w-2.5 h-2.5' : 'w-2 h-2';

    return (
      <span
        ref={ref}
        className={cn(badgeVariants({ variant, size, shape }), className)}
        {...props}
      >
        {/* Dot indicator */}
        {showDot && (
          <span
            className={cn('rounded-full bg-current', dotSize)}
            aria-hidden="true"
          />
        )}

        {/* Icon before */}
        {iconBefore && (
          <span
            className={cn('flex items-center', iconSize)}
            aria-hidden="true"
          >
            {iconBefore}
          </span>
        )}

        {/* Badge content */}
        {children}

        {/* Icon after */}
        {iconAfter && (
          <span
            className={cn('flex items-center', iconSize)}
            aria-hidden="true"
          >
            {iconAfter}
          </span>
        )}
      </span>
    );
  },
);

Badge.displayName = 'Badge';

export default Badge;
