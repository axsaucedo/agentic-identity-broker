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

  /** Optional email address extracted from JWT claims */
  email?: string;

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
 * External OAuth2 service delegated to an agent (scoped variant).
 * Used when the backend returns a plain service with available scopes.
 */
export interface ServiceWithScopes {
  kind: 'scoped';
  serviceId: string;
  displayName?: string;
  logoUrl?: string;
  scopes?: ServiceScope[];
}

export type ThirdpartyService = ServiceWithScopes | ServiceRequirement;

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
 * CIMD metadata included in the agent detail response when the authorization
 * request originates from a Client ID Metadata Document URL (client_id).
 * Null/absent for opaque UUID-based client_id values.
 */
export interface CIMDMetadata {
  /** The CIMD URL used as client_id */
  client_id_url: string;
  /** The requested redirect_uri */
  redirect_uri: string;
  /** Hostname from client_id_url, pre-registered and verified */
  verified_domain: string;
  /** OAuth2 scopes requested by this authorization */
  requested_scopes: string[];
}

/**
 * Response from GET /api/consent/agent/:agent-id
 */
export interface GetAgentDetailResponse {
  data: {
    agent: AgentDetail;
    services: ThirdpartyService[];
    cimd_metadata?: CIMDMetadata | null;
  };
}

/**
 * Response from GET /api/consent/agent/:agent-id/grants
 * Returns a single grant (or null if no grant exists) due to 1:1 relationship per (principal, agent_id)
 */
export interface GetAgentGrantsResponse {
  data: UserGrant | null;
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
  redirect_url?: string;
}

export type GrantResult =
  | { kind: 'created'; grant: UserGrant }
  | { kind: 'noContent' }
  | { kind: 'redirect'; redirectUrl: string };

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

/**
 * Service requirement for an agent (Phase 6).
 * Specifies which services an agent needs access to.
 */
export interface ServiceRequirement {
  kind: 'requirement';
  serviceId: string;
  serviceName: string;
  requirementType: 'mandatory' | 'optional';
  requiredScopes: Array<{
    name: string;
    description?: string;
  }>;
  connectionStatus: 'connected' | 'not_connected';
  logoUrl?: string;
}

/**
 * Agent with service requirements (Phase 6).
 * Response from GET /api/consent/agent/:agent-id with requirements.
 */
export interface AgentWithServiceRequirements extends AgentDetail {
  /** List of service requirements for this agent */
  serviceRequirements: ServiceRequirement[];
}

// Note: Validation functions moved to utils/validation.ts
