import { computed } from 'vue'
import { useLocalStorage, usePreferredDark } from '@vueuse/core'
export const locale = useLocalStorage<'zh-CN' | 'en'>('octopulse.locale', 'zh-CN')
export const theme = useLocalStorage<'light' | 'dark' | 'system'>('octopulse.theme', 'system')
export const timezone = useLocalStorage(
  'octopulse.timezone',
  Intl.DateTimeFormat().resolvedOptions().timeZone,
)
const prefersDark = usePreferredDark()
export const dark = computed(
  () => theme.value === 'dark' || (theme.value === 'system' && prefersDark.value),
)
export const t = (zh: string, en: string) => (locale.value === 'zh-CN' ? zh : en)
export function formatDate(value: number | string | undefined) {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(locale.value, {
    dateStyle: 'medium',
    timeStyle: 'short',
    timeZone: timezone.value,
  }).format(date)
}
export function formatPercent(value: number | undefined | null) {
  return value == null ? '—' : `${(value > 1 ? value : value * 100).toFixed(2)}%`
}
export function duration(milliseconds: number | undefined) {
  if (milliseconds == null) return '—'
  if (milliseconds < 1000) return `${Math.round(milliseconds)} ms`
  if (milliseconds < 60000) return `${(milliseconds / 1000).toFixed(1)} s`
  if (milliseconds < 3600000) return `${Math.round(milliseconds / 60000)} min`
  if (milliseconds < 86400000) return `${(milliseconds / 3600000).toFixed(1)} h`
  return `${(milliseconds / 86400000).toFixed(1)} d`
}
export function datetimeInput(value: number, tz = timezone.value) {
  if (!value) return ''
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
  if (!match) throw new Error(t('日期时间无效', 'Invalid date and time'))
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
    const n = (type: string) => Number(parts.find((p) => p.type === type)?.value)
    const local = Date.UTC(n('year'), n('month') - 1, n('day'), n('hour'), n('minute'), n('second'))
    guess += target - local
  }
  if (datetimeInput(guess, tz) !== value.slice(0, 16))
    throw new Error(
      t(
        '该时区中不存在此时间，请检查夏令时转换。',
        'This wall-clock time does not exist in the selected time zone.',
      ),
    )
  return guess
}
export function statusLabel(value: string) {
  const labels: Record<string, [string, string]> = {
    admin: ['管理员', 'Administrator'],
    operator: ['操作员', 'Operator'],
    viewer: ['只读', 'Viewer'],
    investigating: ['调查中', 'Investigating'],
    identified: ['已定位', 'Identified'],
    monitoring: ['观察中', 'Monitoring'],
    resolved: ['已解决', 'Resolved'],
    none: ['仅信息', 'Informational'],
    partial: ['部分影响', 'Partial impact'],
    outage: ['全面故障', 'Major outage'],
    pending: ['待投递', 'Pending'],
    leased: ['投递中', 'Delivering'],
    retry: ['等待重试', 'Retry pending'],
    sent: ['已投递', 'Sent'],
    delivered: ['已投递', 'Delivered'],
    discarded: ['已跳过', 'Skipped'],
    failed: ['失败', 'Failed'],
    test: ['渠道测试', 'Channel test'],
    down: ['故障', 'Down'],
    up: ['正常', 'Up'],
    certificate_expiring: ['证书即将到期', 'Certificate expiring'],
    certificate_expired: ['证书已过期', 'Certificate expired'],
    certificate_renewed: ['证书续期', 'Certificate renewed'],
  }
  return labels[value] ? t(...labels[value]!) : value
}
