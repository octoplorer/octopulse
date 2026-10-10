<script setup lang="ts">
import { cva } from 'cva'

const props = withDefaults(defineProps<{ as?: string, variant?: 'heading' | 'secondary', size?: 'base' | 'lg' }>(), { as: 'h3', variant: 'heading', size: 'lg' })
const styles = cva({ base: 'text-balance', variants: { variant: { heading: 'font-semibold text-strong', secondary: 'font-normal text-subtle' }, size: { base: 'text-size-base', lg: 'text-size-lg' } }, defaultVariants: { variant: 'heading', size: 'lg' } })
</script>

<template>
  <component :is="props.as" :class="styles({ variant: props.variant, size: props.size })">
    <slot />
  </component>
</template>
