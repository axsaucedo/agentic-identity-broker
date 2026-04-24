interface CIMDDomainBadgeProps {
  domain: string;
}

export function CIMDDomainBadge({ domain }: CIMDDomainBadgeProps) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <svg className="w-4 h-4 text-success-primary flex-shrink-0" fill="currentColor" viewBox="0 0 20 20" aria-hidden="true">
        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
      </svg>
      <span className="text-neutral-700">
        Verified domain: <span className="font-mono font-medium text-trust-deep">{domain}</span>
      </span>
    </div>
  );
}
