<script setup lang="ts">
const props = withDefaults(defineProps<{ variant?: 'default' | 'compact', sticky?: boolean }>(), { variant: 'default' })
</script>

<template>
  <thead class="group/header" :data-compact="props.variant === 'compact' ? '' : undefined" :class="[props.variant === 'compact' ? '[&_th]:bg-elevated [&_th]:py-2 [&_th]:text-size-xs' : '', props.sticky ? '[&_th]:sticky [&_th]:top-0 [&_th]:z-1' : '']">
    <slot />
  </thead>
</template>
