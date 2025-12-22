/**
 * Skeleton Component
 *
 * Loading state placeholder component for creating skeleton screens.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 4 shape variants: line, circle, rectangle, rounded
 * - Optional pulse animation
 * - Flexible width and height
 * - Support for multiple skeleton instances via count prop
 * - Configurable gap between multiple skeletons
 * - WCAG 2.1 AA compliant with proper ARIA attributes
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const skeletonVariants = cva(
  // Base styles - applied to all variants
  'bg-gray-200 overflow-hidden',
  {
    variants: {
      variant: {
        // Line: Default rectangle with full width
        line: 'h-4 w-full rounded',

        // Circle: Perfect circle shape (equal width/height)
        circle: 'rounded-full',

        // Rectangle: Basic rectangle with sharp corners
        rectangle: '',

        // Rounded: Rectangle with rounded corners
        rounded: 'rounded-lg',
      },
      animate: {
        true: 'animate-pulse',
        false: '',
      },
    },
    defaultVariants: {
      variant: 'line',
      animate: true,
    },
  }
);

export interface SkeletonProps
  extends Omit<React.ComponentPropsWithoutRef<'div'>, 'children'> {
  /** Shape variant of the skeleton */
  variant?: 'line' | 'circle' | 'rectangle' | 'rounded';
  /** Whether to animate with pulse effect */
  animate?: boolean;
  /** Custom width (CSS value: px, %, rem, etc.) */
  width?: string;
  /** Custom height (CSS value: px, %, rem, etc.) */
  height?: string;
  /** Number of skeletons to render */
  count?: number;
  /** Gap between multiple skeletons (CSS value: px, rem, etc.) */
  gap?: string;
}

/**
 * Skeleton component for displaying loading state placeholders.
 * Supports multiple shapes, animations, and flexible sizing.
 *
 * @example
 * ```tsx
 * // Basic line skeleton
 * <Skeleton />
 *
 * // Avatar skeleton
 * <Skeleton variant="circle" width="48px" height="48px" />
 *
 * // Card image skeleton
 * <Skeleton variant="rounded" width="100%" height="200px" />
 *
 * // Multiple text lines
 * <Skeleton count={3} gap="0.5rem" />
 *
 * // Without animation
 * <Skeleton animate={false} />
 * ```
 */
export const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>(
  (
    {
      variant = 'line',
      animate = true,
      width,
      height,
      count = 1,
      gap = '0.5rem',
      className,
      style,
      ...props
    },
    ref
  ) => {
    // Calculate default dimensions based on variant
    const getDefaultDimensions = () => {
      const dimensions: React.CSSProperties = {};

      // Apply custom width/height if provided
      if (width) dimensions.width = width;
      if (height) dimensions.height = height;

      // Set defaults for circle variant if not specified
      if (variant === 'circle') {
        if (!width && !height) {
          dimensions.width = '40px';
          dimensions.height = '40px';
        } else if (width && !height) {
          dimensions.height = width;
        } else if (height && !width) {
          dimensions.width = height;
        }
      }

      // Set defaults for rectangle/rounded variants if not specified
      if ((variant === 'rectangle' || variant === 'rounded') && !height) {
        dimensions.height = '100px';
      }
      if ((variant === 'rectangle' || variant === 'rounded') && !width) {
        dimensions.width = '100%';
      }

      return dimensions;
    };

    const dimensions = getDefaultDimensions();
    const combinedStyle = { ...dimensions, ...style };

    // Single skeleton
    if (count === 1) {
      return (
        <div
          ref={ref}
          role="status"
          aria-label="Loading"
          aria-live="polite"
          className={cn(skeletonVariants({ variant, animate }), className)}
          style={combinedStyle}
          {...props}
        >
          <span className="sr-only">Loading...</span>
        </div>
      );
    }

    // Multiple skeletons
    return (
      <div
        ref={ref}
        role="status"
        aria-label="Loading"
        aria-live="polite"
        className={className}
        style={style}
        {...props}
      >
        <div className="flex flex-col" style={{ gap }}>
          {Array.from({ length: count }).map((_, index) => (
            <div
              key={index}
              className={cn(skeletonVariants({ variant, animate }))}
              style={dimensions}
            />
          ))}
        </div>
        <span className="sr-only">Loading...</span>
      </div>
    );
  }
);

Skeleton.displayName = 'Skeleton';

export default Skeleton;
