import {
  defineConfig,
  presetWind4,
  presetAttributify,
  transformerDirectives,
  transformerVariantGroup
} from 'unocss'

export default defineConfig({
  presets: [
    presetWind4(),
    presetAttributify()
  ],
  transformers: [
    transformerDirectives(),
    transformerVariantGroup()
  ]
})
