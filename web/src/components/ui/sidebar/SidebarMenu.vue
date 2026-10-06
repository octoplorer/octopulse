<script setup lang="ts">
defineOptions({ name: 'SidebarMenu' })
</script>

<template>
  <ul data-sidebar="menu" class="m-0 flex min-w-0 list-none flex-col items-stretch gap-y-px p-0">
    <slot />
  </ul>
</template>
