/**
 * GrantValidityControl component - Controls grant expiration settings.
 *
 * Features:
 * - Checkbox to enable/disable expiration
 * - DatePicker for selecting expiration date
 * - Suggested date shortcuts (1 month, 3 months, 1 year)
 * - Current date display
 * - Validation
 */

import React, { useMemo } from 'react';
import { addMonths, addYears, format } from 'date-fns';
import { DatePicker } from '../ui/DatePicker';
import type { GrantValidityState } from '../../types/consent';

interface GrantValidityControlProps {
  /** Current validity state */
  value: GrantValidityState;
  /** Callback when validity state changes */
  onChange: (state: GrantValidityState) => void;
  /** Additional CSS classes */
  className?: string;
}

/**
 * GrantValidityControl allows users to set grant expiration date.
 * Provides checkbox to enable expiration and date picker with suggested dates.
 */
export function GrantValidityControl({
  value,
  onChange,
  className = '',
}: GrantValidityControlProps) {
  const today = useMemo(() => new Date(), []);
  const tomorrow = useMemo(() => {
    const date = new Date();
    date.setDate(date.getDate() + 1);
    return date;
  }, []);

  // Suggested dates
  const suggestedDates = useMemo(
    () => [
      { label: '1 month', date: addMonths(today, 1) },
      { label: '3 months', date: addMonths(today, 3) },
      { label: '1 year', date: addYears(today, 1) },
    ],
    [today]
  );

  // Handle checkbox toggle
  const handleCheckboxChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const checked = e.target.checked;
    onChange({
      noExpiration: !checked,
      expiresAt: checked ? addMonths(today, 3) : undefined,
    });
  };

  // Handle date change
  const handleDateChange = (date: Date | null) => {
    onChange({
      noExpiration: false,
      expiresAt: date || undefined,
    });
  };

  // Handle suggested date click
  const handleSuggestedDateClick = (date: Date) => {
    onChange({
      noExpiration: false,
      expiresAt: date,
    });
  };

  const hasExpiration = !value.noExpiration;

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Current date info */}
      <div className="text-sm text-gray-600">
        Today: <span className="font-medium">{format(today, 'MMMM d, yyyy')}</span>
      </div>

      {/* Expiration checkbox */}
      <div className="flex items-start gap-3">
        <input
          type="checkbox"
          id="grant-expires"
          checked={hasExpiration}
          onChange={handleCheckboxChange}
          className="mt-1 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
        />
        <label htmlFor="grant-expires" className="text-sm font-medium text-gray-700">
          Grant expires on a specific date
        </label>
      </div>

      {/* Date picker (only shown when checkbox is checked) */}
      {hasExpiration && (
        <div className="pl-7 space-y-3">
          <DatePicker
            value={value.expiresAt || null}
            onChange={handleDateChange}
            minDate={tomorrow}
            label="Expiration date"
            error={
              value.expiresAt && value.expiresAt <= today
                ? 'Expiration date must be in the future'
                : undefined
            }
          />

          {/* Suggested dates */}
          <div>
            <p className="text-xs text-gray-600 mb-2">Quick suggestions:</p>
            <div className="flex flex-wrap gap-2">
              {suggestedDates.map((suggestion) => (
                <button
                  key={suggestion.label}
                  type="button"
                  onClick={() => handleSuggestedDateClick(suggestion.date)}
                  className="px-3 py-1.5 text-xs font-medium bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                >
                  {suggestion.label}
                  <span className="ml-1.5 text-gray-500">
                    ({format(suggestion.date, 'MMM d, yyyy')})
                  </span>
                </button>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* No expiration message */}
      {!hasExpiration && (
        <div className="pl-7 text-sm text-gray-600">
          This grant will remain active indefinitely until manually revoked.
        </div>
      )}
    </div>
  );
}

export default GrantValidityControl;
