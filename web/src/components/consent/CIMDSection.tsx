import type { CIMDMetadata, ThirdpartyService } from '../../types/consent';
import { CIMDConsentSummary } from './CIMDConsentSummary';
import { CIMDLocalhostWarning } from './CIMDLocalhostWarning';
import { CIMDAdvancedDetails } from './CIMDAdvancedDetails';

interface CIMDSectionProps {
  cimdMeta: CIMDMetadata;
  agentDisplayName: string;
  agentLogoUrl?: string;
  services: ThirdpartyService[];
}

export function CIMDSection({ cimdMeta, agentDisplayName, agentLogoUrl, services }: CIMDSectionProps) {
  const accessTarget =
    services
      .filter(s =>
        s.requiredScopes?.some(sc => cimdMeta.requested_scopes.includes(sc.name)) ||
        s.scopes?.some(sc => cimdMeta.requested_scopes.includes(sc.value))
      )
      .map(s => s.displayName || s.serviceName || s.serviceId)
      .join(', ') || cimdMeta.requested_scopes.join(', ') || 'requested services';

  return (
    <div className="space-y-3">
      <CIMDConsentSummary
        domain={cimdMeta.verified_domain}
        agentName={agentDisplayName}
        accessTarget={accessTarget}
        agentLogoUrl={agentLogoUrl}
      />
      {cimdMeta.is_localhost_redirect && (
        <CIMDLocalhostWarning agentDisplayName={agentDisplayName} />
      )}
      <CIMDAdvancedDetails
        clientIdUrl={cimdMeta.client_id_url}
        redirectUri={cimdMeta.redirect_uri}
        requestedScopes={cimdMeta.requested_scopes}
      />
    </div>
  );
}
