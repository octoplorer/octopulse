<script setup lang="ts">
import { ref } from 'vue'
import { useSidebar } from './context'

defineOptions({ inheritAttrs: false })

const emit = defineEmits<{ click: [event: MouseEvent] }>()
const { toggleSidebar } = useSidebar()
const element = ref<HTMLButtonElement | null>(null)

function handleClick(event: MouseEvent) {
  emit('click', event)
  toggleSidebar()
}

defineExpose({ element })
</script>

<template>
  <button
    ref="element"
    type="button"
    data-sidebar="rail"
    data-component="Sidebar"
    data-part="rail"
    aria-label="Toggle sidebar"
    :tabindex="-1"
    class="absolute inset-y-0 z-1 hidden w-4 -translate-x-1/2 cursor-pointer transition-all after:absolute after:inset-y-0 after:left-1/2 after:w-0.5 hover:after:bg-brand/20 group-data-[side=left]/sidebar-wrapper:right-0 group-data-[side=right]/sidebar-wrapper:left-0 sm:flex"
    v-bind="$attrs"
    @click="handleClick"
  >
    <slot />
  </button>
</template>
