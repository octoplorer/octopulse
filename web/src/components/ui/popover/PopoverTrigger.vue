<script setup lang="ts">
import { Popover as ArkPopover } from '@ark-ui/vue/popover'

const props = defineProps<{ asChild?: boolean }>()
</script>

<template>
  <ArkPopover.Trigger :as-child="props.asChild" :class="props.asChild ? undefined : 'inline-flex items-center justify-center rounded-lg'">
    <slot />
  </ArkPopover.Trigger>
</template>
