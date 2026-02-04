/**
 * Axios-based API client for consent management backend.
 *
 * Configured with:
 * - Base URL: /api (relative, proxied by Vite dev server)
 * - Request interceptor: Adds session token to headers
 * - Response interceptor: Handles errors and redirects on auth failures
 */

import axios, {
  AxiosError,
  AxiosInstance,
  InternalAxiosRequestConfig,
} from 'axios';
import type { ApiError } from '../../types/consent';

/**
 * Create and configure the axios instance with interceptors.
 */
function createApiClient(): AxiosInstance {
  const client = axios.create({
    baseURL: '/api',
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // Request interceptor: Reserved for future use
  // Note: Authentication is handled by Vite proxy in development (injects X-Remote-User header)
  // In production, authentication is handled by the reverse proxy/API gateway
  client.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      // No authentication logic here - handled externally
      return config;
    },
    (error) => {
      return Promise.reject(error);
    },
  );

  // Response interceptor: Handle errors with enhanced error messages
  client.interceptors.response.use(
    (response) => {
      // Success response - return as-is
      return response;
    },
    (error: AxiosError<ApiError>) => {
      // Handle error responses
      if (error.response) {
        const { status, data } = error.response;

        // 401 Unauthorized - Authentication failed
        if (status === 401) {
          console.warn('Unauthorized request - authentication failed');

          // In development, provide more debugging info
          const isDevMode = import.meta.env.DEV;
          const errorMsg = isDevMode
            ? 'Authentication failed. Vite proxy should inject X-Remote-User header automatically.'
            : 'Authentication failed. Please contact your administrator.';

          return Promise.reject({
            status,
            code: 'UNAUTHORIZED',
            message: errorMsg,
            retryable: false,
          } as ApiError & { retryable: boolean });
        }

        // 403 Forbidden - User doesn't have permission
        if (status === 403) {
          console.warn('Forbidden request - insufficient permissions');
          return Promise.reject({
            status,
            code: 'FORBIDDEN',
            message: "You don't have permission to access this resource.",
          } as ApiError);
        }

        // 404 Not Found - Resource doesn't exist
        if (status === 404) {
          return Promise.reject({
            status,
            code: 'NOT_FOUND',
            message: 'The requested resource was not found.',
          } as ApiError);
        }

        // 500+ Server Error - Service temporarily unavailable
        if (status >= 500) {
          return Promise.reject({
            status,
            code: 'SERVER_ERROR',
            message: 'Service temporarily unavailable. Please try again later.',
            retryable: true,
          } as ApiError & { retryable: boolean });
        }

        // Return structured error from backend
        if (data && typeof data === 'object') {
          return Promise.reject(data as ApiError);
        }
      }

      // Network error or no response (timeout, connection refused, etc.)
      if (!error.response) {
        const networkError: ApiError & { retryable: boolean } = {
          status: 0,
          code: 'NETWORK_ERROR',
          message:
            'Unable to connect to server. Please check your internet connection.',
          retryable: true,
        };
        return Promise.reject(networkError);
      }

      // Generic error fallback
      const genericError: ApiError = {
        status: error.response?.status || 500,
        code: 'UNKNOWN_ERROR',
        message: error.message || 'An unexpected error occurred',
      };
      return Promise.reject(genericError);
    },
  );

  return client;
}

/**
 * Configured axios instance for making API requests.
 * Use this instance for all API calls in the application.
 *
 * @example
 * ```typescript
 * import { apiClient } from '@services/api/client';
 *
 * const response = await apiClient.get('/consent/agents');
 * console.log(response.data);
 * ```
 */
export const apiClient = createApiClient();

/**
 * Type guard to check if an error is an ApiError.
 */
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    'code' in error &&
    'message' in error
  );
}
