/**
 * TextArea Component
 *
 * Multi-line text input with validation and character count.
 * Follows the "Refined Trust Architecture" design system.
 *
 * Features:
 * - Flexible height (rows prop)
 * - Auto-grow support
 * - Character count
 * - Validation states (error, success)
 * - Helper text
 * - Full keyboard accessibility
 * - WCAG 2.1 AA compliant
 */

import React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '@design-system/utils';

const textareaVariants = cva(
  'w-full px-4 py-3 rounded-md border-1.5 transition-colors font-sans resize-none',
  {
    variants: {
      variant: {
        default: 'border-neutral-300 text-neutral-900 placeholder-neutral-500 focus:border-trust-deep focus:ring-1 focus:ring-trust',
        error: 'border-error-primary bg-error-light/20 text-neutral-900 placeholder-neutral-500 focus:border-error-primary focus:ring-1 focus:ring-error-primary',
        success: 'border-success-primary bg-success-light/20 text-neutral-900 placeholder-neutral-500 focus:border-success-primary focus:ring-1 focus:ring-success-primary',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export interface TextAreaProps
  extends Omit<React.TextareaHTMLAttributes<HTMLTextAreaElement>, 'rows'>,
    VariantProps<typeof textareaVariants> {
  /** Label text displayed above textarea */
  label?: string;
  /** Show required indicator on label */
  required?: boolean;
  /** Helper text displayed below textarea */
  helperText?: string;
  /** Error message (changes variant to error) */
  errorMessage?: string;
  /** Success message */
  successMessage?: string;
  /** Number of visible text lines */
  rows?: number;
  /** Auto-grow textarea as user types */
  autoGrow?: boolean;
  /** Show character count */
  showCharCount?: boolean;
  /** Unique identifier for form association */
  id?: string;
  /** Callback when field state changes */
  onChange?: (e: React.ChangeEvent<HTMLTextAreaElement>) => void;
}

/**
 * TextArea component for multi-line text entry.
 * Supports labels, validation states, character count, and optional auto-grow.
 *
 * @example
 * ```tsx
 * <TextArea
 *   label="Comments"
 *   placeholder="Enter your feedback..."
 *   helperText="Your feedback helps us improve"
 * />
 *
 * <TextArea
 *   label="Description"
 *   maxLength={500}
 *   showCharCount
 *   autoGrow
 * />
 * ```
 */
export const TextArea = React.forwardRef<HTMLTextAreaElement, TextAreaProps>(
  (
    {
      label,
      required = false,
      helperText,
      errorMessage,
      successMessage,
      showCharCount = false,
      maxLength,
      className,
      id,
      rows = 4,
      value = '',
      autoGrow = false,
      disabled = false,
      onChange,
      ...props
    },
    ref
  ) => {
    // Determine variant based on error/success state
    const variant = errorMessage ? 'error' : successMessage ? 'success' : 'default';

    // Character count
    const charCount = typeof value === 'string' ? value.length : 0;
    const displayCharCount = showCharCount && maxLength ? `${charCount}/${maxLength}` : null;

    // Handle auto-grow
    const textareaRef = React.useRef<HTMLTextAreaElement>(null);
    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onChange?.(e);

      if (autoGrow && textareaRef.current) {
        textareaRef.current.style.height = 'auto';
        textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
      }
    };

    // Initialize auto-grow height
    React.useEffect(() => {
      if (autoGrow && textareaRef.current) {
        textareaRef.current.style.height = 'auto';
        textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
      }
    }, [autoGrow]);

    return (
      <div className="w-full">
        {/* Label */}
        {label && (
          <label
            htmlFor={id}
            className="block text-sm font-medium text-neutral-900 mb-2"
          >
            {label}
            {required && <span className="ml-1 text-error-primary">*</span>}
          </label>
        )}

        {/* Textarea */}
        <textarea
          ref={(node) => {
            if (node) {
              (textareaRef as React.MutableRefObject<HTMLTextAreaElement>).current = node;
              if (typeof ref === 'function') ref(node);
              else if (ref) ref.current = node;
            }
          }}
          id={id}
          rows={autoGrow ? 1 : rows}
          value={value}
          maxLength={maxLength}
          disabled={disabled}
          onChange={handleChange}
          className={cn(
            textareaVariants({ variant }),
            disabled && 'bg-neutral-50 cursor-not-allowed opacity-60',
            autoGrow && 'min-h-[2.5rem] overflow-hidden',
            className
          )}
          {...props}
        />

        {/* Helper text, error message, or character count */}
        <div className="mt-1.5 flex items-center justify-between">
          {errorMessage && (
            <span className="text-xs text-error-primary font-medium">{errorMessage}</span>
          )}
          {successMessage && !errorMessage && (
            <span className="text-xs text-success-primary font-medium">{successMessage}</span>
          )}
          {helperText && !errorMessage && !successMessage && (
            <span className="text-xs text-neutral-600">{helperText}</span>
          )}

          {/* Character count on right */}
          {displayCharCount && (
            <span
              className={cn(
                'text-xs ml-auto',
                charCount > maxLength! * 0.8 ? 'text-warning-primary' : 'text-neutral-500'
              )}
            >
              {displayCharCount}
            </span>
          )}
        </div>
      </div>
    );
  }
);

TextArea.displayName = 'TextArea';

export default TextArea;
