<script setup lang="ts">
import { Select as ArkSelect } from '@ark-ui/vue/select'

withDefaults(defineProps<{ label?: string }>(), { label: 'Clear selection' })
</script>

<template>
  <ArkSelect.ClearTrigger :aria-label="label" class="inline-flex size-24px shrink-0 items-center justify-center rounded-md text-subtle hover:bg-tint outline-none focus:ring-focus focus-visible:ring-2 focus-visible:ring-brand disabled:opacity-50">
    <slot><span class="i-lucide-x size-3.5" aria-hidden="true" /></slot>
  </ArkSelect.ClearTrigger>
</template>
