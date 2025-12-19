/**
 * TypeScript types for consent management feature.
 * Maps to backend API responses from data-model.md.
 */

/**
 * Current authenticated user information.
 * Returned by GET /api/me
 */
export interface UserInfo {
  /** User's principal identifier (e.g., email, UUID) */
  principal: string;

  /** Human-readable display name */
  displayName: string;

  /** URL to user's profile picture/avatar */
  pictureUrl?: string;
}

/**
 * Summary of an agent the user has delegated access to.
 * Returned by GET /api/consent/agents
 */
export interface AgentDelegation {
  /** Unique agent identifier */
  agentId: string;

  /** Agent's display name shown to users */
  displayName: string;

  /** URL to agent's logo/avatar */
  logoUrl?: string;

  /** Number of active service grants for this agent */
  activeGrantCount: number;

  /** ISO 8601 timestamp of last grant modification */
  lastModifiedAt: string;

  /** Optional grant expiration (null = indefinite) */
  expiresAt?: string | null;
}

/**
 * Detailed agent information for grant management page.
 * Returned by GET /api/consent/agent/:agent-id
 */
export interface AgentDetail {
  /** Unique agent identifier */
  agentId: string;

  /** Agent's display name */
  displayName: string;

  /** Agent description/purpose */
  description: string;

  /** URL to agent's logo/avatar */
  logoUrl?: string;

  /** Link to agent governance documentation */
  governanceUrl?: string;

  /** Link to user-facing agent documentation */
  userDocumentationUrl?: string;

  /** Link to agent's public interface (if applicable) */
  agentInterfaceUrl?: string;
}

/**
 * External OAuth2 service that can be delegated to agents.
 * Part of GET /api/consent/agent/:agent-id response.
 */
export interface ThirdpartyService {
  /** Unique service identifier */
  serviceId: string;

  /** Service display name (e.g., "Google Drive", "GitHub") */
  displayName: string;

  /** URL to service logo */
  logoUrl?: string;

  /** Available OAuth2 scopes for this service */
  scopes: ServiceScope[];
}

/**
 * OAuth2 scope within a third-party service.
 */
export interface ServiceScope {
  /** OAuth2 scope value (e.g., "read:email") */
  value: string;

  /** Human-readable scope description */
  description: string;
}

/**
 * User's existing grant to an agent.
 * Returned by GET /api/consent/agent/:agent-id/grants
 * Note: Uses snake_case to match backend API response
 */
export interface UserGrant {
  /** Unique grant identifier */
  id: string;

  /** Agent receiving the grant */
  agent_id: string;

  /** User principal who created the grant */
  principal: string;

  /** Delegated service access tokens */
  delegated_oauth2_tokens: DelegatedToken[];

  /** Optional expiration timestamp (null = indefinite) */
  valid_until?: string | null;

  /** ISO 8601 timestamp of grant creation */
  created_at: string;

  /** ISO 8601 timestamp of last update */
  updated_at: string;
}

/**
 * Service-specific delegation within a grant.
 */
export interface DelegatedToken {
  /** Third-party service identifier */
  thirdparty_oauth2_service_id: string;

  /** OAuth2 scopes granted for this service */
  scopes: string[];
}

// API Request/Response Types

/**
 * Response from GET /api/me
 */
export interface GetUserInfoResponse {
  data: UserInfo;
}

/**
 * Response from GET /api/consent/agents
 */
export interface GetAgentDelegationsResponse {
  data: AgentDelegation[];
}

/**
 * Response from GET /api/consent/agent/:agent-id
 */
export interface GetAgentDetailResponse {
  data: {
    agent: AgentDetail;
    services: ThirdpartyService[];
  };
}

/**
 * Response from GET /api/consent/agent/:agent-id/grants
 */
export interface GetAgentGrantsResponse {
  data: UserGrant[];
}

/**
 * Request body for POST /api/consent/agent/:agent-id/grants
 */
export interface CreateOrUpdateGrantRequest {
  /** Service delegations (empty array = revoke grant) */
  delegated_oauth2_tokens: {
    thirdparty_oauth2_service_id: string;
    scopes: string[];
  }[];

  /** Optional expiration timestamp (omit for indefinite) */
  valid_until?: string | null;
}

/**
 * Response from POST /api/consent/agent/:agent-id/grants
 */
export interface CreateOrUpdateGrantResponse {
  data: UserGrant;
}

/**
 * Standard error response from backend APIs.
 */
export interface ApiError {
  /** HTTP status code */
  status: number;

  /** Error code (e.g., "INVALID_AGENT_ID") */
  code: string;

  /** Human-readable error message */
  message: string;

  /** Optional field-level validation errors */
  details?: Record<string, string[]>;
}

// UI State Types

/**
 * UI state for service grant toggle.
 */
export interface ServiceGrantState {
  serviceId: string;
  isEnabled: boolean;
  isExpanded: boolean;
  selectedScopes: Set<string>;
}

/**
 * Grant validity form state.
 */
export interface GrantValidityState {
  noExpiration: boolean;
  expiresAt?: Date;
}

/**
 * Loading state for async operations.
 */
export interface LoadingState {
  isLoading: boolean;
  error?: ApiError;
}

// Note: Validation functions moved to utils/validation.ts
