<script setup lang="ts">
import { useSidebar } from './context'
import { Button } from './placeholders'

defineOptions({ inheritAttrs: false })

const emit = defineEmits<{ click: [event: MouseEvent] }>()
const { setOpenMobile } = useSidebar()

function handleClick(event: MouseEvent) {
  emit('click', event)
  setOpenMobile(false)
}
</script>

<template>
  <Button
    shape="square"
    data-sidebar="close"
    data-component="Sidebar"
    data-part="close"
    aria-label="Close navigation"
    class="shrink-0"
    v-bind="$attrs"
    @click="handleClick"
  >
    <slot>
      <span class="i-lucide-x size-4.5" aria-hidden="true" />
    </slot>
  </Button>
</template>
