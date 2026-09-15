/**
 * Tests for useToggleGrant hook.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import axios from 'axios';
import { useToggleGrant } from './useToggleGrant';
import { consentApi } from '../services/api/consent';
import type { UserGrant, GrantResult } from '../types/consent';

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

    expect(result.current.grantedPermissionSets).toEqual({});
    expect(result.current.isSubmitting).toBe(false);
    expect(result.current.error).toBeNull();
    expect(result.current.isSuccess).toBe(false);
  });

  it('should update granted permission sets', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    const ps: Record<string, string[]> = {
      'ps-id-1': ['svc-1', 'svc-2'],
      'ps-id-2': ['svc-1'],
    };

    act(() => {
      result.current.setGrantedPermissionSets(ps);
    });

    expect(result.current.grantedPermissionSets).toEqual(ps);
  });

  it('should submit grant successfully', async () => {
    const mockGrant: UserGrant = {
      id: 'grant-123',
      agent_id: agentId,
      principal: 'user@example.com',
      granted_permission_sets: { 'ps-id-1': ['svc-1'] },
      valid_until: null,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    const mockResult: GrantResult = { kind: 'created', grant: mockGrant };
    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue(mockResult);

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set permission sets
    act(() => {
      result.current.setGrantedPermissionSets(
        mockGrant.granted_permission_sets,
      );
    });

    // Submit
    let returnedResult: GrantResult | undefined;
    await act(async () => {
      returnedResult = await result.current.submit();
    });

    expect(returnedResult).toEqual(mockResult);
    expect(result.current.isSuccess).toBe(true);
    expect(result.current.error).toBeNull();
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should handle submission error', async () => {
    const errorResponse = new axios.AxiosError('Bad Request');
    errorResponse.response = { status: 400, data: { message: 'Invalid request' } } as never;

    vi.mocked(consentApi.createOrUpdateGrant).mockRejectedValue(errorResponse);

    const { result } = renderHook(() => useToggleGrant(agentId));

    act(() => {
      result.current.setGrantedPermissionSets({ 'ps-id-1': ['svc-1'] });
    });

    // Submit — expect a throw since submit() now re-throws on API errors.
    let caughtError: Error | undefined;
    await act(async () => {
      try {
        await result.current.submit();
      } catch (err) {
        caughtError = err as Error;
      }
    });

    expect(caughtError).toBeInstanceOf(Error);
    expect(caughtError?.message).toBe('Invalid request');
    expect(result.current.isSuccess).toBe(false);
    expect(result.current.error).toBe('Invalid request');
    expect(result.current.isSubmitting).toBe(false);
  });

  it('should handle validation errors', async () => {
    const errorResponse = new axios.AxiosError('Unprocessable Entity');
    errorResponse.response = {
      status: 422,
      data: {
        message: 'Validation failed',
        details: {
          granted_permission_sets: ['At least one entry required'],
        },
      },
    } as never;

    vi.mocked(consentApi.createOrUpdateGrant).mockRejectedValue(errorResponse);

    const { result } = renderHook(() => useToggleGrant(agentId));

    act(() => {
      result.current.setGrantedPermissionSets({ 'ps-id-1': ['svc-1'] });
    });

    let thrownError: Error | null = null;
    await act(async () => {
      try {
        await result.current.submit();
      } catch (err) {
        thrownError = err as Error;
      }
    });

    expect(thrownError).not.toBeNull();
    expect(result.current.error).toContain('Validation error');
    expect(result.current.error).toContain('granted_permission_sets');
  });

  it('uses caller-supplied grantedPermissionSets without prior state update', async () => {
    const mockGrant: UserGrant = {
      id: 'grant-123',
      agent_id: agentId,
      principal: 'user@example.com',
      granted_permission_sets: { 'ps-direct': ['svc-x'] },
      valid_until: null,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue({ kind: 'created', grant: mockGrant });

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Do NOT call setGrantedPermissionSets — pass directly to submit
    await act(async () => {
      await result.current.submit(null, undefined, { 'ps-direct': ['svc-x'] });
    });

    expect(consentApi.createOrUpdateGrant).toHaveBeenCalledWith(
      agentId,
      expect.objectContaining({
        granted_permission_sets: { 'ps-direct': ['svc-x'] },
      }),
      undefined,
    );
    expect(result.current.isSuccess).toBe(true);
  });

  it('should reset state', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set some state
    act(() => {
      result.current.setGrantedPermissionSets({ 'ps-id-1': ['svc-1'] });
    });

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.grantedPermissionSets).toEqual({});
    expect(result.current.error).toBeNull();
    expect(result.current.isSuccess).toBe(false);
  });

  it('should clear error', () => {
    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set some state
    act(() => {
      result.current.setGrantedPermissionSets({ 'ps-id-1': ['svc-1'] });
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
      granted_permission_sets: { 'ps-id-1': ['svc-1'] },
      valid_until: '2025-01-01T00:00:00Z',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    };

    vi.mocked(consentApi.createOrUpdateGrant).mockResolvedValue({ kind: 'created', grant: mockGrant });

    const { result } = renderHook(() => useToggleGrant(agentId));

    // Set permission sets
    act(() => {
      result.current.setGrantedPermissionSets(
        mockGrant.granted_permission_sets,
      );
    });

    // Submit with expiration
    await act(async () => {
      await result.current.submit('2025-01-01T00:00:00Z');
    });

    // Verify the API was called
    expect(consentApi.createOrUpdateGrant).toHaveBeenCalled();
  });
});
