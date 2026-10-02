<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed } from 'vue'

const { t, n } = useI18n({ useScope: 'global' })

const props = defineProps<{
  values: number[]
  timestamps?: number[]
  height?: number
  color?: string
  showScale?: boolean
}>()
const scale = computed(() => {
  const values = props.values.filter(Number.isFinite)
  return { max: Math.max(...values, 1), min: Math.min(...values, 0), last: values.at(-1) }
})
const metric = (value: number) => n(value, { maximumFractionDigits: 2 })
const segments = computed(() => {
  const entries = props.values
    .map((value, index) => ({ value, at: props.timestamps?.[index] ?? index }))
    .filter((p) => Number.isFinite(p.value) && Number.isFinite(p.at))
  if (!entries.length) return []
  const max = Math.max(...entries.map((p) => p.value), 1),
    min = Math.min(...entries.map((p) => p.value), 0),
    spread = max - min || 1,
    from = Math.min(...entries.map((p) => p.at)),
    to = Math.max(...entries.map((p) => p.at)),
    span = to - from || 1
  const gaps = entries
      .slice(1)
      .map((p, i) => p.at - entries[i]!.at)
      .filter((g) => g > 0)
      .sort((a, b) => a - b),
    gapLimit = props.timestamps && gaps.length ? gaps[Math.floor(gaps.length / 2)]! * 3 : Infinity
  const paths: string[][] = [[]]
  entries.forEach((p, index) => {
    if (index && p.at - entries[index - 1]!.at > gapLimit) paths.push([])
    paths.at(-1)!.push(`${((p.at - from) / span) * 300},${54 - ((p.value - min) / spread) * 44}`)
  })
  return paths.map((points) => points.join(' '))
})
</script>
<template>
  <svg
    v-if="segments.length"
    class="sparkline"
    :style="{ height: `${height || 60}px`, color: color || 'var(--accent)' }"
    viewBox="0 0 300 60"
    preserveAspectRatio="none"
    role="img"
    :aria-label="t('sparkline.observedMetricTrend')"
  >
    <title>
      {{
        t('sparkline.summary', {
          count: values.length,
          value: scale.last == null ? '—' : metric(scale.last),
        })
      }}
    </title>
    <template v-if="showScale">
      <text x="2" y="9" fill="var(--muted)" font-size="8">{{ metric(scale.max) }}</text>
      <text x="2" y="59" fill="var(--muted)" font-size="8">{{ metric(scale.min) }}</text>
    </template>
    <circle
      v-for="(points, index) in segments.filter((segment) => !segment.includes(' '))"
      :key="`point-${index}`"
      :cx="points.split(',')[0]"
      :cy="points.split(',')[1]"
      r="2"
      fill="currentColor"
    />
    <polyline
      v-for="(points, index) in segments"
      :key="index"
      :points="points"
      fill="none"
      stroke="currentColor"
      stroke-width="2.5"
      stroke-linecap="round"
      vector-effect="non-scaling-stroke"
    />
  </svg>
  <div v-else class="chart-empty">—</div>
</template>
