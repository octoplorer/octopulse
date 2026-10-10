<script setup lang="ts">
import { Popover as ArkPopover } from '@ark-ui/vue/popover'
</script>

<template>
  <ArkPopover.Description class="mt-1 text-size-sm text-subtle">
    <slot />
  </ArkPopover.Description>
</template>
