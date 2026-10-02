<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'
const { t } = useI18n({ useScope: 'global' })

const props = defineProps<{ state?: string; paused?: boolean; maintenance?: boolean }>()
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
const label = computed(() => (labels[key.value] ? t(labels[key.value]!) : props.state))
</script>
<template>
  <span class="state-badge" :data-state="key"><i />{{ label }}</span>
</template>
