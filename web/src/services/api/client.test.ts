import { AxiosHeaders, type InternalAxiosRequestConfig } from 'axios';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { apiClient } from './client';

describe('apiClient CSRF request interceptor', () => {
  let cookieValue = '';
  let originalCookieDescriptor: PropertyDescriptor | undefined;

  beforeEach(() => {
    cookieValue = '';
    originalCookieDescriptor = Object.getOwnPropertyDescriptor(document, 'cookie');

    Object.defineProperty(document, 'cookie', {
      configurable: true,
      get: () => cookieValue,
      set: (value: string) => {
        cookieValue = value;
      },
    });
  });

  afterEach(() => {
    if (originalCookieDescriptor) {
      Object.defineProperty(document, 'cookie', originalCookieDescriptor);
      return;
    }

    delete (document as Document & { cookie?: string }).cookie;
  });

  async function runRequestInterceptor(method: string) {
    const interceptor = apiClient.interceptors.request.handlers?.[0];

    if (!interceptor) {
      throw new Error('Expected request interceptor to be registered');
    }

    return interceptor.fulfilled({
      headers: new AxiosHeaders(),
      method,
      url: '/consent/agent/agent-1/grant',
    } as InternalAxiosRequestConfig);
  }

  it.each(['POST', 'PUT', 'PATCH', 'DELETE'])(
    'sets X-CSRF-Token header for %s requests',
    async (method) => {
      document.cookie = 'csrf_token=token%20value';

      const config = await runRequestInterceptor(method);

      expect(config.headers.get('X-CSRF-Token')).toBe('token value');
    },
  );

  it.each(['GET', 'HEAD', 'OPTIONS'])(
    'does not set X-CSRF-Token header for %s requests',
    async (method) => {
      document.cookie = 'csrf_token=token%20value';

      const config = await runRequestInterceptor(method);

      expect(config.headers.has('X-CSRF-Token')).toBe(false);
    },
  );
});
