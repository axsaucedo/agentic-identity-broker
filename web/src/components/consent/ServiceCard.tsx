/**
 * ServiceCard component displays a third-party service with its scopes and grant status.
 *
 * Features:
 * - Service logo and name
 * - Description/purpose
 * - Expandable/collapsible scope list
 * - Shows which scopes user has granted (if any)
 * - Grant status badge
 * - Interactive mode with toggles and checkboxes (Phase 5)
 * - View-only mode (Phase 4)
 *
 * Refactored to use design system primitives (Phase 9).
 */

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { ThirdpartyService, DelegatedToken } from '../../types/consent';
import { Card } from '@design-system/components/data-display/Card';
import { Stack } from '@design-system/components/layout/Stack';
import { Button } from '@design-system/components/primitives/Button';
import { Avatar } from '@design-system/components/primitives/Avatar';
import { Switch } from '@design-system/components/inputs/Switch';
import { StatusIndicator } from '@design-system/components/data-display/StatusIndicator';
import { Alert } from '@design-system/components/feedback/Alert';
import { ScopeList } from './ScopeList';

interface ServiceCardProps {
  /** Third-party service information */
  service: ThirdpartyService;
  /** Existing grants for this service (if any) */
  grants?: DelegatedToken[];
  /** Loading state */
  isLoading?: boolean;
  /** Whether the card is in editable mode */
  isEditable?: boolean;
  /** Callback when service is toggled on/off */
  onServiceToggle?: (serviceId: string, enabled: boolean) => void;
  /** Callback when scope selection changes */
  onScopeChange?: (serviceId: string, scopes: string[]) => void;
  /** Pre-selected state for interactive mode */
  isServiceEnabled?: boolean;
  /** Pre-selected scopes for interactive mode */
  selectedScopes?: string[];
}

/**
 * ServiceCard displays service details with expandable scope list.
 * Supports both view-only and interactive editing modes.
 */
export function ServiceCard({
  service,
  grants,
  isEditable = false,
  onServiceToggle,
  onScopeChange,
  isServiceEnabled = false,
  selectedScopes = [],
}: ServiceCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);

  // Use selected scopes directly from props (controlled component)
  const localSelectedScopes = new Set(selectedScopes);

  // Find the grant for this specific service
  const serviceGrant = grants?.find((g) => g.thirdparty_oauth2_service_id === service.serviceId);
  const hasGrant = !!serviceGrant;
  const grantedScopes = serviceGrant?.scopes || [];

  // Determine grant status
  const getGrantStatus = (): 'active' | 'expired' | 'pending' | null => {
    if (!hasGrant) return null;
    // For Phase 5, we only show active grants (no expiration checking yet)
    return 'active';
  };

  const grantStatus = getGrantStatus();

  // Handle service toggle
  const handleServiceToggle = (enabled: boolean) => {
    if (onServiceToggle) {
      onServiceToggle(service.serviceId, enabled);
    }
    // Auto-expand when enabling service
    if (enabled) {
      setIsExpanded(true);
    }
  };

  // Handle scope checkbox change
  const handleScopeToggle = (scopeValue: string, checked: boolean) => {
    const newScopes = new Set(localSelectedScopes);
    if (checked) {
      newScopes.add(scopeValue);
    } else {
      newScopes.delete(scopeValue);
    }

    if (onScopeChange) {
      onScopeChange(service.serviceId, Array.from(newScopes));
    }
  };

  // Select all scopes
  const handleSelectAll = () => {
    const allScopes = service.scopes.map((s) => s.value);
    if (onScopeChange) {
      onScopeChange(service.serviceId, allScopes);
    }
  };

  // Deselect all scopes
  const handleDeselectAll = () => {
    if (onScopeChange) {
      onScopeChange(service.serviceId, []);
    }
  };

  return (
    <Card padding="none" hover="none" border="subtle">
      {/* Service header */}
      <div className="p-6">
        <Stack direction="row" gap="md" align="start">
          {/* Service logo using Avatar component */}
          <Avatar
            src={service.logoUrl}
            alt={`${service.displayName} logo`}
            initials={service.displayName.charAt(0).toUpperCase()}
            size="lg"
            shape="rounded"
          />

          {/* Service info */}
          <Stack gap="sm" className="flex-1 min-w-0">
            <Stack direction="row" justify="space-between" align="start" gap="md">
              <Stack gap="xs" className="flex-1 min-w-0">
                <h3 className="text-lg font-semibold text-navy-900">
                  {service.displayName}
                </h3>
                {/* Status and scopes indicators */}
                {!isEditable && hasGrant && grantStatus && (
                  <Stack direction="row" gap="md" align="center" className="flex-wrap">
                    <StatusIndicator
                      label={grantStatus.charAt(0).toUpperCase() + grantStatus.slice(1)}
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
                    <StatusIndicator
                      label={
                        grantedScopes.length === 1
                          ? '1 scope granted'
                          : `${grantedScopes.length} scopes granted`
                      }
                    />
                  </Stack>
                )}
              </Stack>

              {/* Interactive toggle (editable mode only) */}
              {isEditable && (
                <Switch
                  checked={isServiceEnabled}
                  onChange={handleServiceToggle}
                  label={isServiceEnabled ? 'Enabled' : 'Disabled'}
                  size="md"
                />
              )}
            </Stack>

            {/* Grant summary (editable mode) */}
            {isEditable && isServiceEnabled && localSelectedScopes.size > 0 && (
              <StatusIndicator
                label={
                  localSelectedScopes.size === 1
                    ? '1 scope selected'
                    : `${localSelectedScopes.size} scopes selected`
                }
                variant="info"
              />
            )}
          </Stack>
        </Stack>
      </div>

      {/* Expandable scope section */}
      {(!isEditable || isServiceEnabled) && (
        <>
          {/* Divider between header and scope section */}
          {isExpanded && <div className="border-t border-dotted border-slate/10" />}

          <button
            type="button"
            onClick={() => setIsExpanded(!isExpanded)}
            className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-slate/5 transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-navy-500"
            aria-expanded={isExpanded}
            aria-controls={`scopes-${service.serviceId}`}
            aria-label={
              isEditable
                ? `Select Scopes (${service.scopes.length})`
                : `Available Scopes (${service.scopes.length})`
            }
          >
            {isExpanded && (
              <span className="text-sm font-medium text-slate-700">
                {isEditable ? 'Select Scopes' : 'Available Scopes'} ({service.scopes.length})
              </span>
            )}
            <motion.svg
              className="w-5 h-5 text-slate-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
              animate={{ rotate: isExpanded ? 180 : 0 }}
              transition={{ duration: 0.2 }}
              aria-hidden="true"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M19 9l-7 7-7-7"
              />
            </motion.svg>
          </button>

          <AnimatePresence initial={false}>
            {isExpanded && (
              <motion.div
                id={`scopes-${service.serviceId}`}
                initial={{ height: 0, opacity: 0 }}
                animate={{ height: 'auto', opacity: 1 }}
                exit={{ height: 0, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="overflow-hidden"
              >
                <div className="px-6 pb-4">
                  {/* Select all/none buttons (editable mode only) */}
                  {isEditable && (
                    <Stack direction="row" gap="xs" className="mb-3" align="center">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleSelectAll}
                        className="text-xs font-medium text-emerald-600 hover:text-emerald-700 h-auto py-0 px-0"
                      >
                        Select all
                      </Button>
                      <span className="text-xs text-slate-400">|</span>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={handleDeselectAll}
                        className="text-xs font-medium text-emerald-600 hover:text-emerald-700 h-auto py-0 px-0"
                      >
                        Deselect all
                      </Button>
                    </Stack>
                  )}

                  {/* Scope list */}
                  {isEditable ? (
                    <Stack gap="xs" className="divide-y divide-dotted divide-slate/10">
                      {service.scopes.map((scope) => {
                        const isChecked = localSelectedScopes.has(scope.value);

                        return (
                          <div
                            key={scope.value}
                            className="flex items-start gap-3 py-3 first:pt-0 last:pb-0"
                          >
                            {/* Checkbox using design system */}
                            <div className="flex-shrink-0 pt-0.5">
                              <input
                                type="checkbox"
                                id={`scope-${service.serviceId}-${scope.value}`}
                                checked={isChecked}
                                onChange={(e) => handleScopeToggle(scope.value, e.target.checked)}
                                className="h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                              />
                            </div>

                            {/* Scope details */}
                            <div className="flex-1 min-w-0">
                              <label
                                htmlFor={`scope-${service.serviceId}-${scope.value}`}
                                className="cursor-pointer"
                              >
                                <code
                                  className={`text-sm font-mono font-medium break-all ${
                                    isChecked ? 'text-emerald-700' : 'text-navy-900'
                                  }`}
                                >
                                  {scope.value}
                                </code>
                                <p
                                  className={`mt-1 text-sm ${
                                    isChecked ? 'text-emerald-600' : 'text-slate-600'
                                  }`}
                                >
                                  {scope.description}
                                </p>
                              </label>
                            </div>
                          </div>
                        );
                      })}
                    </Stack>
                  ) : (
                    <ScopeList
                      scopes={service.scopes}
                      grantedScopes={grantedScopes}
                      isExpanded={isExpanded}
                    />
                  )}
                </div>
              </motion.div>
            )}
          </AnimatePresence>
        </>
      )}

      {/* View-only notice (only in non-editable mode with grants) */}
      {!isEditable && hasGrant && (
        <div className="px-6 pb-3">
          <Alert variant="info">
            Enable edit mode to modify this grant
          </Alert>
        </div>
      )}
    </Card>
  );
}

export default ServiceCard;
