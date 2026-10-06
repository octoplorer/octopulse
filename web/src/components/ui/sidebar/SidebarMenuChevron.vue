<script setup lang="ts">
import { useSidebarCollapsible } from './collapsible-context'

const { isOpen, isCollapsible } = useSidebarCollapsible()
</script>

<template>
  <span
    aria-hidden="true"
    class="i-lucide-chevron-right ml-auto size-3 shrink-0 opacity-40 [transition:transform_200ms,opacity_200ms] group-hover/menu-button:opacity-100 group-data-[state=collapsed]/sidebar:hidden motion-reduce:transition-none"
    :class="{ '[transform:rotate(90deg)]': isCollapsible && isOpen }"
  />
</template>
