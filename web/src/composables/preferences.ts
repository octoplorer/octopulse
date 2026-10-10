import { useLocalStorage, usePreferredDark } from '@vueuse/core'
import { computed } from 'vue'
import { i18n, t } from './i18n.ts'

export const theme = useLocalStorage<'light' | 'dark' | 'system'>('octopulse.theme', 'system')
export const timezone = useLocalStorage(
  'octopulse.timezone',
  new Intl.DateTimeFormat().resolvedOptions().timeZone,
)
const prefersDark = usePreferredDark()
export const dark = computed(
  () => theme.value === 'dark' || (theme.value === 'system' && prefersDark.value),
)
export function formatDate(value: number | string | undefined | null) {
  if (value == null || value === '' || value === 0)
    return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime()))
    return '—'
  return i18n.global.d(date, {
    key: 'short',
    timeZone: timezone.value,
  })
}
export function formatPercent(value: number | undefined | null) {
  if (value == null || !Number.isFinite(value))
    return '—'
  return i18n.global.n(value > 1 ? value / 100 : value, 'percent')
}
export function duration(milliseconds: number | undefined | null) {
  if (milliseconds == null || !Number.isFinite(milliseconds))
    return '—'
  if (milliseconds < 1000)
    return t('duration.milliseconds', { value: i18n.global.n(milliseconds, 'integer') })
  if (milliseconds < 60000)
    return t('duration.seconds', { value: i18n.global.n(milliseconds / 1000, 'decimal') })
  if (milliseconds < 3600000)
    return t('duration.minutes', { value: i18n.global.n(milliseconds / 60000, 'integer') })
  if (milliseconds < 86400000)
    return t('duration.hours', { value: i18n.global.n(milliseconds / 3600000, 'decimal') })
  return t('duration.days', { value: i18n.global.n(milliseconds / 86400000, 'decimal') })
}
export function datetimeInput(value: number, tz = timezone.value) {
  if (!value)
    return ''
  const parts = new Intl.DateTimeFormat('sv-SE', {
    timeZone: tz,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).format(new Date(value))
  return parts.replace(' ', 'T')
}
export function datetimeMilliseconds(value: string, tz = timezone.value) {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})/.exec(value)
  if (!match)
    throw new Error(t('preferences.invalid-date-and-time'))
  const [, year, month, day, hour, minute] = match
  const target = Date.UTC(+year!, +month! - 1, +day!, +hour!, +minute!)
  let guess = target
  for (let i = 0; i < 3; i++) {
    const parts = new Intl.DateTimeFormat('en-GB', {
      timeZone: tz,
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      hourCycle: 'h23',
    }).formatToParts(new Date(guess))
    const n = (type: string) => Number(parts.find(p => p.type === type)?.value)
    const local = Date.UTC(n('year'), n('month') - 1, n('day'), n('hour'), n('minute'), n('second'))
    guess += target - local
  }
  if (datetimeInput(guess, tz) !== value.slice(0, 16))
    throw new Error(t('preferences.this-wall-clock-time-does-not-exist-in'))
  return guess
}
export function statusLabel(value: string) {
  const labels: Record<string, string> = {
    admin: 'status.admin',
    operator: 'status.operator',
    viewer: 'status.viewer',
    investigating: 'status.investigating',
    identified: 'status.identified',
    monitoring: 'status.monitoring',
    resolved: 'status.resolved',
    none: 'status.none',
    partial: 'status.partial',
    outage: 'status.outage',
    pending: 'status.pending',
    leased: 'status.leased',
    retry: 'status.retry',
    sent: 'status.sent',
    delivered: 'status.delivered',
    discarded: 'status.discarded',
    failed: 'status.failed',
    test: 'status.test',
    down: 'status.down',
    up: 'status.up',
    certificate_expiring: 'status.certificate-expiring',
    certificate_expired: 'status.certificate-expired',
    certificate_renewed: 'status.certificate-renewed',
  }
  return Object.hasOwn(labels, value) ? t(labels[value]!) : value
}
