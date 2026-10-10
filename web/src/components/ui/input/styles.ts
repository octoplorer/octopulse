export type InputSize = 'xs' | 'sm' | 'base' | 'lg'

export const inputSizes = {
  xs: 'h-5 px-1.5 rounded-sm text-size-xs max-sm:h-7 max-sm:text-16px',
  sm: 'h-6.5 px-2 rounded-md text-size-xs max-sm:h-8 max-sm:text-16px',
  base: 'h-9 px-3 rounded-lg text-size-base max-sm:h-10 max-sm:text-16px',
  lg: 'h-10 px-4 rounded-lg text-size-base max-sm:text-16px',
} satisfies Record<InputSize, string>

export const inputClasses = 'ui-input w-full min-w-0 border-0 bg-control text-default shadow-none ring-1 ring-control-border outline-none focus:outline-none placeholder:text-placeholder disabled:opacity-50 disabled:cursor-not-allowed focus:ring-2 focus:ring-focus data-[invalid]:ring-action-danger data-[invalid]:focus:ring-action-danger'
