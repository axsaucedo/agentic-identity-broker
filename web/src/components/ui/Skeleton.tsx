/**
 * Skeleton loading components for displaying placeholder UI while data loads.
 *
 * Components:
 * - Skeleton: Generic shimmer skeleton box
 * - DelegationListSkeleton: Specific skeleton for delegation card list
 */

import React from 'react';
import { motion } from 'framer-motion';

interface SkeletonProps {
  /** Width of the skeleton (CSS value) */
  width?: string;
  /** Height of the skeleton (CSS value) */
  height?: string;
  /** Additional Tailwind classes */
  className?: string;
}

/**
 * Generic skeleton component with shimmer animation.
 * Used as a placeholder while content is loading.
 */
export function Skeleton({ width, height, className = '' }: SkeletonProps) {
  const style = {
    width: width || '100%',
    height: height || '1rem',
  };

  return (
    <div
      className={`bg-gray-200 rounded animate-pulse ${className}`}
      style={style}
      aria-hidden="true"
    />
  );
}

/**
 * Skeleton card component matching the DelegationCard layout.
 * Shows placeholder content while delegation data loads.
 */
function DelegationCardSkeleton() {
  return (
    <div className="card p-6 space-y-4">
      {/* Agent logo and name row */}
      <div className="flex items-center gap-4">
        <Skeleton width="48px" height="48px" className="rounded-lg" />
        <div className="flex-1 space-y-2">
          <Skeleton width="60%" height="1.5rem" />
          <Skeleton width="40%" height="1rem" />
        </div>
      </div>

      {/* Service count and date row */}
      <div className="flex items-center justify-between">
        <Skeleton width="30%" height="1rem" />
        <Skeleton width="35%" height="1rem" />
      </div>
    </div>
  );
}

/**
 * DelegationListSkeleton displays 3-4 skeleton cards with stagger animation.
 * Used while fetching the agent delegation list.
 */
export function DelegationListSkeleton() {
  const cardCount = 4;

  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
      },
    },
  };

  const itemVariants = {
    hidden: { opacity: 0, y: 20 },
    visible: {
      opacity: 1,
      y: 0,
      transition: {
        duration: 0.3,
      },
    },
  };

  return (
    <motion.div
      className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6"
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      {Array.from({ length: cardCount }).map((_, index) => (
        <motion.div key={index} variants={itemVariants}>
          <DelegationCardSkeleton />
        </motion.div>
      ))}
    </motion.div>
  );
}

export default Skeleton;
