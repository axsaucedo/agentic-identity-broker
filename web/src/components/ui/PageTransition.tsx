/**
 * PageTransition component provides animated transitions between pages.
 *
 * Features:
 * - Smooth fade + slide animations
 * - Framer Motion powered
 * - Consistent animation timing across pages
 * - Supports custom animation variants
 *
 * Usage:
 * ```tsx
 * import { PageTransition } from '@components/ui/PageTransition';
 *
 * function MyPage() {
 *   return (
 *     <PageTransition>
 *       <div>Page content here</div>
 *     </PageTransition>
 *   );
 * }
 * ```
 */

import React from 'react';
import { motion, Variants } from 'framer-motion';

interface PageTransitionProps {
  /** Page content to animate */
  children: React.ReactNode;
  /** Custom animation variants (optional) */
  variants?: Variants;
  /** Animation duration in seconds */
  duration?: number;
}

/**
 * Default animation variants for page transitions.
 * - initial: Page starts slightly below and invisible
 * - animate: Page slides up and fades in
 * - exit: Page fades out and slides up slightly
 */
const defaultVariants: Variants = {
  initial: {
    opacity: 0,
    y: 20,
  },
  animate: {
    opacity: 1,
    y: 0,
  },
  exit: {
    opacity: 0,
    y: -20,
  },
};

/**
 * PageTransition wraps page content with smooth entry/exit animations.
 * Use this component to wrap each page for consistent transitions.
 */
export function PageTransition({
  children,
  variants = defaultVariants,
  duration = 0.3,
}: PageTransitionProps) {
  return (
    <motion.div
      initial="initial"
      animate="animate"
      exit="exit"
      variants={variants}
      transition={{
        duration,
        ease: 'easeInOut',
      }}
    >
      {children}
    </motion.div>
  );
}

/**
 * Fade-only page transition (no slide).
 */
export const fadeVariants: Variants = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
};

/**
 * Slide from right page transition.
 */
export const slideFromRightVariants: Variants = {
  initial: { opacity: 0, x: 100 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: -100 },
};

/**
 * Slide from left page transition.
 */
export const slideFromLeftVariants: Variants = {
  initial: { opacity: 0, x: -100 },
  animate: { opacity: 1, x: 0 },
  exit: { opacity: 0, x: 100 },
};

/**
 * Scale + fade page transition.
 */
export const scaleVariants: Variants = {
  initial: { opacity: 0, scale: 0.95 },
  animate: { opacity: 1, scale: 1 },
  exit: { opacity: 0, scale: 1.05 },
};

export default PageTransition;
