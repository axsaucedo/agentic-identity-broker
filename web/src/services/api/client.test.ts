import { describe, expect, it } from 'vitest';
import { apiClient, isApiError } from './client';

describe('apiClient', () => {
  it('should be configured with /api base URL', () => {
    expect(apiClient.defaults.baseURL).toBe('/api');
  });

  it('should have 30s timeout', () => {
    expect(apiClient.defaults.timeout).toBe(30000);
  });
});

describe('isApiError', () => {
  it('returns true for valid ApiError objects', () => {
    expect(isApiError({ status: 404, code: 'NOT_FOUND', message: 'not found' })).toBe(true);
  });

  it('returns false for non-ApiError values', () => {
    expect(isApiError(null)).toBe(false);
    expect(isApiError('string')).toBe(false);
    expect(isApiError({ status: 404 })).toBe(false);
  });
});
