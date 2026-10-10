<script setup lang="ts">
import { Popover as ArkPopover } from '@ark-ui/vue/popover'

const props = defineProps<{ asChild?: boolean, label?: string }>()
</script>

<template>
  <ArkPopover.CloseTrigger :as-child="props.asChild" :class="props.asChild ? undefined : 'inline-flex h-7 w-7 items-center justify-center rounded-md text-subtle hover:bg-tint'" :aria-label="props.label ?? 'Close popover'">
    <slot><span class="i-lucide-x size-4" aria-hidden="true" /></slot>
  </ArkPopover.CloseTrigger>
</template>
