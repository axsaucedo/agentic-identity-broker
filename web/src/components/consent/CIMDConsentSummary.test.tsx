import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CIMDConsentSummary } from './CIMDConsentSummary';

describe('CIMDConsentSummary', () => {
  const baseProps = {
    domain: 'agent.example.com',
    agentName: 'My Agent',
    accessTarget: 'GitHub',
  };

  it('renders domain as the primary identifier', () => {
    render(<CIMDConsentSummary {...baseProps} />);
    expect(screen.getByText('agent.example.com')).toBeInTheDocument();
  });

  it('renders client name and access target in the secondary sentence', () => {
    render(<CIMDConsentSummary {...baseProps} />);
    expect(screen.getByText(/My Agent/)).toBeInTheDocument();
    expect(screen.getByText(/GitHub/)).toBeInTheDocument();
    expect(screen.getByText(/wants to access/)).toBeInTheDocument();
  });

  it('renders logo when agentLogoUrl is provided', () => {
    render(<CIMDConsentSummary {...baseProps} agentLogoUrl="https://example.com/logo.png" />);
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'https://example.com/logo.png');
    expect(img).toHaveAttribute('alt', 'My Agent logo');
    expect(img).toHaveAttribute('referrerpolicy', 'no-referrer');
  });

  it('renders without logo when agentLogoUrl is not provided', () => {
    render(<CIMDConsentSummary {...baseProps} />);
    expect(screen.queryByRole('img')).toBeNull();
  });
});
