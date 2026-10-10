<script setup lang="ts">
import { useMutationObserver, useResizeObserver } from '@vueuse/core'
import { onMounted, ref, useId } from 'vue'

defineOptions({ inheritAttrs: false })
withDefaults(defineProps<{ scrollLabel?: string }>(), { scrollLabel: 'Scroll to see more columns' })
const viewport = ref<HTMLElement | null>(null)
const overflowing = ref(false)
const hintId = useId()
function stringAttribute(value: unknown) {
  return typeof value === 'string' ? value : undefined
}
function tabIndexAttribute(value: unknown) {
  return typeof value === 'string' || typeof value === 'number' ? value : undefined
}
function updateOverflow() {
  overflowing.value = !!viewport.value && viewport.value.scrollWidth > viewport.value.clientWidth + 1
}
useResizeObserver(viewport, updateOverflow)
useMutationObserver(viewport, updateOverflow, { childList: true, subtree: true, characterData: true })
onMounted(updateOverflow)
</script>

<template>
  <div class="min-w-0">
    <p v-if="overflowing" :id="hintId" class="flex items-center gap-2 border-b border-line px-4 py-2 text-size-xs text-subtle">
      <span class="i-lucide-arrow-left-right size-4 shrink-0" aria-hidden="true" />{{ scrollLabel }}
    </p>
    <div ref="viewport" v-bind="$attrs" class="table-wrap max-w-full overflow-auto overscroll-contain focus-visible:outline-2 focus-visible:outline-solid focus-visible:outline-focus focus-visible:outline-offset-[-2px]" :tabindex="tabIndexAttribute($attrs.tabindex) ?? (overflowing ? 0 : undefined)" :role="stringAttribute($attrs.role) ?? (overflowing ? 'region' : undefined)" :aria-label="stringAttribute($attrs['aria-label']) ?? (overflowing ? scrollLabel : undefined)" :aria-describedby="overflowing ? [stringAttribute($attrs['aria-describedby']), hintId].filter(Boolean).join(' ') : stringAttribute($attrs['aria-describedby'])">
      <slot />
    </div>
  </div>
</template>
