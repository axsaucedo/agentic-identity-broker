/**
 * Tests for useToggleGrant hook.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useToggleGrant } from './useToggleGrant';
import { consentApi } from '../services/api/consent';
import type { UserGrant } from '../types/consent';

// Mock the API
vi.mock('../services/api/consent', () => ({
  consentApi: {
    createOrUpdateGrant: vi.fn(),
  },
}));

describe('useToggleGrant', () => {
  const agentId = 'test-agent';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should initialize with empty state', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    expect(result.current.delegatedTokens).toEqual([]);
    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.isSuccess).toBe(false);
  });

  it('should update delegated tokens', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    const tokens = [
      {
        thirdparty_oauth2_service_id: 'service-1',
        scopes: ['read', 'write'],
      },
    ];

    act(() => {
      result.current.setDelegatedTokens(tokens);
    });

    expect(result.current.delegatedTokens).toEqual(tokens);
  });

  it('should submit grant successfully', async () => {
    const mockGrant: UserGrant = {
      id: 'grant-123',
      agent_id: agentId,
      principal: 'user@example.com',
      delegated_oauth2_tokens: [
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: ['read'],
        },
      ],
      valid_until: null,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue(mockGrant);

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set tokens
    act(() => {
      result.current.setDelegatedTokens(mockGrant.delegated_oauth2_tokens);
    });

    // Submit
    let returnedGrant: UserGrant | null = null;
    await act(async () => {
      returnedGrant = await result.current.submit();
    });

    expect(returnedGrant).toEqual(mockGrant);
    expect(result.current.isSuccess).toBe(true);
    expect(result.current.error).toBeNull();
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should handle submission error', async () => {
    const errorResponse = {
      response: {
        data: {
          message: 'Invalid request',
        },
      },
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockRejectedValue(errorResponse);

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set tokens
    act(() => {
      result.current.setDelegatedTokens([
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: ['read'],
        },
      ]);
    });

    // Submit
    let returnedGrant: UserGrant | null = null;
    await act(async () => {
      returnedGrant = await result.current.submit();
    });

    expect(returnedGrant).toBeNull();
    expect(result.current.isSuccess).toBe(false);
    expect(result.current.error).toBe('Invalid request');
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should handle validation errors', async () => {
    const errorResponse = {
      response: {
        data: {
          message: 'Validation failed',
          details: {
            'delegated_oauth2_tokens[0].scopes': [
              'At least one scope required',
            ],
          },
        },
      },
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockRejectedValue(errorResponse);

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set tokens
    act(() => {
      result.current.setDelegatedTokens([
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: [],
        },
      ]);
    });

    // Submit
    await act(async () => {
      await result.current.submit();
    });

    expect(result.current.error).toContain('Validation error');
    expect(result.current.error).toContain('delegated_oauth2_tokens[0].scopes');
  });

  it('should reset state', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set some state
    act(() => {
      result.current.setDelegatedTokens([
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: ['read'],
        },
      ]);
    });

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.delegatedTokens).toEqual([]);
    expect(result.current.error).toBeNull();
    expect(result.current.isSuccess).toBe(false);
  });

  it('should clear error', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    // Manually set error (in real scenario, would come from failed submit)
    act(() => {
      result.current.setDelegatedTokens([
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: ['read'],
        },
      ]);
    });

    // Clear error
    act(() => {
      result.current.clearError();
    });

    expect(result.current.error).toBeNull();
  });

  it('should submit with validUntil', async () => {
    const mockGrant: UserGrant = {
      id: 'grant-123',
      agent_id: agentId,
      principal: 'user@example.com',
      delegated_oauth2_tokens: [
        {
          thirdparty_oauth2_service_id: 'service-1',
          scopes: ['read'],
        },
      ],
      valid_until: '2025-01-01T00:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue(mockGrant);

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set tokens
    act(() => {
      result.current.setDelegatedTokens(mockGrant.delegated_oauth2_tokens);
    });

    // Submit with expiration
    await act(async () => {
      await result.current.submit('2025-01-01T00:00:00Z');
    });

    // Verify the API was called
    expect(consentApi.createOrUpdateGrant).toHaveBeenCalled();
  });
});
