/**
 * Session storage utilities for managing authentication tokens.
 *
 * These utilities handle storing and retrieving session tokens
 * from browser storage (sessionStorage by default, with localStorage fallback).
 */

const SESSION_TOKEN_KEY = 'aib_session_token';
const USE_SESSION_STORAGE = true; // Set to false to use localStorage

/**
 * Get the session token from storage.
 * Checks sessionStorage first, then falls back to localStorage.
 *
 * @returns Session token string, or null if not found
 */
export function getSessionToken(): string | null {
  try {
    // Try sessionStorage first
    if (USE_SESSION_STORAGE && typeof sessionStorage !== 'undefined') {
      const token = sessionStorage.getItem(SESSION_TOKEN_KEY);
      if (token) return token;
    }

    // Fallback to localStorage
    if (typeof localStorage !== 'undefined') {
      return localStorage.getItem(SESSION_TOKEN_KEY);
    }

    return null;
  } catch (error) {
    console.error('Error reading session token:', error);
    return null;
  }
}

/**
 * Store the session token in storage.
 * Stores in sessionStorage by default, or localStorage if configured.
 *
 * @param token - Session token to store
 */
export function setSessionToken(token: string): void {
  try {
    if (USE_SESSION_STORAGE && typeof sessionStorage !== 'undefined') {
      sessionStorage.setItem(SESSION_TOKEN_KEY, token);
    } else if (typeof localStorage !== 'undefined') {
      localStorage.setItem(SESSION_TOKEN_KEY, token);
    } else {
      console.warn('No storage mechanism available');
    }
  } catch (error) {
    console.error('Error storing session token:', error);
  }
}

/**
 * Clear the session token from both sessionStorage and localStorage.
 * Used during logout or session expiration.
 */
export function clearSessionToken(): void {
  try {
    if (typeof sessionStorage !== 'undefined') {
      sessionStorage.removeItem(SESSION_TOKEN_KEY);
    }
    if (typeof localStorage !== 'undefined') {
      localStorage.removeItem(SESSION_TOKEN_KEY);
    }
  } catch (error) {
    console.error('Error clearing session token:', error);
  }
}

/**
 * Check if a session token exists in storage.
 *
 * @returns true if a token exists, false otherwise
 */
export function hasSessionToken(): boolean {
  return getSessionToken() !== null;
}
