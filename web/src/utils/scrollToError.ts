/**
 * Utility function to smoothly scroll to the first error field.
 *
 * Features:
 * - Smooth scroll behavior
 * - Focus on error field for accessibility
 * - Configurable offset for fixed headers
 * - Works with any error container or input
 *
 * Usage:
 * ```typescript
 * import { scrollToError } from '@utils/scrollToError';
 *
 * // Scroll to first element with [data-error] attribute
 * scrollToError();
 *
 * // Scroll to specific selector
 * scrollToError('[aria-invalid="true"]');
 * ```
 */

interface ScrollToErrorOptions {
  /** CSS selector for error elements (default: '[data-error="true"]') */
  selector?: string;
  /** Offset from top in pixels (useful for fixed headers) */
  offset?: number;
  /** Scroll behavior ('smooth' | 'auto') */
  behavior?: ScrollBehavior;
  /** Whether to focus the element after scrolling */
  focus?: boolean;
}

/**
 * Scroll to the first error element on the page.
 * Useful for form validation to bring user attention to errors.
 *
 * @param options - Configuration options
 * @returns True if an error element was found and scrolled to, false otherwise
 */
export function scrollToError(options: ScrollToErrorOptions = {}): boolean {
  const {
    selector = '[data-error="true"], [aria-invalid="true"], .error',
    offset = 80,
    behavior = 'smooth',
    focus = true,
  } = options;

  // Find the first error element
  const errorElement = document.querySelector<HTMLElement>(selector);

  if (!errorElement) {
    return false;
  }

  // Calculate scroll position with offset
  const elementPosition = errorElement.getBoundingClientRect().top;
  const offsetPosition = elementPosition + window.pageYOffset - offset;

  // Scroll to element
  window.scrollTo({
    top: offsetPosition,
    behavior,
  });

  // Focus the element for accessibility (if it's focusable)
  if (focus) {
    setTimeout(() => {
      if (errorElement.tabIndex >= 0 || isNaturallyFocusable(errorElement)) {
        errorElement.focus({ preventScroll: true });
      } else {
        // Make element focusable if it's not naturally
        errorElement.tabIndex = -1;
        errorElement.focus({ preventScroll: true });
      }
    }, 100); // Small delay to ensure scroll completes first
  }

  return true;
}

/**
 * Check if an element is naturally focusable (input, button, etc.).
 */
function isNaturallyFocusable(element: HTMLElement): boolean {
  const focusableTags = ['INPUT', 'TEXTAREA', 'SELECT', 'BUTTON', 'A'];
  return focusableTags.includes(element.tagName);
}

/**
 * Scroll to a specific element by ID or selector.
 *
 * @param elementOrSelector - Element or CSS selector
 * @param options - Configuration options
 * @returns True if element was found and scrolled to, false otherwise
 */
export function scrollToElement(
  elementOrSelector: HTMLElement | string,
  options: Omit<ScrollToErrorOptions, 'selector'> = {}
): boolean {
  const element =
    typeof elementOrSelector === 'string'
      ? document.querySelector<HTMLElement>(elementOrSelector)
      : elementOrSelector;

  if (!element) {
    return false;
  }

  const { offset = 80, behavior = 'smooth', focus = false } = options;

  // Calculate scroll position with offset
  const elementPosition = element.getBoundingClientRect().top;
  const offsetPosition = elementPosition + window.pageYOffset - offset;

  // Scroll to element
  window.scrollTo({
    top: offsetPosition,
    behavior,
  });

  // Focus if requested
  if (focus) {
    setTimeout(() => {
      if (element.tabIndex >= 0 || isNaturallyFocusable(element)) {
        element.focus({ preventScroll: true });
      }
    }, 100);
  }

  return true;
}

/**
 * Hook to create a scroll-to-error callback function.
 * Useful in React components.
 *
 * @example
 * ```tsx
 * import { useScrollToError } from '@utils/scrollToError';
 *
 * function MyForm() {
 *   const scrollToError = useScrollToError();
 *
 *   const handleSubmit = async () => {
 *     try {
 *       await validateForm();
 *     } catch (error) {
 *       scrollToError(); // Scroll to first error
 *     }
 *   };
 * }
 * ```
 */
export function useScrollToError(options: ScrollToErrorOptions = {}) {
  return () => scrollToError(options);
}

export default scrollToError;
