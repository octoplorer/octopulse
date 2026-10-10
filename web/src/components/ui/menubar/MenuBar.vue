<script setup lang="ts">
import { Tabs as ArkTabs } from '@ark-ui/vue/tabs'

const props = withDefaults(defineProps<{ label?: string, orientation?: 'horizontal' | 'vertical' }>(), { orientation: 'horizontal' })
const value = defineModel<string | undefined>({ default: undefined })
</script>

<template>
  <ArkTabs.Root v-model="value" :orientation="props.orientation" class="min-w-0">
    <ArkTabs.List :aria-label="props.label" class="flex w-fit gap-1 rounded-lg bg-tint p-1" :class="props.orientation === 'vertical' ? 'flex-col' : 'items-center'">
      <slot />
    </ArkTabs.List><slot name="content" />
  </ArkTabs.Root>
</template>
