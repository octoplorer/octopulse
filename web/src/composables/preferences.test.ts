import test from 'node:test'
import assert from 'node:assert/strict'
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
  assert.equal(timestamp, Date.UTC(2026, 9, 2, 1, 30))
  assert.equal(datetimeInput(timestamp, 'Asia/Shanghai'), '2026-10-02T09:30')
  assert.equal(datetimeInput(timestamp, 'UTC'), '2026-10-02T01:30')
})

test('nonexistent spring DST times are rejected', () => {
  assert.throws(() => datetimeMilliseconds('2026-03-08T02:30', 'America/New_York'))
})

test('normal DST time round-trips in the selected zone', () => {
  const timestamp = datetimeMilliseconds('2026-07-01T15:45', 'America/New_York')
  assert.equal(timestamp, Date.UTC(2026, 6, 1, 19, 45))
  assert.equal(datetimeInput(timestamp, 'America/New_York'), '2026-07-01T15:45')
})

test('missing statistics never appear as perfect uptime', () => {
  assert.equal(formatPercent(null), '—')
  assert.equal(formatPercent(undefined), '—')
  assert.equal(formatPercent(Number.NaN), '—')
  assert.equal(formatPercent(Infinity), '—')
  assert.equal(formatPercent(0), '0.00%')
  assert.equal(formatPercent(1), '100.00%')
  assert.equal(formatPercent(0.123456), '12.35%')
  assert.equal(formatPercent(99.99), '99.99%')
  assert.equal(duration(null), '—')
  assert.equal(duration(Number.NaN), '—')
  assert.equal(duration(Infinity), '—')
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
        assert.equal(formatDate(timestamp), expected)
        assert.equal(formatDate(new Date(timestamp).toISOString()), expected)
      }
      assert.equal(formatDate(0), '—', 'the API uses zero for an unset timestamp')
    }
    assert.equal(formatDate(undefined), '—')
    assert.equal(formatDate(null), '—')
    assert.equal(formatDate(''), '—')
    assert.equal(formatDate('not a timestamp'), '—')
    assert.equal(formatDate(Number.NaN), '—')
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
    assert.equal(status.value, 'Investigating')
    assert.equal(duration(42), '42 ms')
    assert.equal(duration(1250), '1.3 s')
    assert.equal(duration(90000), '2 min')
    assert.equal(duration(5400000), '1.5 h')
    assert.equal(duration(86400000), '1.0 d')
    locale.value = 'zh-CN'
    assert.equal(status.value, '调查中')
    assert.equal(duration(42), '42 毫秒')
    assert.equal(duration(1250), '1.3 秒')
    assert.equal(duration(90000), '2 分钟')
    assert.equal(duration(5400000), '1.5 小时')
    assert.equal(duration(86400000), '1.0 天')
    assert.equal(statusLabel('new_api_state'), 'new_api_state')
    assert.equal(statusLabel('constructor'), 'constructor')
  } finally {
    locale.value = originalLocale
  }
})

test('date validation errors use the current language', () => {
  const originalLocale = locale.value
  try {
    locale.value = 'en'
    assert.throws(() => datetimeMilliseconds('invalid'), /Invalid date and time/)
    locale.value = 'zh-CN'
    assert.throws(() => datetimeMilliseconds('invalid'), /日期时间无效/)
  } finally {
    locale.value = originalLocale
  }
})
