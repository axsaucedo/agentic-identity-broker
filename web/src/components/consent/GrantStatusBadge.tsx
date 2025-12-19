/**
 * GrantStatusBadge component displays the status of a grant.
 *
 * Features:
 * - Color-coded badge (active: green, expired: gray, pending: yellow)
 * - Expiration date display
 * - Icon support
 * - Accessible with proper ARIA labels
 */

import React from 'react';
import { format, isPast } from 'date-fns';

type GrantStatus = 'active' | 'expired' | 'pending';

interface GrantStatusBadgeProps {
  /** Grant status */
  status: GrantStatus;
  /** Optional expiration date */
  expiresAt?: Date | string | null;
}

/**
 * GrantStatusBadge displays a colored badge indicating grant status.
 * Shows expiration date if available.
 */
export function GrantStatusBadge({ status, expiresAt }: GrantStatusBadgeProps) {
  // Parse expiration date if provided
  const expirationDate = expiresAt ? new Date(expiresAt) : null;
  const isExpired = expirationDate ? isPast(expirationDate) : false;

  // Override status if expiration date is in the past
  const effectiveStatus: GrantStatus = isExpired ? 'expired' : status;

  // Badge styling based on status
  const getBadgeStyles = () => {
    switch (effectiveStatus) {
      case 'active':
        return {
          container: 'bg-green-100 text-green-800 border-green-200',
          icon: 'text-green-600',
          label: 'Active',
        };
      case 'expired':
        return {
          container: 'bg-gray-100 text-gray-600 border-gray-200',
          icon: 'text-gray-500',
          label: 'Expired',
        };
      case 'pending':
        return {
          container: 'bg-yellow-100 text-yellow-800 border-yellow-200',
          icon: 'text-yellow-600',
          label: 'Pending',
        };
      default:
        return {
          container: 'bg-gray-100 text-gray-600 border-gray-200',
          icon: 'text-gray-500',
          label: 'Unknown',
        };
    }
  };

  const styles = getBadgeStyles();

  // Format expiration text
  const getExpirationText = () => {
    if (!expirationDate) return null;

    if (isExpired) {
      return `Expired ${format(expirationDate, 'MMM d, yyyy')}`;
    }

    return `Expires ${format(expirationDate, 'MMM d, yyyy')}`;
  };

  const expirationText = getExpirationText();

  return (
    <div className="inline-flex items-center gap-2">
      {/* Status badge */}
      <span
        className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium border ${styles.container}`}
        role="status"
        aria-label={`Grant status: ${styles.label}`}
      >
        {/* Status icon */}
        {effectiveStatus === 'active' && (
          <svg
            className={`w-3.5 h-3.5 ${styles.icon}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )}
        {effectiveStatus === 'expired' && (
          <svg
            className={`w-3.5 h-3.5 ${styles.icon}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )}
        {effectiveStatus === 'pending' && (
          <svg
            className={`w-3.5 h-3.5 ${styles.icon}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
            />
          </svg>
        )}
        <span>{styles.label}</span>
      </span>

      {/* Expiration date */}
      {expirationText && (
        <span
          className={`text-xs ${isExpired ? 'text-gray-500 line-through' : 'text-gray-600'}`}
          aria-label={expirationText}
        >
          {expirationText}
        </span>
      )}
    </div>
  );
}

export default GrantStatusBadge;
