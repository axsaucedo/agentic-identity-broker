import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CIMDConsentSummary } from './CIMDConsentSummary';

describe('CIMDConsentSummary', () => {
  const baseProps = {
    clientName: 'My Agent',
    accessTarget: 'GitHub',
  };

  it('renders the summary statement with client_name and access target', () => {
    render(<CIMDConsentSummary {...baseProps} />);
    expect(screen.getByText(/The application/)).toBeInTheDocument();
    expect(screen.getByText(/My Agent/)).toBeInTheDocument();
    expect(screen.getByText(/GitHub/)).toBeInTheDocument();
  });

  it('renders logo when logoUri is provided', () => {
    render(<CIMDConsentSummary {...baseProps} logoUri="https://example.com/logo.png" />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'https://example.com/logo.png');
    expect(img).toHaveAttribute('alt', 'My Agent logo');
    expect(img).toHaveAttribute('referrerpolicy', 'no-referrer');
  });

  it('renders without logo when logoUri is not provided', () => {
    render(<CIMDConsentSummary {...baseProps} />);
    expect(screen.queryByRole('img')).toBeNull();
  });
});
