<script setup lang="ts">
import { ark } from '@ark-ui/vue/factory'
import { cva } from 'cva'
import { computed, onBeforeUnmount, onMounted, provide, useTemplateRef } from 'vue'
import { toolbarSizeKey } from './context'

const props = withDefaults(defineProps<{ label: string, orientation?: 'horizontal' | 'vertical', size?: 'sm' | 'base' | 'lg', loopFocus?: boolean }>(), { orientation: 'horizontal', size: 'base', loopFocus: true })
const element = useTemplateRef<HTMLElement>('element')
const styles = cva({ base: 'flex rounded-lg bg-base text-default ring-1 ring-line', variants: { size: { sm: 'gap-1 p-1', base: 'gap-1.5 p-1.5', lg: 'gap-2 p-2' }, orientation: { horizontal: 'items-center', vertical: 'flex-col items-stretch' } }, defaultVariants: { size: 'base', orientation: 'horizontal' } })
provide(toolbarSizeKey, computed(() => props.size))
let observer: MutationObserver | undefined
function controls() {
  return Array.from(element.value?.querySelectorAll<HTMLElement>('[data-toolbar-control]') ?? []).filter(control => !control.hasAttribute('disabled') && control.getAttribute('aria-disabled') !== 'true' && !control.hidden)
}
function updateTabStops(active?: HTMLElement) {
  const enabled = controls()
  const current = active ?? enabled.find(control => control === document.activeElement) ?? enabled.find(control => control.tabIndex === 0) ?? enabled[0]
  element.value?.querySelectorAll<HTMLElement>('[data-toolbar-control]').forEach((control) => {
    control.tabIndex = control === current ? 0 : -1
  })
}
function onFocus(event: FocusEvent) {
  const target = event.target as HTMLElement
  if (target.matches('[data-toolbar-control]'))
    updateTabStops(target)
}
function onKeydown(event: KeyboardEvent) {
  const target = event.target as HTMLElement
  if (target.matches('input, textarea, select') || event.altKey || event.ctrlKey || event.metaKey)
    return
  const enabled = controls()
  const index = enabled.indexOf(target)
  if (index < 0)
    return
  const forward = props.orientation === 'horizontal' ? 'ArrowRight' : 'ArrowDown'
  const backward = props.orientation === 'horizontal' ? 'ArrowLeft' : 'ArrowUp'
  let next = index
  if (event.key === 'Home')
    next = 0
  else if (event.key === 'End')
    next = enabled.length - 1
  else if (event.key === forward)
    next++
  else if (event.key === backward)
    next--
  else return
  event.preventDefault()
  next = props.loopFocus ? (next + enabled.length) % enabled.length : Math.max(0, Math.min(next, enabled.length - 1))
  enabled[next]?.focus()
  updateTabStops(enabled[next])
}
onMounted(() => {
  updateTabStops()
  observer = new MutationObserver(() => updateTabStops())
  if (element.value)
    observer.observe(element.value, { childList: true, subtree: true, attributes: true, attributeFilter: ['disabled', 'aria-disabled', 'hidden'] })
})
onBeforeUnmount(() => observer?.disconnect())
</script>

<template>
  <ark.div as-child>
    <div ref="element" role="toolbar" :aria-label="props.label" :aria-orientation="props.orientation" :class="styles({ size: props.size, orientation: props.orientation })" @keydown="onKeydown" @focusin="onFocus">
      <slot />
    </div>
  </ark.div>
</template>
