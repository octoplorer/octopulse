import { defineConfig, presetAttributify, presetIcons, presetWind4 } from 'unocss'

export default defineConfig({
  presets: [
    presetWind4(),
    presetAttributify({
      prefix: 'un-',
      prefixedOnly: false,
      // Keep component text props and SVG presentation attributes out of utility extraction.
      ignoreAttributes: ['text', 'font-size', 'stroke-width'],
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
