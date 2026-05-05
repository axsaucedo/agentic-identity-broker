interface CIMDConsentSummaryProps {
  domain: string;
  agentName: string;
  accessTarget: string;
  agentLogoUrl?: string | null;
}

export function CIMDConsentSummary({ domain, agentName, accessTarget, agentLogoUrl }: CIMDConsentSummaryProps) {
  return (
    <div className="flex items-start gap-4 p-4 bg-neutral-50 rounded-lg border border-neutral-200">
      {agentLogoUrl && (
        <img
          src={agentLogoUrl}
          alt={`${agentName} logo`}
          className="w-12 h-12 rounded-lg object-contain flex-shrink-0"
          referrerPolicy="no-referrer"
        />
      )}
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-1.5">
          <svg className="w-4 h-4 text-success-primary flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
          <span className="text-xs text-neutral-500">Verified domain:</span>
          <span className="font-mono font-semibold text-trust-deep">{domain}</span>
        </div>
        <p className="text-neutral-600 text-sm">
          <span>{agentName}</span>{' '}
          wants to access{' '}
          <span className="font-medium text-neutral-900">{accessTarget}</span>.
        </p>
      </div>
    </div>
  );
}
