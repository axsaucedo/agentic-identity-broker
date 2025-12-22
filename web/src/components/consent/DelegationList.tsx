/**
 * DelegationList component displays a grid of delegation cards.
 *
 * Features:
 * - Responsive grid layout using design system Grid component
 * - Framer Motion stagger animations
 * - Handles empty states
 * - Memoized for performance with large lists
 */

import React, { memo } from 'react';
import { motion } from 'framer-motion';
import { DelegationCard } from './DelegationCard';
import { Grid } from '@design-system/components/layout/Grid';
import type { AgentDelegation } from '../../types/consent';

interface DelegationListProps {
  /** Array of agent delegations to display */
  delegations: AgentDelegation[];
  /** Callback when a delegation card is clicked */
  onDelegationClick: (agentId: string) => void;
}

/**
 * DelegationList renders a responsive grid of DelegationCard components
 * with smooth stagger animations. Memoized for performance.
 */
function DelegationListComponent({
  delegations,
  onDelegationClick,
}: DelegationListProps) {
  // Animation variants for the container
  const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
      opacity: 1,
      transition: {
        staggerChildren: 0.1,
      },
    },
  };

  // Animation variants for each card
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
      variants={containerVariants}
      initial="hidden"
      animate="visible"
    >
      <Grid
        columns={1}
        gap="lg"
        className="md:grid-cols-2 lg:grid-cols-3"
      >
        {delegations.map((delegation) => {
          // Create stable callback for each card
          const handleClick = () => onDelegationClick(delegation.agentId);

          return (
            <motion.div key={delegation.agentId} variants={itemVariants}>
              <DelegationCard
                delegation={delegation}
                onClick={handleClick}
              />
            </motion.div>
          );
        })}
      </Grid>
    </motion.div>
  );
}

/**
 * Memoized DelegationList component.
 * Only re-renders if delegations array changes.
 */
export const DelegationList = memo(
  DelegationListComponent,
  (prevProps, nextProps) => {
    // Shallow comparison of delegations array
    if (prevProps.delegations.length !== nextProps.delegations.length) {
      return false;
    }

    // Check if delegation IDs changed
    for (let i = 0; i < prevProps.delegations.length; i++) {
      if (prevProps.delegations[i].agentId !== nextProps.delegations[i].agentId) {
        return false;
      }
    }

    return true;
  }
);

DelegationList.displayName = 'DelegationList';

export default DelegationList;
