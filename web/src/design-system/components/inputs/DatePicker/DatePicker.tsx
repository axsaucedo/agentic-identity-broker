/**
 * DatePicker Component
 *
 * Date input with validation and semantic styling.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - HTML5 date input with native picker
 * - Min/max date validation
 * - Error/success states
 * - Helper text and error messages
 * - Label support with required indicator
 * - Full keyboard accessibility
 * - WCAG 2.1 AA compliant
 */

import React from 'react';
import { format, parse, isValid, isBefore, isAfter } from 'date-fns';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const dateInputVariants = cva(
  'w-full px-4 py-3 rounded-md border-1.5 transition-colors font-sans',
  {
    variants: {
      size: {
        sm: 'px-3 py-2 text-sm h-9',
        md: 'px-4 py-3 text-base h-11',
        lg: 'px-4 py-3.5 text-lg h-13',
      },
      variant: {
        default:
          'border-neutral-300 text-neutral-900 placeholder-neutral-500 focus:border-trust-deep focus:ring-1 focus:ring-trust',
        error:
          'border-error-primary bg-error-light/20 text-neutral-900 placeholder-neutral-500 focus:border-error-primary focus:ring-1 focus:ring-error-primary',
        success:
          'border-success-primary bg-success-light/20 text-neutral-900 placeholder-neutral-500 focus:border-success-primary focus:ring-1 focus:ring-success-primary',
      },
    },
    defaultVariants: {
      size: 'md',
      variant: 'default',
    },
  },
);

export interface DatePickerProps
  extends
    Omit<
      React.InputHTMLAttributes<HTMLInputElement>,
      'size' | 'type' | 'value' | 'onChange'
    >,
    VariantProps<typeof dateInputVariants> {
  /** Current date value */
  value: Date | null;
  /** Callback when date changes */
  onChange: (date: Date | null) => void;
  /** Minimum allowed date */
  minDate?: Date;
  /** Maximum allowed date */
  maxDate?: Date;
  /** Label text displayed above input */
  label?: string;
  /** Show required indicator on label */
  required?: boolean;
  /** Helper text displayed below input */
  helperText?: string;
  /** Error message (changes variant to error) */
  errorMessage?: string;
  /** Success message */
  successMessage?: string;
  /** Unique identifier for form association */
  id?: string;
}

/**
 * DatePicker component for date selection.
 * Wraps HTML5 date input with design system styling and validation.
 *
 * @example
 * ```tsx
 * <DatePicker
 *   label="Grant expiration"
 *   value={expiryDate}
 *   onChange={setExpiryDate}
 *   minDate={new Date()}
 * />
 *
 * <DatePicker
 *   label="Birth date"
 *   value={birthDate}
 *   onChange={setBirthDate}
 *   helperText="Required to verify age"
 * />
 * ```
 */
export const DatePicker = React.forwardRef<HTMLDivElement, DatePickerProps>(
  (
    {
      size = 'md',
      label,
      required = false,
      helperText,
      errorMessage,
      successMessage,
      value,
      onChange,
      minDate,
      maxDate,
      disabled = false,
      id,
      className,
      ...props
    },
    ref,
  ) => {
    // Determine variant based on error/success state
    const variant = errorMessage
      ? 'error'
      : successMessage
        ? 'success'
        : 'default';

    // Format date for input (YYYY-MM-DD)
    const formatForInput = (date: Date | null): string => {
      if (!date || !isValid(date)) return '';
      return format(date, 'yyyy-MM-dd');
    };

    // Parse input value to Date
    const parseInputValue = (inputValue: string): Date | null => {
      if (!inputValue) return null;
      const parsed = parse(inputValue, 'yyyy-MM-dd', new Date());
      if (!isValid(parsed)) return null;
      return parsed;
    };

    // Handle input change
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const inputValue = e.target.value;
      if (!inputValue) {
        onChange(null);
        return;
      }

      const parsed = parseInputValue(inputValue);
      if (!parsed) {
        // Invalid date format
        return;
      }

      // Validate against minDate
      if (minDate && isBefore(parsed, minDate)) {
        return;
      }

      // Validate against maxDate
      if (maxDate && isAfter(parsed, maxDate)) {
        return;
      }

      onChange(parsed);
    };

    const inputValue = formatForInput(value);
    const minDateValue = minDate ? formatForInput(minDate) : undefined;
    const maxDateValue = maxDate ? formatForInput(maxDate) : undefined;
    const generatedId =
      id || `date-picker-${Math.random().toString(36).substr(2, 9)}`;

    return (
      <div ref={ref} className="w-full">
        {/* Label */}
        {label && (
          <label
            htmlFor={generatedId}
            className="block text-sm font-medium text-neutral-900 mb-2"
          >
            {label}
            {required && <span className="ml-1 text-error-primary">*</span>}
          </label>
        )}

        {/* Input field */}
        <input
          id={generatedId}
          type="date"
          value={inputValue}
          onChange={handleChange}
          min={minDateValue}
          max={maxDateValue}
          disabled={disabled}
          className={cn(
            dateInputVariants({ size, variant }),
            disabled && 'bg-neutral-50 cursor-not-allowed opacity-60',
            className,
          )}
          aria-invalid={!!errorMessage}
          aria-describedby={
            errorMessage || successMessage || helperText
              ? `${generatedId}-description`
              : undefined
          }
          {...props}
        />

        {/* Helper text, error message, or success message */}
        <div className="mt-1.5">
          {errorMessage && (
            <p
              id={`${generatedId}-description`}
              className="text-xs text-error-primary font-medium"
              role="alert"
            >
              {errorMessage}
            </p>
          )}

          {successMessage && !errorMessage && (
            <p
              id={`${generatedId}-description`}
              className="text-xs text-success-primary font-medium"
            >
              {successMessage}
            </p>
          )}

          {helperText && !errorMessage && !successMessage && (
            <p
              id={`${generatedId}-description`}
              className="text-xs text-neutral-600"
            >
              {helperText}
            </p>
          )}
        </div>
      </div>
    );
  },
);

DatePicker.displayName = 'DatePicker';

export default DatePicker;
