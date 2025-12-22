/**
 * GlobalErrorBoundary - Enhanced error boundary for the entire application.
 *
 * Features:
 * - Catches all React rendering errors
 * - Shows user-friendly error page with fallback UI
 * - Displays error details in development mode only
 * - Provides "Reload page" and "Go home" actions
 * - Logs errors to console (ready for external error tracking)
 * - Error icon/illustration
 *
 * Usage:
 * ```tsx
 * import { GlobalErrorBoundary } from '@components/ui/GlobalErrorBoundary';
 *
 * function App() {
 *   return (
 *     <GlobalErrorBoundary>
 *       <YourApp />
 *     </GlobalErrorBoundary>
 *   );
 * }
 * ```
 */

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

/**
 * Global error boundary that catches all React errors.
 * Displays a comprehensive fallback UI with actionable options.
 */
export class GlobalErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    // Update state to trigger fallback UI
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    // Log error details to console
    console.error('GlobalErrorBoundary caught an error:', error, errorInfo);

    // Update state with error details
    this.setState({
      error,
      errorInfo,
    });

    // TODO: Send error to external tracking service
    // Example: Sentry, LogRocket, etc.
    // if (import.meta.env.PROD) {
    //   logErrorToService(error, errorInfo);
    // }
  }

  handleReset = (): void => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  handleReload = (): void => {
    window.location.reload();
  };

  handleGoHome = (): void => {
    window.location.href = '/consent';
  };

  render(): ReactNode {
    if (this.state.hasError) {
      // Custom fallback if provided
      if (this.props.fallback) {
        return this.props.fallback;
      }

      // Check if running in development mode
      const isDevelopment = import.meta.env.DEV;

      // Default comprehensive fallback UI
      return (
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-neutral-50 to-neutral-100 px-4 py-8">
          <div className="max-w-2xl w-full bg-white rounded-2xl shadow-2xl overflow-hidden border border-neutral-200">
            {/* Header Section with Icon */}
            <div className="bg-gradient-to-r from-red-500 to-red-600 px-8 py-10 text-center">
              <div className="inline-flex items-center justify-center w-20 h-20 bg-white rounded-full mb-4 shadow-lg">
                <svg
                  className="w-12 h-12 text-red-500"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  aria-hidden="true"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                  />
                </svg>
              </div>
              <h1 className="text-3xl font-bold text-white mb-2">
                Oops! Something went wrong
              </h1>
              <p className="text-red-100 text-lg">
                We encountered an unexpected error
              </p>
            </div>

            {/* Content Section */}
            <div className="px-8 py-8">
              <p className="text-neutral-700 text-lg mb-6 leading-relaxed">
                We're sorry for the inconvenience. An unexpected error occurred while
                rendering this page. Our team has been notified and we'll look into it.
              </p>

              {/* Error Details (Development Only) */}
              {isDevelopment && this.state.error && (
                <details className="mb-6 p-4 bg-neutral-50 rounded-lg border border-neutral-200">
                  <summary className="cursor-pointer text-sm font-semibold text-neutral-700 hover:text-neutral-900 mb-3 flex items-center gap-2">
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
                        d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                      />
                    </svg>
                    Error Details (Development Mode)
                  </summary>
                  <div className="space-y-3">
                    <div>
                      <p className="text-xs font-semibold text-neutral-600 mb-1">
                        Error Message:
                      </p>
                      <pre className="text-xs font-mono text-red-700 whitespace-pre-wrap break-words bg-red-50 p-3 rounded border border-red-200">
                        {this.state.error.toString()}
                      </pre>
                    </div>
                    {this.state.errorInfo && this.state.errorInfo.componentStack && (
                      <div>
                        <p className="text-xs font-semibold text-neutral-600 mb-1">
                          Component Stack:
                        </p>
                        <pre className="text-xs font-mono text-neutral-700 whitespace-pre-wrap break-words bg-neutral-100 p-3 rounded border border-neutral-300 max-h-64 overflow-y-auto">
                          {this.state.errorInfo.componentStack}
                        </pre>
                      </div>
                    )}
                  </div>
                </details>
              )}

              {/* Production Error Hint */}
              {!isDevelopment && (
                <div className="mb-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
                  <div className="flex gap-3">
                    <svg
                      className="w-5 h-5 text-blue-500 flex-shrink-0 mt-0.5"
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
                    <div>
                      <p className="text-sm font-medium text-blue-900 mb-1">
                        What can you do?
                      </p>
                      <ul className="text-sm text-blue-800 space-y-1 list-disc list-inside">
                        <li>Try reloading the page</li>
                        <li>Return to the home page and try again</li>
                        <li>Clear your browser cache</li>
                        <li>Contact support if the issue persists</li>
                      </ul>
                    </div>
                  </div>
                </div>
              )}

              {/* Action Buttons */}
              <div className="flex flex-col sm:flex-row gap-3">
                <button
                  onClick={this.handleReload}
                  className="flex-1 flex items-center justify-center gap-2 bg-red-600 text-white px-6 py-3 rounded-lg hover:bg-red-700 transition-colors font-semibold shadow-md hover:shadow-lg"
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
                      d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                    />
                  </svg>
                  Reload Page
                </button>
                <button
                  onClick={this.handleGoHome}
                  className="flex-1 flex items-center justify-center gap-2 bg-neutral-200 text-neutral-800 px-6 py-3 rounded-lg hover:bg-neutral-300 transition-colors font-semibold shadow-sm hover:shadow-md"
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
                      d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                    />
                  </svg>
                  Go Home
                </button>
              </div>

              {/* Additional Help Text */}
              <div className="mt-6 pt-6 border-t border-neutral-200">
                <p className="text-sm text-neutral-500 text-center">
                  If this problem persists, please contact our support team
                </p>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default GlobalErrorBoundary;
