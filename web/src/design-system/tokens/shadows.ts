/**
 * Design System Shadow Tokens
 * Premium elevation system for the "Refined Trust Architecture"
 */

export const shadows = {
  // Premium subtle shadows (multi-layer with inset highlight)
  'sm-premium': '0 1px 2px 0 rgb(0 0 0 / 0.03)',
  'md-premium': '0 4px 12px 0 rgb(0 0 0 / 0.08)',
  'lg-premium': '0 10px 25px -5px rgb(0 0 0 / 0.1)',
  'xl-premium': '0 20px 40px -10px rgb(0 0 0 / 0.12)',

  // Elevated card styling
  card: '0 2px 8px 0 rgb(0 0 0 / 0.04), inset 0 1px 0 0 rgb(255 255 255 / 1)',
  'card-hover':
    '0 8px 20px -4px rgb(0 0 0 / 0.1), inset 0 1px 0 0 rgb(255 255 255 / 1)',

  // Standard elevation scale
  none: 'none',
  sm: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
  base: '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)',
  md: '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)',
  lg: '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)',
  xl: '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)',
  '2xl': '0 25px 50px -12px rgb(0 0 0 / 0.25)',

  // Focus rings
  'focus-ring': '0 0 0 3px rgb(30 77 107 / 0.5)',
  'focus-ring-error': '0 0 0 3px rgb(220 38 38 / 0.5)',
} as const;

/**
 * Semantic shadow mapping for common use cases
 */
export const elevationLevels = {
  flat: shadows.none,
  raised: shadows['sm-premium'],
  floating: shadows['md-premium'],
  overlay: shadows['lg-premium'],
  modal: shadows['xl-premium'],
} as const;

export type Shadows = typeof shadows;
export type ElevationLevels = typeof elevationLevels;
