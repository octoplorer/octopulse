import { computed, watch, type Ref, type WatchStopHandle } from 'vue'
import { createI18n } from 'vue-i18n'
import en from '../locales/en.json' with { type: 'json' }
import zhCN from '../locales/zh-CN.json' with { type: 'json' }

type AppLocale = 'zh-CN' | 'en'
type MessageSchema = typeof en

declare module 'vue-i18n' {
  export interface DefineLocaleMessage extends MessageSchema {}
}

export const localeStorageKey = 'octopulse.locale'
export const supportedLocales = ['zh-CN', 'en'] as const
export const languageOptions = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en', label: 'English' },
] as const
export const messages = { 'zh-CN': zhCN, en } satisfies Record<AppLocale, MessageSchema>

type LocaleStorage = Pick<Storage, 'getItem' | 'setItem'>

/** Saved preferences use the same locale identifiers as the API. */
export function resolveLocale(
  savedLocale: unknown,
  browserLocales: readonly string[] = [],
): AppLocale {
  if (savedLocale != null) return savedLocale === 'en' ? 'en' : 'zh-CN'
  for (const language of browserLocales) {
    const base = language.toLowerCase().split('-')[0]
    if (base === 'zh') return 'zh-CN'
    if (base === 'en') return 'en'
  }
  return 'zh-CN'
}

/** Keep the composer's locale as the only reactive language preference. */
export function syncLocalePreference(
  localeRef: Ref<string>,
  storage?: LocaleStorage,
  browserLocales: readonly string[] = [],
): WatchStopHandle {
  let savedLocale: string | null = null
  try {
    savedLocale = storage?.getItem(localeStorageKey) ?? null
  } catch {
    // Browsers can disable storage; translations still work in memory.
  }
  localeRef.value = resolveLocale(savedLocale, browserLocales)
  return watch(
    localeRef,
    (value) => {
      const supportedLocale = resolveLocale(value)
      if (value !== supportedLocale) {
        localeRef.value = supportedLocale
        return
      }
      try {
        storage?.setItem(localeStorageKey, supportedLocale)
      } catch {
        // A blocked or full storage area must not prevent language changes.
      }
    },
    { immediate: true, flush: 'sync' },
  )
}

const dateTimeFormat = { dateStyle: 'medium', timeStyle: 'short' } as const
const numberFormats = {
  percent: { style: 'percent', minimumFractionDigits: 2, maximumFractionDigits: 2 },
  integer: { maximumFractionDigits: 0, useGrouping: false },
  decimal: { minimumFractionDigits: 1, maximumFractionDigits: 1, useGrouping: false },
} as const

export const i18n = createI18n({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages,
  datetimeFormats: {
    'zh-CN': { short: dateTimeFormat },
    en: { short: dateTimeFormat },
  },
  numberFormats: {
    'zh-CN': numberFormats,
    en: numberFormats,
  },
})

export const locale = computed<AppLocale>({
  get: () => resolveLocale(i18n.global.locale.value),
  set: (value) => {
    i18n.global.locale.value = resolveLocale(value)
  },
})
export const t = i18n.global.t

if (typeof window !== 'undefined') {
  let storage: LocaleStorage | undefined
  try {
    storage = window.localStorage
  } catch {
    // Storage access can throw before getItem in privacy restricted browsers.
  }
  syncLocalePreference(i18n.global.locale, storage, window.navigator.languages)
  window.addEventListener('storage', (event) => {
    if (event.key === localeStorageKey || event.key === null) {
      locale.value = resolveLocale(event.newValue, window.navigator.languages)
    }
  })
}
