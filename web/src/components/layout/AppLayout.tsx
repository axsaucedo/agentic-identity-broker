/**
 * AppLayout component provides consistent layout structure across the app.
 *
 * Includes:
 * - Header with navigation
 * - Main content area with responsive padding
 * - Footer
 */

import React, { ReactNode } from 'react';

interface AppLayoutProps {
  children: ReactNode;
}

/**
 * Header component with branding and navigation.
 * Premium design with gradient backdrop and refined typography.
 */
function Header() {
  return (
    <header className="sticky top-0 z-50 bg-white/80 backdrop-blur-md border-b border-taupe">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 gap-8">
          {/* Logo and branding */}
          <div className="flex items-center gap-3 flex-shrink-0">
            <div className="w-9 h-9 bg-gradient-to-br from-navy-600 to-navy-700 rounded-lg flex items-center justify-center shadow-md">
              <svg
                className="w-5 h-5 text-white"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"
                />
              </svg>
            </div>
            <div>
              <h1 className="font-display text-base font-bold text-navy-900 leading-tight">
                Consent Management
              </h1>
              <p className="text-xs text-slate-600 font-medium">
                Agentic Identity Broker
              </p>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex items-center gap-8 ml-auto">
            <a
              href="/consent"
              className="text-sm font-medium text-slate-700 hover:text-navy-700 transition-colors duration-200"
            >
              My Agents
            </a>
            <a
              href="/docs"
              className="text-sm font-medium text-slate-700 hover:text-navy-700 transition-colors duration-200"
              target="_blank"
              rel="noopener noreferrer"
            >
              Documentation
            </a>
          </nav>
        </div>
      </div>
    </header>
  );
}

/**
 * Footer component with links and copyright.
 * Refined minimal design with premium spacing.
 */
function Footer() {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="bg-white/50 backdrop-blur-sm border-t border-taupe mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex flex-col md:flex-row justify-between items-center gap-6">
          <div className="text-sm text-slate-600">
            <p>&copy; {currentYear} Agentic Identity Broker. All rights reserved.</p>
          </div>

          <div className="flex gap-8">
            <a
              href="/privacy"
              className="text-sm font-medium text-slate-600 hover:text-navy-700 transition-colors duration-200"
            >
              Privacy Policy
            </a>
            <a
              href="/terms"
              className="text-sm font-medium text-slate-600 hover:text-navy-700 transition-colors duration-200"
            >
              Terms of Service
            </a>
            <a
              href="/support"
              className="text-sm font-medium text-slate-600 hover:text-navy-700 transition-colors duration-200"
            >
              Support
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}

/**
 * Main layout wrapper that provides consistent structure.
 * Premium layout with refined spacing and background gradient.
 */
export function AppLayout({ children }: AppLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col bg-gradient-to-br from-cream via-sand to-taupe/20">
      <Header />

      <main className="flex-1 w-full">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="animate-fade-in-up">
            {children}
          </div>
        </div>
      </main>

      <Footer />
    </div>
  );
}

export default AppLayout;
