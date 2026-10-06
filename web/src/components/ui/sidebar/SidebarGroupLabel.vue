<script setup lang="ts">
defineOptions({ name: 'SidebarGroupLabel' })
</script>

<template>
  <div
    data-sidebar="group-label"
    class="my-3 grid grid-rows-[0fr] overflow-hidden border-b border-line transition-[grid-template-rows,margin,border-color] duration-$sidebar-animation-duration ease-$sidebar-easing [[data-sidebar=group]:first-child_&]:my-0 [[data-sidebar=group]:first-child_&]:border-transparent group-not-data-[state=collapsed]/sidebar:my-0 group-not-data-[state=collapsed]/sidebar:grid-rows-[1fr] group-not-data-[state=collapsed]/sidebar:border-transparent group-data-[mobile=true]/sidebar:my-0 group-data-[mobile=true]/sidebar:grid-rows-[1fr] group-data-[mobile=true]/sidebar:border-transparent group-data-[mobile=true]/sidebar:transition-none"
  >
    <div class="min-h-0 min-w-0">
      <div class="mb-2 mt-4 truncate px-3 text-sm font-medium text-subtle [[data-sidebar=group]:first-child_&]:mt-2">
        <slot />
      </div>
    </div>
  </div>
</template>
