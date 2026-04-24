interface CIMDConsentSummaryProps {
  clientName: string;
  accessTarget: string;
  logoUri?: string;
}

export function CIMDConsentSummary({ clientName, accessTarget, logoUri }: CIMDConsentSummaryProps) {
  return (
    <div className="flex items-start gap-4 p-4 bg-neutral-50 rounded-lg border border-neutral-200">
      {logoUri && (
        <img
          src={logoUri}
          alt={`${clientName} logo`}
          className="w-12 h-12 rounded-lg object-contain flex-shrink-0"
        />
      )}
      <p className="text-neutral-900 text-base">
        The application{' '}
        <span className="font-semibold text-trust-deep">{clientName}</span>{' '}
        wants to access{' '}
        <span className="font-semibold">{accessTarget}</span>.
      </p>
    </div>
  );
}
