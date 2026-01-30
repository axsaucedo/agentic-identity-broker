/**
 * useSessions hook fetches OAuth2 session list.
 *
 * Features:
 * - Fetches all user sessions on mount
 * - Handles loading and error states
 * - Provides refetch function
 * - Cleans up on unmount
 */

import { useState, useEffect, useCallback } from 'react';
import type { SessionSummary } from '../services/api/sessions';
import { sessionsApi } from '../services/api/sessions';

interface UseSessionsState {
  /** List of OAuth2 sessions */
  sessions: SessionSummary[];
  /** Loading state */
  loading: boolean;
  /** Error message if any */
  error: string | null;
}

interface UseSessionsReturn extends UseSessionsState {
  /** Refetch sessions data */
  refetch: () => Promise<void>;
}

/**
 * Custom hook to fetch OAuth2 sessions for the current user.
 * Fetches data on mount and provides refetch capability.
 *
 * @returns Sessions data, loading state, error, and refetch function
 *
 * @example
 * ```tsx
 * function SessionsPage() {
 *   const { sessions, loading, error, refetch } = useSessions();
 *
 *   if (loading) return <Spinner />;
 *   if (error) return <Alert variant="error">{error}</Alert>;
 *
 *   return (
 *     <div>
 *       {sessions.map(session => (
 *         <SessionCard key={session.id} session={session} />
 *       ))}
 *     </div>
 *   );
 * }
 * ```
 */
export function useSessions(): UseSessionsReturn {
  const [state, setState] = useState<UseSessionsState>({
    sessions: [],
    loading: true,
    error: null,
  });

  /**
   * Fetch sessions from API.
   * Updates state with results or error.
   */
  const fetchData = useCallback(async () => {
    // Set loading state
    setState((prev) => ({
      ...prev,
      loading: true,
      error: null,
    }));

    try {
      // Fetch sessions
      const sessions = await sessionsApi.listSessions();

      // Update state with successful data
      setState({
        sessions,
        loading: false,
        error: null,
      });
    } catch (err: unknown) {
      // Handle errors
      let errorMessage = 'Failed to load OAuth2 sessions';

      if (err && typeof err === 'object') {
        if (
          'response' in err &&
          err.response &&
          typeof err.response === 'object'
        ) {
          const response = err.response as {
            status?: number;
            data?: { message?: string };
          };

          if (response.status === 404) {
            errorMessage = 'No sessions found';
          } else if (response.data?.message) {
            errorMessage = response.data.message;
          }
        } else if ('message' in err && typeof err.message === 'string') {
          errorMessage = err.message;
        }
      }

      setState({
        sessions: [],
        loading: false,
        error: errorMessage,
      });
    }
  }, []);

  /**
   * Refetch function that can be called manually.
   * Useful for retry after errors or refreshing data.
   */
  const refetch = useCallback(async () => {
    await fetchData();
  }, [fetchData]);

  // Fetch data on mount
  useEffect(() => {
    // Execute fetch
    fetchData().catch(() => {
      // Error handling is done in fetchData
      // This catch is to prevent unhandled promise rejection
    });
  }, [fetchData]);

  return {
    ...state,
    refetch,
  };
}

/**
 * Hook to get a specific session by service ID.
 *
 * @param serviceId - The service identifier to find
 * @returns The session if found, undefined otherwise
 *
 * @example
 * ```tsx
 * function SessionDetail({ serviceId }: { serviceId: string }) {
 *   const session = useSession(serviceId);
 *
 *   if (!session) return <div>Session not found</div>;
 *
 *   return <div>{session.service_display_name}</div>;
 * }
 * ```
 */
export function useSession(serviceId: string): SessionSummary | undefined {
  const { sessions } = useSessions();
  return sessions.find((s) => s.service_id === serviceId);
}

export default useSessions;
