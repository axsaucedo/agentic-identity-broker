/**
 * Design System Z-Index Tokens
 * Stacked layer system for overlays and modals
 */

export const zIndex = {
  base: 0,
  dropdown: 1000,
  sticky: 1100,
  fixed: 1200,
  overlay: 1300,
  modal: 1400,
  popover: 1500,
  tooltip: 1600,
  toast: 1700,
} as const;

export type ZIndex = typeof zIndex;
