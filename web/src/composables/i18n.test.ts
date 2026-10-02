import { expect, it } from 'vitest'
import { computed } from 'vue'
import { createI18n } from 'vue-i18n'
import { i18n, locale, messages, supportedLocales, t } from './i18n.ts'

function flattenMessages(catalog: object, prefix = ''): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(catalog)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'string')
      result[path] = value
    else Object.assign(result, flattenMessages(value, path))
  }
  return result
}

it('every supported catalog has the same keys and compiles through Vue I18n', () => {
  const originalLocale = locale.value
  const warnings: unknown[][] = []
  const errors: unknown[][] = []
  const originalWarn = console.warn
  const originalError = console.error
  console.warn = (...args) => warnings.push(args)
  console.error = (...args) => errors.push(args)
  try {
    const englishKeys = Object.keys(flattenMessages(messages.en)).sort()
    for (const language of supportedLocales) {
      const catalog = flattenMessages(messages[language])
      expect(Object.keys(catalog).sort()).toStrictEqual(englishKeys)
      locale.value = language
      for (const key of Object.keys(catalog)) {
        const translated = t(key, { count: 2, label: 'Example', value: 42, shown: 1, total: 2 })
        expect(translated.length, `${language}: ${key} must translate`).toBeGreaterThan(0)
        expect(translated, `${language}: ${key} must resolve its message`).not.toBe(key)
      }
    }
    expect(warnings, 'catalogs must compile without warnings').toStrictEqual([])
    expect(errors, 'catalogs must compile without errors').toStrictEqual([])
  }
  finally {
    locale.value = originalLocale
    console.warn = originalWarn
    console.error = originalError
  }
})

it('translations react to locale changes, interpolate values, and pluralize counts', () => {
  const originalLocale = locale.value
  const translated = computed(() => t('common.monitors'))
  try {
    locale.value = 'en'
    expect(translated.value).toBe('Monitors')
    expect(t('overview.showingMonitors', { shown: 2, total: 5 })).toBe('Showing 2 of 5 monitors')
    expect(t('counts.monitors', 0)).toBe('No monitors')
    expect(t('counts.monitors', 1)).toBe('1 monitor')
    expect(t('counts.monitors', 2)).toBe('2 monitors')
    locale.value = 'zh-CN'
    expect(translated.value).toBe('监控项')
    expect(t('counts.monitors', 2)).toBe('2 个监控项')
  }
  finally {
    locale.value = originalLocale
  }
})

it('literal JSON and at signs are handled by the message compiler', () => {
  const originalLocale = locale.value
  try {
    locale.value = 'en'
    expect(t('monitorDetails.postReportsStatusUpOrStatusDownWith')).toBe(
      'POST reports: {"status":"up"} or {"status":"down"}, with optional description.',
    )
    const composer = createI18n({
      legacy: false,
      locale: 'en',
      messages: {
        en: {
          ...messages.en,
          common: { ...messages.en.common, monitors: 'support{\'@\'}example.com' },
        },
      },
    }).global
    expect(composer.t('common.monitors')).toBe('support@example.com')
  }
  finally {
    locale.value = originalLocale
  }
})

it('the configured fallback supplies translations when a locale has no catalog', () => {
  const composer = createI18n({
    legacy: false,
    locale: 'fr',
    fallbackLocale: i18n.global.fallbackLocale.value,
    messages,
    missingWarn: false,
    fallbackWarn: false,
  }).global
  expect(composer.t('common.monitors')).toBe('监控项')
})
