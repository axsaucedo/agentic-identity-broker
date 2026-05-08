import { isAxiosError } from 'axios';

export function extractApiError(err: unknown, fallback: string): string {
  if (isAxiosError(err)) {
    const msg = err.response?.data?.message;
    return typeof msg === 'string' ? msg : fallback;
  }
  if (err instanceof Error) return err.message;
  return fallback;
}
