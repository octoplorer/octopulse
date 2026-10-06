<script setup lang="ts">
defineOptions({ name: 'SidebarFooter' })
</script>

<template>
  <div
    data-sidebar="footer"
    class="sticky bottom-0 flex h-12 w-$sidebar-width shrink-0 items-center gap-4 overflow-hidden border-t border-line bg-$sidebar-bg bg-clip-padding px-[11px] whitespace-nowrap transition-[width,padding] duration-$sidebar-animation-duration ease-$sidebar-easing group-not-data-[state=collapsed]/sidebar:px-4 group-data-[state=collapsed]/sidebar:w-$sidebar-width-icon group-data-[state=collapsed]/sidebar:border-r group-data-[state=collapsed]/sidebar:border-line motion-reduce:transition-none"
  >
    <slot />
  </div>
</template>
