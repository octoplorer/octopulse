<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'
import { computed, useId } from 'vue'

const props = withDefaults(defineProps<{
  label: string
  value: number
  min?: number
  max?: number
  showValue?: boolean
  customValue?: string
  variant?: 'brand' | 'success' | 'warning' | 'danger'
}>(), { min: 0, max: 100, showValue: true, variant: 'brand' })
const labelId = useId()
const upper = computed(() => Math.max(props.min, props.max))
const measured = computed(() => Math.min(upper.value, Math.max(props.min, Number.isFinite(props.value) ? props.value : props.min)))
const percent = computed(() => upper.value === props.min ? 0 : (measured.value - props.min) / (upper.value - props.min) * 100)
const fills = { brand: 'bg-brand', success: 'bg-success', warning: 'bg-warning', danger: 'bg-danger' }
</script>

<template>
  <ark.div role="meter" :aria-labelledby="labelId" :aria-valuenow="measured" :aria-valuemin="min" :aria-valuemax="upper" :aria-valuetext="customValue" class="flex w-full flex-col gap-2">
    <div class="flex items-center justify-between gap-4">
      <span :id="labelId" class="text-size-xs text-subtle">{{ label }}</span>
      <span v-if="customValue || showValue" class="text-size-sm font-medium tabular-nums">{{ customValue || `${Math.round(percent)}%` }}</span>
    </div>
    <div class="h-2 overflow-hidden rounded-full bg-fill" aria-hidden="true">
      <div class="h-full rounded-full transition-[width] duration-300" :class="[fills[variant]]" :style="{ width: `${percent}%` }" />
    </div>
  </ark.div>
</template>
