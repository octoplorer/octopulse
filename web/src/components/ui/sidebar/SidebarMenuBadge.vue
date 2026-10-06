<script setup lang="ts">
defineOptions({ name: 'SidebarMenuBadge' })
</script>

<template>
  <span
    data-sidebar="menu-badge"
    class="inline-flex shrink-0 select-none items-center rounded-full border border-dashed border-line px-1.5 py-0.5 text-[11px]/none font-medium text-strong group-data-[state=collapsed]/sidebar:hidden"
  >
    <slot />
  </span>
</template>
