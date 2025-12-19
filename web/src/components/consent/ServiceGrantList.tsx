/**
 * ServiceGrantList component - Manages multiple service grants interactively.
 *
 * Features:
 * - Renders ServiceCard components in interactive mode
 * - Manages state of which services/scopes are selected
 * - Returns selected scopes as grant request format
 * - Supports both editable and view-only modes
 */

import React, { useState, useEffect, useCallback } from 'react';
import type { ThirdpartyService, UserGrant, DelegatedToken } from '../../types/consent';
import { ServiceCard } from './ServiceCard';

interface ServiceGrantListProps {
  /** Available third-party services */
  services: ThirdpartyService[];
  /** Existing user grants */
  grants: UserGrant[];
  /** Callback when grants change */
  onGrantsChange: (delegatedTokens: DelegatedToken[]) => void;
  /** Whether the list is in editable mode */
  isEditable: boolean;
}

interface ServiceState {
  isEnabled: boolean;
  selectedScopes: string[];
}

/**
 * ServiceGrantList manages interactive service grant selection.
 * Aggregates service/scope selections and provides grant request data.
 */
export function ServiceGrantList({
  services,
  grants,
  onGrantsChange,
  isEditable,
}: ServiceGrantListProps) {
  // Track state for each service
  const [serviceStates, setServiceStates] = useState<Map<string, ServiceState>>(new Map());

  // Initialize service states from existing grants
  useEffect(() => {
    const initialStates = new Map<string, ServiceState>();

    // Get all delegated tokens from grants (filter out undefined/null)
    const allDelegatedTokens = grants
      .flatMap((grant) => grant.delegated_oauth2_tokens || [])
      .filter((token) => token != null);

    // Initialize state for each service
    services.forEach((service) => {
      const existingToken = allDelegatedTokens.find((t) => t.thirdparty_oauth2_service_id === service.serviceId);

      initialStates.set(service.serviceId, {
        isEnabled: !!existingToken,
        selectedScopes: existingToken?.scopes || [],
      });
    });

    setServiceStates(initialStates);
  }, [services, grants]);

  // Handle service toggle
  const handleServiceToggle = useCallback(
    (serviceId: string, enabled: boolean) => {
      setServiceStates((prev) => {
        const newStates = new Map(prev);
        const currentState = newStates.get(serviceId) || { isEnabled: false, selectedScopes: [] };

        newStates.set(serviceId, {
          ...currentState,
          isEnabled: enabled,
          // Clear scopes when disabling
          selectedScopes: enabled ? currentState.selectedScopes : [],
        });

        return newStates;
      });
    },
    []
  );

  // Handle scope selection change
  const handleScopeChange = useCallback((serviceId: string, scopes: string[]) => {
    setServiceStates((prev) => {
      const newStates = new Map(prev);
      const currentState = newStates.get(serviceId) || { isEnabled: false, selectedScopes: [] };

      newStates.set(serviceId, {
        ...currentState,
        selectedScopes: scopes,
      });

      return newStates;
    });
  }, []);

  // Notify parent of changes when service states change
  useEffect(() => {
    const delegatedTokens: DelegatedToken[] = [];

    serviceStates.forEach((state, serviceId) => {
      // Only include enabled services with at least one scope
      if (state.isEnabled && state.selectedScopes.length > 0) {
        delegatedTokens.push({
          thirdparty_oauth2_service_id: serviceId,
          scopes: state.selectedScopes,
        });
      }
    });

    onGrantsChange(delegatedTokens);
  }, [serviceStates, onGrantsChange]);

  // Get existing grants for a service
  const getGrantsForService = (serviceId: string): DelegatedToken[] => {
    return grants
      .flatMap((grant) => grant.delegated_oauth2_tokens || [])
      .filter((token) => token != null && token.thirdparty_oauth2_service_id === serviceId);
  };

  return (
    <div className="space-y-4">
      {services.map((service) => {
        const serviceState = serviceStates.get(service.serviceId) || {
          isEnabled: false,
          selectedScopes: [],
        };

        return (
          <ServiceCard
            key={service.serviceId}
            service={service}
            grants={getGrantsForService(service.serviceId)}
            isEditable={isEditable}
            onServiceToggle={handleServiceToggle}
            onScopeChange={handleScopeChange}
            isServiceEnabled={serviceState.isEnabled}
            selectedScopes={serviceState.selectedScopes}
          />
        );
      })}
    </div>
  );
}

export default ServiceGrantList;
