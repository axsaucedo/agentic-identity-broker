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
import type {
  AgentDetail,
  ThirdpartyService,
  CIMDMetadata,
  UserGrant,
} from '../types/consent';
import { consentApi } from '../services/api/consent';

interface UseAgentGrantsState {
  /** Agent detail information */
  agent: AgentDetail | null;
  /** Available third-party services */
  services: ThirdpartyService[];
  /** CIMD metadata when the authorization request uses a URL-based client_id */
  cimdMeta: CIMDMetadata | null;
  /** User's existing grant for this agent (null if no grant exists) */
  grants: UserGrant | null;
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
 * @param cimdParams - Optional CIMD query params for URL-based client_id requests
 * @returns Agent data, grants, CIMD metadata, loading state, error, and refetch function
 */
export function useAgentGrants(
  agentId: string,
  cimdParams?: { clientId: string; redirectUri: string; scope: string },
): UseAgentGrantsReturn {
  const [state, setState] = useState<UseAgentGrantsState>({
    agent: null,
    services: [],
    cimdMeta: null,
    grants: null,
    loading: true,
    error: null,
  });

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setState({
        agent: null,
        services: [],
        cimdMeta: null,
        grants: null,
        loading: false,
        error: 'Invalid agent ID',
      });
      return;
    }

    setState((prev) => ({
      ...prev,
      loading: true,
      error: null,
    }));

    try {
      const [agentDetailData, grantsData] = await Promise.all([
        consentApi.getAgentDetail(agentId, cimdParams),
        consentApi.getAgentGrants(agentId),
      ]);

      setState({
        agent: agentDetailData.agent,
        services: agentDetailData.services,
        cimdMeta: agentDetailData.cimd_metadata ?? null,
        grants: grantsData,
        loading: false,
        error: null,
      });
    } catch (err: unknown) {
      let errorMessage = 'Failed to load agent details';

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
        cimdMeta: null,
        grants: null,
        loading: false,
        error: errorMessage,
      });
    }
  }, [agentId, cimdParams]);

  /**
   * Refetch function that can be called manually.
   * Useful for retry after errors or refreshing data.
   */
  const refetch = useCallback(async () => {
    await fetchData();
  }, [fetchData]);

  // Fetch data on mount or when agentId changes
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

export default useAgentGrants;
