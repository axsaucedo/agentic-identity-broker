/**
 * Button component - Reusable button with variants and loading states.
 *
 * Features:
 * - Multiple variants (primary, secondary, outline)
 * - Multiple sizes (sm, md, lg)
 * - Loading state with spinner
 * - Disabled state
 * - Keyboard accessible
 */

import React from 'react';

interface ButtonProps {
  /** Button variant */
  variant?: 'primary' | 'secondary' | 'outline';
  /** Button size */
  size?: 'sm' | 'md' | 'lg';
  /** Whether button is in loading state */
  isLoading?: boolean;
  /** Whether button is disabled */
  disabled?: boolean;
  /** Click handler */
  onClick?: () => void;
  /** Button type */
  type?: 'button' | 'submit' | 'reset';
  /** Button children */
  children: React.ReactNode;
  /** Additional CSS classes */
  className?: string;
}

/**
 * Button component with consistent styling and behavior.
 * Shows spinner when loading and disables interactions.
 */
export function Button({
  variant = 'primary',
  size = 'md',
  isLoading = false,
  disabled = false,
  onClick,
  type = 'button',
  children,
  className = '',
}: ButtonProps) {
  const isDisabled = disabled || isLoading;

  // Variant styles - Premium navy-based color scheme
  const variantClasses = {
    primary: `
      text-white border-transparent
      shadow-md hover:shadow-lg transition-shadow font-semibold
    `,
    secondary: `
      bg-emerald-600 text-white border-transparent
      hover:bg-emerald-700
      focus:ring-emerald-500
      disabled:bg-emerald-400
      shadow-md hover:shadow-lg transition-shadow
    `,
    outline: `
      bg-white text-navy-900 border-slate
      hover:bg-cream hover:border-taupe
      focus:ring-navy-500
      disabled:bg-sand disabled:text-slate-400 disabled:border-taupe
      shadow-sm hover:shadow-card transition-all
    `,
  };

  // Size styles
  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg',
  };

  const baseClasses = `
    inline-flex items-center justify-center
    font-medium rounded-md border
    transition-colors duration-200
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:cursor-not-allowed disabled:opacity-60
  `;

  // For primary button, use inline style for guaranteed dark navy
  const buttonStyle = variant === 'primary'
    ? {
        backgroundColor: '#0d1829',
        color: '#ffffff'
      }
    : undefined;

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={isDisabled}
      style={buttonStyle}
      className={`${baseClasses} ${variant !== 'primary' ? variantClasses[variant] : ''} ${sizeClasses[size]} ${className}`}
    >
      {isLoading && (
        <svg
          className="animate-spin -ml-1 mr-2 h-4 w-4"
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          aria-hidden="true"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      )}
      {children}
    </button>
  );
}

export default Button;
