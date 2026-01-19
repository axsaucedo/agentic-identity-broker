/**
 * Tests for ServiceRequirementsList component.
 *
 * Tests grouping by requirement type, rendering of sections,
 * and callback propagation.
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ServiceRequirementsList } from './ServiceRequirementsList';

describe('ServiceRequirementsList', () => {
  const mockOnLogin = vi.fn();
  const mockOnDisconnect = vi.fn();

  const mockScopes = [
    {
      name: 'read:email',
      description: 'Read your email address',
    },
  ];

  beforeEach(() => {
    mockOnLogin.mockClear();
    mockOnDisconnect.mockClear();
  });

  describe('Empty state', () => {
    it('shows empty state message when no requirements', () => {
      render(
        <ServiceRequirementsList
          requirements={[]}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(
        screen.getByText('No service requirements configured for this agent.')
      ).toBeInTheDocument();
    });
  });

  describe('Grouping by requirement type', () => {
    it('groups mandatory services first, then optional', () => {
      const requirements = [
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'google',
          serviceName: 'Google Drive',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      const { container } = render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      // Should show both section headers
      expect(screen.getByText('Required Services')).toBeInTheDocument();
      expect(screen.getByText('Optional Services')).toBeInTheDocument();

      // Mandatory services (GitHub, Google Drive) should appear before optional (Stripe)
      const allText = container.textContent || '';
      const githubIndex = allText.indexOf('GitHub');
      const googleIndex = allText.indexOf('Google Drive');
      const stripeIndex = allText.indexOf('Stripe');

      expect(githubIndex).toBeLessThan(stripeIndex);
      expect(googleIndex).toBeLessThan(stripeIndex);
    });

    it('renders only mandatory section when no optional services', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('Required Services')).toBeInTheDocument();
      expect(screen.queryByText('Optional Services')).not.toBeInTheDocument();
    });

    it('renders only optional section when no mandatory services', () => {
      const requirements = [
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('Optional Services')).toBeInTheDocument();
      expect(screen.queryByText('Required Services')).not.toBeInTheDocument();
    });
  });

  describe('Section headers', () => {
    it('displays "Required Services" header for mandatory services', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('Required Services')).toBeInTheDocument();
    });

    it('displays "Optional Services" header for optional services', () => {
      const requirements = [
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('Optional Services')).toBeInTheDocument();
    });
  });

  describe('Service rendering', () => {
    it('renders all services with their cards', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('GitHub')).toBeInTheDocument();
      expect(screen.getByText('Stripe')).toBeInTheDocument();
    });

    it('renders ServiceRequirementCard for each requirement', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'google',
          serviceName: 'Google Drive',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      // Should show Login button for not connected service
      expect(
        screen.getByRole('button', { name: /login/i })
      ).toBeInTheDocument();

      // Should show Active Session badge and Disconnect button for connected service
      expect(screen.getByText('Active Session')).toBeInTheDocument();
      expect(
        screen.getByRole('button', { name: /disconnect/i })
      ).toBeInTheDocument();
    });
  });

  describe('Callback propagation', () => {
    it('passes onLogin callback to ServiceRequirementCard', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      const loginButton = screen.getByRole('button', { name: /login/i });
      fireEvent.click(loginButton);

      expect(mockOnLogin).toHaveBeenCalledWith('github');
    });

    it('passes onDisconnect callback to ServiceRequirementCard', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      const disconnectButton = screen.getByRole('button', { name: /disconnect/i });
      fireEvent.click(disconnectButton);

      expect(mockOnDisconnect).toHaveBeenCalledWith('github');
    });

    it('handles multiple services with callbacks', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      // Click login for GitHub
      const loginButtons = screen.getAllByRole('button', { name: /login/i });
      fireEvent.click(loginButtons[0]);

      // Click disconnect for Stripe
      const disconnectButton = screen.getByRole('button', { name: /disconnect/i });
      fireEvent.click(disconnectButton);

      expect(mockOnLogin).toHaveBeenCalledWith('github');
      expect(mockOnDisconnect).toHaveBeenCalledWith('stripe');
    });
  });

  describe('Multiple services in sections', () => {
    it('renders multiple mandatory services in correct order', () => {
      const requirements = [
        {
          serviceId: 'github',
          serviceName: 'GitHub',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'google',
          serviceName: 'Google Drive',
          requirementType: 'mandatory' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('GitHub')).toBeInTheDocument();
      expect(screen.getByText('Google Drive')).toBeInTheDocument();

      // Both should be in Required Services section
      const requiredServicesHeader = screen.getByText('Required Services');
      const nextElement = requiredServicesHeader.nextElementSibling;
      const sectionText = nextElement?.textContent || '';

      expect(sectionText).toContain('GitHub');
      expect(sectionText).toContain('Google Drive');
    });

    it('renders multiple optional services in correct order', () => {
      const requirements = [
        {
          serviceId: 'stripe',
          serviceName: 'Stripe',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
        {
          serviceId: 'slack',
          serviceName: 'Slack',
          requirementType: 'optional' as const,
          requiredScopes: mockScopes,
          connectionStatus: 'not_connected' as const,
        },
      ];

      render(
        <ServiceRequirementsList
          requirements={requirements}
          onLogin={mockOnLogin}
          onDisconnect={mockOnDisconnect}
        />
      );

      expect(screen.getByText('Stripe')).toBeInTheDocument();
      expect(screen.getByText('Slack')).toBeInTheDocument();
    });
  });
});
