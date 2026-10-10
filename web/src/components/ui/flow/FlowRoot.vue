<script setup lang="ts">
import type { FlowAlign, FlowOrientation, FlowState } from './flow-layout'
import { computed, onBeforeUnmount, onMounted, onUpdated, provide, shallowRef, useId, useTemplateRef } from 'vue'
import { createFlowGroup, flowGroupChildren, flowKey, provideFlowGroup } from './context'
import { computeDiagramRect, computeEdges, computePositions } from './flow-layout'

const props = withDefaults(defineProps<{
  orientation?: FlowOrientation
  canvas?: boolean
  align?: FlowAlign
  padding?: { x?: number, y?: number }
  columnGap?: number
  rowGap?: number
  ariaLabel?: string
}>(), { orientation: 'horizontal', canvas: true, align: 'start', columnGap: 64, rowGap: 16, ariaLabel: 'Workflow diagram' })
const emit = defineEmits<{ overflowChange: [overflow: { x: boolean, y: boolean }] }>()
const group = createFlowGroup()
const nodes = shallowRef<FlowState['nodes']>({})
const orderVersion = shallowRef(0)
const wrapper = useTemplateRef('wrapper')
const content = useTemplateRef('content')
const markerId = `flow-${useId()}`
const dragging = shallowRef(false)
const state = computed<FlowState>(() => {
  // DOM order is authoritative after keyed slot children are moved.
  void orderVersion.value
  return { nodes: nodes.value, tree: { kind: 'list', children: flowGroupChildren(group) }, orientation: props.orientation, align: props.align }
})
const positions = computed(() => computePositions(state.value, { columnGap: props.columnGap, rowGap: props.rowGap }))
const bounds = computed(() => computeDiagramRect(positions.value, state.value))
const padding = computed(() => ({ x: props.padding?.x ?? 16, y: props.padding?.y ?? 64 }))
const connectors = computed(() => computeEdges(state.value).flatMap(([source, target]) => {
  const from = positions.value[source]
  const to = positions.value[target]
  const fromSize = nodes.value[source]
  const toSize = nodes.value[target]
  if (!from || !to || !fromSize || !toSize)
    return []
  const vertical = props.orientation === 'vertical'
  const x1 = from.x + (vertical ? fromSize.width / 2 : fromSize.width)
  const y1 = from.y + (vertical ? fromSize.height : fromSize.startAnchorOffset ?? fromSize.height / 2)
  const x2 = to.x + (vertical ? toSize.width / 2 : 0)
  const y2 = to.y + (vertical ? 0 : toSize.endAnchorOffset ?? toSize.height / 2)
  const middle = vertical ? (y1 + y2) / 2 : (x1 + x2) / 2
  return [{ id: `${source}:${target}`, disabled: fromSize.disabled || toSize.disabled, path: vertical ? `M${x1},${y1}C${x1},${middle} ${x2},${middle} ${x2},${y2}` : `M${x1},${y1}C${middle},${y1} ${middle},${y2} ${x2},${y2}` }]
}))

provideFlowGroup(group)
provide(flowKey, {
  orderVersion,
  positions,
  orientation: computed(() => props.orientation),
  reportNode: (id, measurement) => {
    const previous = nodes.value[id]
    if (!previous || Object.keys(measurement).some(key => previous[key as keyof typeof previous] !== measurement[key as keyof typeof measurement]))
      nodes.value = { ...nodes.value, [id]: measurement }
  },
  removeNode: (id) => {
    const next = { ...nodes.value }
    delete next[id]
    nodes.value = next
  },
})

let observer: ResizeObserver | undefined
let dragStart: { x: number, y: number, left: number, top: number } | undefined
let previousOverflow = { x: false, y: false }
let previousOrder = ''

onUpdated(() => {
  const order = JSON.stringify(Array.from(content.value?.querySelectorAll('[data-node-id]') ?? [], node => node.getAttribute('data-node-id')))
  if (order !== previousOrder) {
    previousOrder = order
    orderVersion.value++
  }
})

function measureOverflow() {
  const element = wrapper.value
  if (!element)
    return
  const next = { x: element.scrollWidth > element.clientWidth + 1, y: element.scrollHeight > element.clientHeight + 1 }
  if (next.x !== previousOverflow.x || next.y !== previousOverflow.y) {
    previousOverflow = next
    emit('overflowChange', next)
  }
}

function startPan(event: PointerEvent) {
  if (!props.canvas || event.button !== 0 || (event.target as Element).closest('[data-node-id]') || !wrapper.value)
    return
  dragStart = { x: event.clientX, y: event.clientY, left: wrapper.value.scrollLeft, top: wrapper.value.scrollTop }
  dragging.value = true
  wrapper.value.setPointerCapture(event.pointerId)
}

function pan(event: PointerEvent) {
  if (dragStart && wrapper.value) {
    wrapper.value.scrollLeft = dragStart.left + dragStart.x - event.clientX
    wrapper.value.scrollTop = dragStart.top + dragStart.y - event.clientY
  }
}

function stopPan() {
  dragStart = undefined
  dragging.value = false
}

function keydown(event: KeyboardEvent) {
  if (event.target !== wrapper.value || !['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key))
    return
  event.preventDefault()
  wrapper.value?.scrollBy({ left: event.key === 'ArrowLeft' ? -64 : event.key === 'ArrowRight' ? 64 : 0, top: event.key === 'ArrowUp' ? -64 : event.key === 'ArrowDown' ? 64 : 0 })
}

onMounted(() => {
  observer = new ResizeObserver(measureOverflow)
  if (wrapper.value)
    observer.observe(wrapper.value)
  if (content.value)
    observer.observe(content.value)
  measureOverflow()
})
onBeforeUnmount(() => observer?.disconnect())

defineExpose({ scrollTo: (options: ScrollToOptions) => wrapper.value?.scrollTo(options) })
</script>

<template>
  <div ref="wrapper" class="relative max-w-full outline-none focus:outline-none focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand" :class="canvas ? ['overflow-auto rounded-xl border border-line bg-recessed', { 'cursor-grabbing select-none': dragging }] : 'overflow-visible'" :tabindex="canvas ? 0 : undefined" :role="canvas ? 'region' : undefined" :aria-label="ariaLabel" :style="{ padding: `${padding.y}px ${padding.x}px` }" @pointerdown="startPan" @pointermove="pan" @pointerup="stopPan" @pointercancel="stopPan" @lostpointercapture="stopPan" @keydown="keydown">
    <div ref="content" class="relative" :style="{ width: `${bounds.width}px`, height: `${bounds.height}px`, minWidth: '100%' }" role="list">
      <svg class="pointer-events-none absolute inset-0 overflow-visible" :width="bounds.width" :height="bounds.height" aria-hidden="true">
        <defs>
          <marker :id="markerId" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse"><path d="M0 0L10 5L0 10Z" fill="var(--color-brand)" /></marker>
          <marker :id="`${markerId}-disabled`" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="5" markerHeight="5" orient="auto-start-reverse"><path d="M0 0L10 5L0 10Z" fill="var(--color-interact)" /></marker>
        </defs>
        <path v-for="edge in connectors" :key="edge.id" :d="edge.path" fill="none" :stroke="edge.disabled ? 'var(--color-interact)' : 'var(--color-brand)'" stroke-width="1.5" :marker-end="`url(#${markerId}${edge.disabled ? '-disabled' : ''})`" />
      </svg>
      <slot />
    </div>
  </div>
</template>
