/**
 * useConsent hook manages fetching and state for user consent data.
 *
 * Fetches:
 * - User information (principal, display name)
 * - Agent delegations list
 *
 * Features:
 * - Parallel data fetching
 * - Loading and error states
 * - Refetch capability
 * - Cleanup on unmount
 */

import { useState, useEffect, useCallback } from 'react';
import { consentApi } from '@services/api/consent';
import type { UserInfo, AgentDelegation, ApiError } from '../types/consent';

interface UseConsentResult {
  /** List of agent delegations */
  delegations: AgentDelegation[];
  /** Current user information */
  userInfo: UserInfo | null;
  /** Loading state */
  loading: boolean;
  /** Error state if fetch failed */
  error: ApiError | null;
  /** Function to re-fetch data */
  refetch: () => Promise<void>;
}

/**
 * Custom hook to fetch and manage consent-related data.
 * Fetches user info and agent delegations in parallel on mount.
 *
 * @returns Consent data, loading state, error state, and refetch function
 *
 * @example
 * ```tsx
 * const { delegations, userInfo, loading, error, refetch } = useConsent();
 *
 * if (loading) return <Skeleton />;
 * if (error) return <Error onRetry={refetch} />;
 * return <DelegationList delegations={delegations} />;
 * ```
 */
export function useConsent(): UseConsentResult {
  const [delegations, setDelegations] = useState<AgentDelegation[]>([]);
  const [userInfo, setUserInfo] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<ApiError | null>(null);

  /**
   * Fetches user info and delegations in parallel.
   */
  const fetchData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch both in parallel for better performance
      const [userInfoData, delegationsData] = await Promise.all([
        consentApi.getUserInfo(),
        consentApi.getAgentDelegations(),
      ]);

      setUserInfo(userInfoData);
      setDelegations(delegationsData);
    } catch (err) {
      console.error('Failed to fetch consent data:', err);
      setError(err as ApiError);
    } finally {
      setLoading(false);
    }
  }, []);

  /**
   * Refetch function that can be called manually (e.g., on retry).
   */
  const refetch = useCallback(async () => {
    await fetchData();
  }, [fetchData]);

  // Fetch data on mount
  useEffect(() => {
    let isMounted = true;

    const loadData = async () => {
      if (isMounted) {
        await fetchData();
      }
    };

    loadData();

    // Cleanup function
    return () => {
      isMounted = false;
    };
  }, [fetchData]);

  return {
    delegations,
    userInfo,
    loading,
    error,
    refetch,
  };
}

export default useConsent;
