import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import { CIMDSection } from './CIMDSection';
import type { CIMDMetadata, ThirdpartyService } from '../../types/consent';

const baseMeta: CIMDMetadata = {
  client_id_url: 'https://agent.example.com/cimd.json',
  redirect_uri: 'https://agent.example.com/callback',
  verified_domain: 'agent.example.com',
  is_localhost_redirect: false,
  requested_scopes: ['read:email'],
};

const matchingService: ThirdpartyService = {
  kind: 'scoped',
  serviceId: 'svc-1',
  displayName: 'GitHub',
  scopes: [{ value: 'read:email', description: 'Read email' }],
};

describe('CIMDSection', () => {
  it('renders summary and advanced details without localhost warning when is_localhost_redirect is false', () => {
    render(
      <CIMDSection
        cimdMeta={baseMeta}
        agentDisplayName="My Agent"
        services={[]}
      />,
    );

    expect(screen.getByText('agent.example.com')).toBeInTheDocument();
    expect(screen.getByText(/Advanced Details/i)).toBeInTheDocument();
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('renders summary, localhost warning, and advanced details when is_localhost_redirect is true', () => {
    const meta = { ...baseMeta, is_localhost_redirect: true };

    render(
      <CIMDSection
        cimdMeta={meta}
        agentDisplayName="My Agent"
        services={[]}
      />,
    );

    expect(screen.getByText('agent.example.com')).toBeInTheDocument();
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/Advanced Details/i)).toBeInTheDocument();
  });

  it('derives accessTarget from a service whose scope matches requested_scopes', () => {
    render(
      <CIMDSection
        cimdMeta={baseMeta}
        agentDisplayName="My Agent"
        services={[matchingService]}
      />,
    );

    expect(screen.getByText(/GitHub/)).toBeInTheDocument();
    expect(screen.getByText(/wants to access/)).toBeInTheDocument();
  });
});
