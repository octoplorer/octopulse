import test from 'node:test'
import assert from 'node:assert/strict'
import { computed } from 'vue'
import { createI18n } from 'vue-i18n'
import { i18n, locale, messages, supportedLocales, t } from './i18n.ts'

function flattenMessages(catalog: object, prefix = ''): Record<string, string> {
  const result: Record<string, string> = {}
  for (const [key, value] of Object.entries(catalog)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (typeof value === 'string') result[path] = value
    else Object.assign(result, flattenMessages(value, path))
  }
  return result
}

test('every supported catalog has the same keys and compiles through Vue I18n', () => {
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
      assert.deepEqual(Object.keys(catalog).sort(), englishKeys)
      locale.value = language
      for (const key of Object.keys(catalog)) {
        const translated = t(key, { count: 2, label: 'Example', value: 42, shown: 1, total: 2 })
        assert.ok(translated.length > 0, `${language}: ${key} must translate`)
        assert.notEqual(translated, key, `${language}: ${key} must resolve its message`)
      }
    }
    assert.deepEqual(warnings, [], 'catalogs must compile without warnings')
    assert.deepEqual(errors, [], 'catalogs must compile without errors')
  } finally {
    locale.value = originalLocale
    console.warn = originalWarn
    console.error = originalError
  }
})

test('translations react to locale changes, interpolate values, and pluralize counts', () => {
  const originalLocale = locale.value
  const translated = computed(() => t('common.monitors'))
  try {
    locale.value = 'en'
    assert.equal(translated.value, 'Monitors')
    assert.equal(t('overview.showingMonitors', { shown: 2, total: 5 }), 'Showing 2 of 5 monitors')
    assert.equal(t('counts.monitors', 0), 'No monitors')
    assert.equal(t('counts.monitors', 1), '1 monitor')
    assert.equal(t('counts.monitors', 2), '2 monitors')
    locale.value = 'zh-CN'
    assert.equal(translated.value, '监控项')
    assert.equal(t('counts.monitors', 2), '2 个监控项')
  } finally {
    locale.value = originalLocale
  }
})

test('literal JSON and at signs are handled by the message compiler', () => {
  const originalLocale = locale.value
  try {
    locale.value = 'en'
    assert.equal(
      t('monitorDetails.postReportsStatusUpOrStatusDownWith'),
      'POST reports: {"status":"up"} or {"status":"down"}, with optional description.',
    )
    const composer = createI18n({
      legacy: false,
      locale: 'en',
      messages: {
        en: {
          ...messages.en,
          common: { ...messages.en.common, monitors: "support{'@'}example.com" },
        },
      },
    }).global
    assert.equal(composer.t('common.monitors'), 'support@example.com')
  } finally {
    locale.value = originalLocale
  }
})

test('the configured fallback supplies translations when a locale has no catalog', () => {
  const composer = createI18n({
    legacy: false,
    locale: 'fr',
    fallbackLocale: i18n.global.fallbackLocale.value,
    messages,
    missingWarn: false,
    fallbackWarn: false,
  }).global
  assert.equal(composer.t('common.monitors'), '监控项')
})
