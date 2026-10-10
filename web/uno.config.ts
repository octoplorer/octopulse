import { defineConfig, escapeSelector, presetAttributify, presetIcons, presetWind4 } from 'unocss'

export default defineConfig({
  content: {
    pipeline: {
      include: [
        /\.(vue|svelte|[jt]sx|vine\.ts|mdx?|astro|elm|php|phtml|marko|html)($|\?)/,
        /\/src\/.*\.[jt]s($|\?)/,
      ],
      exclude: [/\/(node_modules|dist)\//],
    },
    filesystem: ['src/components/ui/**/*.{vue,ts}', '!src/components/ui/**/*.test.ts'],
  },
  variants: [
    {
      name: 'group-not-data',
      multiPass: true,
      match(matcher) {
        const match = matcher.match(/^group-not-data-\[((?:\\.|[^\]\\])+)\](?:\/([^:\s]+))?:/)
        if (!match)
          return

        const [, attribute, groupName] = match
        const group = escapeSelector(groupName ? `group/${groupName}` : 'group')

        return {
          matcher: matcher.slice(match[0].length),
          // Match Wind4's tagged-data selector structure and specificity.
          handle: (input, next) => next({
            ...input,
            parent: `${input.parent ? `${input.parent} $$ ` : ''}${input.selector}`,
            selector: `&:is(:where(.${group}):not([data-${attribute}]) *)`,
          }),
        }
      },
    },
  ],
  theme: {
    text: {
      xs: { fontSize: '12px', lineHeight: 'calc(1 / 0.75)' },
      sm: { fontSize: '13px', lineHeight: '1.5' },
      base: { fontSize: '14px', lineHeight: '1.5' },
      lg: { fontSize: '16px', lineHeight: '1.5' },
    },
    shadow: {
      control: 'var(--shadow-control)',
      panel: 'var(--shadow-panel)',
    },
    colors: {
      'default': 'var(--text-color-default)',
      'inverse': 'var(--text-color-inverse)',
      'strong': 'var(--text-color-strong)',
      'subtle': 'var(--text-color-subtle)',
      'inactive': 'var(--text-color-inactive)',
      'placeholder': 'var(--text-color-placeholder)',
      'link': 'var(--text-color-link)',
      'on-brand': 'var(--text-color-on-brand)',
      'on-danger': 'var(--text-color-on-danger)',
      // Foreground status colors are separate from their solid surface colors.
      'fg': {
        'brand': 'var(--text-color-brand)',
        'info': 'var(--text-color-info)',
        'success': 'var(--text-color-success)',
        'danger': 'var(--text-color-danger)',
        'warning': 'var(--text-color-warning)',
        'badge-orange-subtle': 'var(--text-color-badge-orange-subtle)',
        'badge-teal-subtle': 'var(--text-color-badge-teal-subtle)',
        'badge-neutral-subtle': 'var(--text-color-badge-neutral-subtle)',
        'badge-inverted': 'var(--text-color-badge-inverted)',
      },
      'canvas': 'var(--color-canvas)',
      'elevated': 'var(--color-elevated)',
      'recessed': 'var(--color-recessed)',
      'base': 'var(--color-base)',
      'tint': 'var(--color-tint)',
      'contrast': 'var(--color-contrast)',
      'overlay': 'var(--color-overlay)',
      'control': 'var(--color-control)',
      'control-border': 'var(--color-control-border)',
      'control-thumb': 'var(--color-control-thumb)',
      'backdrop': 'var(--color-backdrop)',
      'interact': 'var(--color-interact)',
      'fill': 'var(--color-fill)',
      'fill-hover': 'var(--color-fill-hover)',
      'brand': 'var(--color-brand)',
      'brand-hover': 'var(--color-brand-hover)',
      'line': 'var(--color-line)',
      'hairline': 'var(--color-hairline)',
      'focus': 'var(--color-focus)',
      'shadow-edge': 'var(--color-shadow-edge)',
      'shadow-drop': 'var(--color-shadow-drop)',
      'arrow-edge': 'var(--color-arrow-edge)',
      'arrow-stroke': 'var(--color-arrow-stroke)',
      'info-tint': 'var(--color-info-tint)',
      'info': 'var(--color-info)',
      'warning-tint': 'var(--color-warning-tint)',
      'warning': 'var(--color-warning)',
      'danger-tint': 'var(--color-danger-tint)',
      'danger': 'var(--color-danger)',
      'action-danger': 'var(--color-action-danger)',
      'success-tint': 'var(--color-success-tint)',
      'success': 'var(--color-success)',
      'banner-info': 'var(--color-banner-info)',
      'banner-warning': 'var(--color-banner-warning)',
      'badge-red': 'var(--color-badge-red)',
      'badge-green': 'var(--color-badge-green)',
      'badge-orange': 'var(--color-badge-orange)',
      'badge-purple': 'var(--color-badge-purple)',
      'badge-teal': 'var(--color-badge-teal)',
      'badge-blue': 'var(--color-badge-blue)',
      'badge-neutral': 'var(--color-badge-neutral)',
      'badge-inverted': 'var(--color-badge-inverted)',
    },
  },
  presets: [
    presetWind4({
      preflights: {
        reset: true,
      },
    }),
    presetAttributify({
      prefix: 'un-',
      prefixedOnly: false,
      // Preserve native SVG presentation attributes.
      ignoreAttributes: ['font-size', 'stroke-width'],
    }),
    presetIcons({
      extraProperties: {
        'display': 'inline-block',
        'vertical-align': 'middle',
        'flex-shrink': '0',
      },
      warn: true,
    }),
  ],
})
