/**
 * Container Component
 *
 * Layout component for constraining content width and centering.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 5 size variants: sm, md, lg, xl, full
 * - Centered or left-aligned content
 * - Optional internal padding (sm, md, lg, or boolean)
 * - Responsive design with mobile-first approach
 * - Fully accessible with semantic HTML
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const containerVariants = cva(
  // Base styles - applied to all variants
  'w-full',
  {
    variants: {
      size: {
        // Small: 640px - Compact content, sidebars
        sm: 'max-w-screen-sm',

        // Medium: 768px - Default for most content
        md: 'max-w-screen-md',

        // Large: 1024px - Wide content sections
        lg: 'max-w-screen-lg',

        // Extra Large: 1280px - Marketing pages, dashboards
        xl: 'max-w-screen-xl',

        // Full: 100% - No constraint, full viewport width
        full: 'max-w-full',
      },
      centered: {
        true: 'mx-auto',
        false: '',
      },
      padding: {
        false: '',
        true: 'px-4 py-4',
        sm: 'px-3 py-2',
        md: 'px-4 py-4',
        lg: 'px-6 py-6',
      },
    },
    defaultVariants: {
      size: 'lg',
      centered: true,
      padding: false,
    },
  }
);

export interface ContainerProps
  extends React.ComponentPropsWithoutRef<'div'>,
    VariantProps<typeof containerVariants> {
  /** Maximum width constraint */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  /** Center content horizontally */
  centered?: boolean;
  /** Internal padding - boolean or size variant */
  padding?: boolean | 'sm' | 'md' | 'lg';
  /** Content to render inside container */
  children: React.ReactNode;
}

/**
 * Container component for constraining content width.
 * Provides consistent max-width constraints and centering.
 *
 * @example
 * ```tsx
 * // Default centered container with large max-width
 * <Container>
 *   <h1>Page Title</h1>
 *   <p>Content goes here...</p>
 * </Container>
 *
 * // Small container with padding
 * <Container size="sm" padding="md">
 *   <p>Compact content</p>
 * </Container>
 *
 * // Full width, left-aligned
 * <Container size="full" centered={false}>
 *   <nav>Navigation items</nav>
 * </Container>
 *
 * // Nested containers for complex layouts
 * <Container size="xl">
 *   <Container size="md" padding="lg">
 *     <article>Nested content</article>
 *   </Container>
 * </Container>
 * ```
 */
export const Container = React.forwardRef<HTMLDivElement, ContainerProps>(
  (
    {
      size,
      centered,
      padding,
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(containerVariants({ size, centered, padding }), className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Container.displayName = 'Container';

export default Container;
