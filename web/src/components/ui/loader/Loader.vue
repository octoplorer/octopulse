<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ size?: 'sm' | 'base' | 'lg' | number, label?: string }>(), { size: 'base', label: 'Loading' })
const pixels = computed(() => typeof props.size === 'number' ? props.size : { sm: 16, base: 24, lg: 32 }[props.size])
</script>

<template>
  <span role="status" class="inline-flex items-center gap-2 text-subtle">
    <svg class="kumo-loader shrink-0" :width="pixels" :height="pixels" viewBox="0 0 24 24" stroke="currentColor" aria-hidden="true">
      <circle cx="12" cy="12" r="9.5" fill="none" stroke-width="2" stroke-linecap="round" />
    </svg>
    <span class="sr-only">{{ label }}</span><slot />
  </span>
</template>

<style scoped>
.kumo-loader {
  animation: kumo-loader-rotate 2s linear infinite;
}
.kumo-loader circle {
  stroke-dasharray: 42 150;
  animation: kumo-loader-dash 1.5s ease-in-out infinite;
}
@keyframes kumo-loader-rotate {
  to {
    transform: rotate(360deg);
  }
}
@keyframes kumo-loader-dash {
  0% {
    stroke-dasharray: 0 150;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 42 150;
    stroke-dashoffset: -15;
  }
  100% {
    stroke-dasharray: 42 150;
    stroke-dashoffset: -60;
  }
}
@media (prefers-reduced-motion: reduce) {
  .kumo-loader,
  .kumo-loader circle {
    animation: none;
  }
}
</style>
