/**
 * Switch component - Toggle switch control using Headless UI.
 *
 * Features:
 * - Accessible with ARIA labels
 * - Smooth transition animations
 * - Blue when active, gray when inactive
 * - Disabled state support
 * - Keyboard accessible
 */

import React from 'react';
import { Switch as HeadlessSwitch } from '@headlessui/react';

interface SwitchProps {
  /** Whether the switch is checked */
  checked: boolean;
  /** Callback when switch state changes */
  onChange: (checked: boolean) => void;
  /** Optional label text */
  label?: string;
  /** Whether the switch is disabled */
  disabled?: boolean;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Switch component wraps Headless UI Switch with consistent styling.
 * Use for binary on/off controls.
 */
export function Switch({
  checked,
  onChange,
  label,
  disabled = false,
  className = '',
}: SwitchProps) {
  return (
    <HeadlessSwitch.Group>
      <div className={`flex items-center gap-3 ${className}`}>
        <HeadlessSwitch
          checked={checked}
          onChange={onChange}
          disabled={disabled}
          className={`
            relative inline-flex h-6 w-11 flex-shrink-0 items-center rounded-full
            transition-all duration-200 ease-in-out
            focus:outline-none focus:ring-2 focus:ring-navy-500 focus:ring-offset-2
            shadow-sm hover:shadow-md
            ${checked ? 'bg-navy-600' : 'bg-slate-300'}
            ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
          `}
        >
          <span
            className={`
              inline-block h-5 w-5 transform rounded-full bg-white shadow-sm
              transition-transform duration-200 ease-in-out
              ${checked ? 'translate-x-5' : 'translate-x-0.5'}
            `}
          />
        </HeadlessSwitch>
        {label && (
          <HeadlessSwitch.Label
            className={`text-sm font-medium whitespace-nowrap ${
              disabled ? 'text-slate-500 cursor-not-allowed' : 'text-navy-900 cursor-pointer'
            }`}
          >
            {label}
          </HeadlessSwitch.Label>
        )}
      </div>
    </HeadlessSwitch.Group>
  );
}

export default Switch;
