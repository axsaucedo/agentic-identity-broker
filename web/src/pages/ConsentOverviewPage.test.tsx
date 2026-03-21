/**
 * Tests for ConsentOverviewPage — revoke flow (T029).
 *
 * Covers:
 * - RevokeGrantDialog opens when onRevoke called on a card
 * - card removed from list after successful confirm (deleteGrant resolves)
 * - success toast shown
 * - card remains on cancel (dialog closed without confirm)
 * - error state shown when deleteGrant rejects
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BrowserRouter } from 'react-router-dom';
import { ConsentOverviewPage } from './ConsentOverviewPage';
import { ToastProvider } from '@components/ui/Toast';
import * as useConsentModule from '@hooks/useConsent';
import * as consentApiModule from '@services/api/consent';
import type { AgentDelegation } from '../types/consent';

vi.mock('@hooks/useConsent');
vi.mock('@services/api/consent', () => ({
  consentApi: {
    deleteGrant: vi.fn(),
  },
  ConsentApiService: vi.fn(),
}));

const makeDelegation = (overrides?: Partial<AgentDelegation>): AgentDelegation => ({
  agentId: 'agent-123',
  displayName: 'Research Assistant',
  activeGrantCount: 2,
  lastModifiedAt: new Date('2024-01-15T10:00:00Z').toISOString(),
  expiresAt: null,
  ...overrides,
});

const Wrapper = ({ children }: { children: React.ReactNode }) => (
  <ToastProvider>
    <BrowserRouter>{children}</BrowserRouter>
  </ToastProvider>
);

describe('ConsentOverviewPage — revoke flow', () => {
  const mockRefetch = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    mockRefetch.mockResolvedValue(undefined);
  });

  it('RevokeGrantDialog opens when onRevoke called via card Revoke button', async () => {
    const delegations = [makeDelegation()];
    const user = userEvent.setup();

    vi.spyOn(useConsentModule, 'useConsent').mockReturnValue({
      delegations,
      userInfo: null,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(
      <Wrapper>
        <ConsentOverviewPage />
      </Wrapper>,
    );

    // Dialog should not be visible initially
    expect(
      screen.queryByRole('heading', { name: /revoke all access/i }),
    ).not.toBeInTheDocument();

    // Click the Revoke button on the card
    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    // Dialog should now be open
    expect(
      screen.getByRole('heading', { name: /revoke all access/i }),
    ).toBeInTheDocument();
  });

  it('shows success toast and calls refetch after successful confirm', async () => {
    const delegations = [makeDelegation()];
    const user = userEvent.setup();

    vi.spyOn(useConsentModule, 'useConsent').mockReturnValue({
      delegations,
      userInfo: null,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    vi.spyOn(consentApiModule.consentApi, 'deleteGrant').mockResolvedValue(
      undefined,
    );

    render(
      <Wrapper>
        <ConsentOverviewPage />
      </Wrapper>,
    );

    // Open dialog
    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    // Confirm
    await user.click(
      screen.getByRole('button', { name: /^revoke all access$/i }),
    );

    await waitFor(() => {
      expect(consentApiModule.consentApi.deleteGrant).toHaveBeenCalledWith(
        'agent-123',
      );
      expect(mockRefetch).toHaveBeenCalled();
    });

    // Success toast
    await waitFor(() => {
      expect(screen.getByText(/all access has been revoked/i)).toBeInTheDocument();
    });
  });

  it('dialog closes on cancel without calling deleteGrant', async () => {
    const delegations = [makeDelegation()];
    const user = userEvent.setup();

    vi.spyOn(useConsentModule, 'useConsent').mockReturnValue({
      delegations,
      userInfo: null,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    render(
      <Wrapper>
        <ConsentOverviewPage />
      </Wrapper>,
    );

    // Open dialog
    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    expect(
      screen.getByRole('heading', { name: /revoke all access/i }),
    ).toBeInTheDocument();

    // Cancel
    await user.click(screen.getByRole('button', { name: /^cancel$/i }));

    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: /revoke all access/i }),
      ).not.toBeInTheDocument();
    });

    expect(consentApiModule.consentApi.deleteGrant).not.toHaveBeenCalled();
  });

  it('shows error toast when deleteGrant rejects', async () => {
    const delegations = [makeDelegation()];
    const user = userEvent.setup();

    vi.spyOn(useConsentModule, 'useConsent').mockReturnValue({
      delegations,
      userInfo: null,
      loading: false,
      error: null,
      refetch: mockRefetch,
    });

    vi.spyOn(consentApiModule.consentApi, 'deleteGrant').mockRejectedValue({
      status: 500,
      code: 'SERVER_ERROR',
      message: 'Service temporarily unavailable.',
    });

    render(
      <Wrapper>
        <ConsentOverviewPage />
      </Wrapper>,
    );

    // Open dialog
    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    // Confirm
    await user.click(
      screen.getByRole('button', { name: /^revoke all access$/i }),
    );

    await waitFor(() => {
      expect(
        screen.getByText(/service temporarily unavailable/i),
      ).toBeInTheDocument();
    });

    expect(mockRefetch).not.toHaveBeenCalled();
  });
});
