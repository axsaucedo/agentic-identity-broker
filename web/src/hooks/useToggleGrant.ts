/**
 * useToggleGrant hook - Manages grant creation/modification state.
 *
 * Features:
 * - Track granted permission sets with per-PS included service IDs
 * - Submit grant requests with optimistic updates
 * - Rollback on error
 * - Loading state management
 */

import { useState, useCallback } from 'react';
import { isAxiosError } from 'axios';
import type {
  CreateOrUpdateGrantRequest,
  GrantResult,
} from '../types/consent';
import { consentApi } from '../services/api/consent';
import { extractApiError } from '../utils/api';

interface UseToggleGrantState {
  /** Granted permission sets: map of PS ID → included service IDs */
  grantedPermissionSets: Record<string, string[]>;
  /** Whether submission is in progress */
  isSubmitting: boolean;
  /** Error message if submission failed */
  error: string | null;
  /** Success state after submission */
  isSuccess: boolean;
}

interface UseToggleGrantReturn extends UseToggleGrantState {
  /** Update granted permission sets */
  setGrantedPermissionSets: (ps: Record<string, string[]>) => void;
  /** Submit grant request */
  submit: (
    validUntil?: string | null,
    submitOptions?: { redirectUri?: string; sessionToken?: string },
    grantedPermissionSets?: Record<string, string[]>,
  ) => Promise<GrantResult | undefined>;
  /** Reset state */
  reset: () => void;
  /** Clear error */
  clearError: () => void;
}

/**
 * Custom hook for managing grant toggle operations.
 * Handles form state, submission, and error handling.
 *
 * @param agentId - Unique agent identifier
 * @returns Grant state and control methods
 */
export function useToggleGrant(agentId: string): UseToggleGrantReturn {
  const [state, setState] = useState<UseToggleGrantState>({
    grantedPermissionSets: {},
    isSubmitting: false,
    error: null,
    isSuccess: false,
  });

  /**
   * Update granted permission sets selection.
   */
  const setGrantedPermissionSets = useCallback((ps: Record<string, string[]>) => {
    setState((prev) => ({
      ...prev,
      grantedPermissionSets: ps,
      isSuccess: false,
      error: null,
    }));
  }, []);

  /**
   * Submit grant request to backend.
   * Uses optimistic updates and rolls back on error.
   * If redirectUri is provided and backend returns 303, navigates to the redirect URL.
   */
  const submit = useCallback(
    async (
      validUntil?: string | null,
      submitOptions?: { redirectUri?: string; sessionToken?: string },
      grantedPermissionSets?: Record<string, string[]>,
    ): Promise<GrantResult | undefined> => {
      // Set submitting state
      setState((prev) => ({
        ...prev,
        isSubmitting: true,
        error: null,
        isSuccess: false,
      }));

      try {
        // Prepare request payload — prefer the caller-supplied value to avoid stale closure
        const request: CreateOrUpdateGrantRequest = {
          granted_permission_sets: grantedPermissionSets ?? state.grantedPermissionSets,
          valid_until: validUntil || undefined,
        };

        const result = await consentApi.createOrUpdateGrant(
          agentId,
          request,
          submitOptions,
        );

        // Update state with success
        setState((prev) => ({
          ...prev,
          isSubmitting: false,
          isSuccess: true,
          error: null,
        }));

        return result;
      } catch (err: unknown) {
        let errorMessage = extractApiError(err, 'Failed to update grant');

        if (isAxiosError(err) && err.response?.data?.details) {
          const details = Object.entries(
            err.response.data.details as Record<string, string[]>,
          )
            .map(([field, errors]) => `${field}: ${errors.join(', ')}`)
            .join('; ');
          errorMessage = `Validation error: ${details}`;
        }

        // Update state with error and throw so the caller can react without
        // reading stale closed-over hook state.
        setState((prev) => ({
          ...prev,
          isSubmitting: false,
          error: errorMessage,
          isSuccess: false,
        }));

        throw new Error(errorMessage);
      }
    },
    [agentId, state.grantedPermissionSets],
  );

  /**
   * Reset state to initial values.
   */
  const reset = useCallback(() => {
    setState({
      grantedPermissionSets: {},
      isSubmitting: false,
      error: null,
      isSuccess: false,
    });
  }, []);

  /**
   * Clear error message.
   */
  const clearError = useCallback(() => {
    setState((prev) => ({
      ...prev,
      error: null,
    }));
  }, []);

  return {
    ...state,
    setGrantedPermissionSets,
    submit,
    reset,
    clearError,
  };
}

export default useToggleGrant;
