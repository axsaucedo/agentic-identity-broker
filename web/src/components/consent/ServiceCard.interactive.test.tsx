/**
 * Tests for ServiceCard interactive mode.
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ServiceCard } from './ServiceCard';
import type { ThirdpartyService } from '../../types/consent';

describe('ServiceCard - Interactive Mode', () => {
  const mockService: ThirdpartyService = {
    serviceId: 'github',
    displayName: 'GitHub',
    logoUrl: 'https://example.com/github-logo.png',
    scopes: [
      { value: 'read:user', description: 'Read user profile' },
      { value: 'read:repo', description: 'Read repository data' },
      { value: 'write:repo', description: 'Write repository data' },
    ],
  };

  const mockOnServiceToggle = vi.fn();
  const mockOnScopeChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render service toggle when editable', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={false}
        selectedScopes={[]}
      />
    );

    expect(screen.getByText('Disabled')).toBeInTheDocument();
  });

  it('should call onServiceToggle when switch is clicked', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={false}
        selectedScopes={[]}
      />
    );

    const toggle = screen.getByRole('switch');
    fireEvent.click(toggle);

    expect(mockOnServiceToggle).toHaveBeenCalledWith('github', true);
  });

  it('should show scope checkboxes when service is enabled', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={[]}
      />
    );

    // Expand scopes - use getByRole with aria-label
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    // Check that checkboxes are rendered
    const checkboxes = screen.getAllByRole('checkbox');
    expect(checkboxes.length).toBeGreaterThan(0);
  });

  it('should call onScopeChange when scope checkbox is toggled', async () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={[]}
      />
    );

    // Expand scopes
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    // Find and click first scope checkbox
    const checkbox = screen.getByLabelText(/Read user profile/);
    fireEvent.click(checkbox);

    expect(mockOnScopeChange).toHaveBeenCalledWith('github', ['read:user']);
  });

  it('should show selected scopes count', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={['read:user', 'read:repo']}
      />
    );

    expect(screen.getByText('2 scopes selected')).toBeInTheDocument();
  });

  it('should render select all and deselect all buttons', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={[]}
      />
    );

    // Expand scopes
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    expect(screen.getByText('Select all')).toBeInTheDocument();
    expect(screen.getByText('Deselect all')).toBeInTheDocument();
  });

  it('should select all scopes when "Select all" is clicked', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={[]}
      />
    );

    // Expand scopes
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    // Click select all
    const selectAllButton = screen.getByText('Select all');
    fireEvent.click(selectAllButton);

    expect(mockOnScopeChange).toHaveBeenCalledWith('github', [
      'read:user',
      'read:repo',
      'write:repo',
    ]);
  });

  it('should deselect all scopes when "Deselect all" is clicked', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={['read:user', 'read:repo']}
      />
    );

    // Expand scopes
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    // Click deselect all
    const deselectAllButton = screen.getByText('Deselect all');
    fireEvent.click(deselectAllButton);

    expect(mockOnScopeChange).toHaveBeenCalledWith('github', []);
  });

  it('should not show scope section when service is disabled in edit mode', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={false}
        selectedScopes={[]}
      />
    );

    expect(screen.queryByText(/Select Scopes/)).not.toBeInTheDocument();
  });

  it('should highlight selected scopes', () => {
    render(
      <ServiceCard
        service={mockService}
        isEditable={true}
        onServiceToggle={mockOnServiceToggle}
        onScopeChange={mockOnScopeChange}
        isServiceEnabled={true}
        selectedScopes={['read:user']}
      />
    );

    // Expand scopes
    const expandButton = screen.getByRole('button', { name: /select scopes/i });
    fireEvent.click(expandButton);

    // Check that the selected scope is checked
    const checkbox = screen.getByLabelText(/Read user profile/) as HTMLInputElement;
    expect(checkbox.checked).toBe(true);
  });
});
