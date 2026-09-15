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
import { isAxiosError } from 'axios';
import type {
  AgentDetail,
  ThirdpartyService,
  CIMDMetadata,
  UserGrant,
} from '../types/consent';
import { consentApi } from '../services/api/consent';
import { extractApiError } from '../utils/api';

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
  /** True when the backend returned session_expired — token cannot be retried */
  sessionExpired: boolean;
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
 * @param options - Optional: sessionToken for session-based CIMD flows
 * @returns Agent data, grants, CIMD metadata, loading state, error, and refetch function
 */
export function useAgentGrants(
  agentId: string,
  options?: {
    sessionToken?: string;
  },
): UseAgentGrantsReturn {
  const [state, setState] = useState<UseAgentGrantsState>({
    agent: null,
    services: [],
    cimdMeta: null,
    grants: null,
    loading: true,
    error: null,
    sessionExpired: false,
  });

  const sessionToken = options?.sessionToken;

  const fetchData = useCallback(async () => {
    if (!agentId) {
      setState({
        agent: null,
        services: [],
        cimdMeta: null,
        grants: null,
        loading: false,
        error: 'Invalid agent ID',
        sessionExpired: false,
      });
      return;
    }

    setState((prev) => ({
      ...prev,
      loading: true,
      error: null,
      sessionExpired: false,
    }));

    try {
      const [agentDetailData, grantsData] = await Promise.all([
        consentApi.getAgentDetail(agentId, sessionToken ? { sessionToken } : undefined),
        consentApi.getAgentGrants(agentId),
      ]);

      setState({
        agent: agentDetailData.agent,
        services: agentDetailData.services,
        cimdMeta: agentDetailData.cimd_metadata ?? null,
        grants: grantsData,
        loading: false,
        error: null,
        sessionExpired: false,
      });
    } catch (err: unknown) {
      if (isAxiosError(err) && err.response?.data?.error === 'session_expired') {
        setState({
          agent: null,
          services: [],
          cimdMeta: null,
          grants: null,
          loading: false,
          error: null,
          sessionExpired: true,
        });
        return;
      }

      let errorMessage: string;
      if (isAxiosError(err) && err.response?.status === 404) {
        errorMessage = 'Agent not found';
      } else {
        errorMessage = extractApiError(err, 'Failed to load agent details');
      }

      setState({
        agent: null,
        services: [],
        cimdMeta: null,
        grants: null,
        loading: false,
        error: errorMessage,
        sessionExpired: false,
      });
    }
  }, [agentId, sessionToken]);

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
