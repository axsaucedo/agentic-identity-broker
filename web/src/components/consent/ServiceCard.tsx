/**
 * ServiceCard component displays a third-party service with its scopes and grant status.
 *
 * Features:
 * - Service logo and name
 * - Requirement type badge (Required/Optional)
 * - Connection status indicator
 * - Delegate/Revoke action button
 * - Scopes displayed as StatusIndicators with tooltip descriptions
 * - Expandable scope list for detailed view
 *
 * Refactored for Phase 8 (Simplified UI without edit mode).
 */

import type { ThirdpartyService, DelegatedToken } from '../../types/consent';
import { Card } from '@design-system/components/data-display/Card';
import { Stack } from '@design-system/components/layout/Stack';
import { Button } from '@design-system/components/primitives/Button';
import { Avatar } from '@design-system/components/primitives/Avatar';
import { StatusIndicator } from '@design-system/components/data-display/StatusIndicator';
import { Tooltip } from '@design-system/components/overlays/Tooltip';

interface ServiceCardProps {
  /** Third-party service information */
  service: ThirdpartyService;
  /** Existing grants for this service (if any) */
  grants?: DelegatedToken[];
  /** Loading state */
  isLoading?: boolean;
  /** Whether service is currently selected for delegation in this session */
  isDelegated?: boolean;
  /** Callback when Delegate button is clicked */
  onDelegate?: (serviceId: string) => void;
  /** Callback when Revoke button is clicked */
  onRevoke?: (serviceId: string) => void;
}

/**
 * ServiceCard displays service details with requirement type, connection status, and delegate action.
 * Scopes are shown as StatusIndicators with tooltip descriptions.
 * Card is highlighted when service is delegated (selected).
 */
export function ServiceCard({
  service,
  grants,
  isLoading = false,
  isDelegated = false,
  onDelegate,
  onRevoke,
}: ServiceCardProps) {

  // Get service display name
  // When service is a requirement, backend uses serviceName field
  const serviceDisplayName = service.displayName || service.serviceName || 'Unknown Service';

  // Normalize scopes to have consistent structure (value, description)
  // Support both ServiceScope (with 'value') and requiredScopes (with 'name')
  const normalizeScopes = (): Array<{ value: string; description?: string }> => {
    if (service.scopes) {
      return service.scopes;
    }
    if (service.requiredScopes) {
      return service.requiredScopes.map((s) => ({
        value: s.name,
        description: s.description,
      }));
    }
    return [];
  };

  const serviceScopes = normalizeScopes();

  // Find the grant for this specific service
  const serviceGrant = grants?.find((g) => g.thirdparty_oauth2_service_id === service.serviceId);
  const grantedScopes = serviceGrant?.scopes || [];

  // Determine connection status
  const isConnected = service.connectionStatus === 'connected';

  return (
    <article
      role="article"
      aria-label={`Service: ${serviceDisplayName}`}
      data-testid={`service-card-${service.serviceId}`}
    >
      <Card
        padding="none"
        hover="none"
        border="subtle"
        className={isDelegated ? 'ring-2 ring-success-primary bg-success-50' : ''}
      >
      {/* Service header */}
      <div className="p-6">
        <Stack direction="row" gap="md" align="start">
          {/* Avatar with status badge below */}
          <Stack gap="md" align="center">
            <Avatar
              src={service.logoUrl}
              alt={`${serviceDisplayName} logo`}
              initials={serviceDisplayName.charAt(0).toUpperCase()}
              size="lg"
              shape="rounded"
            />
            {service.requirementType && (
              <span
                className={
                  service.requirementType === 'mandatory'
                    ? 'px-2 py-1 text-xs font-semibold text-white bg-trust-deep rounded'
                    : 'px-2 py-1 text-xs font-semibold text-trust-deep border border-trust-deep rounded'
                }
              >
                {service.requirementType === 'mandatory' ? 'Required' : 'Optional'}
              </span>
            )}
          </Stack>

          {/* Service info */}
          <Stack gap="sm" className="flex-1 min-w-0">
            {/* Title */}
            <h3 className="text-lg font-semibold text-trust-deep">
              {serviceDisplayName}
            </h3>

            {/* Connection status */}
            {isConnected && (
              <StatusIndicator
                label="Active Session"
                variant="success"
                icon={
                  <svg
                    className="w-full h-full"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                    />
                  </svg>
                }
              />
            )}

            {/* Permissions (Scopes) as StatusIndicators with tooltips */}
            {serviceScopes.length > 0 && (
              <Stack gap="xs">
                <span className="text-xs font-semibold text-neutral-600 uppercase tracking-wider">
                  Permissions
                </span>
                <div className="flex flex-wrap gap-2">
                  {serviceScopes.map((scope) => (
                    <Tooltip
                      key={scope.value}
                      content={scope.description || `Scope: ${scope.value}`}
                    >
                      <StatusIndicator
                        label={scope.value}
                        variant={grantedScopes.includes(scope.value) ? 'success' : 'default'}
                        interactive
                      />
                    </Tooltip>
                  ))}
                </div>
              </Stack>
            )}
          </Stack>

          {/* Action button - Login / Delegate / Revoke */}
          <div className="flex-shrink-0" data-testid={`service-actions-${service.serviceId}`}>
            {!isConnected ? (
              <Button
                variant="primary"
                size="sm"
                onClick={() => onDelegate?.(service.serviceId)}
                disabled={isLoading}
                title="Connect to this service"
                className="bg-success-primary hover:bg-success-hover text-white"
                data-testid="service-login-button"
              >
                Login
              </Button>
            ) : isDelegated ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={() => onRevoke?.(service.serviceId)}
                disabled={isLoading}
                title="Revoke delegation for this service"
                data-testid="service-revoke-button"
              >
                Revoke
              </Button>
            ) : (
              <Button
                variant="primary"
                size="sm"
                onClick={() => onDelegate?.(service.serviceId)}
                disabled={isLoading}
                title="Delegate this service"
                data-testid="service-delegate-button"
              >
                Delegate
              </Button>
            )}
          </div>
        </Stack>
      </div>
      </Card>
    </article>
  );
}

export default ServiceCard;
