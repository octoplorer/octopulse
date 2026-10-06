<script setup lang="ts">
defineOptions({ name: 'SidebarHeader' })
</script>

<template>
  <div
    data-sidebar="header"
    class="flex h-[58px] shrink-0 items-center gap-1 overflow-hidden border-b border-line px-3"
  >
    <slot />
  </div>
</template>
