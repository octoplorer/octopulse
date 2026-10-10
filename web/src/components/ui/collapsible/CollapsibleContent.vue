<script setup lang="ts">
import { Collapsible as ArkCollapsible } from '@ark-ui/vue/collapsible'
</script>

<template>
  <ArkCollapsible.Content class="overflow-hidden text-size-sm text-default">
    <slot />
  </ArkCollapsible.Content>
</template>
