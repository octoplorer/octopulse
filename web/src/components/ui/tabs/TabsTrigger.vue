<script setup lang="ts">
import { TabTrigger } from '@ark-ui/vue/tabs'
import { cva } from 'cva'

withDefaults(defineProps<{ value: string, disabled?: boolean, variant?: 'line' | 'segmented', asChild?: boolean }>(), { variant: 'segmented' })
const variants = cva({ base: 'shrink-0 whitespace-nowrap text-size-base font-medium outline-none disabled:opacity-50 focus:outline-none focus:ring-focus/50 focus-visible:ring-2 focus-visible:ring-brand', variants: { variant: { line: 'border-0 border-b-2 border-solid border-transparent bg-transparent py-3 text-subtle focus-visible:ring-inset data-[selected]:border-b-current data-[selected]:text-strong', segmented: 'rounded-md border-0 bg-transparent px-3 py-1.5 text-subtle data-[selected]:bg-base data-[selected]:text-default data-[selected]:shadow-control' } } })
</script>

<template>
  <TabTrigger :value="value" :disabled="disabled" :as-child="asChild" :class="variants({ variant })">
    <slot />
  </TabTrigger>
</template>
