import { t } from '../composables/i18n'

type StateVariant = 'success' | 'danger' | 'info' | 'warning' | 'muted'

const labels: Record<string, string> = {
  up: 'states.up',
  down: 'states.down',
  unknown: 'states.unknown',
  paused: 'states.paused',
  maintenance: 'states.maintenance',
  healthy: 'states.healthy',
  expiring: 'states.expiring',
  expired: 'states.expired',
  check_failed: 'states.check-failed',
  normal: 'states.normal',
  operational: 'states.operational',
  partial_outage: 'states.partial-outage',
  full_outage: 'states.full-outage',
  insufficient_data: 'states.insufficient-data',
}

const variants: Record<string, StateVariant> = {
  up: 'success',
  normal: 'success',
  operational: 'success',
  healthy: 'success',
  down: 'danger',
  expired: 'danger',
  full_outage: 'danger',
  outage: 'danger',
  maintenance: 'info',
  expiring: 'warning',
  partial_outage: 'warning',
  partial: 'warning',
}

export function monitorStateDisplay(
  value: string | null | undefined,
  options: { paused?: boolean, maintenance?: boolean } = {},
) {
  const state = options.paused
    ? 'paused'
    : options.maintenance
      ? 'maintenance'
      : (value || 'unknown').toLowerCase()
  return {
    state,
    label: Object.hasOwn(labels, state) ? t(labels[state]!) : value,
    variant: Object.hasOwn(variants, state) ? variants[state]! : 'muted' as const,
  }
}
