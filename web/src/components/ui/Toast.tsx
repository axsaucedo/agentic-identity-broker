/**
 * Toast notification component with auto-dismiss and animations.
 *
 * Features:
 * - Auto-dismiss after configurable duration (default 4s)
 * - Manual dismiss with close button
 * - Multiple variants: success, error, warning, info
 * - Framer Motion animations (fade + slide)
 * - Positioned at top-right corner
 * - Accessible with ARIA attributes
 *
 * Usage:
 * ```tsx
 * import { Toast, ToastProvider, useToast } from '@components/ui/Toast';
 *
 * function MyComponent() {
 *   const { showToast } = useToast();
 *   return (
 *     <button onClick={() => showToast('Success!', 'success')}>
 *       Show Toast
 *     </button>
 *   );
 * }
 * ```
 */

import React, {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
} from 'react';
import { motion, AnimatePresence } from 'framer-motion';

export type ToastVariant = 'success' | 'error' | 'warning' | 'info';

export interface ToastMessage {
  id: string;
  message: string;
  variant: ToastVariant;
  duration?: number;
}

interface ToastContextValue {
  showToast: (
    message: string,
    variant?: ToastVariant,
    duration?: number,
  ) => void;
  hideToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

/**
 * Hook to access toast notifications.
 * Must be used within a ToastProvider.
 */
export function useToast(): ToastContextValue {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
}

interface ToastProviderProps {
  children: React.ReactNode;
  maxToasts?: number;
}

/**
 * ToastProvider manages toast notification state and provides context.
 * Wrap your app or component tree with this provider.
 */
export function ToastProvider({ children, maxToasts = 5 }: ToastProviderProps) {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  /**
   * Show a new toast notification.
   */
  const showToast = useCallback(
    (
      message: string,
      variant: ToastVariant = 'info',
      duration: number = 4000,
    ) => {
      const id = `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
      const newToast: ToastMessage = { id, message, variant, duration };

      setToasts((prev) => {
        // Limit number of toasts
        const updated = [...prev, newToast];
        if (updated.length > maxToasts) {
          return updated.slice(updated.length - maxToasts);
        }
        return updated;
      });
    },
    [maxToasts],
  );

  /**
   * Hide a specific toast by ID.
   */
  const hideToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((toast) => toast.id !== id));
  }, []);

  return (
    <ToastContext.Provider value={{ showToast, hideToast }}>
      {children}
      <ToastContainer toasts={toasts} onDismiss={hideToast} />
    </ToastContext.Provider>
  );
}

interface ToastContainerProps {
  toasts: ToastMessage[];
  onDismiss: (id: string) => void;
}

/**
 * ToastContainer renders all active toasts in a fixed position.
 */
function ToastContainer({ toasts, onDismiss }: ToastContainerProps) {
  return (
    <div
      className="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none"
      role="region"
      aria-live="polite"
      aria-label="Notifications"
    >
      <AnimatePresence>
        {toasts.map((toast) => (
          <Toast key={toast.id} toast={toast} onDismiss={onDismiss} />
        ))}
      </AnimatePresence>
    </div>
  );
}

interface ToastProps {
  toast: ToastMessage;
  onDismiss: (id: string) => void;
}

/**
 * Individual Toast component with auto-dismiss timer.
 */
function Toast({ toast, onDismiss }: ToastProps) {
  const { id, message, variant, duration = 4000 } = toast;

  // Auto-dismiss after duration
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onDismiss(id);
      }, duration);

      return () => clearTimeout(timer);
    }
  }, [id, duration, onDismiss]);

  // Variant-specific styles
  const variantStyles = {
    success: {
      bg: 'bg-green-50',
      border: 'border-green-200',
      icon: 'text-green-500',
      text: 'text-green-900',
    },
    error: {
      bg: 'bg-red-50',
      border: 'border-red-200',
      icon: 'text-red-500',
      text: 'text-red-900',
    },
    warning: {
      bg: 'bg-yellow-50',
      border: 'border-yellow-200',
      icon: 'text-yellow-500',
      text: 'text-yellow-900',
    },
    info: {
      bg: 'bg-blue-50',
      border: 'border-blue-200',
      icon: 'text-blue-500',
      text: 'text-blue-900',
    },
  };

  const styles = variantStyles[variant];

  return (
    <motion.div
      initial={{ opacity: 0, x: 100, scale: 0.95 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 100, scale: 0.95 }}
      transition={{ duration: 0.2 }}
      className={`
        ${styles.bg} ${styles.border} ${styles.text}
        border rounded-lg shadow-lg p-4 pr-10
        min-w-[320px] max-w-md
        pointer-events-auto
        relative
      `}
      role={variant === 'error' || variant === 'warning' ? 'alert' : 'status'}
    >
      <div className="flex items-start gap-3">
        {/* Icon */}
        <div className={`${styles.icon} flex-shrink-0`}>
          {variant === 'success' && <SuccessIcon />}
          {variant === 'error' && <ErrorIcon />}
          {variant === 'warning' && <WarningIcon />}
          {variant === 'info' && <InfoIcon />}
        </div>

        {/* Message */}
        <p className="text-sm font-medium flex-1 pt-0.5">{message}</p>

        {/* Close button */}
        <button
          onClick={() => onDismiss(id)}
          className={`
            ${styles.icon} hover:opacity-70
            absolute top-3 right-3
            transition-opacity
            focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
            rounded
          `}
          aria-label="Close notification"
        >
          <svg
            className="w-5 h-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        </button>
      </div>
    </motion.div>
  );
}

// Icon components
function SuccessIcon() {
  return (
    <svg
      className="w-6 h-6"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
}

function ErrorIcon() {
  return (
    <svg
      className="w-6 h-6"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
}

function WarningIcon() {
  return (
    <svg
      className="w-6 h-6"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
      />
    </svg>
  );
}

function InfoIcon() {
  return (
    <svg
      className="w-6 h-6"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
      />
    </svg>
  );
}

export default Toast;
