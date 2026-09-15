/**
 * Tests for DelegationCard component — revoke action (T028).
 *
 * Covers:
 * - "Revoke" button rendered when onRevoke prop provided
 * - button absent when onRevoke not provided
 * - click calls onRevoke with correct agentId
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DelegationCard } from './DelegationCard';
import type { AgentDelegation } from '../../types/consent';

const mockDelegation: AgentDelegation = {
  agentId: 'agent-abc',
  displayName: 'Research Assistant',
  logoUrl: undefined,
  activeGrantCount: 2,
  lastModifiedAt: new Date('2024-01-15T10:00:00Z').toISOString(),
  expiresAt: null,
};

describe('DelegationCard — revoke action', () => {
  it('renders Revoke button when onRevoke prop is provided', () => {
    const onRevoke = vi.fn();

    render(
      <DelegationCard
        delegation={mockDelegation}
        onClick={vi.fn()}
        onRevoke={onRevoke}
      />,
    );

    expect(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    ).toBeInTheDocument();
  });

  it('does not render Revoke button when onRevoke is not provided', () => {
    render(
      <DelegationCard delegation={mockDelegation} onClick={vi.fn()} />,
    );

    expect(
      screen.queryByRole('button', { name: /revoke/i }),
    ).not.toBeInTheDocument();
  });

  it('click calls onRevoke with the correct agentId', async () => {
    const onRevoke = vi.fn();
    const user = userEvent.setup();

    render(
      <DelegationCard
        delegation={mockDelegation}
        onClick={vi.fn()}
        onRevoke={onRevoke}
      />,
    );

    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    expect(onRevoke).toHaveBeenCalledTimes(1);
    expect(onRevoke).toHaveBeenCalledWith('agent-abc');
  });

  it('clicking Revoke does not trigger the card onClick', async () => {
    const onClick = vi.fn();
    const onRevoke = vi.fn();
    const user = userEvent.setup();

    render(
      <DelegationCard
        delegation={mockDelegation}
        onClick={onClick}
        onRevoke={onRevoke}
      />,
    );

    await user.click(
      screen.getByRole('button', { name: /revoke access for research assistant/i }),
    );

    expect(onRevoke).toHaveBeenCalledTimes(1);
    // stopPropagation should prevent card click
    expect(onClick).not.toHaveBeenCalled();
  });
});
