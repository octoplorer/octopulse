<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'span' })
</script>

<template>
  <component :is="as" class="field-error" un-text="12px $danger">
    <slot />
  </component>
</template>
