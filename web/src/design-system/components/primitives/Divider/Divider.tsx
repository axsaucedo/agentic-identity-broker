/**
 * Divider Component
 *
 * Horizontal or vertical separator for visual content division.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Horizontal and vertical orientations
 * - Solid or dashed styles
 * - Customizable spacing
 * - Optional label/text in the middle
 * - Semantic color variants
 * - WCAG 2.1 AA compliant
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const dividerVariants = cva('', {
  variants: {
    orientation: {
      horizontal: 'w-full',
      vertical: 'h-full',
    },
    variant: {
      default: '',
      subtle: '',
      muted: '',
    },
  },
  defaultVariants: {
    orientation: 'horizontal',
    variant: 'default',
  },
});

const dividerLineVariants = cva('', {
  variants: {
    orientation: {
      horizontal: 'h-px w-full',
      vertical: 'w-px h-full',
    },
    style: {
      solid: '',
      dashed: '',
    },
    variant: {
      default: 'bg-neutral-300',
      subtle: 'bg-neutral-200',
      muted: 'bg-neutral-100',
    },
  },
  compoundVariants: [
    {
      style: 'dashed',
      orientation: 'horizontal',
      className: 'border-t border-dashed border-neutral-300',
      // Keep bg-neutral-300 for solid fallback
    },
    {
      style: 'dashed',
      orientation: 'vertical',
      className: 'border-l border-dashed border-neutral-300',
    },
  ],
  defaultVariants: {
    orientation: 'horizontal',
    style: 'solid',
    variant: 'default',
  },
});

export interface DividerProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'style'>,
    VariantProps<typeof dividerVariants> {
  /** Style of the divider: solid or dashed */
  style?: 'solid' | 'dashed';
  /** Optional text label to display in the middle (horizontal only) */
  label?: React.ReactNode;
  /** Spacing around the label */
  labelSpacing?: 'sm' | 'md' | 'lg';
}

/**
 * Divider component for visual content separation.
 * Can display horizontally or vertically with optional centered text.
 *
 * @example
 * ```tsx
 * <Divider />
 *
 * <Divider label="Or" />
 *
 * <Divider variant="subtle" />
 *
 * <Divider orientation="vertical" />
 *
 * <Divider style="dashed" labelSpacing="lg" label="Section Break" />
 * ```
 */
export const Divider = React.forwardRef<HTMLDivElement, DividerProps>(
  (
    {
      orientation = 'horizontal',
      variant = 'default',
      className,
      style: dividerStyle = 'solid',
      label,
      labelSpacing = 'md',
      ...props
    },
    ref
  ) => {
    const spacing = {
      sm: 'px-2',
      md: 'px-3',
      lg: 'px-4',
    }[labelSpacing];

    // Vertical divider - no label support
    if (orientation === 'vertical') {
      return (
        <div
          ref={ref}
          className={cn(
            dividerVariants({ orientation, variant }),
            'flex items-center'
          )}
          role="separator"
          aria-orientation="vertical"
          {...props}
        >
          <div
            className={cn(dividerLineVariants({ orientation, style: dividerStyle, variant }))}
            style={dividerStyle === 'solid' ? {} : undefined}
          />
        </div>
      );
    }

    // Horizontal divider without label
    if (!label) {
      return (
        <div
          ref={ref}
          className={cn(
            dividerVariants({ orientation, variant }),
            'flex items-center'
          )}
          role="separator"
          aria-orientation="horizontal"
          {...props}
        >
          <div
            className={cn(dividerLineVariants({ orientation, style: dividerStyle, variant }))}
            style={dividerStyle === 'solid' ? {} : undefined}
          />
        </div>
      );
    }

    // Horizontal divider with label
    return (
      <div
        ref={ref}
        className={cn(
          dividerVariants({ orientation, variant }),
          'flex items-center gap-0',
          className
        )}
        role="separator"
        aria-label={typeof label === 'string' ? label : undefined}
        {...props}
      >
        {/* Left line */}
        <div
          className={cn(
            dividerLineVariants({ orientation, style: dividerStyle, variant }),
            'flex-1'
          )}
          style={dividerStyle === 'solid' ? {} : undefined}
        />

        {/* Center label */}
        <span
          className={cn(
            spacing,
            'text-sm text-neutral-600 font-medium whitespace-nowrap'
          )}
        >
          {label}
        </span>

        {/* Right line */}
        <div
          className={cn(
            dividerLineVariants({ orientation, style: dividerStyle, variant }),
            'flex-1'
          )}
          style={dividerStyle === 'solid' ? {} : undefined}
        />
      </div>
    );
  }
);

Divider.displayName = 'Divider';

export default Divider;
