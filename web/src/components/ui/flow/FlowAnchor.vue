<script setup lang="ts">
import { inject, onBeforeUnmount, onMounted, useTemplateRef } from 'vue'
import { flowNodeKey } from './context'

const props = withDefaults(defineProps<{ type?: 'start' | 'end' | 'both' }>(), { type: 'both' })
const context = inject(flowNodeKey)
const element = useTemplateRef<HTMLElement>('element')
let observer: ResizeObserver | undefined
onMounted(() => {
  observer = new ResizeObserver(() => context?.reportAnchor(props.type, element.value ?? undefined))
  if (element.value)
    observer.observe(element.value)
  context?.reportAnchor(props.type, element.value ?? undefined)
})
onBeforeUnmount(() => {
  observer?.disconnect()
  context?.reportAnchor(props.type, undefined)
})
</script>

<template>
  <span ref="element" class="inline-block"><slot /></span>
</template>
