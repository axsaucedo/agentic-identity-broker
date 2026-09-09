import { render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { consentApi } from '@services/api/consent';
import { sessionsApi } from '@services/api/sessions';
import App from './App';

vi.mock('@services/api/consent', () => ({
  consentApi: {
    getUserInfo: vi.fn(),
    getAgentDelegations: vi.fn(),
  },
}));

vi.mock('@services/api/sessions', () => ({
  sessionsApi: {
    listSessions: vi.fn(),
  },
}));

vi.mock('./pages/AgentGrantDetailPage', () => ({
  default: () => <h1>Agent grant detail</h1>,
}));

vi.mock('./pages/ToolAuthorizationsPage', () => ({
  default: () => <h1>Tool approvals</h1>,
}));

describe('App routes', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.history.replaceState({}, '', '/sessions');
    vi.mocked(consentApi.getUserInfo).mockResolvedValue({
      principal: 'user@example.com',
      displayName: 'Test User',
      pictureUrl: 'https://example.com/avatar.png',
    });
    vi.mocked(consentApi.getAgentDelegations).mockResolvedValue([]);
    vi.mocked(sessionsApi.listSessions).mockResolvedValue([]);
  });

  afterEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('renders the sessions view at /sessions', async () => {
    render(<App />);

    expect(
      await screen.findByRole(
        'heading',
        { name: 'No Sessions Found' },
        { timeout: 5_000 },
      ),
    ).toBeInTheDocument();
  });

  it('renders agent delegations at /delegations', async () => {
    window.history.replaceState({}, '', '/delegations');

    render(<App />);

    expect(
      await screen.findByRole(
        'heading',
        { name: 'My Agent Delegations' },
        { timeout: 5_000 },
      ),
    ).toBeInTheDocument();
  });

  it('redirects the root route to /delegations', async () => {
    window.history.replaceState({}, '', '/');

    render(<App />);

    await waitFor(() => {
      expect(window.location.pathname).toBe('/delegations');
    });
  });

  it('renders agent detail at /agents/:agentId', async () => {
    window.history.replaceState({}, '', '/agents/agent-123');

    render(<App />);

    expect(
      await screen.findByRole('heading', { name: 'Agent grant detail' }),
    ).toBeInTheDocument();
  });

  it('renders tool approvals at /approvals', async () => {
    window.history.replaceState({}, '', '/approvals');

    render(<App />);

    expect(
      await screen.findByRole('heading', { name: 'Tool approvals' }),
    ).toBeInTheDocument();
  });
});
