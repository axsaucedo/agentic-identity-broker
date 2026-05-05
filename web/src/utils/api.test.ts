import { describe, it, expect } from 'vitest';
import axios from 'axios';
import { extractApiError } from './api';

describe('extractApiError', () => {
  it('returns response.data.message from AxiosError', () => {
    const err = new axios.AxiosError('request failed');
    err.response = { data: { message: 'Not authorized' }, status: 403 } as never;
    expect(extractApiError(err, 'fallback')).toBe('Not authorized');
  });

  it('returns fallback when AxiosError has no message in response', () => {
    const err = new axios.AxiosError('request failed');
    err.response = { data: {}, status: 500 } as never;
    expect(extractApiError(err, 'default error')).toBe('default error');
  });

  it('returns err.message for plain Error', () => {
    const err = new Error('something broke');
    expect(extractApiError(err, 'fallback')).toBe('something broke');
  });

  it('returns fallback for unrecognized value', () => {
    expect(extractApiError('not an error', 'fallback')).toBe('fallback');
    expect(extractApiError(null, 'fallback')).toBe('fallback');
    expect(extractApiError(42, 'fallback')).toBe('fallback');
  });
});
