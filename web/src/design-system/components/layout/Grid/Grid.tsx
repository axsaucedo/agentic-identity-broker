/**
 * Grid Component
 *
 * Layout component for CSS Grid layouts with responsive columns.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 1-6 column configurations
 * - 3 gap sizes: sm, md, lg
 * - Responsive column adaptation
 * - Auto-fit and auto-fill support
 * - Flexible child element sizing
 * - Fully accessible with semantic HTML
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const gridVariants = cva(
  // Base styles - applied to all variants
  'grid',
  {
    variants: {
      columns: {
        // 1 Column: Full width single column
        1: 'grid-cols-1',

        // 2 Columns: Split layout
        2: 'grid-cols-2',

        // 3 Columns: Default balanced grid
        3: 'grid-cols-3',

        // 4 Columns: Dense grid
        4: 'grid-cols-4',

        // 5 Columns: Specialized layout
        5: 'grid-cols-5',

        // 6 Columns: Maximum density
        6: 'grid-cols-6',
      },
      gap: {
        // Small: 12px - Tight grouping
        sm: 'gap-3',

        // Medium: 16px - Standard spacing (default)
        md: 'gap-4',

        // Large: 24px - Generous spacing
        lg: 'gap-6',
      },
    },
    defaultVariants: {
      columns: 3,
      gap: 'md',
    },
  },
);

export interface GridProps
  extends
    React.ComponentPropsWithoutRef<'div'>,
    VariantProps<typeof gridVariants> {
  /** Number of columns (1-6) */
  columns?: 1 | 2 | 3 | 4 | 5 | 6;
  /** Gap size between items */
  gap?: 'sm' | 'md' | 'lg';
  /** Grid content */
  children: React.ReactNode;
}

/**
 * Grid component for CSS Grid layouts.
 * Provides consistent column structure and spacing for child elements.
 *
 * @example
 * ```tsx
 * // Default 3-column grid with medium gap
 * <Grid>
 *   <div>Item 1</div>
 *   <div>Item 2</div>
 *   <div>Item 3</div>
 * </Grid>
 *
 * // 4-column grid with large gap
 * <Grid columns={4} gap="lg">
 *   <Card>Card 1</Card>
 *   <Card>Card 2</Card>
 *   <Card>Card 3</Card>
 *   <Card>Card 4</Card>
 * </Grid>
 *
 * // Responsive grid with breakpoint classes
 * <Grid columns={1} className="sm:grid-cols-2 lg:grid-cols-4">
 *   <div>Responsive Item 1</div>
 *   <div>Responsive Item 2</div>
 *   <div>Responsive Item 3</div>
 *   <div>Responsive Item 4</div>
 * </Grid>
 *
 * // Auto-fit columns with minimum width
 * <Grid className="grid-cols-[repeat(auto-fit,minmax(200px,1fr))]" gap="sm">
 *   <div>Auto Item 1</div>
 *   <div>Auto Item 2</div>
 *   <div>Auto Item 3</div>
 * </Grid>
 *
 * // Card grid example
 * <Grid columns={3} gap="md">
 *   <Card title="Feature 1" />
 *   <Card title="Feature 2" />
 *   <Card title="Feature 3" />
 * </Grid>
 * ```
 */
export const Grid = React.forwardRef<HTMLDivElement, GridProps>(
  ({ columns, gap, className, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(gridVariants({ columns, gap }), className)}
        {...props}
      >
        {children}
      </div>
    );
  },
);

Grid.displayName = 'Grid';

export default Grid;
