import antfu from '@antfu/eslint-config'

export default antfu({
  typescript: true,
  vue: true,
  ignores: ['dist/**', 'src/client/**', 'typed-router.d.ts', 'aube-lock.yaml'],
  formatters: {
    css: true,
    html: true,
    prettierOptions: { printWidth: 100 },
  },
  rules: {
    'vue/quote-props': 'off',
  },
})
