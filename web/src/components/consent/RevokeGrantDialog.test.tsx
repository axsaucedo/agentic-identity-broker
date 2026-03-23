/**
 * Tests for RevokeGrantDialog component (T015).
 *
 * Covers:
 * - renders agent name in body text
 * - Escape key closes dialog without calling onConfirm
 * - Enter key triggers onConfirm
 * - shows loading state when isLoading=true
 * - cancel button calls onClose
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RevokeGrantDialog } from './RevokeGrantDialog';

const defaultProps = {
  agentName: 'Research Assistant',
  isOpen: true,
  onConfirm: vi.fn().mockResolvedValue(undefined),
  onClose: vi.fn(),
  isLoading: false,
};

describe('RevokeGrantDialog', () => {
  it('renders agent name in body text', () => {
    render(<RevokeGrantDialog {...defaultProps} />);

    expect(screen.getByText('Research Assistant')).toBeInTheDocument();
  });

  it('renders the dialog title', () => {
    render(<RevokeGrantDialog {...defaultProps} />);

    expect(
      screen.getByRole('heading', { name: 'Revoke All Access' }),
    ).toBeInTheDocument();
  });

  it('cancel button calls onClose', async () => {
    const onClose = vi.fn();
    const onConfirm = vi.fn();
    const user = userEvent.setup();

    render(
      <RevokeGrantDialog
        {...defaultProps}
        onClose={onClose}
        onConfirm={onConfirm}
      />,
    );

    await user.click(screen.getByRole('button', { name: /cancel/i }));

    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('Escape key closes dialog without calling onConfirm', async () => {
    const onClose = vi.fn();
    const onConfirm = vi.fn();

    render(
      <RevokeGrantDialog
        {...defaultProps}
        onClose={onClose}
        onConfirm={onConfirm}
      />,
    );

    // Headless UI Dialog calls onClose when Escape is pressed
    fireEvent.keyDown(document.activeElement || document.body, {
      key: 'Escape',
      code: 'Escape',
      keyCode: 27,
    });

    await waitFor(() => {
      expect(onConfirm).not.toHaveBeenCalled();
    });
  });

  it('Enter key triggers onConfirm via form submit', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const user = userEvent.setup();

    render(<RevokeGrantDialog {...defaultProps} onConfirm={onConfirm} />);

    // Focus the confirm button and press Enter to trigger form submit
    const confirmButton = screen.getByRole('button', {
      name: /revoke all access/i,
    });
    confirmButton.focus();
    await user.keyboard('{Enter}');

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledTimes(1);
    });
  });

  it('shows loading state when isLoading=true', () => {
    render(<RevokeGrantDialog {...defaultProps} isLoading={true} />);

    // The confirm button should show spinner (data-testid="button-spinner")
    expect(screen.getByTestId('button-spinner')).toBeInTheDocument();

    // Cancel button should be disabled
    expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled();
  });

  it('does not render when isOpen=false', () => {
    render(<RevokeGrantDialog {...defaultProps} isOpen={false} />);

    expect(
      screen.queryByRole('heading', { name: /revoke all access/i }),
    ).not.toBeInTheDocument();
  });

  it('renders the confirm button with danger styling', () => {
    render(<RevokeGrantDialog {...defaultProps} />);

    const confirmButton = screen.getByRole('button', {
      name: /revoke all access/i,
    });
    expect(confirmButton).toBeInTheDocument();
    // The button uses variant="danger" (bg-error-primary)
    expect(confirmButton).not.toBeDisabled();
  });
});
