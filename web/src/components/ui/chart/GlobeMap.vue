<script setup lang="ts">
import type { GlobeMapMarker, GlobeMapProps } from './types'
import { computed, getCurrentInstance, onBeforeUnmount, onMounted, shallowRef, watch } from 'vue'
import Chart from './Chart.vue'
import { createGlobeOption, globeViewport, globeVisibleMarkers } from './globe-model'

const props = withDefaults(defineProps<GlobeMapProps>(), {
  markers: () => [],
  draggable: true,
  autoRotate: false,
  autoRotateSpeed: 4,
  markerRadius: 7,
  showGraticule: false,
  showTooltip: true,
  ariaLabel: 'Interactive globe map',
  landHatchSpacing: 10,
})
const emit = defineEmits<{ markerClick: [marker: GlobeMapMarker], userRotationChange: [rotation: [number, number, number]] }>()
const root = shallowRef<HTMLElement>()
const chart = shallowRef<InstanceType<typeof Chart>>()
const rotation = shallowRef<[number, number, number]>(props.defaultRotation ? [...props.defaultRotation] : [0, 0, 0])
const dimensions = shallowRef({ width: 400, height: props.height ?? 400 })
const dragging = shallowRef(false)
const reducedMotion = shallowRef(false)
const selected = shallowRef<GlobeMapMarker>()
const instance = getCurrentInstance()
const interactiveMarkers = computed(() => !!instance?.vnode.props?.onMarkerClick)
const option = computed(() => createGlobeOption(props, rotation.value))
const visibleMarkers = computed(() => globeVisibleMarkers(props, rotation.value))
const viewport = computed(() => globeViewport(dimensions.value.width, dimensions.value.height))
let pointer: { x: number, y: number, rotation: [number, number, number] } | undefined
let frame = 0
let lastFrame = 0
let media: MediaQueryList | undefined
let observer: ResizeObserver | undefined
let dragged = false

function updateRotation(next: [number, number, number]) {
  rotation.value = [((next[0] + 180) % 360 + 360) % 360 - 180, Math.max(-90, Math.min(90, next[1])), next[2]]
  selected.value = undefined
  emit('userRotationChange', [...rotation.value])
}

function markerAtEvent(event: PointerEvent) {
  const bounds = root.value?.getBoundingClientRect()
  if (!bounds)
    return false
  const x = event.clientX - bounds.left
  const y = event.clientY - bounds.top
  return visibleMarkers.value.some(point => Math.hypot(x - viewport.value.x - point.x * viewport.value.scale, y - viewport.value.y - point.y * viewport.value.scale) <= (point.marker.radius ?? props.markerRadius) * viewport.value.scale)
}

function startDrag(event: PointerEvent) {
  if (!props.draggable || event.button !== 0 || (event.target as Element).closest('[data-globe-marker]') || markerAtEvent(event))
    return
  pointer = { x: event.clientX, y: event.clientY, rotation: [...rotation.value] }
  dragging.value = true
  dragged = false
  ;(event.currentTarget as Element).setPointerCapture?.(event.pointerId)
}

function drag(event: PointerEvent) {
  if (!pointer)
    return
  dragged ||= Math.hypot(event.clientX - pointer.x, event.clientY - pointer.y) > 3
  updateRotation([pointer.rotation[0] + (event.clientX - pointer.x) * 0.4, pointer.rotation[1] - (event.clientY - pointer.y) * 0.4, pointer.rotation[2]])
}

function stopDrag() {
  pointer = undefined
  dragging.value = false
}

function keydown(event: KeyboardEvent) {
  if (!props.draggable || (event.target as Element).closest('[data-globe-marker]') || !['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home'].includes(event.key))
    return
  event.preventDefault()
  const [longitude, latitude, roll] = rotation.value
  updateRotation(event.key === 'Home' ? [...props.defaultRotation ?? [0, 0, 0]] : [longitude + (event.key === 'ArrowLeft' ? -5 : event.key === 'ArrowRight' ? 5 : 0), latitude + (event.key === 'ArrowUp' ? 5 : event.key === 'ArrowDown' ? -5 : 0), roll])
}

interface MarkerChartEvent { seriesId?: string, data?: { datum?: GlobeMapMarker } }

function selectChartMarker(event: MarkerChartEvent) {
  if (event.seriesId === 'globe-markers' && event.data?.datum && !dragged)
    emit('markerClick', event.data.datum)
  dragged = false
}

function hoverChartMarker(event: MarkerChartEvent) {
  selected.value = event.seriesId === 'globe-markers' ? event.data?.datum : undefined
}

function markerStyle(point: typeof visibleMarkers.value[number]) {
  const size = Math.max(24, (point.marker.radius ?? props.markerRadius) * viewport.value.scale * 2)
  return { left: `${viewport.value.x + point.x * viewport.value.scale}px`, top: `${viewport.value.y + point.y * viewport.value.scale}px`, width: `${size}px`, height: `${size}px` }
}

function animate(time: number) {
  if (props.autoRotate && !reducedMotion.value && !dragging.value && lastFrame) {
    const elapsed = Math.min((time - lastFrame) / 1000, 0.1)
    rotation.value = [rotation.value[0] + elapsed * props.autoRotateSpeed, rotation.value[1], rotation.value[2]]
  }
  lastFrame = time
  frame = requestAnimationFrame(animate)
}

function readMotion() {
  reducedMotion.value = media?.matches ?? false
}

function syncAnimation() {
  cancelAnimationFrame(frame)
  lastFrame = 0
  frame = props.autoRotate && !reducedMotion.value ? requestAnimationFrame(animate) : 0
}
watch([() => props.autoRotate, reducedMotion], syncAnimation)

onMounted(() => {
  media = window.matchMedia('(prefers-reduced-motion: reduce)')
  readMotion()
  media.addEventListener('change', readMotion)
  observer = new ResizeObserver(([entry]) => {
    if (entry && entry.contentRect.width && entry.contentRect.height)
      dimensions.value = { width: entry.contentRect.width, height: entry.contentRect.height }
  })
  if (root.value)
    observer.observe(root.value)
  syncAnimation()
})
onBeforeUnmount(() => {
  cancelAnimationFrame(frame)
  media?.removeEventListener('change', readMotion)
  observer?.disconnect()
})

defineExpose({
  getRotation: () => [...rotation.value],
  setRotation: updateRotation,
  getEchartsInstance: () => chart.value?.getEchartsInstance(),
})
</script>

<template>
  <div
    ref="root"
    class="relative w-full select-none text-default outline-none focus-visible:rounded-xl focus:outline-none focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand"
    :class="{ 'cursor-grab touch-none': draggable, 'cursor-grabbing': dragging }"
    :style="{ height: height ? `${height}px` : undefined, aspectRatio: height ? undefined : '1' }"
    :role="draggable || interactiveMarkers ? 'group' : 'img'"
    :tabindex="draggable ? 0 : undefined"
    :aria-label="ariaLabel"
    @pointerdown="startDrag"
    @pointermove="drag"
    @pointerup="stopDrag"
    @pointercancel="stopDrag"
    @lostpointercapture="stopDrag"
    @keydown="keydown"
  >
    <Chart ref="chart" :option="option" :height="height" :style="{ height: '100%' }" class="absolute inset-0" :update-options="{ lazyUpdate: true }" @click="selectChartMarker" @mouseover="hoverChartMarker" @mouseout="selected = undefined" />
    <template v-if="interactiveMarkers">
      <button
        v-for="point in visibleMarkers"
        :key="point.index"
        type="button"
        data-globe-marker
        :aria-label="point.marker.description ? `${point.marker.name}: ${point.marker.description}` : point.marker.name"
        :style="markerStyle(point)"
        class="absolute cursor-pointer rounded-full border-none bg-transparent p-0 outline-none -translate-x-1/2 -translate-y-1/2 focus:outline-none focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand"
        @click="emit('markerClick', point.marker)"
        @pointerenter="selected = point.marker"
        @pointerleave="selected = undefined"
        @focus="selected = point.marker"
        @blur="selected = undefined"
      />
    </template>
    <div v-if="showTooltip && selected" role="tooltip" class="pointer-events-none absolute bottom-3 left-1/2 max-w-64 rounded-lg border border-line bg-base px-3 py-2 text-size-sm shadow-panel -translate-x-1/2">
      <div class="font-medium">
        {{ selected.name }}
      </div>
      <div v-if="selected.description" class="mt-1 text-subtle">
        {{ selected.description }}
      </div>
    </div>
  </div>
</template>
