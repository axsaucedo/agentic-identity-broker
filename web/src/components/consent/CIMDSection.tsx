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

function isLocalhostURI(uri: string): boolean {
  try {
    const host = new URL(uri).hostname;
    return host === 'localhost' || host === '127.0.0.1' || host === '::1';
  } catch {
    return false;
  }
}

export function CIMDSection({ cimdMeta, agentDisplayName, agentLogoUrl, services }: CIMDSectionProps) {
  const isLocalhostRedirect = isLocalhostURI(cimdMeta.redirect_uri);
  const accessTarget =
    services
      .filter(s =>
        s.kind === 'requirement'
          ? s.requiredScopes.some(sc => cimdMeta.requested_scopes.includes(sc.name))
          : (s.scopes?.some(sc => cimdMeta.requested_scopes.includes(sc.value)) ?? false)
      )
      .map(s => (s.kind === 'scoped' ? (s.displayName ?? s.serviceId) : (s.serviceName ?? s.serviceId)))
      .join(', ') || cimdMeta.requested_scopes.join(', ') || 'requested services';

  return (
    <div className="space-y-3">
      <CIMDConsentSummary
        domain={cimdMeta.verified_domain}
        agentName={agentDisplayName}
        accessTarget={accessTarget}
        agentLogoUrl={agentLogoUrl}
      />
      {isLocalhostRedirect && (
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
