/**
 * StatusIndicator Component
 *
 * Displays a status label with optional icon for permission metadata.
 * Often wrapped with Tooltip for additional context.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Optional icon before text
 * - Semantic color variants
 * - Small, subtle styling for metadata display
 * - Accessible markup with proper ARIA labels
 * - Works well with Tooltip for contextual help
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const statusIndicatorVariants = cva(
  'inline-flex items-center gap-1 text-xs text-neutral-500',
  {
    variants: {
      variant: {
        default: 'text-neutral-500',
        success: 'text-success-primary',
        warning: 'text-warning-primary',
        error: 'text-error-primary',
        info: 'text-info-primary',
      },
      interactive: {
        true: 'cursor-help',
        false: '',
      },
    },
    defaultVariants: {
      variant: 'default',
      interactive: false,
    },
  },
);

export interface StatusIndicatorProps
  extends
    React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof statusIndicatorVariants> {
  /** Icon to display before the text */
  icon?: React.ReactNode;
  /** Status label text */
  label: string;
}

/**
 * StatusIndicator for displaying permission metadata.
 * Compact indicator that works well within Tooltip wrappers.
 *
 * @example
 * ```tsx
 * <Tooltip content="Data is encrypted in transit and at rest">
 *   <StatusIndicator
 *     icon={<LockIcon />}
 *     label="Encrypted"
 *     interactive
 *   />
 * </Tooltip>
 *
 * <StatusIndicator label="Active" variant="success" />
 * ```
 */
export const StatusIndicator = React.forwardRef<
  HTMLDivElement,
  StatusIndicatorProps
>(
  (
    {
      icon,
      label,
      variant = 'default',
      interactive = false,
      className,
      ...props
    },
    ref,
  ) => {
    // Generate data-testid for scope indicators for E2E testing
    // Pattern: scope-indicator-{scope-value}
    const testId = `scope-indicator-${label.toLowerCase().replace(/\s+/g, '-')}`;

    return (
      <div
        ref={ref}
        role="img"
        aria-label={label}
        data-testid={testId}
        className={cn(
          statusIndicatorVariants({ variant, interactive }),
          className,
        )}
        {...props}
      >
        {/* Icon */}
        {icon && (
          <span
            className="inline-flex flex-shrink-0 w-4 h-4"
            aria-hidden="true"
          >
            {icon}
          </span>
        )}

        {/* Label */}
        <span>{label}</span>
      </div>
    );
  },
);

StatusIndicator.displayName = 'StatusIndicator';

export default StatusIndicator;
