import antfu from '@antfu/eslint-config'

export default antfu({
  typescript: true,
  vue: true,
  ignores: ['web/dist/**', 'web/src/client/**', 'web/typed-router.d.ts'],
  formatters: {
    css: true,
    html: true,
    prettierOptions: { printWidth: 100 },
  },
  rules: {
    'vue/quote-props': 'off',
  },
})
