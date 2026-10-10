<script setup lang="ts">
import type { SankeyChartProps, SankeyLinkData, SankeyNodeData } from './types'
import { computed, useTemplateRef } from 'vue'
import Chart from './Chart.vue'
import { createSankeyOption } from './model'

const props = defineProps<SankeyChartProps>()
const emit = defineEmits<{ nodeClick: [node: SankeyNodeData], linkClick: [link: SankeyLinkData] }>()
const chart = useTemplateRef('chart')
const option = computed(() => createSankeyOption(props))

function select(event: { dataType?: string, data?: { datum?: SankeyNodeData | SankeyLinkData } }) {
  if (!event.data?.datum)
    return
  if (event.dataType === 'edge')
    emit('linkClick', event.data.datum as SankeyLinkData)
  else
    emit('nodeClick', event.data.datum as SankeyNodeData)
}

defineExpose({ getEchartsInstance: () => chart.value?.getEchartsInstance(), resize: () => chart.value?.resize() })
</script>

<template>
  <Chart ref="chart" :option="option" :height="height" :loading="loading" :aria-label="ariaLabel" @click="select" />
</template>
