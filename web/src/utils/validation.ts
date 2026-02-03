/**
 * Validation utilities for grant requests.
 *
 * Provides validation functions for:
 * - Grant requests
 * - Service IDs
 * - Scopes
 * - Expiration dates
 */

import type { DelegatedToken } from '../types/consent';

/**
 * Validates a grant creation/update request.
 * Returns array of error messages (empty if valid).
 * Works with internal frontend format (camelCase).
 *
 * @param request - Grant request to validate with internal format
 * @returns Array of validation error messages
 */
export function validateGrantRequest(request: {
  delegatedTokens: DelegatedToken[];
  validUntil?: string | null;
}): string[] {
  const errors: string[] = [];

  // Must have delegatedTokens field
  if (!request.delegatedTokens) {
    errors.push('delegatedTokens field is required');
    return errors;
  }

  // Must have at least one delegated token
  if (request.delegatedTokens.length === 0) {
    errors.push('Please select at least one service with scopes');
    return errors;
  }

  // Validate each delegated token
  request.delegatedTokens.forEach((token, index) => {
    // Service ID is required
    if (!token.thirdparty_oauth2_service_id) {
      errors.push(
        `Service ${index + 1}: thirdparty_oauth2_service_id is required`,
      );
    } else if (!validateServiceId(token.thirdparty_oauth2_service_id)) {
      errors.push(
        `Service ${index + 1}: invalid thirdparty_oauth2_service_id format`,
      );
    }

    // Scopes are required
    if (!token.scopes) {
      errors.push(`Service ${index + 1}: scopes field is required`);
    } else if (token.scopes.length === 0) {
      errors.push(`Service ${index + 1}: at least one scope must be selected`);
    } else if (!validateScopes(token.scopes)) {
      errors.push(`Service ${index + 1}: invalid scope format`);
    }
  });

  // Validate expiration date if provided
  if (request.validUntil) {
    const validationError = validateExpirationDate(
      new Date(request.validUntil),
    );
    if (validationError) {
      errors.push(validationError);
    }
  }

  return errors;
}

/**
 * Validates that a service ID is in valid format.
 *
 * @param serviceId - Service ID to validate
 * @returns True if valid, false otherwise
 */
export function validateServiceId(serviceId: string): boolean {
  // Service ID must be non-empty string
  if (!serviceId || typeof serviceId !== 'string') {
    return false;
  }

  // Service ID should not contain whitespace or special characters
  // Allow alphanumeric, hyphens, underscores, dots
  const serviceIdPattern = /^[a-zA-Z0-9._-]+$/;
  return serviceIdPattern.test(serviceId);
}

/**
 * Validates an array of OAuth scopes.
 *
 * @param scopes - Array of scope strings to validate
 * @returns True if all scopes are valid, false otherwise
 */
export function validateScopes(scopes: string[]): boolean {
  if (!Array.isArray(scopes)) {
    return false;
  }

  // Each scope must be non-empty string
  return scopes.every((scope) => {
    if (!scope || typeof scope !== 'string') {
      return false;
    }

    // Scope should not be just whitespace
    if (scope.trim().length === 0) {
      return false;
    }

    return true;
  });
}

/**
 * Validates an expiration date.
 *
 * @param date - Date to validate
 * @returns Error message if invalid, null if valid
 */
export function validateExpirationDate(date: Date): string | null {
  // Must be a valid date
  if (!(date instanceof Date) || isNaN(date.getTime())) {
    return 'Invalid date format';
  }

  // Must be in the future
  const now = new Date();
  if (date <= now) {
    return 'Expiration date must be in the future';
  }

  // Should not be more than 10 years in the future (sanity check)
  const tenYearsFromNow = new Date();
  tenYearsFromNow.setFullYear(tenYearsFromNow.getFullYear() + 10);
  if (date > tenYearsFromNow) {
    return 'Expiration date cannot be more than 10 years in the future';
  }

  return null;
}

/**
 * Formats validation errors for display.
 *
 * @param errors - Array of error messages
 * @returns Formatted error string
 */
export function formatValidationErrors(errors: string[]): string {
  if (errors.length === 0) {
    return '';
  }

  if (errors.length === 1) {
    return errors[0];
  }

  // Multiple errors: format as numbered list
  return errors.map((error, index) => `${index + 1}. ${error}`).join('\n');
}

/**
 * Validates that a URL is safe for redirection (same-origin or relative).
 *
 * This provides defense-in-depth protection against open redirect vulnerabilities.
 * While the backend performs the primary validation (SR-003), frontend validation
 * adds an additional security layer.
 *
 * @param redirectUrl - URL to validate
 * @returns True if URL is safe (relative or same-origin), false otherwise
 */
export function isSafeRedirectUrl(redirectUrl: string): boolean {
  if (!redirectUrl) {
    return false;
  }

  // Check for obviously malicious or dangerous schemes first
  const dangerousSchemes = [
    'javascript:',
    'data:',
    'vbscript:',
    'file:',
    'about:',
    'blob:',
  ];
  const lowerUrl = redirectUrl.toLowerCase();
  if (dangerousSchemes.some((scheme) => lowerUrl.startsWith(scheme))) {
    return false;
  }

  // Check for protocol-relative URLs (e.g., "//evil.com")
  if (redirectUrl.startsWith('//')) {
    return false;
  }

  try {
    // Try to parse as absolute URL
    const url = new URL(redirectUrl);

    // It's an absolute URL - check if it's same-origin
    // Normalize ports: http default 80, https default 443
    const normalizePort = (protocol: string, port: string): string => {
      if (port) return port;
      return protocol === 'https:' ? '443' : '80';
    };

    const urlPort = normalizePort(url.protocol, url.port);
    const locationPort = normalizePort(
      window.location.protocol,
      window.location.port,
    );

    return (
      url.protocol === window.location.protocol &&
      url.hostname === window.location.hostname &&
      urlPort === locationPort
    );
  } catch {
    // Not a valid absolute URL - treat as relative URL
    // Relative URLs are safe (e.g., "/consent", "consent", "./consent")
    // But double-check it's not trying to be a URL with protocol
    // Use regex to match protocol pattern: scheme followed by colon
    // (e.g., "http:", "https:", "ftp:", "javascript:")
    if (redirectUrl.match(/^[a-zA-Z][a-zA-Z0-9+.-]*:/)) {
      return false;
    }
    return true;
  }
}
