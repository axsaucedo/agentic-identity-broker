import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import userEvent from '@testing-library/user-event';
import { CIMDAdvancedDetails } from './CIMDAdvancedDetails';

describe('CIMDAdvancedDetails', () => {
  const baseProps = {
    clientIdUrl: 'https://agent.example.com/client',
    redirectUri: 'https://agent.example.com/callback',
    requestedScopes: ['repo', 'user:email'],
  };

  it('renders collapsed by default', () => {
    render(<CIMDAdvancedDetails {...baseProps} />);
    expect(screen.getByText(/Advanced Details/i)).toBeInTheDocument();
    expect(screen.queryByText('https://agent.example.com/client')).toBeNull();
  });

  it('expands when clicked to reveal advanced details', async () => {
    const user = userEvent.setup();
    render(<CIMDAdvancedDetails {...baseProps} />);

    await user.click(screen.getByRole('button', { name: /Advanced Details/i }));

    expect(screen.getByText('https://agent.example.com/client')).toBeInTheDocument();
    expect(screen.getByText('https://agent.example.com/callback')).toBeInTheDocument();
    expect(screen.getByText('repo')).toBeInTheDocument();
    expect(screen.getByText('user:email')).toBeInTheDocument();
  });
});
