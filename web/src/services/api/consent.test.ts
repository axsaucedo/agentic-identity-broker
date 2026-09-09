import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ConsentApiService } from './consent';
import type { CreateOrUpdateGrantRequest, UserGrant } from '../../types/consent';

vi.mock('./client', () => ({
  apiClient: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

vi.mock('./cache', () => ({
  apiCache: {
    get: vi.fn().mockReturnValue(null),
    set: vi.fn(),
    invalidate: vi.fn(),
    invalidatePattern: vi.fn(),
  },
}));

import { apiClient } from './client';

const mockGrant: UserGrant = {
  id: 'grant-1',
  agent_id: 'agent-1',
  principal: 'user@example.com',
  delegated_oauth2_tokens: [{ thirdparty_oauth2_service_id: 'svc-1', scopes: ['read'] }],
  valid_until: null,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const request: CreateOrUpdateGrantRequest = {
  delegated_oauth2_tokens: [{ thirdparty_oauth2_service_id: 'svc-1', scopes: ['read'] }],
};

describe('ConsentApiService.createOrUpdateGrant', () => {
  let service: ConsentApiService;
  let originalLocation: Location;

  beforeEach(() => {
    service = new ConsentApiService();
    vi.clearAllMocks();

    originalLocation = window.location;
    Object.defineProperty(window, 'location', {
      value: { ...originalLocation, href: '' },
      writable: true,
      configurable: true,
    });
  });

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  it('returns { kind: "created", grant } on 201 without redirect_url', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({
      status: 201,
      data: { data: mockGrant },
    });

    const result = await service.createOrUpdateGrant('agent-1', request);

    expect(result).toEqual({ kind: 'created', grant: mockGrant });
    expect(window.location.href).toBe('');
  });

  it('returns { kind: "redirect", redirectUrl } on 201 with redirect_url and does not navigate', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({
      status: 201,
      data: { data: mockGrant, redirect_url: '/agents/agent-1/done' },
    });

    const result = await service.createOrUpdateGrant('agent-1', request);

    expect(result).toEqual({ kind: 'redirect', redirectUrl: '/agents/agent-1/done' });
    expect(window.location.href).toBe('');
  });

  it('returns { kind: "noContent" } on 204', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({
      status: 204,
      data: {},
    });

    const result = await service.createOrUpdateGrant('agent-1', request);

    expect(result).toEqual({ kind: 'noContent' });
  });

  it('throws on 201 with an unsafe redirect_url', async () => {
    vi.mocked(apiClient.post).mockResolvedValue({
      status: 201,
      data: { data: mockGrant, redirect_url: 'https://evil.example.com/steal' },
    });

    await expect(service.createOrUpdateGrant('agent-1', request)).rejects.toThrow(
      'Redirect URL validation failed',
    );
    expect(window.location.href).toBe('');
  });
});
