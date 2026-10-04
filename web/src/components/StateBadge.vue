<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Badge } from './ui/badge'

const props = defineProps<{ state?: string, paused?: boolean, maintenance?: boolean }>()

const { t } = useI18n({ useScope: 'global' })

const key = computed(() =>
  props.paused
    ? 'paused'
    : props.maintenance
      ? 'maintenance'
      : (props.state || 'unknown').toLowerCase(),
)
const labels: Record<string, string> = {
  up: 'states.up',
  down: 'states.down',
  unknown: 'states.unknown',
  paused: 'states.paused',
  maintenance: 'states.maintenance',
  healthy: 'states.healthy',
  expiring: 'states.expiring',
  expired: 'states.expired',
  check_failed: 'states.check_failed',
  normal: 'states.normal',
  operational: 'states.operational',
  partial_outage: 'states.partial_outage',
  full_outage: 'states.full_outage',
  insufficient_data: 'states.insufficient_data',
}
const appearance: Record<string, 'success' | 'danger' | 'info' | 'warning' | 'muted'> = {
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
const label = computed(() => (labels[key.value] ? t(labels[key.value]!) : props.state))
</script>

<template>
  <Badge :variant="appearance[key] || 'muted'" :data-state="key" dot>
    {{ label }}
  </Badge>
</template>
