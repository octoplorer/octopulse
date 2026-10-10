<script setup lang="ts">
import type { Component, ComponentPublicInstance } from 'vue'
import type { TreeNode } from './flow-layout'
import { unrefElement } from '@vueuse/core'
import { computed, inject, onBeforeUnmount, onMounted, provide, useId, useTemplateRef, watch } from 'vue'
import { flowGroupKey, flowNodeKey, useFlowContext } from './context'

const props = withDefaults(defineProps<{ id?: string, disabled?: boolean, as?: string | Component }>(), { disabled: false, as: 'div' })
const context = useFlowContext()
const group = inject(flowGroupKey)!
const generated = useId()
const id = computed(() => props.id ?? generated)
const root = useTemplateRef<HTMLElement | ComponentPublicInstance>('element')
const element = computed(() => {
  const node = unrefElement(root)
  return node?.nodeType === 1 ? node as HTMLElement : undefined
})
const position = computed(() => context.positions.value[id.value] ?? { x: 0, y: 0 })
let observer: ResizeObserver | undefined
let startAnchor: HTMLElement | undefined
let endAnchor: HTMLElement | undefined
const unregister = group.register({ tree: computed<TreeNode>(() => ({ kind: 'node', id: id.value })), element: () => element.value ?? undefined })

function measure() {
  const rect = element.value?.getBoundingClientRect()
  context.reportNode(id.value, {
    width: rect?.width ?? 160,
    height: rect?.height ?? 48,
    disabled: props.disabled,
    startAnchorOffset: rect && startAnchor ? startAnchor.getBoundingClientRect().top + startAnchor.offsetHeight / 2 - rect.top : undefined,
    endAnchorOffset: rect && endAnchor ? endAnchor.getBoundingClientRect().top + endAnchor.offsetHeight / 2 - rect.top : undefined,
  })
}

provide(flowNodeKey, { reportAnchor: (type, anchor) => {
  if (type === 'start' || type === 'both')
    startAnchor = anchor
  if (type === 'end' || type === 'both')
    endAnchor = anchor
  measure()
} })
watch([id, () => props.disabled], (_value, oldValue) => {
  if (oldValue?.[0] && oldValue[0] !== id.value)
    context.removeNode(oldValue[0])
  measure()
}, { immediate: true })

onMounted(() => {
  observer = new ResizeObserver(measure)
  if (element.value)
    observer.observe(element.value)
  measure()
})
onBeforeUnmount(() => {
  observer?.disconnect()
  unregister()
  context.removeNode(id.value)
})
</script>

<template>
  <component :is="as" ref="element" :data-node-id="id" :data-disabled="disabled ? '' : undefined" role="listitem" class="absolute z-1 w-max min-w-32 rounded-lg border border-line bg-base px-4 py-3 text-size-sm text-default shadow-control" :class="{ 'opacity-60': disabled }" :style="{ left: `${position.x}px`, top: `${position.y}px` }">
    <slot />
  </component>
</template>
