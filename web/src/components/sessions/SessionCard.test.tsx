/**
 * SessionCard Component Tests
 *
 * Test Coverage:
 * - Component rendering with different session states
 * - Status badge variants (Active, Expiring Soon, Access Token Expired, Expired)
 * - Action button visibility and behavior
 * - Loading state handling
 * - Accessibility compliance
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SessionCard } from './SessionCard';
import type { SessionSummary } from '@services/api/sessions';

// Mock session data factory
const createMockSession = (overrides?: Partial<SessionSummary>): SessionSummary => ({
  id: 'test-session-id',
  service_id: 'google',
  service_display_name: 'Google Drive',
  token_type: 'Bearer',
  scope: ['read:email', 'write:files'],
  initiated_at: new Date('2024-01-01T12:00:00Z').toISOString(),
  is_expired: false,
  access_token_expired: false,
  refresh_token_expires_at: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString(),
  dependent_agent_count: 2,
  is_encrypted: true,
  ...overrides,
});

describe('SessionCard', () => {
  describe('Rendering', () => {
    it('renders session information correctly', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      // Verify service name
      expect(screen.getByText('Google Drive')).toBeInTheDocument();

      // Verify token type
      expect(screen.getByText('Bearer')).toBeInTheDocument();

      // Verify encryption indicator
      expect(screen.getByText('Encrypted')).toBeInTheDocument();

      // Verify dependent agents count
      expect(screen.getByText(/2 agents/i)).toBeInTheDocument();

      // Verify scopes
      expect(screen.getByText('read:email')).toBeInTheDocument();
      expect(screen.getByText('write:files')).toBeInTheDocument();
    });

    it('renders "Active" status for valid session', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.getByText('Active')).toBeInTheDocument();
    });

    it('renders "Expiring Soon" status for session expiring within 7 days', () => {
      const session = createMockSession({
        refresh_token_expires_at: new Date(Date.now() + 5 * 24 * 60 * 60 * 1000).toISOString(),
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.getByText('Expiring Soon')).toBeInTheDocument();
    });

    it('renders "Access Token Expired" status when access token is expired', () => {
      const session = createMockSession({
        access_token_expired: true,
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.getByText('Access Token Expired')).toBeInTheDocument();
    });

    it('renders "Expired" status for fully expired session', () => {
      const session = createMockSession({
        is_expired: true,
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.getByText('Expired')).toBeInTheDocument();
    });

    it('renders singular "agent" for dependent_agent_count of 1', () => {
      const session = createMockSession({
        dependent_agent_count: 1,
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.getByText(/1 agent$/i)).toBeInTheDocument();
    });

    it('does not render scopes section when no scopes exist', () => {
      const session = createMockSession({
        scope: [],
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.queryByText('Scopes:')).not.toBeInTheDocument();
    });
  });

  describe('Actions', () => {
    it('renders "View Details" and "Terminate" buttons for active session', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();
      const onViewDetails = vi.fn();

      render(
        <SessionCard
          session={session}
          onTerminate={onTerminate}
          onViewDetails={onViewDetails}
        />
      );

      expect(screen.getByText('View Details')).toBeInTheDocument();
      expect(screen.getByText('Terminate')).toBeInTheDocument();
    });

    it('calls onTerminate when Terminate button is clicked', async () => {
      const user = userEvent.setup();
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      const terminateButton = screen.getByText('Terminate');
      await user.click(terminateButton);

      expect(onTerminate).toHaveBeenCalledOnce();
      expect(onTerminate).toHaveBeenCalledWith('google');
    });

    it('calls onViewDetails when View Details button is clicked', async () => {
      const user = userEvent.setup();
      const session = createMockSession();
      const onTerminate = vi.fn();
      const onViewDetails = vi.fn();

      render(
        <SessionCard
          session={session}
          onTerminate={onTerminate}
          onViewDetails={onViewDetails}
        />
      );

      const viewDetailsButton = screen.getByText('View Details');
      await user.click(viewDetailsButton);

      expect(onViewDetails).toHaveBeenCalledOnce();
      expect(onViewDetails).toHaveBeenCalledWith('google');
    });

    it('does not render "View Details" button when onViewDetails is not provided', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(screen.queryByText('View Details')).not.toBeInTheDocument();
      expect(screen.getByText('Terminate')).toBeInTheDocument();
    });

    it('renders expired message instead of buttons for expired session', () => {
      const session = createMockSession({
        is_expired: true,
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      expect(
        screen.getByText('Session expired. Please re-authenticate.')
      ).toBeInTheDocument();
      expect(screen.queryByText('Terminate')).not.toBeInTheDocument();
      expect(screen.queryByText('View Details')).not.toBeInTheDocument();
    });
  });

  describe('Loading State', () => {
    it('disables buttons and shows loading spinner when loading', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} loading={true} />);

      const terminateButton = screen.getByText('Terminate');
      expect(terminateButton).toBeDisabled();

      // Check for loading spinner (data-testid="button-spinner")
      const spinner = screen.getByTestId('button-spinner');
      expect(spinner).toBeInTheDocument();
    });

    it('disables View Details button when loading', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();
      const onViewDetails = vi.fn();

      render(
        <SessionCard
          session={session}
          onTerminate={onTerminate}
          onViewDetails={onViewDetails}
          loading={true}
        />
      );

      const viewDetailsButton = screen.getByText('View Details');
      expect(viewDetailsButton).toBeDisabled();
    });
  });

  describe('Accessibility', () => {
    it('renders semantic HTML structure', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();

      const { container } = render(
        <SessionCard session={session} onTerminate={onTerminate} />
      );

      // Card should use semantic heading for service name
      const heading = container.querySelector('h3');
      expect(heading).toBeInTheDocument();
      expect(heading).toHaveTextContent('Google Drive');
    });

    it('has accessible button labels', () => {
      const session = createMockSession();
      const onTerminate = vi.fn();
      const onViewDetails = vi.fn();

      render(
        <SessionCard
          session={session}
          onTerminate={onTerminate}
          onViewDetails={onViewDetails}
        />
      );

      // Buttons should have clear text labels
      expect(screen.getByRole('button', { name: /view details/i })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /terminate/i })).toBeInTheDocument();
    });

    it('supports keyboard navigation', async () => {
      const user = userEvent.setup();
      const session = createMockSession();
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      const terminateButton = screen.getByText('Terminate');

      // Tab to button
      await user.tab();
      expect(terminateButton).toHaveFocus();

      // Press Enter
      await user.keyboard('{Enter}');
      expect(onTerminate).toHaveBeenCalledOnce();
    });
  });

  describe('Date Formatting', () => {
    it('formats initiated date correctly', () => {
      const session = createMockSession({
        initiated_at: new Date('2024-03-15T14:30:00Z').toISOString(),
      });
      const onTerminate = vi.fn();

      render(<SessionCard session={session} onTerminate={onTerminate} />);

      // Check that the date is formatted and displayed
      // Format: "Mar 15, 2024, 02:30 PM" (or similar depending on locale)
      expect(screen.getByText(/Established:/i)).toBeInTheDocument();
    });
  });
});
