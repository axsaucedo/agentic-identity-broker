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
      errors.push(`Service ${index + 1}: thirdparty_oauth2_service_id is required`);
    } else if (!validateServiceId(token.thirdparty_oauth2_service_id)) {
      errors.push(`Service ${index + 1}: invalid thirdparty_oauth2_service_id format`);
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
    const validationError = validateExpirationDate(new Date(request.validUntil));
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
