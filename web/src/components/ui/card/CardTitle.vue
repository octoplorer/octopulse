<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'h2' })
</script>

<template>
  <component :is="as" class="card-title">
    <slot />
  </component>
</template>
