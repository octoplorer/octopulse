<script setup lang="ts">
import { computed, ref } from 'vue'
import { useSidebar } from './context'
import SidebarPanelIcon from './SidebarPanelIcon.vue'

defineOptions({ inheritAttrs: false })

const emit = defineEmits<{ click: [event: MouseEvent] }>()
const { open, openMobile, isMobile, toggleSidebar } = useSidebar()
const element = ref<HTMLButtonElement | null>(null)
const expanded = computed(() => isMobile.value ? openMobile.value : open.value)

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
    data-sidebar="trigger"
    data-component="Sidebar"
    data-part="trigger"
    :aria-expanded="expanded"
    :aria-label="expanded ? 'Collapse sidebar' : 'Expand sidebar'"
    class="flex size-8.5 shrink-0 cursor-pointer items-center justify-center rounded-lg text-subtle hover:bg-$sidebar-active-bg hover:text-default focus:outline-none focus-visible:ring-2 focus-visible:ring-brand focus-visible:ring-inset"
    v-bind="$attrs"
    @click="handleClick"
  >
    <slot>
      <SidebarPanelIcon />
    </slot>
  </button>
</template>
