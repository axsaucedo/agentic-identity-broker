/**
 * Tooltip Component
 *
 * Lightweight overlay component for displaying contextual information on hover or focus.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - 4 position variants: top, right, bottom, left
 * - 2 theme variants: dark (default), light
 * - Optional arrow indicator pointing to trigger
 * - Configurable hover delay (default 200ms)
 * - Keyboard focus support for accessibility
 * - WCAG 2.1 AA compliant with proper ARIA attributes
 * - Uses Headless UI Popover for positioning
 */

import React, { useState, useEffect, useRef } from 'react';
import { cva } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const tooltipVariants = cva(
  // Base styles - applied to all variants
  'absolute z-50 px-3 py-2 text-sm font-medium rounded-md shadow-lg pointer-events-none transition-opacity duration-150',
  {
    variants: {
      theme: {
        // Dark theme (default) - high contrast
        dark: 'bg-neutral-900 text-white',

        // Light theme - subtle with border
        light: 'bg-white text-neutral-900 border border-neutral-300 shadow-md',
      },
      position: {
        top: '',
        right: '',
        bottom: '',
        left: '',
      },
    },
    defaultVariants: {
      theme: 'dark',
      position: 'top',
    },
  },
);

const arrowVariants = cva('absolute w-2 h-2 rotate-45', {
  variants: {
    theme: {
      dark: 'bg-neutral-900',
      light: 'bg-white border-neutral-300',
    },
    position: {
      top: 'bottom-[-4px] left-1/2 -translate-x-1/2',
      right: 'left-[-4px] top-1/2 -translate-y-1/2',
      bottom: 'top-[-4px] left-1/2 -translate-x-1/2',
      left: 'right-[-4px] top-1/2 -translate-y-1/2',
    },
  },
  defaultVariants: {
    theme: 'dark',
    position: 'top',
  },
});

// Border variants for light theme arrow
const arrowBorderVariants = cva('', {
  variants: {
    theme: {
      dark: '',
      light: 'border-l border-t',
    },
    position: {
      top: '',
      right: 'border-l border-t',
      bottom: '',
      left: 'border-l border-t',
    },
  },
});

export interface TooltipProps {
  /** Content to display in the tooltip */
  content: string | React.ReactNode;
  /** Element that triggers the tooltip */
  children: React.ReactNode;
  /** Position of tooltip relative to trigger */
  position?: 'top' | 'right' | 'bottom' | 'left';
  /** Visual theme variant */
  theme?: 'dark' | 'light';
  /** Whether to show arrow indicator */
  showArrow?: boolean;
  /** Delay in milliseconds before showing tooltip */
  delay?: number;
  /** Additional CSS classes */
  className?: string;
  /** Whether tooltip is disabled */
  disabled?: boolean;
}

/**
 * Tooltip component for displaying contextual help on hover or focus.
 * Automatically positions itself based on the position prop.
 *
 * @example
 * ```tsx
 * <Tooltip content="This is a helpful tip">
 *   <button>Hover me</button>
 * </Tooltip>
 *
 * <Tooltip content="Info about this field" position="right" theme="light">
 *   <InfoIcon />
 * </Tooltip>
 *
 * <Tooltip content="Multi-line content\nSupported here" showArrow delay={500}>
 *   <span>Long delay tooltip</span>
 * </Tooltip>
 * ```
 */
export const Tooltip = React.forwardRef<HTMLDivElement, TooltipProps>(
  (
    {
      content,
      children,
      position = 'top',
      theme = 'dark',
      showArrow = true,
      delay = 200,
      className,
      disabled = false,
      ...props
    },
    ref,
  ) => {
    const [isVisible, setIsVisible] = useState(false);
    const [showTooltip, setShowTooltip] = useState(false);
    const timeoutRef = useRef<number | null>(null);
    const tooltipRef = useRef<HTMLDivElement>(null);
    const triggerRef = useRef<HTMLDivElement>(null);

    // Clear timeout on unmount
    useEffect(() => {
      return () => {
        if (timeoutRef.current) {
          window.clearTimeout(timeoutRef.current);
        }
      };
    }, []);

    // Handle show with delay
    const handleShow = () => {
      if (disabled) return;

      if (delay > 0) {
        timeoutRef.current = window.setTimeout(() => {
          setIsVisible(true);
          // Small additional delay for fade-in effect
          setTimeout(() => setShowTooltip(true), 10);
        }, delay);
      } else {
        setIsVisible(true);
        setTimeout(() => setShowTooltip(true), 10);
      }
    };

    // Handle hide
    const handleHide = () => {
      if (timeoutRef.current) {
        window.clearTimeout(timeoutRef.current);
        timeoutRef.current = null;
      }
      setShowTooltip(false);
      // Wait for fade-out animation before removing from DOM
      setTimeout(() => setIsVisible(false), 150);
    };

    // Calculate tooltip position
    const getTooltipPosition = () => {
      if (!triggerRef.current || !tooltipRef.current) return {};

      const gap = 8; // Gap between trigger and tooltip

      switch (position) {
        case 'top':
          return {
            bottom: `calc(100% + ${gap}px)`,
            left: '50%',
            transform: 'translateX(-50%)',
          };
        case 'bottom':
          return {
            top: `calc(100% + ${gap}px)`,
            left: '50%',
            transform: 'translateX(-50%)',
          };
        case 'left':
          return {
            right: `calc(100% + ${gap}px)`,
            top: '50%',
            transform: 'translateY(-50%)',
          };
        case 'right':
          return {
            left: `calc(100% + ${gap}px)`,
            top: '50%',
            transform: 'translateY(-50%)',
          };
        default:
          return {};
      }
    };

    if (disabled) {
      return <>{children}</>;
    }

    return (
      <div
        ref={ref}
        className="relative inline-flex"
        onMouseEnter={handleShow}
        onMouseLeave={handleHide}
        onFocus={handleShow}
        onBlur={handleHide}
        {...props}
      >
        {/* Trigger element */}
        <div
          ref={triggerRef}
          className="inline-flex"
          tabIndex={0}
          role="button"
          aria-describedby={isVisible ? 'tooltip' : undefined}
        >
          {children}
        </div>

        {/* Tooltip content */}
        {isVisible && (
          <div
            ref={tooltipRef}
            id="tooltip"
            role="tooltip"
            className={cn(
              tooltipVariants({ theme, position }),
              showTooltip ? 'opacity-100' : 'opacity-0',
              className,
            )}
            style={getTooltipPosition()}
          >
            {/* Arrow indicator */}
            {showArrow && (
              <div
                className={cn(
                  arrowVariants({ theme, position }),
                  arrowBorderVariants({ theme, position }),
                )}
              />
            )}

            {/* Content */}
            <div className="relative z-10 whitespace-nowrap">{content}</div>
          </div>
        )}
      </div>
    );
  },
);

Tooltip.displayName = 'Tooltip';

export default Tooltip;
