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
 */

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import type { ThirdpartyService, DelegatedToken } from '../../types/consent';
import { ScopeList } from './ScopeList';
import { GrantStatusBadge } from './GrantStatusBadge';
import { Switch } from '../ui/Switch';

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
    <div className="card overflow-hidden">
      {/* Service header */}
      <div className={`p-6 ${isExpanded ? 'border-b border-dotted border-slate/10' : ''}`}>
        <div className="flex items-start gap-4">
          {/* Service logo */}
          <div className="flex-shrink-0">
            {service.logoUrl ? (
              <img
                src={service.logoUrl}
                alt={`${service.displayName} logo`}
                className="w-12 h-12 rounded-lg object-cover"
              />
            ) : (
              <div className="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white text-lg font-semibold">
                  {service.displayName.charAt(0).toUpperCase()}
                </span>
              </div>
            )}
          </div>

          {/* Service info */}
          <div className="flex-1 min-w-0">
            <div className="flex items-start justify-between gap-4">
              <div className="flex-1 min-w-0">
                <h3 className="text-lg font-semibold text-navy-900">
                  {service.displayName}
                </h3>
                {!isEditable && hasGrant && grantStatus && (
                  <div className="mt-2">
                    <GrantStatusBadge status={grantStatus} expiresAt={null} />
                  </div>
                )}
              </div>

              {/* Interactive toggle (editable mode only) */}
              {isEditable && (
                <Switch
                  checked={isServiceEnabled}
                  onChange={handleServiceToggle}
                  label={isServiceEnabled ? 'Enabled' : 'Disabled'}
                />
              )}
            </div>

            {/* Grant summary */}
            {!isEditable && hasGrant && (
              <p className="mt-2 text-sm text-slate-600">
                {grantedScopes.length === 1
                  ? '1 scope granted'
                  : `${grantedScopes.length} scopes granted`}
              </p>
            )}
            {isEditable && isServiceEnabled && localSelectedScopes.size > 0 && (
              <p className="mt-2 text-sm text-slate-600">
                {localSelectedScopes.size === 1
                  ? '1 scope selected'
                  : `${localSelectedScopes.size} scopes selected`}
              </p>
            )}
          </div>
        </div>
      </div>

      {/* Expandable scope section */}
      {(!isEditable || isServiceEnabled) && (
        <div>
          <button
            type="button"
            onClick={() => setIsExpanded(!isExpanded)}
            className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-slate/5 transition-colors focus:outline-none focus:ring-2 focus:ring-inset focus:ring-navy-500"
            aria-expanded={isExpanded}
            aria-controls={`scopes-${service.serviceId}`}
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
                    <div className="flex gap-2 mb-3">
                      <button
                        type="button"
                        onClick={handleSelectAll}
                        className="text-xs font-medium text-emerald-600 hover:text-emerald-700 focus:outline-none focus:underline"
                      >
                        Select all
                      </button>
                      <span className="text-xs text-slate-400">|</span>
                      <button
                        type="button"
                        onClick={handleDeselectAll}
                        className="text-xs font-medium text-emerald-600 hover:text-emerald-700 focus:outline-none focus:underline"
                      >
                        Deselect all
                      </button>
                    </div>
                  )}

                  {/* Scope list */}
                  {isEditable ? (
                    <div className="divide-y divide-dotted divide-slate/10">
                      {service.scopes.map((scope) => {
                        const isChecked = localSelectedScopes.has(scope.value);

                        return (
                          <div
                            key={scope.value}
                            className="flex items-start gap-3 py-3 first:pt-0 last:pb-0"
                          >
                            {/* Checkbox */}
                            <input
                              type="checkbox"
                              id={`scope-${service.serviceId}-${scope.value}`}
                              checked={isChecked}
                              onChange={(e) => handleScopeToggle(scope.value, e.target.checked)}
                              className="mt-1 h-4 w-4 rounded border-slate-300 text-emerald-600 focus:ring-emerald-500"
                            />

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
                    </div>
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
        </div>
      )}

      {/* View-only notice (only in non-editable mode with grants) */}
      {!isEditable && hasGrant && (
        <div className="px-6 py-3 bg-emerald-50 text-sm text-emerald-700">
          <div className="flex items-center gap-2">
            <svg
              className="w-4 h-4 flex-shrink-0"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>Enable edit mode to modify this grant</span>
          </div>
        </div>
      )}
    </div>
  );
}

export default ServiceCard;
