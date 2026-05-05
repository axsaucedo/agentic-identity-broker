/**
 * useToggleGrant hook - Manages grant creation/modification state.
 *
 * Features:
 * - Track selected services and scopes
 * - Submit grant requests with optimistic updates
 * - Rollback on error
 * - Loading state management
 */

import { useState, useCallback } from 'react';
import { isAxiosError } from 'axios';
import type {
  DelegatedToken,
  CreateOrUpdateGrantRequest,
  UserGrant,
} from '../types/consent';
import { consentApi } from '../services/api/consent';
import { extractApiError } from '../utils/api';

interface UseToggleGrantState {
  /** Selected delegated tokens */
  delegatedTokens: DelegatedToken[];
  /** Whether submission is in progress */
  isSubmitting: boolean;
  /** Error message if submission failed */
  error: string | null;
  /** Success state after submission */
  isSuccess: boolean;
}

interface UseToggleGrantReturn extends UseToggleGrantState {
  /** Update selected delegated tokens */
  setDelegatedTokens: (tokens: DelegatedToken[]) => void;
  /** Submit grant request */
  submit: (
    validUntil?: string | null,
    submitOptions?: { redirectUri?: string; sessionToken?: string },
  ) => Promise<UserGrant | null>;
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
    delegatedTokens: [],
    isSubmitting: false,
    error: null,
    isSuccess: false,
  });

  /**
   * Update delegated tokens selection.
   */
  const setDelegatedTokens = useCallback((tokens: DelegatedToken[]) => {
    setState((prev) => ({
      ...prev,
      delegatedTokens: tokens,
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
    ): Promise<UserGrant | null> => {
      // Set submitting state
      setState((prev) => ({
        ...prev,
        isSubmitting: true,
        error: null,
        isSuccess: false,
      }));

      try {
        // Prepare request payload - already in snake_case format
        const request: CreateOrUpdateGrantRequest = {
          delegated_oauth2_tokens: state.delegatedTokens.map((token) => ({
            thirdparty_oauth2_service_id: token.thirdparty_oauth2_service_id,
            scopes: token.scopes,
          })),
          valid_until: validUntil || undefined,
        };

        const grant = await consentApi.createOrUpdateGrant(
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

        return grant;
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

        // Update state with error
        setState((prev) => ({
          ...prev,
          isSubmitting: false,
          error: errorMessage,
          isSuccess: false,
        }));

        throw new Error(errorMessage);
      }
    },
    [agentId, state.delegatedTokens],
  );

  /**
   * Reset state to initial values.
   */
  const reset = useCallback(() => {
    setState({
      delegatedTokens: [],
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
    setDelegatedTokens,
    submit,
    reset,
    clearError,
  };
}

export default useToggleGrant;
