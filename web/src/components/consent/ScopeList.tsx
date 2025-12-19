/**
 * ScopeList component displays OAuth scopes for a service.
 *
 * Features:
 * - Badge-style display
 * - Scope value and description
 * - Visual indicator for granted scopes
 * - Responsive layout
 */

import React from 'react';
import type { ServiceScope } from '../../types/consent';

interface ScopeListProps {
  /** List of OAuth scopes */
  scopes: ServiceScope[];
  /** List of scope values that have been granted */
  grantedScopes?: string[];
  /** Whether the parent container is expanded */
  isExpanded: boolean;
}

/**
 * ScopeList displays a list of OAuth scopes with badge styling.
 * Highlights scopes that have been granted to the agent.
 */
export function ScopeList({ scopes, grantedScopes = [], isExpanded }: ScopeListProps) {
  if (!isExpanded || scopes.length === 0) {
    return null;
  }

  const isGranted = (scopeValue: string) => grantedScopes.includes(scopeValue);

  return (
    <div className="divide-y divide-dotted divide-slate/10">
      {scopes.map((scope, index) => {
        const granted = isGranted(scope.value);

        return (
          <div
            key={scope.value}
            className="flex items-start gap-3 py-3 first:pt-0 last:pb-0"
          >
            {/* Status indicator */}
            <div className="flex-shrink-0 mt-0.5">
              {granted ? (
                <svg
                  className="w-5 h-5 text-emerald-600"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                  aria-label="Granted"
                >
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              ) : (
                <div
                  className="w-5 h-5 rounded-full border-2 border-slate-300"
                  aria-label="Not granted"
                />
              )}
            </div>

            {/* Scope details */}
            <div className="flex-1 min-w-0">
              <div className="flex items-start justify-between gap-2">
                <code
                  className={`text-sm font-mono font-medium break-all ${
                    granted ? 'text-emerald-700' : 'text-navy-900'
                  }`}
                >
                  {scope.value}
                </code>
                {granted && (
                  <span className="flex-shrink-0 inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700">
                    Granted
                  </span>
                )}
              </div>
              <p
                className={`mt-1 text-sm ${
                  granted ? 'text-emerald-600' : 'text-slate-600'
                }`}
              >
                {scope.description}
              </p>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default ScopeList;
