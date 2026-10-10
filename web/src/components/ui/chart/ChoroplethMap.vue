<script setup lang="ts" generic="T">
import type { ChoroplethMapProps } from './types'
import { registerMap } from 'echarts/core'
import { computed, useId, useTemplateRef, watch } from 'vue'
import Chart from './Chart.vue'
import { createChoroplethMapOption } from './model'

const props = defineProps<ChoroplethMapProps<T>>()
const emit = defineEmits<{ regionClick: [row: T], regionHover: [row: T | undefined] }>()
const chart = useTemplateRef('chart')
const id = useId()
const mapName = computed(() => props.mapName ?? `choropleth-map-${id}`)
watch(() => [mapName.value, props.geoJson] as const, ([name, geoJson]) => registerMap(name, geoJson as Parameters<typeof registerMap>[1]), { immediate: true })
const option = computed(() => createChoroplethMapOption({ ...props, mapName: mapName.value }))

function select(event: { data?: { datum?: T } }) {
  if (event.data?.datum !== undefined)
    emit('regionClick', event.data.datum)
}
function hover(event: { data?: { datum?: T } }) {
  emit('regionHover', event.data?.datum)
}

defineExpose({ getEchartsInstance: () => chart.value?.getEchartsInstance(), resize: () => chart.value?.resize() })
</script>

<template>
  <Chart ref="chart" :option="option" :height="height" :loading="loading" :aria-label="ariaLabel" :style="aspectRatio ? { aspectRatio, height: height === undefined ? 'auto' : `${height}px` } : undefined" @click="select" @mouseover="hover" @mouseout="emit('regionHover', undefined)" />
</template>
