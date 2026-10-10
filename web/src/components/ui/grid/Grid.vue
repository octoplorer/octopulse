<script setup lang="ts">
import { cva } from 'cva'

withDefaults(defineProps<{ variant?: '2up' | 'side-by-side' | '2-1' | '1-2' | '1-3up' | '3up' | '4up' | '6up' | '1-2-4up', gap?: 'none' | 'sm' | 'base' | 'lg' }>(), { variant: '2up', gap: 'base' })
const gridVariants = cva({ base: 'grid min-w-0', variants: {
  variant: { '2up': 'grid-cols-1 md:grid-cols-2', 'side-by-side': 'grid-cols-2', '2-1': 'grid-cols-1 md:grid-cols-[2fr_1fr]', '1-2': 'grid-cols-1 md:grid-cols-[1fr_2fr]', '1-3up': 'grid-cols-1 lg:grid-cols-3', '3up': 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3', '4up': 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4', '6up': 'grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6', '1-2-4up': 'grid-cols-1 md:grid-cols-2 lg:grid-cols-4' },
  gap: { none: 'gap-0', sm: 'gap-3', base: 'gap-2 md:gap-6 lg:gap-8', lg: 'gap-8' },
} })
</script>

<template>
  <div :class="gridVariants({ variant, gap })">
    <slot />
  </div>
</template>
