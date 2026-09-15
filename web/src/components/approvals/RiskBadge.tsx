import type { RiskLevel } from '../../types/approval';

interface RiskBadgeProps {
  level?: RiskLevel | null;
}

type KnownRiskLevel = 'low' | 'medium' | 'critical';

const RISK_BADGES: Record<KnownRiskLevel, { label: string; className: string }> = {
  low: {
    label: 'Low Risk',
    className: 'border-success-primary/30 bg-success-light text-success-dark',
  },
  medium: {
    label: 'Medium Risk',
    className: 'border-warning-primary/30 bg-warning-light text-warning-dark',
  },
  critical: {
    label: 'Critical Risk',
    className: 'border-error-primary/30 bg-error-light text-error-dark',
  },
};

const UNKNOWN_RISK_BADGE_CLASSNAME =
  'border-warning-primary/30 bg-warning-light text-warning-dark';

export function RiskBadge({ level }: RiskBadgeProps) {
  if (!level) {
    return null;
  }

  const normalizedLevel = level.toLowerCase().trim();

  const config =
    normalizedLevel === 'low' ||
    normalizedLevel === 'medium' ||
    normalizedLevel === 'critical'
      ? RISK_BADGES[normalizedLevel]
      : {
          label: `${normalizedLevel
            .replace(/[_-]+/g, ' ')
            .replace(/\b\w/g, (char) => char.toUpperCase())} Risk`,
          className: UNKNOWN_RISK_BADGE_CLASSNAME,
        };

  return (
    <span
      className={`inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium ${config.className}`}
    >
      {config.label}
    </span>
  );
}

export default RiskBadge;
