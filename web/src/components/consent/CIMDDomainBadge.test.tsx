import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CIMDDomainBadge } from './CIMDDomainBadge';

describe('CIMDDomainBadge', () => {
  it('renders the verified domain badge', () => {
    render(<CIMDDomainBadge domain="agent.example.com" />);
    expect(screen.getByText(/Verified domain/i)).toBeInTheDocument();
    expect(screen.getByText(/agent\.example\.com/)).toBeInTheDocument();
  });
});
