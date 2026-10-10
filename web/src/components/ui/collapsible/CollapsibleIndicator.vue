<script setup lang="ts">
import { Collapsible as ArkCollapsible } from '@ark-ui/vue/collapsible'
</script>

<template>
  <ArkCollapsible.Indicator class="inline-flex text-subtle transition-transform data-[state=open]:rotate-90">
    <slot><span class="i-lucide-chevron-right size-4" aria-hidden="true" /></slot>
  </ArkCollapsible.Indicator>
</template>
