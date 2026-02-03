/**
 * Design System Animation Tokens
 * Timing, easing, and keyframes for "motion that guides, not entertains"
 */

export const duration = {
  instant: '0ms',
  fast: '150ms', // Hover color transitions, focus ring appearance
  base: '200ms', // Button state changes, dropdown open/close
  slow: '300ms', // Card elevation changes, modal overlays
  slower: '500ms', // Page transitions, full-screen loading states
} as const;

export const easing = {
  linear: 'linear',
  ease: 'ease',
  easeIn: 'ease-in',
  easeOut: 'ease-out',
  easeInOut: 'ease-in-out',
  spring: 'cubic-bezier(0.34, 1.56, 0.64, 1)', // Spring feeling for important actions
  smooth: 'cubic-bezier(0.16, 1, 0.3, 1)', // Smooth, premium transitions
} as const;

export const keyframes = {
  fadeIn: {
    '0%': { opacity: '0' },
    '100%': { opacity: '1' },
  },
  fadeOut: {
    '0%': { opacity: '1' },
    '100%': { opacity: '0' },
  },
  slideUp: {
    '0%': { transform: 'translateY(12px)', opacity: '0' },
    '100%': { transform: 'translateY(0)', opacity: '1' },
  },
  slideDown: {
    '0%': { transform: 'translateY(-12px)', opacity: '0' },
    '100%': { transform: 'translateY(0)', opacity: '1' },
  },
  scaleIn: {
    '0%': { transform: 'scale(0.95)', opacity: '0' },
    '100%': { transform: 'scale(1)', opacity: '1' },
  },
  scaleOut: {
    '0%': { transform: 'scale(1)', opacity: '1' },
    '100%': { transform: 'scale(0.95)', opacity: '0' },
  },
  spin: {
    '0%': { transform: 'rotate(0deg)' },
    '100%': { transform: 'rotate(360deg)' },
  },
} as const;

export type Duration = typeof duration;
export type Easing = typeof easing;
export type Keyframes = typeof keyframes;
