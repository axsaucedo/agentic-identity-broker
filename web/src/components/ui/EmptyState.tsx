/**
 * EmptyState component displays a centered message when no data is available.
 *
 * Features:
 * - Optional icon/illustration
 * - Title and description text
 * - Optional call-to-action button
 * - Centered layout with proper spacing
 */

import React, { ReactNode } from 'react';

interface EmptyStateProps {
  /** Optional icon or image element to display */
  icon?: ReactNode;
  /** Main heading text */
  title: string;
  /** Descriptive text explaining the empty state */
  description: string;
  /** Optional action button configuration */
  action?: {
    label: string;
    onClick: () => void;
  };
}

/**
 * EmptyState displays a user-friendly message when there's no content to show.
 * Commonly used for empty lists, search results, etc.
 */
export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="card p-12 text-center">
      {/* Icon/Illustration */}
      {icon ? (
        <div className="flex justify-center">{icon}</div>
      ) : (
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
          />
        </svg>
      )}

      {/* Title */}
      <h3 className="mt-4 text-lg font-semibold text-gray-900">{title}</h3>

      {/* Description */}
      <p className="mt-2 text-sm text-gray-600 max-w-md mx-auto">{description}</p>

      {/* Optional action button */}
      {action && (
        <div className="mt-6">
          <button
            type="button"
            onClick={action.onClick}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors"
          >
            {action.label}
          </button>
        </div>
      )}
    </div>
  );
}

export default EmptyState;
