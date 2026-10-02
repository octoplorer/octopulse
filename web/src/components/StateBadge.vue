<script setup lang="ts">
import { computed } from 'vue'
import { t } from '../lib/preferences'
const props = defineProps<{ state?: string; paused?: boolean; maintenance?: boolean }>()
const key = computed(() =>
  props.paused
    ? 'paused'
    : props.maintenance
      ? 'maintenance'
      : (props.state || 'unknown').toLowerCase(),
)
const labels: Record<string, [string, string]> = {
  up: ['正常', 'Operational'],
  down: ['故障', 'Down'],
  unknown: ['等待数据', 'Unknown'],
  paused: ['已暂停', 'Paused'],
  maintenance: ['维护中', 'Maintenance'],
  healthy: ['证书有效', 'Valid certificate'],
  expiring: ['即将到期', 'Expiring'],
  expired: ['已过期', 'Expired'],
  check_failed: ['检查失败', 'Check failed'],
  normal: ['全部正常', 'All systems operational'],
  operational: ['全部正常', 'All systems operational'],
  partial_outage: ['部分故障', 'Partial outage'],
  full_outage: ['全面故障', 'Major outage'],
  insufficient_data: ['数据不足', 'Insufficient data'],
}
const label = computed(() => (labels[key.value] ? t(...labels[key.value]!) : props.state))
</script>
<template>
  <span class="state-badge" :data-state="key"><i />{{ label }}</span>
</template>
