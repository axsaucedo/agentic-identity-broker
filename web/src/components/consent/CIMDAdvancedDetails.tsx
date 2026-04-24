import { useState } from 'react';

interface CIMDAdvancedDetailsProps {
  clientName: string;
  clientIdUrl: string;
  redirectUri: string;
  requestedScopes: string[];
}

export function CIMDAdvancedDetails({ clientName, clientIdUrl, redirectUri, requestedScopes }: CIMDAdvancedDetailsProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className="border border-neutral-200 rounded-lg overflow-hidden">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="flex w-full items-center justify-between px-4 py-3 text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors duration-150 focus:outline-none focus-visible:ring-2 focus-visible:ring-trust focus-visible:ring-offset-2"
        aria-expanded={expanded}
      >
        <span>Advanced Details</span>
        <svg
          className={`w-4 h-4 transition-transform duration-150 ${expanded ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {expanded && (
        <div className="px-4 pb-4 space-y-3 border-t border-neutral-200 pt-3">
          <DetailRow label="Client Name" value={clientName} />
          <DetailRow label="Client ID" value={clientIdUrl} mono />
          <DetailRow label="Redirect URI" value={redirectUri} mono />
          <div>
            <dt className="text-xs font-medium text-neutral-500 uppercase tracking-wide mb-1">Requested Scopes</dt>
            <dd className="flex flex-wrap gap-1.5">
              {requestedScopes.map((scope) => (
                <span key={scope} className="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono bg-neutral-100 text-neutral-700">
                  {scope}
                </span>
              ))}
              {requestedScopes.length === 0 && (
                <span className="text-sm text-neutral-500">None requested</span>
              )}
            </dd>
          </div>
        </div>
      )}
    </div>
  );
}

function DetailRow({ label, value, mono = false }: { label: string; value: string; mono?: boolean }) {
  return (
    <div>
      <dt className="text-xs font-medium text-neutral-500 uppercase tracking-wide mb-0.5">{label}</dt>
      <dd className={`text-sm text-neutral-800 break-all ${mono ? 'font-mono' : ''}`}>{value}</dd>
    </div>
  );
}
