<script setup lang="ts">
import { computed } from 'vue'
import { useSidebar } from './context'

const { open, openMobile, isMobile, side } = useSidebar()
const expanded = computed(() => isMobile.value ? openMobile.value : open.value)
const iconClass = computed(() => {
  if (side.value === 'right') {
    return expanded.value ? 'i-lucide-panel-right-close' : 'i-lucide-panel-right-open'
  }
  return expanded.value ? 'i-lucide-panel-left-close' : 'i-lucide-panel-left-open'
})
</script>

<template>
  <span class="inline-flex size-4.5 shrink-0" aria-hidden="true">
    <slot>
      <span :class="iconClass" class="size-full" />
    </slot>
  </span>
</template>
