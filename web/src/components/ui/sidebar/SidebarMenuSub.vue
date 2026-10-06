<script setup lang="ts">
defineOptions({ name: 'SidebarMenuSub' })
</script>

<template>
  <ul
    data-sidebar="menu-sub"
    class="relative m-0 flex min-w-0 list-none flex-col gap-y-px overflow-hidden p-0 pl-7 pr-0"
  >
    <div aria-hidden="true" class="absolute inset-y-px left-[19px] z-10 w-px bg-line" />
    <slot />
  </ul>
</template>
