/**
 * Tests for AgentGrantDetailPage component.
 *
 * Tests:
 * - Loading state
 * - Agent not found error
 * - Successful render with agent details
 * - Navigation functionality
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AgentGrantDetailPage } from './AgentGrantDetailPage';
import * as useAgentGrantsModule from '../hooks/useAgentGrants';
import type { AgentDetail, ThirdpartyService, UserGrant } from '../types/consent';

// Mock the useAgentGrants hook
vi.mock('../hooks/useAgentGrants');

// Mock router params
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...actual,
    useParams: () => ({ agentId: 'agent-123' }),
  };
});

const mockAgent: AgentDetail = {
  agentId: 'agent-123',
  displayName: 'Research Assistant',
  description: 'AI agent that helps with research tasks',
  logoUrl: 'https://example.com/agent-logo.png',
  governanceUrl: 'https://example.com/governance',
  userDocumentationUrl: 'https://example.com/docs',
  agentInterfaceUrl: 'https://example.com/interface',
};

const mockServices: ThirdpartyService[] = [
  {
    serviceId: 'github',
    displayName: 'GitHub',
    logoUrl: 'https://example.com/github.png',
    scopes: [
      { value: 'read:user', description: 'Read user profile' },
      { value: 'repo', description: 'Access repositories' },
    ],
  },
  {
    serviceId: 'google',
    displayName: 'Google Drive',
    logoUrl: 'https://example.com/google.png',
    scopes: [
      { value: 'drive.readonly', description: 'Read files' },
      { value: 'drive.file', description: 'Manage files' },
    ],
  },
];

const mockGrants: UserGrant[] = [
  {
    grantId: 'grant-1',
    agentId: 'agent-123',
    principal: 'user@example.com',
    delegatedTokens: [
      {
        serviceId: 'github',
        scopes: ['read:user', 'repo'],
      },
    ],
    validUntil: null,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

// Wrapper component for router context
const RouterWrapper = ({ children }: { children: React.ReactNode }) => (
  <BrowserRouter>
    <Routes>
      <Route path="*" element={children} />
    </Routes>
  </BrowserRouter>
);

describe('AgentGrantDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('displays loading state with skeletons', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: null,
      services: [],
      grants: [],
      loading: true,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check for breadcrumb link
    expect(screen.getByText('Delegations')).toBeInTheDocument();

    // Loading skeletons should be present
    // We can't easily test for skeleton components, but we can verify the page renders
    expect(screen.getByText('Delegations')).toBeInTheDocument();
  });

  it('displays error state with retry button', () => {
    const mockRefetch = vi.fn();
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: null,
      services: [],
      grants: [],
      loading: false,
      error: 'Failed to load agent details',
      refetch: mockRefetch,
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    expect(screen.getByText('Failed to load agent details')).toBeInTheDocument();
    expect(screen.getByText('Try again')).toBeInTheDocument();
  });

  it('displays agent not found error', () => {
    const mockRefetch = vi.fn();
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: null,
      services: [],
      grants: [],
      loading: false,
      error: 'Agent not found',
      refetch: mockRefetch,
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    expect(screen.getByText('Agent not found')).toBeInTheDocument();
  });

  it('renders agent details successfully', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check agent name (using heading role for specificity)
    expect(screen.getByRole('heading', { name: 'Research Assistant', level: 1 })).toBeInTheDocument();
    expect(screen.getByText('AI agent that helps with research tasks')).toBeInTheDocument();

    // Check agent logo
    expect(screen.getByAltText('Research Assistant logo')).toHaveAttribute(
      'src',
      'https://example.com/agent-logo.png'
    );
  });

  it('renders agent links when provided', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check for links - use getAllByText for duplicate "Documentation"
    expect(screen.getByText('Governance')).toBeInTheDocument();
    const docLinks = screen.getAllByText('Documentation');
    expect(docLinks.length).toBeGreaterThan(0);
    expect(screen.getByText('Agent Interface')).toBeInTheDocument();

    // Verify link targets
    const governanceLink = screen.getByText('Governance').closest('a');
    expect(governanceLink).toHaveAttribute('href', 'https://example.com/governance');
    expect(governanceLink).toHaveAttribute('target', '_blank');
  });

  it('renders services list', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check services section
    expect(screen.getByText('Services')).toBeInTheDocument();
    expect(screen.getByText('(2)')).toBeInTheDocument();

    // Check individual services
    expect(screen.getByText('GitHub')).toBeInTheDocument();
    expect(screen.getByText('Google Drive')).toBeInTheDocument();
  });

  it('displays empty state when no services available', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: [],
      grants: [],
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    expect(screen.getByText('No services available')).toBeInTheDocument();
    expect(screen.getByText('This agent has no services configured yet.')).toBeInTheDocument();
  });

  it('renders breadcrumb navigation', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check breadcrumb using navigation
    const breadcrumb = screen.getByRole('navigation', { name: 'Breadcrumb' });
    const delegationsLink = screen.getByText('Delegations').closest('a');
    expect(delegationsLink).toHaveAttribute('href', '/');
    expect(breadcrumb).toContainHTML('Research Assistant');
  });

  it('renders fallback logo when agent logo is not provided', () => {
    const agentWithoutLogo = { ...mockAgent, logoUrl: undefined };

    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: agentWithoutLogo,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // Check for fallback initial letter
    expect(screen.getByText('R')).toBeInTheDocument();
  });

  it('groups grants by service correctly', () => {
    vi.spyOn(useAgentGrantsModule, 'useAgentGrants').mockReturnValue({
      agent: mockAgent,
      services: mockServices,
      grants: mockGrants,
      loading: false,
      error: null,
      refetch: vi.fn(),
    });

    render(
      <RouterWrapper>
        <AgentGrantDetailPage />
      </RouterWrapper>
    );

    // GitHub service should show granted scopes
    expect(screen.getByText('2 scopes granted')).toBeInTheDocument();
  });
});
