<script setup lang="ts">
import { Popover as ArkPopover } from '@ark-ui/vue/popover'
</script>

<template>
  <ArkPopover.Title class="text-size-base font-semibold text-strong">
    <slot />
  </ArkPopover.Title>
</template>
