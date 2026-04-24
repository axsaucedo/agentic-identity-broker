import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CIMDLocalhostWarning } from './CIMDLocalhostWarning';

describe('CIMDLocalhostWarning', () => {
  it('renders the localhost warning with agent display name', () => {
    render(<CIMDLocalhostWarning agentDisplayName="My Test Agent" />);
    expect(screen.getByText(/redirect to your local machine/i)).toBeInTheDocument();
    expect(screen.getByText(/My Test Agent/)).toBeInTheDocument();
  });

  it('has role="alert" for accessibility', () => {
    render(<CIMDLocalhostWarning agentDisplayName="Test Agent" />);
    expect(screen.getByRole('alert')).toBeInTheDocument();
  });
});
