/**
 * Simple in-memory cache for API responses with TTL.
 *
 * Features:
 * - In-memory storage with timestamp-based TTL
 * - Configurable TTL per cache entry (default 5 minutes)
 * - Automatic cache invalidation on expiry
 * - Cache key generation from URL and params
 * - Manual cache clearing
 * - Typed cache entries
 *
 * Usage:
 * ```typescript
 * import { apiCache } from '@services/api/cache';
 *
 * // Get from cache
 * const cached = apiCache.get('/consent/agents');
 * if (cached) return cached;
 *
 * // Set cache
 * const data = await fetchData();
 * apiCache.set('/consent/agents', data, 300000); // 5 min TTL
 *
 * // Invalidate cache
 * apiCache.invalidate('/consent/agents');
 * ```
 */

interface CacheEntry<T = unknown> {
  /** Cached data */
  data: T;
  /** Timestamp when cached (milliseconds) */
  timestamp: number;
  /** Time-to-live in milliseconds */
  ttl: number;
}

type CacheKey = string;

/**
 * In-memory cache storage.
 * Keys are generated from API endpoints and parameters.
 */
class ApiCache {
  private cache: Map<CacheKey, CacheEntry> = new Map();
  private readonly DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes

  /**
   * Generate a cache key from URL and optional parameters.
   */
  private generateKey(url: string, params?: Record<string, unknown>): CacheKey {
    if (!params || Object.keys(params).length === 0) {
      return url;
    }

    // Sort params for consistent cache keys
    const sortedParams = Object.keys(params)
      .sort()
      .map((key) => `${key}=${JSON.stringify(params[key])}`)
      .join('&');

    return `${url}?${sortedParams}`;
  }

  /**
   * Check if a cache entry is expired.
   */
  private isExpired(entry: CacheEntry): boolean {
    const now = Date.now();
    return now - entry.timestamp > entry.ttl;
  }

  /**
   * Get data from cache if available and not expired.
   *
   * @param url - API endpoint URL
   * @param params - Optional query parameters
   * @returns Cached data or null if not found/expired
   */
  get<T = unknown>(url: string, params?: Record<string, unknown>): T | null {
    const key = this.generateKey(url, params);
    const entry = this.cache.get(key);

    if (!entry) {
      return null;
    }

    // Check if expired
    if (this.isExpired(entry)) {
      this.cache.delete(key);
      return null;
    }

    return entry.data as T;
  }

  /**
   * Set data in cache with TTL.
   *
   * @param url - API endpoint URL
   * @param data - Data to cache
   * @param ttl - Time-to-live in milliseconds (default 5 minutes)
   * @param params - Optional query parameters
   */
  set<T = unknown>(
    url: string,
    data: T,
    ttl: number = this.DEFAULT_TTL,
    params?: Record<string, unknown>
  ): void {
    const key = this.generateKey(url, params);
    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      ttl,
    };

    this.cache.set(key, entry);
  }

  /**
   * Check if a cache entry exists and is valid.
   *
   * @param url - API endpoint URL
   * @param params - Optional query parameters
   * @returns True if cached data exists and is not expired
   */
  has(url: string, params?: Record<string, unknown>): boolean {
    const key = this.generateKey(url, params);
    const entry = this.cache.get(key);

    if (!entry) {
      return false;
    }

    if (this.isExpired(entry)) {
      this.cache.delete(key);
      return false;
    }

    return true;
  }

  /**
   * Invalidate (remove) a specific cache entry.
   *
   * @param url - API endpoint URL
   * @param params - Optional query parameters
   */
  invalidate(url: string, params?: Record<string, unknown>): void {
    const key = this.generateKey(url, params);
    this.cache.delete(key);
  }

  /**
   * Invalidate all cache entries matching a URL pattern.
   *
   * @param pattern - URL pattern to match (supports wildcards with *)
   * @example
   * ```typescript
   * // Invalidate all agent-related caches
   * apiCache.invalidatePattern('/consent/agent/*');
   * ```
   */
  invalidatePattern(pattern: string): void {
    const regex = new RegExp(
      '^' + pattern.replace(/\*/g, '.*').replace(/\?/g, '\\?') + '$'
    );

    for (const key of this.cache.keys()) {
      if (regex.test(key)) {
        this.cache.delete(key);
      }
    }
  }

  /**
   * Clear all cache entries.
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Get current cache size (number of entries).
   */
  size(): number {
    return this.cache.size;
  }

  /**
   * Clean up expired cache entries.
   * Call this periodically to free memory.
   */
  cleanup(): void {
    for (const [key, entry] of this.cache.entries()) {
      if (this.isExpired(entry)) {
        this.cache.delete(key);
      }
    }
  }

  /**
   * Get cache statistics for debugging.
   */
  getStats(): {
    totalEntries: number;
    expiredEntries: number;
    validEntries: number;
  } {
    let expired = 0;
    let valid = 0;

    for (const entry of this.cache.values()) {
      if (this.isExpired(entry)) {
        expired++;
      } else {
        valid++;
      }
    }

    return {
      totalEntries: this.cache.size,
      expiredEntries: expired,
      validEntries: valid,
    };
  }
}

/**
 * Singleton instance of the API cache.
 * Use this instance throughout the application.
 */
export const apiCache = new ApiCache();

// Cleanup expired entries every 2 minutes
if (typeof window !== 'undefined') {
  setInterval(() => {
    apiCache.cleanup();
  }, 2 * 60 * 1000);
}

export default apiCache;
