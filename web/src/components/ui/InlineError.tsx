/**
 * InlineError component displays error messages with retry functionality.
 *
 * Features:
 * - Red alert styling
 * - Error icon
 * - Retry button
 * - Dismissible with close button
 */

import React, { useState } from 'react';

interface InlineErrorProps {
  /** Error message to display */
  error: string;
  /** Callback function for retry action */
  onRetry: () => void;
  /** Label for the retry button. Defaults to "Try again". */
  retryLabel?: string;
}

/**
 * InlineError displays an error message with an option to retry the failed action.
 * Can be dismissed by the user.
 */
export function InlineError({ error, onRetry, retryLabel = 'Try again' }: InlineErrorProps) {
  const [isDismissed, setIsDismissed] = useState(false);

  if (isDismissed) {
    return null;
  }

  return (
    <div
      className="bg-red-50 border border-red-200 rounded-lg p-4"
      role="alert"
      aria-live="assertive"
    >
      <div className="flex items-start gap-3">
        {/* Error icon */}
        <svg
          className="w-5 h-5 text-red-600 flex-shrink-0 mt-0.5"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>

        {/* Error content */}
        <div className="flex-1 min-w-0">
          <h3 className="text-sm font-medium text-red-800">Error</h3>
          <p className="mt-1 text-sm text-red-700">{error}</p>

          {/* Actions */}
          <div className="mt-3 flex items-center gap-3">
            <button
              type="button"
              onClick={onRetry}
              className="text-sm font-medium text-red-800 hover:text-red-900 underline focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 rounded"
            >
              {retryLabel}
            </button>
          </div>
        </div>

        {/* Dismiss button */}
        <button
          type="button"
          onClick={() => setIsDismissed(true)}
          className="flex-shrink-0 text-red-400 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-2 rounded"
          aria-label="Dismiss error"
        >
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
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    </div>
  );
}

export default InlineError;
