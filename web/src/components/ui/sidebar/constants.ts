export const SIDEBAR_STYLING = {
  width: { expanded: '16.25rem', icon: '57px' },
  mobile: { breakpoint: 768 },
} as const

export const SIDEBAR_EASING = 'cubic-bezier(0.77, 0, 0.175, 1)'
export const SIDEBAR_ANIMATION_DURATION_MS = 250
export const SIDEBAR_DEFAULT_WIDTH = 256
export const SIDEBAR_MIN_WIDTH = 200
export const SIDEBAR_MAX_WIDTH = 480
export const SIDEBAR_FOCUSABLE_SELECTOR = 'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
