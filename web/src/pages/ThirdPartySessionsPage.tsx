/**
 * ThirdPartySessionsPage
 *
 * Displays all OAuth2 sessions for the current user with management capabilities.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Grid layout of session cards
 * - Loading state with skeletons
 * - Error state with retry option
 * - Empty state when no sessions exist
 * - Confirmation dialog before termination
 * - Real-time status updates
 * - WCAG 2.1 AA compliant
 */

import React, { useState } from 'react';
import { AppLayout } from '@components/layout/AppLayout';
import { Stack } from '@design-system/components/layout/Stack';
import { Grid } from '@design-system/components/layout/Grid';
import { EmptyState } from '@design-system/components/feedback/EmptyState';
import { Skeleton } from '@design-system/components/feedback/Skeleton';
import { Alert } from '@design-system/components/feedback/Alert';
import { Button } from '@design-system/components/primitives/Button';
import { useSessions } from '@hooks/useSessions';
import { SessionCard } from '@components/sessions/SessionCard';
import { sessionsApi } from '@services/api/sessions';

/**
 * ThirdPartySessionsPage displays all OAuth2 sessions for the current user.
 *
 * Users can:
 * - View all their active sessions with status information
 * - See which agents depend on each session
 * - Terminate sessions (with confirmation warning)
 * - Re-authenticate if a session has expired
 *
 * @example
 * ```tsx
 * <Route path="/oauth2/sessions" element={<ThirdPartySessionsPage />} />
 * ```
 */
export const ThirdPartySessionsPage: React.FC = () => {
  const { sessions, loading, error, refetch } = useSessions();
  const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null);
  const [terminatingLoading, setTerminatingLoading] = useState(false);
  const [terminationError, setTerminationError] = useState<string | null>(null);

  const handleTerminate = (serviceId: string) => {
    setSelectedSessionId(serviceId);
    setTerminationError(null);
  };

  const handleTerminationConfirm = async () => {
    if (!selectedSessionId) return;

    try {
      setTerminatingLoading(true);
      setTerminationError(null);

      // Call API to terminate session
      await sessionsApi.terminateSession(selectedSessionId);

      // Close dialog and refetch sessions
      setSelectedSessionId(null);
      await refetch();
    } catch (err) {
      console.error('Failed to terminate session', err);

      // Extract error message
      let errorMessage = 'Failed to terminate session. Please try again.';
      if (err && typeof err === 'object' && 'message' in err) {
        errorMessage = (err as { message: string }).message;
      }

      setTerminationError(errorMessage);
    } finally {
      setTerminatingLoading(false);
    }
  };

  const handleTerminationCancel = () => {
    setSelectedSessionId(null);
    setTerminationError(null);
  };

  // Loading state with skeleton cards
  if (loading) {
    return (
      <AppLayout>
        <div className="min-h-screen bg-neutral-50 py-8 px-4">
          <div className="max-w-7xl mx-auto">
          <Stack gap="lg">
            <div>
              <h1 className="text-3xl font-bold text-neutral-900 mb-2">
                Third-Party Sessions
              </h1>
              <p className="text-base text-neutral-600">
                Manage your OAuth2 sessions with third-party services.
              </p>
            </div>

            <Grid columns={1} gap="md" className="sm:grid-cols-2 lg:grid-cols-3">
              {[1, 2, 3].map((i) => (
                <Skeleton key={i} variant="rounded" height="320px" />
              ))}
            </Grid>
          </Stack>
          </div>
        </div>
      </AppLayout>
    );
  }

  // Error state with retry button
  if (error) {
    return (
      <AppLayout>
        <div className="min-h-screen bg-neutral-50 py-8 px-4">
          <div className="max-w-7xl mx-auto">
          <Stack gap="lg">
            <div>
              <h1 className="text-3xl font-bold text-neutral-900 mb-2">
                Third-Party Sessions
              </h1>
              <p className="text-base text-neutral-600">
                Manage your OAuth2 sessions with third-party services.
              </p>
            </div>

            <Alert variant="error" title="Failed to Load Sessions">
              <p className="mb-4">{error}</p>
              <Button variant="outline" size="sm" onClick={refetch}>
                Retry
              </Button>
            </Alert>
          </Stack>
          </div>
        </div>
      </AppLayout>
    );
  }

  // Empty state when no sessions exist
  if (sessions.length === 0) {
    return (
      <AppLayout>
        <div className="min-h-screen bg-neutral-50 py-8 px-4">
          <div className="max-w-7xl mx-auto">
          <Stack gap="lg">
            <div>
              <h1 className="text-3xl font-bold text-neutral-900 mb-2">
                Third-Party Sessions
              </h1>
              <p className="text-base text-neutral-600">
                Manage your OAuth2 sessions with third-party services.
              </p>
            </div>

            <EmptyState
              title="No Sessions Found"
              description="You haven't authenticated with any third-party services yet. Sessions will appear here once you grant permissions to agents."
            >
              <Button variant="primary" onClick={refetch}>
                Refresh
              </Button>
            </EmptyState>
          </Stack>
          </div>
        </div>
      </AppLayout>
    );
  }

  // Find the selected session for the confirmation dialog
  const selectedSession = sessions.find((s) => s.service_id === selectedSessionId);

  // Main content with session cards
  return (
    <AppLayout>
      <div className="min-h-screen bg-neutral-50 py-8 px-4">
        <div className="max-w-7xl mx-auto">
        <Stack gap="lg">
          {/* Page header */}
          <div>
            <h1 className="text-3xl font-bold text-neutral-900 mb-2">
              Third-Party Sessions
            </h1>
            <p className="text-base text-neutral-600 leading-relaxed">
              Manage your OAuth2 sessions with third-party services. View active sessions,
              dependent agents, and terminate sessions when needed.
            </p>
          </div>

          {/* Session cards grid */}
          <Grid columns={1} gap="md" className="sm:grid-cols-2 lg:grid-cols-3">
            {sessions.map((session) => (
              <SessionCard
                key={session.id}
                session={session}
                onTerminate={handleTerminate}
                loading={terminatingLoading && selectedSessionId === session.service_id}
              />
            ))}
          </Grid>
        </Stack>
      </div>

      {/* Termination confirmation dialog */}
      {selectedSessionId && selectedSession && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
          role="dialog"
          aria-modal="true"
          aria-labelledby="dialog-title"
          onClick={handleTerminationCancel}
        >
          <div
            className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4 p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <Stack gap="md">
              {/* Dialog header */}
              <div>
                <h2 id="dialog-title" className="text-xl font-semibold text-neutral-900">
                  Terminate Session?
                </h2>
                <p className="mt-2 text-sm text-neutral-600">
                  Are you sure you want to terminate your session with{' '}
                  <strong>{selectedSession.service_display_name}</strong>?
                </p>
              </div>

              {/* Warning about dependent agents */}
              {selectedSession.dependent_agent_count > 0 && (
                <Alert variant="warning" title="Warning">
                  <p className="text-sm">
                    This session is currently used by{' '}
                    <strong>{selectedSession.dependent_agent_count}</strong> agent
                    {selectedSession.dependent_agent_count !== 1 ? 's' : ''}. Terminating
                    this session will revoke their access.
                  </p>
                </Alert>
              )}

              {/* Termination error */}
              {terminationError && (
                <Alert variant="error" title="Termination Failed">
                  <p className="text-sm">{terminationError}</p>
                </Alert>
              )}

              {/* Dialog actions */}
              <Stack direction="row" gap="md" justify="end">
                <Button
                  variant="outline"
                  size="md"
                  onClick={handleTerminationCancel}
                  disabled={terminatingLoading}
                >
                  Cancel
                </Button>
                <Button
                  variant="danger"
                  size="md"
                  onClick={handleTerminationConfirm}
                  isLoading={terminatingLoading}
                  disabled={terminatingLoading}
                >
                  Terminate Session
                </Button>
              </Stack>
            </Stack>
          </div>
        </div>
      )}
        </div>
    </AppLayout>
  );
};

export default ThirdPartySessionsPage;
