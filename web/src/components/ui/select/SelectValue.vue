<script setup lang="ts">
import { Select as ArkSelect } from '@ark-ui/vue/select'

defineProps<{ placeholder?: string }>()
</script>

<template>
  <ArkSelect.ValueText class="select-value" :placeholder="placeholder" />
</template>

<style scoped>
.select-value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.select-value[data-placeholder] {
  color: var(--text-color-placeholder);
}
</style>
