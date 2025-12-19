/**
 * useUpdateValidity hook - Manages grant validity/expiration state.
 *
 * Features:
 * - Track expiration date
 * - Validate date is in future
 * - Reset functionality
 */

import { useState, useCallback } from 'react';
import type { UserGrant, GrantValidityState } from '../types/consent';

interface UseUpdateValidityReturn {
  /** Current validity state */
  validityState: GrantValidityState;
  /** Update validity state */
  setValidityState: (state: GrantValidityState) => void;
  /** Get validity as ISO string for API */
  getValidUntil: () => string | null;
  /** Validate current state */
  validate: () => string | null;
  /** Reset to initial state */
  reset: () => void;
}

/**
 * Custom hook for managing grant validity/expiration.
 *
 * @param grant - Optional existing grant to initialize from
 * @returns Validity state and control methods
 */
export function useUpdateValidity(grant: UserGrant | null = null): UseUpdateValidityReturn {
  // Initialize from existing grant if provided
  const initialState: GrantValidityState = grant?.valid_until
    ? {
        noExpiration: false,
        expiresAt: new Date(grant.valid_until),
      }
    : {
        noExpiration: true,
        expiresAt: undefined,
      };

  const [validityState, setValidityStateInternal] =
    useState<GrantValidityState>(initialState);

  /**
   * Update validity state.
   */
  const setValidityState = useCallback((state: GrantValidityState) => {
    setValidityStateInternal(state);
  }, []);

  /**
   * Get validUntil as ISO string for API request.
   * Returns null if no expiration.
   */
  const getValidUntil = useCallback((): string | null => {
    if (validityState.noExpiration || !validityState.expiresAt) {
      return null;
    }
    return validityState.expiresAt.toISOString();
  }, [validityState]);

  /**
   * Validate current validity state.
   * Returns error message if invalid, null if valid.
   */
  const validate = useCallback((): string | null => {
    // If no expiration, always valid
    if (validityState.noExpiration) {
      return null;
    }

    // Must have expiration date if checkbox is checked
    if (!validityState.expiresAt) {
      return 'Expiration date is required when grant has expiration';
    }

    // Date must be in the future
    const now = new Date();
    if (validityState.expiresAt <= now) {
      return 'Expiration date must be in the future';
    }

    return null;
  }, [validityState]);

  /**
   * Reset to initial state.
   */
  const reset = useCallback(() => {
    setValidityStateInternal(initialState);
  }, [initialState]);

  return {
    validityState,
    setValidityState,
    getValidUntil,
    validate,
    reset,
  };
}

export default useUpdateValidity;
