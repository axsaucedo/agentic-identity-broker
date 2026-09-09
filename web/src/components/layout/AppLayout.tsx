/**
 * AppLayout component wraps the design system AppLayout with a custom header.
 *
 * Includes:
 * - Sticky header with branding, navigation, and user info
 * - Main content area with responsive padding and premium styling
 */

import React, { ReactNode } from 'react';
import { useLocation, Link } from 'react-router-dom';
import { useConsent } from '@hooks/useConsent';
import { Avatar } from '@design-system/components/primitives/Avatar';
import { AppLayout as DesignSystemAppLayout } from '@design-system/components/layout/AppLayout/AppLayout';

interface AppLayoutProps {
  children: ReactNode;
  header?: React.ComponentType<Record<string, never>> | null;
}

/**
 * Header component with branding, navigation, and user info.
 * Premium design with gradient backdrop and refined typography.
 */
function Header() {
  const { userInfo } = useConsent();
  const location = useLocation();

  const navLinks = [
    { label: 'Agent Delegations', href: '/delegations' },
    { label: 'Tool Authorizations', href: '/approvals' },
    { label: 'Third-Party Sessions', href: '/sessions' },
  ];


  return (
    <div className="bg-white border-b border-neutral-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-20 gap-8">
          {/* Logo and branding */}
          <div className="flex items-center gap-3 flex-shrink-0">
            <div className="w-9 h-9 bg-gradient-to-br from-trust-hover to-trust rounded-lg flex items-center justify-center shadow-md">
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
              <h1 className="text-base font-bold text-trust-deep leading-tight">
                Consent Management
              </h1>
              <p className="text-xs text-slate-600 font-medium">
                Agentic Identity Broker
              </p>
            </div>
          </div>

          {/* Navigation */}
          <nav className="flex items-center gap-6 ml-auto">
            {/* Main navigation links */}
            <div className="hidden md:flex items-center gap-1">
              {navLinks.map((link) => {
                const isActive = location.pathname === link.href;

                return (
                  <Link
                    key={link.href}
                    to={link.href}
                    className={`px-3 py-2 rounded-md text-sm font-medium transition-colors duration-200 ${
                      isActive
                        ? 'bg-trust-light text-trust-deep'
                        : 'text-slate-700 hover:bg-slate-100 hover:text-slate-900'
                    }`}
                    aria-current={isActive ? 'page' : undefined}
                  >
                    {link.label}
                  </Link>
                );
              })}
            </div>

            {/* User info with avatar */}
            {userInfo && (
              <div className="flex items-center gap-3 pl-6 border-l border-slate-200">
                <Avatar
                  src={userInfo.pictureUrl}
                  alt={userInfo.displayName}
                  initials={userInfo.displayName.charAt(0).toUpperCase()}
                  size="sm"
                  shape="circle"
                />
                <div className="flex flex-col min-w-0">
                  <p className="text-sm font-medium text-trust-deep truncate">
                    {userInfo.displayName}
                  </p>
                  <p className="text-xs text-slate-600 truncate">
                    {userInfo.email || userInfo.principal}
                  </p>
                </div>
              </div>
            )}
          </nav>
        </div>
      </div>
    </div>
  );
}

/**
 * Main layout wrapper that extends the design system AppLayout with custom styling.
 * Provides consistent structure with premium gradient background and spacing.
 */
export function AppLayout({ children, header }: AppLayoutProps) {
  const HeaderComponent = header || Header;

  return (
    <div className="bg-gradient-to-br from-cream via-sand to-taupe/20">
      <DesignSystemAppLayout
        header={<HeaderComponent />}
        stickyHeader={true}
        className="bg-transparent"
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
          <div className="animate-fade-in-up">{children}</div>
        </div>
      </DesignSystemAppLayout>
    </div>
  );
}

export default AppLayout;
