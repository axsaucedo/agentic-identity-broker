/**
 * PersistenceSelector - Shared design-system-radio-backed persistence picker.
 *
 * Used by both the approval detail page and the Tool Authorizations list so the
 * persistence control behaves and looks the same in both places.
 */

import { Radio } from '@design-system/components/inputs/Radio';
import type { ApprovalPersistence } from '../../types/approval';

interface PersistenceSelectorProps {
  value: ApprovalPersistence;
  onChange: (value: ApprovalPersistence) => void;
  disabled?: boolean;
  name?: string;
}

interface PersistenceOption {
  value: ApprovalPersistence;
  label: string;
  description: string;
}

const OPTIONS: PersistenceOption[] = [
  {
    value: 'once',
    label: 'Just this once',
    description:
      'Allow this specific tool call only. The agent will need to request approval again next time.',
  },
  {
    value: 'session',
    label: 'For this session',
    description: 'Allow this tool for the duration of the current session.',
  },
  {
    value: 'permanent',
    label: 'Always allow',
    description:
      'Permanently allow this agent to use this tool without asking.',
  },
];

export function PersistenceSelector({
  value,
  onChange,
  disabled = false,
  name = 'approval-persistence',
}: PersistenceSelectorProps) {
  return (
    <fieldset className="space-y-3" disabled={disabled}>
      <legend className="mb-2 text-sm font-medium text-neutral-700">
        How long should this be allowed?
      </legend>
      <div className="grid gap-2 md:grid-cols-3">
        {OPTIONS.map((option) => {
          const checked = value === option.value;
          const optionId = `${name}-${option.value}`;

          return (
            <label
              key={option.value}
              htmlFor={optionId}
              className={[
                'block rounded-lg border p-3 transition-colors',
                checked
                  ? 'border-trust-deep bg-trust-light'
                  : 'border-neutral-200 bg-white hover:border-neutral-300',
                disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer',
              ].join(' ')}
            >
              <Radio
                id={optionId}
                name={name}
                value={option.value}
                checked={checked}
                disabled={disabled}
                onChange={() => onChange(option.value)}
                label={option.label}
                description={option.description}
              />
            </label>
          );
        })}
      </div>

      {value === 'permanent' && (
        <div
          className="rounded-lg border border-warning-primary/30 bg-warning-light p-3 text-sm text-warning-dark"
          role="alert"
        >
          <strong>Warning:</strong> This grants permanent access. You can revoke
          it any time from this page.
        </div>
      )}
    </fieldset>
  );
}

export default PersistenceSelector;
