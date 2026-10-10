<script setup lang="ts">
import type { TimeseriesChartProps } from './types'
import { computed, getCurrentInstance, useTemplateRef } from 'vue'
import Chart from './Chart.vue'
import { createTimeseriesOption } from './model'

const props = defineProps<TimeseriesChartProps>()
const emit = defineEmits<{ timeRangeChange: [from: number, to: number] }>()
const chart = useTemplateRef('chart')
const instance = getCurrentInstance()
const option = computed(() => createTimeseriesOption({ ...props, selectable: props.selectable ?? !!instance?.vnode.props?.onTimeRangeChange }))

function selectRange(event: { areas?: { coordRange?: number[] }[] }) {
  const range = event.areas?.[0]?.coordRange
  if (range && range.length === 2 && range.every(Number.isFinite))
    emit('timeRangeChange', Math.min(range[0]!, range[1]!), Math.max(range[0]!, range[1]!))
}

defineExpose({
  getEchartsInstance: () => chart.value?.getEchartsInstance(),
  resize: () => chart.value?.resize(),
  dispatchAction: (...args: Parameters<NonNullable<typeof chart.value>['dispatchAction']>) => chart.value?.dispatchAction(...args),
})
</script>

<template>
  <Chart ref="chart" :option="option" :height="height" :loading="loading" :aria-label="ariaDescription" :update-options="optionUpdateBehavior" @brushend="selectRange" />
</template>
