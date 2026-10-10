<script setup lang="ts">
import { cva } from 'cva'
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  as?: keyof HTMLElementTagNameMap
  variant?: 'heading' | 'heading1' | 'heading2' | 'heading3' | 'body' | 'secondary' | 'success' | 'error' | 'mono' | 'mono-secondary'
  size?: 'xs' | 'sm' | 'base' | 'lg'
}>(), { as: 'span', variant: 'body', size: 'base' })
const textVariants = cva({
  base: 'text-default',
  variants: {
    variant: {
      'heading': 'font-semibold tracking-tight',
      'heading1': 'text-3xl font-semibold tracking-tight',
      'heading2': 'text-2xl font-semibold tracking-tight',
      'heading3': 'text-lg font-semibold',
      'body': '',
      'secondary': 'text-subtle',
      'success': 'text-fg-success',
      'error': 'text-fg-danger',
      'mono': 'font-mono',
      'mono-secondary': 'font-mono text-subtle',
    },
    size: { xs: 'text-size-xs', sm: 'text-size-sm', base: 'text-size-base', lg: 'text-size-lg' },
  },
})
const classes = computed(() => {
  const variant = props.variant
  const customSize = variant === 'heading' || variant.startsWith('heading') || variant.startsWith('mono')
  return [textVariants({ variant, size: customSize ? undefined : props.size }), variant === 'heading' && (props.size === 'lg' ? 'text-xl' : 'text-size-lg'), variant.startsWith('mono') && (props.size === 'lg' ? 'text-size-base' : 'text-size-sm')]
})
</script>

<template>
  <component :is="as" :class="classes">
    <slot />
  </component>
</template>
