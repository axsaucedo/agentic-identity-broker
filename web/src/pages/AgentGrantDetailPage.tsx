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

import { useState, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { AppLayout } from '@components/layout/AppLayout';
import { PageTransition } from '@components/ui/PageTransition';
import { Skeleton } from '@components/ui/Skeleton';
import { InlineError } from '@components/ui/InlineError';
import { Button } from '@components/ui/Button';
import { Switch } from '@components/ui/Switch';
import { useToast } from '@components/ui/Toast';
import { Breadcrumb } from '@design-system/components/navigation/Breadcrumb';
import { Card } from '@design-system/components/data-display/Card';
import { Stack } from '@design-system/components/layout/Stack';
import { ServiceGrantList } from '@components/consent/ServiceGrantList';
import { ServiceCard } from '@components/consent/ServiceCard';
import { GrantValidityControl } from '@components/consent/GrantValidityControl';
import { useAgentGrants } from '../hooks/useAgentGrants';
import { useToggleGrant } from '../hooks/useToggleGrant';
import { useUpdateValidity } from '../hooks/useUpdateValidity';
import { validateGrantRequest, formatValidationErrors } from '../utils/validation';
import { scrollToError } from '../utils/scrollToError';
import type { DelegatedToken } from '../types/consent';

/**
 * AgentGrantDetailPage displays detailed agent information and service grants.
 * Allows users to view and edit which services and scopes are granted to the agent.
 */
export function AgentGrantDetailPage() {
  const { agentId } = useParams<{ agentId: string }>();
  const { showToast } = useToast();

  // State for edit mode toggle
  const [isEditMode, setIsEditMode] = useState(false);

  // Validate agentId parameter
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

  // Fetch agent data and grants
  const { agent, services, grants, loading, error, refetch } = useAgentGrants(agentId);

  // Grant toggle hook
  const {
    delegatedTokens,
    setDelegatedTokens,
    isSubmitting,
    error: submitError,
    isSuccess,
    submit,
    clearError,
  } = useToggleGrant(agentId);

  // Validity hook (initialized from first grant if exists)
  const { validityState, setValidityState, getValidUntil, validate: validateValidity } =
    useUpdateValidity(grants[0] || null);

  // Track if form has changes
  const [hasChanges, setHasChanges] = useState(false);

  // Validation errors
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  // Handle grants change
  const handleGrantsChange = useCallback((tokens: DelegatedToken[]) => {
    console.log('[AgentGrantDetailPage] handleGrantsChange called with:', tokens);
    setDelegatedTokens(tokens);
    setHasChanges(true);
    setValidationErrors([]);
  }, []);

  // Handle validity change
  const handleValidityChange = (state: typeof validityState) => {
    setValidityState(state);
    setHasChanges(true);
    setValidationErrors([]);
  };

  // Validate form
  const validateForm = (): boolean => {
    const errors: string[] = [];

    // Validate grant request
    const grantErrors = validateGrantRequest({
      delegatedTokens,
      validUntil: getValidUntil(),
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
    console.log('[AgentGrantDetailPage] handleSubmit - delegatedTokens:', delegatedTokens);

    // Clear previous errors
    clearError();
    setValidationErrors([]);

    // Validate form
    if (!validateForm()) {
      return;
    }

    // Submit grant
    const validUntil = getValidUntil();
    console.log('[AgentGrantDetailPage] submitting grant request with validUntil:', validUntil);
    const result = await submit(validUntil);
    console.log('[AgentGrantDetailPage] submit result:', result);

    // Check for errors (result can be null for successful revocation)
    if (submitError) {
      // Show error toast
      showToast(submitError, 'error');
      return;
    }

    // Success - refetch data and show toast notification
    await refetch();
    setIsEditMode(false);
    setHasChanges(false);

    // Show appropriate success message
    if (result) {
      showToast('Grant updated successfully!', 'success');
    } else {
      showToast('Grant revoked successfully!', 'success');
    }
  };

  // Handle cancel
  const handleCancel = () => {
    setIsEditMode(false);
    setHasChanges(false);
    setValidationErrors([]);
    clearError();
  };

  // Handle edit mode toggle
  const handleEditModeToggle = (enabled: boolean) => {
    if (!enabled && hasChanges) {
      // Confirm before discarding changes
      const confirmed = window.confirm(
        'You have unsaved changes. Are you sure you want to discard them?'
      );
      if (!confirmed) {
        return;
      }
    }
    setIsEditMode(enabled);
    setHasChanges(false);
    setValidationErrors([]);
    clearError();
  };


  // Loading state
  if (loading) {
    return (
      <AppLayout>
        <div className="space-y-6">
          {/* Breadcrumb skeleton */}
          <Breadcrumb items={[
            { label: 'Delegations', href: '/' },
            { label: <Skeleton width="120px" height="1rem" /> },
          ]} />

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
          <Breadcrumb items={[
            { label: 'Delegations', href: '/' },
            { label: 'Agent Details' },
          ]} />

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

  // Get delegated tokens grouped by service
  const getDelegatedTokensForService = (serviceId: string): DelegatedToken[] => {
    return grants
      .flatMap((grant) => grant.delegated_oauth2_tokens || [])
      .filter((token) => token != null && token.thirdparty_oauth2_service_id === serviceId);
  };

  return (
    <AppLayout>
      <PageTransition>
        <div className="space-y-6">
        {/* Breadcrumb */}
        <Breadcrumb items={[
          { label: 'Delegations', href: '/' },
          { label: agent.displayName },
        ]} />

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
                <div className="w-20 h-20 bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-lg flex items-center justify-center">
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
                  <h1 className="text-2xl font-bold text-navy-900">{agent.displayName}</h1>
                  <p className="mt-2 text-slate-600">{agent.description}</p>
                </div>

                {/* Edit mode toggle */}
                <Switch
                  checked={isEditMode}
                  onChange={handleEditModeToggle}
                  label={isEditMode ? 'Edit Mode' : 'View Mode'}
                />
              </div>

              {/* Agent links */}
              <div className="mt-4 flex flex-wrap items-center gap-4">
                {agent.governanceUrl && (
                  <a
                    href={agent.governanceUrl}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-2 text-sm font-medium text-navy-600 hover:text-navy-700 focus:outline-none focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 rounded"
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
                    className="inline-flex items-center gap-2 text-sm font-medium text-navy-600 hover:text-navy-700 focus:outline-none focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 rounded"
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
                    className="inline-flex items-center gap-2 text-sm font-medium text-navy-600 hover:text-navy-700 focus:outline-none focus:ring-2 focus:ring-navy-500 focus:ring-offset-2 rounded"
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
                <h3 className="text-sm font-medium text-amber-800">Validation Error</h3>
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

        {/* Grant validity control (only in edit mode) */}
        {isEditMode && (
          <div className="card p-6">
            <h3 className="text-lg font-semibold text-navy-900 mb-4">Grant Validity</h3>
            <GrantValidityControl value={validityState} onChange={handleValidityChange} />
          </div>
        )}

        {/* Services section */}
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-semibold text-navy-900">
              Services
              <span className="ml-2 text-sm font-normal text-slate-500">
                ({services.length})
              </span>
            </h2>
          </div>

          {services.length === 0 ? (
            <div className="card p-8 text-center">
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
              <h3 className="mt-4 text-lg font-medium text-navy-900">
                No services available
              </h3>
              <p className="mt-2 text-slate-600">
                This agent has no services configured yet.
              </p>
            </div>
          ) : isEditMode ? (
            <ServiceGrantList
              services={services}
              grants={grants}
              onGrantsChange={handleGrantsChange}
              isEditable={true}
            />
          ) : (
            <div className="space-y-4">
              {services.map((service) => (
                <ServiceCard
                  key={service.serviceId}
                  service={service}
                  grants={getDelegatedTokensForService(service.serviceId)}
                />
              ))}
            </div>
          )}
        </div>

        {/* Action buttons (only in edit mode) */}
        {isEditMode && (
          <div className="flex items-center justify-end gap-3 pt-4">
            <Button variant="outline" onClick={handleCancel} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleSubmit}
              isLoading={isSubmitting}
              disabled={!hasChanges || isSubmitting}
            >
              Approve & Delegate
            </Button>
          </div>
        )}
      </div>
      </PageTransition>
    </AppLayout>
  );
}

export default AgentGrantDetailPage;
