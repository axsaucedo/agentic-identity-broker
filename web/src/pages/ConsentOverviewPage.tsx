/**
 * ConsentOverviewPage - Main consent management page.
 *
 * Shows list of agents the user has delegated access to.
 * Implements User Story 1: View Active Delegations.
 *
 * States:
 * 1. Loading: Shows skeleton cards while fetching data
 * 2. Error: Shows error message with retry button
 * 3. Success: Shows list of delegations or empty state
 *
 * Features page transition animations for smooth UX.
 */

import React from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '@components/layout/AppLayout';
import { PageTransition } from '@components/ui/PageTransition';
import { DelegationListSkeleton } from '@components/ui/Skeleton';
import { InlineError } from '@components/ui/InlineError';
import { EmptyState } from '@components/ui/EmptyState';
import { DelegationList } from '@components/consent/DelegationList';
import { useConsent } from '@hooks/useConsent';

export function ConsentOverviewPage() {
  const navigate = useNavigate();
  const { delegations, userInfo, loading, error, refetch } = useConsent();

  /**
   * Navigate to agent detail page when delegation card is clicked.
   */
  const handleDelegationClick = (agentId: string) => {
    navigate(`/agent/${agentId}`);
  };

  return (
    <AppLayout>
      <PageTransition>
        <div className="space-y-6">
        {/* Page header with user info */}
        <div>
          <div className="flex items-center justify-between">
            <div>
              <h2 className="text-2xl font-semibold text-gray-900">
                My Agent Delegations
              </h2>
              <p className="mt-2 text-gray-600">
                Manage which agents have access to your third-party services.
              </p>
            </div>

            {/* Display user principal when available */}
            {userInfo && (
              <div className="flex items-center gap-3">
                {userInfo.pictureUrl && (
                  <img
                    src={userInfo.pictureUrl}
                    alt={userInfo.displayName}
                    className="w-10 h-10 rounded-full"
                  />
                )}
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900">
                    {userInfo.displayName}
                  </p>
                  <p className="text-xs text-gray-500">{userInfo.principal}</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Content area with three states */}
        <div>
          {/* Loading State */}
          {loading && (
            <div>
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Loading delegations...
              </h3>
              <DelegationListSkeleton />
            </div>
          )}

          {/* Error State */}
          {!loading && error && (
            <InlineError
              error={error.message || 'Failed to load delegations'}
              onRetry={refetch}
            />
          )}

          {/* Success State */}
          {!loading && !error && (
            <div>
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Your Delegations
              </h3>

              {/* Empty state - no delegations */}
              {delegations.length === 0 && (
                <EmptyState
                  icon={
                    <svg
                      className="mx-auto h-12 w-12 text-gray-400"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                      />
                    </svg>
                  }
                  title="No active delegations"
                  description="You haven't granted access to any agents yet. When you authorize an agent to access your services, they will appear here."
                />
              )}

              {/* Delegation list */}
              {delegations.length > 0 && (
                <DelegationList
                  delegations={delegations}
                  onDelegationClick={handleDelegationClick}
                />
              )}
            </div>
          )}
        </div>
      </div>
      </PageTransition>
    </AppLayout>
  );
}

export default ConsentOverviewPage;
