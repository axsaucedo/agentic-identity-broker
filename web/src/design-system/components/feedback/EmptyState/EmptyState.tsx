/**
 * EmptyState Component
 *
 * Feedback component for displaying empty state UIs when there's no data,
 * no results, or no permissions. Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Centered layout with icon, title, and description
 * - Optional primary and secondary action buttons
 * - 3 size variants: compact, default, expanded
 * - Flexible icon slot for custom illustrations
 * - Semantic HTML with proper accessibility
 * - WCAG 2.1 AA compliant
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';
import { Button } from '@design-system/components/primitives/Button';

const emptyStateVariants = cva(
  // Base styles - centered layout with consistent structure
  'flex flex-col items-center justify-center text-center max-w-md mx-auto',
  {
    variants: {
      size: {
        // Compact: Minimal spacing for inline or sidebar use
        compact: 'py-6 px-4 space-y-2',

        // Default: Standard spacing for most use cases
        default: 'py-10 px-6 space-y-4',

        // Expanded: Generous spacing for prominent display
        expanded: 'py-16 px-8 space-y-6',
      },
    },
    defaultVariants: {
      size: 'default',
    },
  }
);

const iconContainerVariants = cva('flex items-center justify-center rounded-full', {
  variants: {
    size: {
      compact: 'w-12 h-12 mb-2',
      default: 'w-16 h-16 mb-3',
      expanded: 'w-20 h-20 mb-4',
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

const titleVariants = cva('font-semibold text-gray-900', {
  variants: {
    size: {
      compact: 'text-base',
      default: 'text-lg',
      expanded: 'text-xl',
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

const descriptionVariants = cva('text-gray-600 leading-relaxed', {
  variants: {
    size: {
      compact: 'text-sm',
      default: 'text-base',
      expanded: 'text-base',
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

const actionsContainerVariants = cva('flex gap-3', {
  variants: {
    size: {
      compact: 'flex-col w-full mt-3',
      default: 'flex-row items-center justify-center mt-4',
      expanded: 'flex-row items-center justify-center mt-6',
    },
  },
  defaultVariants: {
    size: 'default',
  },
});

export interface EmptyStateProps
  extends Omit<React.ComponentPropsWithoutRef<'div'>, 'title'> {
  /** Icon or illustration to display (decorative) */
  icon?: React.ReactNode;
  /** Main heading text (required) */
  title: string;
  /** Supporting description text (optional) */
  description?: string;
  /** Primary call-to-action button */
  primaryAction?: {
    label: string;
    onClick: () => void;
  };
  /** Secondary action button */
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  /** Size variant for spacing and typography */
  size?: 'compact' | 'default' | 'expanded';
}

// Default icon for empty states (folder icon)
const DefaultIcon = () => (
  <svg
    className="w-full h-full text-gray-400"
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
    aria-hidden="true"
  >
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={1.5}
      d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
    />
  </svg>
);

/**
 * EmptyState component for displaying friendly empty state UIs.
 * Shows when there's no data, no search results, or user lacks permissions.
 *
 * @example
 * ```tsx
 * <EmptyState
 *   title="No consents yet"
 *   description="You haven't granted any permissions to applications."
 * />
 *
 * <EmptyState
 *   icon={<SearchIcon />}
 *   title="No results found"
 *   description="Try adjusting your search or filters."
 *   primaryAction={{
 *     label: "Clear filters",
 *     onClick: handleClear
 *   }}
 * />
 *
 * <EmptyState
 *   size="expanded"
 *   icon={<IllustrationComponent />}
 *   title="Get started"
 *   description="Connect your first application to begin."
 *   primaryAction={{
 *     label: "Add application",
 *     onClick: handleAdd
 *   }}
 *   secondaryAction={{
 *     label: "Learn more",
 *     onClick: handleLearnMore
 *   }}
 * />
 * ```
 */
export const EmptyState = React.forwardRef<HTMLDivElement, EmptyStateProps>(
  (
    {
      icon,
      title,
      description,
      primaryAction,
      secondaryAction,
      size = 'default',
      className,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        role="status"
        aria-live="polite"
        className={cn(emptyStateVariants({ size }), className)}
        {...props}
      >
        {/* Icon container */}
        {(icon || !icon) && (
          <div className={iconContainerVariants({ size })}>
            {icon || <DefaultIcon />}
          </div>
        )}

        {/* Title */}
        <h3 className={titleVariants({ size })}>{title}</h3>

        {/* Description */}
        {description && (
          <p className={descriptionVariants({ size })}>{description}</p>
        )}

        {/* Actions */}
        {(primaryAction || secondaryAction) && (
          <div className={actionsContainerVariants({ size })}>
            {primaryAction && (
              <Button
                variant="primary"
                size={size === 'compact' ? 'sm' : 'md'}
                onClick={primaryAction.onClick}
                fullWidth={size === 'compact'}
              >
                {primaryAction.label}
              </Button>
            )}

            {secondaryAction && (
              <Button
                variant="outline"
                size={size === 'compact' ? 'sm' : 'md'}
                onClick={secondaryAction.onClick}
                fullWidth={size === 'compact'}
              >
                {secondaryAction.label}
              </Button>
            )}
          </div>
        )}
      </div>
    );
  }
);

EmptyState.displayName = 'EmptyState';

export default EmptyState;
