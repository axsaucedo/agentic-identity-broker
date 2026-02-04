/**
 * API service for consent management endpoints.
 *
 * Provides type-safe methods for interacting with the consent backend APIs.
 * Includes caching layer for GET requests to reduce network load.
 */

import { apiClient } from './client';
import { apiCache } from './cache';
import { isSafeRedirectUrl } from '../../utils/validation';
import type {
  UserInfo,
  AgentDelegation,
  AgentDetail,
  ThirdpartyService,
  UserGrant,
  GetUserInfoResponse,
  GetAgentDelegationsResponse,
  GetAgentDetailResponse,
  GetAgentGrantsResponse,
  CreateOrUpdateGrantRequest,
  CreateOrUpdateGrantResponse,
} from '../../types/consent';

/**
 * Consent API service class providing all consent-related endpoints.
 */
export class ConsentApiService {
  /**
   * Get current authenticated user information.
   * Results are cached for 5 minutes.
   *
   * @returns User information
   * @throws {ApiError} if request fails
   */
  async getUserInfo(): Promise<UserInfo> {
    const cacheKey = '/me';

    // Check cache first
    const cached = apiCache.get<UserInfo>(cacheKey);
    if (cached) {
      return cached;
    }

    // Fetch from API
    const response = await apiClient.get<GetUserInfoResponse>('/me');
    const data = response.data.data;

    // Cache the result (5 minutes TTL)
    apiCache.set(cacheKey, data, 5 * 60 * 1000);

    return data;
  }

  /**
   * Get all agent delegations for the current user.
   * Results are cached for 5 minutes.
   *
   * @returns Array of agent delegations
   * @throws {ApiError} if request fails
   */
  async getAgentDelegations(): Promise<AgentDelegation[]> {
    const cacheKey = '/consent/agents';

    // Check cache first
    const cached = apiCache.get<AgentDelegation[]>(cacheKey);
    if (cached) {
      return cached;
    }

    // Fetch from API
    const response =
      await apiClient.get<GetAgentDelegationsResponse>('/consent/agents');
    const data = response.data.data;

    // Cache the result (5 minutes TTL)
    apiCache.set(cacheKey, data, 5 * 60 * 1000);

    return data;
  }

  /**
   * Get detailed information about a specific agent and available services.
   * Results are cached for 5 minutes.
   *
   * @param agentId - Unique agent identifier
   * @returns Agent details and available services
   * @throws {ApiError} if request fails or agent not found
   */
  async getAgentDetail(agentId: string): Promise<{
    agent: AgentDetail;
    services: ThirdpartyService[];
  }> {
    const cacheKey = `/consent/agent/${agentId}`;

    // Check cache first
    const cached = apiCache.get<{
      agent: AgentDetail;
      services: ThirdpartyService[];
    }>(cacheKey);
    if (cached) {
      return cached;
    }

    // Fetch from API
    const response = await apiClient.get<GetAgentDetailResponse>(
      `/consent/agent/${agentId}`,
    );
    const data = response.data.data;

    // Cache the result (5 minutes TTL)
    apiCache.set(cacheKey, data, 5 * 60 * 1000);

    return data;
  }

  /**
   * Get grant for a specific agent for the current user.
   * Results are cached for 5 minutes.
   *
   * @param agentId - Unique agent identifier
   * @returns User grant for this agent (null if no grant exists)
   * @throws {ApiError} if request fails
   */
  async getAgentGrants(agentId: string): Promise<UserGrant | null> {
    const cacheKey = `/consent/agent/${agentId}/grants`;

    // Check cache first
    const cached = apiCache.get<UserGrant | null>(cacheKey);
    if (cached) {
      return cached;
    }

    // Fetch from API
    const response = await apiClient.get<GetAgentGrantsResponse>(
      `/consent/agent/${agentId}/grants`,
    );
    const data = response.data.data;

    // Cache the result (5 minutes TTL)
    apiCache.set(cacheKey, data, 5 * 60 * 1000);

    return data;
  }

  /**
   * Create or update a grant for a specific agent.
   * Invalidates relevant caches after successful update.
   *
   * @param agentId - Unique agent identifier
   * @param request - Grant configuration
   * @param redirectUri - Optional redirect URI for seamless flow continuation (FR-025)
   * @returns Created or updated grant (or null if redirected)
   * @throws {ApiError} if request fails or validation errors
   */
  async createOrUpdateGrant(
    agentId: string,
    request: CreateOrUpdateGrantRequest,
    redirectUri?: string,
  ): Promise<UserGrant | null> {
    // Build URL with optional redirect_uri query parameter (FR-025, T055)
    let url = `/consent/agent/${agentId}/grants`;
    if (redirectUri) {
      url += `?redirect_uri=${encodeURIComponent(redirectUri)}`;
    }

    const response = await apiClient.post<CreateOrUpdateGrantResponse>(
      url,
      request,
    );

    // Invalidate caches for this agent since data changed
    apiCache.invalidatePattern(`/consent/agent/${agentId}*`);
    apiCache.invalidate('/consent/agents');

    // Handle 204 No Content response (grant revoked with empty tokens)
    if (response.status === 204) {
      return null;
    }

    // Handle 201 Created response
    if (response.status === 201) {
      // Check if backend provided a redirect_url in the response body (FR-025, T056)
      // Backend returns redirect_url instead of HTTP 303 to avoid CORS issues with cross-origin redirects
      const redirectUrl = response.data.redirect_url;
      if (redirectUrl) {
        // Defense-in-depth: Validate redirect URL is same-origin before following
        // Backend already validates (SR-003), but frontend validation adds security layer
        if (!isSafeRedirectUrl(redirectUrl)) {
          throw new Error(
            'Redirect URL validation failed: URL must be same-origin',
          );
        }

        // Use window.location.href to navigate (not XHR/fetch)
        // This allows proper handling of redirect chains including cross-origin redirects
        window.location.href = redirectUrl;
        return null; // We're navigating away
      }

      return response.data.data;
    }

    // Unexpected status
    throw new Error(`Unexpected status code: ${response.status}`);
  }

  /**
   * Revoke a grant by creating an empty grant (no delegated tokens).
   *
   * @param agentId - Unique agent identifier
   * @returns Updated grant (empty delegations) or null if revoked
   * @throws {ApiError} if request fails
   */
  async revokeGrant(agentId: string): Promise<UserGrant | null> {
    return this.createOrUpdateGrant(agentId, {
      delegated_oauth2_tokens: [],
    });
  }
}

/**
 * Singleton instance of the consent API service.
 * Use this throughout the application for API calls.
 */
export const consentApi = new ConsentApiService();
