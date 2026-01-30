/**
 * ServiceGrantList component - Manages multiple service grants interactively.
 *
 * Features:
 * - Renders ServiceCard components in interactive mode
 * - Manages state of which services/scopes are selected
 * - Returns selected scopes as grant request format
 * - Supports both editable and view-only modes
 */

import React, { useState, useEffect } from 'react';
import type {
  ThirdpartyService,
  UserGrant,
  DelegatedToken,
} from '../../types/consent';
import { ServiceCard } from './ServiceCard';

interface ServiceGrantListProps {
  /** Available third-party services */
  services: ThirdpartyService[];
  /** Existing user grants */
  grants: UserGrant[];
  /** Callback when grants change */
  onGrantsChange: (delegatedTokens: DelegatedToken[]) => void;
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
}: ServiceGrantListProps) {
  // Track state for each service
  const [serviceStates, setServiceStates] = useState<Map<string, ServiceState>>(
    new Map(),
  );

  // Initialize service states from existing grants
  useEffect(() => {
    const initialStates = new Map<string, ServiceState>();

    // Get all delegated tokens from grants (filter out undefined/null)
    const allDelegatedTokens = grants
      .flatMap((grant) => grant.delegated_oauth2_tokens || [])
      .filter((token) => token != null);

    // Initialize state for each service
    services.forEach((service) => {
      const existingToken = allDelegatedTokens.find(
        (t) => t.thirdparty_oauth2_service_id === service.serviceId,
      );

      initialStates.set(service.serviceId, {
        isEnabled: !!existingToken,
        selectedScopes: existingToken?.scopes || [],
      });
    });

    setServiceStates(initialStates);
  }, [services, grants]);

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
      .filter(
        (token) =>
          token != null && token.thirdparty_oauth2_service_id === serviceId,
      );
  };

  return (
    <div className="space-y-4">
      {services.map((service) => {
        return (
          <ServiceCard
            key={service.serviceId}
            service={service}
            grants={getGrantsForService(service.serviceId)}
          />
        );
      })}
    </div>
  );
}

export default ServiceGrantList;
