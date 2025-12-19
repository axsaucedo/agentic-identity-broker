/**
 * Integration tests for AgentGrantDetailPage grant submission flow.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { AgentGrantDetailPage } from './AgentGrantDetailPage';
import { consentApi } from '../services/api/consent';
import type { AgentDetail, ThirdpartyService, UserGrant } from '../types/consent';

// Mock the API
vi.mock('../services/api/consent', () => ({
  consentApi: {
    getAgentDetail: vi.fn(),
    getAgentGrants: vi.fn(),
    createOrUpdateGrant: vi.fn(),
  },
}));

describe('AgentGrantDetailPage - Integration', () => {
  const mockAgent: AgentDetail = {
    agentId: 'test-agent',
    displayName: 'Test Agent',
    description: 'A test agent',
    logoUrl: 'https://example.com/logo.png',
    governanceUrl: 'https://example.com/governance',
    userDocumentationUrl: 'https://example.com/docs',
    agentInterfaceUrl: 'https://example.com/interface',
  };

  const mockServices: ThirdpartyService[] = [
    {
      serviceId: 'github',
      displayName: 'GitHub',
      scopes: [
        { value: 'read:user', description: 'Read user profile' },
        { value: 'read:repo', description: 'Read repositories' },
      ],
    },
  ];

  const mockGrant: UserGrant = {
    grantId: 'grant-123',
    agentId: 'test-agent',
    principal: 'user@example.com',
    delegatedTokens: [
      {
        serviceId: 'github',
        scopes: ['read:user'],
      },
    ],
    validUntil: null,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  };

  beforeEach(() => {
    vi.clearAllMocks();

    // Setup default mock responses
    vi.mocked(consentApi.getAgentDetail).mockResolvedValue({
      agent: mockAgent,
      services: mockServices,
    });

    vi.mocked(consentApi.getAgentGrants).mockResolvedValue([mockGrant]);
  });

  const renderPage = () => {
    return render(
      <MemoryRouter initialEntries={['/agent/test-agent']}>
        <Routes>
          <Route path="/agent/:agentId" element={<AgentGrantDetailPage />} />
        </Routes>
      </MemoryRouter>
    );
  };

  it('should load and display agent information', async () => {
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
      expect(screen.getByText('A test agent')).toBeInTheDocument();
    });
  });

  it('should enable edit mode when toggle is clicked', async () => {
    renderPage();

    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
    });

    // Find and click the edit mode toggle
    const toggle = screen.getByLabelText(/View Mode/);
    fireEvent.click(toggle);

    await waitFor(() => {
      expect(screen.getByText('Edit Mode')).toBeInTheDocument();
    });
  });

  it('should submit grant successfully', async () => {
    const updatedGrant: UserGrant = {
      ...mockGrant,
      delegatedTokens: [
        {
          serviceId: 'github',
          scopes: ['read:user', 'read:repo'],
        },
      ],
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue(updatedGrant);

    renderPage();

    // Wait for page to load
    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
    });

    // Enable edit mode
    const editToggle = screen.getByLabelText(/View Mode/);
    fireEvent.click(editToggle);

    await waitFor(() => {
      expect(screen.getByText('Edit Mode')).toBeInTheDocument();
    });

    // Enable the service (it should already be enabled from existing grant)
    // Expand scopes
    const expandButton = screen.getByText(/Select Scopes/);
    fireEvent.click(expandButton);

    // Select additional scope
    await waitFor(() => {
      const repoCheckbox = screen.getByLabelText(/Read repositories/);
      fireEvent.click(repoCheckbox);
    });

    // Submit
    const submitButton = screen.getByText('Approve & Delegate');
    fireEvent.click(submitButton);

    // Wait for API call
    await waitFor(() => {
      expect(consentApi.createOrUpdateGrant).toHaveBeenCalledWith('test-agent', {
        delegatedTokens: expect.arrayContaining([
          expect.objectContaining({
            serviceId: 'github',
            scopes: expect.arrayContaining(['read:user', 'read:repo']),
          }),
        ]),
        validUntil: null,
      });
    });

    // Should show success message
    await waitFor(() => {
      expect(screen.getByText(/Grant updated successfully/)).toBeInTheDocument();
    });

    // Should exit edit mode
    await waitFor(() => {
      expect(screen.getByText('View Mode')).toBeInTheDocument();
    });
  });

  it('should display error when submission fails', async () => {
    vi.mocked(consentApi.createOrUpdateGrant).mockRejectedValue({
      response: {
        data: {
          message: 'Failed to update grant',
        },
      },
    });

    renderPage();

    // Wait for page to load
    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
    });

    // Enable edit mode
    const editToggle = screen.getByLabelText(/View Mode/);
    fireEvent.click(editToggle);

    await waitFor(() => {
      expect(screen.getByText('Edit Mode')).toBeInTheDocument();
    });

    // Enable a service
    const serviceToggle = screen.getByLabelText(/Disabled/);
    fireEvent.click(serviceToggle);

    // Expand and select a scope
    await waitFor(() => {
      const expandButton = screen.getByText(/Select Scopes/);
      fireEvent.click(expandButton);
    });

    await waitFor(() => {
      const checkbox = screen.getByLabelText(/Read user profile/);
      fireEvent.click(checkbox);
    });

    // Submit
    const submitButton = screen.getByText('Approve & Delegate');
    fireEvent.click(submitButton);

    // Should show error message
    await waitFor(() => {
      expect(screen.getByText(/Failed to update grant/)).toBeInTheDocument();
    });
  });

  it('should refetch data after successful submission', async () => {
    const updatedGrant: UserGrant = {
      ...mockGrant,
      delegatedTokens: [
        {
          serviceId: 'github',
          scopes: ['read:user', 'read:repo'],
        },
      ],
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue(updatedGrant);

    renderPage();

    // Wait for initial load
    await waitFor(() => {
      expect(consentApi.getAgentDetail).toHaveBeenCalledTimes(1);
      expect(consentApi.getAgentGrants).toHaveBeenCalledTimes(1);
    });

    // Enable edit mode
    await waitFor(() => {
      const editToggle = screen.getByLabelText(/View Mode/);
      fireEvent.click(editToggle);
    });

    // Enable service and select scope
    await waitFor(() => {
      const serviceToggle = screen.getByLabelText(/Disabled/);
      fireEvent.click(serviceToggle);
    });

    await waitFor(() => {
      const expandButton = screen.getByText(/Select Scopes/);
      fireEvent.click(expandButton);
    });

    await waitFor(() => {
      const checkbox = screen.getByLabelText(/Read user profile/);
      fireEvent.click(checkbox);
    });

    // Submit
    await waitFor(() => {
      const submitButton = screen.getByText('Approve & Delegate');
      fireEvent.click(submitButton);
    });

    // Should refetch data
    await waitFor(() => {
      expect(consentApi.getAgentDetail).toHaveBeenCalledTimes(2);
      expect(consentApi.getAgentGrants).toHaveBeenCalledTimes(2);
    });
  });

  it('should validate form before submission', async () => {
    renderPage();

    // Wait for page to load
    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
    });

    // Enable edit mode
    const editToggle = screen.getByLabelText(/View Mode/);
    fireEvent.click(editToggle);

    await waitFor(() => {
      expect(screen.getByText('Edit Mode')).toBeInTheDocument();
    });

    // Try to submit without making changes (button should be disabled)
    const submitButton = screen.getByText('Approve & Delegate');
    expect(submitButton).toBeDisabled();
  });

  it('should cancel edit mode', async () => {
    renderPage();

    // Wait for page to load
    await waitFor(() => {
      expect(screen.getByText('Test Agent')).toBeInTheDocument();
    });

    // Enable edit mode
    const editToggle = screen.getByLabelText(/View Mode/);
    fireEvent.click(editToggle);

    await waitFor(() => {
      expect(screen.getByText('Edit Mode')).toBeInTheDocument();
    });

    // Click cancel
    const cancelButton = screen.getByText('Cancel');
    fireEvent.click(cancelButton);

    // Should exit edit mode
    await waitFor(() => {
      expect(screen.getByText('View Mode')).toBeInTheDocument();
    });
  });
});
