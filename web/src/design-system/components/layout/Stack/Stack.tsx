/**
 * Stack Component
 *
 * Layout component for flexible stacking (rows or columns with gaps).
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Row and column directions
 * - 5 gap sizes: xs, sm, md, lg, xl
 * - Alignment control (start, center, end, stretch)
 * - Justify content options (start, center, end, space-between, space-around)
 * - Optional wrapping for multi-line layouts
 * - Responsive direction changes
 * - Fully accessible with semantic HTML
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const stackVariants = cva(
  // Base styles - applied to all variants
  'flex',
  {
    variants: {
      direction: {
        // Row: Horizontal stacking (default)
        row: 'flex-row',

        // Column: Vertical stacking
        column: 'flex-col',
      },
      gap: {
        // Extra Small: 4px - Minimal spacing
        xs: 'gap-1',

        // Small: 8px - Tight grouping
        sm: 'gap-2',

        // Medium: 16px - Standard spacing (default)
        md: 'gap-4',

        // Large: 24px - Generous spacing
        lg: 'gap-6',

        // Extra Large: 32px - Maximum spacing
        xl: 'gap-8',
      },
      align: {
        // Start: Align items to start of cross axis
        start: 'items-start',

        // Center: Align items to center of cross axis
        center: 'items-center',

        // End: Align items to end of cross axis
        end: 'items-end',

        // Stretch: Stretch items to fill cross axis (default)
        stretch: 'items-stretch',
      },
      justify: {
        // Start: Pack items to start of main axis (default)
        start: 'justify-start',

        // Center: Pack items to center of main axis
        center: 'justify-center',

        // End: Pack items to end of main axis
        end: 'justify-end',

        // Space Between: Distribute items evenly with first item at start, last at end
        'space-between': 'justify-between',

        // Space Around: Distribute items evenly with equal space around each item
        'space-around': 'justify-around',
      },
      wrap: {
        true: 'flex-wrap',
        false: 'flex-nowrap',
      },
    },
    defaultVariants: {
      direction: 'column',
      gap: 'md',
      align: 'stretch',
      justify: 'start',
      wrap: false,
    },
  }
);

export interface StackProps
  extends React.ComponentPropsWithoutRef<'div'>,
    VariantProps<typeof stackVariants> {
  /** Direction of stack - row (horizontal) or column (vertical) */
  direction?: 'row' | 'column';
  /** Gap size between items */
  gap?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  /** Alignment on cross axis */
  align?: 'start' | 'center' | 'end' | 'stretch';
  /** Justification on main axis */
  justify?: 'start' | 'center' | 'end' | 'space-between' | 'space-around';
  /** Enable wrapping for multi-line layouts */
  wrap?: boolean;
  /** Content to stack */
  children: React.ReactNode;
}

/**
 * Stack component for flexible row and column layouts.
 * Provides consistent spacing and alignment for child elements.
 *
 * @example
 * ```tsx
 * // Default vertical stack with medium gap
 * <Stack>
 *   <div>Item 1</div>
 *   <div>Item 2</div>
 *   <div>Item 3</div>
 * </Stack>
 *
 * // Horizontal stack with large gap and centered items
 * <Stack direction="row" gap="lg" align="center">
 *   <button>Action 1</button>
 *   <button>Action 2</button>
 *   <button>Action 3</button>
 * </Stack>
 *
 * // Responsive stack with wrapping
 * <Stack direction="row" wrap={true} gap="sm">
 *   <div>Tag 1</div>
 *   <div>Tag 2</div>
 *   <div>Tag 3</div>
 * </Stack>
 *
 * // Space-between layout for header/footer patterns
 * <Stack direction="row" justify="space-between" align="center">
 *   <h1>Logo</h1>
 *   <nav>Navigation</nav>
 * </Stack>
 *
 * // Nested stacks for complex layouts
 * <Stack gap="xl">
 *   <Stack direction="row" justify="space-between">
 *     <h1>Title</h1>
 *     <button>Action</button>
 *   </Stack>
 *   <Stack gap="sm">
 *     <p>Content paragraph 1</p>
 *     <p>Content paragraph 2</p>
 *   </Stack>
 * </Stack>
 * ```
 */
export const Stack = React.forwardRef<HTMLDivElement, StackProps>(
  (
    {
      direction,
      gap,
      align,
      justify,
      wrap,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          stackVariants({ direction, gap, align, justify, wrap }),
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Stack.displayName = 'Stack';

export default Stack;
