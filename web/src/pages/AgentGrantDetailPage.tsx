/**
 * AgentGrantDetailPage - Detailed view for managing grants to a specific agent.
 *
 * Shows:
 * - Agent details (logo, name, description)
 * - Links to governance and documentation
 * - List of available services with scopes
 * - Existing grants and their status
 * - Interactive grant editing (Phase 5)
 *
 * Features:
 * - Toast notifications for success/error feedback
 * - Page transition animations
 * - Smooth scroll to errors on validation failure
 */

import { useState, useCallback, useEffect, useMemo } from 'react';
import { useParams, useLocation, useNavigate } from 'react-router-dom';
import { AppLayout } from '@components/layout/AppLayout';
import { PageTransition } from '@components/ui/PageTransition';
import { Skeleton } from '@components/ui/Skeleton';
import { InlineError } from '@components/ui/InlineError';
import { Button } from '@components/ui/Button';
import { useToast } from '@components/ui/Toast';
import { Breadcrumb } from '@design-system/components/navigation/Breadcrumb';
import { Card } from '@design-system/components/data-display/Card';
import { ServiceCard } from '@components/consent/ServiceCard';
import { RevokeGrantButton } from '@components/consent/RevokeGrantButton';
import { CIMDConsentSummary } from '@components/consent/CIMDConsentSummary';
import { CIMDDomainBadge } from '@components/consent/CIMDDomainBadge';
import { CIMDLocalhostWarning } from '@components/consent/CIMDLocalhostWarning';
import { CIMDAdvancedDetails } from '@components/consent/CIMDAdvancedDetails';
import { GrantValidityControl } from '@components/consent/GrantValidityControl';
import { useAgentGrants } from '../hooks/useAgentGrants';
import { useToggleGrant } from '../hooks/useToggleGrant';
import { useUpdateValidity } from '../hooks/useUpdateValidity';
import { validateGrantRequest } from '../utils/validation';
import { scrollToError } from '../utils/scrollToError';
import type { DelegatedToken } from '../types/consent';

/**
 * AgentGrantDetailPage displays detailed agent information and service grants.
 * Allows users to view and edit which services and scopes are granted to the agent.
 */
export function AgentGrantDetailPage() {
  const { agentId } = useParams<{ agentId: string }>();
  const { showToast } = useToast();
  const location = useLocation();
  const navigate = useNavigate();

  // Extract redirect_uri from query parameters (FR-025)
  const searchParams = new URLSearchParams(location.search);
  const redirectUri = searchParams.get('redirect_uri') || undefined;

  // Parse the original OAuth2 authorize URL to extract CIMD params.
  // Memoized so the object reference is stable across renders — prevents infinite
  // refetch loop in useAgentGrants (which includes cimdParams in useCallback deps).
  const cimdParams = useMemo(() => {
    if (!redirectUri) return undefined;
    try {
      const authorizeUrl = new URL(redirectUri, window.location.origin);
      const authorizeParams = authorizeUrl.searchParams;
      const maybeClientId = authorizeParams.get('client_id') || undefined;
      if (maybeClientId?.startsWith('https://')) {
        return {
          clientId: maybeClientId,
          redirectUri: authorizeParams.get('redirect_uri') || '',
          scope: authorizeParams.get('scope') || '',
        };
      }
    } catch {
      // ignore parse errors
    }
    return undefined;
  }, [redirectUri]);

  const resolvedAgentId = agentId ?? '';

  // Fetch agent data and grants
  const { agent, services, cimdMeta, grants, loading, error, refetch } =
    useAgentGrants(resolvedAgentId, cimdParams);

  // Grant toggle hook
  const {
    delegatedTokens,
    setDelegatedTokens,
    isSubmitting,
    error: submitError,
    submit,
    clearError,
  } = useToggleGrant(resolvedAgentId);

  // Initialize delegatedTokens from loaded grant on mount or when grant changes
  useEffect(() => {
    if (grants) {
      // Extract delegated tokens from grant to populate the form state
      const initialTokens = grants.delegated_oauth2_tokens || [];
      if (initialTokens.length > 0) {
        setDelegatedTokens(initialTokens);
      }
    }
  }, [grants, setDelegatedTokens]);

  // Validity hook (initialized from grant if exists)
  const {
    validityState,
    setValidityState,
    getValidUntil,
    validate: validateValidity,
  } = useUpdateValidity(grants || null);

  // Track if form has changes
  // Validation errors
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  // Track unmet mandatory requirements (FR-024: disable Approve button until satisfied)
  const [hasUnmetMandatoryRequirements, setHasUnmetMandatoryRequirements] =
    useState(false);

  // Handle grants change
  // Handle validity change
  const handleValidityChange = (state: typeof validityState) => {
    setValidityState(state);
    setValidationErrors([]);
  };

  // Validate form
  const validateForm = (): boolean => {
    const errors: string[] = [];

    // Require at least one service only when the agent has mandatory service requirements.
    // Optional services are never required — the user may approve without delegating any.
    const hasMandatoryRequirements = services.some(
      (s) => s.requirementType === 'mandatory',
    );
    const requireAtLeastOneService = hasMandatoryRequirements;

    // Validate grant request
    const grantErrors = validateGrantRequest({
      delegatedTokens,
      validUntil: getValidUntil(),
      requireAtLeastOneService,
    });
    errors.push(...grantErrors);

    // Validate validity
    const validityError = validateValidity();
    if (validityError) {
      errors.push(validityError);
    }

    setValidationErrors(errors);

    // Scroll to first error if validation failed
    if (errors.length > 0) {
      setTimeout(() => {
        scrollToError();
      }, 100);
    }

    return errors.length === 0;
  };

  // Handle form submission
  const handleSubmit = async () => {
    console.log(
      '[AgentGrantDetailPage] handleSubmit - delegatedTokens:',
      delegatedTokens,
    );
    console.log('[AgentGrantDetailPage] redirectUri:', redirectUri);

    // Clear previous errors
    clearError();
    setValidationErrors([]);

    // Validate form
    if (!validateForm()) {
      return;
    }

    // Submit grant
    const validUntil = getValidUntil();
    console.log(
      '[AgentGrantDetailPage] submitting grant request with validUntil:',
      validUntil,
    );
    const result = await submit(validUntil, redirectUri);
    console.log('[AgentGrantDetailPage] submit result:', result);

    // Check for errors (result can be null for successful revocation or redirect)
    if (submitError) {
      // Show error toast
      showToast(submitError, 'error');
      return;
    }

    // If redirectUri was provided and result is null, we've been redirected
    // (handled by window.location.href in the API service)
    if (!result && redirectUri) {
      return;
    }

    // Success - refetch data and show toast notification
    await refetch();

    // Show appropriate success message
    if (result) {
      showToast('Grant updated successfully!', 'success');
    } else {
      showToast('Grant revoked successfully!', 'success');
    }
  };

  // Handle service login (FR-020a: redirect to third-party OAuth2 flow)
  // Extract redirect logic to separate function for testability
  const buildServiceLoginUrl = useCallback((serviceId: string): string => {
    const currentUrl = window.location.href;
    return `/api/third-party/${serviceId}/oauth2/authorize?redirect_uri=${encodeURIComponent(currentUrl)}`;
  }, []);

  const handleServiceLogin = useCallback(
    (serviceId: string) => {
      const loginUrl = buildServiceLoginUrl(serviceId);
      window.location.href = loginUrl;
    },
    [buildServiceLoginUrl],
  );

  // Handle service delegation (select service for grant)
  const handleDelegate = useCallback(
    (serviceId: string) => {
      const service = services.find((s) => s.serviceId === serviceId);
      if (!service) return;

      // If service is not connected, initiate login first
      if (service.connectionStatus !== 'connected') {
        handleServiceLogin(serviceId);
        return;
      }

      // Otherwise, add to delegated tokens with all required scopes
      const scopesToDelegate = service.requiredScopes
        ? service.requiredScopes.map((s) => s.name)
        : service.scopes
          ? service.scopes.map((s) => s.value)
          : [];

      const newToken = {
        thirdparty_oauth2_service_id: serviceId,
        scopes: scopesToDelegate,
      };

      // Add or update this service in delegated tokens
      const existingIndex = delegatedTokens.findIndex(
        (t) => t.thirdparty_oauth2_service_id === serviceId,
      );

      const updated = [...delegatedTokens];
      if (existingIndex >= 0) {
        updated[existingIndex] = newToken;
      } else {
        updated.push(newToken);
      }

      setDelegatedTokens(updated);
    showToast('Service selected for delegation', 'success');
  },
    [services, delegatedTokens, handleServiceLogin, setDelegatedTokens, showToast],
  );

  // Handle full grant deletion (Revoke All Access button)
  const handleGrantRevoked = useCallback(() => {
    navigate('/');
    showToast('All access for this agent has been revoked.', 'success');
  }, [navigate, showToast]);

  // Handle service revocation (deselect service from grant)
  const handleRevoke = useCallback(
    (serviceId: string) => {
      const updated = delegatedTokens.filter(
        (t) => t.thirdparty_oauth2_service_id !== serviceId,
      );
      setDelegatedTokens(updated);
      showToast('Service removed from delegation', 'success');
    },
    [delegatedTokens, setDelegatedTokens, showToast],
  );

  // Update hasUnmetMandatoryRequirements whenever agent or services change
  useEffect(() => {
    if (!agent || !services) {
      setHasUnmetMandatoryRequirements(false);
      return;
    }

    // Filter for services that are requirements (have requirementType) and are mandatory
    const mandatoryRequirements = services.filter(
      (s) =>
        s.requirementType === 'mandatory' &&
        s.connectionStatus === 'not_connected',
    );

    const hasUnmet = mandatoryRequirements.length > 0;
    setHasUnmetMandatoryRequirements(hasUnmet);
  }, [agent, services]);

  // Validate agentId parameter (after hooks to keep hook order stable)
  if (!agentId) {
    return (
      <AppLayout>
        <div className="space-y-6">
          <InlineError
            error="Invalid agent ID in URL"
            onRetry={() => window.history.back()}
          />
        </div>
      </AppLayout>
    );
  }

  // Loading state
  if (loading) {
    return (
      <AppLayout>
        <div className="space-y-6">
          {/* Breadcrumb skeleton */}
          <Breadcrumb
            items={[
              { label: 'Delegations', href: '/' },
              { label: 'Loading...' },
            ]}
          />

          {/* Agent header skeleton */}
          <Card padding="default">
            <div className="flex items-start gap-4">
              <Skeleton width="80px" height="80px" className="rounded-lg" />
              <div className="flex-1 space-y-3">
                <Skeleton width="60%" height="2rem" />
                <Skeleton width="80%" height="1rem" />
                <Skeleton width="40%" height="1rem" />
              </div>
            </div>
          </Card>

          {/* Services skeleton */}
          <div className="space-y-4">
            <Skeleton width="150px" height="1.5rem" />
            <div className="space-y-4">
              <Skeleton width="100%" height="200px" className="rounded-lg" />
              <Skeleton width="100%" height="200px" className="rounded-lg" />
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  // Error state
  if (error) {
    return (
      <AppLayout>
        <div className="space-y-6">
          {/* Breadcrumb */}
          <Breadcrumb
            items={[
              { label: 'Delegations', href: '/' },
              { label: 'Agent Details' },
            ]}
          />

          <InlineError error={error} onRetry={refetch} />
        </div>
      </AppLayout>
    );
  }

  // No agent data (should not happen if no error)
  if (!agent) {
    return (
      <AppLayout>
        <div className="space-y-6">
          <InlineError error="Agent data not available" onRetry={refetch} />
        </div>
      </AppLayout>
    );
  }

  // Get delegated tokens for a specific service from the current grant
  const getDelegatedTokensForService = (
    serviceId: string,
  ): DelegatedToken[] => {
    if (!grants) {
      return [];
    }
    return (grants.delegated_oauth2_tokens || []).filter(
      (token) =>
        token != null && token.thirdparty_oauth2_service_id === serviceId,
    );
  };

  // Check if service is currently selected for delegation in this session
  const isServiceCurrentlyDelegated = (serviceId: string): boolean => {
    return delegatedTokens.some(
      (token) => token.thirdparty_oauth2_service_id === serviceId,
    );
  };

  return (
    <AppLayout>
      <PageTransition>
        <div className="space-y-6">
          {/* Breadcrumb */}
          <Breadcrumb
            items={[
              { label: 'Delegations', href: '/' },
              { label: agent.displayName },
            ]}
          />

          {/* Agent header card */}
          <Card padding="default">
            <div className="flex items-start gap-6">
              {/* Agent logo */}
              <div className="flex-shrink-0">
                {agent.logoUrl ? (
                  <img
                    src={agent.logoUrl}
                    alt={`${agent.displayName} logo`}
                    className="w-20 h-20 rounded-lg object-cover"
                  />
                ) : (
                  <div className="w-20 h-20 bg-gradient-to-br from-success-primary to-success-primary rounded-lg flex items-center justify-center">
                    <span className="text-white text-2xl font-semibold">
                      {agent.displayName.charAt(0).toUpperCase()}
                    </span>
                  </div>
                )}
              </div>

              {/* Agent info */}
              <div className="flex-1 min-w-0">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <h1
                      className="text-2xl font-bold text-trust-deep"
                      data-testid="agent-name-heading"
                    >
                      {agent.displayName}
                    </h1>
                    <p className="mt-2 text-slate-600">{agent.description}</p>
                  </div>
                </div>

                {/* Agent links */}
                <div className="mt-4 flex flex-wrap items-center gap-4">
                  {agent.governanceUrl && (
                    <a
                      href={agent.governanceUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 text-sm font-medium text-trust-hover hover:text-trust focus:outline-none focus:ring-2 focus:ring-trust focus:ring-offset-2 rounded"
                    >
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                        />
                      </svg>
                      Governance
                    </a>
                  )}
                  {agent.userDocumentationUrl && (
                    <a
                      href={agent.userDocumentationUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 text-sm font-medium text-trust-hover hover:text-trust focus:outline-none focus:ring-2 focus:ring-trust focus:ring-offset-2 rounded"
                    >
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"
                        />
                      </svg>
                      Documentation
                    </a>
                  )}
                  {agent.agentInterfaceUrl && (
                    <a
                      href={agent.agentInterfaceUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-2 text-sm font-medium text-trust-hover hover:text-trust focus:outline-none focus:ring-2 focus:ring-trust focus:ring-offset-2 rounded"
                    >
                      <svg
                        className="w-4 h-4"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                        />
                      </svg>
                      Agent Interface
                    </a>
                  )}
                </div>
              </div>
            </div>
          </Card>

          {/* CIMD metadata section — shown only for URL-based (CIMD) agents */}
          {cimdMeta && (
            <div className="space-y-3">
              <CIMDConsentSummary
                clientName={cimdMeta.client_name}
                accessTarget={services.map((s) => s.displayName || s.serviceName || s.serviceId).join(', ') || 'requested services'}
                logoUri={cimdMeta.logo_uri}
              />
              <CIMDDomainBadge domain={cimdMeta.verified_domain} />
              {cimdMeta.is_localhost_redirect && (
                <CIMDLocalhostWarning agentDisplayName={agent.displayName} />
              )}
              <CIMDAdvancedDetails
                clientName={cimdMeta.client_name}
                clientIdUrl={cimdMeta.client_id_url}
                redirectUri={cimdMeta.redirect_uri}
                requestedScopes={cimdMeta.requested_scopes}
              />
            </div>
          )}

          {/* Validation errors */}
          {validationErrors.length > 0 && (
            <div
              data-error="true"
              className="bg-amber-50 border border-amber-200 rounded-lg p-4"
              role="alert"
              aria-live="assertive"
            >
              <div className="flex items-start gap-3">
                <svg
                  className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <div className="flex-1">
                  <h3 className="text-sm font-medium text-amber-800">
                    Validation Error
                  </h3>
                  <ul className="mt-2 text-sm text-amber-700 list-disc list-inside">
                    {validationErrors.map((error, index) => (
                      <li key={index}>{error}</li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Submit error */}
          {submitError && (
            <InlineError error={submitError} onRetry={handleSubmit} />
          )}

          {/* Grant validity control */}
          <div className="card p-6">
            <h3 className="text-lg font-semibold text-trust-deep mb-4">
              Grant Validity
            </h3>
            <GrantValidityControl
              value={validityState}
              onChange={handleValidityChange}
            />
          </div>

          {/* Services section - Flattened layout with mandatory services first */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <h2 className="text-xl font-semibold text-trust-deep">
                  Services
                  <span className="ml-2 text-sm font-normal text-slate-500">
                    ({services.length})
                  </span>
                </h2>
                <p className="mt-2 text-sm text-slate-600">
                  Delegate your permissions in these services to{' '}
                  {agent.displayName}. The agent will use these services on your
                  behalf.
                </p>
              </div>
            </div>

            {services.length === 0 ? (
              <Card padding="default">
                <div className="text-center">
                  <svg
                    className="mx-auto h-12 w-12 text-slate-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"
                    />
                  </svg>
                  <h3 className="mt-4 text-lg font-medium text-trust-deep">
                    No services available
                  </h3>
                  <p className="mt-2 text-slate-600">
                    This agent has no services configured yet.
                  </p>
                </div>
              </Card>
            ) : (
              <div className="space-y-4">
                {/* Sort services: mandatory first, then optional, then others */}
                {services
                  .sort((a, b) => {
                    if (
                      a.requirementType === 'mandatory' &&
                      b.requirementType !== 'mandatory'
                    )
                      return -1;
                    if (
                      a.requirementType !== 'mandatory' &&
                      b.requirementType === 'mandatory'
                    )
                      return 1;
                    if (
                      a.requirementType === 'optional' &&
                      b.requirementType !== 'optional'
                    )
                      return -1;
                    if (
                      a.requirementType !== 'optional' &&
                      b.requirementType === 'optional'
                    )
                      return 1;
                    return 0;
                  })
                  .map((service) => (
                    <ServiceCard
                      key={service.serviceId}
                      service={service}
                      grants={getDelegatedTokensForService(service.serviceId)}
                      isDelegated={isServiceCurrentlyDelegated(
                        service.serviceId,
                      )}
                      onDelegate={handleDelegate}
                      onRevoke={handleRevoke}
                    />
                  ))}
              </div>
            )}
          </div>

          {/* Action buttons - Always visible (FR-023), Approve disabled until mandatory requirements met (FR-024) */}
          <div className="flex items-center justify-between gap-3 pt-4">
            {/* Revoke All Access — only visible when user has an active grant */}
            {grants !== null &&
              grants.delegated_oauth2_tokens &&
              grants.delegated_oauth2_tokens.length > 0 && (
                <RevokeGrantButton
                  agentId={resolvedAgentId}
                  agentName={agent.displayName}
                  onRevoked={handleGrantRevoked}
                />
              )}

            <Button
              variant="primary"
              onClick={handleSubmit}
              isLoading={isSubmitting}
              disabled={hasUnmetMandatoryRequirements || isSubmitting}
              title={
                hasUnmetMandatoryRequirements
                  ? 'Please connect all required services first'
                  : 'Approve and delegate'
              }
            >
              Approve & Delegate
            </Button>
          </div>
        </div>
      </PageTransition>
    </AppLayout>
  );
}

export default AgentGrantDetailPage;
