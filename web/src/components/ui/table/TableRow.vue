<script setup lang="ts">
import { cva } from 'cva'

const props = withDefaults(defineProps<{ variant?: 'default' | 'selected' }>(), { variant: 'default' })
const styles = cva({ base: 'transition-colors motion-reduce:transition-none', variants: { variant: { default: 'bg-base even:bg-elevated [--table-row-bg:var(--color-base)] even:[--table-row-bg:var(--color-elevated)]', selected: 'bg-tint [--table-row-bg:var(--color-tint)]' } }, defaultVariants: { variant: 'default' } })
</script>

<template>
  <tr :data-selected="props.variant === 'selected' || undefined" :class="styles({ variant: props.variant })">
    <slot />
  </tr>
</template>
