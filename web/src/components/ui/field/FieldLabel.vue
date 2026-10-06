<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'span' })
</script>

<template>
  <component :is="as" class="field-label" un-text="14px default" font="500">
    <slot />
  </component>
</template>
