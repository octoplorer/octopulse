export type ChartSemanticColorName = 'Attention' | 'Warning' | 'Success' | 'Neutral' | 'Disabled' | 'Skeleton'

const categoricalColors = ['var(--color-brand)', 'var(--color-warning)', 'var(--color-badge-purple)', 'var(--color-badge-blue)', 'var(--color-badge-teal)', 'var(--color-badge-orange)']
const semanticColors: Record<ChartSemanticColorName, string> = {
  Attention: 'var(--color-danger)',
  Warning: 'var(--color-warning)',
  Success: 'var(--color-success)',
  Neutral: 'var(--color-info)',
  Disabled: 'var(--color-interact)',
  Skeleton: 'var(--color-fill)',
}

export const ChartPalette = {
  categorical: (index: number) => categoricalColors[((Math.trunc(index) % categoricalColors.length) + categoricalColors.length) % categoricalColors.length]!,
  semantic: (name: ChartSemanticColorName) => semanticColors[name],
  sequential: (_palette: 'blues' = 'blues') => [15, 35, 55, 75, 100].map(weight => `color-mix(in oklch, var(--color-brand) ${weight}%, var(--color-base))`),
  text: (variant: 'primary' | 'secondary') => variant === 'primary' ? 'var(--text-color-default)' : 'var(--text-color-subtle)',
  mapColors: () => ({ area: 'var(--color-fill)', bubble: categoricalColors[0]!, scale: ChartPalette.sequential() }),
}
