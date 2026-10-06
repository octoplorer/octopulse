<script setup lang="ts">
defineOptions({ name: 'SidebarGroup' })
</script>

<template>
  <div data-sidebar="group" class="flex min-w-0 flex-col gap-y-px">
    <slot />
  </div>
</template>
