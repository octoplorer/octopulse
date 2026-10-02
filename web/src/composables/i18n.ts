import { useStorage } from '@vueuse/core'
import { createI18n } from 'vue-i18n'
import en from '../locales/en.json' with { type: 'json' }
import zhCN from '../locales/zh-CN.json' with { type: 'json' }

type MessageSchema = typeof en

declare module 'vue-i18n' {
  export interface DefineLocaleMessage extends MessageSchema {}
}

export const localeStorageKey = 'octopulse.locale'
export const supportedLocales = ['zh-CN', 'en'] as const
export type AppLocale = (typeof supportedLocales)[number]
export const languageOptions = [
  { value: 'zh-CN', label: '简体中文' },
  { value: 'en', label: 'English' },
] as const satisfies readonly { value: AppLocale, label: string }[]
export const messages = { 'zh-CN': zhCN, en } satisfies Record<AppLocale, MessageSchema>

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
    'en': { short: dateTimeFormat },
  },
  numberFormats: {
    'zh-CN': numberFormats,
    'en': numberFormats,
  },
})

export const locale = useStorage(localeStorageKey, i18n.global.locale)
export const t = i18n.global.t
