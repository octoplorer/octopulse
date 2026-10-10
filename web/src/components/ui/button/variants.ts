import { cva } from 'cva'

export const buttonVariants = cva({
  base: 'button inline-flex shrink-0 items-center justify-center font-medium whitespace-nowrap select-none border-0 outline-none transition-colors disabled:cursor-not-allowed disabled:opacity-50 focus-visible:relative focus-visible:z-1 focus:outline-none focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand',
  variants: {
    variant: {
      'primary': 'button-emphasis ring-1 text-on-brand shadow-control',
      'destructive': 'button-emphasis ring-1 text-on-danger shadow-control',
      'secondary': 'bg-base text-default ring-1 ring-line shadow-control not-disabled:hover:bg-tint',
      'secondary-destructive': 'bg-base text-fg-danger ring-1 ring-line shadow-control not-disabled:hover:bg-danger-tint',
      'ghost': 'bg-transparent text-default not-disabled:hover:bg-tint',
      'outline': 'bg-transparent text-default ring-1 ring-line not-disabled:hover:bg-tint',
    },
    size: { xs: 'h-5 gap-1 text-size-xs', sm: 'h-6.5 gap-1 text-size-xs', base: 'h-9 gap-1.5 text-size-base', lg: 'h-10 gap-2 text-size-base' },
    shape: { base: '', square: 'px-0', circle: 'rounded-full px-0' },
  },
  compoundVariants: [
    { size: 'xs', shape: 'base', class: 'px-1.5' },
    { size: 'sm', shape: 'base', class: 'px-2' },
    { size: 'base', shape: 'base', class: 'px-3' },
    { size: 'lg', shape: 'base', class: 'px-4' },
    { size: 'xs', shape: ['base', 'square'], class: 'rounded-sm' },
    { size: 'sm', shape: ['base', 'square'], class: 'rounded-md' },
    { size: ['base', 'lg'], shape: ['base', 'square'], class: 'rounded-lg' },
    { size: 'xs', shape: ['square', 'circle'], class: 'w-5' },
    { size: 'sm', shape: ['square', 'circle'], class: 'w-6.5' },
    { size: 'base', shape: ['square', 'circle'], class: 'w-9' },
    { size: 'lg', shape: ['square', 'circle'], class: 'w-10' },
  ],
  defaultVariants: { variant: 'secondary', size: 'base', shape: 'base' },
})
