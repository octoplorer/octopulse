<script setup lang="ts">
import { cva } from 'cva'

withDefaults(defineProps<{
  variant?: 'primary' | 'secondary' | 'error' | 'danger' | 'warning' | 'success' | 'info' | 'beta' | 'outline' | 'muted' | 'red' | 'green' | 'neutral' | 'orange' | 'purple' | 'teal' | 'teal-subtle' | 'blue'
  appearance?: 'filled' | 'dot'
  dot?: boolean
}>(), { variant: 'outline', appearance: 'filled' })
const badgeVariants = cva({ base: 'inline-flex w-fit shrink-0 items-center gap-1.5 whitespace-nowrap rounded-full px-2 py-0.5 text-size-xs font-medium', variants: { variant: {
  'primary': 'bg-badge-inverted text-fg-badge-inverted',
  'secondary': 'bg-fill text-fg-badge-neutral-subtle',
  'error': 'bg-danger-tint text-fg-danger',
  'danger': 'bg-danger-tint text-fg-danger',
  'warning': 'bg-warning-tint text-fg-warning',
  'success': 'bg-success-tint text-fg-success',
  'info': 'bg-info-tint text-fg-info',
  'beta': 'border border-dashed border-brand text-link',
  'outline': 'border border-solid border-line bg-base text-default',
  'muted': 'bg-recessed text-subtle',
  'red': 'bg-badge-red text-white',
  'green': 'bg-badge-green text-white',
  'neutral': 'bg-badge-neutral text-white',
  'orange': 'bg-badge-orange text-black',
  'purple': 'bg-badge-purple text-white',
  'teal': 'bg-badge-teal text-white',
  'teal-subtle': 'bg-badge-teal/10 text-fg-badge-teal-subtle',
  'blue': 'bg-badge-blue text-white',
} } })
const dots = { 'primary': 'bg-contrast', 'secondary': 'bg-subtle', 'error': 'bg-danger', 'danger': 'bg-danger', 'warning': 'bg-warning', 'success': 'bg-success', 'info': 'bg-info', 'beta': 'bg-brand', 'outline': 'bg-subtle', 'muted': 'bg-subtle', 'red': 'bg-badge-red', 'green': 'bg-badge-green', 'neutral': 'bg-badge-neutral', 'orange': 'bg-badge-orange', 'purple': 'bg-badge-purple', 'teal': 'bg-badge-teal', 'teal-subtle': 'bg-badge-teal', 'blue': 'bg-badge-blue' }
</script>

<template>
  <span :class="appearance === 'dot' ? 'inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-size-xs ring-1 ring-hairline text-default' : badgeVariants({ variant })"><span v-if="dot || appearance === 'dot'" class="size-1.5 shrink-0 rounded-full" :class="[dots[variant]]" aria-hidden="true" /><slot /></span>
</template>
