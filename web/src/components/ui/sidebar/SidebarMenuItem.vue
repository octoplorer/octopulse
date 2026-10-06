<script setup lang="ts">
import type { SidebarMenuItemProps } from './menu-types'
import { ref, watchEffect } from 'vue'
import { useSidebar } from './context'
import { provideMenuItemContext } from './menu-context'

defineOptions({ name: 'SidebarMenuItem' })

const props = defineProps<SidebarMenuItemProps>()
const element = ref<HTMLLIElement | null>(null)
const { registerItem } = useSidebar()

provideMenuItemContext()

watchEffect((onCleanup) => {
  if (!props.itemId)
    return
  const id = props.itemId
  registerItem(id, element.value)
  onCleanup(() => registerItem(id, null))
})

defineExpose({ element })
</script>

<template>
  <li
    ref="element"
    data-sidebar="menu-item"
    :data-sidebar-item-id="itemId"
    class="relative group-data-[state=collapsed]/sidebar:overflow-hidden"
  >
    <slot />
  </li>
</template>
