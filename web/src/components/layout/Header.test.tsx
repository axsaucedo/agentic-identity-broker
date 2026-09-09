import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import * as useConsentModule from '@hooks/useConsent';
import { AppLayout } from './AppLayout';

vi.mock('@hooks/useConsent');

describe('AppLayout header', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(useConsentModule, 'useConsent').mockReturnValue({
      delegations: [],
      userInfo: null,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });
  });

  it('links top-level navigation to canonical browser routes', () => {
    render(
      <MemoryRouter>
        <AppLayout>content</AppLayout>
      </MemoryRouter>,
    );

    expect(
      screen.getByRole('link', { name: 'Third-Party Sessions' }),
    ).toHaveAttribute('href', '/sessions');

    expect(
      screen.getByRole('link', { name: 'Agent Delegations' }),
    ).toHaveAttribute('href', '/delegations');
    expect(
      screen.getByRole('link', { name: 'Tool Authorizations' }),
    ).toHaveAttribute('href', '/approvals');
  });

  it('marks agent delegations active at /delegations', () => {
    render(
      <MemoryRouter initialEntries={['/delegations']}>
        <AppLayout>content</AppLayout>
      </MemoryRouter>,
    );

    expect(
      screen.getByRole('link', { name: 'Agent Delegations' }),
    ).toHaveAttribute('aria-current', 'page');
  });
});
