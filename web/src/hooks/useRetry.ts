/**
 * useRetry hook implements exponential backoff retry logic.
 *
 * Features:
 * - Exponential backoff (delay doubles each retry)
 * - Max retry limit (3 attempts)
 * - Max delay cap (30 seconds)
 * - Tracks retry count and next retry delay
 * - Reset functionality
 */

import { useState, useCallback } from 'react';

interface UseRetryOptions {
  /** Initial retry delay in milliseconds */
  initialDelay?: number;
  /** Maximum number of retry attempts */
  maxRetries?: number;
  /** Maximum delay cap in milliseconds */
  maxDelay?: number;
}

interface UseRetryResult {
  /** Whether a retry is currently in progress */
  isRetrying: boolean;
  /** Milliseconds until next retry (0 if not applicable) */
  nextRetryIn: number;
  /** Current retry attempt count */
  retryCount: number;
  /** Execute the retry function with exponential backoff */
  retry: () => Promise<void>;
  /** Reset retry state */
  reset: () => void;
}

/**
 * Custom hook for retrying async operations with exponential backoff.
 * Useful for resilient API calls and handling transient failures.
 *
 * @param fn - Async function to retry
 * @param options - Retry configuration options
 * @returns Retry state and control functions
 *
 * @example
 * ```tsx
 * const { retry, isRetrying, retryCount, reset } = useRetry(
 *   async () => await fetchData(),
 *   { initialDelay: 1000 }
 * );
 *
 * <button onClick={retry} disabled={isRetrying}>
 *   Retry {retryCount > 0 && `(${retryCount})`}
 * </button>
 * ```
 */
export function useRetry(
  fn: () => Promise<void>,
  options: UseRetryOptions = {},
): UseRetryResult {
  const { initialDelay = 1000, maxRetries = 3, maxDelay = 30000 } = options;

  const [isRetrying, setIsRetrying] = useState(false);
  const [retryCount, setRetryCount] = useState(0);
  const [nextRetryIn, setNextRetryIn] = useState(0);

  /**
   * Calculate exponential backoff delay.
   */
  const calculateDelay = useCallback(
    (attemptNumber: number): number => {
      const exponentialDelay = initialDelay * Math.pow(2, attemptNumber);
      return Math.min(exponentialDelay, maxDelay);
    },
    [initialDelay, maxDelay],
  );

  /**
   * Execute retry with exponential backoff.
   */
  const retry = useCallback(async () => {
    if (isRetrying) {
      console.warn('Retry already in progress');
      return;
    }

    if (retryCount >= maxRetries) {
      console.warn(`Max retries (${maxRetries}) reached`);
      return;
    }

    setIsRetrying(true);

    try {
      // Calculate delay for this attempt
      const delay = calculateDelay(retryCount);
      setNextRetryIn(delay);

      // Wait before retrying
      await new Promise((resolve) => setTimeout(resolve, delay));

      // Execute the function
      await fn();

      // Success - reset retry count
      setRetryCount(0);
      setNextRetryIn(0);
    } catch (err) {
      // Increment retry count on failure
      console.error(`Retry attempt ${retryCount + 1} failed:`, err);
      setRetryCount((prev) => prev + 1);
    } finally {
      setIsRetrying(false);
    }
  }, [fn, isRetrying, retryCount, maxRetries, calculateDelay]);

  /**
   * Reset retry state to initial values.
   */
  const reset = useCallback(() => {
    setIsRetrying(false);
    setRetryCount(0);
    setNextRetryIn(0);
  }, []);

  return {
    isRetrying,
    nextRetryIn,
    retryCount,
    retry,
    reset,
  };
}

export default useRetry;
