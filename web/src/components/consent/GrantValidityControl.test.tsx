/**
 * Tests for GrantValidityControl component.
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { GrantValidityControl } from './GrantValidityControl';
import type { GrantValidityState } from '../../types/consent';

describe('GrantValidityControl', () => {
  const mockOnChange = vi.fn();

  const defaultState: GrantValidityState = {
    noExpiration: true,
    expiresAt: undefined,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render with no expiration by default', () => {
    render(<GrantValidityControl value={defaultState} onChange={mockOnChange} />);

    const checkbox = screen.getByRole('checkbox') as HTMLInputElement;
    expect(checkbox.checked).toBe(false);

    // Date picker should not be visible
    expect(screen.queryByLabelText('Expiration date')).not.toBeInTheDocument();
  });

  it('should show date picker when checkbox is checked', () => {
    render(<GrantValidityControl value={defaultState} onChange={mockOnChange} />);

    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);

    expect(mockOnChange).toHaveBeenCalledWith(
      expect.objectContaining({
        noExpiration: false,
        expiresAt: expect.any(Date),
      })
    );
  });

  it('should hide date picker when checkbox is unchecked', () => {
    // Use a future date
    const futureDate = new Date();
    futureDate.setFullYear(futureDate.getFullYear() + 1);

    const stateWithExpiration: GrantValidityState = {
      noExpiration: false,
      expiresAt: futureDate,
    };

    const { rerender } = render(
      <GrantValidityControl value={stateWithExpiration} onChange={mockOnChange} />
    );

    // Date picker should be visible
    expect(screen.getByLabelText('Expiration date')).toBeInTheDocument();

    // Uncheck
    const checkbox = screen.getByRole('checkbox');
    fireEvent.click(checkbox);

    expect(mockOnChange).toHaveBeenCalledWith({
      noExpiration: true,
      expiresAt: undefined,
    });
  });

  it('should call onChange when date is selected', () => {
    // Use a future date
    const futureDate = new Date();
    futureDate.setFullYear(futureDate.getFullYear() + 1);
    const futureDateString = futureDate.toISOString().split('T')[0];

    const stateWithExpiration: GrantValidityState = {
      noExpiration: false,
      expiresAt: futureDate,
    };

    render(<GrantValidityControl value={stateWithExpiration} onChange={mockOnChange} />);

    const dateInput = screen.getByLabelText('Expiration date') as HTMLInputElement;

    // Change to a different future date
    const newFutureDate = new Date(futureDate);
    newFutureDate.setMonth(newFutureDate.getMonth() + 3);
    const newFutureDateString = newFutureDate.toISOString().split('T')[0];

    fireEvent.change(dateInput, { target: { value: newFutureDateString } });

    expect(mockOnChange).toHaveBeenCalledWith({
      noExpiration: false,
      expiresAt: expect.any(Date),
    });
  });

  it('should render suggested dates', () => {
    const futureDate = new Date();
    futureDate.setFullYear(futureDate.getFullYear() + 1);

    const stateWithExpiration: GrantValidityState = {
      noExpiration: false,
      expiresAt: futureDate,
    };

    render(<GrantValidityControl value={stateWithExpiration} onChange={mockOnChange} />);

    expect(screen.getByText('1 month', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('3 months', { exact: false })).toBeInTheDocument();
    expect(screen.getByText('1 year', { exact: false })).toBeInTheDocument();
  });

  it('should call onChange when suggested date is clicked', () => {
    const futureDate = new Date();
    futureDate.setFullYear(futureDate.getFullYear() + 1);

    const stateWithExpiration: GrantValidityState = {
      noExpiration: false,
      expiresAt: futureDate,
    };

    render(<GrantValidityControl value={stateWithExpiration} onChange={mockOnChange} />);

    const oneMonthButton = screen.getByText('1 month', { exact: false });
    fireEvent.click(oneMonthButton);

    expect(mockOnChange).toHaveBeenCalledWith({
      noExpiration: false,
      expiresAt: expect.any(Date),
    });

    // Check that the date is roughly 1 month from now
    const calledDate = mockOnChange.mock.calls[0][0].expiresAt;
    const now = new Date();
    const expectedDate = new Date(now);
    expectedDate.setMonth(expectedDate.getMonth() + 1);

    // Allow 1 day difference for test execution time
    const diff = Math.abs(calledDate.getTime() - expectedDate.getTime());
    expect(diff).toBeLessThan(24 * 60 * 60 * 1000);
  });

  it('should show current date', () => {
    render(<GrantValidityControl value={defaultState} onChange={mockOnChange} />);

    expect(screen.getByText(/Today:/)).toBeInTheDocument();
  });

  it('should show indefinite message when no expiration', () => {
    render(<GrantValidityControl value={defaultState} onChange={mockOnChange} />);

    expect(screen.getByText(/remain active indefinitely/)).toBeInTheDocument();
  });

  it('should validate future dates', () => {
    const pastDate = new Date('2020-01-01');
    const stateWithPastDate: GrantValidityState = {
      noExpiration: false,
      expiresAt: pastDate,
    };

    render(<GrantValidityControl value={stateWithPastDate} onChange={mockOnChange} />);

    expect(screen.getByText(/must be in the future/)).toBeInTheDocument();
  });
});
