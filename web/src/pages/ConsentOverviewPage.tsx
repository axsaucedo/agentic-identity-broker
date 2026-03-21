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

import React, { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { AppLayout } from '@components/layout/AppLayout';
import { PageTransition } from '@components/ui/PageTransition';
import { DelegationListSkeleton } from '@components/ui/Skeleton';
import { InlineError } from '@components/ui/InlineError';
import { EmptyState } from '@components/ui/EmptyState';
import { DelegationList } from '@components/consent/DelegationList';
import { RevokeGrantDialog } from '@components/consent/RevokeGrantDialog';
import { useConsent } from '@hooks/useConsent';
import { useToast } from '@components/ui/Toast';
import { consentApi } from '@services/api';

export function ConsentOverviewPage() {
  const navigate = useNavigate();
  const { delegations, loading, error, refetch } = useConsent();
  const { showToast } = useToast();

  // Agent being revoked (null when dialog is closed)
  const [revokingAgentId, setRevokingAgentId] = useState<string | null>(null);
  const [isRevoking, setIsRevoking] = useState(false);

  // Derive agent name from the delegation list
  const revokingAgentName = revokingAgentId
    ? (delegations.find((d) => d.agentId === revokingAgentId)?.displayName ??
      '')
    : '';

  const handleRevoke = useCallback((agentId: string) => {
    setRevokingAgentId(agentId);
  }, []);

  const handleRevokeConfirm = async () => {
    if (!revokingAgentId) return;
    setIsRevoking(true);
    try {
      await consentApi.deleteGrant(revokingAgentId);
      setRevokingAgentId(null);
      showToast('All access has been revoked.', 'success');
      await refetch();
    } catch (err: unknown) {
      const message =
        err && typeof err === 'object' && 'message' in err
          ? String((err as { message: string }).message)
          : 'Failed to revoke access. Please try again.';
      showToast(message, 'error');
    } finally {
      setIsRevoking(false);
    }
  };

  const handleRevokeClose = () => {
    if (!isRevoking) setRevokingAgentId(null);
  };

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
          {/* Page header */}
          <div>
            <h2 className="text-2xl font-semibold text-neutral-900">
              My Agent Delegations
            </h2>
            <p className="mt-2 text-neutral-600">
              Manage which agents have access to your third-party services.
            </p>
          </div>

          {/* Content area with three states */}
          <div>
            {/* Loading State */}
            {loading && (
              <div>
                <h3 className="text-lg font-medium text-neutral-900 mb-4">
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
                <h3 className="text-lg font-medium text-neutral-900 mb-4">
                  Your Delegations
                </h3>

                {/* Empty state - no delegations */}
                {delegations.length === 0 && (
                  <EmptyState
                    icon={
                      <svg
                        className="mx-auto h-12 w-12 text-neutral-400"
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
                    onRevoke={handleRevoke}
                  />
                )}
              </div>
            )}
          </div>
        </div>
      </PageTransition>

      {/* Revoke confirmation dialog — controlled by revokingAgentId state */}
      <RevokeGrantDialog
        agentName={revokingAgentName}
        isOpen={revokingAgentId !== null}
        onConfirm={handleRevokeConfirm}
        onClose={handleRevokeClose}
        isLoading={isRevoking}
      />
    </AppLayout>
  );
}

export default ConsentOverviewPage;
