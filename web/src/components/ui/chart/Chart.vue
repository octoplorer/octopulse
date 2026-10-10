<script setup lang="ts">
import type { EChartsOption } from 'echarts'
import type { init, SetOptionOpts } from 'echarts/core'
import type { StyleValue } from 'vue'
import type { Exposed } from 'vue-echarts'
import { computed, onBeforeUnmount, onMounted, shallowRef, useAttrs } from 'vue'
import VChart from 'vue-echarts'
import {
  createChartColorResolver,
  defaultChartAppearance,
  prepareChartOption,
  readChartAppearance,
  resolveChartColors,
} from '@/lib/echarts.ts'
import './register'

defineOptions({ inheritAttrs: false })

const props = withDefaults(defineProps<{
  option: EChartsOption
  height?: number
  ariaLabel?: string
  loading?: boolean
  theme?: Parameters<typeof init>[1]
  initOptions?: Omit<NonNullable<Parameters<typeof init>[2]>, 'renderer'> & { renderer?: 'canvas' }
  updateOptions?: SetOptionOpts
}>(), { height: 350, loading: false })

const root = shallowRef<HTMLElement>()
const chart = shallowRef<Exposed>()
const ready = shallowRef(false)
const appearance = shallowRef(defaultChartAppearance)
const attrs = useAttrs()

function forwardedAttrs() {
  return Object.fromEntries(Object.entries(attrs).filter(([key]) => key !== 'class' && key !== 'style'))
}

function rootStyle(): StyleValue {
  return [{ height: `${props.height}px` }, attrs.style as StyleValue]
}
const option = computed(() => prepareChartOption(
  props.option,
  appearance.value,
  root.value ? createChartColorResolver(root.value) : undefined,
  props.ariaLabel,
))
const theme = computed(() => {
  const appTheme = appearance.value.dark ? 'dark' : undefined
  if (props.theme && typeof props.theme === 'object' && root.value)
    return resolveChartColors(props.theme, createChartColorResolver(root.value))
  return props.theme ?? appTheme
})
const initOptions = computed(() => ({ ...props.initOptions, renderer: 'canvas' as const }))
const loadingOptions = computed(() => ({
  text: '',
  color: appearance.value.palette[0],
  textColor: appearance.value.text,
  maskColor: 'transparent',
  showSpinner: !appearance.value.reducedMotion,
}))

let observer: MutationObserver | undefined
let motion: MediaQueryList | undefined

function refreshAppearance() {
  if (root.value)
    appearance.value = readChartAppearance(root.value, motion?.matches ?? false)
}

onMounted(() => {
  motion = window.matchMedia('(prefers-reduced-motion: reduce)')
  motion.addEventListener('change', refreshAppearance)
  observer = new MutationObserver(refreshAppearance)
  // Watching ancestors catches both the app color mode and locally edited brand colors.
  for (let element = root.value; element; element = element.parentElement ?? undefined)
    observer.observe(element, { attributes: true, attributeFilter: ['style', 'class', 'data-mode'] })
  refreshAppearance()
  ready.value = true
})

onBeforeUnmount(() => {
  observer?.disconnect()
  motion?.removeEventListener('change', refreshAppearance)
})

defineExpose({
  getEchartsInstance: () => chart.value?.chart,
  resize: (...args: Parameters<Exposed['resize']>) => chart.value?.resize(...args),
  dispatchAction: (...args: Parameters<Exposed['dispatchAction']>) => chart.value?.dispatchAction(...args),
})
</script>

<template>
  <div ref="root" class="echart" :class="attrs.class" :style="rootStyle()" :aria-busy="loading" :aria-label="ariaLabel" :role="ariaLabel ? 'img' : undefined">
    <VChart
      v-if="ready"
      ref="chart"
      v-bind="forwardedAttrs()"
      :option="option"
      :theme="theme"
      :init-options="initOptions"
      :update-options="updateOptions"
      :autoresize="{ throttle: 100 }"
      :loading="loading"
      :loading-options="loadingOptions"
    />
  </div>
</template>

<style scoped>
.echart {
  width: 100%;
  min-width: 0;
}
</style>
