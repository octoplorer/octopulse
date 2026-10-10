<script setup lang="ts">
import Button from './Button.vue'

withDefaults(defineProps<{ loading?: boolean, label?: string }>(), { label: 'Refresh' })
</script>

<template>
  <Button shape="square" :aria-label="label" :disabled="loading">
    <span class="i-lucide-refresh-cw size-4" :class="loading && 'animate-spin'" aria-hidden="true" />
  </Button>
</template>
