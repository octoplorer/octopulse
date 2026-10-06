<script setup lang="ts">
import { provideMenuSubItemContext } from './menu-context'

defineOptions({ name: 'SidebarMenuSubItem' })

provideMenuSubItemContext()
</script>

<template>
  <li data-sidebar="menu-sub-item" class="relative">
    <slot />
  </li>
</template>
