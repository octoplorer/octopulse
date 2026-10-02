import { expect, test } from 'vitest'
import { computed } from 'vue'
import {
  datetimeInput,
  datetimeMilliseconds,
  duration,
  formatDate,
  formatPercent,
  statusLabel,
  timezone,
} from './preferences.ts'
import { locale } from './i18n.ts'

test('maintenance wall-clock input uses the selected zone, not the process zone', () => {
  const timestamp = datetimeMilliseconds('2026-10-02T09:30', 'Asia/Shanghai')
  expect(timestamp).toBe(Date.UTC(2026, 9, 2, 1, 30))
  expect(datetimeInput(timestamp, 'Asia/Shanghai')).toBe('2026-10-02T09:30')
  expect(datetimeInput(timestamp, 'UTC')).toBe('2026-10-02T01:30')
})

test('nonexistent spring DST times are rejected', () => {
  expect(() => datetimeMilliseconds('2026-03-08T02:30', 'America/New_York')).toThrow()
})

test('normal DST time round-trips in the selected zone', () => {
  const timestamp = datetimeMilliseconds('2026-07-01T15:45', 'America/New_York')
  expect(timestamp).toBe(Date.UTC(2026, 6, 1, 19, 45))
  expect(datetimeInput(timestamp, 'America/New_York')).toBe('2026-07-01T15:45')
})

test('missing statistics never appear as perfect uptime', () => {
  expect(formatPercent(null)).toBe('—')
  expect(formatPercent(undefined)).toBe('—')
  expect(formatPercent(Number.NaN)).toBe('—')
  expect(formatPercent(Infinity)).toBe('—')
  expect(formatPercent(0)).toBe('0.00%')
  expect(formatPercent(1)).toBe('100.00%')
  expect(formatPercent(0.123456)).toBe('12.35%')
  expect(formatPercent(99.99)).toBe('99.99%')
  expect(duration(null)).toBe('—')
  expect(duration(Number.NaN)).toBe('—')
  expect(duration(Infinity)).toBe('—')
})

test('human dates follow the active locale and selected display zone', () => {
  const originalLocale = locale.value
  const originalTimezone = timezone.value
  const timestamp = Date.UTC(2026, 9, 2, 1, 30)
  try {
    for (const language of ['zh-CN', 'en'] as const) {
      locale.value = language
      for (const zone of ['Asia/Shanghai', 'America/New_York', 'UTC']) {
        timezone.value = zone
        const expected = new Intl.DateTimeFormat(language, {
          dateStyle: 'medium',
          timeStyle: 'short',
          timeZone: zone,
        }).format(new Date(timestamp))
        expect(formatDate(timestamp)).toBe(expected)
        expect(formatDate(new Date(timestamp).toISOString())).toBe(expected)
      }
      expect(formatDate(0), 'the API uses zero for an unset timestamp').toBe('—')
    }
    expect(formatDate(undefined)).toBe('—')
    expect(formatDate(null)).toBe('—')
    expect(formatDate('')).toBe('—')
    expect(formatDate('not a timestamp')).toBe('—')
    expect(formatDate(Number.NaN)).toBe('—')
  } finally {
    locale.value = originalLocale
    timezone.value = originalTimezone
  }
})

test('duration units and status labels react to the shared locale', () => {
  const originalLocale = locale.value
  const status = computed(() => statusLabel('investigating'))
  try {
    locale.value = 'en'
    expect(status.value).toBe('Investigating')
    expect(duration(42)).toBe('42 ms')
    expect(duration(1250)).toBe('1.3 s')
    expect(duration(90000)).toBe('2 min')
    expect(duration(5400000)).toBe('1.5 h')
    expect(duration(86400000)).toBe('1.0 d')
    locale.value = 'zh-CN'
    expect(status.value).toBe('调查中')
    expect(duration(42)).toBe('42 毫秒')
    expect(duration(1250)).toBe('1.3 秒')
    expect(duration(90000)).toBe('2 分钟')
    expect(duration(5400000)).toBe('1.5 小时')
    expect(duration(86400000)).toBe('1.0 天')
    expect(statusLabel('new_api_state')).toBe('new_api_state')
    expect(statusLabel('constructor')).toBe('constructor')
  } finally {
    locale.value = originalLocale
  }
})

test('date validation errors use the current language', () => {
  const originalLocale = locale.value
  try {
    locale.value = 'en'
    expect(() => datetimeMilliseconds('invalid')).toThrow(/Invalid date and time/)
    locale.value = 'zh-CN'
    expect(() => datetimeMilliseconds('invalid')).toThrow(/日期时间无效/)
  } finally {
    locale.value = originalLocale
  }
})
