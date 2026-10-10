<script setup lang="ts">
withDefaults(defineProps<{ orientation?: 'horizontal' | 'vertical', variant?: 'joined' | 'separated', label?: string }>(), { orientation: 'horizontal', variant: 'joined' })
</script>

<template>
  <div role="group" :aria-label="label" class="inline-flex" :class="[orientation === 'vertical' && 'flex-col', variant === 'separated' ? 'gap-2' : orientation === 'horizontal' ? '[&>button:not(:first-child)]:rounded-l-none [&>button:not(:last-child)]:rounded-r-none [&>button+button]:-ml-px' : '[&>button:not(:first-child)]:rounded-t-none [&>button:not(:last-child)]:rounded-b-none [&>button+button]:-mt-px']">
    <slot />
  </div>
</template>
