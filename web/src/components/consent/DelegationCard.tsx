/**
 * DelegationCard component displays a summary of an agent delegation.
 *
 * Shows:
 * - Agent logo and name
 * - Number of active service grants
 * - Last modified date
 * - Optional expiration date
 *
 * Performance: Memoized to prevent unnecessary re-renders in lists.
 */

import React, { memo } from 'react';
import { formatDistanceToNow, format } from 'date-fns';
import type { AgentDelegation } from '../../types/consent';

interface DelegationCardProps {
  /** Agent delegation data */
  delegation: AgentDelegation;
  /** Callback when card is clicked */
  onClick: () => void;
}

/**
 * DelegationCard displays a clickable card with agent delegation summary.
 * Clicking the card navigates to the detailed grant management page.
 * Memoized for performance in large lists.
 */
function DelegationCardComponent({ delegation, onClick }: DelegationCardProps) {
  const {
    displayName,
    logoUrl,
    activeGrantCount,
    lastModifiedAt,
    expiresAt,
  } = delegation;

  // Format last modified time as relative (e.g., "2 days ago")
  const lastModifiedText = formatDistanceToNow(new Date(lastModifiedAt), {
    addSuffix: true,
  });

  // Format expiration date if present
  const expirationText = expiresAt
    ? `Expires ${format(new Date(expiresAt), 'MMM d, yyyy')}`
    : null;

  return (
    <button
      type="button"
      onClick={onClick}
      className="card p-6 w-full text-left transition-all duration-200 hover:shadow-lg hover:border-blue-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
    >
      {/* Agent logo and name */}
      <div className="flex items-start gap-4">
        {/* Logo */}
        <div className="flex-shrink-0">
          {logoUrl ? (
            <img
              src={logoUrl}
              alt={`${displayName} logo`}
              className="w-12 h-12 rounded-lg object-cover"
            />
          ) : (
            <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-white text-lg font-semibold">
                {displayName.charAt(0).toUpperCase()}
              </span>
            </div>
          )}
        </div>

        {/* Agent info */}
        <div className="flex-1 min-w-0">
          <h3 className="text-lg font-semibold text-gray-900 truncate">
            {displayName}
          </h3>
          <p className="mt-1 text-sm text-gray-600">
            {activeGrantCount === 1
              ? '1 service'
              : `${activeGrantCount} services`}
          </p>
        </div>

        {/* Arrow icon */}
        <div className="flex-shrink-0 text-gray-400">
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 5l7 7-7 7"
            />
          </svg>
        </div>
      </div>

      {/* Metadata row */}
      <div className="mt-4 flex items-center justify-between text-xs text-gray-500">
        <span>Updated {lastModifiedText}</span>
        {expirationText && (
          <span className="flex items-center gap-1">
            <svg
              className="w-4 h-4"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            {expirationText}
          </span>
        )}
      </div>
    </button>
  );
}

/**
 * Memoized DelegationCard component.
 * Only re-renders if delegation data or onClick changes.
 */
export const DelegationCard = memo(
  DelegationCardComponent,
  (prevProps, nextProps) => {
    // Custom comparison: only re-render if delegation or onClick changed
    return (
      prevProps.delegation.agentId === nextProps.delegation.agentId &&
      prevProps.delegation.lastModifiedAt === nextProps.delegation.lastModifiedAt &&
      prevProps.delegation.activeGrantCount === nextProps.delegation.activeGrantCount &&
      prevProps.delegation.expiresAt === nextProps.delegation.expiresAt &&
      prevProps.onClick === nextProps.onClick
    );
  }
);

DelegationCard.displayName = 'DelegationCard';

export default DelegationCard;
