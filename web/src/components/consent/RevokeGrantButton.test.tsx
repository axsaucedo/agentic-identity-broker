/**
 * Tests for RevokeGrantButton component (T016).
 *
 * Covers:
 * - renders with aria-label containing agent name
 * - click opens RevokeGrantDialog
 * - dialog calls onRevoked after successful confirm
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RevokeGrantButton } from './RevokeGrantButton';
import { ToastProvider } from '@components/ui/Toast';
import * as consentApiModule from '@services/api/consent';

// Mock the consent API module
vi.mock('@services/api/consent', () => ({
  consentApi: {
    deleteGrant: vi.fn(),
  },
  ConsentApiService: vi.fn(),
}));

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <ToastProvider>{children}</ToastProvider>
);

const defaultProps = {
  agentId: 'agent-123',
  agentName: 'Research Assistant',
  onRevoked: vi.fn(),
};

describe('RevokeGrantButton', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders with aria-label containing agent name', () => {
    render(
      <Wrapper>
        <RevokeGrantButton {...defaultProps} />
      </Wrapper>,
    );

    const button = screen.getByRole('button', {
      name: /revoke all access for research assistant/i,
    });
    expect(button).toBeInTheDocument();
  });

  it('click opens RevokeGrantDialog', async () => {
    const user = userEvent.setup();

    render(
      <Wrapper>
        <RevokeGrantButton {...defaultProps} />
      </Wrapper>,
    );

    // Dialog should not be visible initially
    expect(
      screen.queryByRole('heading', { name: /revoke all access/i }),
    ).not.toBeInTheDocument();

    // Click the button
    await user.click(
      screen.getByRole('button', {
        name: /revoke all access for research assistant/i,
      }),
    );

    // Dialog should now be open
    expect(
      screen.getByRole('heading', { name: /revoke all access/i }),
    ).toBeInTheDocument();
  });

  it('dialog calls onRevoked after successful confirm', async () => {
    const onRevoked = vi.fn();
    const user = userEvent.setup();

    vi.spyOn(consentApiModule.consentApi, 'deleteGrant').mockResolvedValue(
      undefined,
    );

    render(
      <Wrapper>
        <RevokeGrantButton {...defaultProps} onRevoked={onRevoked} />
      </Wrapper>,
    );

    // Open the dialog
    await user.click(
      screen.getByRole('button', {
        name: /revoke all access for research assistant/i,
      }),
    );

    // Click Revoke All Access in the dialog
    await user.click(
      screen.getByRole('button', { name: /^revoke all access$/i }),
    );

    await waitFor(() => {
      expect(consentApiModule.consentApi.deleteGrant).toHaveBeenCalledWith(
        'agent-123',
      );
      expect(onRevoked).toHaveBeenCalledTimes(1);
    });
  });

  it('shows error toast and does not call onRevoked on API failure', async () => {
    const onRevoked = vi.fn();
    const user = userEvent.setup();

    vi.spyOn(consentApiModule.consentApi, 'deleteGrant').mockRejectedValue({
      status: 500,
      code: 'SERVER_ERROR',
      message: 'Service temporarily unavailable.',
    });

    render(
      <Wrapper>
        <RevokeGrantButton {...defaultProps} onRevoked={onRevoked} />
      </Wrapper>,
    );

    // Open the dialog
    await user.click(
      screen.getByRole('button', {
        name: /revoke all access for research assistant/i,
      }),
    );

    // Confirm in the dialog
    await user.click(
      screen.getByRole('button', { name: /^revoke all access$/i }),
    );

    await waitFor(() => {
      expect(onRevoked).not.toHaveBeenCalled();
    });
  });
});
