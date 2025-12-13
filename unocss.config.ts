import {
  defineConfig,
  presetAttributify,
  presetIcons,
  presetWind4,
  transformerDirectives,
  transformerVariantGroup,
} from 'unocss'

export default defineConfig({
  presets: [
    presetWind4(),
    presetAttributify({
      ignoreAttributes: [
        'text',
      ],
    }),
    presetIcons(),
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup(),
  ],
})
