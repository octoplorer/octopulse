<script setup lang="ts">
import { TabList } from '@ark-ui/vue/tabs'
import { cva } from 'cva'

withDefaults(defineProps<{ variant?: 'line' | 'segmented', asChild?: boolean }>(), { variant: 'segmented' })
const variants = cva({ base: 'flex min-w-0 max-w-full items-center overflow-auto', variants: { variant: { line: 'gap-6 border-b border-solid border-line bg-base px-5 sm:px-6 max-sm:gap-4', segmented: 'w-fit gap-1 rounded-lg bg-recessed p-1' } } })
</script>

<template>
  <TabList :as-child="asChild" :class="variants({ variant })">
    <slot />
  </TabList>
</template>
