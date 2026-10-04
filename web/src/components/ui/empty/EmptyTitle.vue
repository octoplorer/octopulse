<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'h3' })
</script>

<template>
  <component :is="as" :class="as === 'h3' ? 'text-16px' : undefined">
    <slot />
  </component>
</template>
