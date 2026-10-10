export const stickyCellClasses = {
  left: 'sticky left-0 z-1 bg-[var(--table-row-bg,var(--color-base))] before:pointer-events-none before:absolute before:inset-y-0 before:-right-6 before:w-6 before:bg-[linear-gradient(to_right,var(--table-row-bg,var(--color-base)),transparent)] before:content-empty',
  right: 'sticky right-0 z-1 bg-[var(--table-row-bg,var(--color-base))] before:pointer-events-none before:absolute before:inset-y-0 before:-left-6 before:w-6 before:bg-[linear-gradient(to_left,var(--table-row-bg,var(--color-base)),transparent)] before:content-empty',
} as const
export const stickyHeadClasses = {
  left: 'sticky left-0 z-2 before:pointer-events-none before:absolute before:inset-y-0 before:-right-6 before:w-6 before:bg-[linear-gradient(to_right,var(--color-base),transparent)] before:content-empty',
  right: 'sticky right-0 z-2 before:pointer-events-none before:absolute before:inset-y-0 before:-left-6 before:w-6 before:bg-[linear-gradient(to_left,var(--color-base),transparent)] before:content-empty',
} as const
