/**
 * useAgentGrants hook fetches agent details and grants.
 *
 * Features:
 * - Fetches agent detail and grants in parallel
 * - Handles loading and error states
 * - Provides refetch function
 * - Cleans up on unmount
 */

import { useState, useEffect, useCallback } from 'react';
import type { AgentDetail, ThirdpartyService, UserGrant } from '../types/consent';
import { consentApi } from '../services/api/consent';

interface UseAgentGrantsState {
  /** Agent detail information */
  agent: AgentDetail | null;
  /** Available third-party services */
  services: ThirdpartyService[];
  /** User's existing grants for this agent */
  grants: UserGrant[];
  /** Loading state */
  loading: boolean;
  /** Error message if any */
  error: string | null;
}

interface UseAgentGrantsReturn extends UseAgentGrantsState {
  /** Refetch agent and grants data */
  refetch: () => Promise<void>;
}

/**
 * Custom hook to fetch agent details and grants.
 * Fetches data in parallel on mount and provides refetch capability.
 *
 * @param agentId - Unique agent identifier
 * @returns Agent data, grants, loading state, error, and refetch function
 */
export function useAgentGrants(agentId: string): UseAgentGrantsReturn {
  const [state, setState] = useState<UseAgentGrantsState>({
    agent: null,
    services: [],
    grants: [],
    loading: true,
    error: null,
  });

  /**
   * Fetch agent details and grants in parallel.
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
      // Fetch agent detail and grants in parallel
      const [agentDetailData, grantsData] = await Promise.all([
        consentApi.getAgentDetail(agentId),
        consentApi.getAgentGrants(agentId),
      ]);

      // Update state with successful data
      setState({
        agent: agentDetailData.agent,
        services: agentDetailData.services,
        grants: grantsData,
        loading: false,
        error: null,
      });
    } catch (err: unknown) {
      // Handle errors
      let errorMessage = 'Failed to load agent details';

      if (err && typeof err === 'object') {
        if ('response' in err && err.response && typeof err.response === 'object') {
          const response = err.response as { status?: number; data?: { message?: string } };

          if (response.status === 404) {
            errorMessage = 'Agent not found';
          } else if (response.data?.message) {
            errorMessage = response.data.message;
          }
        } else if ('message' in err && typeof err.message === 'string') {
          errorMessage = err.message;
        }
      }

      setState({
        agent: null,
        services: [],
        grants: [],
        loading: false,
        error: errorMessage,
      });
    }
  }, [agentId]);

  /**
   * Refetch function that can be called manually.
   * Useful for retry after errors or refreshing data.
   */
  const refetch = useCallback(async () => {
    await fetchData();
  }, [fetchData]);

  // Fetch data on mount or when agentId changes
  useEffect(() => {
    let isMounted = true;

    // Execute fetch
    fetchData().catch(() => {
      // Error handling is done in fetchData
      // This catch is to prevent unhandled promise rejection
    });

    // Cleanup function
    return () => {
      isMounted = false;
    };
  }, [fetchData]);

  return {
    ...state,
    refetch,
  };
}

export default useAgentGrants;
