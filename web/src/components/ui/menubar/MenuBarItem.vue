<script setup lang="ts">
import { Tabs as ArkTabs } from '@ark-ui/vue/tabs'

const props = defineProps<{ value: string, disabled?: boolean }>()
</script>

<template>
  <ArkTabs.Trigger :value="props.value" :disabled="props.disabled" class="relative rounded-md px-3 py-1.5 text-size-sm font-medium text-subtle outline-none hover:text-default data-[selected]:bg-base data-[selected]:text-default data-[selected]:shadow-control focus:z-3 focus:ring-focus/50 focus-visible:z-3 focus-visible:ring-2 focus-visible:ring-brand disabled:opacity-50">
    <slot />
  </ArkTabs.Trigger>
</template>
