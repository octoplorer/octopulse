<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'span' })
</script>

<template>
  <component :is="as" class="field-error" un-text="12px fg-danger">
    <slot />
  </component>
</template>
