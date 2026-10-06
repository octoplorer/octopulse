<script setup lang="ts">
withDefaults(defineProps<{ as?: string }>(), { as: 'span' })
</script>

<template>
  <component :is="as" class="field-hint" un-text="12px subtle" leading="[1.6]">
    <slot />
  </component>
</template>
